package router

import "github.com/gin-gonic/gin"

func registerEmpregabilidadeRoutes(apiV1 *gin.RouterGroup, deps *Dependencies) {
	empregabilidade := apiV1.Group("/empregabilidade")

	// Lookup tables
	registerCRUDRoutes(empregabilidade.Group("/regimes-contratacao"), deps.EmpRegimeContratacaoHandler)
	registerCRUDRoutes(empregabilidade.Group("/modelos-trabalho"), deps.EmpModeloTrabalhoHandler)
	registerCRUDRoutes(empregabilidade.Group("/tipos-pcd"), deps.EmpTipoPCDHandler)
	registerCRUDRoutes(empregabilidade.Group("/idiomas"), deps.EmpIdiomaHandler)
	registerCRUDRoutes(empregabilidade.Group("/niveis-idioma"), deps.EmpNivelIdiomaHandler)
	registerCRUDRoutes(empregabilidade.Group("/escolaridades"), deps.EmpEscolaridadeHandler)
	registerCRUDRoutes(empregabilidade.Group("/tipos-conquista"), deps.EmpTipoConquistaHandler)
	registerCRUDRoutes(empregabilidade.Group("/situacoes-atual"), deps.EmpSituacaoAtualHandler)
	registerCRUDRoutes(empregabilidade.Group("/disponibilidades"), deps.EmpDisponibilidadeHandler)

	// Empresas
	empEmpresas := empregabilidade.Group("/empresas")
	{
		empEmpresas.POST("", deps.EmpEmpresaHandler.Create)
		empEmpresas.GET("", deps.EmpEmpresaHandler.List)
		empEmpresas.GET("/consulta-cnpj/:cnpj", deps.EmpEmpresaHandler.ConsultaCNPJ)
		empEmpresas.GET("/:cnpj", deps.EmpEmpresaHandler.GetByID)
		empEmpresas.PUT("/:cnpj", deps.EmpEmpresaHandler.Update)
		empEmpresas.DELETE("/:cnpj", deps.EmpEmpresaHandler.Delete)
	}

	// Vagas
	empVagas := empregabilidade.Group("/vagas")
	{
		empVagas.POST("", deps.EmpVagaHandler.Create)
		empVagas.POST("/draft", deps.EmpVagaHandler.CreateDraft)
		empVagas.GET("", deps.EmpVagaHandler.List)
		empVagas.GET("/:id", deps.EmpVagaHandler.GetByID)
		empVagas.PUT("/:id", deps.EmpVagaHandler.Update)
		empVagas.DELETE("/:id", deps.EmpVagaHandler.Delete)
		empVagas.PUT("/:id/publish", deps.EmpVagaHandler.Publish)
		empVagas.PUT("/:id/tipos-pcd", deps.EmpVagaHandler.UpdateTiposPCD)

		// Etapas (nested)
		empVagas.POST("/:id/etapas", deps.EmpEtapaHandler.Create)
		empVagas.GET("/:id/etapas", deps.EmpEtapaHandler.ListByVaga)
		empVagas.GET("/:id/etapas/:etapaId", deps.EmpEtapaHandler.GetByID)
		empVagas.PUT("/:id/etapas/:etapaId", deps.EmpEtapaHandler.Update)
		empVagas.DELETE("/:id/etapas/:etapaId", deps.EmpEtapaHandler.Delete)
	}

	// Candidaturas
	empCandidaturas := empregabilidade.Group("/candidaturas")
	{
		empCandidaturas.POST("", deps.EmpCandidaturaHandler.Create)
		empCandidaturas.GET("", deps.EmpCandidaturaHandler.List)
		empCandidaturas.GET("/:id", deps.EmpCandidaturaHandler.GetByID)
		empCandidaturas.PUT("/:id", deps.EmpCandidaturaHandler.Update)
		empCandidaturas.DELETE("/:id", deps.EmpCandidaturaHandler.Delete)
		empCandidaturas.PUT("/:id/status", deps.EmpCandidaturaHandler.UpdateStatus)
		empCandidaturas.PUT("/:id/approve", deps.EmpCandidaturaHandler.Approve)
		empCandidaturas.PUT("/:id/reject", deps.EmpCandidaturaHandler.Reject)
	}

	// Currículo
	empCurriculo := empregabilidade.Group("/curriculo")
	{
		empCurriculo.GET("/:cpf", deps.EmpCurriculoHandler.GetCurriculoCompleto)

		empCurriculo.POST("/formacoes", deps.EmpCurriculoHandler.CreateFormacao)
		empCurriculo.GET("/formacoes/:id", deps.EmpCurriculoHandler.GetFormacaoByID)
		empCurriculo.PUT("/formacoes/:id", deps.EmpCurriculoHandler.UpdateFormacao)
		empCurriculo.DELETE("/formacoes/:id", deps.EmpCurriculoHandler.DeleteFormacao)
		empCurriculo.GET("/:cpf/formacoes", deps.EmpCurriculoHandler.ListFormacoesByCPF)

		empCurriculo.POST("/idiomas", deps.EmpCurriculoHandler.CreateIdioma)
		empCurriculo.GET("/idiomas/:id", deps.EmpCurriculoHandler.GetIdiomaByID)
		empCurriculo.PUT("/idiomas/:id", deps.EmpCurriculoHandler.UpdateIdioma)
		empCurriculo.DELETE("/idiomas/:id", deps.EmpCurriculoHandler.DeleteIdioma)
		empCurriculo.GET("/:cpf/idiomas", deps.EmpCurriculoHandler.ListIdiomasByCPF)

		empCurriculo.POST("/cursos-complementares", deps.EmpCurriculoHandler.CreateCursoComplementar)
		empCurriculo.GET("/cursos-complementares/:id", deps.EmpCurriculoHandler.GetCursoComplementarByID)
		empCurriculo.PUT("/cursos-complementares/:id", deps.EmpCurriculoHandler.UpdateCursoComplementar)
		empCurriculo.DELETE("/cursos-complementares/:id", deps.EmpCurriculoHandler.DeleteCursoComplementar)
		empCurriculo.GET("/:cpf/cursos-complementares", deps.EmpCurriculoHandler.ListCursosComplementaresByCPF)

		empCurriculo.POST("/experiencias", deps.EmpCurriculoHandler.CreateExperiencia)
		empCurriculo.GET("/experiencias/:id", deps.EmpCurriculoHandler.GetExperienciaByID)
		empCurriculo.PUT("/experiencias/:id", deps.EmpCurriculoHandler.UpdateExperiencia)
		empCurriculo.DELETE("/experiencias/:id", deps.EmpCurriculoHandler.DeleteExperiencia)
		empCurriculo.GET("/:cpf/experiencias", deps.EmpCurriculoHandler.ListExperienciasByCPF)

		empCurriculo.POST("/conquistas", deps.EmpCurriculoHandler.CreateConquista)
		empCurriculo.GET("/conquistas/:id", deps.EmpCurriculoHandler.GetConquistaByID)
		empCurriculo.PUT("/conquistas/:id", deps.EmpCurriculoHandler.UpdateConquista)
		empCurriculo.DELETE("/conquistas/:id", deps.EmpCurriculoHandler.DeleteConquista)
		empCurriculo.GET("/:cpf/conquistas", deps.EmpCurriculoHandler.ListConquistasByCPF)

		empCurriculo.PUT("/situacao-interesses", deps.EmpCurriculoHandler.UpsertSituacaoInteresses)
		empCurriculo.GET("/:cpf/situacao-interesses", deps.EmpCurriculoHandler.GetSituacaoInteressesByCPF)
	}

	// Onboarding
	empOnboarding := empregabilidade.Group("/onboarding")
	{
		empOnboarding.GET("/:cpf", deps.EmpOnboardingHandler.IsFirstLogin)
		empOnboarding.PUT("/:cpf/complete", deps.EmpOnboardingHandler.MarkFirstLoginCompleted)
	}

	// Termos de Uso
	empTermosUso := empregabilidade.Group("/termos-uso")
	{
		empTermosUso.GET("/:cpf", deps.EmpTermosUsoHandler.HasAcceptedTerms)
		empTermosUso.PUT("/:cpf/accept", deps.EmpTermosUsoHandler.AcceptTerms)
		empTermosUso.GET("/:cpf/details", deps.EmpTermosUsoHandler.GetDetails)
	}
}
