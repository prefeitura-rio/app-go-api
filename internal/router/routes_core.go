package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/wire"
)

// registerCoreRoutes registers all core (non-empregabilidade) API routes.
func registerCoreRoutes(apiV1, apiPublic *gin.RouterGroup, app *wire.ApplicationContainer, cfg *config.AppConfig) {
	registerLookup(apiV1.Group("/empregos"), app.EmpregoHandler)
	registerLookup(apiV1.Group("/acessibilidades"), app.AcessibilidadeHandler)
	registerLookup(apiV1.Group("/categorias"), app.CategoriaHandler)
	registerLookup(apiV1.Group("/empresas"), app.EmpresaHandler)
	registerLookup(apiV1.Group("/escolaridades"), app.EscolaridadeHandler)
	registerLookup(apiV1.Group("/instituicoes"), app.InstituicaoHandler)

	if app.TypesenseHandler != nil {
		apiV1.POST("/typesense/multi-search", app.TypesenseHandler.SearchMultiCollection)
		apiV1.POST("/typesense/cursos/search", app.TypesenseHandler.SearchCursos)
		apiV1.POST("/typesense/empregos/search", app.TypesenseHandler.SearchEmpregos)
		apiV1.Group("/typesense/collections").POST("/:collection/documents/search", app.TypesenseHandler.SearchDocuments)
	}

	cursoLoader := func(ctx context.Context, id int) (orgaoID string, found bool, err error) {
		curso, err := app.CursoService.GetByID(ctx, id)
		if err != nil {
			return "", false, err
		}
		if curso == nil {
			return "", false, nil
		}
		return curso.OrgaoID, true, nil
	}

	ownershipCheck := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseOwnershipCheck(cursoLoader))
	courseAuth := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseAuthorization())
	courseOrgaoInjector := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseOrgaoInjector())
	courseListFilter := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseListFilter())
	enrollmentListAccess := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseEnrollmentListAccess(cursoLoader))

	courses := apiV1.Group("/courses")
	{
		courses.POST("", courseAuth, courseOrgaoInjector, app.CourseHandler.Create)
		courses.POST("/draft", courseAuth, courseOrgaoInjector, app.CourseHandler.CreateDraft)
		courses.GET("", courseListFilter, app.CourseHandler.List)
		courses.GET("/drafts", courseListFilter, app.CourseHandler.ListDrafts)
		courses.GET("/:courseId", courseAuth, ownershipCheck, app.CourseHandler.GetByID)
		courses.DELETE("/:courseId", courseAuth, ownershipCheck, app.CourseHandler.Delete)
		courses.PUT("/:courseId", courseAuth, ownershipCheck, app.CourseHandler.Update)
		courses.PUT("/:courseId/send-to-review", courseAuth, ownershipCheck, app.CourseHandler.SendToReview)
		courses.PUT("/:courseId/approve", courseAuth, ownershipCheck, app.CourseHandler.Approve)
		courses.PUT("/:courseId/request-changes", courseAuth, ownershipCheck, app.CourseHandler.RequestChanges)
		courses.PUT("/:courseId/request-deletion", courseAuth, ownershipCheck, app.CourseHandler.RequestDeletion)
		courses.POST("/:courseId/enrollments", app.InscricaoHandler.Create)
		courses.POST("/:courseId/enrollments/manual", courseAuth, ownershipCheck, app.InscricaoHandler.CreateManual)
		courses.POST("/:courseId/enrollments/import", courseAuth, ownershipCheck, app.InscricaoHandler.Import)
		courses.GET("/:courseId/enrollments", enrollmentListAccess, app.InscricaoHandler.List)
		courses.PUT("/:courseId/enrollments/status", courseAuth, ownershipCheck, app.InscricaoHandler.UpdateStatus)
		courses.PUT("/:courseId/enrollments/:enrollmentId", app.InscricaoHandler.Update)
		courses.PUT("/:courseId/enrollments/:enrollmentId/status", courseAuth, ownershipCheck, app.InscricaoHandler.UpdateIndividualStatus)
		courses.GET("/:courseId/enrollments/:enrollmentId", app.InscricaoHandler.GetByID)
		courses.PUT("/:courseId/enrollments/:enrollmentId/certificate", courseAuth, ownershipCheck, app.InscricaoHandler.UpdateCertificate)
		courses.DELETE("/:courseId/enrollments/:enrollmentId", app.InscricaoHandler.Delete)
	}

	apiV1.Group("/jobs").GET("/:jobId/status", app.JobHandler.GetStatus)
	apiV1.Group("/users").GET("/:userId/courses", app.CourseHandler.ListByUser)
	enrollments := apiV1.Group("/enrollments")
	enrollments.GET("/user/:cpf", app.InscricaoHandler.ListByUser)
	enrollments.PUT("/:enrollmentId/schedule", app.InscricaoHandler.ChangeSchedule)

	oportunidades := apiV1.Group("/oportunidades-mei")
	{
		oportunidades.POST("", app.OportunidadeMEIHandler.Create)
		oportunidades.POST("/draft", app.OportunidadeMEIHandler.CreateDraft)
		oportunidades.GET("", app.OportunidadeMEIHandler.List)
		oportunidades.GET("/drafts", app.OportunidadeMEIHandler.ListDrafts)
		oportunidades.GET("/:id", app.OportunidadeMEIHandler.GetByID)
		oportunidades.PUT("/:id", app.OportunidadeMEIHandler.Update)
		oportunidades.PUT("/:id/publish", app.OportunidadeMEIHandler.Publish)
		oportunidades.DELETE("/:id", app.OportunidadeMEIHandler.Delete)
		oportunidades.POST("/:id/propostas", app.PropostaMEIHandler.Create)
		oportunidades.GET("/:id/propostas", app.PropostaMEIHandler.List)
		oportunidades.PUT("/:id/propostas/status", app.PropostaMEIHandler.UpdateStatusBulk)
		oportunidades.GET("/:id/propostas/:propostaId", app.PropostaMEIHandler.GetByID)
		oportunidades.PUT("/:id/propostas/:propostaId", app.PropostaMEIHandler.Update)
		oportunidades.PUT("/:id/propostas/:propostaId/status", app.PropostaMEIHandler.UpdateStatus)
		oportunidades.DELETE("/:id/propostas/:propostaId", app.PropostaMEIHandler.Delete)
	}
	apiV1.Group("/propostas-mei").GET("/por-empresa", app.PropostaMEIHandler.ListByMEIEmpresa)

	apiPublic.GET("/courses", app.CourseHandler.ListPublic)
	apiPublic.GET("/courses/:courseId", app.CourseHandler.GetByID)
	apiPublic.GET("/oportunidades-mei", app.OportunidadeMEIHandler.List)
	apiPublic.GET("/oportunidades-mei/:id", app.OportunidadeMEIHandler.GetByID)
}
