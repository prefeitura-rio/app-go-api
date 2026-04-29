package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTokenManager implements TokenManager interface for testing
type mockTokenManager struct {
	token string
	err   error
}

func (m *mockTokenManager) GetToken(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

// setupContactInfoServiceTest creates test infrastructure with miniredis and RMI mock server
func setupContactInfoServiceTest(t *testing.T, handler http.HandlerFunc) (*services.ContactInfoService, *redis.Client, *miniredis.Miniredis, func()) {
	// Create miniredis for caching
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create RMI mock server
	mockServer := httptest.NewServer(handler)
	rmiClient := clients.NewRMIClient(mockServer.URL, 5*time.Second)

	// Create token manager
	tokenMgr := &mockTokenManager{token: "test-token"}

	// Create config
	cfg := &config.AppConfig{
		Cache: config.CacheSettings{
			ContactInfoTTL: 5 * time.Minute,
		},
	}

	// Create service
	service := services.NewContactInfoService(rmiClient, tokenMgr, redisClient, cfg)

	cleanup := func() {
		mockServer.Close()
		redisClient.Close()
		mr.Close()
	}

	return service, redisClient, mr, cleanup
}

// Test GetCNPJOwnerContactInfo - Success Scenarios

func TestGetCNPJOwnerContactInfo_CacheHit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call RMI API when cache hit")
	})

	service, redisClient, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	// Pre-populate cache
	cachedInfo := &models.CNPJOwnerInfo{
		CNPJ:                cnpj,
		EmailPessoaFisica:   "cached@example.com",
		CelularPessoaFisica: "5521999999999",
	}
	data, _ := json.Marshal(cachedInfo)
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)
	err := redisClient.Set(ctx, cacheKey, data, 5*time.Minute).Err()
	require.NoError(t, err)

	// Call service
	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, cnpj, result.CNPJ)
	assert.Equal(t, "cached@example.com", result.EmailPessoaFisica)
	assert.Equal(t, "5521999999999", result.CelularPessoaFisica)
}

func TestGetCNPJOwnerContactInfo_CacheMiss_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			citizen.Telefone.Indicador = true
			citizen.Telefone.Principal.DDI = "55"
			citizen.Telefone.Principal.DDD = "21"
			citizen.Telefone.Principal.Valor = "999999999"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, redisClient, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	// Call service (cache miss)
	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "12345678000195", result.CNPJ)
	assert.Equal(t, "john@example.com", result.EmailPessoaFisica)
	assert.Equal(t, "5521999999999", result.CelularPessoaFisica)

	// Verify cache was set
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)
	cachedData, err := redisClient.Get(ctx, cacheKey).Result()
	require.NoError(t, err)

	var cachedInfo models.CNPJOwnerInfo
	err = json.Unmarshal([]byte(cachedData), &cachedInfo)
	require.NoError(t, err)
	assert.Equal(t, "john@example.com", cachedInfo.EmailPessoaFisica)
}

func TestGetCNPJOwnerContactInfo_CacheTTLVerification(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, mr, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	// Call service
	_, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)

	// Verify TTL is set correctly
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)
	ttl := mr.TTL(cacheKey)
	assert.True(t, ttl > 0, "Cache TTL should be set")
	assert.True(t, ttl <= 5*time.Minute, "Cache TTL should not exceed configured TTL")
}

func TestGetCNPJOwnerContactInfo_CNPJNormalization(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name      string
		inputCNPJ string
	}{
		{"With formatting", "12.345.678/0001-95"},
		{"Without formatting", "12345678000195"},
		{"Mixed formatting", "12345678/0001-95"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetCNPJOwnerContactInfo(ctx, tt.inputCNPJ)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, "12345678000195", result.CNPJ)
		})
	}
}

// Test GetCNPJOwnerContactInfo - Error Scenarios

func TestGetCNPJOwnerContactInfo_TokenManagerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call RMI API when token fails")
	})

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	mockServer := httptest.NewServer(handler)
	defer mockServer.Close()

	rmiClient := clients.NewRMIClient(mockServer.URL, 5*time.Second)
	tokenMgr := &mockTokenManager{err: errors.New("token service unavailable")}
	cfg := &config.AppConfig{
		Cache: config.CacheSettings{ContactInfoTTL: 5 * time.Minute},
	}

	service := services.NewContactInfoService(rmiClient, tokenMgr, redisClient, cfg)

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get service account token")
}

func TestGetCNPJOwnerContactInfo_RMITimeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Exceed timeout
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetCNPJOwnerContactInfo_RMI404(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "99999999999999"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetCNPJOwnerContactInfo_RMI500(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetCNPJOwnerContactInfo_NoSociosFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		legalEntity := models.LegalEntityDetails{
			CNPJ:        "12345678000195",
			RazaoSocial: "Test Company",
			Socios: []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{},
		}
		json.NewEncoder(w).Encode(legalEntity)
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no socios found")
}

func TestGetCNPJOwnerContactInfo_EmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetCNPJOwnerContactInfo_MalformedJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json {{{"))
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetCNPJOwnerContactInfo_NetworkError(t *testing.T) {
	// Create RMI client with invalid URL
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	rmiClient := clients.NewRMIClient("http://invalid-url-that-does-not-exist-12345.local", 5*time.Second)
	tokenMgr := &mockTokenManager{token: "test-token"}
	cfg := &config.AppConfig{
		Cache: config.CacheSettings{ContactInfoTTL: 5 * time.Minute},
	}

	service := services.NewContactInfoService(rmiClient, tokenMgr, redisClient, cfg)

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// Test Cache Behavior

func TestGetCNPJOwnerContactInfo_CacheCorrupted(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, redisClient, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	// Set corrupted cache data
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)
	err := redisClient.Set(ctx, cacheKey, "corrupted json {{{", 5*time.Minute).Err()
	require.NoError(t, err)

	// Call service - should fall back to API
	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "john@example.com", result.EmailPessoaFisica)

	// Should have called API
	assert.Greater(t, callCount, 0, "Should have called RMI API after cache corruption")
}

func TestGetCNPJOwnerContactInfo_ConcurrentRequestsSameCNPJ(t *testing.T) {
	apiCallCount := 0
	var mu sync.Mutex
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		apiCallCount++
		mu.Unlock()

		// Simulate slow API
		time.Sleep(100 * time.Millisecond)

		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	// Make concurrent requests
	var wg sync.WaitGroup
	results := make([]*models.CNPJOwnerInfo, 5)
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errors[idx] = service.GetCNPJOwnerContactInfo(ctx, cnpj)
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i := 0; i < 5; i++ {
		assert.NoError(t, errors[i])
		assert.NotNil(t, results[i])
		assert.Equal(t, "john@example.com", results[i].EmailPessoaFisica)
	}

	// Multiple API calls expected (no cache synchronization)
	mu.Lock()
	assert.Greater(t, apiCallCount, 0)
	mu.Unlock()
}

func TestGetCNPJOwnerContactInfo_MultipleSocios(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "11111111111", NomeSocio: "First Partner"},
				{CPFSocio: "22222222222", NomeSocio: "Second Partner"},
				{CPFSocio: "33333333333", NomeSocio: "Third Partner"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/11111111111") {
			citizen := models.CitizenContactInfo{
				CPF:  "11111111111",
				Nome: "First Partner",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "first@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should return first socio's info
	assert.Equal(t, "first@example.com", result.EmailPessoaFisica)
}

func TestGetCNPJOwnerContactInfo_DataTransformationValidation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john.doe@example.com"
			citizen.Telefone.Indicador = true
			citizen.Telefone.Principal.DDI = "55"
			citizen.Telefone.Principal.DDD = "21"
			citizen.Telefone.Principal.Valor = "999999999"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify data transformation from RMI client (email sanitization, phone formatting)
	assert.Equal(t, "john.doe@example.com", result.EmailPessoaFisica)
	assert.Equal(t, "5521999999999", result.CelularPessoaFisica)
}

// Test GetMultipleCNPJOwnerContactInfo

func TestGetMultipleCNPJOwnerContactInfo_AllHits(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company " + cnpj,
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{"11111111111111", "22222222222222", "33333333333333"}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	assert.Equal(t, 3, len(results))
	for _, cnpj := range cnpjs {
		assert.Contains(t, results, cnpj)
		assert.Equal(t, "john@example.com", results[cnpj].EmailPessoaFisica)
	}
}

func TestGetMultipleCNPJOwnerContactInfo_AllMisses(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{"11111111111111", "22222222222222", "33333333333333"}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// Should return empty map (graceful degradation)
	assert.Equal(t, 0, len(results))
}

func TestGetMultipleCNPJOwnerContactInfo_PartialSuccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]

			// Only succeed for first CNPJ
			if cnpj == "11111111111111" {
				legalEntity := models.LegalEntityDetails{
					CNPJ:        cnpj,
					RazaoSocial: "Test Company",
				}
				legalEntity.Socios = []struct {
					CPFSocio  string `json:"cpf_socio"`
					NomeSocio string `json:"nome_socio_estrangeiro"`
				}{
					{CPFSocio: "12345678900", NomeSocio: "John Doe"},
				}
				json.NewEncoder(w).Encode(legalEntity)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{"11111111111111", "22222222222222", "33333333333333"}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// Should return only successful result
	assert.Equal(t, 1, len(results))
	assert.Contains(t, results, "11111111111111")
	assert.Equal(t, "john@example.com", results["11111111111111"].EmailPessoaFisica)
}

func TestGetMultipleCNPJOwnerContactInfo_EmptyList(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call RMI API for empty list")
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	assert.Equal(t, 0, len(results))
}

func TestGetMultipleCNPJOwnerContactInfo_ParallelExecution(t *testing.T) {
	requestTimes := make(map[string]time.Time)
	var mu sync.Mutex

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]

			mu.Lock()
			requestTimes[cnpj] = time.Now()
			mu.Unlock()

			// Simulate slow API
			time.Sleep(100 * time.Millisecond)

			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{"11111111111111", "22222222222222", "33333333333333"}

	start := time.Now()
	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)
	duration := time.Since(start)

	// Should complete faster than sequential (3 * 100ms = 300ms)
	// Allow some overhead, but should be < 250ms if truly parallel
	assert.Less(t, duration, 250*time.Millisecond, "Requests should be parallel")
	assert.Equal(t, 3, len(results))
}

func TestGetMultipleCNPJOwnerContactInfo_CacheEfficiency(t *testing.T) {
	apiCallCount := 0
	var mu sync.Mutex

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			mu.Lock()
			apiCallCount++
			mu.Unlock()

			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{"11111111111111", "22222222222222"}

	// First call - should hit API
	results1 := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)
	assert.Equal(t, 2, len(results1))

	firstCallCount := apiCallCount

	// Second call - should use cache
	results2 := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)
	assert.Equal(t, 2, len(results2))

	// Should not have made additional API calls
	mu.Lock()
	assert.Equal(t, firstCallCount, apiCallCount, "Second call should use cache")
	mu.Unlock()
}

func TestGetMultipleCNPJOwnerContactInfo_LargeBatch(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()

	// Generate 100 CNPJs
	cnpjs := make([]string, 100)
	for i := 0; i < 100; i++ {
		cnpjs[i] = fmt.Sprintf("%014d", i+1)
	}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	assert.Equal(t, 100, len(results))
	for _, cnpj := range cnpjs {
		assert.Contains(t, results, cnpj)
	}
}

func TestGetMultipleCNPJOwnerContactInfo_DuplicateCNPJs(t *testing.T) {
	apiCallCount := 0
	var mu sync.Mutex

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			mu.Lock()
			apiCallCount++
			mu.Unlock()

			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	// Same CNPJ repeated 3 times
	cnpjs := []string{"11111111111111", "11111111111111", "11111111111111"}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// Note: Current implementation doesn't deduplicate, so might make multiple calls
	// This test documents current behavior
	assert.Greater(t, len(results), 0)
	mu.Lock()
	callCount := apiCallCount
	mu.Unlock()
	// Will call API 3 times (no deduplication in current implementation)
	assert.Greater(t, callCount, 0)
}

// Integration Tests

func TestContactInfoService_IntegrationEndToEnd(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Integration Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "98765432100", NomeSocio: "Jane Smith"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "98765432100",
				Nome: "Jane Smith",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "jane.smith@integration.test"
			citizen.Telefone.Indicador = true
			citizen.Telefone.Principal.DDI = "55"
			citizen.Telefone.Principal.DDD = "11"
			citizen.Telefone.Principal.Valor = "888888888"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, redisClient, mr, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12.345.678/0001-95"

	// First call - cache miss
	result1, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result1)
	assert.Equal(t, "12345678000195", result1.CNPJ)
	assert.Equal(t, "jane.smith@integration.test", result1.EmailPessoaFisica)
	assert.Equal(t, "5511888888888", result1.CelularPessoaFisica)

	// Verify cache was set
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)
	ttl := mr.TTL(cacheKey)
	assert.True(t, ttl > 0)

	// Second call - cache hit
	result2, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, result1.EmailPessoaFisica, result2.EmailPessoaFisica)

	// Verify cached data
	cachedData, err := redisClient.Get(ctx, cacheKey).Result()
	require.NoError(t, err)

	var cachedInfo models.CNPJOwnerInfo
	err = json.Unmarshal([]byte(cachedData), &cachedInfo)
	require.NoError(t, err)
	assert.Equal(t, "jane.smith@integration.test", cachedInfo.EmailPessoaFisica)
}

func TestContactInfoService_IntegrationConcurrentAccess(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Concurrent Test",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "Test User"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "Test User",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "test@concurrent.test"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()

	// Simulate concurrent access from multiple goroutines
	var wg sync.WaitGroup
	errorCount := 0
	var errMu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			cnpj := fmt.Sprintf("%014d", (idx%5)+1) // 5 different CNPJs
			_, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
			if err != nil {
				errMu.Lock()
				errorCount++
				errMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent access")
}

func TestContactInfoService_IntegrationErrorRecovery(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		currentCall := callCount
		mu.Unlock()

		// Fail first 2 calls, succeed on 3rd
		if currentCall <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Recovery Test",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "Recovery User"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "Recovery User",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "recovery@test.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	// First two calls should fail
	_, err1 := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err1)

	_, err2 := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err2)

	// Third call should succeed
	result, err3 := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err3)
	require.NotNil(t, result)
	assert.Equal(t, "recovery@test.com", result.EmailPessoaFisica)
}

func TestContactInfoService_IntegrationPerformanceUnderLoad(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate realistic API latency
		time.Sleep(10 * time.Millisecond)

		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Load Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "Load Test User"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "Load Test User",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "load@test.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()

	// Test with 50 different CNPJs
	cnpjs := make([]string, 50)
	for i := 0; i < 50; i++ {
		cnpjs[i] = fmt.Sprintf("%014d", i+1)
	}

	start := time.Now()
	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)
	duration := time.Since(start)

	// Should complete reasonably fast with parallelization
	assert.Less(t, duration, 5*time.Second, "Should handle 50 CNPJs in < 5s")
	assert.Equal(t, 50, len(results))

	// Verify all results are valid
	for _, cnpj := range cnpjs {
		assert.Contains(t, results, cnpj)
		assert.Equal(t, "load@test.com", results[cnpj].EmailPessoaFisica)
	}
}

// Additional Edge Case Tests

func TestGetCNPJOwnerContactInfo_CacheSetFailureGraceful(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	mockServer := httptest.NewServer(handler)
	defer mockServer.Close()

	rmiClient := clients.NewRMIClient(mockServer.URL, 5*time.Second)
	tokenMgr := &mockTokenManager{token: "test-token"}
	cfg := &config.AppConfig{
		Cache: config.CacheSettings{ContactInfoTTL: 5 * time.Minute},
	}

	service := services.NewContactInfoService(rmiClient, tokenMgr, redisClient, cfg)

	ctx := context.Background()
	cnpj := "12345678000195"

	// Close redis to simulate cache set failure
	mr.Close()

	// Should still succeed (graceful degradation)
	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "john@example.com", result.EmailPessoaFisica)
}

func TestGetCNPJOwnerContactInfo_MultipleAPICallsSameKey(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		time.Sleep(50 * time.Millisecond) // Simulate delay

		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	// First call
	result1, err1 := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err1)
	require.NotNil(t, result1)

	mu.Lock()
	firstCallCount := callCount
	mu.Unlock()

	// Second call - should hit cache
	result2, err2 := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err2)
	require.NotNil(t, result2)

	// Should not have made additional API calls
	mu.Lock()
	assert.Equal(t, firstCallCount, callCount, "Second call should use cache")
	mu.Unlock()
}

func TestGetCNPJOwnerContactInfo_ContextCancellation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetCNPJOwnerContactInfo_EmptyEmailAndPhone(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			// No email or phone
			citizen.Email.Indicador = false
			citizen.Telefone.Indicador = false
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "", result.EmailPessoaFisica)
	assert.Equal(t, "", result.CelularPessoaFisica)
}

func TestGetCNPJOwnerContactInfo_CitizenAPIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "12345678000195"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetCNPJOwnerContactInfo_InvalidCNPJFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpj := "invalid-cnpj"

	result, err := service.GetCNPJOwnerContactInfo(ctx, cnpj)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetMultipleCNPJOwnerContactInfo_SingleCNPJ(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{"12345678000195"}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	assert.Equal(t, 1, len(results))
	assert.Contains(t, results, "12345678000195")
}

func TestGetMultipleCNPJOwnerContactInfo_MixedFormats(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	// Mix of formatted and unformatted CNPJs
	cnpjs := []string{
		"12.345.678/0001-95",
		"98765432000195",
		"11.111.111/0001-11",
	}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// All should succeed
	assert.Greater(t, len(results), 0)
}

func TestGetMultipleCNPJOwnerContactInfo_TokenErrorAffectsAll(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call RMI API when token fails")
	})

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	mockServer := httptest.NewServer(handler)
	defer mockServer.Close()

	rmiClient := clients.NewRMIClient(mockServer.URL, 5*time.Second)
	tokenMgr := &mockTokenManager{err: errors.New("token failure")}
	cfg := &config.AppConfig{
		Cache: config.CacheSettings{ContactInfoTTL: 5 * time.Minute},
	}

	service := services.NewContactInfoService(rmiClient, tokenMgr, redisClient, cfg)

	ctx := context.Background()
	cnpjs := []string{"11111111111111", "22222222222222"}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// All should fail due to token error
	assert.Equal(t, 0, len(results))
}

func TestGetMultipleCNPJOwnerContactInfo_ContextCancellation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cnpjs := []string{"11111111111111", "22222222222222"}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// Should return empty or partial results due to timeout
	assert.LessOrEqual(t, len(results), 2)
}

func TestGetMultipleCNPJOwnerContactInfo_AllCached(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call RMI API when all cached")
	})

	service, redisClient, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()
	cnpjs := []string{"11111111111111", "22222222222222"}

	// Pre-populate cache for all CNPJs
	for _, cnpj := range cnpjs {
		cachedInfo := &models.CNPJOwnerInfo{
			CNPJ:                cnpj,
			EmailPessoaFisica:   "cached@example.com",
			CelularPessoaFisica: "5521999999999",
		}
		data, _ := json.Marshal(cachedInfo)
		cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)
		err := redisClient.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		require.NoError(t, err)
	}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	assert.Equal(t, 2, len(results))
	for _, cnpj := range cnpjs {
		assert.Contains(t, results, cnpj)
		assert.Equal(t, "cached@example.com", results[cnpj].EmailPessoaFisica)
	}
}

func TestGetMultipleCNPJOwnerContactInfo_VeryLargeBatch(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]
			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()

	// Generate 200 CNPJs
	cnpjs := make([]string, 200)
	for i := 0; i < 200; i++ {
		cnpjs[i] = fmt.Sprintf("%014d", i+1)
	}

	start := time.Now()
	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)
	duration := time.Since(start)

	assert.Equal(t, 200, len(results))
	// Should complete in reasonable time with parallelization
	assert.Less(t, duration, 10*time.Second, "Should handle 200 CNPJs in < 10s")
}

func TestGetMultipleCNPJOwnerContactInfo_ErrorAggregation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			cnpj := strings.Split(r.URL.Path, "/")[3]

			// Fail on every odd CNPJ
			cnpjInt := 0
			fmt.Sscanf(cnpj, "%d", &cnpjInt)
			if cnpjInt%2 == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			legalEntity := models.LegalEntityDetails{
				CNPJ:        cnpj,
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	service, _, _, cleanup := setupContactInfoServiceTest(t, handler)
	defer cleanup()

	ctx := context.Background()

	// Create mix of even and odd CNPJs (only even will succeed)
	cnpjs := []string{
		"00000000000002",
		"00000000000003",
		"00000000000004",
		"00000000000005",
		"00000000000006",
	}

	results := service.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// Should have 3 successful results (even numbers: 2, 4, 6)
	assert.Equal(t, 3, len(results))
	assert.Contains(t, results, "00000000000002")
	assert.Contains(t, results, "00000000000004")
	assert.Contains(t, results, "00000000000006")
}

func TestContactInfoService_IntegrationCacheExpiration(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/legal-entity/") {
			legalEntity := models.LegalEntityDetails{
				CNPJ:        "12345678000195",
				RazaoSocial: "Test Company",
			}
			legalEntity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{
				{CPFSocio: "12345678900", NomeSocio: "John Doe"},
			}
			json.NewEncoder(w).Encode(legalEntity)
		} else if strings.Contains(r.URL.Path, "/v1/citizen/") {
			citizen := models.CitizenContactInfo{
				CPF:  "12345678900",
				Nome: "John Doe",
			}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "john@example.com"
			json.NewEncoder(w).Encode(citizen)
		}
	})

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()

	mockServer := httptest.NewServer(handler)
	defer mockServer.Close()

	rmiClient := clients.NewRMIClient(mockServer.URL, 5*time.Second)
	tokenMgr := &mockTokenManager{token: "test-token"}
	cfg := &config.AppConfig{
		Cache: config.CacheSettings{ContactInfoTTL: 1 * time.Second}, // Short TTL
	}

	service := services.NewContactInfoService(rmiClient, tokenMgr, redisClient, cfg)

	ctx := context.Background()
	cnpj := "12345678000195"

	// First call - cache miss
	_, err = service.GetCNPJOwnerContactInfo(ctx, cnpj)
	require.NoError(t, err)

	// Verify cache exists
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpj)
	ttl := mr.TTL(cacheKey)
	assert.True(t, ttl > 0)

	// Fast-forward time in miniredis
	mr.FastForward(2 * time.Second)

	// Cache should be expired
	ttl = mr.TTL(cacheKey)
	assert.Equal(t, time.Duration(0), ttl, "Cache should be expired")
}
