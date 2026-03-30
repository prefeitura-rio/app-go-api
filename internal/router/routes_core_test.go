package router

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/wire"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestRegisterCoreRoutes tests the complete core routes registration
func TestRegisterCoreRoutes(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	// Expected counts:
	// - 6 lookup resources × 5 routes each = 30
	// - Typesense routes (when handler exists) = 4
	// - Courses group = 17 routes (7 course + 10 enrollment)
	// - Jobs group = 1 route
	// - Users group = 1 route
	// - Enrollments group = 2 routes
	// - Oportunidades MEI group = 15 routes (8 oportunidades + 7 propostas)
	// - Propostas MEI group = 1 route
	// - Public routes = 4 routes
	// Total = 75 routes
	expectedRoutes := 75
	assert.Len(t, routes, expectedRoutes, "Should have all core routes registered")
}

// TestRegisterCoreRoutes_LookupResources tests lookup resource routes
func TestRegisterCoreRoutes_LookupResources(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	lookupResources := []string{
		"/api/v1/empregos",
		"/api/v1/acessibilidades",
		"/api/v1/categorias",
		"/api/v1/empresas",
		"/api/v1/escolaridades",
		"/api/v1/instituicoes",
	}

	routes := router.Routes()

	// Each lookup resource should have 5 routes (POST, GET, GET/:id, PUT/:id, DELETE/:id)
	for _, resource := range lookupResources {
		t.Run(resource, func(t *testing.T) {
			routeCount := 0
			for _, route := range routes {
				if route.Path == resource || route.Path == resource+"/:id" {
					routeCount++
				}
			}
			assert.Equal(t, 5, routeCount, "Each lookup resource should have 5 routes")
		})
	}
}

// TestRegisterCoreRoutes_TypesenseRoutes tests Typesense routes registration
func TestRegisterCoreRoutes_TypesenseRoutes(t *testing.T) {
	t.Run("With TypesenseHandler", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockApplicationContainer()
		registerCoreRoutes(apiV1, apiPublic, app)

		routes := router.Routes()

		expectedTypesenseRoutes := []struct {
			method string
			path   string
		}{
			{"POST", "/api/v1/typesense/multi-search"},
			{"POST", "/api/v1/typesense/cursos/search"},
			{"POST", "/api/v1/typesense/empregos/search"},
			{"POST", "/api/v1/typesense/collections/:collection/documents/search"},
		}

		for _, expected := range expectedTypesenseRoutes {
			found := false
			for _, route := range routes {
				if route.Method == expected.method && route.Path == expected.path {
					found = true
					break
				}
			}
			assert.True(t, found, "Should have route %s %s", expected.method, expected.path)
		}
	})

	t.Run("Without TypesenseHandler", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockApplicationContainer()
		app.TypesenseHandler = nil
		registerCoreRoutes(apiV1, apiPublic, app)

		routes := router.Routes()

		// Verify no Typesense routes exist
		for _, route := range routes {
			assert.NotContains(t, route.Path, "/typesense", "Should not have Typesense routes when handler is nil")
		}
	})
}

// TestRegisterCoreRoutes_CoursesGroup tests courses route group
func TestRegisterCoreRoutes_CoursesGroup(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	expectedCourseRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/courses"},
		{"POST", "/api/v1/courses/draft"},
		{"PUT", "/api/v1/courses/:courseId"},
		{"GET", "/api/v1/courses"},
		{"GET", "/api/v1/courses/drafts"},
		{"GET", "/api/v1/courses/:courseId"},
		{"DELETE", "/api/v1/courses/:courseId"},
		{"POST", "/api/v1/courses/:courseId/enrollments"},
		{"POST", "/api/v1/courses/:courseId/enrollments/manual"},
		{"POST", "/api/v1/courses/:courseId/enrollments/import"},
		{"GET", "/api/v1/courses/:courseId/enrollments"},
		{"PUT", "/api/v1/courses/:courseId/enrollments/status"},
		{"PUT", "/api/v1/courses/:courseId/enrollments/:enrollmentId"},
		{"PUT", "/api/v1/courses/:courseId/enrollments/:enrollmentId/status"},
		{"GET", "/api/v1/courses/:courseId/enrollments/:enrollmentId"},
		{"PUT", "/api/v1/courses/:courseId/enrollments/:enrollmentId/certificate"},
		{"DELETE", "/api/v1/courses/:courseId/enrollments/:enrollmentId"},
	}

	for _, expected := range expectedCourseRoutes {
		t.Run(expected.method+" "+expected.path, func(t *testing.T) {
			found := false
			for _, route := range routes {
				if route.Method == expected.method && route.Path == expected.path {
					found = true
					break
				}
			}
			assert.True(t, found, "Should have route %s %s", expected.method, expected.path)
		})
	}
}

// TestRegisterCoreRoutes_JobsGroup tests jobs route group
func TestRegisterCoreRoutes_JobsGroup(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	found := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/api/v1/jobs/:jobId/status" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have GET /api/v1/jobs/:jobId/status route")
}

// TestRegisterCoreRoutes_UsersGroup tests users route group
func TestRegisterCoreRoutes_UsersGroup(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	found := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/api/v1/users/:userId/courses" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have GET /api/v1/users/:userId/courses route")
}

// TestRegisterCoreRoutes_EnrollmentsGroup tests enrollments route group
func TestRegisterCoreRoutes_EnrollmentsGroup(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	expectedEnrollmentRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/enrollments/user/:cpf"},
		{"PUT", "/api/v1/enrollments/:enrollmentId/schedule"},
	}

	for _, expected := range expectedEnrollmentRoutes {
		found := false
		for _, route := range routes {
			if route.Method == expected.method && route.Path == expected.path {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have route %s %s", expected.method, expected.path)
	}
}

// TestRegisterCoreRoutes_OportunidadesMEIGroup tests oportunidades-mei route group
func TestRegisterCoreRoutes_OportunidadesMEIGroup(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	expectedOportunidadesRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/oportunidades-mei"},
		{"POST", "/api/v1/oportunidades-mei/draft"},
		{"GET", "/api/v1/oportunidades-mei"},
		{"GET", "/api/v1/oportunidades-mei/drafts"},
		{"GET", "/api/v1/oportunidades-mei/:id"},
		{"PUT", "/api/v1/oportunidades-mei/:id"},
		{"PUT", "/api/v1/oportunidades-mei/:id/publish"},
		{"DELETE", "/api/v1/oportunidades-mei/:id"},
		{"POST", "/api/v1/oportunidades-mei/:id/propostas"},
		{"GET", "/api/v1/oportunidades-mei/:id/propostas"},
		{"PUT", "/api/v1/oportunidades-mei/:id/propostas/status"},
		{"GET", "/api/v1/oportunidades-mei/:id/propostas/:propostaId"},
		{"PUT", "/api/v1/oportunidades-mei/:id/propostas/:propostaId"},
		{"PUT", "/api/v1/oportunidades-mei/:id/propostas/:propostaId/status"},
		{"DELETE", "/api/v1/oportunidades-mei/:id/propostas/:propostaId"},
	}

	for _, expected := range expectedOportunidadesRoutes {
		t.Run(expected.method+" "+expected.path, func(t *testing.T) {
			found := false
			for _, route := range routes {
				if route.Method == expected.method && route.Path == expected.path {
					found = true
					break
				}
			}
			assert.True(t, found, "Should have route %s %s", expected.method, expected.path)
		})
	}
}

// TestRegisterCoreRoutes_PropostasMEIGroup tests propostas-mei route group
func TestRegisterCoreRoutes_PropostasMEIGroup(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	found := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/api/v1/propostas-mei/por-empresa" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have GET /api/v1/propostas-mei/por-empresa route")
}

// TestRegisterCoreRoutes_PublicRoutes tests public route group
func TestRegisterCoreRoutes_PublicRoutes(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	expectedPublicRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/public/courses"},
		{"GET", "/api/public/courses/:courseId"},
		{"GET", "/api/public/oportunidades-mei"},
		{"GET", "/api/public/oportunidades-mei/:id"},
	}

	for _, expected := range expectedPublicRoutes {
		found := false
		for _, route := range routes {
			if route.Method == expected.method && route.Path == expected.path {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have public route %s %s", expected.method, expected.path)
	}
}

// TestRegisterCoreRoutes_RouteMethodCounts tests HTTP method distribution
func TestRegisterCoreRoutes_RouteMethodCounts(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	methodCounts := make(map[string]int)
	for _, route := range routes {
		methodCounts[route.Method]++
	}

	assert.Greater(t, methodCounts["POST"], 0, "Should have POST routes")
	assert.Greater(t, methodCounts["GET"], 0, "Should have GET routes")
	assert.Greater(t, methodCounts["PUT"], 0, "Should have PUT routes")
	assert.Greater(t, methodCounts["DELETE"], 0, "Should have DELETE routes")
}

// TestRegisterCoreRoutes_RouteGrouping tests route organization by groups
func TestRegisterCoreRoutes_RouteGrouping(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	groupCounts := map[string]int{
		"/api/v1/courses":           0,
		"/api/v1/jobs":              0,
		"/api/v1/users":             0,
		"/api/v1/enrollments":       0,
		"/api/v1/oportunidades-mei": 0,
		"/api/v1/propostas-mei":     0,
		"/api/public/courses":       0,
		"/api/public/oportunidades": 0,
	}

	for _, route := range routes {
		for prefix := range groupCounts {
			if len(route.Path) >= len(prefix) && route.Path[:len(prefix)] == prefix {
				groupCounts[prefix]++
			}
		}
	}

	assert.Greater(t, groupCounts["/api/v1/courses"], 0, "Should have courses routes")
	assert.Greater(t, groupCounts["/api/v1/jobs"], 0, "Should have jobs routes")
	assert.Greater(t, groupCounts["/api/v1/users"], 0, "Should have users routes")
	assert.Greater(t, groupCounts["/api/v1/enrollments"], 0, "Should have enrollments routes")
	assert.Greater(t, groupCounts["/api/v1/oportunidades-mei"], 0, "Should have oportunidades-mei routes")
	assert.Greater(t, groupCounts["/api/v1/propostas-mei"], 0, "Should have propostas-mei routes")
	assert.Greater(t, groupCounts["/api/public/courses"], 0, "Should have public courses routes")
	assert.Greater(t, groupCounts["/api/public/oportunidades"], 0, "Should have public oportunidades routes")
}

// TestRegisterCoreRoutes_NestedResources tests nested resource routes
func TestRegisterCoreRoutes_NestedResources(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	// Test course enrollments nesting
	nestedEnrollmentRoutes := []string{
		"/api/v1/courses/:courseId/enrollments",
		"/api/v1/courses/:courseId/enrollments/manual",
		"/api/v1/courses/:courseId/enrollments/import",
		"/api/v1/courses/:courseId/enrollments/:enrollmentId",
		"/api/v1/courses/:courseId/enrollments/:enrollmentId/status",
		"/api/v1/courses/:courseId/enrollments/:enrollmentId/certificate",
	}

	for _, path := range nestedEnrollmentRoutes {
		found := false
		for _, route := range routes {
			if route.Path == path {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have nested enrollment route %s", path)
	}

	// Test oportunidades propostas nesting
	nestedPropostaRoutes := []string{
		"/api/v1/oportunidades-mei/:id/propostas",
		"/api/v1/oportunidades-mei/:id/propostas/:propostaId",
		"/api/v1/oportunidades-mei/:id/propostas/:propostaId/status",
	}

	for _, path := range nestedPropostaRoutes {
		found := false
		for _, route := range routes {
			if route.Path == path {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have nested proposta route %s", path)
	}
}

// TestRegisterCoreRoutes_ParameterizedRoutes tests routes with path parameters
func TestRegisterCoreRoutes_ParameterizedRoutes(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	parameterizedRoutes := []string{
		":id",
		":courseId",
		":enrollmentId",
		":jobId",
		":userId",
		":cpf",
		":propostaId",
		":collection",
	}

	for _, param := range parameterizedRoutes {
		found := false
		for _, route := range routes {
			if containsParameter(route.Path, param) {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have routes with parameter %s", param)
	}
}

// TestRegisterCoreRoutes_PublicVsPrivate tests public vs private route separation
func TestRegisterCoreRoutes_PublicVsPrivate(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")

	app := createMockApplicationContainer()
	registerCoreRoutes(apiV1, apiPublic, app)

	routes := router.Routes()

	publicCount := 0
	privateCount := 0

	for _, route := range routes {
		if len(route.Path) >= 11 && route.Path[:11] == "/api/public" {
			publicCount++
		} else if len(route.Path) >= 7 && route.Path[:7] == "/api/v1" {
			privateCount++
		}
	}

	assert.Equal(t, 4, publicCount, "Should have exactly 4 public routes")
	assert.Equal(t, 71, privateCount, "Should have exactly 71 private routes")
}

// Helper function to create a mock ApplicationContainer with all handlers
func createMockApplicationContainer() *wire.ApplicationContainer {
	return &wire.ApplicationContainer{
		EmpregoHandler:         &v1.EmpregoHandler{},
		AcessibilidadeHandler:  &v1.AcessibilidadeHandler{},
		CategoriaHandler:       &v1.CategoriaHandler{},
		EmpresaHandler:         &v1.EmpresaHandler{},
		EscolaridadeHandler:    &v1.EscolaridadeHandler{},
		InstituicaoHandler:     &v1.InstituicaoHandler{},
		TypesenseHandler:       &v1.TypesenseHandler{},
		CourseHandler:          &v1.CourseHandler{},
		InscricaoHandler:       &v1.InscricaoHandler{},
		JobHandler:             &v1.JobHandler{},
		OportunidadeMEIHandler: &v1.OportunidadeMEIHandler{},
		PropostaMEIHandler:     &v1.PropostaMEIHandler{},
	}
}

// Helper function to check if a path contains a specific parameter
func containsParameter(path, param string) bool {
	for i := 0; i < len(path); i++ {
		if i+len(param) <= len(path) && path[i:i+len(param)] == param {
			return true
		}
	}
	return false
}
