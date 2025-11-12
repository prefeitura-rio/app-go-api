package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/redis/go-redis/v9"
)

func TestLegalEntitiesCache_SetAndGet(t *testing.T) {
	// Setup: Create a Redis client for testing
	// Using a test database (DB 15) to avoid conflicts
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use test database
	})

	// Ensure Redis is available
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available for testing: %v", err)
		return
	}

	// Clean up test data before and after
	defer redisClient.FlushDB(ctx)
	redisClient.FlushDB(ctx)

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
