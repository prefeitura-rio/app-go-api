package providers

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/cache"
)

// ReferenceCaches bundles the three reference-data Redis caches to avoid
// Wire ambiguity caused by all three having the same *cache.ReferenceDataCache type.
type ReferenceCaches struct {
	Categorias      *cache.ReferenceDataCache
	Acessibilidades *cache.ReferenceDataCache
	Escolaridades   *cache.ReferenceDataCache
}

// ProvideReferenceCaches creates all three reference-data caches (1-hour TTL each)
func ProvideReferenceCaches(redisClient *redis.Client) *ReferenceCaches {
	return &ReferenceCaches{
		Categorias:      cache.NewReferenceDataCache(redisClient, 1*time.Hour, "categorias"),
		Acessibilidades: cache.NewReferenceDataCache(redisClient, 1*time.Hour, "acessibilidades"),
		Escolaridades:   cache.NewReferenceDataCache(redisClient, 1*time.Hour, "escolaridades"),
	}
}

// ProvideLegalEntitiesCache creates the legal entities Redis cache with a 30-minute TTL
func ProvideLegalEntitiesCache(redisClient *redis.Client) *cache.LegalEntitiesCache {
	return cache.NewLegalEntitiesCache(redisClient, 30*time.Minute)
}

// ProvideCourseCache creates the course cache with a 5-minute TTL.
// Shorter TTL than reference data because courses change more frequently.
func ProvideCourseCache(redisClient *redis.Client) *cache.CourseCache {
	return cache.NewCourseCache(redisClient, 5*time.Minute)
}
