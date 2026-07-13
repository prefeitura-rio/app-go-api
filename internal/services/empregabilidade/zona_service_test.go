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

type mockZonaRepository struct {
	items []*empregabilidade.Zona
	total int
}

func (m *mockZonaRepository) Create(ctx context.Context, entity *empregabilidade.Zona) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockZonaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Zona, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockZonaRepository) Update(ctx context.Context, entity *empregabilidade.Zona) error {
	return nil
}

func (m *mockZonaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockZonaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Zona, int, error) {
	return m.items, m.total, nil
}

func TestNewZonaService(t *testing.T) {
	mockRepo := &mockZonaRepository{}
	service := services.NewZonaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestZonaService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.Zona{
		{ID: id1, Descricao: "Centro", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Zona Sul", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockZonaRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewZonaServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Centro", items[0].Descricao)
	assert.Equal(t, "Zona Sul", items[1].Descricao)
}

func TestZonaService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.Zona{
		{ID: id, Descricao: "Zona Oeste", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockZonaRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewZonaServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Zona Oeste", item.Descricao)
}

func TestZonaService_Create(t *testing.T) {
	mockRepo := &mockZonaRepository{}
	service := services.NewZonaServiceWithInterface(mockRepo)

	id, err := service.Create(context.Background(), &empregabilidade.Zona{Descricao: "Zona Norte"})
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestZonaService_Update(t *testing.T) {
	mockRepo := &mockZonaRepository{}
	service := services.NewZonaServiceWithInterface(mockRepo)

	err := service.Update(context.Background(), &empregabilidade.Zona{ID: uuid.New(), Descricao: "Centro"})
	assert.NoError(t, err)
}

func TestZonaService_Delete(t *testing.T) {
	mockRepo := &mockZonaRepository{}
	service := services.NewZonaServiceWithInterface(mockRepo)

	err := service.Delete(context.Background(), uuid.New())
	assert.NoError(t, err)
}
