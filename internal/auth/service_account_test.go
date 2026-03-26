package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceAccountTokenManager(t *testing.T) {
	manager := NewServiceAccountTokenManager(
		"https://keycloak.example.com",
		"test-realm",
		"test-client",
		"test-secret",
	)

	assert.NotNil(t, manager)
	assert.Equal(t, "https://keycloak.example.com", manager.keycloakURL)
	assert.Equal(t, "test-realm", manager.realm)
	assert.Equal(t, "test-client", manager.clientID)
	assert.Equal(t, "test-secret", manager.clientSecret)
	assert.NotNil(t, manager.httpClient)
}

func TestGetToken_Success(t *testing.T) {
	// Create mock Keycloak server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/realms/test-realm/protocol/openid-connect/token", r.URL.Path)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		// Parse form data
		err := r.ParseForm()
		require.NoError(t, err)
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "test-client", r.FormValue("client_id"))
		assert.Equal(t, "test-secret", r.FormValue("client_secret"))

		// Return mock token response
		response := TokenResponse{
			AccessToken:      "mock-access-token-" + time.Now().Format("20060102150405"),
			ExpiresIn:        3600,
			RefreshExpiresIn: 0,
			TokenType:        "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := NewServiceAccountTokenManager(
		server.URL,
		"test-realm",
		"test-client",
		"test-secret",
	)

	ctx := context.Background()

	// First call should fetch token
	token1, err := manager.GetToken(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, token1)
	assert.Contains(t, token1, "mock-access-token")
	assert.Equal(t, 1, callCount)

	// Second call should use cached token
	token2, err := manager.GetToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, token1, token2)
	assert.Equal(t, 1, callCount) // Should not make another call
}

func TestGetToken_Expiration(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Return token with very short expiry (1 second)
		response := TokenResponse{
			AccessToken:      "token-" + time.Now().Format("150405"),
			ExpiresIn:        1, // 1 second
			RefreshExpiresIn: 0,
			TokenType:        "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := NewServiceAccountTokenManager(
		server.URL,
		"test-realm",
		"test-client",
		"test-secret",
	)

	ctx := context.Background()

	// Get first token
	token1, err := manager.GetToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Wait for token to expire (90% of 1 second = 0.9 seconds)
	time.Sleep(1 * time.Second)

	// Get second token - should fetch new one
	token2, err := manager.GetToken(ctx)
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2)
	assert.Equal(t, 2, callCount)
}

func TestGetToken_ConcurrentAccess(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		// Simulate slow token generation
		time.Sleep(100 * time.Millisecond)

		response := TokenResponse{
			AccessToken:      "shared-token",
			ExpiresIn:        3600,
			RefreshExpiresIn: 0,
			TokenType:        "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := NewServiceAccountTokenManager(
		server.URL,
		"test-realm",
		"test-client",
		"test-secret",
	)

	ctx := context.Background()

	// Launch multiple concurrent requests
	const numGoroutines = 10
	tokens := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)
	var wg sync.WaitGroup

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			token, err := manager.GetToken(ctx)
			tokens[index] = token
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// Verify all goroutines got a token
	for i := 0; i < numGoroutines; i++ {
		assert.NoError(t, errors[i])
		assert.Equal(t, "shared-token", tokens[i])
	}

	// Due to double-checked locking, only one request should have been made
	// (or a small number if timing is unlucky)
	mu.Lock()
	assert.LessOrEqual(t, callCount, 2, "Should minimize redundant token fetches")
	mu.Unlock()
}

func TestGetToken_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	manager := NewServiceAccountTokenManager(
		server.URL,
		"test-realm",
		"invalid-client",
		"invalid-secret",
	)

	ctx := context.Background()

	token, err := manager.GetToken(ctx)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "token request failed with status 401")
}

func TestGetToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	manager := NewServiceAccountTokenManager(
		server.URL,
		"test-realm",
		"test-client",
		"test-secret",
	)

	ctx := context.Background()

	token, err := manager.GetToken(ctx)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to decode token response")
}

func TestGetToken_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(5 * time.Second)

		response := TokenResponse{
			AccessToken:      "token",
			ExpiresIn:        3600,
			RefreshExpiresIn: 0,
			TokenType:        "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := NewServiceAccountTokenManager(
		server.URL,
		"test-realm",
		"test-client",
		"test-secret",
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	token, err := manager.GetToken(ctx)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestFetchNewToken_TokenCaching(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TokenResponse{
			AccessToken:      "cached-token",
			ExpiresIn:        1000, // 1000 seconds
			RefreshExpiresIn: 0,
			TokenType:        "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := NewServiceAccountTokenManager(
		server.URL,
		"test-realm",
		"test-client",
		"test-secret",
	)

	ctx := context.Background()

	// Fetch token
	token, err := manager.GetToken(ctx)
	require.NoError(t, err)
	assert.Equal(t, "cached-token", token)

	// Verify token is cached
	manager.mu.RLock()
	assert.Equal(t, "cached-token", manager.cachedToken)
	assert.True(t, manager.tokenExpiry.After(time.Now()))
	// Should be refreshed at 90% of lifetime (900 seconds from now)
	expectedExpiry := time.Now().Add(900 * time.Second)
	assert.WithinDuration(t, expectedExpiry, manager.tokenExpiry, 5*time.Second)
	manager.mu.RUnlock()
}
