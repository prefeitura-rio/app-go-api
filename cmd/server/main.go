package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/observability"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/router"
	"github.com/prefeitura-rio/app-go-api/internal/workers"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
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
	// Carrega configurações
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

	// Conecta ao banco de dados usando GORM
	dsn := cfg.Database.DSN()

	// Configura o logger do GORM de acordo com o ambiente
	var logMode logger.LogLevel
	if cfg.App.IsDevelopment() {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	}

	log.Println("Tentando conectar ao banco de dados...")
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso!")

	// Configure connection pool for optimal performance under high load
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Erro ao obter database/sql DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.ConnMaxIdleTime) * time.Minute)

	log.Printf("Database connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%dm, MaxIdleTime=%dm",
		cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime, cfg.Database.ConnMaxIdleTime)

	// Add OpenTelemetry instrumentation to GORM (if tracing is enabled)
	if cfg.Tracing.Enabled {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			log.Fatalf("Erro ao adicionar plugin OTEL ao GORM: %v", err)
		}
		log.Println("GORM OpenTelemetry instrumentation enabled")
	}

	// Auto-migração do GORM (opcional)
	if cfg.Migrations.Run {
		log.Println("Iniciando auto-migração...")
		// Migrar apenas as tabelas básicas que já existem
		// Nota: Categoria foi removida do AutoMigrate pois é gerenciada via migrations SQL
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

		// Tentar migrar as novas tabelas individualmente para melhor controle de erros
		log.Println("Migrando tabelas de inscrições...")
		err = db.AutoMigrate(&models.Inscricao{})
		if err != nil {
			log.Printf("Aviso: Erro ao migrar tabela inscricoes: %v", err)
		}

		err = db.AutoMigrate(&models.CustomField{})
		if err != nil {
			log.Printf("Aviso: Erro ao migrar tabela custom_fields: %v", err)
		}

		err = db.AutoMigrate(&models.LocationClass{})
		if err != nil {
			log.Printf("Aviso: Erro ao migrar tabela location_classes: %v", err)
		}

		err = db.AutoMigrate(&models.RemoteClass{})
		if err != nil {
			log.Printf("Aviso: Erro ao migrar tabela remote_classes: %v", err)
		}

		log.Println("Migrando tabela de jobs...")
		err = db.AutoMigrate(&models.Job{})
		if err != nil {
			log.Printf("Aviso: Erro ao migrar tabela jobs: %v", err)
		}

		log.Println("Auto-migração concluída!")
	}

	// Configura o router
	r := router.SetupRouter(db, cfg)

	// Initialize orgao sync worker if enabled
	var workerCtx context.Context
	var workerCancel context.CancelFunc

	if cfg.OrgaoSync.Enabled {
		log.Println("Initializing orgao sync worker...")

		// Create RMI client for worker
		rmiClient := clients.NewRMIClient(cfg.RMI.BaseURL, 30*time.Second)

		// Create Redis client for distributed locking (ensures only one pod runs sync at a time)
		workerRedis := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		// Create repositories
		orgaoSnapshotRepo := repository.NewOrgaoSnapshotRepository(db)
		cursoRepo := repository.NewCursoRepository(db)
		empregoRepo := repository.NewEmpregoRepository(db)
		oportunidadeMEIRepo := repository.NewOportunidadeMEIRepository(db)

		// Create worker
		worker := workers.NewOrgaoSyncWorker(
			db,
			rmiClient,
			workerRedis,
			orgaoSnapshotRepo,
			cursoRepo,
			empregoRepo,
			oportunidadeMEIRepo,
			&cfg.OrgaoSync,
		)

		// Start worker in background
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

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Servidor iniciado em %s", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Cancel worker context if worker is running
	if workerCancel != nil {
		log.Println("Stopping orgao sync worker...")
		workerCancel()
	}

	log.Println("Server shutdown complete")
}
