package providers

import (
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/prefeitura-rio/app-go-api/internal/config"
)

// ProvideDatabase creates GORM database connection with connection pooling and optional tracing
func ProvideDatabase(cfg *config.AppConfig) (*gorm.DB, error) {
	dsn := cfg.Database.DSN()

	// Configure GORM logger based on environment
	var logMode logger.LogLevel
	if cfg.App.IsDevelopment() {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	}

	log.Println("Connecting to database...")
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connection established successfully!")

	// Configure connection pool for optimal performance under high load
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database/sql DB: %w", err)
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
			return nil, fmt.Errorf("failed to add OTEL plugin to GORM: %w", err)
		}
		log.Println("GORM OpenTelemetry instrumentation enabled")
	}

	return db, nil
}

// ProvideRedisClient creates Redis client with connection pooling
func ProvideRedisClient(cfg *config.AppConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})
}
