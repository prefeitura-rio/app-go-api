package empregabilidade_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// Mock Candidatura Repository
type MockCandidaturaRepo struct {
	candidaturas    map[uuid.UUID]*empregabilidade.Candidatura
	existingCheck   bool
	createError     error
	getError        error
	updateError     error
	deleteError     error
	checkError      error
	listError       error
	updateStatusErr error
	bulkGetError    error
}

func NewMockCandidaturaRepo() *MockCandidaturaRepo {
	return &MockCandidaturaRepo{
		candidaturas: make(map[uuid.UUID]*empregabilidade.Candidatura),
	}
}

func (m *MockCandidaturaRepo) Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	if m.createError != nil {
		return uuid.Nil, m.createError
	}
	id := uuid.New()
	entity.ID = id
	m.candidaturas[id] = entity
	return id, nil
}

func (m *MockCandidaturaRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	candidatura, exists := m.candidaturas[id]
	if !exists {
		return nil, nil
	}
	return candidatura, nil
}

func (m *MockCandidaturaRepo) Update(ctx context.Context, entity *empregabilidade.Candidatura) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.candidaturas[entity.ID] = entity
	return nil
}

func (m *MockCandidaturaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.candidaturas, id)
	return nil
}

func (m *MockCandidaturaRepo) List(ctx context.Context, filter empregabilidade.CandidaturaFilter, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	return []*empregabilidade.Candidatura{}, 0, nil
}

func (m *MockCandidaturaRepo) CheckExistingCandidatura(ctx context.Context, cpf string, vagaID uuid.UUID) (bool, error) {
	if m.checkError != nil {
		return false, m.checkError
	}
	return m.existingCheck, nil
}

func (m *MockCandidaturaRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if c, exists := m.candidaturas[id]; exists {
		c.Status = status
	}
	return nil
}

func (m *MockCandidaturaRepo) UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error {
	return nil
}

func (m *MockCandidaturaRepo) BulkUpdateStatus(ctx context.Context, vagaID uuid.UUID, cpfs []string, status empregabilidade.StatusCandidatura) (repository.BulkUpdateResult, error) {
	return repository.BulkUpdateResult{Updated: len(cpfs)}, nil
}

func (m *MockCandidaturaRepo) BulkGetByCPFs(ctx context.Context, vagaID uuid.UUID, cpfs []string) ([]*empregabilidade.Candidatura, error) {
	if m.bulkGetError != nil {
		return nil, m.bulkGetError
	}
	var result []*empregabilidade.Candidatura
	for _, c := range m.candidaturas {
		if c.IDVaga != vagaID {
			continue
		}
		for _, cpf := range cpfs {
			if c.CPF == cpf {
				result = append(result, c)
				break
			}
		}
	}
	return result, nil
}

func (m *MockCandidaturaRepo) BulkUpdateEtapa(ctx context.Context, ids []uuid.UUID, etapaID uuid.UUID) error {
	for _, id := range ids {
		if c, exists := m.candidaturas[id]; exists {
			c.IDEtapaAtual = &etapaID
		}
	}
	return nil
}

func (m *MockCandidaturaRepo) BulkSaveAndUpdateStatusByVagaID(ctx context.Context, vagaID uuid.UUID, status empregabilidade.StatusCandidatura) error {
	return nil
}

func (m *MockCandidaturaRepo) BulkRestoreStatusByVagaID(ctx context.Context, vagaID uuid.UUID) error {
	return nil
}

func (m *MockCandidaturaRepo) CountByStatus(ctx context.Context, filter empregabilidade.CandidaturaFilter) (map[empregabilidade.StatusCandidatura]int64, error) {
	return map[empregabilidade.StatusCandidatura]int64{}, nil
}

// Mock Vaga Repository
type MockVagaRepo struct {
	vagas    map[uuid.UUID]*empregabilidade.Vaga
	getError error
}

func NewMockVagaRepo() *MockVagaRepo {
	return &MockVagaRepo{
		vagas: make(map[uuid.UUID]*empregabilidade.Vaga),
	}
}

func (m *MockVagaRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	vaga, exists := m.vagas[id]
	if !exists {
		return nil, nil
	}
	return vaga, nil
}

// Mock Curriculo Service
type MockCurriculoService struct {
	curriculo *empregabilidade.CurriculoCompleto
	getError  error
}

func NewMockCurriculoService() *MockCurriculoService {
	return &MockCurriculoService{
		curriculo: &empregabilidade.CurriculoCompleto{},
	}
}

func (m *MockCurriculoService) GetCurriculoCompleto(ctx context.Context, cpf string) (*empregabilidade.CurriculoCompleto, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.curriculo, nil
}

func TestCandidaturaService_Create_Success(t *testing.T) {
	t.Run("Successful candidatura creation", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, candidatura)

		if err != nil {
			t.Errorf("Expected successful creation, got error: %v", err)
		}

		if id == uuid.Nil {
			t.Error("Expected non-nil UUID")
		}

		// Verify status was set correctly
		created := mockCandidaturaRepo.candidaturas[id]
		if created.Status != empregabilidade.StatusCandidaturaEnviada {
			t.Errorf("Expected status to be 'candidatura_enviada', got %s", created.Status)
		}
	})
}

func TestCandidaturaService_Create_VagaNotFound(t *testing.T) {
	t.Run("Error when vaga not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: uuid.New(), // Non-existent vaga
		}

		ctx := context.Background()
		id, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error for non-existent vaga")
		}

		expectedMsg := "vaga não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})
}

func TestCandidaturaService_Create_VagaNotActive(t *testing.T) {
	t.Run("Error when vaga is in draft status", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaEmEdicao,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error for inactive vaga")
		}

		expectedMsg := "vaga não está ativa para candidaturas"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})

	t.Run("Error when vaga is in approval status", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaEmAprovacao,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		_, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error for vaga in approval")
		}

		expectedMsg := "vaga não está ativa para candidaturas"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestCandidaturaService_Create_VagaExpired(t *testing.T) {
	t.Run("Error when vaga is expired by date", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		pastDate := time.Now().Add(-24 * time.Hour) // Yesterday
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:         vagaID,
			Status:     empregabilidade.StatusVagaPublicadoAtivo, // Still active in DB
			DataLimite: &pastDate,                                // But expired by date
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error for expired vaga")
		}

		expectedMsg := "vaga expirada, não aceita mais candidaturas"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})

	t.Run("Error when vaga status is already expired", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoExpirado,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		_, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error for expired vaga")
		}

		expectedMsg := "vaga expirada, não aceita mais candidaturas"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestCandidaturaService_Create_DuplicateCandidatura(t *testing.T) {
	t.Run("Error when candidatura already exists", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.existingCheck = true // Simulate existing candidatura

		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error for duplicate candidatura")
		}

		expectedMsg := "candidatura já existe para esta vaga"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})
}

func TestCandidaturaService_Create_RepositoryErrors(t *testing.T) {
	t.Run("Error when vaga repository fails", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()
		mockVagaRepo.getError = errors.New("database connection error")

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: uuid.New(),
		}

		ctx := context.Background()
		_, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error when vaga repository fails")
		}

		expectedMsg := "database connection error"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Error when check existing candidatura fails", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.checkError = errors.New("database error")

		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		_, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error when check fails")
		}
	})
}

func TestCandidaturaService_UpdateStatus(t *testing.T) {
	t.Run("Valid status update", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaAprovada)

		if err != nil {
			t.Errorf("Expected successful status update, got error: %v", err)
		}
	})

	t.Run("Invalid status update", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateStatus(ctx, uuid.New(), empregabilidade.StatusCandidatura("invalid_status"))

		if err == nil {
			t.Error("Expected error for invalid status")
		}

		expectedMsg := "status inválido"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("UpdateStatus candidatura not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateStatus(ctx, uuid.New(), empregabilidade.StatusCandidaturaAprovada)

		if err == nil {
			t.Error("Expected error when candidatura not found")
		}

		expectedMsg := "candidatura não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("UpdateStatus repository error", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.getError = errors.New("database error")
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateStatus(ctx, uuid.New(), empregabilidade.StatusCandidaturaAprovada)

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})

	t.Run("UpdateStatus invalid transition", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaDescontinuada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaAprovada)

		if err == nil {
			t.Error("Expected error for invalid state transition")
		}

		if !strings.Contains(err.Error(), "transição de status inválida") {
			t.Errorf("Expected transition error, got '%s'", err.Error())
		}
	})

	t.Run("UpdateStatus update error", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}
		mockCandidaturaRepo.updateStatusErr = errors.New("update failed")

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaAprovada)

		if err == nil {
			t.Error("Expected error when UpdateStatus fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

func TestCandidaturaService_ApproveReject(t *testing.T) {
	t.Run("Approve candidatura", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Approve(ctx, id)

		if err != nil {
			t.Errorf("Expected successful approval, got error: %v", err)
		}

		if mockCandidaturaRepo.candidaturas[id].Status != empregabilidade.StatusCandidaturaAprovada {
			t.Error("Expected status to be approved")
		}
	})

	t.Run("Reject candidatura", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Reject(ctx, id)

		if err != nil {
			t.Errorf("Expected successful rejection, got error: %v", err)
		}

		if mockCandidaturaRepo.candidaturas[id].Status != empregabilidade.StatusCandidaturaReprovada {
			t.Error("Expected status to be rejected")
		}
	})

	t.Run("Approve candidatura not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Approve(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when candidatura not found")
		}

		expectedMsg := "candidatura não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Reject candidatura not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Reject(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when candidatura not found")
		}

		expectedMsg := "candidatura não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Approve with repository error", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.getError = errors.New("database error")
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Approve(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})

	t.Run("Reject with repository error", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.getError = errors.New("database error")
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Reject(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})

	t.Run("Approve invalid state transition", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaDescontinuada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Approve(ctx, id)

		if err == nil {
			t.Error("Expected error for invalid state transition")
		}

		if !strings.Contains(err.Error(), "não pode ser aprovada") {
			t.Errorf("Expected state transition error, got '%s'", err.Error())
		}
	})

	t.Run("Reject invalid state transition", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaDescontinuada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Reject(ctx, id)

		if err == nil {
			t.Error("Expected error for invalid state transition")
		}

		if !strings.Contains(err.Error(), "não pode ser reprovada") {
			t.Errorf("Expected state transition error, got '%s'", err.Error())
		}
	})

	t.Run("Approve update status error", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}
		mockCandidaturaRepo.updateStatusErr = errors.New("update failed")

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Approve(ctx, id)

		if err == nil {
			t.Error("Expected error when UpdateStatus fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})

	t.Run("Reject update status error", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:     id,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}
		mockCandidaturaRepo.updateStatusErr = errors.New("update failed")

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Reject(ctx, id)

		if err == nil {
			t.Error("Expected error when UpdateStatus fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

func TestCandidaturaService_Update(t *testing.T) {
	t.Run("Update preserves controlled fields", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		vagaID := uuid.New()
		etapaID := uuid.New()
		originalCPF := "12345678901"
		originalStatus := empregabilidade.StatusCandidaturaAprovada

		originalSnapshot := &empregabilidade.CurriculoCompleto{
			Formacoes: []*empregabilidade.CurriculoFormacao{
				{NomeInstituicao: "UFRJ"},
			},
		}

		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:                id,
			CPF:               originalCPF,
			IDVaga:            vagaID,
			Status:            originalStatus,
			IDEtapaAtual:      &etapaID,
			CurriculoSnapshot: originalSnapshot,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		differentVagaID := uuid.New()
		updated := &empregabilidade.Candidatura{
			ID:     id,
			CPF:    "99999999999",
			IDVaga: differentVagaID,
			Status: empregabilidade.StatusCandidaturaReprovada,
		}

		ctx := context.Background()
		err := service.Update(ctx, updated)

		if err != nil {
			t.Errorf("Expected successful update, got error: %v", err)
		}

		result := mockCandidaturaRepo.candidaturas[id]

		if result.CPF != originalCPF {
			t.Errorf("Expected CPF to be preserved as '%s', got '%s'", originalCPF, result.CPF)
		}

		if result.IDVaga != vagaID {
			t.Errorf("Expected IDVaga to be preserved as '%s', got '%s'", vagaID, result.IDVaga)
		}

		if result.Status != originalStatus {
			t.Errorf("Expected Status to be preserved as '%s', got '%s'", originalStatus, result.Status)
		}

		if result.IDEtapaAtual == nil || *result.IDEtapaAtual != etapaID {
			t.Error("Expected IDEtapaAtual to be preserved")
		}

		if result.CurriculoSnapshot == nil || len(result.CurriculoSnapshot.Formacoes) != 1 {
			t.Error("Expected CurriculoSnapshot to be preserved")
		}
	})

	t.Run("Update returns error when candidatura not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		updated := &empregabilidade.Candidatura{
			ID: uuid.New(),
		}

		ctx := context.Background()
		err := service.Update(ctx, updated)

		if err == nil {
			t.Error("Expected error when candidatura not found")
		}

		expectedMsg := "candidatura não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Update returns error when GetByID fails", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.getError = errors.New("database error")
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		updated := &empregabilidade.Candidatura{
			ID: uuid.New(),
		}

		ctx := context.Background()
		err := service.Update(ctx, updated)

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})
}

func TestCandidaturaService_Create_CurriculoSnapshot(t *testing.T) {
	t.Run("Successful candidatura includes curriculo snapshot", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		expectedCurriculo := &empregabilidade.CurriculoCompleto{
			Formacoes: []*empregabilidade.CurriculoFormacao{
				{NomeInstituicao: "Test University"},
			},
		}
		mockCurriculoService.curriculo = expectedCurriculo

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, candidatura)

		if err != nil {
			t.Errorf("Expected successful creation, got error: %v", err)
		}

		created := mockCandidaturaRepo.candidaturas[id]
		if created.CurriculoSnapshot == nil {
			t.Error("Expected curriculo snapshot to be set")
		}

		if len(created.CurriculoSnapshot.Formacoes) != 1 {
			t.Error("Expected curriculo snapshot to have formacoes")
		}
	})

	t.Run("Candidatura created without snapshot when curriculo service fails", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()
		mockCurriculoService.getError = errors.New("curriculo service error")

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		id, err := service.Create(ctx, candidatura)

		if err != nil {
			t.Errorf("Expected candidatura to be created even when curriculo fails, got error: %v", err)
		}

		if id == uuid.Nil {
			t.Error("Expected non-nil UUID")
		}

		created := mockCandidaturaRepo.candidaturas[id]
		if created.CurriculoSnapshot != nil {
			t.Error("Expected curriculo snapshot to be nil when curriculo service fails")
		}
	})
}

// ==================== UpdateEtapa Tests ====================

func TestCandidaturaService_UpdateEtapa_Success(t *testing.T) {
	t.Run("Successfully update etapa", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		candID := uuid.New()
		vagaID := uuid.New()
		etapaID := uuid.New()

		mockCandidaturaRepo.candidaturas[candID] = &empregabilidade.Candidatura{
			ID:     candID,
			IDVaga: vagaID,
		}

		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID: vagaID,
			Etapas: []empregabilidade.Etapa{
				{ID: etapaID, Titulo: "Entrevista"},
			},
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateEtapa(ctx, candID, etapaID)

		if err != nil {
			t.Errorf("Expected successful etapa update, got error: %v", err)
		}
	})
}

func TestCandidaturaService_UpdateEtapa_Errors(t *testing.T) {
	t.Run("Error when candidatura not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateEtapa(ctx, uuid.New(), uuid.New())

		if err == nil || err.Error() != "candidatura não encontrada" {
			t.Errorf("Expected 'candidatura não encontrada', got: %v", err)
		}
	})

	t.Run("Error when etapa not belongs to vaga", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		candID := uuid.New()
		vagaID := uuid.New()
		etapaID := uuid.New()

		mockCandidaturaRepo.candidaturas[candID] = &empregabilidade.Candidatura{
			ID:     candID,
			IDVaga: vagaID,
		}

		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Etapas: []empregabilidade.Etapa{},
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.UpdateEtapa(ctx, candID, etapaID)

		if err == nil {
			t.Error("Expected error when etapa doesn't belong to vaga")
		}
	})
}

// ==================== BulkUpdateStatus Tests ====================

func TestCandidaturaService_BulkUpdateStatus_Success(t *testing.T) {
	t.Run("Successfully bulk update status", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		cpf1 := "11111111111"
		cpf2 := "22222222222"

		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf1,
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}
		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf2,
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		result, err := service.BulkUpdateStatus(ctx, vagaID, []string{cpf1, cpf2}, empregabilidade.StatusCandidaturaAprovada)

		if err != nil {
			t.Errorf("Expected successful bulk update, got error: %v", err)
		}

		if result.Updated != 2 {
			t.Errorf("Expected 2 updates, got %d", result.Updated)
		}
	})
}

func TestCandidaturaService_BulkUpdateStatus_Errors(t *testing.T) {
	t.Run("Error with invalid status", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateStatus(ctx, uuid.New(), []string{"12345678901"}, "invalid_status")

		if err == nil {
			t.Error("Expected error with invalid status")
		}
	})

	t.Run("Error with empty CPF list", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateStatus(ctx, uuid.New(), []string{}, empregabilidade.StatusCandidaturaAprovada)

		if err == nil || err.Error() != "lista de CPFs não pode ser vazia" {
			t.Errorf("Expected 'lista de CPFs não pode ser vazia', got: %v", err)
		}
	})

	t.Run("Invalid state transitions are filtered out", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		cpf1 := "11111111111"
		cpf2 := "22222222222"

		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf1,
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaDescontinuada, // Cannot transition from discontinued
		}
		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf2,
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaEnviada, // Can transition
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		result, err := service.BulkUpdateStatus(ctx, vagaID, []string{cpf1, cpf2}, empregabilidade.StatusCandidaturaAprovada)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Only cpf2 should be updated
		if result.Updated != 1 {
			t.Errorf("Expected 1 update, got %d", result.Updated)
		}

		if len(result.FailedCPFs) != 1 {
			t.Errorf("Expected 1 failed CPF, got %d", len(result.FailedCPFs))
		}

		if len(result.FailedCPFs) > 0 && result.FailedCPFs[0] != cpf1 {
			t.Errorf("Expected failed CPF to be %s, got %s", cpf1, result.FailedCPFs[0])
		}
	})

	t.Run("BulkGetByCPFs error", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.bulkGetError = errors.New("database error")
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateStatus(ctx, uuid.New(), []string{"12345678901"}, empregabilidade.StatusCandidaturaAprovada)

		if err == nil {
			t.Error("Expected error when BulkGetByCPFs fails")
		}
	})

	t.Run("CPFs not found are marked as failed", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		cpf1 := "11111111111"
		cpf2 := "22222222222" // Not in database

		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf1,
			IDVaga: vagaID,
			Status: empregabilidade.StatusCandidaturaEnviada,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		result, err := service.BulkUpdateStatus(ctx, vagaID, []string{cpf1, cpf2}, empregabilidade.StatusCandidaturaAprovada)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Updated != 1 {
			t.Errorf("Expected 1 update, got %d", result.Updated)
		}

		if len(result.FailedCPFs) != 1 {
			t.Errorf("Expected 1 failed CPF, got %d", len(result.FailedCPFs))
		}
	})
}

// ==================== BulkUpdateEtapa Tests ====================

func TestCandidaturaService_BulkUpdateEtapa_Success(t *testing.T) {
	t.Run("Successfully bulk update etapa", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		etapaID := uuid.New()
		cpf1 := "11111111111"
		cpf2 := "22222222222"

		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf1,
			IDVaga: vagaID,
		}
		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf2,
			IDVaga: vagaID,
		}

		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID: vagaID,
			Etapas: []empregabilidade.Etapa{
				{ID: etapaID, Titulo: "Entrevista"},
			},
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		result, err := service.BulkUpdateEtapa(ctx, vagaID, []string{cpf1, cpf2}, etapaID)

		if err != nil {
			t.Errorf("Expected successful bulk update, got error: %v", err)
		}

		if result.Updated != 2 {
			t.Errorf("Expected 2 updates, got %d", result.Updated)
		}
	})
}

func TestCandidaturaService_BulkUpdateEtapa_Errors(t *testing.T) {
	t.Run("Error when vaga not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateEtapa(ctx, uuid.New(), []string{"12345678901"}, uuid.New())

		if err == nil {
			t.Error("Expected error when vaga not found")
		}
	})

	t.Run("Error when etapa not belongs to vaga", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Etapas: []empregabilidade.Etapa{},
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateEtapa(ctx, vagaID, []string{"12345678901"}, uuid.New())

		if err == nil || err.Error() != "etapa não pertence à vaga" {
			t.Errorf("Expected 'etapa não pertence à vaga', got: %v", err)
		}
	})

	t.Run("Error with empty CPF list", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateEtapa(ctx, uuid.New(), []string{}, uuid.New())

		if err == nil || err.Error() != "lista de CPFs não pode ser vazia" {
			t.Errorf("Expected 'lista de CPFs não pode ser vazia', got: %v", err)
		}
	})

	t.Run("Error when vaga repository fails", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockVagaRepo.getError = errors.New("database error")
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateEtapa(ctx, uuid.New(), []string{"12345678901"}, uuid.New())

		if err == nil {
			t.Error("Expected error when vaga repository fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})

	t.Run("Error when BulkGetByCPFs fails", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockCandidaturaRepo.bulkGetError = errors.New("database error")
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		etapaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Etapas: []empregabilidade.Etapa{{ID: etapaID}},
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateEtapa(ctx, vagaID, []string{"12345678901"}, etapaID)

		if err == nil {
			t.Error("Expected error when BulkGetByCPFs fails")
		}
	})

	t.Run("Error when candidaturas have different etapas", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		etapaID := uuid.New()
		etapa1 := uuid.New()
		etapa2 := uuid.New()
		cpf1 := "11111111111"
		cpf2 := "22222222222"

		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:          cpf1,
			IDVaga:       vagaID,
			IDEtapaAtual: &etapa1,
		}
		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:          cpf2,
			IDVaga:       vagaID,
			IDEtapaAtual: &etapa2,
		}

		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Etapas: []empregabilidade.Etapa{{ID: etapaID}},
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		_, err := service.BulkUpdateEtapa(ctx, vagaID, []string{cpf1, cpf2}, etapaID)

		if err == nil {
			t.Error("Expected error when candidaturas have different etapas")
		}

		if !strings.Contains(err.Error(), "mesma etapa") {
			t.Errorf("Expected error about same etapa, got: %s", err.Error())
		}
	})

	t.Run("Success with failed CPFs (not found)", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		vagaID := uuid.New()
		etapaID := uuid.New()
		cpf1 := "11111111111"
		cpf2 := "22222222222" // Not in database

		mockCandidaturaRepo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
			CPF:    cpf1,
			IDVaga: vagaID,
		}

		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Etapas: []empregabilidade.Etapa{{ID: etapaID}},
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		result, err := service.BulkUpdateEtapa(ctx, vagaID, []string{cpf1, cpf2}, etapaID)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Updated != 1 {
			t.Errorf("Expected 1 update, got %d", result.Updated)
		}

		if len(result.FailedCPFs) != 1 {
			t.Errorf("Expected 1 failed CPF, got %d", len(result.FailedCPFs))
		}

		if len(result.FailedCPFs) > 0 && result.FailedCPFs[0] != cpf2 {
			t.Errorf("Expected failed CPF to be %s, got %s", cpf2, result.FailedCPFs[0])
		}
	})
}

// ==================== Delete Tests ====================

func TestCandidaturaService_Delete_Success(t *testing.T) {
	t.Run("Successfully delete candidatura", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		id := uuid.New()
		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID: id,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		err := service.Delete(ctx, id)

		if err != nil {
			t.Errorf("Expected successful delete, got error: %v", err)
		}

		if _, exists := mockCandidaturaRepo.candidaturas[id]; exists {
			t.Error("Expected candidatura to be deleted")
		}
	})
}

// ==================== List and CountByStatus Tests ====================

func TestCandidaturaService_List_Success(t *testing.T) {
	t.Run("Successfully list candidaturas", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		result, total, err := service.List(ctx, empregabilidade.CandidaturaFilter{}, 1, 10)

		if err != nil {
			t.Errorf("Expected successful list, got error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		if total < 0 {
			t.Errorf("Expected non-negative total, got %d", total)
		}
	})
}

func TestCandidaturaService_CountByStatus_Success(t *testing.T) {
	t.Run("Successfully count by status", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService, nil, nil)

		ctx := context.Background()
		result, err := service.CountByStatus(ctx, empregabilidade.CandidaturaFilter{})

		if err != nil {
			t.Errorf("Expected successful count, got error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}
	})
}

func TestCandidaturaService_EnrichRespostasWithTitulo(t *testing.T) {
	infoID1 := uuid.New()
	infoID2 := uuid.New()
	infoIDOrfao := uuid.New()

	vaga := &empregabilidade.Vaga{
		ID: uuid.New(),
		InformacoesComplementares: []empregabilidade.InformacaoComplementar{
			{ID: infoID1, Titulo: "Tem CNH?"},
			{ID: infoID2, Titulo: "Anos de experiência"},
		},
	}

	tests := []struct {
		name        string
		candidatura *empregabilidade.Candidatura
		wantTitulos []string
	}{
		{
			name: "popula titulo a partir da vaga",
			candidatura: &empregabilidade.Candidatura{
				Vaga: vaga,
				RespostasInfoComplementares: []empregabilidade.RespostaInfoComplementar{
					{IDInfo: infoID1, Resposta: "Sim"},
					{IDInfo: infoID2, Resposta: "3"},
				},
			},
			wantTitulos: []string{"Tem CNH?", "Anos de experiência"},
		},
		{
			name: "id_info sem correspondencia fica com titulo vazio",
			candidatura: &empregabilidade.Candidatura{
				Vaga: vaga,
				RespostasInfoComplementares: []empregabilidade.RespostaInfoComplementar{
					{IDInfo: infoIDOrfao, Resposta: "alguma coisa"},
				},
			},
			wantTitulos: []string{""},
		},
		{
			name: "candidatura sem vaga nao panics",
			candidatura: &empregabilidade.Candidatura{
				Vaga: nil,
				RespostasInfoComplementares: []empregabilidade.RespostaInfoComplementar{
					{IDInfo: infoID1, Resposta: "Sim"},
				},
			},
			wantTitulos: []string{""},
		},
		{
			name:        "candidatura nil nao panics",
			candidatura: nil,
			wantTitulos: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := services.NewCandidaturaService(NewMockCandidaturaRepo(), NewMockVagaRepo(), NewMockCurriculoService(), nil, nil)
			svc.EnrichRespostasWithTitulo(tt.candidatura)

			if tt.candidatura == nil {
				return
			}
			for i, resposta := range tt.candidatura.RespostasInfoComplementares {
				if resposta.Titulo != tt.wantTitulos[i] {
					t.Errorf("resposta[%d].Titulo = %q, want %q", i, resposta.Titulo, tt.wantTitulos[i])
				}
			}
		})
	}
}
