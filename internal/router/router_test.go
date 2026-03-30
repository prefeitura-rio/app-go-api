package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Set Gin to test mode for cleaner output
	gin.SetMode(gin.TestMode)
}

// TestRegisterLookup tests the registerLookup helper function
func TestRegisterLookup(t *testing.T) {
	router := gin.New()
	group := router.Group("/test")

	// Create a mock lookup handler
	mockHandler := &mockLookupHandler{}

	// Register routes
	registerLookup(group, mockHandler)

	// Verify routes are registered
	routes := router.Routes()
	assert.Len(t, routes, 5, "Should have 5 CRUD routes")

	// Verify each HTTP method
	methods := map[string]bool{
		"POST":   false,
		"GET":    false,
		"PUT":    false,
		"DELETE": false,
	}

	for _, route := range routes {
		methods[route.Method] = true
	}

	assert.True(t, methods["POST"], "Should have POST route")
	assert.True(t, methods["GET"], "Should have GET routes")
	assert.True(t, methods["PUT"], "Should have PUT route")
	assert.True(t, methods["DELETE"], "Should have DELETE route")
}

// TestRegisterCurriculoSection tests the registerCurriculoSection helper
func TestRegisterCurriculoSection(t *testing.T) {
	router := gin.New()
	group := router.Group("/curriculo")

	// Mock handlers
	create := func(c *gin.Context) {}
	getByID := func(c *gin.Context) {}
	update := func(c *gin.Context) {}
	delete := func(c *gin.Context) {}
	listByCPF := func(c *gin.Context) {}
	replaceAll := func(c *gin.Context) {}

	// Register section
	registerCurriculoSection(group, "formacoes", create, getByID, update, delete, listByCPF, replaceAll)

	// Verify routes are registered
	routes := router.Routes()
	assert.Len(t, routes, 6, "Should have 6 routes for a curriculo section")

	// Check paths
	expectedPaths := map[string]bool{
		"/curriculo/formacoes":     false,
		"/curriculo/formacoes/:id": false,
		"/curriculo/:cpf/formacoes": false,
	}

	for _, route := range routes {
		if _, exists := expectedPaths[route.Path]; exists {
			expectedPaths[route.Path] = true
		}
	}

	for path, found := range expectedPaths {
		assert.True(t, found, "Path %s should be registered", path)
	}
}

// TestLookupHandlerInterface tests the lookupHandlerRoutes interface
func TestLookupHandlerInterface(t *testing.T) {
	// Verify that our mock handler implements the interface
	var _ lookupHandlerRoutes = &mockLookupHandler{}
}

// TestRouteRegistrationPatterns tests common route registration patterns
func TestRouteRegistrationPatterns(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*gin.Engine)
		expectedCount int
		description   string
	}{
		{
			name: "Basic CRUD routes",
			setupFunc: func(r *gin.Engine) {
				group := r.Group("/test")
				registerLookup(group, &mockLookupHandler{})
			},
			expectedCount: 5,
			description:   "Standard CRUD operations (POST, GET, GET/:id, PUT/:id, DELETE/:id)",
		},
		{
			name: "Nested route groups",
			setupFunc: func(r *gin.Engine) {
				api := r.Group("/api")
				v1 := api.Group("/v1")
				v1.GET("/test", func(c *gin.Context) {})
			},
			expectedCount: 1,
			description:   "Nested groups should register routes correctly",
		},
		{
			name: "Multiple groups",
			setupFunc: func(r *gin.Engine) {
				group1 := r.Group("/group1")
				group1.GET("/resource", func(c *gin.Context) {})

				group2 := r.Group("/group2")
				group2.GET("/resource", func(c *gin.Context) {})
			},
			expectedCount: 2,
			description:   "Multiple groups should coexist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			tt.setupFunc(router)

			routes := router.Routes()
			assert.Len(t, routes, tt.expectedCount, tt.description)
		})
	}
}

// TestRouteMethodMapping tests that routes are mapped to correct HTTP methods
func TestRouteMethodMapping(t *testing.T) {
	router := gin.New()
	group := router.Group("/resources")

	registerLookup(group, &mockLookupHandler{})

	routes := router.Routes()

	// Map routes by method
	routesByMethod := make(map[string][]gin.RouteInfo)
	for _, route := range routes {
		routesByMethod[route.Method] = append(routesByMethod[route.Method], route)
	}

	// Verify expected methods
	assert.Len(t, routesByMethod["POST"], 1, "Should have 1 POST route (Create)")
	assert.Len(t, routesByMethod["GET"], 2, "Should have 2 GET routes (List + GetByID)")
	assert.Len(t, routesByMethod["PUT"], 1, "Should have 1 PUT route (Update)")
	assert.Len(t, routesByMethod["DELETE"], 1, "Should have 1 DELETE route (Delete)")
}

// TestRoutePathPatterns tests route path patterns
func TestRoutePathPatterns(t *testing.T) {
	router := gin.New()
	group := router.Group("/items")

	registerLookup(group, &mockLookupHandler{})

	routes := router.Routes()

	// Check for expected path patterns
	hasList := false
	hasGetByID := false
	hasUpdate := false
	hasDelete := false
	hasCreate := false

	for _, route := range routes {
		switch {
		case route.Method == "GET" && route.Path == "/items":
			hasList = true
		case route.Method == "GET" && route.Path == "/items/:id":
			hasGetByID = true
		case route.Method == "PUT" && route.Path == "/items/:id":
			hasUpdate = true
		case route.Method == "DELETE" && route.Path == "/items/:id":
			hasDelete = true
		case route.Method == "POST" && route.Path == "/items":
			hasCreate = true
		}
	}

	assert.True(t, hasList, "Should have List route")
	assert.True(t, hasGetByID, "Should have GetByID route")
	assert.True(t, hasUpdate, "Should have Update route")
	assert.True(t, hasDelete, "Should have Delete route")
	assert.True(t, hasCreate, "Should have Create route")
}

// TestRegisterLookupWithMultipleResources tests multiple lookup resources
func TestRegisterLookupWithMultipleResources(t *testing.T) {
	router := gin.New()

	// Register multiple lookup resources
	registerLookup(router.Group("/empregos"), &mockLookupHandler{})
	registerLookup(router.Group("/acessibilidades"), &mockLookupHandler{})
	registerLookup(router.Group("/categorias"), &mockLookupHandler{})

	routes := router.Routes()
	assert.Len(t, routes, 15, "Should have 5 routes × 3 resources = 15 routes")

	// Verify all resources are registered
	resourcePaths := []string{"/empregos", "/acessibilidades", "/categorias"}
	for _, resourcePath := range resourcePaths {
		found := false
		for _, route := range routes {
			if route.Path == resourcePath {
				found = true
				break
			}
		}
		assert.True(t, found, "Resource %s should be registered", resourcePath)
	}
}

// TestCurriculoSectionMultipleSections tests multiple curriculo sections
func TestCurriculoSectionMultipleSections(t *testing.T) {
	router := gin.New()
	group := router.Group("/curriculo")

	// Mock handlers
	create := func(c *gin.Context) {}
	getByID := func(c *gin.Context) {}
	update := func(c *gin.Context) {}
	delete := func(c *gin.Context) {}
	listByCPF := func(c *gin.Context) {}
	replaceAll := func(c *gin.Context) {}

	// Register multiple sections
	sections := []string{"formacoes", "idiomas", "cursos-complementares", "experiencias", "conquistas"}
	for _, section := range sections {
		registerCurriculoSection(group, section, create, getByID, update, delete, listByCPF, replaceAll)
	}

	routes := router.Routes()
	assert.Len(t, routes, 30, "Should have 6 routes × 5 sections = 30 routes")
}

// TestRouteParameterExtraction tests parameter extraction from routes
func TestRouteParameterExtraction(t *testing.T) {
	router := gin.New()
	group := router.Group("/items")

	registerLookup(group, &mockLookupHandler{})

	// Count routes with parameters
	parametricRoutes := 0
	for _, route := range router.Routes() {
		if route.Path == "/items/:id" {
			parametricRoutes++
		}
	}

	assert.Equal(t, 3, parametricRoutes, "Should have 3 parametric routes (GET, PUT, DELETE)")
}

// TestRoutesGroupNesting tests nested route groups
func TestRoutesGroupNesting(t *testing.T) {
	router := gin.New()

	// Create nested groups
	api := router.Group("/api")
	v1 := api.Group("/v1")
	empregabilidade := v1.Group("/empregabilidade")

	registerLookup(empregabilidade.Group("/vagas"), &mockLookupHandler{})

	routes := router.Routes()

	// Verify nested path
	foundNestedPath := false
	for _, route := range routes {
		if route.Path == "/api/v1/empregabilidade/vagas" && route.Method == "GET" {
			foundNestedPath = true
			break
		}
	}

	assert.True(t, foundNestedPath, "Should have nested route /api/v1/empregabilidade/vagas")
}

// TestRouteHandlerExecution tests that handlers are called correctly
func TestRouteHandlerExecution(t *testing.T) {
	router := gin.New()
	group := router.Group("/test")

	handler := &mockLookupHandler{}
	registerLookup(group, handler)

	tests := []struct {
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"POST", "/test", http.StatusOK, `"action":"create"`},
		{"GET", "/test", http.StatusOK, `"action":"list"`},
		{"GET", "/test/123", http.StatusOK, `"action":"getByID"`},
		{"PUT", "/test/123", http.StatusOK, `"action":"update"`},
		{"DELETE", "/test/123", http.StatusOK, `"action":"delete"`},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestRouteParameterPassing tests parameter values are passed correctly
func TestRouteParameterPassing(t *testing.T) {
	router := gin.New()
	group := router.Group("/items")

	handler := &mockLookupHandler{}
	registerLookup(group, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items/42", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"42"`)
}

// TestHTTPMethodSegregation tests that different methods are segregated correctly
func TestHTTPMethodSegregation(t *testing.T) {
	router := gin.New()
	group := router.Group("/resources")

	registerLookup(group, &mockLookupHandler{})

	// Test POST to create
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/resources", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"action":"create"`)

	// Test GET to list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/resources", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"action":"list"`)

	// Test GET with ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/resources/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"action":"getByID"`)

	// Test PUT
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/resources/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"action":"update"`)

	// Test DELETE
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/resources/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"action":"delete"`)
}

// TestInvalidHTTPMethod tests 404 for invalid methods
func TestInvalidHTTPMethod(t *testing.T) {
	router := gin.New()
	group := router.Group("/items")

	registerLookup(group, &mockLookupHandler{})

	// PATCH is not registered
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/items/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// mockLookupHandler implements lookupHandlerRoutes for testing
type mockLookupHandler struct{}

func (m *mockLookupHandler) Create(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "create"})
}

func (m *mockLookupHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "list"})
}

func (m *mockLookupHandler) GetByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "getByID", "id": c.Param("id")})
}

func (m *mockLookupHandler) Update(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "update", "id": c.Param("id")})
}

func (m *mockLookupHandler) Delete(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "delete", "id": c.Param("id")})
}
