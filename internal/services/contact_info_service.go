package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/redis/go-redis/v9"
)

// TokenManager is an interface for getting service account tokens
type TokenManager interface {
	GetToken(ctx context.Context) (string, error)
}

// ContactInfoService handles fetching and caching of CNPJ owner contact information
type ContactInfoService struct {
	rmiClient    *clients.RMIClient
	tokenManager TokenManager
	redisClient  *redis.Client
	cacheTTL     time.Duration
}

// NewContactInfoService creates a new contact info service
func NewContactInfoService(
	rmiClient *clients.RMIClient,
	tokenManager TokenManager,
	redisClient *redis.Client,
	cfg *config.AppConfig,
) *ContactInfoService {
	return &ContactInfoService{
		rmiClient:    rmiClient,
		tokenManager: tokenManager,
		redisClient:  redisClient,
		cacheTTL:     cfg.Cache.ContactInfoTTL,
	}
}

// GetCNPJOwnerContactInfo fetches contact information for a CNPJ owner
// Uses Redis cache with configurable TTL, falls back to RMI API if cache miss
// Returns empty strings for email/phone on error (graceful degradation)
func (s *ContactInfoService) GetCNPJOwnerContactInfo(ctx context.Context, cnpj string) (*models.CNPJOwnerInfo, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		var ownerInfo models.CNPJOwnerInfo
		if err := json.Unmarshal([]byte(cachedData), &ownerInfo); err == nil {
			log.Printf("[CONTACT_INFO_CACHE_HIT] CNPJ: %s", cnpj)
			return &ownerInfo, nil
		}
		// Cache data is corrupted, continue to fetch from API
		log.Printf("[CONTACT_INFO_CACHE_CORRUPTED] CNPJ: %s, error: %v", cnpj, err)
	} else if err != redis.Nil {
		// Redis error (not a cache miss), log but continue
		log.Printf("[CONTACT_INFO_CACHE_ERROR] CNPJ: %s, error: %v", cnpj, err)
	}

	// Cache miss or error - fetch from RMI API
	log.Printf("[CONTACT_INFO_CACHE_MISS] CNPJ: %s", cnpj)

	// Get service account token
	token, err := s.tokenManager.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get service account token: %w", err)
	}

	// Fetch from RMI API
	ownerInfo, err := s.rmiClient.GetCNPJOwnerInfo(ctx, token, cnpj)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CNPJ owner info from RMI: %w", err)
	}

	// Cache the result
	dataToCache, err := json.Marshal(ownerInfo)
	if err != nil {
		log.Printf("[CONTACT_INFO_CACHE_MARSHAL_ERROR] CNPJ: %s, error: %v", cnpj, err)
		// Don't fail, just return the data without caching
		return ownerInfo, nil
	}

	err = s.redisClient.Set(ctx, cacheKey, dataToCache, s.cacheTTL).Err()
	if err != nil {
		log.Printf("[CONTACT_INFO_CACHE_SET_ERROR] CNPJ: %s, error: %v", cnpj, err)
		// Don't fail, just return the data without caching
	} else {
		log.Printf("[CONTACT_INFO_CACHED] CNPJ: %s, TTL: %v", cnpj, s.cacheTTL)
	}

	return ownerInfo, nil
}

// GetMultipleCNPJOwnerContactInfo fetches contact info for multiple CNPJs in parallel
// Returns a map of CNPJ -> CNPJOwnerInfo
// Errors for individual CNPJs are logged but don't stop the entire operation
func (s *ContactInfoService) GetMultipleCNPJOwnerContactInfo(ctx context.Context, cnpjs []string) map[string]*models.CNPJOwnerInfo {
	results := make(map[string]*models.CNPJOwnerInfo)
	resultChan := make(chan struct {
		cnpj string
		info *models.CNPJOwnerInfo
		err  error
	}, len(cnpjs))

	// Fetch in parallel
	for _, cnpj := range cnpjs {
		go func(c string) {
			info, err := s.GetCNPJOwnerContactInfo(ctx, c)
			resultChan <- struct {
				cnpj string
				info *models.CNPJOwnerInfo
				err  error
			}{cnpj: c, info: info, err: err}
		}(cnpj)
	}

	// Collect results
	for i := 0; i < len(cnpjs); i++ {
		result := <-resultChan
		if result.err != nil {
			log.Printf("[CONTACT_INFO_ERROR] CNPJ: %s, error: %v", result.cnpj, result.err)
			// Don't add to results map (graceful degradation)
			continue
		}
		results[result.cnpj] = result.info
	}

	return results
}
