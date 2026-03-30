package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockTipoPCDRepository struct{}

func (m *mockTipoPCDRepository) Create(ctx context.Context, entity *empregabilidade.TipoPCD) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockTipoPCDRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoPCD, error) {
	return nil, nil
}

func (m *mockTipoPCDRepository) Update(ctx context.Context, entity *empregabilidade.TipoPCD) error {
	return nil
}

func (m *mockTipoPCDRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockTipoPCDRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoPCD, int, error) {
	return nil, 0, nil
}

func TestNewTipoPCDService(t *testing.T) {
	mockRepo := &mockTipoPCDRepository{}
	service := services.NewTipoPCDServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
