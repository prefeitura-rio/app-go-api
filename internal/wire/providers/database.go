package providers

import (
	"fmt"
	"log"
	"os"
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

	gormConfig := &gorm.Config{
		Logger: GormLogger(cfg),
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

// GormLogger returns the GORM logger shared by every database connection: SQL at
// Info level in development (APP_ENV=development, also used by staging) and silent
// otherwise. Queries are always logged with their placeholders ($1, $2…) instead of
// the bound values, so search terms such as CPF, name or e-mail never reach the logs.
func GormLogger(cfg *config.AppConfig) logger.Interface {
	return newGormLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogLevel(cfg))
}

// gormLogLevel maps the environment to the GORM log level.
func gormLogLevel(cfg *config.AppConfig) logger.LogLevel {
	if cfg.App.IsDevelopment() {
		return logger.Info
	}
	return logger.Silent
}

// newGormLogger builds a GORM logger with the defaults of logger.Default, but
// without interpolating query parameters into the logged SQL.
func newGormLogger(writer logger.Writer, level logger.LogLevel) logger.Interface {
	return logger.New(writer, logger.Config{
		SlowThreshold:        200 * time.Millisecond,
		LogLevel:             level,
		Colorful:             true,
		ParameterizedQueries: true,
	})
}

// ProvideRedisClient creates Redis client with connection pooling
func ProvideRedisClient(cfg *config.AppConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		ConnMaxLifetime: 4 * time.Minute,
		ConnMaxIdleTime: 3 * time.Minute,
	})
}
