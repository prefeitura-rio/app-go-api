package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/wire"
)

// registerEmpregabilidadeRoutes registers all empregabilidade API routes.
func registerEmpregabilidadeRoutes(apiV1, apiPublic *gin.RouterGroup, app *wire.ApplicationContainer, cfg *config.AppConfig) {
	emp := apiV1.Group("/empregabilidade")

	// Lookup tables
	registerLookup(emp.Group("/regimes-contratacao"), app.EmpRegimeContratacaoHandler)
	registerLookup(emp.Group("/modelos-trabalho"), app.EmpModeloTrabalhoHandler)
	registerLookup(emp.Group("/tipos-pcd"), app.EmpTipoPCDHandler)
	registerLookup(emp.Group("/idiomas"), app.EmpIdiomaHandler)
	registerLookup(emp.Group("/niveis-idioma"), app.EmpNivelIdiomaHandler)
	registerLookup(emp.Group("/escolaridades"), app.EmpEscolaridadeHandler)
	registerLookup(emp.Group("/tipos-conquista"), app.EmpTipoConquistaHandler)
	registerLookup(emp.Group("/situacoes-atual"), app.EmpSituacaoAtualHandler)
	registerLookup(emp.Group("/disponibilidades"), app.EmpDisponibilidadeHandler)

	// Empresas
	empEmpresas := emp.Group("/empresas")
	{
		empEmpresas.POST("", app.EmpEmpresaHandler.Create)
		empEmpresas.GET("", app.EmpEmpresaHandler.List)
		empEmpresas.GET("/consulta-cnpj/:cnpj", app.EmpEmpresaHandler.ConsultaCNPJ)
		empEmpresas.GET("/:cnpj", app.EmpEmpresaHandler.GetByID)
		empEmpresas.PUT("/:cnpj", app.EmpEmpresaHandler.Update)
		empEmpresas.DELETE("/:cnpj", app.EmpEmpresaHandler.Delete)
	}

	var vagaAuth gin.HandlerFunc
	var vagaOrgaoInjector gin.HandlerFunc
	var vagaListFilter gin.HandlerFunc

	if cfg.App.Environment == "development" || cfg.App.Environment == "test" {
		vagaAuth = middlewares.VagaAuthorization()
		vagaOrgaoInjector = middlewares.VagaOrgaoInjector()
		vagaListFilter = middlewares.VagaListFilter()
	} else {
		vagaAuth = noOpHandler
		vagaOrgaoInjector = noOpHandler
		vagaListFilter = noOpHandler
	}

	// Vagas
	empVagas := emp.Group("/vagas")
	{
		empVagas.POST("", vagaAuth, vagaOrgaoInjector, app.EmpVagaHandler.Create)
		empVagas.POST("/draft", vagaAuth, vagaOrgaoInjector, app.EmpVagaHandler.Create)
		empVagas.GET("", vagaAuth, vagaListFilter, app.EmpVagaHandler.List)
		empVagas.GET("/:id", app.EmpVagaHandler.GetByID)
		empVagas.PUT("/:id", vagaAuth, app.EmpVagaHandler.Update)
		empVagas.DELETE("/:id", vagaAuth, app.EmpVagaHandler.Delete)
		empVagas.PUT("/:id/send-to-draft", vagaAuth, app.EmpVagaHandler.SendToDraft)
		empVagas.PUT("/:id/send-to-approval", vagaAuth, app.EmpVagaHandler.SendToApproval)
		empVagas.PUT("/:id/publish", vagaAuth, app.EmpVagaHandler.Publish)
		empVagas.PUT("/:id/freeze", vagaAuth, app.EmpVagaHandler.Freeze)
		empVagas.PUT("/:id/unfreeze", vagaAuth, app.EmpVagaHandler.Unfreeze)
		empVagas.PUT("/:id/discontinue", vagaAuth, app.EmpVagaHandler.Discontinue)
		empVagas.PUT("/:id/reactivate", vagaAuth, app.EmpVagaHandler.Reactivate)
		empVagas.PUT("/:id/tipos-pcd", vagaAuth, app.EmpVagaHandler.UpdateTiposPCD)
		empVagas.POST("/:id/etapas", vagaAuth, app.EmpEtapaHandler.Create)
		empVagas.GET("/:id/etapas", app.EmpEtapaHandler.ListByVaga)
		empVagas.GET("/:id/etapas/:etapaId", app.EmpEtapaHandler.GetByID)
		empVagas.PUT("/:id/etapas/:etapaId", vagaAuth, app.EmpEtapaHandler.Update)
		empVagas.DELETE("/:id/etapas/:etapaId", vagaAuth, app.EmpEtapaHandler.Delete)
	}

	// Candidaturas
	empCandidaturas := emp.Group("/candidaturas")
	{
		empCandidaturas.POST("", app.EmpCandidaturaHandler.Create)
		empCandidaturas.GET("", vagaAuth, app.EmpCandidaturaHandler.List)
		empCandidaturas.PUT("/bulk-status", vagaAuth, app.EmpCandidaturaHandler.BulkUpdateStatus)
		empCandidaturas.PUT("/bulk-etapa", vagaAuth, app.EmpCandidaturaHandler.BulkUpdateEtapa)
		empCandidaturas.GET("/usuario/:cpf", app.EmpCandidaturaHandler.ListByCPF)
		empCandidaturas.GET("/:id", app.EmpCandidaturaHandler.GetByID)
		empCandidaturas.PUT("/:id", app.EmpCandidaturaHandler.Update)
		empCandidaturas.DELETE("/:id", app.EmpCandidaturaHandler.Delete)
		empCandidaturas.PUT("/:id/status", vagaAuth, app.EmpCandidaturaHandler.UpdateStatus)
		empCandidaturas.PUT("/:id/approve", vagaAuth, app.EmpCandidaturaHandler.Approve)
		empCandidaturas.PUT("/:id/reject", vagaAuth, app.EmpCandidaturaHandler.Reject)
		empCandidaturas.PUT("/:id/etapa", vagaAuth, app.EmpCandidaturaHandler.UpdateEtapa)
	}

	// Curriculo
	registerEmpCurriculoRoutes(emp, app)

	// Onboarding
	empOnboarding := emp.Group("/onboarding")
	empOnboarding.GET("/:cpf", app.EmpOnboardingHandler.IsFirstLogin)
	empOnboarding.PUT("/:cpf/complete", app.EmpOnboardingHandler.MarkFirstLoginCompleted)

	// Termos de Uso
	empTermosUso := emp.Group("/termos-uso")
	empTermosUso.GET("/:cpf", app.EmpTermosUsoHandler.HasAcceptedTerms)
	empTermosUso.PUT("/:cpf/accept", app.EmpTermosUsoHandler.AcceptTerms)
	empTermosUso.GET("/:cpf/details", app.EmpTermosUsoHandler.GetDetails)

	// Public empregabilidade
	apiPublicEmp := apiPublic.Group("/empregabilidade")
	apiPublicEmp.GET("/vagas", app.EmpVagaHandler.PublicList)
	apiPublicEmp.GET("/vagas/slug/:slug", app.EmpVagaHandler.PublicGetBySlug)
	apiPublicEmp.GET("/vagas/:id", app.EmpVagaHandler.PublicGetByID)
}

// registerEmpCurriculoRoutes registers curriculo sub-routes.
func registerEmpCurriculoRoutes(emp *gin.RouterGroup, app *wire.ApplicationContainer) {
	c := emp.Group("/curriculo")
	c.GET("/:cpf", app.EmpCurriculoHandler.GetCurriculoCompleto)

	registerCurriculoSection(c, "formacoes", app.EmpCurriculoHandler.CreateFormacao,
		app.EmpCurriculoHandler.GetFormacaoByID, app.EmpCurriculoHandler.UpdateFormacao,
		app.EmpCurriculoHandler.DeleteFormacao, app.EmpCurriculoHandler.ListFormacoesByCPF,
		app.EmpCurriculoHandler.ReplaceAllFormacoesByCPF)

	registerCurriculoSection(c, "idiomas", app.EmpCurriculoHandler.CreateIdioma,
		app.EmpCurriculoHandler.GetIdiomaByID, app.EmpCurriculoHandler.UpdateIdioma,
		app.EmpCurriculoHandler.DeleteIdioma, app.EmpCurriculoHandler.ListIdiomasByCPF,
		app.EmpCurriculoHandler.ReplaceAllIdiomasByCPF)

	registerCurriculoSection(c, "cursos-complementares", app.EmpCurriculoHandler.CreateCursoComplementar,
		app.EmpCurriculoHandler.GetCursoComplementarByID, app.EmpCurriculoHandler.UpdateCursoComplementar,
		app.EmpCurriculoHandler.DeleteCursoComplementar, app.EmpCurriculoHandler.ListCursosComplementaresByCPF,
		app.EmpCurriculoHandler.ReplaceAllCursosComplementaresByCPF)

	registerCurriculoSection(c, "experiencias", app.EmpCurriculoHandler.CreateExperiencia,
		app.EmpCurriculoHandler.GetExperienciaByID, app.EmpCurriculoHandler.UpdateExperiencia,
		app.EmpCurriculoHandler.DeleteExperiencia, app.EmpCurriculoHandler.ListExperienciasByCPF,
		app.EmpCurriculoHandler.ReplaceAllExperienciasByCPF)

	registerCurriculoSection(c, "conquistas", app.EmpCurriculoHandler.CreateConquista,
		app.EmpCurriculoHandler.GetConquistaByID, app.EmpCurriculoHandler.UpdateConquista,
		app.EmpCurriculoHandler.DeleteConquista, app.EmpCurriculoHandler.ListConquistasByCPF,
		app.EmpCurriculoHandler.ReplaceAllConquistasByCPF)

	c.PUT("/situacao-interesses", app.EmpCurriculoHandler.UpsertSituacaoInteresses)
	c.GET("/:cpf/situacao-interesses", app.EmpCurriculoHandler.GetSituacaoInteressesByCPF)
}

// registerCurriculoSection registers CRUD + list-by-CPF + replace-all routes for one curriculo sub-resource.
func registerCurriculoSection(
	c *gin.RouterGroup,
	resource string,
	create, getByID, update, delete gin.HandlerFunc,
	listByCPF, replaceAll gin.HandlerFunc,
) {
	c.POST("/"+resource, create)
	c.GET("/"+resource+"/:id", getByID)
	c.PUT("/"+resource+"/:id", update)
	c.DELETE("/"+resource+"/:id", delete)
	c.GET("/:cpf/"+resource, listByCPF)
	c.PUT("/:cpf/"+resource, replaceAll)
}

// lookupHandlerRoutes is the common CRUD interface for all lookup handlers.
type lookupHandlerRoutes interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

// registerLookup registers standard CRUD routes on the given group.
func registerLookup(group *gin.RouterGroup, h lookupHandlerRoutes) {
	group.POST("", h.Create)
	group.GET("", h.List)
	group.GET("/:id", h.GetByID)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}
