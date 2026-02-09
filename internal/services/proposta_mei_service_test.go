package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// Mock PropostaMEI Repository
type MockPropostaRepo struct {
	propostas        map[uuid.UUID]*models.PropostaMEI
	existingCheck    bool
	createError      error
	getError         error
	updateError      error
	deleteError      error
	checkError       error
	updateMultiCount int
}

func NewMockPropostaRepo() *MockPropostaRepo {
	return &MockPropostaRepo{
		propostas: make(map[uuid.UUID]*models.PropostaMEI),
	}
}

func (m *MockPropostaRepo) Create(ctx context.Context, proposta *models.PropostaMEI) (uuid.UUID, error) {
	if m.createError != nil {
		return uuid.Nil, m.createError
	}
	id := uuid.New()
	proposta.ID = id
	m.propostas[id] = proposta
	return id, nil
}

func (m *MockPropostaRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	proposta, exists := m.propostas[id]
	if !exists {
		return nil, nil
	}
	return proposta, nil
}

func (m *MockPropostaRepo) Update(ctx context.Context, proposta *models.PropostaMEI) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.propostas[proposta.ID] = proposta
	return nil
}

func (m *MockPropostaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.propostas, id)
	return nil
}

func (m *MockPropostaRepo) CheckExistingProposta(ctx context.Context, oportunidadeID int, meiEmpresaID string) (bool, error) {
	if m.checkError != nil {
		return false, m.checkError
	}
	return m.existingCheck, nil
}

func (m *MockPropostaRepo) ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, limit, offset int) ([]*models.PropostaMEI, int, error) {
	return []*models.PropostaMEI{}, 0, nil
}

func (m *MockPropostaRepo) ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, limit, offset int) ([]*models.PropostaMEI, int, error) {
	return []*models.PropostaMEI{}, 0, nil
}

func (m *MockPropostaRepo) ListByStatus(ctx context.Context, status models.StatusPropostaCidadao, limit, offset int) ([]*models.PropostaMEI, int, error) {
	return []*models.PropostaMEI{}, 0, nil
}

func (m *MockPropostaRepo) UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error) {
	if m.updateError != nil {
		return 0, m.updateError
	}
	return m.updateMultiCount, nil
}

// Mock OportunidadeMEI Repository
type MockOportunidadeRepo struct {
	oportunidades map[int]*models.OportunidadeMEI
	getError      error
}

func NewMockOportunidadeRepo() *MockOportunidadeRepo {
	return &MockOportunidadeRepo{
		oportunidades: make(map[int]*models.OportunidadeMEI),
	}
}

func (m *MockOportunidadeRepo) Create(ctx context.Context, oportunidade *models.OportunidadeMEI) (int, error) {
	return 0, nil
}

func (m *MockOportunidadeRepo) GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	oportunidade, exists := m.oportunidades[id]
	if !exists {
		return nil, nil
	}
	return oportunidade, nil
}

func (m *MockOportunidadeRepo) Update(ctx context.Context, oportunidade *models.OportunidadeMEI) error {
	return nil
}

func (m *MockOportunidadeRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *MockOportunidadeRepo) List(ctx context.Context, filters map[string]interface{}, titulo string, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	return nil, 0, nil
}

func (m *MockOportunidadeRepo) ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	return nil, 0, nil
}

func (m *MockOportunidadeRepo) ListByOrgao(ctx context.Context, orgaoID string, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	return nil, 0, nil
}

func (m *MockOportunidadeRepo) UpdateExpiredOpportunities(ctx context.Context) error {
	return nil
}

// Mock CNAE Validation Service
type MockCNAEValidation struct {
	validateError        error
	checkOwnershipResult bool
	checkOwnershipError  error
}

func (m *MockCNAEValidation) ValidatePropostaForCNAE(ctx context.Context, authToken string, cnpj string, opportunityCNAEIDs []string) error {
	return m.validateError
}

func (m *MockCNAEValidation) CheckCNPJOwnership(ctx context.Context, authToken string, cpf string, cnpj string) (bool, error) {
	return m.checkOwnershipResult, m.checkOwnershipError
}

func TestPropostaMEIService_Create_Success(t *testing.T) {
	t.Run("Successful proposal creation", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		// Setup active opportunity
		mockOportunidadeRepo.oportunidades[1] = &models.OportunidadeMEI{
			ID:      1,
			Status:  models.StatusOportunidadeActive,
			CNAEIDs: []string{"4110700"},
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		ctx := context.Background()
		id, err := service.Create(ctx, proposta, "Bearer test-token")

		if err != nil {
			t.Errorf("Expected successful creation, got error: %v", err)
		}

		if id == uuid.Nil {
			t.Error("Expected non-nil UUID")
		}

		// Verify status was set correctly
		created := mockPropostaRepo.propostas[id]
		if created.StatusAdmin != models.StatusPropostaAdminActive {
			t.Errorf("Expected StatusAdmin to be Active, got %s", created.StatusAdmin)
		}
		if created.StatusCidadao != models.StatusPropostaCidadaoSubmitted {
			t.Errorf("Expected StatusCidadao to be Submitted, got %s", created.StatusCidadao)
		}
	})
}

func TestPropostaMEIService_Create_Failures(t *testing.T) {
	t.Run("Invalid proposal data", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		// Missing required fields
		proposta := &models.PropostaMEI{}

		ctx := context.Background()
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err == nil {
			t.Error("Expected validation error for invalid proposal")
		}
	})

	t.Run("Opportunity not found", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 999, // Non-existent
			MEIEmpresaID:      "12345678000190",
		}

		ctx := context.Background()
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err == nil {
			t.Error("Expected error for non-existent opportunity")
		}

		expectedMsg := "oportunidade não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Opportunity not active", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		// Setup inactive opportunity
		mockOportunidadeRepo.oportunidades[1] = &models.OportunidadeMEI{
			ID:      1,
			Status:  models.StatusOportunidadeDraft,
			CNAEIDs: []string{"4110700"},
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		ctx := context.Background()
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err == nil {
			t.Error("Expected error for inactive opportunity")
		}

		expectedMsg := "oportunidade não está ativa"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("CNAE validation fails", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{
			validateError: errors.New("o CNPJ 12345678000190 não possui CNAE compatível com esta oportunidade"),
		}

		mockOportunidadeRepo.oportunidades[1] = &models.OportunidadeMEI{
			ID:      1,
			Status:  models.StatusOportunidadeActive,
			CNAEIDs: []string{"4110700"},
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		ctx := context.Background()
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err == nil {
			t.Error("Expected CNAE validation error")
		}

		if err.Error() != "o CNPJ 12345678000190 não possui CNAE compatível com esta oportunidade" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("Duplicate proposal", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockPropostaRepo.existingCheck = true // Simulate existing proposal

		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		mockOportunidadeRepo.oportunidades[1] = &models.OportunidadeMEI{
			ID:      1,
			Status:  models.StatusOportunidadeActive,
			CNAEIDs: []string{"4110700"},
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		ctx := context.Background()
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err == nil {
			t.Error("Expected error for duplicate proposal")
		}

		expectedMsg := "já existe uma proposta desta empresa para esta oportunidade"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestPropostaMEIService_UpdateProposta(t *testing.T) {
	t.Run("Update with valid positive value", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		// Create existing proposal
		id := uuid.New()
		mockPropostaRepo.propostas[id] = &models.PropostaMEI{
			ID:                id,
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
			ValorProposta:     nil,
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		newValue := 1500.00
		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 1, &newValue, nil, nil)

		if err != nil {
			t.Errorf("Expected successful update, got error: %v", err)
		}

		// Verify value was updated
		updated := mockPropostaRepo.propostas[id]
		if updated.ValorProposta == nil || *updated.ValorProposta != 1500.00 {
			t.Error("Expected valor_proposta to be updated to 1500.00")
		}
	})

	t.Run("Update with negative value fails", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		id := uuid.New()
		mockPropostaRepo.propostas[id] = &models.PropostaMEI{
			ID:                id,
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		negativeValue := -100.00
		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 1, &negativeValue, nil, nil)

		if err == nil {
			t.Error("Expected error for negative value")
		}

		expectedMsg := "valor_proposta deve ser positivo"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Update non-existent proposal", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		nonExistentID := uuid.New()
		newValue := 1500.00
		ctx := context.Background()
		err := service.UpdateProposta(ctx, nonExistentID, 1, &newValue, nil, nil)

		if err == nil {
			t.Error("Expected error for non-existent proposal")
		}

		expectedMsg := "proposta não encontrada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Update with wrong opportunity ID", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		id := uuid.New()
		mockPropostaRepo.propostas[id] = &models.PropostaMEI{
			ID:                id,
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		newValue := 1500.00
		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 999, &newValue, nil, nil) // Wrong opportunity ID

		if err == nil {
			t.Error("Expected error for wrong opportunity ID")
		}

		expectedMsg := "proposta não pertence à oportunidade especificada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestPropostaMEIService_UpdateStatusCidadao(t *testing.T) {
	t.Run("Update status successfully", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		id := uuid.New()
		mockPropostaRepo.propostas[id] = &models.PropostaMEI{
			ID:            id,
			StatusCidadao: models.StatusPropostaCidadaoSubmitted,
			StatusAdmin:   models.StatusPropostaAdminActive,
			MEIEmpresaID:  "12345678000190",
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()
		err := service.UpdateStatusCidadao(ctx, id, models.StatusPropostaCidadaoApproved)

		if err != nil {
			t.Errorf("Expected successful status update, got error: %v", err)
		}

		// Verify status was updated
		updated := mockPropostaRepo.propostas[id]
		if updated.StatusCidadao != models.StatusPropostaCidadaoApproved {
			t.Errorf("Expected status to be Approved, got %s", updated.StatusCidadao)
		}
	})

	t.Run("Approve and Reject helper methods", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		id := uuid.New()
		mockPropostaRepo.propostas[id] = &models.PropostaMEI{
			ID:            id,
			StatusCidadao: models.StatusPropostaCidadaoSubmitted,
			MEIEmpresaID:  "12345678000190",
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()

		// Test Approve
		err := service.Approve(ctx, id)
		if err != nil {
			t.Errorf("Expected successful approval, got error: %v", err)
		}
		if mockPropostaRepo.propostas[id].StatusCidadao != models.StatusPropostaCidadaoApproved {
			t.Error("Expected status to be Approved")
		}

		// Test Reject
		err = service.Reject(ctx, id)
		if err != nil {
			t.Errorf("Expected successful rejection, got error: %v", err)
		}
		if mockPropostaRepo.propostas[id].StatusCidadao != models.StatusPropostaCidadaoRejected {
			t.Error("Expected status to be Rejected")
		}
	})
}

func TestPropostaMEIService_UpdateMultipleStatus(t *testing.T) {
	t.Run("Update multiple propostas successfully", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockPropostaRepo.updateMultiCount = 5 // Simulate 5 updates

		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
		ctx := context.Background()

		count, err := service.UpdateMultipleStatus(ctx, ids, models.StatusPropostaCidadaoApproved)

		if err != nil {
			t.Errorf("Expected successful bulk update, got error: %v", err)
		}

		if count != 5 {
			t.Errorf("Expected 5 updates, got %d", count)
		}
	})

	t.Run("Empty proposal list", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()
		_, err := service.UpdateMultipleStatus(ctx, []uuid.UUID{}, models.StatusPropostaCidadaoApproved)

		if err == nil {
			t.Error("Expected error for empty proposal list")
		}

		expectedMsg := "nenhuma proposta selecionada"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestPropostaMEIService_Delete(t *testing.T) {
	t.Run("Delete proposal successfully", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		id := uuid.New()
		mockPropostaRepo.propostas[id] = &models.PropostaMEI{
			ID:           id,
			MEIEmpresaID: "12345678000190",
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()
		err := service.Delete(ctx, id)

		if err != nil {
			t.Errorf("Expected successful deletion, got error: %v", err)
		}

		// Verify proposal was deleted
		if _, exists := mockPropostaRepo.propostas[id]; exists {
			t.Error("Expected proposal to be deleted")
		}
	})
}
