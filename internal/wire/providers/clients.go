package providers

import (
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
)

// ProvideRMIClient creates the RMI API client with a 15-second request timeout
func ProvideRMIClient(cfg *config.AppConfig) *clients.RMIClient {
	return clients.NewRMIClient(cfg.RMI.BaseURL, 15*time.Second)
}

// ProvideDataRelayClient creates the DataRelay client for email sending
func ProvideDataRelayClient(cfg *config.AppConfig) *clients.DataRelayClient {
	return clients.NewDataRelayClient(cfg.DataRelay.BaseURL, cfg.DataRelay.APIKey, 30*time.Second)
}

// ProvideServiceAccountTokenManager creates the Keycloak service account token manager.
// Returns nil when Keycloak is not configured, which is a valid no-op state.
func ProvideServiceAccountTokenManager(cfg *config.AppConfig) *auth.ServiceAccountTokenManager {
	if cfg.Keycloak.URL == "" || cfg.Keycloak.ClientID == "" || cfg.Keycloak.ClientSecret == "" {
		return nil
	}
	return auth.NewServiceAccountTokenManager(
		cfg.Keycloak.URL,
		cfg.Keycloak.Realm,
		cfg.Keycloak.ClientID,
		cfg.Keycloak.ClientSecret,
	)
}
