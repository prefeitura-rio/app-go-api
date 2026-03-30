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

type mockIdiomaRepository struct {
	items []*empregabilidade.Idioma
	total int
}

func (m *mockIdiomaRepository) Create(ctx context.Context, entity *empregabilidade.Idioma) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockIdiomaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Idioma, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockIdiomaRepository) Update(ctx context.Context, entity *empregabilidade.Idioma) error {
	return nil
}

func (m *mockIdiomaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockIdiomaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Idioma, int, error) {
	return m.items, m.total, nil
}

func TestNewIdiomaService(t *testing.T) {
	mockRepo := &mockIdiomaRepository{}
	service := services.NewIdiomaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestIdiomaService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.Idioma{
		{ID: id1, Descricao: "Português", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Inglês", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockIdiomaRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewIdiomaServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Português", items[0].Descricao)
	assert.Equal(t, "Inglês", items[1].Descricao)
}

func TestIdiomaService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.Idioma{
		{ID: id, Descricao: "Espanhol", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockIdiomaRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewIdiomaServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Espanhol", item.Descricao)
}
