package repository_test

import (
	"context"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

func getDBForIntegration(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("RUN_REPOSITORY_INTEGRATION") == "" && os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: set RUN_REPOSITORY_INTEGRATION=1 or DATABASE_URL to run")
		return nil
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		cfg, err := config.Load()
		if err != nil {
			t.Skipf("Skipping integration test: config load failed: %v", err)
			return nil
		}
		dsn = cfg.Database.DSN()
		if cfg.Database.Host == "" {
			t.Skip("Skipping integration test: DB_HOST not set")
			return nil
		}
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to database: %v", err)
		return nil
	}
	return db
}

func TestCategoriaRepository_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	ctx := context.Background()
	repo := repository.NewCategoriaRepository(db)

	t.Run("Create and GetByID", func(t *testing.T) {
		c := &models.Categoria{Nome: "Test Categoria Integration"}
		id, err := repo.Create(ctx, c)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id <= 0 {
			t.Errorf("Create: expected positive id, got %d", id)
		}
		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got == nil || got.Nome != c.Nome {
			t.Errorf("GetByID: got %+v", got)
		}
		_ = repo.Delete(ctx, id)
	})

	t.Run("List", func(t *testing.T) {
		items, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if items == nil {
			t.Error("List: expected non-nil slice")
		}
		if total < 0 {
			t.Errorf("List: total must be >= 0, got %d", total)
		}
	})
}

func TestAcessibilidadeRepository_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	ctx := context.Background()
	repo := repository.NewAcessibilidadeRepository(db)

	t.Run("Create and GetByID", func(t *testing.T) {
		a := &models.Acessibilidade{Nome: "Test Acessibilidade Integration"}
		id, err := repo.Create(ctx, a)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id <= 0 {
			t.Errorf("Create: expected positive id, got %d", id)
		}
		got, err := repo.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got == nil || got.Nome != a.Nome {
			t.Errorf("GetByID: got %+v", got)
		}
		_ = repo.Delete(ctx, id)
	})
}
