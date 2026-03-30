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

type mockDisponibilidadeRepository struct {
	items []*empregabilidade.Disponibilidade
	total int
}

func (m *mockDisponibilidadeRepository) Create(ctx context.Context, entity *empregabilidade.Disponibilidade) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockDisponibilidadeRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Disponibilidade, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockDisponibilidadeRepository) Update(ctx context.Context, entity *empregabilidade.Disponibilidade) error {
	return nil
}

func (m *mockDisponibilidadeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockDisponibilidadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Disponibilidade, int, error) {
	return m.items, m.total, nil
}

func TestNewDisponibilidadeService(t *testing.T) {
	mockRepo := &mockDisponibilidadeRepository{}
	service := services.NewDisponibilidadeServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestDisponibilidadeService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.Disponibilidade{
		{ID: id1, Descricao: "Manhã", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Tarde", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockDisponibilidadeRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewDisponibilidadeServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Manhã", items[0].Descricao)
	assert.Equal(t, "Tarde", items[1].Descricao)
}

func TestDisponibilidadeService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.Disponibilidade{
		{ID: id, Descricao: "Noite", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockDisponibilidadeRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewDisponibilidadeServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Noite", item.Descricao)
}
