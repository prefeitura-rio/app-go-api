package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/jobs"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"

	_ "github.com/prefeitura-rio/app-go-api/docs"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Middleware global
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middlewares.CorsMiddleware())

	// Configuração do Swagger com host dinâmico
	r.GET("/docs/*any", middlewares.DynamicSwaggerHandler())

	// Rota de saúde
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1
	apiV1 := r.Group("/api/v1")

	// Middleware para extrair contexto do usuário dos headers injetados pelo Istio
	apiV1.Use(middlewares.ExtractUserContext())

	// Inicializando repositórios
	cursoRepo := repository.NewCursoRepository(db)
	empregoRepo := repository.NewEmpregoRepository(db)
	acessibilidadeRepo := repository.NewAcessibilidadeRepository(db)
	categoriaRepo := repository.NewCategoriaRepository(db)
	empresaRepo := repository.NewEmpresaRepository(db)
	escolaridadeRepo := repository.NewEscolaridadeRepository(db)
	instituicaoRepo := repository.NewInstituicaoRepository(db)
	orgaoRepo := repository.NewOrgaoRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	jobRepo := repository.NewJobRepository(db)
	cnaeRepo := repository.NewCNAERepository(db)
	meiEmpresaRepo := repository.NewMEIEmpresaRepository(db)
	oportunidadeMEIRepo := repository.NewOportunidadeMEIRepository(db)
	propostaMEIRepo := repository.NewPropostaMEIRepository(db)

	// Inicializando serviços
	cursoService := services.NewCursoService(cursoRepo)
	empregoService := services.NewEmpregoService(empregoRepo)
	acessibilidadeService := services.NewAcessibilidadeService(acessibilidadeRepo)
	categoriaService := services.NewCategoriaService(categoriaRepo)
	empresaService := services.NewEmpresaService(empresaRepo)
	escolaridadeService := services.NewEscolaridadeService(escolaridadeRepo)
	instituicaoService := services.NewInstituicaoService(instituicaoRepo)
	orgaoService := services.NewOrgaoService(orgaoRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo)
	jobService := services.NewJobService(jobRepo)
	cnaeService := services.NewCNAEService(cnaeRepo)
	meiEmpresaService := services.NewMEIEmpresaService(meiEmpresaRepo, cnaeRepo)
	oportunidadeMEIService := services.NewOportunidadeMEIService(oportunidadeMEIRepo, cnaeRepo, orgaoRepo)
	propostaMEIService := services.NewPropostaMEIService(propostaMEIRepo, oportunidadeMEIRepo, meiEmpresaRepo)

	// Initialize job processor
	jobs.InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)

	// Inicializando handlers
	empregoHandler := v1.NewEmpregoHandler(empregoService)
	acessibilidadeHandler := v1.NewAcessibilidadeHandler(acessibilidadeService)
	categoriaHandler := v1.NewCategoriaHandler(categoriaService)
	empresaHandler := v1.NewEmpresaHandler(empresaService)
	escolaridadeHandler := v1.NewEscolaridadeHandler(escolaridadeService)
	instituicaoHandler := v1.NewInstituicaoHandler(instituicaoService)
	orgaoHandler := v1.NewOrgaoHandler(orgaoService)
	inscricaoHandler := v1.NewInscricaoHandler(inscricaoService, jobService)
	courseHandler := v1.NewCourseHandler(cursoService, inscricaoService)
	jobHandler := v1.NewJobHandler(jobService)
	cnaeHandler := v1.NewCNAEHandler(cnaeService)
	meiEmpresaHandler := v1.NewMEIEmpresaHandler(meiEmpresaService)
	oportunidadeMEIHandler := v1.NewOportunidadeMEIHandler(oportunidadeMEIService)
	propostaMEIHandler := v1.NewPropostaMEIHandler(propostaMEIService)
	typesenseHandler, err := v1.NewTypesenseHandler()
	if err != nil {
		fmt.Printf("Erro ao inicializar o Typesense: %v\n", err)
	}

	// Rotas de empregos
	empregos := apiV1.Group("/empregos")
	{
		empregos.POST("", empregoHandler.Create)
		empregos.GET("", empregoHandler.List)
		empregos.GET("/:id", empregoHandler.GetByID)
		empregos.PUT("/:id", empregoHandler.Update)
		empregos.DELETE("/:id", empregoHandler.Delete)
	}

	// Rotas de acessibilidades
	acessibilidades := apiV1.Group("/acessibilidades")
	{
		acessibilidades.POST("", acessibilidadeHandler.Create)
		acessibilidades.GET("", acessibilidadeHandler.List)
		acessibilidades.GET("/:id", acessibilidadeHandler.GetByID)
		acessibilidades.PUT("/:id", acessibilidadeHandler.Update)
		acessibilidades.DELETE("/:id", acessibilidadeHandler.Delete)
	}

	// Rotas de categorias
	categorias := apiV1.Group("/categorias")
	{
		categorias.POST("", categoriaHandler.Create)
		categorias.GET("", categoriaHandler.List)
		categorias.GET("/:id", categoriaHandler.GetByID)
		categorias.PUT("/:id", categoriaHandler.Update)
		categorias.DELETE("/:id", categoriaHandler.Delete)
	}

	// Rotas de empresas
	empresas := apiV1.Group("/empresas")
	{
		empresas.POST("", empresaHandler.Create)
		empresas.GET("", empresaHandler.List)
		empresas.GET("/:id", empresaHandler.GetByID)
		empresas.PUT("/:id", empresaHandler.Update)
		empresas.DELETE("/:id", empresaHandler.Delete)
	}

	// Rotas de escolaridades
	escolaridades := apiV1.Group("/escolaridades")
	{
		escolaridades.POST("", escolaridadeHandler.Create)
		escolaridades.GET("", escolaridadeHandler.List)
		escolaridades.GET("/:id", escolaridadeHandler.GetByID)
		escolaridades.PUT("/:id", escolaridadeHandler.Update)
		escolaridades.DELETE("/:id", escolaridadeHandler.Delete)
	}

	// Rotas de instituições de ensino
	instituicoes := apiV1.Group("/instituicoes")
	{
		instituicoes.POST("", instituicaoHandler.Create)
		instituicoes.GET("", instituicaoHandler.List)
		instituicoes.GET("/:id", instituicaoHandler.GetByID)
		instituicoes.PUT("/:id", instituicaoHandler.Update)
		instituicoes.DELETE("/:id", instituicaoHandler.Delete)
	}

	// Rotas de órgãos
	orgaos := apiV1.Group("/orgaos")
	{
		orgaos.POST("", orgaoHandler.Create)
		orgaos.GET("", orgaoHandler.List)
		orgaos.GET("/:id", orgaoHandler.GetByID)
		orgaos.PUT("/:id", orgaoHandler.Update)
		orgaos.DELETE("/:id", orgaoHandler.Delete)
	}

	// Rotas de Typesense
	if typesenseHandler != nil {
		apiV1.POST("/typesense/multi-search", typesenseHandler.SearchMultiCollection)
		apiV1.POST("/typesense/cursos/search", typesenseHandler.SearchCursos)
		apiV1.POST("/typesense/empregos/search", typesenseHandler.SearchEmpregos)
		typesenseCollections := apiV1.Group("/typesense/collections")
		typesenseCollections.POST("/:collection/documents/search", typesenseHandler.SearchDocuments)
	}

	// New Course API endpoints following specification (in v1)
	courses := apiV1.Group("/courses")
	{
		courses.POST("", courseHandler.Create)
		courses.POST("/draft", courseHandler.CreateDraft)
		courses.PUT("/:courseId", courseHandler.Update)
		courses.GET("", courseHandler.List)
		courses.GET("/drafts", courseHandler.ListDrafts)
		courses.GET("/:courseId", courseHandler.GetByID)
		courses.DELETE("/:courseId", courseHandler.Delete)

		// Enrollment endpoints
		courses.POST("/:courseId/enrollments", inscricaoHandler.Create)
		courses.POST("/:courseId/enrollments/manual", inscricaoHandler.CreateManual)
		courses.POST("/:courseId/enrollments/import", inscricaoHandler.Import)
		courses.GET("/:courseId/enrollments", inscricaoHandler.List)
		courses.PUT("/:courseId/enrollments/status", inscricaoHandler.UpdateStatus)
		courses.PUT("/:courseId/enrollments/:enrollmentId", inscricaoHandler.Update)
		courses.PUT("/:courseId/enrollments/:enrollmentId/status", inscricaoHandler.UpdateIndividualStatus)
		courses.GET("/:courseId/enrollments/:enrollmentId", inscricaoHandler.GetByID)
		courses.PUT("/:courseId/enrollments/:enrollmentId/certificate", inscricaoHandler.UpdateCertificate)
		courses.DELETE("/:courseId/enrollments/:enrollmentId", inscricaoHandler.Delete)
	}

	// Job status endpoints
	jobsGroup := apiV1.Group("/jobs")
	{
		jobsGroup.GET("/:jobId/status", jobHandler.GetStatus)
	}

	// Endpoints públicos (sem autenticação) para busca e listagem de cursos
	apiPublic := r.Group("/api/public")
	{
		// Endpoints públicos para cursos - reutilizam os handlers mas sem autenticação
		apiPublic.GET("/courses", courseHandler.List)
		apiPublic.GET("/courses/:courseId", courseHandler.GetByID)
	}

	// Endpoints para usuários - cursos por usuário (orgão)
	users := apiV1.Group("/users")
	{
		users.GET("/:userId/courses", courseHandler.ListByUser)
	}

	// Endpoints para inscrições por CPF
	enrollments := apiV1.Group("/enrollments")
	{
		enrollments.GET("/user/:cpf", inscricaoHandler.ListByUser)
	}

	// Rotas de CNAEs
	cnaes := apiV1.Group("/cnaes")
	{
		cnaes.GET("", cnaeHandler.List)
		cnaes.GET("/*codigo", cnaeHandler.GetByCodigo)
	}

	// Rotas admin de CNAEs
	adminCnaes := apiV1.Group("/admin/cnaes")
	{
		adminCnaes.POST("", cnaeHandler.Create)
		adminCnaes.PUT("/*codigo", cnaeHandler.Update)
		adminCnaes.DELETE("/*codigo", cnaeHandler.Delete)
	}

	// Rotas de MEI Empresas
	meiEmpresas := apiV1.Group("/mei-empresas")
	{
		meiEmpresas.POST("", meiEmpresaHandler.Create)
		meiEmpresas.GET("", meiEmpresaHandler.List)
		meiEmpresas.GET("/:id", meiEmpresaHandler.GetByID)
		meiEmpresas.GET("/cnpj/:cnpj", meiEmpresaHandler.GetByCNPJ)
		meiEmpresas.PUT("/:id", meiEmpresaHandler.Update)
	}

	// Rotas de Oportunidades MEI
	oportunidadesMEI := apiV1.Group("/oportunidades-mei")
	{
		oportunidadesMEI.POST("", oportunidadeMEIHandler.Create)
		oportunidadesMEI.POST("/draft", oportunidadeMEIHandler.CreateDraft)
		oportunidadesMEI.GET("", oportunidadeMEIHandler.List)
		oportunidadesMEI.GET("/drafts", oportunidadeMEIHandler.ListDrafts)
		oportunidadesMEI.GET("/:id", oportunidadeMEIHandler.GetByID)
		oportunidadesMEI.PUT("/:id", oportunidadeMEIHandler.Update)
		oportunidadesMEI.PUT("/:id/publish", oportunidadeMEIHandler.Publish)
		oportunidadesMEI.DELETE("/:id", oportunidadeMEIHandler.Delete)

		// Rotas de Propostas MEI (nested)
		oportunidadesMEI.POST("/:id/propostas", propostaMEIHandler.Create)
		oportunidadesMEI.GET("/:id/propostas", propostaMEIHandler.List)
		oportunidadesMEI.PUT("/:id/propostas/status", propostaMEIHandler.UpdateStatusBulk)
		oportunidadesMEI.GET("/:id/propostas/:propostaId", propostaMEIHandler.GetByID)
		oportunidadesMEI.PUT("/:id/propostas/:propostaId", propostaMEIHandler.Update)
		oportunidadesMEI.PUT("/:id/propostas/:propostaId/status", propostaMEIHandler.UpdateStatus)
		oportunidadesMEI.DELETE("/:id/propostas/:propostaId", propostaMEIHandler.Delete)
	}

	// Rota adicional para listar propostas por MEI empresa
	propostasMEI := apiV1.Group("/propostas-mei")
	{
		propostasMEI.GET("/por-empresa", propostaMEIHandler.ListByMEIEmpresa)
	}

	// Endpoints públicos para oportunidades MEI
	apiPublic.GET("/oportunidades-mei", oportunidadeMEIHandler.List)
	apiPublic.GET("/oportunidades-mei/:id", oportunidadeMEIHandler.GetByID)
	apiPublic.GET("/cnaes", cnaeHandler.List)
	apiPublic.GET("/cnaes/*codigo", cnaeHandler.GetByCodigo)

	return r
}
