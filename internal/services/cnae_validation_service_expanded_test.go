package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// EXPANDED TESTS FOR CNAE VALIDATION SERVICE
// This file adds 15+ edge cases for CNAE validation

// ==================== Complex Ownership Scenarios ====================

func TestCNAEValidationService_ComplexOwnershipScenarios(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("User owns multiple CNPJs with different CNAEs", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "11111111000111",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company 1",
				},
				{
					CNPJ:            "22222222000122",
					CNAEFiscal:      "6201500",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company 2",
				},
				{
					CNPJ:            "33333333000133",
					CNAEFiscal:      "6202300",
					CNAESecundarias: []string{"4110700"},
					RazaoSocial:     "Company 3",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// Validate each CNPJ with its CNAE
		testCases := []struct {
			cnpj      string
			cnaeIDs   []string
			shouldFail bool
		}{
			{"11111111000111", []string{"4110700"}, false},
			{"22222222000122", []string{"6201500"}, false},
			{"33333333000133", []string{"6202300"}, false},
			{"33333333000133", []string{"4110700"}, false}, // In secundarias
			{"11111111000111", []string{"6201500"}, true},  // Wrong CNAE
		}

		for _, tc := range testCases {
			err := service.ValidatePropostaForCNAE(ctx, validToken, tc.cnpj, tc.cnaeIDs)
			if tc.shouldFail && err == nil {
				t.Errorf("Expected error for CNPJ %s with CNAE %v", tc.cnpj, tc.cnaeIDs)
			}
			if !tc.shouldFail && err != nil {
				t.Errorf("Expected success for CNPJ %s with CNAE %v, got: %v", tc.cnpj, tc.cnaeIDs, err)
			}
		}
	})

	t.Run("CNPJ with multiple secundarias matching", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{"6201500", "6202300", "6203100"},
					RazaoSocial:     "Multi-CNAE Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// All secundarias should be valid
		testCases := []string{"6201500", "6202300", "6203100"}

		for _, cnaeID := range testCases {
			err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{cnaeID})
			if err != nil {
				t.Errorf("CNAE %s should be valid: %v", cnaeID, err)
			}
		}
	})

	t.Run("CNPJ fiscal matches one of multiple opportunity CNAEs", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "6201500",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// Opportunity accepts multiple CNAEs
		err := service.ValidatePropostaForCNAE(
			ctx,
			validToken,
			"12345678000190",
			[]string{"4110700", "6201500", "6202300"}, // Second one matches
		)

		if err != nil {
			t.Errorf("Expected success when fiscal matches one of multiple CNAEs: %v", err)
		}
	})
}

// ==================== Multiple CNPJ Checking ====================

func TestCNAEValidationService_MultipleCNPJChecking(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("CheckCNPJOwnership with multiple CNPJs returns correct result", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{CNPJ: "11111111000111", CNAEFiscal: "4110700", RazaoSocial: "Company 1"},
				{CNPJ: "22222222000122", CNAEFiscal: "6201500", RazaoSocial: "Company 2"},
				{CNPJ: "33333333000133", CNAEFiscal: "6202300", RazaoSocial: "Company 3"},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		testCases := []struct {
			cnpj     string
			expected bool
		}{
			{"11111111000111", true},
			{"22222222000122", true},
			{"33333333000133", true},
			{"99999999000199", false},
		}

		for _, tc := range testCases {
			isOwner, err := service.CheckCNPJOwnership(ctx, validToken, "12345678900", tc.cnpj)
			if err != nil {
				t.Errorf("CheckCNPJOwnership failed: %v", err)
			}
			if isOwner != tc.expected {
				t.Errorf("CNPJ %s: expected %v, got %v", tc.cnpj, tc.expected, isOwner)
			}
		}
	})

	t.Run("CheckCNPJOwnership with formatted CNPJ", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:        "12.345.678/0001-90", // Formatted
					CNAEFiscal:  "4110700",
					RazaoSocial: "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		testCases := []string{
			"12345678000190",       // Raw
			"12.345.678/0001-90",   // Formatted
			" 12345678000190 ",     // With spaces
		}

		for _, cnpj := range testCases {
			isOwner, err := service.CheckCNPJOwnership(ctx, validToken, "12345678900", cnpj)
			if err != nil {
				t.Errorf("CheckCNPJOwnership failed for %s: %v", cnpj, err)
			}
			if !isOwner {
				t.Errorf("CNPJ %s should be recognized as owned", cnpj)
			}
		}
	})
}

// ==================== RMI API Error Handling ====================

func TestCNAEValidationService_RMIErrorHandling(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("RMI timeout error", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			err: errors.New("RMI timeout: request took too long"),
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err == nil {
			t.Error("Expected error from RMI timeout")
		}

		expectedMsg := "não foi possível validar seus CNPJs no momento. Tente novamente em instantes"
		if err.Error() != expectedMsg {
			t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("RMI returns empty result", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{}, // Empty
			err:      nil,
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err == nil {
			t.Error("Expected error for no CNPJs")
		}

		expectedMsg := "nenhum CNPJ encontrado vinculado ao seu CPF"
		if err.Error() != expectedMsg {
			t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("RMI connection refused", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			err: errors.New("connection refused"),
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()
		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err == nil {
			t.Error("Expected error from connection refused")
		}
	})
}

// ==================== Cache Invalidation Scenarios ====================

func TestCNAEValidationService_CacheScenarios(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("Cache miss then cache hit", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}

		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// First call - cache miss, populates cache
		_ = service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		// Verify cache was populated
		cached, _ := mockCache.Get(ctx, "12345678900")
		if len(cached) != 1 {
			t.Errorf("Expected cache to be populated with 1 entity, got %d", len(cached))
		}

		// Second call - should use cache
		_ = service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})
	})

	t.Run("Cache Get error falls back to RMI", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()
		mockCache.err = errors.New("cache get error")

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// Should still succeed by falling back to RMI
		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err != nil {
			t.Errorf("Expected validation to succeed despite cache error: %v", err)
		}
	})

	t.Run("Cache Set error does not prevent validation", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()
		mockCache.err = nil // No error on Get
		// Create new mock with Set error
		mockCacheWithSetError := &MockCache{
			data: mockCache.data,
			err:  errors.New("cache set error"),
		}

		service := services.NewCNAEValidationService(mockRMI, mockCacheWithSetError)

		ctx := context.Background()

		// Should succeed despite cache set error
		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err != nil {
			t.Errorf("Expected validation to succeed despite cache set error: %v", err)
		}
	})
}

// ==================== Concurrent Validation Requests ====================

func TestCNAEValidationService_ConcurrentRequests(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("Multiple concurrent validations for same CPF", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// Simulate concurrent requests
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})
				if err != nil {
					t.Errorf("Concurrent validation failed: %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// ==================== Edge Cases Extended (Invalid CNPJ format, etc.) ====================

func TestCNAEValidationService_EdgeCasesExtended(t *testing.T) {
	validToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

	t.Run("CNPJ with special characters gets normalized", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12-345-678/0001.90", // Unusual formatting
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err != nil {
			t.Errorf("CNPJ normalization should handle special characters: %v", err)
		}
	})

	t.Run("Empty opportunity CNAEs list", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "4110700",
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// Empty CNAE list - nothing to validate
		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{})

		// Should fail - no CNAEs to match
		if err == nil {
			t.Error("Expected error for empty opportunity CNAE list")
		}
	})

	t.Run("CNAE with leading/trailing spaces", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      " 4110700 ", // With spaces
					CNAESecundarias: []string{},
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err != nil {
			t.Errorf("CNAE normalization should handle spaces: %v", err)
		}
	})

	t.Run("Very long CNAE list", func(t *testing.T) {
		// Create entity with many secundarias
		secundarias := make([]string, 100)
		for i := 0; i < 100; i++ {
			secundarias[i] = "620" + string(rune('0'+i%10)) + "500"
		}
		secundarias[99] = "4110700" // Last one matches

		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:            "12345678000190",
					CNAEFiscal:      "1111111",
					CNAESecundarias: secundarias,
					RazaoSocial:     "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		err := service.ValidatePropostaForCNAE(ctx, validToken, "12345678000190", []string{"4110700"})

		if err != nil {
			t.Errorf("Should find CNAE in long secundarias list: %v", err)
		}
	})

	t.Run("CheckCNPJOwnership with empty CNPJ", func(t *testing.T) {
		mockRMI := &MockRMIClient{
			entities: []models.LegalEntity{
				{
					CNPJ:        "12345678000190",
					CNAEFiscal:  "4110700",
					RazaoSocial: "Company",
				},
			},
		}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		isOwner, err := service.CheckCNPJOwnership(ctx, validToken, "12345678900", "")

		if err != nil {
			t.Errorf("CheckCNPJOwnership should handle empty CNPJ: %v", err)
		}
		if isOwner {
			t.Error("Empty CNPJ should not be owned")
		}
	})
}

// ==================== Token Validation ====================

func TestCNAEValidationService_TokenValidation(t *testing.T) {
	t.Run("Token without Bearer prefix", func(t *testing.T) {
		mockRMI := &MockRMIClient{}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		// Token without "Bearer " prefix
		invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"

		err := service.ValidatePropostaForCNAE(ctx, invalidToken, "12345678000190", []string{"4110700"})

		if err == nil {
			t.Error("Expected error for token without Bearer prefix")
		}
	})

	t.Run("Empty token", func(t *testing.T) {
		mockRMI := &MockRMIClient{}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		err := service.ValidatePropostaForCNAE(ctx, "", "12345678000190", []string{"4110700"})

		if err == nil {
			t.Error("Expected error for empty token")
		}
	})

	t.Run("Malformed JWT token", func(t *testing.T) {
		mockRMI := &MockRMIClient{}
		mockCache := NewMockCache()

		service := services.NewCNAEValidationService(mockRMI, mockCache)

		ctx := context.Background()

		err := service.ValidatePropostaForCNAE(ctx, "Bearer malformed.jwt.token", "12345678000190", []string{"4110700"})

		if err == nil {
			t.Error("Expected error for malformed JWT")
		}
	})
}
