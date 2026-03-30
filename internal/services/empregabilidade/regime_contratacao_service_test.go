package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockRegimeContratacaoRepository struct{}

func (m *mockRegimeContratacaoRepository) Create(ctx context.Context, entity *empregabilidade.RegimeContratacao) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockRegimeContratacaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.RegimeContratacao, error) {
	return nil, nil
}

func (m *mockRegimeContratacaoRepository) Update(ctx context.Context, entity *empregabilidade.RegimeContratacao) error {
	return nil
}

func (m *mockRegimeContratacaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockRegimeContratacaoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.RegimeContratacao, int, error) {
	return nil, 0, nil
}

func TestNewRegimeContratacaoService(t *testing.T) {
	mockRepo := &mockRegimeContratacaoRepository{}
	service := services.NewRegimeContratacaoServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
