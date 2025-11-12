package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// Mock RMI Client
type MockRMIClient struct {
	entities []models.LegalEntity
	err      error
}

func (m *MockRMIClient) GetUserLegalEntities(ctx context.Context, authToken string, cpf string) ([]models.LegalEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.entities, nil
}

// Mock Cache
type MockCache struct {
	data map[string][]models.LegalEntity
	err  error
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]models.LegalEntity),
	}
}

func (m *MockCache) Get(ctx context.Context, cpf string) ([]models.LegalEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	entities, exists := m.data[cpf]
	if !exists {
		return nil, nil // Cache miss
	}
	return entities, nil
}

func (m *MockCache) Set(ctx context.Context, cpf string, entities []models.LegalEntity) error {
	if m.err != nil {
		return m.err
	}
	m.data[cpf] = entities
	return nil
}


func TestCNAEValidationService_ValidatePropostaForCNAE_Success(t *testing.T) {
	// Valid JWT token with CPF in preferred_username
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("Successful validation - exact CNAE match", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"4110700"},
		)

		if err != nil {
			t.Errorf("Expected successful validation, got error: %v", err)
		}

		// Verify cache was populated
		cached, _ := mockCache.Get(ctx, "12345678900")
		if len(cached) != 1 {
			t.Errorf("Expected 1 entity in cache, got %d", len(cached))
		}
	})

	t.Run("Successful validation - CNAE in secundarias", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{"6201500", "6202300"},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"6201500"}, // Match secundaria
		)

		if err != nil {
			t.Errorf("Expected successful validation with secundaria CNAE, got error: %v", err)
		}
	})

	t.Run("Successful validation - formatted CNAE matching", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12.345.678/0001-90",
					CNAEFiscal:      "4110-7/00", // Formatted
					CNAESecundarias: []string{},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190", // No formatting
			[]string{"4110700"},
		)

		if err != nil {
			t.Errorf("Expected successful validation with formatted CNAE, got error: %v", err)
		}
	})

	t.Run("Populates cache after first call", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// Cache should be empty initially
		cached, _ := mockCache.Get(ctx, "12345678900")
		if cached != nil {
			t.Error("Expected cache to be empty initially")
		}

		// Make first call
		_ = service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		// Cache should now be populated
		cached, _ = mockCache.Get(ctx, "12345678900")
		if len(cached) != 1 {
			t.Errorf("Expected 1 entity in cache after first call, got %d", len(cached))
		}
		if len(cached) > 0 && cached[0].CNPJ != "12345678000190" {
			t.Errorf("Expected cached CNPJ to be 12345678000190, got %s", cached[0].CNPJ)
		}
	})
}

func TestCNAEValidationService_ValidatePropostaForCNAE_Failures(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("CNPJ not belonging to user", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "98765432000199", // Different CNPJ
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Other Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190", // User trying to use different CNPJ
			[]string{"4110700"},
		)

		if err == nil {
			t.Error("Expected error for CNPJ not belonging to user, got nil")
		}

		expectedMsg := "o CNPJ 12345678000190 não pertence ao seu CPF"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("CNAE mismatch", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700", // Different CNAE
					CNAESecundarias: []string{},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"6201500", "6202300"}, // Different CNAEs
		)

		if err == nil {
			t.Error("Expected error for CNAE mismatch, got nil")
		}

		if err.Error() != "o CNPJ 12345678000190 não possui CNAE compatível com esta oportunidade. CNAEs aceitos: 6201500, 6202300" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("User has no CNPJs", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{}, // Empty list
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"4110700"},
		)

		if err == nil {
			t.Error("Expected error for no CNPJs, got nil")
		}

		expectedMsg := "nenhum CNPJ encontrado vinculado ao seu CPF"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Invalid JWT token", func(t *testing.T) {
		mockRMI := &MockRMIClient{}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			"invalid-token",
			"12345678000190",
			[]string{"4110700"},
		)

		if err == nil {
			t.Error("Expected error for invalid token, got nil")
		}

		if err.Error() != "não foi possível extrair CPF do token de autenticação" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("RMI API error", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			err: errors.New("RMI API connection failed"),
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"4110700"},
		)

		if err == nil {
			t.Error("Expected error for RMI API failure, got nil")
		}

		expectedMsg := "não foi possível validar seus CNPJs no momento. Tente novamente em instantes"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Cache error does not block validation", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()
		mockCache.err = errors.New("Redis connection failed")

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"4110700"},
		)

		// Should still succeed despite cache error
		if err != nil {
			t.Errorf("Expected validation to succeed despite cache error, got: %v", err)
		}
	})
}

func TestCNAEValidationService_EdgeCases(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("Multiple CNPJs - only one matches", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "11111111000111",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company 1",
				},
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "6201500",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company 2",
				},
				{
					CNPJ:            "33333333000133",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company 3",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"6201500"},
		)

		if err != nil {
			t.Errorf("Expected successful validation, got error: %v", err)
		}
	})

	t.Run("Empty CNAE fiscal but has secundaria matching", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "", // Empty fiscal
					CNAESecundarias: []string{"6201500", "6202300"},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"6201500"},
		)

		if err != nil {
			t.Errorf("Expected successful validation with secundaria only, got error: %v", err)
		}
	})

	t.Run("Multiple opportunity CNAEs - one matches", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"6201500", "4110700", "8888888"}, // Second one matches
		)

		if err != nil {
			t.Errorf("Expected successful validation with one matching CNAE, got error: %v", err)
		}
	})

	t.Run("CNPJ normalization - formatted vs raw", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12.345.678/0001-90", // Formatted
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Test Company LTDA",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190", // Raw
			[]string{"4110700"},
		)

		if err != nil {
			t.Errorf("Expected successful validation with CNPJ normalization, got error: %v", err)
		}
	})
}
