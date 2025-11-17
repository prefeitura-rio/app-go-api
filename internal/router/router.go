package router

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/jobs"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"

	_ "github.com/prefeitura-rio/app-go-api/docs"
)

func SetupRouter(db *gorm.DB, cfg *config.AppConfig) *gin.Engine {
	r := gin.Default()

	// Middleware global
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// OpenTelemetry middleware (if tracing is enabled)
	if cfg.Tracing.Enabled {
		r.Use(otelgin.Middleware(cfg.Tracing.ServiceName))
	}

	// Request timeout middleware (prevents long-running requests from accumulating)
	r.Use(middlewares.TimeoutMiddleware(time.Duration(cfg.Server.RequestTimeout) * time.Second))

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
	inscricaoRepo := repository.NewInscricaoRepository(db)
	jobRepo := repository.NewJobRepository(db)
	oportunidadeMEIRepo := repository.NewOportunidadeMEIRepository(db)
	propostaMEIRepo := repository.NewPropostaMEIRepository(db)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Add OpenTelemetry instrumentation to Redis (if tracing is enabled)
	if cfg.Tracing.Enabled {
		if err := redisotel.InstrumentTracing(redisClient); err != nil {
			panic(fmt.Sprintf("Error instrumenting Redis with OTEL: %v", err))
		}
	}

	// Initialize RMI client (15s timeout per request)
	rmiClient := clients.NewRMIClient(cfg.RMI.BaseURL, 15*time.Second)

	// Initialize Redis cache for legal entities (30 min TTL)
	legalEntitiesCache := cache.NewLegalEntitiesCache(redisClient, 30*time.Minute)

	// Initialize Redis caches for static reference data (1 hour TTL)
	// These are rarely-changing lookup tables accessed frequently
	categoriasCache := cache.NewReferenceDataCache(redisClient, 1*time.Hour, "categorias")
	acessibilidadesCache := cache.NewReferenceDataCache(redisClient, 1*time.Hour, "acessibilidades")
	escolaridadesCache := cache.NewReferenceDataCache(redisClient, 1*time.Hour, "escolaridades")

	// Initialize CNAE validation service
	cnaeValidationService := services.NewCNAEValidationService(rmiClient, legalEntitiesCache)

	// Inicializando serviços
	cursoService := services.NewCursoService(cursoRepo)
	empregoService := services.NewEmpregoService(empregoRepo)
	acessibilidadeService := services.NewAcessibilidadeService(acessibilidadeRepo)
	categoriaService := services.NewCategoriaService(categoriaRepo)
	empresaService := services.NewEmpresaService(empresaRepo)
	escolaridadeService := services.NewEscolaridadeService(escolaridadeRepo)
	instituicaoService := services.NewInstituicaoService(instituicaoRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo)
	jobService := services.NewJobService(jobRepo)
	oportunidadeMEIService := services.NewOportunidadeMEIService(oportunidadeMEIRepo)
	propostaMEIService := services.NewPropostaMEIService(propostaMEIRepo, oportunidadeMEIRepo, cnaeValidationService)

	// Initialize job processor
	jobs.InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)

	// Inicializando handlers
	empregoHandler := v1.NewEmpregoHandler(empregoService)
	acessibilidadeHandler := v1.NewAcessibilidadeHandler(acessibilidadeService).WithCache(acessibilidadesCache)
	categoriaHandler := v1.NewCategoriaHandler(categoriaService).WithCache(categoriasCache)
	empresaHandler := v1.NewEmpresaHandler(empresaService)
	escolaridadeHandler := v1.NewEscolaridadeHandler(escolaridadeService).WithCache(escolaridadesCache)
	instituicaoHandler := v1.NewInstituicaoHandler(instituicaoService)
	inscricaoHandler := v1.NewInscricaoHandler(inscricaoService, jobService)
	courseHandler := v1.NewCourseHandler(cursoService, inscricaoService, cursoRepo)
	jobHandler := v1.NewJobHandler(jobService)
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

	return r
}
