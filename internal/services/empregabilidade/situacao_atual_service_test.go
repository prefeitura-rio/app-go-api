package empregabilidade_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockSituacaoAtualRepository struct {
	items []*empregabilidade.SituacaoAtual
	total int
}

func (m *mockSituacaoAtualRepository) Create(ctx context.Context, entity *empregabilidade.SituacaoAtual) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockSituacaoAtualRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.SituacaoAtual, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockSituacaoAtualRepository) Update(ctx context.Context, entity *empregabilidade.SituacaoAtual) error {
	return nil
}

func (m *mockSituacaoAtualRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockSituacaoAtualRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.SituacaoAtual, int, error) {
	return m.items, m.total, nil
}

func TestNewSituacaoAtualService(t *testing.T) {
	mockRepo := &mockSituacaoAtualRepository{}
	service := services.NewSituacaoAtualServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestSituacaoAtualService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.SituacaoAtual{
		{ID: id1, Descricao: "Empregado", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Desempregado", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockSituacaoAtualRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewSituacaoAtualServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Empregado", items[0].Descricao)
	assert.Equal(t, "Desempregado", items[1].Descricao)
}

func TestSituacaoAtualService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.SituacaoAtual{
		{ID: id, Descricao: "Estudante", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockSituacaoAtualRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewSituacaoAtualServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Estudante", item.Descricao)
}
