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

type mockNivelIdiomaRepository struct {
	items []*empregabilidade.NivelIdioma
	total int
}

func (m *mockNivelIdiomaRepository) Create(ctx context.Context, entity *empregabilidade.NivelIdioma) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockNivelIdiomaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.NivelIdioma, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockNivelIdiomaRepository) Update(ctx context.Context, entity *empregabilidade.NivelIdioma) error {
	return nil
}

func (m *mockNivelIdiomaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockNivelIdiomaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.NivelIdioma, int, error) {
	return m.items, m.total, nil
}

func TestNewNivelIdiomaService(t *testing.T) {
	mockRepo := &mockNivelIdiomaRepository{}
	service := services.NewNivelIdiomaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestNivelIdiomaService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.NivelIdioma{
		{ID: id1, Descricao: "Básico", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Intermediário", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockNivelIdiomaRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewNivelIdiomaServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Básico", items[0].Descricao)
	assert.Equal(t, "Intermediário", items[1].Descricao)
}

func TestNivelIdiomaService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.NivelIdioma{
		{ID: id, Descricao: "Avançado", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockNivelIdiomaRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewNivelIdiomaServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Avançado", item.Descricao)
}
