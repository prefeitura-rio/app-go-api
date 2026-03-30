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

type mockModeloTrabalhoRepository struct {
	items []*empregabilidade.ModeloTrabalho
	total int
}

func (m *mockModeloTrabalhoRepository) Create(ctx context.Context, entity *empregabilidade.ModeloTrabalho) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockModeloTrabalhoRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.ModeloTrabalho, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockModeloTrabalhoRepository) Update(ctx context.Context, entity *empregabilidade.ModeloTrabalho) error {
	return nil
}

func (m *mockModeloTrabalhoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockModeloTrabalhoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.ModeloTrabalho, int, error) {
	return m.items, m.total, nil
}

func TestNewModeloTrabalhoService(t *testing.T) {
	mockRepo := &mockModeloTrabalhoRepository{}
	service := services.NewModeloTrabalhoServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestModeloTrabalhoService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.ModeloTrabalho{
		{ID: id1, Descricao: "Presencial", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Remoto", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockModeloTrabalhoRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewModeloTrabalhoServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Presencial", items[0].Descricao)
	assert.Equal(t, "Remoto", items[1].Descricao)
}

func TestModeloTrabalhoService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.ModeloTrabalho{
		{ID: id, Descricao: "Híbrido", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockModeloTrabalhoRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewModeloTrabalhoServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Híbrido", item.Descricao)
}
