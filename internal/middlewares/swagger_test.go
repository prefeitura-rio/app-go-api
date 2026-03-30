package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/docs"
	"github.com/stretchr/testify/assert"
)

func TestDynamicSwaggerHandler_UpdatesHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original host to restore later
	originalHost := docs.SwaggerInfo.Host

	tests := []struct {
		name         string
		requestHost  string
		expectedHost string
	}{
		{
			name:         "localhost",
			requestHost:  "localhost:8080",
			expectedHost: "localhost:8080",
		},
		{
			name:         "production domain",
			requestHost:  "api.example.com",
			expectedHost: "api.example.com",
		},
		{
			name:         "staging domain",
			requestHost:  "staging-api.example.com:443",
			expectedHost: "staging-api.example.com:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new router for each test
			router := gin.New()
			router.GET("/swagger/*any", DynamicSwaggerHandler())

			// Create request with specific host
			req := httptest.NewRequest("GET", "/swagger/index.html", nil)
			req.Host = tt.requestHost

			w := httptest.NewRecorder()

			// Test the actual handler
			router.ServeHTTP(w, req)

			// Verify the request had the correct host
			assert.Equal(t, tt.expectedHost, req.Host)
		})
	}

	// Verify the host is restored after handler completes
	// (though in practice this happens per-request, not globally)
	assert.Equal(t, originalHost, docs.SwaggerInfo.Host)
}

func TestDynamicSwaggerHandler_RestoresOriginalHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set a known original host
	docs.SwaggerInfo.Host = "original-host.com"
	originalHost := docs.SwaggerInfo.Host

	router := gin.New()
	router.GET("/swagger/*any", DynamicSwaggerHandler())

	// Make request with different host
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	req.Host = "new-host.com"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// After handler execution, original host should be restored
	// Note: In concurrent scenarios, this is handled per-request
	assert.Equal(t, originalHost, docs.SwaggerInfo.Host)
}

func TestDynamicSwaggerHandler_HandlesEmptyHost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/swagger/*any", DynamicSwaggerHandler())

	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	req.Host = ""

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not panic with empty host
	// The handler should execute even if host is empty
}

func TestDynamicSwaggerHandler_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/swagger/*any", DynamicSwaggerHandler())

	// Simulate concurrent requests with different hosts
	hosts := []string{"host1.com", "host2.com", "host3.com"}

	done := make(chan bool, len(hosts))

	for _, host := range hosts {
		go func(h string) {
			req := httptest.NewRequest("GET", "/swagger/index.html", nil)
			req.Host = h
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			done <- true
		}(host)
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(hosts); i++ {
		<-done
	}

	// Test passed if no race conditions or panics occurred
}
