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

type mockEtapaRepository struct {
	items   []*empregabilidade.Etapa
	updated []*empregabilidade.Etapa
	deleted []uuid.UUID
}

func (m *mockEtapaRepository) Create(ctx context.Context, entity *empregabilidade.Etapa) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockEtapaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *mockEtapaRepository) Update(ctx context.Context, entity *empregabilidade.Etapa) error {
	m.updated = append(m.updated, entity)
	return nil
}

func (m *mockEtapaRepository) Delete(ctx context.Context, vagaID, id uuid.UUID) error {
	m.deleted = append(m.deleted, id)
	return nil
}

func (m *mockEtapaRepository) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error) {
	return m.items, nil
}

func (m *mockEtapaRepository) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	return nil
}

func TestNewEtapaService(t *testing.T) {
	mockRepo := &mockEtapaRepository{}
	service := services.NewEtapaServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestEtapaService_GetByID(t *testing.T) {
	id := uuid.New()
	vagaID := uuid.New()
	mockItems := []*empregabilidade.Etapa{
		{ID: id, IDVaga: vagaID, Ordem: 1, Titulo: "Análise de Currículo", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockEtapaRepository{
		items: mockItems,
	}
	service := services.NewEtapaServiceWithInterface(mockRepo)

	item, err := service.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, id, item.ID)
	assert.Equal(t, "Análise de Currículo", item.Titulo)
}

func TestEtapaService_ListByVaga(t *testing.T) {
	vagaID := uuid.New()
	mockItems := []*empregabilidade.Etapa{
		{ID: uuid.New(), IDVaga: vagaID, Ordem: 1, Titulo: "Entrevista", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), IDVaga: vagaID, Ordem: 2, Titulo: "Teste Técnico", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockRepo := &mockEtapaRepository{
		items: mockItems,
	}
	service := services.NewEtapaServiceWithInterface(mockRepo)

	items, err := service.ListByVaga(context.Background(), vagaID)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Entrevista", items[0].Titulo)
	assert.Equal(t, "Teste Técnico", items[1].Titulo)
}

// Update e Delete só agem sobre etapa da vaga informada (DEV-139).
func TestEtapaService_Update_EscopoDaVaga(t *testing.T) {
	vagaA := uuid.New()
	vagaB := uuid.New()
	etapaA := uuid.New()
	etapaB := uuid.New()

	tests := []struct {
		name      string
		entity    *empregabilidade.Etapa
		wantErr   error
		wantSaved bool
	}{
		{"etapa da vaga do path", &empregabilidade.Etapa{ID: etapaA, IDVaga: vagaA, Titulo: "Nova"}, nil, true},
		{"etapa de outra vaga", &empregabilidade.Etapa{ID: etapaB, IDVaga: vagaA, Titulo: "Movida"}, services.ErrEtapaNaoEncontrada, false},
		{"etapa inexistente", &empregabilidade.Etapa{ID: uuid.New(), IDVaga: vagaA, Titulo: "Nova"}, services.ErrEtapaNaoEncontrada, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockEtapaRepository{items: []*empregabilidade.Etapa{
				{ID: etapaA, IDVaga: vagaA, Titulo: "Triagem"},
				{ID: etapaB, IDVaga: vagaB, Titulo: "Entrevista"},
			}}
			service := services.NewEtapaServiceWithInterface(mockRepo)

			err := service.Update(context.Background(), tt.entity)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantSaved, len(mockRepo.updated) == 1)
		})
	}
}

func TestEtapaService_Delete_EscopoDaVaga(t *testing.T) {
	vagaA := uuid.New()
	vagaB := uuid.New()
	etapaA := uuid.New()
	etapaB := uuid.New()

	tests := []struct {
		name        string
		etapaID     uuid.UUID
		wantErr     error
		wantDeleted bool
	}{
		{"etapa da vaga do path", etapaA, nil, true},
		{"etapa de outra vaga", etapaB, services.ErrEtapaNaoEncontrada, false},
		{"etapa inexistente", uuid.New(), services.ErrEtapaNaoEncontrada, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockEtapaRepository{items: []*empregabilidade.Etapa{
				{ID: etapaA, IDVaga: vagaA, Titulo: "Triagem"},
				{ID: etapaB, IDVaga: vagaB, Titulo: "Entrevista"},
			}}
			service := services.NewEtapaServiceWithInterface(mockRepo)

			err := service.Delete(context.Background(), vagaA, tt.etapaID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantDeleted, len(mockRepo.deleted) == 1)
		})
	}
}
