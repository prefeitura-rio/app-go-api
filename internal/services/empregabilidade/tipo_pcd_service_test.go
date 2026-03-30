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

type mockTipoPCDRepository struct {
	items []*empregabilidade.TipoPCD
	total int
}

func (m *mockTipoPCDRepository) Create(ctx context.Context, entity *empregabilidade.TipoPCD) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockTipoPCDRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoPCD, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockTipoPCDRepository) Update(ctx context.Context, entity *empregabilidade.TipoPCD) error {
	return nil
}

func (m *mockTipoPCDRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockTipoPCDRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoPCD, int, error) {
	return m.items, m.total, nil
}

func TestNewTipoPCDService(t *testing.T) {
	mockRepo := &mockTipoPCDRepository{}
	service := services.NewTipoPCDServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestTipoPCDService_List(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	mockItems := []*empregabilidade.TipoPCD{
		{ID: id1, Descricao: "Física", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: id2, Descricao: "Visual", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockTipoPCDRepository{
		items: mockItems,
		total: 2,
	}
	service := services.NewTipoPCDServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Física", items[0].Descricao)
	assert.Equal(t, "Visual", items[1].Descricao)
}

func TestTipoPCDService_GetByID(t *testing.T) {
	id := uuid.New()
	mockItems := []*empregabilidade.TipoPCD{
		{ID: id, Descricao: "Auditiva", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockTipoPCDRepository{
		items: mockItems,
		total: 1,
	}
	service := services.NewTipoPCDServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Auditiva", item.Descricao)
}
