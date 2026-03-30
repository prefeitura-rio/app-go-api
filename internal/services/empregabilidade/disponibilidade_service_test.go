package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockDisponibilidadeRepository struct{}

func (m *mockDisponibilidadeRepository) Create(ctx context.Context, entity *empregabilidade.Disponibilidade) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockDisponibilidadeRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Disponibilidade, error) {
	return nil, nil
}

func (m *mockDisponibilidadeRepository) Update(ctx context.Context, entity *empregabilidade.Disponibilidade) error {
	return nil
}

func (m *mockDisponibilidadeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockDisponibilidadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Disponibilidade, int, error) {
	return nil, 0, nil
}

func TestNewDisponibilidadeService(t *testing.T) {
	mockRepo := &mockDisponibilidadeRepository{}
	service := services.NewDisponibilidadeServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
