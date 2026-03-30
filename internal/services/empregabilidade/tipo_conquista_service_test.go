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

type mockTipoConquistaRepository struct {
	items []*empregabilidade.TipoConquista
	total int
}

func (m *mockTipoConquistaRepository) Create(ctx context.Context, entity *empregabilidade.TipoConquista) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockTipoConquistaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoConquista, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockTipoConquistaRepository) Update(ctx context.Context, entity *empregabilidade.TipoConquista) error {
	return nil
}

func (m *mockTipoConquistaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockTipoConquistaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoConquista, int, error) {
	return m.items, m.total, nil
}

func TestNewTipoConquistaService(t *testing.T) {
	mockRepo := &mockTipoConquistaRepository{}
	service := services.NewTipoConquistaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestTipoConquistaService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.TipoConquista{
		{ID: id1, Descricao: "Certificação", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Prêmio", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockTipoConquistaRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewTipoConquistaServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Certificação", items[0].Descricao)
	assert.Equal(t, "Prêmio", items[1].Descricao)
}

func TestTipoConquistaService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.TipoConquista{
		{ID: id, Descricao: "Reconhecimento", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockTipoConquistaRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewTipoConquistaServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Reconhecimento", item.Descricao)
}
