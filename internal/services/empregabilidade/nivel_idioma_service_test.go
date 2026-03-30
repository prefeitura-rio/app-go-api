package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockNivelIdiomaRepository struct{}

func (m *mockNivelIdiomaRepository) Create(ctx context.Context, entity *empregabilidade.NivelIdioma) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockNivelIdiomaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.NivelIdioma, error) {
	return nil, nil
}

func (m *mockNivelIdiomaRepository) Update(ctx context.Context, entity *empregabilidade.NivelIdioma) error {
	return nil
}

func (m *mockNivelIdiomaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockNivelIdiomaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.NivelIdioma, int, error) {
	return nil, 0, nil
}

func TestNewNivelIdiomaService(t *testing.T) {
	mockRepo := &mockNivelIdiomaRepository{}
	service := services.NewNivelIdiomaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
