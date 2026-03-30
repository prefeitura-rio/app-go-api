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

func setupReferenceDataCacheTest(t *testing.T) (*redis.Client, context.Context, func()) {
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

func TestNewReferenceDataCache(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	tests := []struct {
		name   string
		ttl    time.Duration
		prefix string
	}{
		{"Categories cache", 24 * time.Hour, "categorias"},
		{"Accessibility cache", 24 * time.Hour, "acessibilidades"},
		{"Education levels cache", 24 * time.Hour, "escolaridades"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cache.NewReferenceDataCache(redisClient, tt.ttl, tt.prefix)
			assert.NotNil(t, c)
		})
	}
}

func TestReferenceDataCache_ListOperations(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	prefix := "test_categories"
	c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, prefix)

	t.Run("Set and Get list", func(t *testing.T) {
		listData := []map[string]interface{}{
			{"id": 1, "name": "Technology"},
			{"id": 2, "name": "Arts"},
			{"id": 3, "name": "Sports"},
		}

		// Set the list
		err := c.SetList(ctx, listData)
		require.NoError(t, err)

		// Get the list
		cached, err := c.GetList(ctx)
		require.NoError(t, err)
		require.NotNil(t, cached)

		// Verify data integrity
		var retrievedData []map[string]interface{}
		err = json.Unmarshal(cached, &retrievedData)
		require.NoError(t, err)
		assert.Equal(t, len(listData), len(retrievedData))
	})

	t.Run("Get non-existent list returns nil", func(t *testing.T) {
		// Use a different prefix for this test
		emptyCache := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "nonexistent")
		cached, err := emptyCache.GetList(ctx)
		require.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Set empty list", func(t *testing.T) {
		emptyList := []interface{}{}

		err := c.SetList(ctx, emptyList)
		require.NoError(t, err)

		cached, err := c.GetList(ctx)
		require.NoError(t, err)
		require.NotNil(t, cached)

		var retrievedData []interface{}
		err = json.Unmarshal(cached, &retrievedData)
		require.NoError(t, err)
		assert.Equal(t, 0, len(retrievedData))
	})

	t.Run("List key format", func(t *testing.T) {
		data := []string{"item1", "item2"}
		err := c.SetList(ctx, data)
		require.NoError(t, err)

		expectedKey := "ref:" + prefix + ":list"
		exists, err := redisClient.Exists(ctx, expectedKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})

	t.Run("List TTL is set correctly", func(t *testing.T) {
		shortTTLCache := cache.NewReferenceDataCache(redisClient, 5*time.Second, "ttl_test")
		data := []string{"test"}

		err := shortTTLCache.SetList(ctx, data)
		require.NoError(t, err)

		expectedKey := "ref:ttl_test:list"
		ttl, err := redisClient.TTL(ctx, expectedKey).Result()
		require.NoError(t, err)

		assert.Greater(t, ttl, 4*time.Second)
		assert.Less(t, ttl, 6*time.Second)
	})
}

func TestReferenceDataCache_ItemOperations(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	prefix := "test_items"
	c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, prefix)

	t.Run("Set and Get item by ID", func(t *testing.T) {
		itemID := 42
		itemData := map[string]interface{}{
			"id":          itemID,
			"name":        "Test Item",
			"description": "A test reference data item",
		}

		// Set the item
		err := c.SetByID(ctx, itemID, itemData)
		require.NoError(t, err)

		// Get the item
		cached, err := c.GetByID(ctx, itemID)
		require.NoError(t, err)
		require.NotNil(t, cached)

		// Verify data integrity
		var retrievedData map[string]interface{}
		err = json.Unmarshal(cached, &retrievedData)
		require.NoError(t, err)
		assert.Equal(t, "Test Item", retrievedData["name"])
	})

	t.Run("Get non-existent item returns nil", func(t *testing.T) {
		cached, err := c.GetByID(ctx, 9999)
		require.NoError(t, err)
		assert.Nil(t, cached)
	})

	t.Run("Item key format", func(t *testing.T) {
		itemID := 123
		data := map[string]string{"test": "value"}

		err := c.SetByID(ctx, itemID, data)
		require.NoError(t, err)

		expectedKey := "ref:" + prefix + ":item:123"
		exists, err := redisClient.Exists(ctx, expectedKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})

	t.Run("Multiple items with different IDs", func(t *testing.T) {
		items := map[int]string{
			1: "Item One",
			2: "Item Two",
			3: "Item Three",
		}

		// Set multiple items
		for id, name := range items {
			err := c.SetByID(ctx, id, map[string]string{"name": name})
			require.NoError(t, err)
		}

		// Verify all items
		for id, expectedName := range items {
			cached, err := c.GetByID(ctx, id)
			require.NoError(t, err)
			require.NotNil(t, cached)

			var data map[string]string
			err = json.Unmarshal(cached, &data)
			require.NoError(t, err)
			assert.Equal(t, expectedName, data["name"])
		}
	})
}

func TestReferenceDataCache_Invalidate(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	prefix := "test_invalidate"
	c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, prefix)

	t.Run("Invalidate all clears list and items", func(t *testing.T) {
		// Set list
		err := c.SetList(ctx, []string{"item1", "item2"})
		require.NoError(t, err)

		// Set items
		err = c.SetByID(ctx, 1, map[string]string{"name": "Item 1"})
		require.NoError(t, err)
		err = c.SetByID(ctx, 2, map[string]string{"name": "Item 2"})
		require.NoError(t, err)

		// Verify all exist
		list, err := c.GetList(ctx)
		require.NoError(t, err)
		assert.NotNil(t, list)

		item1, err := c.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.NotNil(t, item1)

		// Invalidate all
		err = c.Invalidate(ctx)
		require.NoError(t, err)

		// Verify all are gone
		list, err = c.GetList(ctx)
		require.NoError(t, err)
		assert.Nil(t, list)

		item1, err = c.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.Nil(t, item1)

		item2, err := c.GetByID(ctx, 2)
		require.NoError(t, err)
		assert.Nil(t, item2)
	})

	t.Run("Invalidate with no existing data", func(t *testing.T) {
		emptyCache := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "empty_prefix")
		err := emptyCache.Invalidate(ctx)
		require.NoError(t, err)
	})

	t.Run("Invalidate does not affect other prefixes", func(t *testing.T) {
		cache1 := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "prefix1")
		cache2 := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "prefix2")

		// Set data in both caches
		err := cache1.SetList(ctx, []string{"cache1"})
		require.NoError(t, err)
		err = cache2.SetList(ctx, []string{"cache2"})
		require.NoError(t, err)

		// Invalidate cache1
		err = cache1.Invalidate(ctx)
		require.NoError(t, err)

		// Verify cache1 is gone but cache2 still exists
		list1, err := cache1.GetList(ctx)
		require.NoError(t, err)
		assert.Nil(t, list1)

		list2, err := cache2.GetList(ctx)
		require.NoError(t, err)
		assert.NotNil(t, list2)
	})
}

func TestReferenceDataCache_InvalidateByID(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	prefix := "test_invalidate_by_id"
	c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, prefix)

	t.Run("Invalidate specific item and list", func(t *testing.T) {
		// Set list and items
		err := c.SetList(ctx, []map[string]interface{}{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		})
		require.NoError(t, err)

		err = c.SetByID(ctx, 1, map[string]string{"name": "Item 1"})
		require.NoError(t, err)
		err = c.SetByID(ctx, 2, map[string]string{"name": "Item 2"})
		require.NoError(t, err)

		// Invalidate item 1
		err = c.InvalidateByID(ctx, 1)
		require.NoError(t, err)

		// Verify item 1 is gone
		item1, err := c.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.Nil(t, item1)

		// Verify list is also gone (stale)
		list, err := c.GetList(ctx)
		require.NoError(t, err)
		assert.Nil(t, list)

		// Verify item 2 still exists
		item2, err := c.GetByID(ctx, 2)
		require.NoError(t, err)
		assert.NotNil(t, item2)
	})

	t.Run("Invalidate non-existent item", func(t *testing.T) {
		// Should not error when invalidating non-existent item
		err := c.InvalidateByID(ctx, 9999)
		require.NoError(t, err)
	})
}

func TestReferenceDataCache_ErrorHandling(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "error_test")

	t.Run("SetList with unmarshalable data", func(t *testing.T) {
		badData := make(chan int)
		err := c.SetList(ctx, badData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal")
	})

	t.Run("SetByID with unmarshalable data", func(t *testing.T) {
		badData := make(chan int)
		err := c.SetByID(ctx, 1, badData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal")
	})

	t.Run("Operations with closed Redis connection", func(t *testing.T) {
		closedClient := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			DB:   15,
		})
		closedClient.Close()

		closedCache := cache.NewReferenceDataCache(closedClient, 1*time.Minute, "closed")

		// All operations should fail gracefully
		_, err := closedCache.GetList(ctx)
		assert.Error(t, err)

		_, err = closedCache.GetByID(ctx, 1)
		assert.Error(t, err)

		err = closedCache.SetList(ctx, []string{"test"})
		assert.Error(t, err)

		err = closedCache.SetByID(ctx, 1, "test")
		assert.Error(t, err)
	})
}

func TestReferenceDataCache_DifferentPrefixes(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	prefixes := []string{"categorias", "acessibilidades", "escolaridades"}

	for _, prefix := range prefixes {
		t.Run("Prefix: "+prefix, func(t *testing.T) {
			c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, prefix)

			// Set and get data
			data := map[string]string{"prefix": prefix}
			err := c.SetList(ctx, data)
			require.NoError(t, err)

			cached, err := c.GetList(ctx)
			require.NoError(t, err)
			assert.NotNil(t, cached)

			// Verify key pattern
			expectedKey := "ref:" + prefix + ":list"
			exists, err := redisClient.Exists(ctx, expectedKey).Result()
			require.NoError(t, err)
			assert.Equal(t, int64(1), exists)
		})
	}
}

func TestReferenceDataCache_ConcurrentAccess(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "concurrent")

	t.Run("Concurrent list operations", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		// Concurrent sets
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				data := map[string]interface{}{"index": index}
				err := c.SetList(ctx, data)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify cache is valid
		cached, err := c.GetList(ctx)
		require.NoError(t, err)
		assert.NotNil(t, cached)
	})

	t.Run("Concurrent item operations", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		// Concurrent sets to different IDs
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				data := map[string]interface{}{"id": index}
				err := c.SetByID(ctx, index, data)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify all items exist
		for i := 0; i < numGoroutines; i++ {
			cached, err := c.GetByID(ctx, i)
			require.NoError(t, err)
			assert.NotNil(t, cached)
		}
	})
}

func TestReferenceDataCache_EdgeCases(t *testing.T) {
	redisClient, ctx, cleanup := setupReferenceDataCacheTest(t)
	defer cleanup()

	t.Run("Zero ID", func(t *testing.T) {
		c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "zero_id")
		data := map[string]string{"id": "zero"}

		err := c.SetByID(ctx, 0, data)
		require.NoError(t, err)

		cached, err := c.GetByID(ctx, 0)
		require.NoError(t, err)
		assert.NotNil(t, cached)
	})

	t.Run("Negative ID", func(t *testing.T) {
		c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "negative_id")
		data := map[string]string{"id": "negative"}

		err := c.SetByID(ctx, -1, data)
		require.NoError(t, err)

		cached, err := c.GetByID(ctx, -1)
		require.NoError(t, err)
		assert.NotNil(t, cached)
	})

	t.Run("Very large ID", func(t *testing.T) {
		c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "large_id")
		data := map[string]string{"id": "large"}

		largeID := 2147483647 // Max int32

		err := c.SetByID(ctx, largeID, data)
		require.NoError(t, err)

		cached, err := c.GetByID(ctx, largeID)
		require.NoError(t, err)
		assert.NotNil(t, cached)
	})

	t.Run("Empty prefix", func(t *testing.T) {
		c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "")
		data := []string{"test"}

		err := c.SetList(ctx, data)
		require.NoError(t, err)

		cached, err := c.GetList(ctx)
		require.NoError(t, err)
		assert.NotNil(t, cached)
	})

	t.Run("Prefix with special characters", func(t *testing.T) {
		c := cache.NewReferenceDataCache(redisClient, 1*time.Minute, "test:prefix:with:colons")
		data := []string{"test"}

		err := c.SetList(ctx, data)
		require.NoError(t, err)

		cached, err := c.GetList(ctx)
		require.NoError(t, err)
		assert.NotNil(t, cached)
	})

	t.Run("Zero TTL", func(t *testing.T) {
		c := cache.NewReferenceDataCache(redisClient, 0, "zero_ttl")
		data := []string{"test"}

		err := c.SetList(ctx, data)
		require.NoError(t, err)

		// With zero TTL, key should expire immediately or not have TTL
		_, err = c.GetList(ctx)
		require.NoError(t, err)
		// Data might or might not exist depending on Redis behavior with 0 TTL
		// This test just ensures no panic or error
	})
}
