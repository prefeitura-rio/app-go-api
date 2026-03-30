package cache_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/redis/go-redis/v9"
)

func setupLegalEntitiesCacheTest(t *testing.T) (*redis.Client, context.Context, func()) {
	// Create a miniredis instance for testing
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ctx := context.Background()

	cleanup := func() {
		redisClient.Close()
		mr.Close()
	}

	return redisClient, ctx, cleanup
}

func TestLegalEntitiesCache_SetAndGet(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	// Create cache with 1 minute TTL for testing
	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Set and Get valid data", func(t *testing.T) {
		cpf := "12345678900"
		entities := []models.LegalEntity{
			{
				CNPJ:            "12345678000190",
				CNAEFiscal:      "6201-5/00",
				CNAESecundarias: []string{"6202-3/00", "6203-1/00"},
				RazaoSocial:     "Test Company LTDA",
			},
			{
				CNPJ:            "98765432000199",
				CNAEFiscal:      "4751-2/01",
				CNAESecundarias: []string{},
				RazaoSocial:     "Another Company ME",
			},
		}

		// Set the cache
		err := c.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Get from cache
		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get from cache: %v", err)
		}

		// Verify cache hit
		if cached == nil {
			t.Fatal("Expected cache hit, got nil")
		}

		// Verify data integrity
		if len(cached) != len(entities) {
			t.Errorf("Expected %d entities, got %d", len(entities), len(cached))
		}

		// Verify first entity
		if cached[0].CNPJ != entities[0].CNPJ {
			t.Errorf("Expected CNPJ %s, got %s", entities[0].CNPJ, cached[0].CNPJ)
		}
		if cached[0].CNAEFiscal != entities[0].CNAEFiscal {
			t.Errorf("Expected CNAEFiscal %s, got %s", entities[0].CNAEFiscal, cached[0].CNAEFiscal)
		}
		if len(cached[0].CNAESecundarias) != len(entities[0].CNAESecundarias) {
			t.Errorf("Expected %d CNAESecundarias, got %d", len(entities[0].CNAESecundarias), len(cached[0].CNAESecundarias))
		}

		// Verify second entity
		if cached[1].CNPJ != entities[1].CNPJ {
			t.Errorf("Expected CNPJ %s, got %s", entities[1].CNPJ, cached[1].CNPJ)
		}
	})

	t.Run("Get non-existent key returns nil", func(t *testing.T) {
		cpf := "00000000000"

		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Expected no error for cache miss, got: %v", err)
		}

		if cached != nil {
			t.Error("Expected nil for cache miss, got data")
		}
	})

	t.Run("Set empty array", func(t *testing.T) {
		cpf := "11111111111"
		entities := []models.LegalEntity{}

		err := c.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set empty array: %v", err)
		}

		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get empty array: %v", err)
		}

		if cached == nil {
			t.Error("Expected empty array, got nil")
		} else if len(cached) != 0 {
			t.Errorf("Expected empty array, got %d elements", len(cached))
		}
	})

	t.Run("Overwrite existing cache", func(t *testing.T) {
		cpf := "22222222222"

		// First set
		entities1 := []models.LegalEntity{
			{CNPJ: "11111111000111", CNAEFiscal: "1111-1/11", RazaoSocial: "First Company"},
		}
		err := c.Set(ctx, cpf, entities1)
		if err != nil {
			t.Fatalf("Failed to set first cache: %v", err)
		}

		// Overwrite with new data
		entities2 := []models.LegalEntity{
			{CNPJ: "22222222000222", CNAEFiscal: "2222-2/22", RazaoSocial: "Second Company"},
			{CNPJ: "33333333000333", CNAEFiscal: "3333-3/33", RazaoSocial: "Third Company"},
		}
		err = c.Set(ctx, cpf, entities2)
		if err != nil {
			t.Fatalf("Failed to overwrite cache: %v", err)
		}

		// Verify new data
		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get overwritten cache: %v", err)
		}

		if len(cached) != 2 {
			t.Errorf("Expected 2 entities after overwrite, got %d", len(cached))
		}
		if cached[0].CNPJ != "22222222000222" {
			t.Errorf("Expected first CNPJ to be updated, got %s", cached[0].CNPJ)
		}
	})

	t.Run("Cache key format", func(t *testing.T) {
		cpf := "99999999999"
		entities := []models.LegalEntity{
			{CNPJ: "99999999000199", CNAEFiscal: "9999-9/99", RazaoSocial: "Test"},
		}

		err := c.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Verify key format directly in Redis
		expectedKey := "legal_entities:" + cpf
		exists, err := redisClient.Exists(ctx, expectedKey).Result()
		if err != nil {
			t.Fatalf("Failed to check key existence: %v", err)
		}

		if exists != 1 {
			t.Errorf("Expected key '%s' to exist in Redis", expectedKey)
		}
	})

	t.Run("TTL is set correctly", func(t *testing.T) {
		cpf := "88888888888"
		entities := []models.LegalEntity{
			{CNPJ: "88888888000188", CNAEFiscal: "8888-8/88", RazaoSocial: "TTL Test"},
		}

		// Create cache with 5 second TTL for this test
		shortTTLCache := cache.NewLegalEntitiesCache(redisClient, 5*time.Second)

		err := shortTTLCache.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Check TTL
		expectedKey := "legal_entities:" + cpf
		ttl, err := redisClient.TTL(ctx, expectedKey).Result()
		if err != nil {
			t.Fatalf("Failed to get TTL: %v", err)
		}

		// TTL should be approximately 5 seconds (allow some variance)
		if ttl < 4*time.Second || ttl > 6*time.Second {
			t.Errorf("Expected TTL around 5s, got %v", ttl)
		}
	})
}

func TestLegalEntitiesCache_GetAllCNAEs(t *testing.T) {
	entity := models.LegalEntity{
		CNPJ:            "12345678000190",
		CNAEFiscal:      "6201-5/00",
		CNAESecundarias: []string{"6202-3/00", "6203-1/00", "6204-0/00"},
		RazaoSocial:     "Test Company",
	}

	cnaes := entity.GetAllCNAEs()

	// Should return fiscal + all secundarias
	expectedCount := 4
	if len(cnaes) != expectedCount {
		t.Errorf("Expected %d CNAEs, got %d", expectedCount, len(cnaes))
	}

	// Verify fiscal CNAE is first
	if cnaes[0] != "6201-5/00" {
		t.Errorf("Expected fiscal CNAE first, got %s", cnaes[0])
	}

	// Verify secundarias are included
	if cnaes[1] != "6202-3/00" {
		t.Errorf("Expected first secundaria to be 6202-3/00, got %s", cnaes[1])
	}
}

func TestLegalEntitiesCache_GetAllCNAEs_NoSecundarias(t *testing.T) {
	entity := models.LegalEntity{
		CNPJ:            "12345678000190",
		CNAEFiscal:      "6201-5/00",
		CNAESecundarias: []string{},
		RazaoSocial:     "Test Company",
	}

	cnaes := entity.GetAllCNAEs()

	// Should return only fiscal CNAE
	if len(cnaes) != 1 {
		t.Errorf("Expected 1 CNAE, got %d", len(cnaes))
	}

	if cnaes[0] != "6201-5/00" {
		t.Errorf("Expected fiscal CNAE, got %s", cnaes[0])
	}
}

func TestLegalEntitiesCache_Delete(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Delete existing cache entry", func(t *testing.T) {
		cpf := "12345678900"
		entities := []models.LegalEntity{
			{CNPJ: "12345678000190", CNAEFiscal: "1111-1/11", RazaoSocial: "Test"},
		}

		// Set cache
		err := c.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Verify it exists
		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get cache: %v", err)
		}
		if cached == nil {
			t.Fatal("Expected cache to exist")
		}

		// Delete directly using Redis client
		key := "legal_entities:" + cpf
		result, err := redisClient.Del(ctx, key).Result()
		if err != nil {
			t.Fatalf("Failed to delete key: %v", err)
		}
		if result != 1 {
			t.Errorf("Expected 1 key deleted, got %d", result)
		}

		// Verify it's gone
		cached, err = c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get cache after delete: %v", err)
		}
		if cached != nil {
			t.Error("Expected cache to be deleted")
		}
	})
}

func TestLegalEntitiesCache_Exists(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Check if key exists", func(t *testing.T) {
		cpf := "11111111111"

		// Verify key doesn't exist initially
		key := "legal_entities:" + cpf
		exists, err := redisClient.Exists(ctx, key).Result()
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists != 0 {
			t.Error("Expected key to not exist")
		}

		// Set cache
		entities := []models.LegalEntity{
			{CNPJ: "11111111000111", CNAEFiscal: "1111-1/11", RazaoSocial: "Test"},
		}
		err = c.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Verify key exists
		exists, err = redisClient.Exists(ctx, key).Result()
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists != 1 {
			t.Error("Expected key to exist")
		}
	})
}

func TestLegalEntitiesCache_ErrorHandling(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Get with invalid JSON in cache", func(t *testing.T) {
		cpf := "99999999999"
		key := "legal_entities:" + cpf

		// Set invalid JSON directly in Redis
		err := redisClient.Set(ctx, key, "invalid json {{{", 1*time.Minute).Err()
		if err != nil {
			t.Fatalf("Failed to set invalid JSON: %v", err)
		}

		// Try to get - should fail to unmarshal
		_, err = c.Get(ctx, cpf)
		if err == nil {
			t.Error("Expected error when getting invalid JSON")
		}
		if err != nil && err.Error() != "" {
			// Verify error message mentions unmarshal
			if len(err.Error()) == 0 {
				t.Error("Expected non-empty error message")
			}
		}
	})

	t.Run("Operations with closed connection", func(t *testing.T) {
		// Create a miniredis and immediately close it
		mrClosed, _ := miniredis.Run()
		closedClient := redis.NewClient(&redis.Options{
			Addr: mrClosed.Addr(),
		})
		mrClosed.Close() // Close the miniredis server
		closedClient.Close()

		closedCache := cache.NewLegalEntitiesCache(closedClient, 1*time.Minute)

		// Get should fail
		_, err := closedCache.Get(ctx, "12345678900")
		if err == nil {
			t.Error("Expected error with closed connection")
		}

		// Set should fail
		entities := []models.LegalEntity{
			{CNPJ: "12345678000190", CNAEFiscal: "1111-1/11", RazaoSocial: "Test"},
		}
		err = closedCache.Set(ctx, "12345678900", entities)
		if err == nil {
			t.Error("Expected error with closed connection")
		}
	})
}

func TestLegalEntitiesCache_MultipleEntries(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Store and retrieve multiple CPFs", func(t *testing.T) {
		cpfData := map[string][]models.LegalEntity{
			"11111111111": {
				{CNPJ: "11111111000111", CNAEFiscal: "1111-1/11", RazaoSocial: "Company 1"},
			},
			"22222222222": {
				{CNPJ: "22222222000122", CNAEFiscal: "2222-2/22", RazaoSocial: "Company 2A"},
				{CNPJ: "33333333000133", CNAEFiscal: "3333-3/33", RazaoSocial: "Company 2B"},
			},
			"33333333333": {
				{CNPJ: "44444444000144", CNAEFiscal: "4444-4/44", RazaoSocial: "Company 3"},
			},
		}

		// Set all entries
		for cpf, entities := range cpfData {
			err := c.Set(ctx, cpf, entities)
			if err != nil {
				t.Fatalf("Failed to set cache for CPF %s: %v", cpf, err)
			}
		}

		// Verify all entries
		for cpf, expectedEntities := range cpfData {
			cached, err := c.Get(ctx, cpf)
			if err != nil {
				t.Fatalf("Failed to get cache for CPF %s: %v", cpf, err)
			}
			if cached == nil {
				t.Fatalf("Expected cache for CPF %s, got nil", cpf)
			}
			if len(cached) != len(expectedEntities) {
				t.Errorf("CPF %s: expected %d entities, got %d", cpf, len(expectedEntities), len(cached))
			}
		}
	})
}

func TestLegalEntitiesCache_ConcurrentAccess(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Concurrent sets and gets", func(t *testing.T) {
		const numGoroutines = 20
		done := make(chan bool, numGoroutines)

		// Concurrent writes
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				cpf := "12345678900"
				entities := []models.LegalEntity{
					{
						CNPJ:        "12345678000190",
						CNAEFiscal:  "1111-1/11",
						RazaoSocial: "Concurrent Test",
					},
				}
				err := c.Set(ctx, cpf, entities)
				if err != nil {
					t.Errorf("Goroutine %d: failed to set: %v", index, err)
				}
				done <- true
			}(i)
		}

		// Wait for all writes
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify final state
		cached, err := c.Get(ctx, "12345678900")
		if err != nil {
			t.Fatalf("Failed to get after concurrent writes: %v", err)
		}
		if cached == nil {
			t.Error("Expected cache to exist after concurrent writes")
		}
	})
}

func TestLegalEntitiesCache_NilAndEmptyValues(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Set nil entities", func(t *testing.T) {
		cpf := "00000000000"

		err := c.Set(ctx, cpf, nil)
		if err != nil {
			t.Fatalf("Failed to set nil: %v", err)
		}

		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get nil value: %v", err)
		}
		if cached != nil && len(cached) != 0 {
			t.Errorf("Expected nil or empty, got %d entities", len(cached))
		}
	})

	t.Run("Entity with empty strings", func(t *testing.T) {
		cpf := "11111111111"
		entities := []models.LegalEntity{
			{
				CNPJ:            "",
				CNAEFiscal:      "",
				CNAESecundarias: []string{},
				RazaoSocial:     "",
			},
		}

		err := c.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set empty entity: %v", err)
		}

		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get empty entity: %v", err)
		}
		if cached == nil || len(cached) != 1 {
			t.Errorf("Expected 1 empty entity, got %v", cached)
		}
	})
}

func TestLegalEntitiesCache_SpecialCPFFormats(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	testCases := []struct {
		name string
		cpf  string
	}{
		{"Standard CPF", "12345678900"},
		{"CPF with zeros", "00000000000"},
		{"CPF with all nines", "99999999999"},
		{"CPF with special pattern", "11111111111"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entities := []models.LegalEntity{
				{CNPJ: "12345678000190", CNAEFiscal: "1111-1/11", RazaoSocial: "Test"},
			}

			err := c.Set(ctx, tc.cpf, entities)
			if err != nil {
				t.Fatalf("Failed to set cache for CPF %s: %v", tc.cpf, err)
			}

			cached, err := c.Get(ctx, tc.cpf)
			if err != nil {
				t.Fatalf("Failed to get cache for CPF %s: %v", tc.cpf, err)
			}
			if cached == nil {
				t.Errorf("Expected cache for CPF %s, got nil", tc.cpf)
			}
		})
	}
}

func TestLegalEntitiesCache_LargeDataSets(t *testing.T) {
	redisClient, ctx, cleanup := setupLegalEntitiesCacheTest(t)
	defer cleanup()

	c := cache.NewLegalEntitiesCache(redisClient, 1*time.Minute)

	t.Run("Store large number of entities for single CPF", func(t *testing.T) {
		cpf := "12345678900"

		// Create 100 entities
		entities := make([]models.LegalEntity, 100)
		for i := 0; i < 100; i++ {
			entities[i] = models.LegalEntity{
				CNPJ:        "12345678000190",
				CNAEFiscal:  "1111-1/11",
				RazaoSocial: "Company " + strconv.Itoa(i),
			}
		}

		err := c.Set(ctx, cpf, entities)
		if err != nil {
			t.Fatalf("Failed to set large dataset: %v", err)
		}

		cached, err := c.Get(ctx, cpf)
		if err != nil {
			t.Fatalf("Failed to get large dataset: %v", err)
		}
		if len(cached) != 100 {
			t.Errorf("Expected 100 entities, got %d", len(cached))
		}
	})
}
