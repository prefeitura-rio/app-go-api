package router

import "github.com/gin-gonic/gin"

func registerMEIRoutes(apiV1 *gin.RouterGroup, deps *Dependencies) {
	oportunidadesMEI := apiV1.Group("/oportunidades-mei")
	{
		oportunidadesMEI.POST("", deps.OportunidadeMEIHandler.Create)
		oportunidadesMEI.POST("/draft", deps.OportunidadeMEIHandler.CreateDraft)
		oportunidadesMEI.GET("", deps.OportunidadeMEIHandler.List)
		oportunidadesMEI.GET("/drafts", deps.OportunidadeMEIHandler.ListDrafts)
		oportunidadesMEI.GET("/:id", deps.OportunidadeMEIHandler.GetByID)
		oportunidadesMEI.PUT("/:id", deps.OportunidadeMEIHandler.Update)
		oportunidadesMEI.PUT("/:id/publish", deps.OportunidadeMEIHandler.Publish)
		oportunidadesMEI.DELETE("/:id", deps.OportunidadeMEIHandler.Delete)

		// Propostas MEI (nested)
		oportunidadesMEI.POST("/:id/propostas", deps.PropostaMEIHandler.Create)
		oportunidadesMEI.GET("/:id/propostas", deps.PropostaMEIHandler.List)
		oportunidadesMEI.PUT("/:id/propostas/status", deps.PropostaMEIHandler.UpdateStatusBulk)
		oportunidadesMEI.GET("/:id/propostas/:propostaId", deps.PropostaMEIHandler.GetByID)
		oportunidadesMEI.PUT("/:id/propostas/:propostaId", deps.PropostaMEIHandler.Update)
		oportunidadesMEI.PUT("/:id/propostas/:propostaId/status", deps.PropostaMEIHandler.UpdateStatus)
		oportunidadesMEI.DELETE("/:id/propostas/:propostaId", deps.PropostaMEIHandler.Delete)
	}

	propostasMEI := apiV1.Group("/propostas-mei")
	{
		propostasMEI.GET("/por-empresa", deps.PropostaMEIHandler.ListByMEIEmpresa)
	}
}
