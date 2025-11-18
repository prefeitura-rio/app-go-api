package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CourseCache manages caching of course listings
// Uses shorter TTL than reference data since courses change more frequently
type CourseCache struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewCourseCache creates a new cache instance for course listings
func NewCourseCache(redisClient *redis.Client, ttl time.Duration) *CourseCache {
	return &CourseCache{
		redis: redisClient,
		ttl:   ttl,
	}
}

// GetList retrieves a cached course listing by filter hash
// Returns nil, nil if not found (cache miss is not an error)
func (c *CourseCache) GetList(ctx context.Context, filterHash string) ([]byte, error) {
	key := c.buildListKey(filterHash)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss - not an error
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get course list from cache: %w", err)
	}

	return data, nil
}

// SetList stores a course listing in cache with TTL
func (c *CourseCache) SetList(ctx context.Context, filterHash string, data interface{}) error {
	key := c.buildListKey(filterHash)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal course list: %w", err)
	}

	if err := c.redis.Set(ctx, key, jsonData, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set course list cache: %w", err)
	}

	return nil
}

// InvalidateAll clears all course list caches
// Should be called after Create, Update, or Delete operations
func (c *CourseCache) InvalidateAll(ctx context.Context) error {
	pattern := c.buildKeyPattern()

	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = c.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan course cache keys: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cached course lists: %w", err)
		}
	}

	return nil
}

// buildListKey generates the Redis key for a course list query
// Format: "courses:list:{filterHash}"
func (c *CourseCache) buildListKey(filterHash string) string {
	return fmt.Sprintf("courses:list:%s", filterHash)
}

// buildKeyPattern generates the pattern for scanning all course list keys
// Format: "courses:list:*"
func (c *CourseCache) buildKeyPattern() string {
	return "courses:list:*"
}

// HashFilter creates a deterministic hash from filter parameters
// Used as cache key to differentiate different query combinations
func HashFilter(filter map[string]interface{}, page, limit int) string {
	// Create a deterministic string from the filter
	type filterKey struct {
		Filter map[string]interface{}
		Page   int
		Limit  int
	}

	key := filterKey{
		Filter: filter,
		Page:   page,
		Limit:  limit,
	}

	// Marshal to JSON for consistent ordering
	data, _ := json.Marshal(key)

	// Hash the JSON
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:16]) // Use first 16 bytes for shorter key
}
