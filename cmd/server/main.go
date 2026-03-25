package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Initialize OpenTelemetry tracing
	if err := observability.InitTracer(cfg); err != nil {
		log.Fatalf("Erro ao inicializar tracer: %v", err)
	}
	defer observability.ShutdownTracer()

	// Run database migrations if enabled.
	// Wire owns the primary DB connection; we open a short-lived connection
	// here solely for migrations, then close it.
	if cfg.Migrations.Run {
		runMigrations(cfg)
	}

	// Set up router (Wire initializes all dependencies internally)
	r, err := router.SetupRouter(cfg)
	if err != nil {
		log.Fatalf("Erro ao inicializar router: %v", err)
	}

	// Initialize orgao sync worker if enabled (uses its own DB/Redis connections)
	var workerCtx context.Context
	var workerCancel context.CancelFunc

	if cfg.OrgaoSync.Enabled {
		log.Println("Initializing orgao sync worker...")

		rmiClient := clients.NewRMIClient(cfg.RMI.BaseURL, 30*time.Second)
		workerRedis := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

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

		workerCtx, workerCancel = context.WithCancel(context.Background())
		go func() {
			if err := worker.Start(workerCtx); err != nil && err != context.Canceled {
				log.Printf("Orgao sync worker error: %v", err)
			}
		}()

		log.Printf("Orgao sync worker started (interval: %v, stale threshold: %v)",
			cfg.OrgaoSync.SyncInterval, cfg.OrgaoSync.StaleThreshold)
	} else {
		log.Println("Orgao sync worker disabled")
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Servidor iniciado em %s", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	if workerCancel != nil {
		log.Println("Stopping orgao sync worker...")
		workerCancel()
	}

	log.Println("Server shutdown complete")
}

// openDB opens a GORM database connection using the application configuration.
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

	log.Println("Auto-migração concluída!")

	// Close the migration-only DB connection
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
