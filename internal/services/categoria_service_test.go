package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockCategoriaRepo struct {
	createID  int
	createErr error
	entity    *models.Categoria
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Categoria
	listTotal int
	listErr   error
}

func (m *mockCategoriaRepo) Create(ctx context.Context, c *models.Categoria) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID != 0 {
		return m.createID, nil
	}
	return 1, nil
}

func (m *mockCategoriaRepo) GetByID(ctx context.Context, id int) (*models.Categoria, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockCategoriaRepo) Update(ctx context.Context, c *models.Categoria) error {
	return m.updateErr
}

func (m *mockCategoriaRepo) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockCategoriaRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Categoria, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Categoria{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestCategoriaService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockCategoriaRepo{createID: 10}
		svc := services.NewCategoriaService(repo)
		ctx := context.Background()
		c := &models.Categoria{Nome: "TI"}
		id, err := svc.Create(ctx, c)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id != 10 {
			t.Errorf("Create: expected id 10, got %d", id)
		}
	})
	t.Run("repo error", func(t *testing.T) {
		repo := &mockCategoriaRepo{createErr: errors.New("db error")}
		svc := services.NewCategoriaService(repo)
		ctx := context.Background()
		_, err := svc.Create(ctx, &models.Categoria{Nome: "TI"})
		if err == nil {
			t.Error("Create: expected error")
		}
	})
}

func TestCategoriaService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockCategoriaRepo{entity: &models.Categoria{ID: 1, Nome: "TI"}}
		svc := services.NewCategoriaService(repo)
		ctx := context.Background()
		c, err := svc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if c == nil || c.ID != 1 || c.Nome != "TI" {
			t.Errorf("GetByID: unexpected %+v", c)
		}
	})
	t.Run("not found", func(t *testing.T) {
		repo := &mockCategoriaRepo{entity: nil}
		svc := services.NewCategoriaService(repo)
		ctx := context.Background()
		c, err := svc.GetByID(ctx, 999)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if c != nil {
			t.Error("GetByID: expected nil")
		}
	})
}

func TestCategoriaService_Update(t *testing.T) {
	repo := &mockCategoriaRepo{}
	svc := services.NewCategoriaService(repo)
	ctx := context.Background()
	err := svc.Update(ctx, &models.Categoria{ID: 1, Nome: "TI"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestCategoriaService_Delete(t *testing.T) {
	repo := &mockCategoriaRepo{}
	svc := services.NewCategoriaService(repo)
	ctx := context.Background()
	err := svc.Delete(ctx, 1)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestCategoriaService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockCategoriaRepo{
			listItems: []*models.Categoria{{ID: 1, Nome: "A"}},
			listTotal: 1,
		}
		svc := services.NewCategoriaService(repo)
		ctx := context.Background()
		items, total, err := svc.List(ctx, map[string]interface{}{}, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(items) != 1 || total != 1 {
			t.Errorf("List: expected 1 item total 1, got %d %d", len(items), total)
		}
	})
	t.Run("with filter", func(t *testing.T) {
		repo := &mockCategoriaRepo{listTotal: 0}
		svc := services.NewCategoriaService(repo)
		ctx := context.Background()
		_, total, err := svc.List(ctx, map[string]interface{}{"days_tolerance": 30}, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 0 {
			t.Errorf("List: expected total 0, got %d", total)
		}
	})
}
