package cache_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCourseCacheTest(t *testing.T) (*redis.Client, context.Context, func()) {
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

func TestNewCourseCache(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	ttl := 10 * time.Minute
	c := cache.NewCourseCache(redisClient, ttl)

	assert.NotNil(t, c)
}

func TestCourseCache_SetAndGetList(t *testing.T) {
	redisClient, ctx, cleanup := setupCourseCacheTest(t)
	defer cleanup()

	c := cache.NewCourseCache(redisClient, 1*time.Minute)

	t.Run("Set and Get valid data", func(t *testing.T) {
		filterHash := "test-filter-123"
		courseData := map[string]interface{}{
			"courses": []map[string]string{
				{"id": "1", "name": "Go Programming"},
				{"id": "2", "name": "Kubernetes Basics"},
			},
			"total": 2,
		}

		// Set the cache
		err := c.SetList(ctx, filterHash, courseData)
		require.NoError(t, err)

		// Get from cache
		cached, err := c.GetList(ctx, filterHash)
		require.NoError(t, err)
		require.NotNil(t, cached)

		// Verify data integrity
		var retrievedData map[string]interface{}
		err = json.Unmarshal(cached, &retrievedData)
		require.NoError(t, err)

		assert.Equal(t, float64(2), retrievedData["total"])
	})

	t.Run("Get non-existent key returns nil", func(t *testing.T) {
		cached, err := c.GetList(ctx, "non-existent-filter")
		require.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Set empty list", func(t *testing.T) {
		filterHash := "empty-filter"
		emptyData := map[string]interface{}{
			"courses": []interface{}{},
			"total":   0,
		}

		err := c.SetList(ctx, filterHash, emptyData)
		require.NoError(t, err)

		cached, err := c.GetList(ctx, filterHash)
		require.NoError(t, err)
		require.NotNil(t, cached)

		var retrievedData map[string]interface{}
		err = json.Unmarshal(cached, &retrievedData)
		require.NoError(t, err)
		assert.Equal(t, float64(0), retrievedData["total"])
	})

	t.Run("Overwrite existing cache", func(t *testing.T) {
		filterHash := "overwrite-filter"

		// First set
		data1 := map[string]interface{}{"version": 1}
		err := c.SetList(ctx, filterHash, data1)
		require.NoError(t, err)

		// Overwrite with new data
		data2 := map[string]interface{}{"version": 2}
		err = c.SetList(ctx, filterHash, data2)
		require.NoError(t, err)

		// Verify new data
		cached, err := c.GetList(ctx, filterHash)
		require.NoError(t, err)

		var retrievedData map[string]interface{}
		err = json.Unmarshal(cached, &retrievedData)
		require.NoError(t, err)
		assert.Equal(t, float64(2), retrievedData["version"])
	})

	t.Run("Cache key format", func(t *testing.T) {
		filterHash := "format-test-hash"
		data := map[string]interface{}{"test": true}

		err := c.SetList(ctx, filterHash, data)
		require.NoError(t, err)

		// Verify key format directly in Redis
		expectedKey := "courses:list:" + filterHash
		exists, err := redisClient.Exists(ctx, expectedKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})

	t.Run("TTL is set correctly", func(t *testing.T) {
		shortTTLCache := cache.NewCourseCache(redisClient, 5*time.Second)
		filterHash := "ttl-test-hash"
		data := map[string]interface{}{"test": "ttl"}

		err := shortTTLCache.SetList(ctx, filterHash, data)
		require.NoError(t, err)

		expectedKey := "courses:list:" + filterHash
		ttl, err := redisClient.TTL(ctx, expectedKey).Result()
		require.NoError(t, err)

		// TTL should be approximately 5 seconds (allow some variance)
		assert.Greater(t, ttl, 4*time.Second)
		assert.Less(t, ttl, 6*time.Second)
	})
}

func TestCourseCache_InvalidateAll(t *testing.T) {
	redisClient, ctx, cleanup := setupCourseCacheTest(t)
	defer cleanup()

	c := cache.NewCourseCache(redisClient, 1*time.Minute)

	t.Run("Invalidate multiple caches", func(t *testing.T) {
		// Set multiple caches
		filters := []string{"filter1", "filter2", "filter3"}
		for _, filter := range filters {
			err := c.SetList(ctx, filter, map[string]interface{}{"filter": filter})
			require.NoError(t, err)
		}

		// Verify caches exist
		for _, filter := range filters {
			cached, err := c.GetList(ctx, filter)
			require.NoError(t, err)
			assert.NotNil(t, cached)
		}

		// Invalidate all
		err := c.InvalidateAll(ctx)
		require.NoError(t, err)

		// Verify all caches are gone
		for _, filter := range filters {
			cached, err := c.GetList(ctx, filter)
			require.NoError(t, err)
			assert.Nil(t, cached)
		}
	})

	t.Run("Invalidate with no existing caches", func(t *testing.T) {
		// Should not error when no caches exist
		err := c.InvalidateAll(ctx)
		require.NoError(t, err)
	})

	t.Run("Invalidate does not affect other prefixes", func(t *testing.T) {
		// Set a course cache
		err := c.SetList(ctx, "course-filter", map[string]interface{}{"type": "course"})
		require.NoError(t, err)

		// Set a different type of cache with different prefix
		otherKey := "other:list:filter"
		err = redisClient.Set(ctx, otherKey, "data", 1*time.Minute).Err()
		require.NoError(t, err)

		// Invalidate course caches
		err = c.InvalidateAll(ctx)
		require.NoError(t, err)

		// Verify course cache is gone
		cached, err := c.GetList(ctx, "course-filter")
		require.NoError(t, err)
		assert.Nil(t, cached)

		// Verify other cache still exists
		exists, err := redisClient.Exists(ctx, otherKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})
}

func TestHashFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter map[string]interface{}
		page   int
		limit  int
	}{
		{
			name:   "Simple filter",
			filter: map[string]interface{}{"category": "tech"},
			page:   1,
			limit:  10,
		},
		{
			name:   "Empty filter",
			filter: map[string]interface{}{},
			page:   1,
			limit:  20,
		},
		{
			name: "Complex filter",
			filter: map[string]interface{}{
				"category":      "tech",
				"accessibility": []string{"wheelchair", "audio"},
				"min_duration":  60,
			},
			page:  2,
			limit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := cache.HashFilter(tt.filter, tt.page, tt.limit)
			hash2 := cache.HashFilter(tt.filter, tt.page, tt.limit)

			// Same inputs should produce same hash (deterministic)
			assert.Equal(t, hash1, hash2)

			// Hash should not be empty
			assert.NotEmpty(t, hash1)

			// Hash should be 32 characters (16 bytes as hex)
			assert.Equal(t, 32, len(hash1))
		})
	}

	t.Run("Different filters produce different hashes", func(t *testing.T) {
		filter1 := map[string]interface{}{"category": "tech"}
		filter2 := map[string]interface{}{"category": "arts"}

		hash1 := cache.HashFilter(filter1, 1, 10)
		hash2 := cache.HashFilter(filter2, 1, 10)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("Different pages produce different hashes", func(t *testing.T) {
		filter := map[string]interface{}{"category": "tech"}

		hash1 := cache.HashFilter(filter, 1, 10)
		hash2 := cache.HashFilter(filter, 2, 10)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("Different limits produce different hashes", func(t *testing.T) {
		filter := map[string]interface{}{"category": "tech"}

		hash1 := cache.HashFilter(filter, 1, 10)
		hash2 := cache.HashFilter(filter, 1, 20)

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestCourseCache_ErrorHandling(t *testing.T) {
	redisClient, ctx, cleanup := setupCourseCacheTest(t)
	defer cleanup()

	c := cache.NewCourseCache(redisClient, 1*time.Minute)

	t.Run("Set with unmarshalable data", func(t *testing.T) {
		// Create a channel which cannot be marshaled to JSON
		badData := make(chan int)

		err := c.SetList(ctx, "bad-filter", badData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal")
	})

	t.Run("Get with closed Redis connection", func(t *testing.T) {
		// Create a new Redis client and close it
		closedClient := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			DB:   15,
		})
		closedClient.Close()

		closedCache := cache.NewCourseCache(closedClient, 1*time.Minute)

		_, err := closedCache.GetList(ctx, "test")
		assert.Error(t, err)
	})
}

func TestCourseCache_ConcurrentAccess(t *testing.T) {
	redisClient, ctx, cleanup := setupCourseCacheTest(t)
	defer cleanup()

	c := cache.NewCourseCache(redisClient, 1*time.Minute)

	t.Run("Concurrent sets and gets", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		// Concurrent sets
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				filterHash := "concurrent-filter"
				data := map[string]interface{}{"index": index}
				err := c.SetList(ctx, filterHash, data)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all sets to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify cache exists and is valid
		cached, err := c.GetList(ctx, "concurrent-filter")
		require.NoError(t, err)
		assert.NotNil(t, cached)
	})
}
