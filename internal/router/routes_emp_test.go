package router

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/wire"
)

// newTestEmpConfig returns a minimal AppConfig with RBAC enabled so that the
// real authorization middlewares (VagaAuthorization, VagaOrgaoInjector,
// VagaListFilter, VagaOwnershipCheck) are wired into the routes.
func newTestEmpConfig() *config.AppConfig {
	return &config.AppConfig{App: config.AppSettings{RBACEnabled: true}}
}

// createMockAppContainer creates a minimal ApplicationContainer with mock handlers.
// We only need the handlers to satisfy interface requirements for route registration,
// so we can use empty struct instances (services will be nil but that's OK for testing).
func createMockAppContainer() *wire.ApplicationContainer {
	return &wire.ApplicationContainer{
		// Lookup handlers - all implement lookupHandlerRoutes interface
		EmpRegimeContratacaoHandler: &empHandlers.RegimeContratacaoHandler{},
		EmpModeloTrabalhoHandler:    &empHandlers.ModeloTrabalhoHandler{},
		EmpTipoPCDHandler:           &empHandlers.TipoPCDHandler{},
		EmpIdiomaHandler:            &empHandlers.IdiomaHandler{},
		EmpNivelIdiomaHandler:       &empHandlers.NivelIdiomaHandler{},
		EmpEscolaridadeHandler:      &empHandlers.EscolaridadeHandler{},
		EmpTipoConquistaHandler:     &empHandlers.TipoConquistaHandler{},
		EmpSituacaoAtualHandler:     &empHandlers.SituacaoAtualHandler{},
		EmpDisponibilidadeHandler:   &empHandlers.DisponibilidadeHandler{},
		// Empregabilidade handlers
		EmpEmpresaHandler:     &empHandlers.EmpresaHandler{},
		EmpVagaHandler:        &empHandlers.VagaHandler{},
		EmpEtapaHandler:       &empHandlers.EtapaHandler{},
		EmpCandidaturaHandler: &empHandlers.CandidaturaHandler{},
		EmpCurriculoHandler:   &empHandlers.CurriculoHandler{},
		EmpOnboardingHandler:  &empHandlers.OnboardingHandler{},
		EmpTermosUsoHandler:   &empHandlers.TermosUsoHandler{},
	}
}

// routeExists checks if a route with the given method and path exists in the router.
func routeExists(routes gin.RoutesInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

func TestRegisterEmpregabilidadeRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Lookup routes registration", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		// Test regime contratacao lookup routes
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/regimes-contratacao") {
			t.Error("Expected POST /api/v1/empregabilidade/regimes-contratacao route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/regimes-contratacao") {
			t.Error("Expected GET /api/v1/empregabilidade/regimes-contratacao route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/regimes-contratacao/:id") {
			t.Error("Expected GET /api/v1/empregabilidade/regimes-contratacao/:id route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/regimes-contratacao/:id") {
			t.Error("Expected PUT /api/v1/empregabilidade/regimes-contratacao/:id route")
		}
		if !routeExists(routes, "DELETE", "/api/v1/empregabilidade/regimes-contratacao/:id") {
			t.Error("Expected DELETE /api/v1/empregabilidade/regimes-contratacao/:id route")
		}

		// Test modelos trabalho lookup routes
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/modelos-trabalho") {
			t.Error("Expected POST /api/v1/empregabilidade/modelos-trabalho route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/modelos-trabalho") {
			t.Error("Expected GET /api/v1/empregabilidade/modelos-trabalho route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/modelos-trabalho/:id") {
			t.Error("Expected GET /api/v1/empregabilidade/modelos-trabalho/:id route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/modelos-trabalho/:id") {
			t.Error("Expected PUT /api/v1/empregabilidade/modelos-trabalho/:id route")
		}
		if !routeExists(routes, "DELETE", "/api/v1/empregabilidade/modelos-trabalho/:id") {
			t.Error("Expected DELETE /api/v1/empregabilidade/modelos-trabalho/:id route")
		}

		// Test tipos PCD lookup routes
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/tipos-pcd") {
			t.Error("Expected POST /api/v1/empregabilidade/tipos-pcd route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/tipos-pcd") {
			t.Error("Expected GET /api/v1/empregabilidade/tipos-pcd route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/tipos-pcd/:id") {
			t.Error("Expected GET /api/v1/empregabilidade/tipos-pcd/:id route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/tipos-pcd/:id") {
			t.Error("Expected PUT /api/v1/empregabilidade/tipos-pcd/:id route")
		}
		if !routeExists(routes, "DELETE", "/api/v1/empregabilidade/tipos-pcd/:id") {
			t.Error("Expected DELETE /api/v1/empregabilidade/tipos-pcd/:id route")
		}

		// Test idiomas lookup routes
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/idiomas") {
			t.Error("Expected POST /api/v1/empregabilidade/idiomas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/idiomas") {
			t.Error("Expected GET /api/v1/empregabilidade/idiomas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/idiomas/:id") {
			t.Error("Expected GET /api/v1/empregabilidade/idiomas/:id route")
		}

		// Test niveis idioma lookup routes
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/niveis-idioma") {
			t.Error("Expected GET /api/v1/empregabilidade/niveis-idioma route")
		}

		// Test escolaridades lookup routes
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/escolaridades") {
			t.Error("Expected GET /api/v1/empregabilidade/escolaridades route")
		}

		// Test tipos conquista lookup routes
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/tipos-conquista") {
			t.Error("Expected GET /api/v1/empregabilidade/tipos-conquista route")
		}

		// Test situacoes atual lookup routes
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/situacoes-atual") {
			t.Error("Expected GET /api/v1/empregabilidade/situacoes-atual route")
		}

		// Test disponibilidades lookup routes
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/disponibilidades") {
			t.Error("Expected GET /api/v1/empregabilidade/disponibilidades route")
		}
	})

	t.Run("Empresas routes registration", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		if !routeExists(routes, "POST", "/api/v1/empregabilidade/empresas") {
			t.Error("Expected POST /api/v1/empregabilidade/empresas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/empresas") {
			t.Error("Expected GET /api/v1/empregabilidade/empresas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/empresas/consulta-cnpj/:cnpj") {
			t.Error("Expected GET /api/v1/empregabilidade/empresas/consulta-cnpj/:cnpj route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/empresas/:cnpj") {
			t.Error("Expected GET /api/v1/empregabilidade/empresas/:cnpj route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/empresas/:cnpj") {
			t.Error("Expected PUT /api/v1/empregabilidade/empresas/:cnpj route")
		}
		if !routeExists(routes, "DELETE", "/api/v1/empregabilidade/empresas/:cnpj") {
			t.Error("Expected DELETE /api/v1/empregabilidade/empresas/:cnpj route")
		}
	})

	t.Run("Vagas routes registration", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		// Basic CRUD
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/vagas") {
			t.Error("Expected POST /api/v1/empregabilidade/vagas route")
		}
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/vagas/draft") {
			t.Error("Expected POST /api/v1/empregabilidade/vagas/draft route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/vagas") {
			t.Error("Expected GET /api/v1/empregabilidade/vagas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/vagas/:id") {
			t.Error("Expected GET /api/v1/empregabilidade/vagas/:id route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id route")
		}
		if !routeExists(routes, "DELETE", "/api/v1/empregabilidade/vagas/:id") {
			t.Error("Expected DELETE /api/v1/empregabilidade/vagas/:id route")
		}

		// Workflow routes
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/send-to-draft") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/send-to-draft route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/send-to-approval") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/send-to-approval route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/publish") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/publish route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/freeze") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/freeze route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/unfreeze") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/unfreeze route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/discontinue") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/discontinue route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/reactivate") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/reactivate route")
		}

		// Tipos PCD update
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/tipos-pcd") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/tipos-pcd route")
		}

		// Etapas nested routes
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/vagas/:id/etapas") {
			t.Error("Expected POST /api/v1/empregabilidade/vagas/:id/etapas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/vagas/:id/etapas") {
			t.Error("Expected GET /api/v1/empregabilidade/vagas/:id/etapas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/vagas/:id/etapas/:etapaId") {
			t.Error("Expected GET /api/v1/empregabilidade/vagas/:id/etapas/:etapaId route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/vagas/:id/etapas/:etapaId") {
			t.Error("Expected PUT /api/v1/empregabilidade/vagas/:id/etapas/:etapaId route")
		}
		if !routeExists(routes, "DELETE", "/api/v1/empregabilidade/vagas/:id/etapas/:etapaId") {
			t.Error("Expected DELETE /api/v1/empregabilidade/vagas/:id/etapas/:etapaId route")
		}
	})

	t.Run("Candidaturas routes registration", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		// Basic CRUD
		if !routeExists(routes, "POST", "/api/v1/empregabilidade/candidaturas") {
			t.Error("Expected POST /api/v1/empregabilidade/candidaturas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/candidaturas") {
			t.Error("Expected GET /api/v1/empregabilidade/candidaturas route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/candidaturas/:id") {
			t.Error("Expected GET /api/v1/empregabilidade/candidaturas/:id route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/candidaturas/:id") {
			t.Error("Expected PUT /api/v1/empregabilidade/candidaturas/:id route")
		}
		if !routeExists(routes, "DELETE", "/api/v1/empregabilidade/candidaturas/:id") {
			t.Error("Expected DELETE /api/v1/empregabilidade/candidaturas/:id route")
		}

		// Bulk operations
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/candidaturas/bulk-status") {
			t.Error("Expected PUT /api/v1/empregabilidade/candidaturas/bulk-status route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/candidaturas/bulk-etapa") {
			t.Error("Expected PUT /api/v1/empregabilidade/candidaturas/bulk-etapa route")
		}

		// User-specific listing
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/candidaturas/usuario/:cpf") {
			t.Error("Expected GET /api/v1/empregabilidade/candidaturas/usuario/:cpf route")
		}

		// Status management
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/candidaturas/:id/status") {
			t.Error("Expected PUT /api/v1/empregabilidade/candidaturas/:id/status route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/candidaturas/:id/approve") {
			t.Error("Expected PUT /api/v1/empregabilidade/candidaturas/:id/approve route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/candidaturas/:id/reject") {
			t.Error("Expected PUT /api/v1/empregabilidade/candidaturas/:id/reject route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/candidaturas/:id/etapa") {
			t.Error("Expected PUT /api/v1/empregabilidade/candidaturas/:id/etapa route")
		}
	})

	t.Run("Onboarding routes registration", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		if !routeExists(routes, "GET", "/api/v1/empregabilidade/onboarding/:cpf") {
			t.Error("Expected GET /api/v1/empregabilidade/onboarding/:cpf route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/onboarding/:cpf/complete") {
			t.Error("Expected PUT /api/v1/empregabilidade/onboarding/:cpf/complete route")
		}
	})

	t.Run("Termos de Uso routes registration", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		if !routeExists(routes, "GET", "/api/v1/empregabilidade/termos-uso/:cpf") {
			t.Error("Expected GET /api/v1/empregabilidade/termos-uso/:cpf route")
		}
		if !routeExists(routes, "PUT", "/api/v1/empregabilidade/termos-uso/:cpf/accept") {
			t.Error("Expected PUT /api/v1/empregabilidade/termos-uso/:cpf/accept route")
		}
		if !routeExists(routes, "GET", "/api/v1/empregabilidade/termos-uso/:cpf/details") {
			t.Error("Expected GET /api/v1/empregabilidade/termos-uso/:cpf/details route")
		}
	})

	t.Run("Public API routes registration", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		if !routeExists(routes, "GET", "/api/public/empregabilidade/vagas") {
			t.Error("Expected GET /api/public/empregabilidade/vagas route")
		}
		if !routeExists(routes, "GET", "/api/public/empregabilidade/vagas/slug/:slug") {
			t.Error("Expected GET /api/public/empregabilidade/vagas/slug/:slug route")
		}
		if !routeExists(routes, "GET", "/api/public/empregabilidade/vagas/:id") {
			t.Error("Expected GET /api/public/empregabilidade/vagas/:id route")
		}
	})
}

func TestRegisterEmpCurriculoRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Curriculo main route registration", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		if !routeExists(routes, "GET", "/empregabilidade/curriculo/:cpf") {
			t.Error("Expected GET /empregabilidade/curriculo/:cpf route")
		}
	})

	t.Run("Formacoes section routes registration", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		if !routeExists(routes, "POST", "/empregabilidade/curriculo/formacoes") {
			t.Error("Expected POST /empregabilidade/curriculo/formacoes route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/formacoes/:id") {
			t.Error("Expected GET /empregabilidade/curriculo/formacoes/:id route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/formacoes/:id") {
			t.Error("Expected PUT /empregabilidade/curriculo/formacoes/:id route")
		}
		if !routeExists(routes, "DELETE", "/empregabilidade/curriculo/formacoes/:id") {
			t.Error("Expected DELETE /empregabilidade/curriculo/formacoes/:id route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/:cpf/formacoes") {
			t.Error("Expected GET /empregabilidade/curriculo/:cpf/formacoes route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/:cpf/formacoes") {
			t.Error("Expected PUT /empregabilidade/curriculo/:cpf/formacoes route")
		}
	})

	t.Run("Idiomas section routes registration", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		if !routeExists(routes, "POST", "/empregabilidade/curriculo/idiomas") {
			t.Error("Expected POST /empregabilidade/curriculo/idiomas route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/idiomas/:id") {
			t.Error("Expected GET /empregabilidade/curriculo/idiomas/:id route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/idiomas/:id") {
			t.Error("Expected PUT /empregabilidade/curriculo/idiomas/:id route")
		}
		if !routeExists(routes, "DELETE", "/empregabilidade/curriculo/idiomas/:id") {
			t.Error("Expected DELETE /empregabilidade/curriculo/idiomas/:id route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/:cpf/idiomas") {
			t.Error("Expected GET /empregabilidade/curriculo/:cpf/idiomas route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/:cpf/idiomas") {
			t.Error("Expected PUT /empregabilidade/curriculo/:cpf/idiomas route")
		}
	})

	t.Run("Cursos complementares section routes registration", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		if !routeExists(routes, "POST", "/empregabilidade/curriculo/cursos-complementares") {
			t.Error("Expected POST /empregabilidade/curriculo/cursos-complementares route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/cursos-complementares/:id") {
			t.Error("Expected GET /empregabilidade/curriculo/cursos-complementares/:id route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/cursos-complementares/:id") {
			t.Error("Expected PUT /empregabilidade/curriculo/cursos-complementares/:id route")
		}
		if !routeExists(routes, "DELETE", "/empregabilidade/curriculo/cursos-complementares/:id") {
			t.Error("Expected DELETE /empregabilidade/curriculo/cursos-complementares/:id route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/:cpf/cursos-complementares") {
			t.Error("Expected GET /empregabilidade/curriculo/:cpf/cursos-complementares route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/:cpf/cursos-complementares") {
			t.Error("Expected PUT /empregabilidade/curriculo/:cpf/cursos-complementares route")
		}
	})

	t.Run("Experiencias section routes registration", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		if !routeExists(routes, "POST", "/empregabilidade/curriculo/experiencias") {
			t.Error("Expected POST /empregabilidade/curriculo/experiencias route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/experiencias/:id") {
			t.Error("Expected GET /empregabilidade/curriculo/experiencias/:id route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/experiencias/:id") {
			t.Error("Expected PUT /empregabilidade/curriculo/experiencias/:id route")
		}
		if !routeExists(routes, "DELETE", "/empregabilidade/curriculo/experiencias/:id") {
			t.Error("Expected DELETE /empregabilidade/curriculo/experiencias/:id route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/:cpf/experiencias") {
			t.Error("Expected GET /empregabilidade/curriculo/:cpf/experiencias route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/:cpf/experiencias") {
			t.Error("Expected PUT /empregabilidade/curriculo/:cpf/experiencias route")
		}
	})

	t.Run("Conquistas section routes registration", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		if !routeExists(routes, "POST", "/empregabilidade/curriculo/conquistas") {
			t.Error("Expected POST /empregabilidade/curriculo/conquistas route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/conquistas/:id") {
			t.Error("Expected GET /empregabilidade/curriculo/conquistas/:id route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/conquistas/:id") {
			t.Error("Expected PUT /empregabilidade/curriculo/conquistas/:id route")
		}
		if !routeExists(routes, "DELETE", "/empregabilidade/curriculo/conquistas/:id") {
			t.Error("Expected DELETE /empregabilidade/curriculo/conquistas/:id route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/:cpf/conquistas") {
			t.Error("Expected GET /empregabilidade/curriculo/:cpf/conquistas route")
		}
		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/:cpf/conquistas") {
			t.Error("Expected PUT /empregabilidade/curriculo/:cpf/conquistas route")
		}
	})

	t.Run("Situacao e interesses routes registration", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		if !routeExists(routes, "PUT", "/empregabilidade/curriculo/situacao-interesses") {
			t.Error("Expected PUT /empregabilidade/curriculo/situacao-interesses route")
		}
		if !routeExists(routes, "GET", "/empregabilidade/curriculo/:cpf/situacao-interesses") {
			t.Error("Expected GET /empregabilidade/curriculo/:cpf/situacao-interesses route")
		}
	})
}

func TestEmpCurriculoSectionRouteCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Each curriculo section has 6 routes", func(t *testing.T) {
		router := gin.New()
		emp := router.Group("/empregabilidade")

		app := createMockAppContainer()
		registerEmpCurriculoRoutes(emp, app)

		routes := router.Routes()

		// Each section (formacoes, idiomas, cursos-complementares, experiencias, conquistas) has:
		// POST, GET/:id, PUT/:id, DELETE/:id, GET/:cpf/section, PUT/:cpf/section
		// 5 sections × 6 routes = 30 routes
		// Plus 2 routes for situacao-interesses: PUT, GET/:cpf
		// Plus 1 route for GET/:cpf (curriculo completo)
		// Total: 30 + 2 + 1 = 33 routes
		if len(routes) != 33 {
			t.Errorf("Expected 33 curriculo routes, got %d", len(routes))
		}
	})
}

func TestEmpregabilidadeRouteCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Total empregabilidade routes count", func(t *testing.T) {
		router := gin.New()
		apiV1 := router.Group("/api/v1")
		apiPublic := router.Group("/api/public")

		app := createMockAppContainer()
		cfg := newTestEmpConfig()
		registerEmpregabilidadeRoutes(apiV1, apiPublic, app, cfg)

		routes := router.Routes()

		// Verify we have a substantial number of routes registered
		// This is a sanity check - actual count may vary as routes are added
		if len(routes) < 50 {
			t.Errorf("Expected at least 50 empregabilidade routes, got %d", len(routes))
		}
	})
}
