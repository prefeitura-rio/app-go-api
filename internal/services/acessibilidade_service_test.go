package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockAcessibilidadeRepo struct {
	createID  int
	createErr error
	entity    *models.Acessibilidade
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Acessibilidade
	listTotal int
	listErr   error
}

func (m *mockAcessibilidadeRepo) Create(ctx context.Context, a *models.Acessibilidade) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockAcessibilidadeRepo) GetByID(ctx context.Context, id int) (*models.Acessibilidade, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockAcessibilidadeRepo) Update(ctx context.Context, a *models.Acessibilidade) error {
	return m.updateErr
}

func (m *mockAcessibilidadeRepo) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockAcessibilidadeRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Acessibilidade, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Acessibilidade{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestAcessibilidadeService_Create(t *testing.T) {
	repo := &mockAcessibilidadeRepo{createID: 5}
	svc := services.NewAcessibilidadeService(repo)
	ctx := context.Background()
	id, err := svc.Create(ctx, &models.Acessibilidade{Nome: "Rampa"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 5 {
		t.Errorf("Create: expected id 5, got %d", id)
	}
}

func TestAcessibilidadeService_GetByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := &mockAcessibilidadeRepo{entity: &models.Acessibilidade{ID: 1, Nome: "Rampa"}}
		svc := services.NewAcessibilidadeService(repo)
		ctx := context.Background()
		a, err := svc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if a == nil || a.Nome != "Rampa" {
			t.Errorf("GetByID: unexpected %+v", a)
		}
	})
	t.Run("not found", func(t *testing.T) {
		repo := &mockAcessibilidadeRepo{entity: nil}
		svc := services.NewAcessibilidadeService(repo)
		ctx := context.Background()
		a, err := svc.GetByID(ctx, 999)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if a != nil {
			t.Error("GetByID: expected nil")
		}
	})
}

func TestAcessibilidadeService_List_RepoError(t *testing.T) {
	repo := &mockAcessibilidadeRepo{listErr: errors.New("db error")}
	svc := services.NewAcessibilidadeService(repo)
	ctx := context.Background()
	_, _, err := svc.List(ctx, nil, 1, 10)
	if err == nil {
		t.Error("List: expected error")
	}
}
