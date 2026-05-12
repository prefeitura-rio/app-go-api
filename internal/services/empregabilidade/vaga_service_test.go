package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

func TestNewVagaService(t *testing.T) {
	mockVagaRepo := NewMockVagaRepoForService()
	mockEmpresaRepo := NewMockEmpresaRepo()
	mockCandidaturaRepo := NewMockCandidaturaRepo()
	service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)
	assert.NotNil(t, service)
}

// Mock Vaga Repository for VagaService tests
type MockVagaRepoForService struct {
	vagas       map[uuid.UUID]*empregabilidade.Vaga
	createError error
	getError    error
	updateError error
	deleteError error
	listError   error
}

func NewMockVagaRepoForService() *MockVagaRepoForService {
	return &MockVagaRepoForService{
		vagas: make(map[uuid.UUID]*empregabilidade.Vaga),
	}
}

func (m *MockVagaRepoForService) Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	if m.createError != nil {
		return uuid.Nil, m.createError
	}
	id := uuid.New()
	entity.ID = id
	m.vagas[id] = entity
	return id, nil
}

func (m *MockVagaRepoForService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	vaga, exists := m.vagas[id]
	if !exists {
		return nil, nil
	}
	return vaga, nil
}

func (m *MockVagaRepoForService) GetByIDPrefix(ctx context.Context, idPrefix string) (*empregabilidade.Vaga, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	for id, vaga := range m.vagas {
		if len(id.String()) >= len(idPrefix) && id.String()[:len(idPrefix)] == idPrefix {
			return vaga, nil
		}
	}
	return nil, nil
}

func (m *MockVagaRepoForService) Update(ctx context.Context, entity *empregabilidade.Vaga) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.vagas[entity.ID] = entity
	return nil
}

func (m *MockVagaRepoForService) UpdateWithAssociations(ctx context.Context, entity *empregabilidade.Vaga) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.vagas[entity.ID] = entity
	return nil
}

func (m *MockVagaRepoForService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.vagas, id)
	return nil
}

func (m *MockVagaRepoForService) List(ctx context.Context, filter empregabilidade.VagaFilter, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	var result []*empregabilidade.Vaga
	for _, v := range m.vagas {
		result = append(result, v)
	}
	return result, len(result), nil
}

func (m *MockVagaRepoForService) UpdateTiposPCD(ctx context.Context, vagaID uuid.UUID, tiposPCDIDs []uuid.UUID) error {
	return nil
}

func (m *MockVagaRepoForService) ListByContratante(ctx context.Context, cnpj string, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	return nil, 0, nil
}

func (m *MockVagaRepoForService) ListByOrgaoParceiro(ctx context.Context, orgaoID string, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	return nil, 0, nil
}

func (m *MockVagaRepoForService) ListPublicActive(ctx context.Context, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	return nil, 0, nil
}

// Mock Empresa Repository for VagaService tests
type MockEmpresaRepoForService struct {
	empresas map[string]*empregabilidade.Empresa
	getError error
}

func NewMockEmpresaRepoForService() *MockEmpresaRepoForService {
	return &MockEmpresaRepoForService{
		empresas: make(map[string]*empregabilidade.Empresa),
	}
}

func (m *MockEmpresaRepoForService) GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	empresa, exists := m.empresas[cnpj]
	if !exists {
		return nil, nil
	}
	return empresa, nil
}

// ==================== Publish Tests ====================

func TestVagaService_Publish_Success(t *testing.T) {
	t.Run("Publish from em_edicao status", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaEmEdicao,
		}

		ctx := context.Background()
		err := service.Publish(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful publish, got error: %v", err)
		}

		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaPublicadoAtivo {
			t.Errorf("Expected status to be publicado_ativo, got %s", mockVagaRepo.vagas[vagaID].Status)
		}
	})

	t.Run("Publish from em_aprovacao status", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaEmAprovacao,
		}

		ctx := context.Background()
		err := service.Publish(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful publish, got error: %v", err)
		}

		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaPublicadoAtivo {
			t.Errorf("Expected status to be publicado_ativo, got %s", mockVagaRepo.vagas[vagaID].Status)
		}
	})
}

func TestVagaService_Publish_VagaNotFound(t *testing.T) {
	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.Publish(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when vaga not found")
		}

		expectedMsg := "vaga não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestVagaService_Publish_InvalidStatus(t *testing.T) {
	t.Run("Error when vaga is already published", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		ctx := context.Background()
		err := service.Publish(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when vaga is already published")
		}

		expectedMsg := "vaga não está em estado de edição ou aprovação"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Error when vaga is expired", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaPublicadoExpirado,
		}

		ctx := context.Background()
		err := service.Publish(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when vaga is expired")
		}

		expectedMsg := "vaga não está em estado de edição ou aprovação"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestVagaService_Publish_RepositoryErrors(t *testing.T) {
	t.Run("Error when GetByID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockVagaRepo.getError = errors.New("database connection error")
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.Publish(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database connection error" {
			t.Errorf("Expected 'database connection error', got '%s'", err.Error())
		}
	})

	t.Run("Error when Update fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaEmEdicao,
		}
		mockVagaRepo.updateError = errors.New("update failed")

		ctx := context.Background()
		err := service.Publish(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when Update fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

// ==================== SendToApproval Tests ====================

func TestVagaService_SendToApproval_Success(t *testing.T) {
	t.Run("SendToApproval from em_edicao status", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaEmEdicao,
		}

		ctx := context.Background()
		err := service.SendToApproval(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful send to approval, got error: %v", err)
		}

		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaEmAprovacao {
			t.Errorf("Expected status to be em_aprovacao, got %s", mockVagaRepo.vagas[vagaID].Status)
		}
	})
}

func TestVagaService_SendToApproval_VagaNotFound(t *testing.T) {
	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.SendToApproval(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when vaga not found")
		}

		expectedMsg := "vaga não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestVagaService_SendToApproval_InvalidStatus(t *testing.T) {
	testCases := []struct {
		name   string
		status empregabilidade.StatusVaga
	}{
		{"Error when vaga is in approval", empregabilidade.StatusVagaEmAprovacao},
		{"Error when vaga is published", empregabilidade.StatusVagaPublicadoAtivo},
		{"Error when vaga is expired", empregabilidade.StatusVagaPublicadoExpirado},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockVagaRepo := NewMockVagaRepoForService()
			mockEmpresaRepo := NewMockEmpresaRepoForService()
			service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

			vagaID := uuid.New()
			mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
				ID:     vagaID,
				Titulo: "Desenvolvedor Go",
				Status: tc.status,
			}

			ctx := context.Background()
			err := service.SendToApproval(ctx, vagaID)

			if err == nil {
				t.Errorf("Expected error for status %s", tc.status)
			}

			expectedMsg := "vaga não está em estado de edição"
			if err.Error() != expectedMsg {
				t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
			}
		})
	}
}

func TestVagaService_SendToApproval_RepositoryErrors(t *testing.T) {
	t.Run("Error when GetByID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockVagaRepo.getError = errors.New("database connection error")
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.SendToApproval(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database connection error" {
			t.Errorf("Expected 'database connection error', got '%s'", err.Error())
		}
	})

	t.Run("Error when Update fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaEmEdicao,
		}
		mockVagaRepo.updateError = errors.New("update failed")

		ctx := context.Background()
		err := service.SendToApproval(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when Update fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

// ==================== SendToDraft Tests ====================

func TestVagaService_SendToDraft_Success(t *testing.T) {
	t.Run("SendToDraft from em_aprovacao status", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaEmAprovacao,
		}

		ctx := context.Background()
		err := service.SendToDraft(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful send to draft, got error: %v", err)
		}

		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaEmEdicao {
			t.Errorf("Expected status to be em_edicao, got %s", mockVagaRepo.vagas[vagaID].Status)
		}
	})
}

func TestVagaService_SendToDraft_VagaNotFound(t *testing.T) {
	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.SendToDraft(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when vaga not found")
		}

		expectedMsg := "vaga não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestVagaService_SendToDraft_InvalidStatus(t *testing.T) {
	testCases := []struct {
		name   string
		status empregabilidade.StatusVaga
	}{
		{"Error when vaga is in editing", empregabilidade.StatusVagaEmEdicao},
		{"Error when vaga is published active", empregabilidade.StatusVagaPublicadoAtivo},
		{"Error when vaga is published expired", empregabilidade.StatusVagaPublicadoExpirado},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockVagaRepo := NewMockVagaRepoForService()
			mockEmpresaRepo := NewMockEmpresaRepoForService()
			service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

			vagaID := uuid.New()
			mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
				ID:     vagaID,
				Titulo: "Desenvolvedor Go",
				Status: tc.status,
			}

			ctx := context.Background()
			err := service.SendToDraft(ctx, vagaID)

			if err == nil {
				t.Errorf("Expected error for status %s", tc.status)
			}

			expectedMsg := "vaga não está em estado de aprovação"
			if err.Error() != expectedMsg {
				t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
			}
		})
	}
}

// ==================== Create Tests ====================

func TestVagaService_Create_Success(t *testing.T) {
	t.Run("Successful creation with valid contratante", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		cnpj := "12.345.678/0001-90"
		mockEmpresaRepo.empresas[cnpj] = &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Teste",
		}

		vaga := &empregabilidade.Vaga{
			Titulo:        "Desenvolvedor Go",
			Descricao:     "Descrição da vaga",
			IDContratante: cnpj,
			Status:        empregabilidade.StatusVagaPublicadoAtivo, // Tentando bypass
		}

		ctx := context.Background()
		id, err := service.Create(ctx, vaga)

		if err != nil {
			t.Errorf("Expected successful creation, got error: %v", err)
		}

		if id == uuid.Nil {
			t.Error("Expected non-nil UUID")
		}

		// Verifica que o status foi forçado para em_edicao
		if vaga.Status != empregabilidade.StatusVagaEmEdicao {
			t.Errorf("Expected status to be forced to em_edicao, got %s", vaga.Status)
		}
	})
}

func TestVagaService_Create_ContratanteNotFound(t *testing.T) {
	t.Run("Error when contratante not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vaga := &empregabilidade.Vaga{
			Titulo:        "Desenvolvedor Go",
			Descricao:     "Descrição da vaga",
			IDContratante: "99.999.999/0001-99",
		}

		ctx := context.Background()
		id, err := service.Create(ctx, vaga)

		if err == nil {
			t.Error("Expected error when contratante not found")
		}

		expectedMsg := "empresa contratante não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})
}

func TestVagaService_Create_EmpresaRepoError(t *testing.T) {
	t.Run("Error when empresa repository fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockEmpresaRepo.getError = errors.New("database error")
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vaga := &empregabilidade.Vaga{
			Titulo:        "Desenvolvedor Go",
			Descricao:     "Descrição da vaga",
			IDContratante: "12.345.678/0001-90",
		}

		ctx := context.Background()
		id, err := service.Create(ctx, vaga)

		if err == nil {
			t.Error("Expected error when empresa repository fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})
}

// ==================== CreateDraft Tests ====================

func TestVagaService_CreateDraft_Success(t *testing.T) {
	t.Run("Successful draft creation with valid contratante", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		cnpj := "12.345.678/0001-90"
		mockEmpresaRepo.empresas[cnpj] = &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Teste",
		}

		vaga := &empregabilidade.Vaga{
			Titulo:        "Desenvolvedor Go",
			IDContratante: cnpj,
		}

		ctx := context.Background()
		id, err := service.CreateDraft(ctx, vaga)

		if err != nil {
			t.Errorf("Expected successful creation, got error: %v", err)
		}

		if id == uuid.Nil {
			t.Error("Expected non-nil UUID")
		}

		if vaga.Status != empregabilidade.StatusVagaEmEdicao {
			t.Errorf("Expected status to be em_edicao, got %s", vaga.Status)
		}
	})
}

func TestVagaService_CreateDraft_ContratanteNotFound(t *testing.T) {
	t.Run("Error when contratante not found in draft creation", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vaga := &empregabilidade.Vaga{
			Titulo:        "Desenvolvedor Go",
			IDContratante: "99.999.999/0001-99",
		}

		ctx := context.Background()
		id, err := service.CreateDraft(ctx, vaga)

		if err == nil {
			t.Error("Expected error when contratante not found")
		}

		expectedMsg := "empresa contratante não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}

		if id != uuid.Nil {
			t.Error("Expected nil UUID when creation fails")
		}
	})
}

// ==================== Update Tests ====================

func TestVagaService_Update_Success(t *testing.T) {
	t.Run("Successful update preserves status", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go",
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		updated := &empregabilidade.Vaga{
			ID:     vagaID,
			Titulo: "Desenvolvedor Go Senior",
			Status: empregabilidade.StatusVagaEmEdicao, // Tentando mudar status via Update
		}

		ctx := context.Background()
		err := service.Update(ctx, updated)

		if err != nil {
			t.Errorf("Expected successful update, got error: %v", err)
		}

		// Verifica que o status foi preservado
		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaPublicadoAtivo {
			t.Errorf("Expected status to be preserved as publicado_ativo, got %s", mockVagaRepo.vagas[vagaID].Status)
		}

		// Verifica que o título foi atualizado
		if mockVagaRepo.vagas[vagaID].Titulo != "Desenvolvedor Go Senior" {
			t.Errorf("Expected titulo to be updated")
		}
	})
}

func TestVagaService_Update_VagaNotFound(t *testing.T) {
	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vaga := &empregabilidade.Vaga{
			ID:     uuid.New(),
			Titulo: "Desenvolvedor Go",
		}

		ctx := context.Background()
		err := service.Update(ctx, vaga)

		if err == nil {
			t.Error("Expected error when vaga not found")
		}

		expectedMsg := "vaga não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestVagaService_Update_RepositoryErrors(t *testing.T) {
	t.Run("Error when GetByID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockVagaRepo.getError = errors.New("database error")
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vaga := &empregabilidade.Vaga{
			ID:     uuid.New(),
			Titulo: "Desenvolvedor Go",
		}

		ctx := context.Background()
		err := service.Update(ctx, vaga)

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})
}

// ==================== FreezeVaga Tests ====================

func TestVagaService_FreezeVaga_Success(t *testing.T) {
	t.Run("Successfully freeze active vaga", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		ctx := context.Background()
		err := service.FreezeVaga(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful freeze, got error: %v", err)
		}

		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaCongelada {
			t.Errorf("Expected status vaga_congelada, got %s", mockVagaRepo.vagas[vagaID].Status)
		}
	})
}

func TestVagaService_FreezeVaga_Errors(t *testing.T) {
	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.FreezeVaga(ctx, uuid.New())

		if err == nil || err.Error() != "vaga não encontrada" {
			t.Errorf("Expected 'vaga não encontrada', got: %v", err)
		}
	})

	t.Run("Error when vaga not active", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaEmEdicao,
		}

		ctx := context.Background()
		err := service.FreezeVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when vaga not active")
		}
	})
}

// ==================== UnfreezeVaga Tests ====================

func TestVagaService_UnfreezeVaga_Success(t *testing.T) {
	t.Run("Successfully unfreeze frozen vaga", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaCongelada,
		}

		ctx := context.Background()
		err := service.UnfreezeVaga(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful unfreeze, got error: %v", err)
		}

		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaPublicadoAtivo {
			t.Errorf("Expected status publicado_ativo, got %s", mockVagaRepo.vagas[vagaID].Status)
		}
	})
}

func TestVagaService_UnfreezeVaga_Errors(t *testing.T) {
	t.Run("Error when vaga not frozen", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		ctx := context.Background()
		err := service.UnfreezeVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when vaga not frozen")
		}
	})

	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.UnfreezeVaga(ctx, uuid.New())

		if err == nil || err.Error() != "vaga não encontrada" {
			t.Errorf("Expected 'vaga não encontrada', got: %v", err)
		}
	})

	t.Run("Error when GetByID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockVagaRepo.getError = errors.New("database error")
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.UnfreezeVaga(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})

	t.Run("Error when BulkRestoreStatusByVagaID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{
			restoreError: errors.New("restore failed"),
		}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaCongelada,
		}

		ctx := context.Background()
		err := service.UnfreezeVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when BulkRestoreStatusByVagaID fails")
		}

		if err.Error() != "restore failed" {
			t.Errorf("Expected 'restore failed', got '%s'", err.Error())
		}
	})

	t.Run("Error when Update fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaCongelada,
		}
		mockVagaRepo.updateError = errors.New("update failed")

		ctx := context.Background()
		err := service.UnfreezeVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when Update fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

// ==================== DiscontinueVaga Tests ====================

func TestVagaService_DiscontinueVaga_Success(t *testing.T) {
	testCases := []struct {
		name   string
		status empregabilidade.StatusVaga
	}{
		{"Discontinue active vaga", empregabilidade.StatusVagaPublicadoAtivo},
		{"Discontinue expired vaga", empregabilidade.StatusVagaPublicadoExpirado},
		{"Discontinue frozen vaga", empregabilidade.StatusVagaCongelada},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockVagaRepo := NewMockVagaRepoForService()
			mockEmpresaRepo := NewMockEmpresaRepoForService()
			mockCandidaturaRepo := &MockCandidaturaRepoForVaga{}
			service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

			vagaID := uuid.New()
			mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
				ID:     vagaID,
				Status: tc.status,
			}

			ctx := context.Background()
			err := service.DiscontinueVaga(ctx, vagaID)

			if err != nil {
				t.Errorf("Expected successful discontinue, got error: %v", err)
			}

			if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaDescontinuada {
				t.Errorf("Expected status vaga_descontinuada, got %s", mockVagaRepo.vagas[vagaID].Status)
			}
		})
	}
}

func TestVagaService_DiscontinueVaga_Errors(t *testing.T) {
	t.Run("Error when vaga in draft", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaEmEdicao,
		}

		ctx := context.Background()
		err := service.DiscontinueVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when discontinuing draft vaga")
		}
	})

	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.DiscontinueVaga(ctx, uuid.New())

		if err == nil || err.Error() != "vaga não encontrada" {
			t.Errorf("Expected 'vaga não encontrada', got: %v", err)
		}
	})

	t.Run("Error when GetByID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockVagaRepo.getError = errors.New("database error")
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.DiscontinueVaga(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})

	t.Run("Error when BulkSaveAndUpdateStatusByVagaID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{
			saveError: errors.New("save failed"),
		}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		ctx := context.Background()
		err := service.DiscontinueVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when BulkSaveAndUpdateStatusByVagaID fails")
		}

		if err.Error() != "save failed" {
			t.Errorf("Expected 'save failed', got '%s'", err.Error())
		}
	})

	t.Run("Error when Update fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}
		mockVagaRepo.updateError = errors.New("update failed")

		ctx := context.Background()
		err := service.DiscontinueVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when Update fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

// ==================== ReactivateVaga Tests ====================

func TestVagaService_ReactivateVaga_Success(t *testing.T) {
	t.Run("Successfully reactivate discontinued vaga", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaDescontinuada,
		}

		ctx := context.Background()
		err := service.ReactivateVaga(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful reactivate, got error: %v", err)
		}

		if mockVagaRepo.vagas[vagaID].Status != empregabilidade.StatusVagaPublicadoAtivo {
			t.Errorf("Expected status publicado_ativo, got %s", mockVagaRepo.vagas[vagaID].Status)
		}
	})
}

func TestVagaService_ReactivateVaga_Errors(t *testing.T) {
	t.Run("Error when vaga not discontinued", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaPublicadoAtivo,
		}

		ctx := context.Background()
		err := service.ReactivateVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when reactivating non-discontinued vaga")
		}
	})

	t.Run("Error when vaga not found", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.ReactivateVaga(ctx, uuid.New())

		if err == nil || err.Error() != "vaga não encontrada" {
			t.Errorf("Expected 'vaga não encontrada', got: %v", err)
		}
	})

	t.Run("Error when GetByID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockVagaRepo.getError = errors.New("database error")
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		ctx := context.Background()
		err := service.ReactivateVaga(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error when GetByID fails")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})

	t.Run("Error when BulkRestoreStatusByVagaID fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{
			restoreError: errors.New("restore failed"),
		}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaDescontinuada,
		}

		ctx := context.Background()
		err := service.ReactivateVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when BulkRestoreStatusByVagaID fails")
		}

		if err.Error() != "restore failed" {
			t.Errorf("Expected 'restore failed', got '%s'", err.Error())
		}
	})

	t.Run("Error when Update fails", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		mockCandidaturaRepo := &MockCandidaturaRepoForVaga{}
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, mockCandidaturaRepo)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID:     vagaID,
			Status: empregabilidade.StatusVagaDescontinuada,
		}
		mockVagaRepo.updateError = errors.New("update failed")

		ctx := context.Background()
		err := service.ReactivateVaga(ctx, vagaID)

		if err == nil {
			t.Error("Expected error when Update fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

// ==================== Delete Tests ====================

func TestVagaService_Delete_Success(t *testing.T) {
	t.Run("Successfully delete vaga", func(t *testing.T) {
		mockVagaRepo := NewMockVagaRepoForService()
		mockEmpresaRepo := NewMockEmpresaRepoForService()
		service := services.NewVagaServiceWithInterfaces(mockVagaRepo, mockEmpresaRepo, nil)

		vagaID := uuid.New()
		mockVagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
			ID: vagaID,
		}

		ctx := context.Background()
		err := service.Delete(ctx, vagaID)

		if err != nil {
			t.Errorf("Expected successful delete, got error: %v", err)
		}

		if _, exists := mockVagaRepo.vagas[vagaID]; exists {
			t.Error("Expected vaga to be deleted")
		}
	})
}

// ==================== GetBySlug Tests ====================

func TestVagaService_GetBySlug_Encontra(t *testing.T) {
	mockVagaRepo := NewMockVagaRepoForService()
	service := services.NewVagaServiceWithInterfaces(mockVagaRepo, NewMockEmpresaRepoForService(), nil)

	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	mockVagaRepo.vagas[id] = &empregabilidade.Vaga{
		ID:     id,
		Titulo: "Analista de TI",
		Status: empregabilidade.StatusVagaPublicadoAtivo,
	}

	vaga, err := service.GetBySlug(context.Background(), "analista-de-ti-f3d23675")
	assert.NoError(t, err)
	assert.NotNil(t, vaga)
	assert.Equal(t, id, vaga.ID)
}

func TestVagaService_GetBySlug_NaoEncontra(t *testing.T) {
	mockVagaRepo := NewMockVagaRepoForService()
	service := services.NewVagaServiceWithInterfaces(mockVagaRepo, NewMockEmpresaRepoForService(), nil)

	vaga, err := service.GetBySlug(context.Background(), "qualquer-titulo-aaaabbbb")
	assert.NoError(t, err)
	assert.Nil(t, vaga)
}

func TestVagaService_GetBySlug_SufixoInvalido(t *testing.T) {
	mockVagaRepo := NewMockVagaRepoForService()
	service := services.NewVagaServiceWithInterfaces(mockVagaRepo, NewMockEmpresaRepoForService(), nil)

	vaga, err := service.GetBySlug(context.Background(), "slug-sem-sufixo-valido")
	assert.NoError(t, err)
	assert.Nil(t, vaga)
}

func TestVagaService_GetBySlug_ErroDoRepo(t *testing.T) {
	mockVagaRepo := NewMockVagaRepoForService()
	mockVagaRepo.getError = errors.New("db error")
	service := services.NewVagaServiceWithInterfaces(mockVagaRepo, NewMockEmpresaRepoForService(), nil)

	_, err := service.GetBySlug(context.Background(), "analista-de-ti-f3d23675")
	assert.Error(t, err)
}

// ==================== Mock Candidatura Repo for Vaga ====================

type MockCandidaturaRepoForVaga struct {
	saveError    error
	restoreError error
}

func (m *MockCandidaturaRepoForVaga) BulkSaveAndUpdateStatusByVagaID(ctx context.Context, vagaID uuid.UUID, status empregabilidade.StatusCandidatura) error {
	return m.saveError
}

func (m *MockCandidaturaRepoForVaga) BulkRestoreStatusByVagaID(ctx context.Context, vagaID uuid.UUID) error {
	return m.restoreError
}
