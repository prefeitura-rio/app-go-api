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
			endpoint:       "/docs/index.html",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OpenAPI spec",
			method:         "GET",
			endpoint:       "/docs/doc.json",
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

	// The endpoint answers {"status":"ok"} — see internal/router/router.go
	status, ok := health["status"].(string)
	if !ok {
		t.Fatal("Health response missing 'status' field")
	}
	if status != "ok" {
		t.Errorf("Expected health status \"ok\", got %q", status)
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
	resp, err := client.Get(baseURL + "/docs/index.html")
	if err != nil {
		t.Fatalf("Failed to fetch Swagger UI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected Swagger UI to return 200, got %d", resp.StatusCode)
	}

	// Test OpenAPI spec
	resp2, err := client.Get(baseURL + "/docs/doc.json")
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

// TestBancoCurriculosRequiresAuthorization checks that the banco de currículos
// stays closed without credentials on a running environment. Its authorization
// does not depend on RBAC_ENABLED, so this holds in every environment.
func TestBancoCurriculosRequiresAuthorization(t *testing.T) {
	baseURL := getBaseURL(t)
	client := &http.Client{Timeout: 10 * time.Second}

	for _, endpoint := range []string{
		"/api/v1/empregabilidade/banco-curriculos",
		"/api/v1/empregabilidade/banco-curriculos/11111111111",
	} {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := client.Get(baseURL + endpoint)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
				t.Errorf("Expected 401 or 403 without credentials, got %d", resp.StatusCode)
			}
		})
	}
}

// TestBancoCurriculosFichaContract checks the running API publishes the ficha
// contract the admin consumes: the citizen data (e-mail, raça, PCD), the last
// update date required by the perfil detalhado card and the situação e
// interesses section of the resume. It guards against changing
// the model without regenerating the Swagger docs, which is what the frontend
// client is generated from.
func TestBancoCurriculosFichaContract(t *testing.T) {
	baseURL := getBaseURL(t)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(baseURL + "/docs/doc.json")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 for the OpenAPI spec, got %d", resp.StatusCode)
	}

	var spec struct {
		Definitions map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"definitions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		t.Fatalf("Failed to decode OpenAPI spec: %v", err)
	}

	// A ficha lê também o currículo completo e, dentro dele, a situação e os
	// interesses do cidadão, com as descrições já resolvidas pela API.
	contrato := []struct {
		definition string
		fields     []string
	}{
		{"empregabilidade.BancoCurriculoDetalhe", []string{
			"cpf", "nome", "nome_social", "data_inclusao", "data_atualizacao",
			"profissao", "escolaridade", "bairro", "celular", "email", "genero",
			"raca", "deficiencia", "idade", "curriculo",
		}},
		{"empregabilidade.CurriculoCompleto", []string{"situacao_interesses"}},
		{"empregabilidade.CurriculoSituacaoInteresses", []string{
			"situacao", "disponibilidade", "tempo_procurando_emprego",
			"ids_tipos_vinculo_preferencia",
		}},
		{"empregabilidade.SituacaoAtual", []string{"descricao"}},
		{"empregabilidade.Disponibilidade", []string{"descricao"}},
	}

	for _, item := range contrato {
		definition, ok := spec.Definitions[item.definition]
		if !ok {
			t.Errorf("OpenAPI spec missing %s", item.definition)
			continue
		}
		for _, field := range item.fields {
			if _, ok := definition.Properties[field]; !ok {
				t.Errorf("Ficha contract missing field %s.%s", item.definition, field)
			}
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
