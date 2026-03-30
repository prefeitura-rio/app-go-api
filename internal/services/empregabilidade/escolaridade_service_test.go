package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockEscolaridadeRepository struct{}

func (m *mockEscolaridadeRepository) Create(ctx context.Context, entity *empregabilidade.Escolaridade) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockEscolaridadeRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Escolaridade, error) {
	return nil, nil
}

func (m *mockEscolaridadeRepository) Update(ctx context.Context, entity *empregabilidade.Escolaridade) error {
	return nil
}

func (m *mockEscolaridadeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockEscolaridadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Escolaridade, int, error) {
	return nil, 0, nil
}

func TestNewEscolaridadeService(t *testing.T) {
	mockRepo := &mockEscolaridadeRepository{}
	service := services.NewEscolaridadeServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
