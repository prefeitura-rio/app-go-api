package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenResponse represents the response from Keycloak token endpoint
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

// ServiceAccountTokenManager manages service account tokens with automatic refresh
type ServiceAccountTokenManager struct {
	keycloakURL  string
	realm        string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	// Token cache (in-memory, thread-safe)
	mu          sync.RWMutex
	cachedToken string
	tokenExpiry time.Time
}

// NewServiceAccountTokenManager creates a new service account token manager
func NewServiceAccountTokenManager(keycloakURL, realm, clientID, clientSecret string) *ServiceAccountTokenManager {
	return &ServiceAccountTokenManager{
		keycloakURL:  keycloakURL,
		realm:        realm,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetToken returns a valid token, fetching a new one if needed
// Implements double-checked locking pattern for thread safety
func (m *ServiceAccountTokenManager) GetToken(ctx context.Context) (string, error) {
	// Fast path: read lock to check if cached token is still valid
	m.mu.RLock()
	if m.cachedToken != "" && time.Now().Before(m.tokenExpiry) {
		token := m.cachedToken
		m.mu.RUnlock()
		return token, nil
	}
	m.mu.RUnlock()

	// Slow path: write lock to fetch new token
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have fetched it)
	if m.cachedToken != "" && time.Now().Before(m.tokenExpiry) {
		return m.cachedToken, nil
	}

	// Fetch new token
	return m.fetchNewToken(ctx)
}

// fetchNewToken fetches a new token from Keycloak using client credentials flow
// Must be called with write lock held
func (m *ServiceAccountTokenManager) fetchNewToken(ctx context.Context) (string, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		m.keycloakURL, m.realm)

	// Prepare form data for client credentials grant
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", m.clientID)
	data.Set("client_secret", m.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Cache token until 90% of its lifetime
	expiresIn := time.Duration(tokenResp.ExpiresIn) * time.Second
	refreshAt := time.Duration(float64(expiresIn) * 0.9)

	m.cachedToken = tokenResp.AccessToken
	m.tokenExpiry = time.Now().Add(refreshAt)

	log.Printf("[SERVICE_ACCOUNT_TOKEN] Token cached for %v (expires in %v, refreshing at 90%%)",
		refreshAt, expiresIn)

	return m.cachedToken, nil
}
