package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	// Embute a base de fusos (zoneinfo) no binário, para que
	// time.LoadLocation funcione mesmo em imagens sem o pacote tzdata (ex.: alpine).
	_ "time/tzdata"

	// IMPORTANTE: Import anonimo para carregar a documentação do Swagger gerada no diretório docs/
	_ "github.com/prefeitura-rio/app-go-api/docs"

	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/observability"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/router"
	"github.com/prefeitura-rio/app-go-api/internal/workers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// @title API Go
// @version 1.0
// @description API de serviços para aplicativos da Prefeitura do Rio
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://iplan.rio
// @contact.email frederico.zolio@prefeitura.rio

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Autenticação usando token JWT no formato 'Bearer {token}'

// @Security BearerAuth
func main() {
	log.Println("Carregando configurações...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Fixa o fuso do processo (time.Local). O driver de banco (pgx) decodifica
	// timestamptz usando time.Local; sem isso o processo fica em UTC e a API
	// serializa as datas com "Z" em vez do offset de Brasília (-03:00).
	setupTimezone(cfg.App.Timezone)

	// Initialize OpenTelemetry tracing
	if err := observability.InitTracer(cfg); err != nil {
		log.Fatalf("Erro ao inicializar tracer: %v", err)
	}

	// Run database migrations if enabled.
	// Wire owns the primary DB connection; we open a short-lived connection
	// here solely for migrations, then close it.
	if cfg.Migrations.Run {
		runMigrations(cfg)
	}

	// Lifecycle context: cancelled on SIGINT/SIGTERM, which stops all background workers.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Set up router (Wire initializes all dependencies internally)
	r, err := router.SetupRouter(ctx, cfg)
	if err != nil {
		log.Fatalf("Erro ao inicializar router: %v", err)
	}

	// Initialize orgao sync worker if enabled (uses its own DB/Redis connections)
	if cfg.OrgaoSync.Enabled {
		log.Println("Initializing orgao sync worker...")

		rmiClient := clients.NewRMIClient(cfg.RMI.BaseURL, 30*time.Second)
		workerRedis := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		defer func() {
			if err := workerRedis.Close(); err != nil {
				log.Printf("Erro ao fechar conexão Redis do worker: %v", err)
			}
		}()

		// Open a dedicated DB connection for the worker
		workerDB, err := openDB(cfg)
		if err != nil {
			log.Fatalf("Erro ao conectar ao banco de dados para worker: %v", err)
		}

		orgaoSnapshotRepo := repository.NewOrgaoSnapshotRepository(workerDB)
		cursoRepo := repository.NewCursoRepository(workerDB)
		empregoRepo := repository.NewEmpregoRepository(workerDB)
		oportunidadeMEIRepo := repository.NewOportunidadeMEIRepository(workerDB)

		worker := workers.NewOrgaoSyncWorker(
			workerDB,
			rmiClient,
			workerRedis,
			orgaoSnapshotRepo,
			cursoRepo,
			empregoRepo,
			oportunidadeMEIRepo,
			&cfg.OrgaoSync,
		)

		go func() {
			if err := worker.Start(ctx); err != nil && err != context.Canceled {
				log.Printf("Orgao sync worker error: %v", err)
			}
		}()

		log.Printf("Orgao sync worker started (interval: %v, stale threshold: %v)",
			cfg.OrgaoSync.SyncInterval, cfg.OrgaoSync.StaleThreshold)
	} else {
		log.Println("Orgao sync worker disabled")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Erro ao abrir porta %s: %v", addr, err)
	}
	log.Printf("Servidor iniciado em %s", ln.Addr())

	// Align shutdown drain window with the per-request timeout so in-flight
	// handlers have a chance to finish before the process exits.
	shutdownTimeout := time.Duration(cfg.Server.RequestTimeout) * time.Second
	if shutdownTimeout <= 0 {
		shutdownTimeout = 30 * time.Second
	}

	srv := &http.Server{Handler: r}
	if err := gracefulServe(ctx, srv, ln, shutdownTimeout); err != nil {
		log.Fatalf("Erro no servidor HTTP: %v", err)
	}

	// Flush any pending spans only after all in-flight HTTP requests have
	// finished. Running this as a defer would race with the drain window.
	observability.ShutdownTracer()
}

// gracefulServe serves on the pre-bound ln and blocks until ctx is cancelled,
// then drains in-flight requests via Shutdown bounded by shutdownTimeout.
// New connections are rejected as soon as Shutdown begins.
// Returns only after all connection goroutines have fully exited.
func gracefulServe(ctx context.Context, srv *http.Server, ln net.Listener, shutdownTimeout time.Duration) error {
	var wg sync.WaitGroup
	prev := srv.ConnState
	srv.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			wg.Add(1)
		case http.StateClosed, http.StateHijacked:
			wg.Done()
		}
		if prev != nil {
			prev(conn, state)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("http server failed: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if err := <-errCh; err != nil {
		return err
	}

	wg.Wait()
	log.Println("Server shutdown complete")
	return nil
}

// openDB opens a GORM database connection using the application configuration.
// setupTimezone fixa time.Local no fuso informado. Em caso de fuso inválido,
// registra um aviso e mantém o padrão do processo, sem derrubar a aplicação.
func setupTimezone(name string) {
	if name == "" {
		return
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Printf("Fuso horário inválido %q, mantendo %s: %v", name, time.Local.String(), err)
		return
	}
	time.Local = loc
	log.Printf("Fuso horário do processo: %s", name)
}

func openDB(cfg *config.AppConfig) (*gorm.DB, error) {
	var logMode logger.LogLevel
	if cfg.App.IsDevelopment() {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.ConnMaxIdleTime) * time.Minute)

	return db, nil
}

// runMigrations runs GORM auto-migrations using a temporary database connection.
func runMigrations(cfg *config.AppConfig) {
	log.Println("Iniciando auto-migração...")

	db, err := openDB(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados para migração: %v", err)
	}

	err = db.AutoMigrate(
		&models.Curso{},
		&models.Emprego{},
		&models.Acessibilidade{},
		&models.InstituicaoEnsino{},
		&models.Empresa{},
		&models.Escolaridade{},
		&models.CursoCategoria{},
		&models.CursoAcessibilidade{},
	)
	if err != nil {
		log.Fatalf("Erro ao executar auto-migração básica: %v", err)
	}

	migrateTable := func(model interface{}, name string) {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Aviso: Erro ao migrar tabela %s: %v", name, err)
		}
	}

	log.Println("Migrando tabelas de inscrições...")
	migrateTable(&models.Inscricao{}, "inscricoes")
	migrateTable(&models.CustomField{}, "custom_fields")
	migrateTable(&models.LocationClass{}, "location_classes")
	migrateTable(&models.RemoteClass{}, "remote_classes")

	log.Println("Migrando tabela de jobs...")
	migrateTable(&models.Job{}, "jobs")

	log.Println("Migrando tabela de orgao snapshots...")
	migrateTable(&models.OrgaoSnapshot{}, "orgao_snapshots")

	log.Println("Auto-migração concluída!")

	// Close the migration-only DB connection
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
