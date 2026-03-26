package e2e_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestAPIEndpoints validates various API endpoints
func TestAPIEndpoints(t *testing.T) {
	baseURL := getBaseURL(t)
	client := &http.Client{Timeout: 10 * time.Second}

	tests := []struct {
		name           string
		method         string
		endpoint       string
		expectedStatus int
	}{
		{
			name:           "Health endpoint",
			method:         "GET",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Swagger docs",
			method:         "GET",
			endpoint:       "/swagger/index.html",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OpenAPI spec",
			method:         "GET",
			endpoint:       "/swagger/doc.json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API v1 acessibilidades list",
			method:         "GET",
			endpoint:       "/api/v1/acessibilidades",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API v1 categorias list",
			method:         "GET",
			endpoint:       "/api/v1/categorias",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API v1 escolaridades list",
			method:         "GET",
			endpoint:       "/api/v1/escolaridades",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, baseURL+tt.endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestHealthEndpointStructure validates the health endpoint response structure
func TestHealthEndpointStructure(t *testing.T) {
	baseURL := getBaseURL(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	// Verify expected fields
	if _, ok := health["status"]; !ok {
		t.Error("Health response missing 'status' field")
	}

	if _, ok := health["timestamp"]; !ok {
		t.Error("Health response missing 'timestamp' field")
	}
}

// TestCORSHeaders validates CORS headers are set correctly
func TestCORSHeaders(t *testing.T) {
	baseURL := getBaseURL(t)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("OPTIONS", baseURL+"/api/v1/acessibilidades", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("CORS preflight request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check CORS headers
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" && allowOrigin != "https://example.com" {
		t.Errorf("Expected CORS Allow-Origin header, got: %s", allowOrigin)
	}

	allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Expected CORS Allow-Methods header")
	}
}

// TestResponseTime validates API response time is acceptable
func TestResponseTime(t *testing.T) {
	baseURL := getBaseURL(t)
	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	resp, err := client.Get(baseURL + "/health")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Response should be under 2 seconds
	if duration > 2*time.Second {
		t.Errorf("Response time too slow: %v (expected < 2s)", duration)
	}

	t.Logf("Response time: %v", duration)
}

// TestSwaggerDocumentation validates Swagger documentation is available
func TestSwaggerDocumentation(t *testing.T) {
	baseURL := getBaseURL(t)
	client := &http.Client{Timeout: 10 * time.Second}

	// Test Swagger UI
	resp, err := client.Get(baseURL + "/swagger/index.html")
	if err != nil {
		t.Fatalf("Failed to fetch Swagger UI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected Swagger UI to return 200, got %d", resp.StatusCode)
	}

	// Test OpenAPI spec
	resp2, err := client.Get(baseURL + "/swagger/doc.json")
	if err != nil {
		t.Fatalf("Failed to fetch OpenAPI spec: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected OpenAPI spec to return 200, got %d", resp2.StatusCode)
	}

	// Validate it's valid JSON
	var spec map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&spec); err != nil {
		t.Fatalf("OpenAPI spec is not valid JSON: %v", err)
	}

	// Verify it has OpenAPI structure
	if _, ok := spec["openapi"]; !ok {
		if _, ok := spec["swagger"]; !ok {
			t.Error("OpenAPI spec missing version field")
		}
	}
}

// getBaseURL retrieves the base URL from environment variable
func getBaseURL(t *testing.T) string {
	baseURL := os.Getenv("TEST_BASE_URL")
	if baseURL == "" {
		t.Skip("TEST_BASE_URL not set, skipping E2E test")
	}
	return baseURL
}
