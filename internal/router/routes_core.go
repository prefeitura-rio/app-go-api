package router

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/wire"
)

// registerTypesenseRoutes registers the Typesense search routes. Every route
// here must carry a Swagger @Router annotation: Heimdall builds its
// endpoint→action mappings from the generated spec and denies unmapped routes.
func registerTypesenseRoutes(apiV1 *gin.RouterGroup, h *v1.TypesenseHandler) {
	apiV1.POST("/typesense/multi-search", h.SearchMultiCollection)
	apiV1.POST("/typesense/cursos/search", h.SearchCursos)
	apiV1.POST("/typesense/empregos/search", h.SearchEmpregos)
	apiV1.Group("/typesense/collections").POST("/:collection/documents/search", h.SearchDocuments)
}

// registerCoreRoutes registers all core (non-empregabilidade) API routes.
func registerCoreRoutes(apiV1, apiPublic *gin.RouterGroup, app *wire.ApplicationContainer, cfg *config.AppConfig) {
	registerLookup(apiV1.Group("/empregos"), app.EmpregoHandler)
	registerLookup(apiV1.Group("/acessibilidades"), app.AcessibilidadeHandler)
	registerLookup(apiV1.Group("/categorias"), app.CategoriaHandler)
	registerLookup(apiV1.Group("/empresas"), app.EmpresaHandler)
	registerLookup(apiV1.Group("/escolaridades"), app.EscolaridadeHandler)
	registerLookup(apiV1.Group("/instituicoes"), app.InstituicaoHandler)

	if app.TypesenseHandler != nil {
		registerTypesenseRoutes(apiV1, app.TypesenseHandler)
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
	courseDraftAuth := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseDraftAuthorization())
	courseListFilter := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseListFilter())
	courseOrgaoUpdateScope := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseOrgaoUpdateScope())
	enrollmentListAccess := rbacMiddleware(cfg.App.RBACEnabled, middlewares.CourseEnrollmentListAccess(cursoLoader))

	enrollmentCPFLoader := func(ctx context.Context, id uuid.UUID) (string, bool, error) {
		insc, err := app.InscricaoService.GetByID(ctx, id)
		if err != nil {
			return "", false, err
		}
		if insc == nil {
			return "", false, nil
		}
		return insc.CPF, true, nil
	}
	enrollmentSingleAccess := rbacMiddleware(cfg.App.RBACEnabled,
		middlewares.CourseEnrollmentSingleAccess(cursoLoader, enrollmentCPFLoader))

	courses := apiV1.Group("/courses")
	{
		courses.POST("", courseAuth, courseOrgaoInjector, app.CourseHandler.Create)
		courses.POST("/draft", courseAuth, courseDraftAuth, app.CourseHandler.CreateDraft)
		courses.GET("", courseListFilter, app.CourseHandler.List)
		courses.GET("/drafts", courseListFilter, app.CourseHandler.ListDrafts)
		courses.GET("/:courseId", courseAuth, ownershipCheck, app.CourseHandler.GetByID)
		courses.DELETE("/:courseId", courseAuth, ownershipCheck, app.CourseHandler.Delete)
		courses.PUT("/:courseId", courseAuth, ownershipCheck, courseOrgaoUpdateScope, app.CourseHandler.Update)
		courses.PUT("/:courseId/send-to-review", courseAuth, ownershipCheck, app.CourseHandler.SendToReview)
		courses.PUT("/:courseId/approve", courseAuth, ownershipCheck, app.CourseHandler.Approve)
		courses.PUT("/:courseId/request-changes", courseAuth, ownershipCheck, app.CourseHandler.RequestChanges)
		courses.PUT("/:courseId/request-deletion", courseAuth, ownershipCheck, app.CourseHandler.RequestDeletion)
		courses.POST("/:courseId/enrollments", app.InscricaoHandler.Create)
		courses.POST("/:courseId/enrollments/manual", courseAuth, ownershipCheck, app.InscricaoHandler.CreateManual)
		courses.POST("/:courseId/enrollments/import", courseAuth, ownershipCheck, app.InscricaoHandler.Import)
		courses.GET("/:courseId/enrollments", enrollmentListAccess, app.InscricaoHandler.List)
		courses.PUT("/:courseId/enrollments/status", enrollmentListAccess, app.InscricaoHandler.UpdateStatus)
		courses.PUT("/:courseId/enrollments/:enrollmentId", app.InscricaoHandler.Update)
		courses.PUT("/:courseId/enrollments/:enrollmentId/status", enrollmentSingleAccess, app.InscricaoHandler.UpdateIndividualStatus)
		courses.GET("/:courseId/enrollments/:enrollmentId", enrollmentSingleAccess, app.InscricaoHandler.GetByID)
		courses.PUT("/:courseId/enrollments/:enrollmentId/certificate", enrollmentSingleAccess, app.InscricaoHandler.UpdateCertificate)
		courses.DELETE("/:courseId/enrollments/:enrollmentId", enrollmentSingleAccess, app.InscricaoHandler.Delete)
	}

	jobCursoLoader := func(ctx context.Context, id uuid.UUID) (int, bool, error) {
		job, err := app.JobService.GetByID(ctx, id)
		if err != nil {
			return 0, false, err
		}
		if job == nil {
			return 0, false, nil
		}
		var metadata models.EnrollmentImportMetadata
		if len(job.Metadata) == 0 || json.Unmarshal(job.Metadata, &metadata) != nil {
			return 0, true, nil
		}
		return metadata.CursoID, true, nil
	}
	jobAccess := rbacMiddleware(cfg.App.RBACEnabled, middlewares.JobAccess(jobCursoLoader, cursoLoader))

	apiV1.Group("/jobs").GET("/:jobId/status", jobAccess, app.JobHandler.GetStatus)
	apiV1.Group("/users").GET("/:userId/courses", app.CourseHandler.ListByUser)
	enrollments := apiV1.Group("/enrollments")
	enrollments.GET("/user/:cpf", app.InscricaoHandler.ListByUser)
	enrollments.PUT("/:enrollmentId/schedule", app.InscricaoHandler.ChangeSchedule)

	// A gestão das oportunidades MEI (mutações, rascunhos) e a mudança de status
	// das propostas exigem perfil de gestão MEI mesmo com RBAC_ENABLED=false
	// (DEV-478). O GET por id esconde rascunhos e as listagens de propostas
	// checam o perfil, a posse do CNPJ ou as ReadPermissions no handler.
	meiGestaoAuth := middlewares.MEIGestaoAuthorization()
	oportunidades := apiV1.Group("/oportunidades-mei")
	{
		oportunidades.POST("", meiGestaoAuth, app.OportunidadeMEIHandler.Create)
		oportunidades.POST("/draft", meiGestaoAuth, app.OportunidadeMEIHandler.CreateDraft)
		oportunidades.GET("", app.OportunidadeMEIHandler.List)
		oportunidades.GET("/drafts", meiGestaoAuth, app.OportunidadeMEIHandler.ListDrafts)
		oportunidades.GET("/:id", app.OportunidadeMEIHandler.GetByID)
		oportunidades.PUT("/:id", meiGestaoAuth, app.OportunidadeMEIHandler.Update)
		oportunidades.PUT("/:id/publish", meiGestaoAuth, app.OportunidadeMEIHandler.Publish)
		oportunidades.DELETE("/:id", meiGestaoAuth, app.OportunidadeMEIHandler.Delete)
		oportunidades.POST("/:id/propostas", app.PropostaMEIHandler.Create)
		oportunidades.GET("/:id/propostas", app.PropostaMEIHandler.List)
		oportunidades.PUT("/:id/propostas/status", meiGestaoAuth, app.PropostaMEIHandler.UpdateStatusBulk)
		oportunidades.GET("/:id/propostas/:propostaId", app.PropostaMEIHandler.GetByID)
		oportunidades.PUT("/:id/propostas/:propostaId", app.PropostaMEIHandler.Update)
		oportunidades.PUT("/:id/propostas/:propostaId/status", meiGestaoAuth, app.PropostaMEIHandler.UpdateStatus)
		oportunidades.DELETE("/:id/propostas/:propostaId", app.PropostaMEIHandler.Delete)
	}
	apiV1.Group("/propostas-mei").GET("/por-empresa", app.PropostaMEIHandler.ListByMEIEmpresa)

	apiPublic.GET("/courses", app.CourseHandler.ListPublic)
	apiPublic.GET("/courses/:courseId", app.CourseHandler.GetByIDPublic)
	apiPublic.GET("/oportunidades-mei", app.OportunidadeMEIHandler.List)
	apiPublic.GET("/oportunidades-mei/:id", app.OportunidadeMEIHandler.GetByID)
}
