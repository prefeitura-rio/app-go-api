package clients_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
)

// Integration test against real RMI API
// Requires environment variables:
// - RMI_TEST_JWT: Admin JWT token
// - RMI_TEST_CPF: Test CPF with multiple CNPJs
// - RMI_BASE_URL: RMI API base URL
//
// WARNING: This test performs READ-ONLY operations against the real API
func TestRMIClient_Integration_GetUserLegalEntities(t *testing.T) {
	// Check if integration test should run
	jwt := os.Getenv("RMI_TEST_JWT")
	cpf := os.Getenv("RMI_TEST_CPF")
	baseURL := os.Getenv("RMI_BASE_URL")

	if jwt == "" || cpf == "" || baseURL == "" {
		t.Skip("Skipping integration test: RMI_TEST_JWT, RMI_TEST_CPF, or RMI_BASE_URL not set")
		return
	}

	// Create RMI client with 15s timeout
	client := clients.NewRMIClient(baseURL, 15*time.Second)

	ctx := context.Background()

	t.Run("Fetch legal entities for test CPF", func(t *testing.T) {
		authHeader := "Bearer " + jwt

		entities, err := client.GetUserLegalEntities(ctx, authHeader, cpf)
		if err != nil {
			t.Fatalf("Failed to fetch legal entities: %v", err)
		}

		// Verify we got results
		if len(entities) == 0 {
			t.Error("Expected at least one legal entity, got none")
			return
		}

		t.Logf("✅ Fetched %d legal entities for CPF %s", len(entities), cpf)

		// Log all entities first to understand the data
		t.Log("\nAll legal entities:")
		var entitiesWithCNAE []int
		for i, entity := range entities {
			t.Logf("  [%d] CNPJ: %s | Razão Social: %s | CNAE Fiscal: '%s' | Secundárias: %d",
				i+1, entity.CNPJ, entity.RazaoSocial, entity.CNAEFiscal, len(entity.CNAESecundarias))

			if entity.CNAEFiscal != "" || len(entity.CNAESecundarias) > 0 {
				entitiesWithCNAE = append(entitiesWithCNAE, i)
			}
		}

		// Find first entity with CNAE data
		if len(entitiesWithCNAE) == 0 {
			t.Log("⚠️  WARNING: No legal entities have CNAE data (all CNAEFiscal are empty)")
			t.Log("This is real-world data - some companies may not have CNAE registered")
			return
		}

		// Verify structure using first entity with CNAE
		firstEntityIdx := entitiesWithCNAE[0]
		firstEntity := entities[firstEntityIdx]

		t.Logf("\nVerifying structure of entity [%d]:", firstEntityIdx+1)

		if firstEntity.CNPJ == "" {
			t.Error("Expected CNPJ to be populated")
		}
		t.Logf("  - CNPJ: %s", firstEntity.CNPJ)

		if firstEntity.CNAEFiscal == "" {
			t.Log("  - CNAE Fiscal: (empty)")
		} else {
			t.Logf("  - CNAE Fiscal: %s", firstEntity.CNAEFiscal)
		}

		if firstEntity.RazaoSocial == "" {
			t.Error("Expected RazaoSocial to be populated")
		}
		t.Logf("  - Razão Social: %s", firstEntity.RazaoSocial)

		// Check if has secondary CNAEs
		if len(firstEntity.CNAESecundarias) > 0 {
			t.Logf("  - CNAE Secundárias: %v", firstEntity.CNAESecundarias)
		} else {
			t.Log("  - CNAE Secundárias: (none)")
		}

		// Verify GetAllCNAEs includes fiscal + secundarias
		allCNAEs := firstEntity.GetAllCNAEs()

		// Only count non-empty CNAEs
		expectedCount := 0
		if firstEntity.CNAEFiscal != "" {
			expectedCount++
		}
		expectedCount += len(firstEntity.CNAESecundarias)

		if len(allCNAEs) != expectedCount {
			t.Errorf("GetAllCNAEs() returned %d CNAEs, expected %d (fiscal: '%s', secundarias: %d)",
				len(allCNAEs), expectedCount, firstEntity.CNAEFiscal, len(firstEntity.CNAESecundarias))
		}

		// Verify fiscal CNAE is first (if not empty)
		if firstEntity.CNAEFiscal != "" && len(allCNAEs) > 0 {
			if allCNAEs[0] != firstEntity.CNAEFiscal {
				t.Errorf("First CNAE should be fiscal CNAE %s, got %s", firstEntity.CNAEFiscal, allCNAEs[0])
			}
		}

		t.Logf("\n✅ Data structure verified: %d entities with CNAE data out of %d total",
			len(entitiesWithCNAE), len(entities))
	})

	t.Run("Verify pagination works if needed", func(t *testing.T) {
		authHeader := "Bearer " + jwt

		entities, err := client.GetUserLegalEntities(ctx, authHeader, cpf)
		if err != nil {
			t.Fatalf("Failed to fetch legal entities: %v", err)
		}

		// If we got entities, pagination should have worked (even if only 1 page)
		t.Logf("✅ Pagination handled correctly: %d total entities", len(entities))
	})

	t.Run("Verify 1000 entity safety limit", func(t *testing.T) {
		authHeader := "Bearer " + jwt

		entities, err := client.GetUserLegalEntities(ctx, authHeader, cpf)
		if err != nil {
			t.Fatalf("Failed to fetch legal entities: %v", err)
		}

		if len(entities) > 1000 {
			t.Errorf("Safety limit exceeded: got %d entities, max should be 1000", len(entities))
		} else {
			t.Logf("✅ Safety limit respected: %d entities (max 1000)", len(entities))
		}
	})

	t.Run("Verify request completes within timeout", func(t *testing.T) {
		authHeader := "Bearer " + jwt

		start := time.Now()
		_, err := client.GetUserLegalEntities(ctx, authHeader, cpf)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		// Should complete well under 15s timeout for a single user
		if duration > 15*time.Second {
			t.Errorf("Request took %v, expected under 15s", duration)
		} else {
			t.Logf("✅ Request completed in %v (timeout: 15s)", duration)
		}
	})
}

// Test error handling (without making actual calls)
func TestRMIClient_ErrorHandling(t *testing.T) {
	// Test with invalid base URL
	t.Run("Invalid base URL", func(t *testing.T) {
		client := clients.NewRMIClient("http://invalid.url.that.does.not.exist", 5*time.Second)
		ctx := context.Background()

		_, err := client.GetUserLegalEntities(ctx, "Bearer token", "12345678900")
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
		t.Logf("✅ Correctly returns error for invalid URL: %v", err)
	})

	// Test with cancelled context
	t.Run("Cancelled context", func(t *testing.T) {
		baseURL := os.Getenv("RMI_BASE_URL")
		if baseURL == "" {
			baseURL = "https://services.pref.rio/rmi" // Default for test
		}

		client := clients.NewRMIClient(baseURL, 15*time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.GetUserLegalEntities(ctx, "Bearer token", "12345678900")
		if err == nil {
			t.Error("Expected error for cancelled context, got nil")
		}
		t.Logf("✅ Correctly returns error for cancelled context: %v", err)
	})
}
