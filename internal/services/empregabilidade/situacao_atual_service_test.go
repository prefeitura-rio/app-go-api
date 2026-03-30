package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockSituacaoAtualRepository struct{}

func (m *mockSituacaoAtualRepository) Create(ctx context.Context, entity *empregabilidade.SituacaoAtual) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockSituacaoAtualRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.SituacaoAtual, error) {
	return nil, nil
}

func (m *mockSituacaoAtualRepository) Update(ctx context.Context, entity *empregabilidade.SituacaoAtual) error {
	return nil
}

func (m *mockSituacaoAtualRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockSituacaoAtualRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.SituacaoAtual, int, error) {
	return nil, 0, nil
}

func TestNewSituacaoAtualService(t *testing.T) {
	mockRepo := &mockSituacaoAtualRepository{}
	service := services.NewSituacaoAtualServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
