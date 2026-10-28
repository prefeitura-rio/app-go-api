package providers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/config"
)

func TestProvideRMIClient_UsesResilienceSettings(t *testing.T) {
	cfg := &config.AppConfig{RMI: config.RMISettings{
		BaseURL:                        "http://rmi.invalid",
		SecretariaCacheTTL:             3 * time.Minute,
		CircuitBreakerFailureThreshold: 4,
		CircuitBreakerOpenTimeout:      15 * time.Second,
	}}

	assert.Len(t, RMIClientOptions(cfg, "api"), 3)
	assert.NotNil(t, ProvideRMIClient(cfg))
}
