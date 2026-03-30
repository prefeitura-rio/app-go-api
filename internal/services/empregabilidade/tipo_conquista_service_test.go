package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockTipoConquistaRepository struct{}

func (m *mockTipoConquistaRepository) Create(ctx context.Context, entity *empregabilidade.TipoConquista) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockTipoConquistaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoConquista, error) {
	return nil, nil
}

func (m *mockTipoConquistaRepository) Update(ctx context.Context, entity *empregabilidade.TipoConquista) error {
	return nil
}

func (m *mockTipoConquistaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockTipoConquistaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoConquista, int, error) {
	return nil, 0, nil
}

func TestNewTipoConquistaService(t *testing.T) {
	mockRepo := &mockTipoConquistaRepository{}
	service := services.NewTipoConquistaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
