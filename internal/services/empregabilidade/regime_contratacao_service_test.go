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

type mockRegimeContratacaoRepository struct {
	items []*empregabilidade.RegimeContratacao
	total int
}

func (m *mockRegimeContratacaoRepository) Create(ctx context.Context, entity *empregabilidade.RegimeContratacao) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockRegimeContratacaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.RegimeContratacao, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockRegimeContratacaoRepository) Update(ctx context.Context, entity *empregabilidade.RegimeContratacao) error {
	return nil
}

func (m *mockRegimeContratacaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockRegimeContratacaoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.RegimeContratacao, int, error) {
	return m.items, m.total, nil
}

func TestNewRegimeContratacaoService(t *testing.T) {
	mockRepo := &mockRegimeContratacaoRepository{}
	service := services.NewRegimeContratacaoServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestRegimeContratacaoService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.RegimeContratacao{
		{ID: id1, Descricao: "CLT", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "PJ", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockRegimeContratacaoRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewRegimeContratacaoServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "CLT", items[0].Descricao)
	assert.Equal(t, "PJ", items[1].Descricao)
}

func TestRegimeContratacaoService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.RegimeContratacao{
		{ID: id, Descricao: "Estágio", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockRegimeContratacaoRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewRegimeContratacaoServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Estágio", item.Descricao)
}
