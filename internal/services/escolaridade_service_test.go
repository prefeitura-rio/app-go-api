package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockEscolaridadeRepo struct {
	createID  int
	createErr error
	entity    *models.Escolaridade
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Escolaridade
	listTotal int
	listErr   error
}

func (m *mockEscolaridadeRepo) Create(ctx context.Context, e *models.Escolaridade) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockEscolaridadeRepo) GetByID(ctx context.Context, id int) (*models.Escolaridade, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockEscolaridadeRepo) Update(ctx context.Context, e *models.Escolaridade) error {
	return m.updateErr
}

func (m *mockEscolaridadeRepo) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockEscolaridadeRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Escolaridade, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Escolaridade{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestEscolaridadeService_Create(t *testing.T) {
	repo := &mockEscolaridadeRepo{createID: 3}
	svc := services.NewEscolaridadeService(repo)
	ctx := context.Background()
	id, err := svc.Create(ctx, &models.Escolaridade{Nivel: "Superior"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 3 {
		t.Errorf("Create: expected id 3, got %d", id)
	}
}

func TestEscolaridadeService_GetByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := &mockEscolaridadeRepo{entity: &models.Escolaridade{ID: 1, Nivel: "Médio"}}
		svc := services.NewEscolaridadeService(repo)
		ctx := context.Background()
		e, err := svc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if e == nil || e.Nivel != "Médio" {
			t.Errorf("GetByID: unexpected %+v", e)
		}
	})
	t.Run("not found", func(t *testing.T) {
		repo := &mockEscolaridadeRepo{entity: nil}
		svc := services.NewEscolaridadeService(repo)
		ctx := context.Background()
		e, err := svc.GetByID(ctx, 999)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if e != nil {
			t.Error("GetByID: expected nil")
		}
	})
}

func TestEscolaridadeService_List_RepoError(t *testing.T) {
	repo := &mockEscolaridadeRepo{listErr: errors.New("db error")}
	svc := services.NewEscolaridadeService(repo)
	ctx := context.Background()
	_, _, err := svc.List(ctx, nil, 1, 10)
	if err == nil {
		t.Error("List: expected error")
	}
}
