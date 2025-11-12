package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// LegalEntitiesCache manages caching of legal entities from RMI API
type LegalEntitiesCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewLegalEntitiesCache creates a new cache instance with specified TTL
func NewLegalEntitiesCache(redisClient *redis.Client, ttl time.Duration) *LegalEntitiesCache {
	return &LegalEntitiesCache{
		redis: redisClient,
		ttl:   ttl,
	}
}

// Get retrieves legal entities for a CPF from cache
// Returns nil, nil if not found (cache miss is not an error)
func (c *LegalEntitiesCache) Get(ctx context.Context, cpf string) ([]models.LegalEntity, error) {
	key := c.buildKey(cpf)

	data, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Cache miss - not an error
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var entities []models.LegalEntity
	if err := json.Unmarshal([]byte(data), &entities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return entities, nil
}

// Set stores legal entities for a CPF in cache with TTL
func (c *LegalEntitiesCache) Set(ctx context.Context, cpf string, entities []models.LegalEntity) error {
	key := c.buildKey(cpf)

	data, err := json.Marshal(entities)
	if err != nil {
		return fmt.Errorf("failed to marshal entities: %w", err)
	}

	if err := c.redis.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// buildKey generates the Redis key for a CPF
// Format: "legal_entities:{cpf}"
func (c *LegalEntitiesCache) buildKey(cpf string) string {
	return fmt.Sprintf("legal_entities:%s", cpf)
}
