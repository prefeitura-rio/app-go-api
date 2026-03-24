package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// Mock InformacaoComplementar Repository
type MockInformacaoComplementarRepo struct {
	informacoes map[uuid.UUID]*empregabilidade.InformacaoComplementar
	byVaga      map[uuid.UUID][]*empregabilidade.InformacaoComplementar
	createError error
	getError    error
	updateError error
	deleteError error
	listError   error
}

func NewMockInformacaoComplementarRepo() *MockInformacaoComplementarRepo {
	return &MockInformacaoComplementarRepo{
		informacoes: make(map[uuid.UUID]*empregabilidade.InformacaoComplementar),
		byVaga:      make(map[uuid.UUID][]*empregabilidade.InformacaoComplementar),
	}
}

func (m *MockInformacaoComplementarRepo) Create(ctx context.Context, entity *empregabilidade.InformacaoComplementar) (uuid.UUID, error) {
	if m.createError != nil {
		return uuid.Nil, m.createError
	}
	id := uuid.New()
	entity.ID = id
	m.informacoes[id] = entity
	m.byVaga[entity.IDVaga] = append(m.byVaga[entity.IDVaga], entity)
	return id, nil
}

func (m *MockInformacaoComplementarRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.InformacaoComplementar, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	info, exists := m.informacoes[id]
	if !exists {
		return nil, nil
	}
	return info, nil
}

func (m *MockInformacaoComplementarRepo) Update(ctx context.Context, entity *empregabilidade.InformacaoComplementar) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.informacoes[entity.ID] = entity
	return nil
}

func (m *MockInformacaoComplementarRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.informacoes, id)
	return nil
}

func (m *MockInformacaoComplementarRepo) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.InformacaoComplementar, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.byVaga[vagaID], nil
}

func (m *MockInformacaoComplementarRepo) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.byVaga, vagaID)
	return nil
}

func (m *MockInformacaoComplementarRepo) CreateWithLimitCheck(ctx context.Context, entity *empregabilidade.InformacaoComplementar, maxLimit int) (uuid.UUID, error) {
	if m.listError != nil {
		return uuid.Nil, m.listError
	}

	existing := m.byVaga[entity.IDVaga]
	if len(existing) >= maxLimit {
		return uuid.Nil, errors.New("limite máximo de 5 informações complementares por vaga atingido")
	}

	if m.createError != nil {
		return uuid.Nil, m.createError
	}

	id := uuid.New()
	entity.ID = id
	m.informacoes[id] = entity
	m.byVaga[entity.IDVaga] = append(m.byVaga[entity.IDVaga], entity)
	return id, nil
}

func TestInformacaoComplementarService_Create_Success(t *testing.T) {
	t.Run("Successful creation with less than 5 items", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		vagaID := uuid.New()
		info := &empregabilidade.InformacaoComplementar{
			IDVaga:      vagaID,
			Titulo:      "Pergunta 1",
			TipoCampo:   "resposta_curta",
			Obrigatorio: false,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, info)

		if err != nil {
			t.Errorf("Expected successful creation, got error: %v", err)
		}

		if id == uuid.Nil {
			t.Error("Expected non-nil UUID")
		}
	})

	t.Run("Successful creation at limit (4 existing)", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		vagaID := uuid.New()

		// Add 4 existing items
		for i := 0; i < 4; i++ {
			existing := &empregabilidade.InformacaoComplementar{
				ID:        uuid.New(),
				IDVaga:    vagaID,
				Titulo:    "Pergunta existente",
				TipoCampo: "resposta_curta",
			}
			mockRepo.byVaga[vagaID] = append(mockRepo.byVaga[vagaID], existing)
		}

		info := &empregabilidade.InformacaoComplementar{
			IDVaga:      vagaID,
			Titulo:      "Pergunta 5 (última permitida)",
			TipoCampo:   "resposta_curta",
			Obrigatorio: false,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, info)

		if err != nil {
			t.Errorf("Expected successful creation at limit, got error: %v", err)
		}

		if id == uuid.Nil {
			t.Error("Expected non-nil UUID")
		}
	})
}

func TestInformacaoComplementarService_Create_ExceedsLimit(t *testing.T) {
	t.Run("Error when exceeding 5 items limit", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		vagaID := uuid.New()

		// Add 5 existing items (at the limit)
		for i := 0; i < 5; i++ {
			existing := &empregabilidade.InformacaoComplementar{
				ID:        uuid.New(),
				IDVaga:    vagaID,
				Titulo:    "Pergunta existente",
				TipoCampo: "resposta_curta",
			}
			mockRepo.byVaga[vagaID] = append(mockRepo.byVaga[vagaID], existing)
		}

		info := &empregabilidade.InformacaoComplementar{
			IDVaga:      vagaID,
			Titulo:      "Pergunta 6 (deve falhar)",
			TipoCampo:   "resposta_curta",
			Obrigatorio: false,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, info)

		if err == nil {
			t.Error("Expected error when exceeding limit of 5 items")
		}

		expectedMsg := "limite máximo de 5 informações complementares por vaga atingido"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})
}

func TestInformacaoComplementarService_Create_ListError(t *testing.T) {
	t.Run("Error when listing existing items fails", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		mockRepo.listError = errors.New("database connection error")
		service := services.NewInformacaoComplementarService(mockRepo)

		vagaID := uuid.New()
		info := &empregabilidade.InformacaoComplementar{
			IDVaga:      vagaID,
			Titulo:      "Pergunta",
			TipoCampo:   "resposta_curta",
			Obrigatorio: false,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, info)

		if err == nil {
			t.Error("Expected error when listing fails")
		}

		expectedMsg := "database connection error"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})
}

func TestInformacaoComplementarService_CRUD(t *testing.T) {
	t.Run("GetByID returns existing item", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		id := uuid.New()
		vagaID := uuid.New()
		mockRepo.informacoes[id] = &empregabilidade.InformacaoComplementar{
			ID:        id,
			IDVaga:    vagaID,
			Titulo:    "Pergunta",
			TipoCampo: "resposta_curta",
		}

		ctx := context.Background()
		result, err := service.GetByID(ctx, id)

		if err != nil {
			t.Errorf("Expected successful get, got error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		if result.Titulo != "Pergunta" {
			t.Errorf("Expected titulo 'Pergunta', got '%s'", result.Titulo)
		}
	})

	t.Run("Update modifies existing item", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		id := uuid.New()
		vagaID := uuid.New()
		mockRepo.informacoes[id] = &empregabilidade.InformacaoComplementar{
			ID:        id,
			IDVaga:    vagaID,
			Titulo:    "Pergunta Original",
			TipoCampo: "resposta_curta",
		}

		updated := &empregabilidade.InformacaoComplementar{
			ID:        id,
			IDVaga:    vagaID,
			Titulo:    "Pergunta Atualizada",
			TipoCampo: "resposta_numerica",
		}

		ctx := context.Background()
		err := service.Update(ctx, updated)

		if err != nil {
			t.Errorf("Expected successful update, got error: %v", err)
		}

		if mockRepo.informacoes[id].Titulo != "Pergunta Atualizada" {
			t.Error("Expected titulo to be updated")
		}
	})

	t.Run("Delete removes existing item", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		id := uuid.New()
		mockRepo.informacoes[id] = &empregabilidade.InformacaoComplementar{
			ID:     id,
			Titulo: "Pergunta",
		}

		ctx := context.Background()
		err := service.Delete(ctx, id)

		if err != nil {
			t.Errorf("Expected successful delete, got error: %v", err)
		}

		if _, exists := mockRepo.informacoes[id]; exists {
			t.Error("Expected item to be deleted")
		}
	})

	t.Run("ListByVaga returns items for vaga", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		vagaID := uuid.New()
		for i := 0; i < 3; i++ {
			mockRepo.byVaga[vagaID] = append(mockRepo.byVaga[vagaID], &empregabilidade.InformacaoComplementar{
				ID:     uuid.New(),
				IDVaga: vagaID,
				Titulo: "Pergunta",
			})
		}

		ctx := context.Background()
		result, err := service.ListByVaga(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful list, got error: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 items, got %d", len(result))
		}
	})

	t.Run("DeleteByVaga removes all items for vaga", func(t *testing.T) {
		mockRepo := NewMockInformacaoComplementarRepo()
		service := services.NewInformacaoComplementarService(mockRepo)

		vagaID := uuid.New()
		for i := 0; i < 3; i++ {
			mockRepo.byVaga[vagaID] = append(mockRepo.byVaga[vagaID], &empregabilidade.InformacaoComplementar{
				ID:     uuid.New(),
				IDVaga: vagaID,
				Titulo: "Pergunta",
			})
		}

		ctx := context.Background()
		err := service.DeleteByVaga(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful delete, got error: %v", err)
		}

		if len(mockRepo.byVaga[vagaID]) != 0 {
			t.Error("Expected all items to be deleted")
		}
	})
}
