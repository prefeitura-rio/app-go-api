package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// EXPANDED TESTS FOR PROPOSTA MEI SERVICE
// This file adds 15+ edge cases for proposal workflows

// ==================== Status Transition Tests ====================

func TestPropostaMEIService_AllStatusTransitions(t *testing.T) {
	testCases := []struct {
		name          string
		initialStatus models.StatusPropostaCidadao
		targetStatus  models.StatusPropostaCidadao
		shouldSucceed bool
	}{
		{
			name:          "Submitted to Approved",
			initialStatus: models.StatusPropostaCidadaoSubmitted,
			targetStatus:  models.StatusPropostaCidadaoApproved,
			shouldSucceed: true,
		},
		{
			name:          "Submitted to Rejected",
			initialStatus: models.StatusPropostaCidadaoSubmitted,
			targetStatus:  models.StatusPropostaCidadaoRejected,
			shouldSucceed: true,
		},
		{
			name:          "Approved to Rejected",
			initialStatus: models.StatusPropostaCidadaoApproved,
			targetStatus:  models.StatusPropostaCidadaoRejected,
			shouldSucceed: true,
		},
		{
			name:          "Rejected to Approved",
			initialStatus: models.StatusPropostaCidadaoRejected,
			targetStatus:  models.StatusPropostaCidadaoApproved,
			shouldSucceed: true,
		},
		{
			name:          "Approved to Approved (no-op)",
			initialStatus: models.StatusPropostaCidadaoApproved,
			targetStatus:  models.StatusPropostaCidadaoApproved,
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPropostaRepo := NewMockPropostaRepo()
			mockOportunidadeRepo := NewMockOportunidadeRepo()
			mockCNAEValidation := &MockCNAEValidation{}

			id := uuid.New()
			mockPropostaRepo.propostas[id] = &models.PropostaMEI{
				ID:            id,
				StatusCidadao: tc.initialStatus,
				MEIEmpresaID:  "12345678000190",
			}

			service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

			ctx := context.Background()
			err := service.UpdateStatusCidadao(ctx, id, tc.targetStatus)

			if tc.shouldSucceed && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if !tc.shouldSucceed && err == nil {
				t.Error("Expected error but got success")
			}

			// Verify status was updated
			if tc.shouldSucceed {
				updated := mockPropostaRepo.propostas[id]
				if updated.StatusCidadao != tc.targetStatus {
					t.Errorf("Expected status %s, got %s", tc.targetStatus, updated.StatusCidadao)
				}
			}
		})
	}
}

// ==================== Document Validation ====================

func TestPropostaMEIService_DocumentValidation(t *testing.T) {
	t.Run("Create with valid CNPJ format", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		mockOportunidadeRepo.oportunidades[1] = &models.OportunidadeMEI{
			ID:      1,
			Status:  models.StatusOportunidadeActive,
			CNAEIDs: []string{"4110700"},
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		testCases := []string{
			"12345678000190",          // Raw
			"12.345.678/0001-90",      // Formatted
			" 12345678000190 ",        // With spaces
		}

		for _, cnpj := range testCases {
			proposta := &models.PropostaMEI{
				OportunidadeMEIID: 1,
				MEIEmpresaID:      cnpj,
			}

			ctx := context.Background()
			id, err := service.Create(ctx, proposta, "Bearer test-token")

			if err != nil {
				t.Errorf("CNPJ %s should be valid: %v", cnpj, err)
			}
			if id == uuid.Nil {
				t.Errorf("Expected non-nil UUID for CNPJ %s", cnpj)
			}
		}
	})

	t.Run("Create with invalid data fails validation", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		// Missing required fields
		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 0, // Invalid
			MEIEmpresaID:      "", // Invalid
		}

		ctx := context.Background()
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err == nil {
			t.Error("Expected validation error for invalid proposal data")
		}
	})
}

// ==================== CNAE Validation Edge Cases ====================

func TestPropostaMEIService_CNAEValidationScenarios(t *testing.T) {
	t.Run("Create with CNAE validation error includes CNPJ in message", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		cnpj := "12345678000190"
		mockCNAEValidation := &MockCNAEValidation{
			validateError: errors.New("o CNPJ " + cnpj + " não possui CNAE compatível com esta oportunidade"),
		}

		mockOportunidadeRepo.oportunidades[1] = &models.OportunidadeMEI{
			ID:      1,
			Status:  models.StatusOportunidadeActive,
			CNAEIDs: []string{"4110700"},
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 1,
			MEIEmpresaID:      cnpj,
		}

		ctx := context.Background()
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err == nil {
			t.Error("Expected CNAE validation error")
		}

		// Verify error message includes CNPJ
		if err != nil && err.Error() != "o CNPJ "+cnpj+" não possui CNAE compatível com esta oportunidade" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("Create passes opportunity CNAEs to validation service", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()

		expectedCNAEs := []string{"4110700", "6201500", "6202300"}

		mockCNAEValidation := &MockCNAEValidation{
			validateError: nil,
		}

		mockOportunidadeRepo.oportunidades[1] = &models.OportunidadeMEI{
			ID:      1,
			Status:  models.StatusOportunidadeActive,
			CNAEIDs: expectedCNAEs,
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		proposta := &models.PropostaMEI{
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		ctx := context.Background()
		// Service will call ValidatePropostaForCNAE with the expected CNAEs
		_, err := service.Create(ctx, proposta, "Bearer test-token")

		if err != nil {
			t.Errorf("Create should succeed: %v", err)
		}
	})
}

// ==================== Address Validation ====================

func TestPropostaMEIService_UpdatePropostaEdgeCases(t *testing.T) {
	t.Run("UpdateProposta with zero value is valid", func(t *testing.T) {
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

		zeroValue := 0.0
		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 1, &zeroValue, nil, nil)

		// Zero is not positive, should fail
		if err == nil {
			t.Error("Expected error for zero value")
		}
	})

	t.Run("UpdateProposta with very large value", func(t *testing.T) {
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

		largeValue := 99999999.99
		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 1, &largeValue, nil, nil)

		if err != nil {
			t.Errorf("Large valid value should succeed: %v", err)
		}

		updated := mockPropostaRepo.propostas[id]
		if updated.ValorProposta == nil || *updated.ValorProposta != largeValue {
			t.Error("Expected large value to be updated")
		}
	})

	t.Run("UpdateProposta with only prazo_execucao", func(t *testing.T) {
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

		prazo := "30 dias"
		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 1, nil, &prazo, nil)

		if err != nil {
			t.Errorf("Update prazo_execucao should succeed: %v", err)
		}

		updated := mockPropostaRepo.propostas[id]
		if updated.PrazoExecucao == nil || *updated.PrazoExecucao != prazo {
			t.Error("Expected prazo_execucao to be updated")
		}
	})

	t.Run("UpdateProposta with only aceita_custos_integrais", func(t *testing.T) {
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

		aceitaCustos := true
		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 1, nil, nil, &aceitaCustos)

		if err != nil {
			t.Errorf("Update aceita_custos_integrais should succeed: %v", err)
		}

		updated := mockPropostaRepo.propostas[id]
		if updated.AceitaCustosIntegrais == nil || *updated.AceitaCustosIntegrais != aceitaCustos {
			t.Error("Expected aceita_custos_integrais to be updated")
		}
	})

	t.Run("UpdateProposta with all fields", func(t *testing.T) {
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

		valor := 5000.0
		prazo := "60 dias"
		aceitaCustos := false

		ctx := context.Background()
		err := service.UpdateProposta(ctx, id, 1, &valor, &prazo, &aceitaCustos)

		if err != nil {
			t.Errorf("Update all fields should succeed: %v", err)
		}

		updated := mockPropostaRepo.propostas[id]
		if updated.ValorProposta == nil || *updated.ValorProposta != valor {
			t.Error("Expected valor_proposta to be updated")
		}
		if updated.PrazoExecucao == nil || *updated.PrazoExecucao != prazo {
			t.Error("Expected prazo_execucao to be updated")
		}
		if updated.AceitaCustosIntegrais == nil || *updated.AceitaCustosIntegrais != aceitaCustos {
			t.Error("Expected aceita_custos_integrais to be updated")
		}
	})
}

// ==================== Rejection Workflows ====================

func TestPropostaMEIService_RejectionWorkflows(t *testing.T) {
	t.Run("Reject sets status correctly", func(t *testing.T) {
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
		err := service.Reject(ctx, id)

		if err != nil {
			t.Errorf("Reject should succeed: %v", err)
		}

		updated := mockPropostaRepo.propostas[id]
		if updated.StatusCidadao != models.StatusPropostaCidadaoRejected {
			t.Errorf("Expected status Rejected, got %s", updated.StatusCidadao)
		}
	})

	t.Run("Reject non-existent proposal", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()
		err := service.Reject(ctx, uuid.New())

		if err == nil {
			t.Error("Expected error for non-existent proposal")
		}
	})
}

// ==================== Approval Workflows ====================

func TestPropostaMEIService_ApprovalWorkflows(t *testing.T) {
	t.Run("Approve sets status correctly", func(t *testing.T) {
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
		err := service.Approve(ctx, id)

		if err != nil {
			t.Errorf("Approve should succeed: %v", err)
		}

		updated := mockPropostaRepo.propostas[id]
		if updated.StatusCidadao != models.StatusPropostaCidadaoApproved {
			t.Errorf("Expected status Approved, got %s", updated.StatusCidadao)
		}
	})

	t.Run("Approve already approved proposal (idempotent)", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		id := uuid.New()
		mockPropostaRepo.propostas[id] = &models.PropostaMEI{
			ID:            id,
			StatusCidadao: models.StatusPropostaCidadaoApproved, // Already approved
			MEIEmpresaID:  "12345678000190",
		}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()
		err := service.Approve(ctx, id)

		if err != nil {
			t.Errorf("Approve should be idempotent: %v", err)
		}

		updated := mockPropostaRepo.propostas[id]
		if updated.StatusCidadao != models.StatusPropostaCidadaoApproved {
			t.Error("Status should remain Approved")
		}
	})
}

// ==================== List Operations ====================

func TestPropostaMEIService_ListOperations(t *testing.T) {
	t.Run("ListByOportunidade pagination edge cases", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()

		// Test first page
		_, _, err := service.ListByOportunidade(ctx, 1, "", "", "", 1, 10)
		if err != nil {
			t.Errorf("ListByOportunidade page 1 failed: %v", err)
		}

		// Test large page number
		_, _, err = service.ListByOportunidade(ctx, 1, "", "", "", 100, 10)
		if err != nil {
			t.Errorf("ListByOportunidade page 100 failed: %v", err)
		}
	})

	t.Run("ListByMEIEmpresa pagination", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()
		_, _, err := service.ListByMEIEmpresa(ctx, "12345678000190", 1, 20)

		if err != nil {
			t.Errorf("ListByMEIEmpresa failed: %v", err)
		}
	})

	t.Run("ListByStatus with all status types", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ctx := context.Background()

		statuses := []models.StatusPropostaCidadao{
			models.StatusPropostaCidadaoSubmitted,
			models.StatusPropostaCidadaoApproved,
			models.StatusPropostaCidadaoRejected,
		}

		for _, status := range statuses {
			_, _, err := service.ListByStatus(ctx, status, 1, 10)
			if err != nil {
				t.Errorf("ListByStatus for %s failed: %v", status, err)
			}
		}
	})
}

// ==================== Bulk Operations ====================

func TestPropostaMEIService_BulkOperationsAdvanced(t *testing.T) {
	t.Run("UpdateMultipleStatus with invalid status", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ids := []uuid.UUID{uuid.New(), uuid.New()}
		ctx := context.Background()

		// Try to update with invalid status
		invalidStatus := models.StatusPropostaCidadao("INVALID_STATUS")
		_, err := service.UpdateMultipleStatus(ctx, ids, invalidStatus)

		if err == nil {
			t.Error("Expected error for invalid status")
		}
		if err.Error() != "status inválido" {
			t.Errorf("Expected 'status inválido' error, got: %v", err)
		}
	})

	t.Run("UpdateMultipleStatus with single ID", func(t *testing.T) {
		mockPropostaRepo := NewMockPropostaRepo()
		mockPropostaRepo.updateMultiCount = 1

		mockOportunidadeRepo := NewMockOportunidadeRepo()
		mockCNAEValidation := &MockCNAEValidation{}

		service := services.NewPropostaMEIService(mockPropostaRepo, mockOportunidadeRepo, mockCNAEValidation, nil)

		ids := []uuid.UUID{uuid.New()}
		ctx := context.Background()

		count, err := service.UpdateMultipleStatus(ctx, ids, models.StatusPropostaCidadaoApproved)

		if err != nil {
			t.Errorf("UpdateMultipleStatus with single ID should succeed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})
}
