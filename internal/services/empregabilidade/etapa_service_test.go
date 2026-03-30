package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockEtapaRepository struct{}

func (m *mockEtapaRepository) Create(ctx context.Context, entity *empregabilidade.Etapa) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockEtapaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error) {
	return nil, nil
}

func (m *mockEtapaRepository) Update(ctx context.Context, entity *empregabilidade.Etapa) error {
	return nil
}

func (m *mockEtapaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockEtapaRepository) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error) {
	return nil, nil
}

func (m *mockEtapaRepository) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	return nil
}

func TestNewEtapaService(t *testing.T) {
	mockRepo := &mockEtapaRepository{}
	service := services.NewEtapaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
