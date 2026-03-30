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

type mockEscolaridadeRepository struct {
	items []*empregabilidade.Escolaridade
	total int
}

func (m *mockEscolaridadeRepository) Create(ctx context.Context, entity *empregabilidade.Escolaridade) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockEscolaridadeRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Escolaridade, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockEscolaridadeRepository) Update(ctx context.Context, entity *empregabilidade.Escolaridade) error {
	return nil
}

func (m *mockEscolaridadeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockEscolaridadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Escolaridade, int, error) {
	return m.items, m.total, nil
}

func TestNewEscolaridadeService(t *testing.T) {
	mockRepo := &mockEscolaridadeRepository{}
	service := services.NewEscolaridadeServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestEscolaridadeService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.Escolaridade{
		{ID: id1, Descricao: "Ensino Fundamental", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Ensino Médio", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockEscolaridadeRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewEscolaridadeServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Ensino Fundamental", items[0].Descricao)
	assert.Equal(t, "Ensino Médio", items[1].Descricao)
}

func TestEscolaridadeService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.Escolaridade{
		{ID: id, Descricao: "Ensino Superior", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockEscolaridadeRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewEscolaridadeServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Ensino Superior", item.Descricao)
}
