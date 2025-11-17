package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ReferenceDataCache manages caching of static reference data (categories, accessibility, etc.)
// This cache is designed for data that rarely changes and is frequently accessed
type ReferenceDataCache struct {
	redis  *redis.Client
	ttl    time.Duration
	prefix string
}

// NewReferenceDataCache creates a new cache instance for a specific reference data type
// prefix should be the entity type (e.g., "categorias", "acessibilidades", "escolaridades")
func NewReferenceDataCache(redisClient *redis.Client, ttl time.Duration, prefix string) *ReferenceDataCache {
	return &ReferenceDataCache{
		redis:  redisClient,
		ttl:    ttl,
		prefix: prefix,
	}
}

// GetList retrieves a cached list of items
// Returns nil, nil if not found (cache miss is not an error)
func (c *ReferenceDataCache) GetList(ctx context.Context) ([]byte, error) {
	key := c.buildListKey()

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss - not an error
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list from cache: %w", err)
	}

	return data, nil
}

// SetList stores a list of items in cache with TTL
func (c *ReferenceDataCache) SetList(ctx context.Context, data interface{}) error {
	key := c.buildListKey()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal list: %w", err)
	}

	if err := c.redis.Set(ctx, key, jsonData, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set list cache: %w", err)
	}

	return nil
}

// GetByID retrieves a cached item by ID
// Returns nil, nil if not found (cache miss is not an error)
func (c *ReferenceDataCache) GetByID(ctx context.Context, id int) ([]byte, error) {
	key := c.buildItemKey(id)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss - not an error
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item from cache: %w", err)
	}

	return data, nil
}

// SetByID stores an item by ID in cache with TTL
func (c *ReferenceDataCache) SetByID(ctx context.Context, id int, data interface{}) error {
	key := c.buildItemKey(id)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	if err := c.redis.Set(ctx, key, jsonData, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set item cache: %w", err)
	}

	return nil
}

// Invalidate clears all cached data for this entity type
// Should be called after Create, Update, or Delete operations
func (c *ReferenceDataCache) Invalidate(ctx context.Context) error {
	// Use SCAN to find all keys with this prefix
	pattern := c.buildKeyPattern()

	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = c.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cached keys: %w", err)
		}
	}

	return nil
}

// InvalidateByID clears a specific cached item
// Should be called after Update or Delete operations for a specific item
func (c *ReferenceDataCache) InvalidateByID(ctx context.Context, id int) error {
	key := c.buildItemKey(id)

	if err := c.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cached item: %w", err)
	}

	// Also invalidate the list cache since it's now stale
	listKey := c.buildListKey()
	if err := c.redis.Del(ctx, listKey).Err(); err != nil {
		return fmt.Errorf("failed to delete cached list: %w", err)
	}

	return nil
}

// buildListKey generates the Redis key for the full list
// Format: "ref:{prefix}:list"
func (c *ReferenceDataCache) buildListKey() string {
	return fmt.Sprintf("ref:%s:list", c.prefix)
}

// buildItemKey generates the Redis key for a specific item
// Format: "ref:{prefix}:item:{id}"
func (c *ReferenceDataCache) buildItemKey(id int) string {
	return fmt.Sprintf("ref:%s:item:%d", c.prefix, id)
}

// buildKeyPattern generates the pattern for scanning all keys of this type
// Format: "ref:{prefix}:*"
func (c *ReferenceDataCache) buildKeyPattern() string {
	return fmt.Sprintf("ref:%s:*", c.prefix)
}
