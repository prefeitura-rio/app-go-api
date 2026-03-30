package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockIdiomaRepository struct{}

func (m *mockIdiomaRepository) Create(ctx context.Context, entity *empregabilidade.Idioma) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockIdiomaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Idioma, error) {
	return nil, nil
}

func (m *mockIdiomaRepository) Update(ctx context.Context, entity *empregabilidade.Idioma) error {
	return nil
}

func (m *mockIdiomaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockIdiomaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Idioma, int, error) {
	return nil, 0, nil
}

func TestNewIdiomaService(t *testing.T) {
	mockRepo := &mockIdiomaRepository{}
	service := services.NewIdiomaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
