package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockModeloTrabalhoRepository struct{}

func (m *mockModeloTrabalhoRepository) Create(ctx context.Context, entity *empregabilidade.ModeloTrabalho) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockModeloTrabalhoRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.ModeloTrabalho, error) {
	return nil, nil
}

func (m *mockModeloTrabalhoRepository) Update(ctx context.Context, entity *empregabilidade.ModeloTrabalho) error {
	return nil
}

func (m *mockModeloTrabalhoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockModeloTrabalhoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.ModeloTrabalho, int, error) {
	return nil, 0, nil
}

func TestNewModeloTrabalhoService(t *testing.T) {
	mockRepo := &mockModeloTrabalhoRepository{}
	service := services.NewModeloTrabalhoServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
