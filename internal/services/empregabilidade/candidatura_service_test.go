package empregabilidade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
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

func (m *MockCandidaturaRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	return []*empregabilidade.Candidatura{}, 0, nil
}

func (m *MockCandidaturaRepo) ListByCPF(ctx context.Context, cpf string, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	return []*empregabilidade.Candidatura{}, 0, nil
}

func (m *MockCandidaturaRepo) ListByVaga(ctx context.Context, vagaID uuid.UUID, status string, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

		ctx := context.Background()
		err := service.UpdateStatus(ctx, uuid.New(), empregabilidade.StatusCandidaturaAprovada)

		if err != nil {
			t.Errorf("Expected successful status update, got error: %v", err)
		}
	})

	t.Run("Invalid status update", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

		ctx := context.Background()
		err := service.Reject(ctx, id)

		if err != nil {
			t.Errorf("Expected successful rejection, got error: %v", err)
		}

		if mockCandidaturaRepo.candidaturas[id].Status != empregabilidade.StatusCandidaturaReprovada {
			t.Error("Expected status to be rejected")
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

		mockCandidaturaRepo.candidaturas[id] = &empregabilidade.Candidatura{
			ID:           id,
			CPF:          originalCPF,
			IDVaga:       vagaID,
			Status:       originalStatus,
			IDEtapaAtual: &etapaID,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

		// Try to update with different controlled values
		differentVagaID := uuid.New()
		updated := &empregabilidade.Candidatura{
			ID:     id,
			CPF:    "99999999999",                               // Should be preserved
			IDVaga: differentVagaID,                             // Should be preserved
			Status: empregabilidade.StatusCandidaturaReprovada,  // Should be preserved
		}

		ctx := context.Background()
		err := service.Update(ctx, updated)

		if err != nil {
			t.Errorf("Expected successful update, got error: %v", err)
		}

		result := mockCandidaturaRepo.candidaturas[id]

		// Verify controlled fields are preserved
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
	})

	t.Run("Update returns error when candidatura not found", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

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

	t.Run("Error when curriculo service fails", func(t *testing.T) {
		mockCandidaturaRepo := NewMockCandidaturaRepo()
		mockVagaRepo := NewMockVagaRepo()
		mockCurriculoService := NewMockCurriculoService()
		mockCurriculoService.getError = errors.New("curriculo service error")

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		service := services.NewCandidaturaService(mockCandidaturaRepo, mockVagaRepo, mockCurriculoService)

		candidatura := &empregabilidade.Candidatura{
			CPF:    "12345678901",
			IDVaga: vagaID,
		}

		ctx := context.Background()
		_, err := service.Create(ctx, candidatura)

		if err == nil {
			t.Error("Expected error when curriculo service fails")
		}

		expectedMsg := "curriculo service error"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}
