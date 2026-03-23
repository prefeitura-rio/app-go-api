package router

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/authorization"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/jobs"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/workers"

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
	// Consulta o Heimdall para obter as roles reais do usuário quando configurado
	apiV1.Use(middlewares.ExtractUserContext(cfg.Heimdall.BaseURL))

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
	orgaoSnapshotRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenSnapshotRepo := repository.NewCitizenSnapshotRepository(db)

	// Empregabilidade repositories
	empRegimeContratacaoRepo := empRepository.NewRegimeContratacaoRepository(db)
	empModeloTrabalhoRepo := empRepository.NewModeloTrabalhoRepository(db)
	empTipoPCDRepo := empRepository.NewTipoPCDRepository(db)
	empIdiomaRepo := empRepository.NewIdiomaRepository(db)
	empNivelIdiomaRepo := empRepository.NewNivelIdiomaRepository(db)
	empEscolaridadeRepo := empRepository.NewEscolaridadeRepository(db)
	empTipoConquistaRepo := empRepository.NewTipoConquistaRepository(db)
	empSituacaoAtualRepo := empRepository.NewSituacaoAtualRepository(db)
	empDisponibilidadeRepo := empRepository.NewDisponibilidadeRepository(db)
	empEmpresaRepo := empRepository.NewEmpresaRepository(db)
	empVagaRepo := empRepository.NewVagaRepository(db)
	empEtapaRepo := empRepository.NewEtapaRepository(db)
	empCandidaturaRepo := empRepository.NewCandidaturaRepository(db)
	empCurriculoRepo := empRepository.NewCurriculoRepository(db)
	empOnboardingRepo := empRepository.NewOnboardingRepository(db)
	empTermosUsoRepo := empRepository.NewTermosUsoRepository(db)

	// Initialize Redis client with connection pool
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
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

	// Initialize course cache (5 minute TTL)
	// Shorter TTL since courses change more frequently than reference data
	courseCache := cache.NewCourseCache(redisClient, 5*time.Minute)

	// Initialize CNAE validation service
	cnaeValidationService := services.NewCNAEValidationService(rmiClient, legalEntitiesCache)

	// Initialize service account token manager for Keycloak
	var tokenManager *auth.ServiceAccountTokenManager
	if cfg.Keycloak.URL != "" && cfg.Keycloak.ClientID != "" && cfg.Keycloak.ClientSecret != "" {
		tokenManager = auth.NewServiceAccountTokenManager(
			cfg.Keycloak.URL,
			cfg.Keycloak.Realm,
			cfg.Keycloak.ClientID,
			cfg.Keycloak.ClientSecret,
		)
	}

	// Initialize contact info service (for fetching CNPJ owner contact info)
	var contactInfoService *services.ContactInfoService
	if tokenManager != nil {
		contactInfoService = services.NewContactInfoService(rmiClient, tokenManager, redisClient, cfg)
	}

	// Initialize Data Relay client for email sending
	dataRelayClient := clients.NewDataRelayClient(cfg.DataRelay.BaseURL, cfg.DataRelay.APIKey, 30*time.Second)

	// Initialize email notification service
	emailNotificationEnabled := cfg.DataRelay.BaseURL != "" && cfg.DataRelay.APIKey != ""
	emailNotificationService := services.NewEmailNotificationService(dataRelayClient, cursoRepo, orgaoSnapshotRepo, citizenSnapshotRepo, emailNotificationEnabled, cfg.PrefRio.Domain)

	// Initialize citizen sync worker (for fetching citizen data from RMI)
	// Use interface type to avoid nil pointer in interface issue
	var citizenDataFetcher services.CitizenDataFetcher
	if cfg.CitizenSync.Enabled && tokenManager != nil {
		citizenSyncWorker := workers.NewCitizenSyncWorker(
			rmiClient,
			citizenSnapshotRepo,
			tokenManager,
			&cfg.CitizenSync,
		)
		citizenDataFetcher = citizenSyncWorker

		// Start citizen sync worker in background
		go func() {
			if err := citizenSyncWorker.Start(context.Background()); err != nil {
				log.Printf("[Router] Citizen sync worker stopped: %v", err)
			}
		}()
		log.Println("[Router] Citizen sync worker initialized and started successfully")
	} else {
		if !cfg.CitizenSync.Enabled {
			log.Println("[Router] Citizen sync worker disabled (CITIZEN_SYNC_ENABLED=false)")
		} else if tokenManager == nil {
			log.Println("[Router] Citizen sync worker disabled - Keycloak not configured (KEYCLOAK_URL, KEYCLOAK_CLIENT_ID, or KEYCLOAK_CLIENT_SECRET missing)")
		}
	}

	// Inicializando serviços
	cursoService := services.NewCursoService(cursoRepo)
	empregoService := services.NewEmpregoService(empregoRepo)
	acessibilidadeService := services.NewAcessibilidadeService(acessibilidadeRepo)
	categoriaService := services.NewCategoriaService(categoriaRepo)
	empresaService := services.NewEmpresaService(empresaRepo)
	escolaridadeService := services.NewEscolaridadeService(escolaridadeRepo)
	instituicaoService := services.NewInstituicaoService(instituicaoRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, citizenSnapshotRepo, citizenDataFetcher, emailNotificationService, cfg)
	jobService := services.NewJobService(jobRepo)
	oportunidadeMEIService := services.NewOportunidadeMEIService(oportunidadeMEIRepo)
	propostaMEIService := services.NewPropostaMEIService(propostaMEIRepo, oportunidadeMEIRepo, cnaeValidationService, contactInfoService)

	// Empregabilidade services
	empRegimeContratacaoService := empServices.NewRegimeContratacaoService(empRegimeContratacaoRepo)
	empModeloTrabalhoService := empServices.NewModeloTrabalhoService(empModeloTrabalhoRepo)
	empTipoPCDService := empServices.NewTipoPCDService(empTipoPCDRepo)
	empIdiomaService := empServices.NewIdiomaService(empIdiomaRepo)
	empNivelIdiomaService := empServices.NewNivelIdiomaService(empNivelIdiomaRepo)
	empEscolaridadeService := empServices.NewEscolaridadeService(empEscolaridadeRepo)
	empTipoConquistaService := empServices.NewTipoConquistaService(empTipoConquistaRepo)
	empSituacaoAtualService := empServices.NewSituacaoAtualService(empSituacaoAtualRepo)
	empDisponibilidadeService := empServices.NewDisponibilidadeService(empDisponibilidadeRepo)
	empEmpresaService := empServices.NewEmpresaService(empEmpresaRepo)
	empVagaService := empServices.NewVagaService(empVagaRepo, empEmpresaRepo, empCandidaturaRepo)
	empEtapaService := empServices.NewEtapaService(empEtapaRepo)
	empCurriculoService := empServices.NewCurriculoService(empCurriculoRepo)
	empCandidaturaService := empServices.NewCandidaturaService(empCandidaturaRepo, empVagaRepo, empCurriculoService, citizenSnapshotRepo, citizenDataFetcher)
	empOnboardingService := empServices.NewOnboardingService(empOnboardingRepo)
	empTermosUsoService := empServices.NewTermosUsoService(empTermosUsoRepo)

	// Initialize CNPJ consulta service for empregabilidade (requires service account)
	var empCNPJConsultaService *empServices.CNPJConsultaService
	if tokenManager != nil {
		empCNPJConsultaService = empServices.NewCNPJConsultaService(rmiClient, tokenManager)
	}

	// Initialize job processor
	jobs.InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)

	// Initialize Cerbos authorization checker (if enabled)
	var authChecker *authorization.Checker
	if cfg.Cerbos.Enabled {
		authChecker = authorization.NewChecker(
			cfg.Cerbos.Endpoint,
			time.Duration(cfg.Cerbos.Timeout)*time.Second,
		)
	}

	// Inicializando handlers
	empregoHandler := v1.NewEmpregoHandler(empregoService)
	acessibilidadeHandler := v1.NewAcessibilidadeHandler(acessibilidadeService).WithCache(acessibilidadesCache)
	categoriaHandler := v1.NewCategoriaHandler(categoriaService).WithCache(categoriasCache)
	empresaHandler := v1.NewEmpresaHandler(empresaService)
	escolaridadeHandler := v1.NewEscolaridadeHandler(escolaridadeService).WithCache(escolaridadesCache)
	instituicaoHandler := v1.NewInstituicaoHandler(instituicaoService)
	inscricaoHandler := v1.NewInscricaoHandler(inscricaoService, jobService, cursoRepo)
	courseHandler := v1.NewCourseHandler(cursoService, inscricaoService, cursoRepo).WithCache(courseCache)
	jobHandler := v1.NewJobHandler(jobService)
	oportunidadeMEIHandler := v1.NewOportunidadeMEIHandler(oportunidadeMEIService)
	propostaMEIHandler := v1.NewPropostaMEIHandler(propostaMEIService, cnaeValidationService, authChecker, cfg)
	typesenseHandler, err := v1.NewTypesenseHandler()
	if err != nil {
		fmt.Printf("Erro ao inicializar o Typesense: %v\n", err)
	}

	// Empregabilidade handlers
	empRegimeContratacaoHandler := empHandlers.NewRegimeContratacaoHandler(empRegimeContratacaoService)
	empModeloTrabalhoHandler := empHandlers.NewModeloTrabalhoHandler(empModeloTrabalhoService)
	empTipoPCDHandler := empHandlers.NewTipoPCDHandler(empTipoPCDService)
	empIdiomaHandler := empHandlers.NewIdiomaHandler(empIdiomaService)
	empNivelIdiomaHandler := empHandlers.NewNivelIdiomaHandler(empNivelIdiomaService)
	empEscolaridadeHandler := empHandlers.NewEscolaridadeHandler(empEscolaridadeService)
	empTipoConquistaHandler := empHandlers.NewTipoConquistaHandler(empTipoConquistaService)
	empSituacaoAtualHandler := empHandlers.NewSituacaoAtualHandler(empSituacaoAtualService)
	empDisponibilidadeHandler := empHandlers.NewDisponibilidadeHandler(empDisponibilidadeService)
	empEmpresaHandler := empHandlers.NewEmpresaHandler(empEmpresaService).WithCNPJConsulta(empCNPJConsultaService)
	empVagaHandler := empHandlers.NewVagaHandler(empVagaService)
	empEtapaHandler := empHandlers.NewEtapaHandler(empEtapaService)
	empCandidaturaHandler := empHandlers.NewCandidaturaHandler(empCandidaturaService)
	empCurriculoHandler := empHandlers.NewCurriculoHandler(empCurriculoService)
	empOnboardingHandler := empHandlers.NewOnboardingHandler(empOnboardingService)
	empTermosUsoHandler := empHandlers.NewTermosUsoHandler(empTermosUsoService)

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
		enrollments.PUT("/:enrollmentId/schedule", inscricaoHandler.ChangeSchedule)
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

	// ==========================================
	// Rotas de Empregabilidade
	// ==========================================

	empregabilidade := apiV1.Group("/empregabilidade")

	// Tabelas auxiliares (lookup tables)
	empRegimesContratacao := empregabilidade.Group("/regimes-contratacao")
	{
		empRegimesContratacao.POST("", empRegimeContratacaoHandler.Create)
		empRegimesContratacao.GET("", empRegimeContratacaoHandler.List)
		empRegimesContratacao.GET("/:id", empRegimeContratacaoHandler.GetByID)
		empRegimesContratacao.PUT("/:id", empRegimeContratacaoHandler.Update)
		empRegimesContratacao.DELETE("/:id", empRegimeContratacaoHandler.Delete)
	}

	empModelosTrabalho := empregabilidade.Group("/modelos-trabalho")
	{
		empModelosTrabalho.POST("", empModeloTrabalhoHandler.Create)
		empModelosTrabalho.GET("", empModeloTrabalhoHandler.List)
		empModelosTrabalho.GET("/:id", empModeloTrabalhoHandler.GetByID)
		empModelosTrabalho.PUT("/:id", empModeloTrabalhoHandler.Update)
		empModelosTrabalho.DELETE("/:id", empModeloTrabalhoHandler.Delete)
	}

	empTiposPCD := empregabilidade.Group("/tipos-pcd")
	{
		empTiposPCD.POST("", empTipoPCDHandler.Create)
		empTiposPCD.GET("", empTipoPCDHandler.List)
		empTiposPCD.GET("/:id", empTipoPCDHandler.GetByID)
		empTiposPCD.PUT("/:id", empTipoPCDHandler.Update)
		empTiposPCD.DELETE("/:id", empTipoPCDHandler.Delete)
	}

	empIdiomas := empregabilidade.Group("/idiomas")
	{
		empIdiomas.POST("", empIdiomaHandler.Create)
		empIdiomas.GET("", empIdiomaHandler.List)
		empIdiomas.GET("/:id", empIdiomaHandler.GetByID)
		empIdiomas.PUT("/:id", empIdiomaHandler.Update)
		empIdiomas.DELETE("/:id", empIdiomaHandler.Delete)
	}

	empNiveisIdioma := empregabilidade.Group("/niveis-idioma")
	{
		empNiveisIdioma.POST("", empNivelIdiomaHandler.Create)
		empNiveisIdioma.GET("", empNivelIdiomaHandler.List)
		empNiveisIdioma.GET("/:id", empNivelIdiomaHandler.GetByID)
		empNiveisIdioma.PUT("/:id", empNivelIdiomaHandler.Update)
		empNiveisIdioma.DELETE("/:id", empNivelIdiomaHandler.Delete)
	}

	empEscolaridades := empregabilidade.Group("/escolaridades")
	{
		empEscolaridades.POST("", empEscolaridadeHandler.Create)
		empEscolaridades.GET("", empEscolaridadeHandler.List)
		empEscolaridades.GET("/:id", empEscolaridadeHandler.GetByID)
		empEscolaridades.PUT("/:id", empEscolaridadeHandler.Update)
		empEscolaridades.DELETE("/:id", empEscolaridadeHandler.Delete)
	}

	empTiposConquista := empregabilidade.Group("/tipos-conquista")
	{
		empTiposConquista.POST("", empTipoConquistaHandler.Create)
		empTiposConquista.GET("", empTipoConquistaHandler.List)
		empTiposConquista.GET("/:id", empTipoConquistaHandler.GetByID)
		empTiposConquista.PUT("/:id", empTipoConquistaHandler.Update)
		empTiposConquista.DELETE("/:id", empTipoConquistaHandler.Delete)
	}

	empSituacoesAtual := empregabilidade.Group("/situacoes-atual")
	{
		empSituacoesAtual.POST("", empSituacaoAtualHandler.Create)
		empSituacoesAtual.GET("", empSituacaoAtualHandler.List)
		empSituacoesAtual.GET("/:id", empSituacaoAtualHandler.GetByID)
		empSituacoesAtual.PUT("/:id", empSituacaoAtualHandler.Update)
		empSituacoesAtual.DELETE("/:id", empSituacaoAtualHandler.Delete)
	}

	empDisponibilidades := empregabilidade.Group("/disponibilidades")
	{
		empDisponibilidades.POST("", empDisponibilidadeHandler.Create)
		empDisponibilidades.GET("", empDisponibilidadeHandler.List)
		empDisponibilidades.GET("/:id", empDisponibilidadeHandler.GetByID)
		empDisponibilidades.PUT("/:id", empDisponibilidadeHandler.Update)
		empDisponibilidades.DELETE("/:id", empDisponibilidadeHandler.Delete)
	}

	// Empresas
	empEmpresas := empregabilidade.Group("/empresas")
	{
		empEmpresas.POST("", empEmpresaHandler.Create)
		empEmpresas.GET("", empEmpresaHandler.List)
		empEmpresas.GET("/consulta-cnpj/:cnpj", empEmpresaHandler.ConsultaCNPJ)
		empEmpresas.GET("/:cnpj", empEmpresaHandler.GetByID)
		empEmpresas.PUT("/:cnpj", empEmpresaHandler.Update)
		empEmpresas.DELETE("/:cnpj", empEmpresaHandler.Delete)
	}

	// Vagas
	empVagas := empregabilidade.Group("/vagas")
	{
		empVagas.POST("", empVagaHandler.Create)
		empVagas.POST("/draft", empVagaHandler.Create)
		empVagas.GET("", empVagaHandler.List)
		empVagas.GET("/:id", empVagaHandler.GetByID)
		empVagas.PUT("/:id", empVagaHandler.Update)
		empVagas.DELETE("/:id", empVagaHandler.Delete)
		empVagas.PUT("/:id/send-to-draft", empVagaHandler.SendToDraft)
		empVagas.PUT("/:id/send-to-approval", empVagaHandler.SendToApproval)
		empVagas.PUT("/:id/publish", empVagaHandler.Publish)
		empVagas.PUT("/:id/freeze", empVagaHandler.Freeze)
		empVagas.PUT("/:id/unfreeze", empVagaHandler.Unfreeze)
		empVagas.PUT("/:id/discontinue", empVagaHandler.Discontinue)
		empVagas.PUT("/:id/reactivate", empVagaHandler.Reactivate)
		empVagas.PUT("/:id/tipos-pcd", empVagaHandler.UpdateTiposPCD)

		// Etapas (nested)
		empVagas.POST("/:id/etapas", empEtapaHandler.Create)
		empVagas.GET("/:id/etapas", empEtapaHandler.ListByVaga)
		empVagas.GET("/:id/etapas/:etapaId", empEtapaHandler.GetByID)
		empVagas.PUT("/:id/etapas/:etapaId", empEtapaHandler.Update)
		empVagas.DELETE("/:id/etapas/:etapaId", empEtapaHandler.Delete)
	}

	// Candidaturas
	empCandidaturas := empregabilidade.Group("/candidaturas")
	{
		empCandidaturas.POST("", empCandidaturaHandler.Create)
		empCandidaturas.GET("", empCandidaturaHandler.List)
		empCandidaturas.PUT("/bulk-status", empCandidaturaHandler.BulkUpdateStatus)
		empCandidaturas.PUT("/bulk-etapa", empCandidaturaHandler.BulkUpdateEtapa)
		empCandidaturas.GET("/usuario/:cpf", empCandidaturaHandler.ListByCPF)
		empCandidaturas.GET("/:id", empCandidaturaHandler.GetByID)
		empCandidaturas.PUT("/:id", empCandidaturaHandler.Update)
		empCandidaturas.DELETE("/:id", empCandidaturaHandler.Delete)
		empCandidaturas.PUT("/:id/status", empCandidaturaHandler.UpdateStatus)
		empCandidaturas.PUT("/:id/approve", empCandidaturaHandler.Approve)
		empCandidaturas.PUT("/:id/reject", empCandidaturaHandler.Reject)
		empCandidaturas.PUT("/:id/etapa", empCandidaturaHandler.UpdateEtapa)
	}

	// Currículo
	empCurriculo := empregabilidade.Group("/curriculo")
	{
		empCurriculo.GET("/:cpf", empCurriculoHandler.GetCurriculoCompleto)

		// Formações
		empCurriculo.POST("/formacoes", empCurriculoHandler.CreateFormacao)
		empCurriculo.GET("/formacoes/:id", empCurriculoHandler.GetFormacaoByID)
		empCurriculo.PUT("/formacoes/:id", empCurriculoHandler.UpdateFormacao)
		empCurriculo.DELETE("/formacoes/:id", empCurriculoHandler.DeleteFormacao)
		empCurriculo.GET("/:cpf/formacoes", empCurriculoHandler.ListFormacoesByCPF)
		empCurriculo.PUT("/:cpf/formacoes", empCurriculoHandler.ReplaceAllFormacoesByCPF)

		// Idiomas
		empCurriculo.POST("/idiomas", empCurriculoHandler.CreateIdioma)
		empCurriculo.GET("/idiomas/:id", empCurriculoHandler.GetIdiomaByID)
		empCurriculo.PUT("/idiomas/:id", empCurriculoHandler.UpdateIdioma)
		empCurriculo.DELETE("/idiomas/:id", empCurriculoHandler.DeleteIdioma)
		empCurriculo.GET("/:cpf/idiomas", empCurriculoHandler.ListIdiomasByCPF)
		empCurriculo.PUT("/:cpf/idiomas", empCurriculoHandler.ReplaceAllIdiomasByCPF)

		// Cursos Complementares
		empCurriculo.POST("/cursos-complementares", empCurriculoHandler.CreateCursoComplementar)
		empCurriculo.GET("/cursos-complementares/:id", empCurriculoHandler.GetCursoComplementarByID)
		empCurriculo.PUT("/cursos-complementares/:id", empCurriculoHandler.UpdateCursoComplementar)
		empCurriculo.DELETE("/cursos-complementares/:id", empCurriculoHandler.DeleteCursoComplementar)
		empCurriculo.GET("/:cpf/cursos-complementares", empCurriculoHandler.ListCursosComplementaresByCPF)
		empCurriculo.PUT("/:cpf/cursos-complementares", empCurriculoHandler.ReplaceAllCursosComplementaresByCPF)

		// Experiências
		empCurriculo.POST("/experiencias", empCurriculoHandler.CreateExperiencia)
		empCurriculo.GET("/experiencias/:id", empCurriculoHandler.GetExperienciaByID)
		empCurriculo.PUT("/experiencias/:id", empCurriculoHandler.UpdateExperiencia)
		empCurriculo.DELETE("/experiencias/:id", empCurriculoHandler.DeleteExperiencia)
		empCurriculo.GET("/:cpf/experiencias", empCurriculoHandler.ListExperienciasByCPF)
		empCurriculo.PUT("/:cpf/experiencias", empCurriculoHandler.ReplaceAllExperienciasByCPF)

		// Conquistas
		empCurriculo.POST("/conquistas", empCurriculoHandler.CreateConquista)
		empCurriculo.GET("/conquistas/:id", empCurriculoHandler.GetConquistaByID)
		empCurriculo.PUT("/conquistas/:id", empCurriculoHandler.UpdateConquista)
		empCurriculo.DELETE("/conquistas/:id", empCurriculoHandler.DeleteConquista)
		empCurriculo.GET("/:cpf/conquistas", empCurriculoHandler.ListConquistasByCPF)
		empCurriculo.PUT("/:cpf/conquistas", empCurriculoHandler.ReplaceAllConquistasByCPF)

		// Situação e Interesses
		empCurriculo.PUT("/situacao-interesses", empCurriculoHandler.UpsertSituacaoInteresses)
		empCurriculo.GET("/:cpf/situacao-interesses", empCurriculoHandler.GetSituacaoInteressesByCPF)
	}

	// Onboarding
	empOnboarding := empregabilidade.Group("/onboarding")
	{
		empOnboarding.GET("/:cpf", empOnboardingHandler.IsFirstLogin)
		empOnboarding.PUT("/:cpf/complete", empOnboardingHandler.MarkFirstLoginCompleted)
	}

	// Termos de Uso
	empTermosUso := empregabilidade.Group("/termos-uso")
	{
		empTermosUso.GET("/:cpf", empTermosUsoHandler.HasAcceptedTerms)
		empTermosUso.PUT("/:cpf/accept", empTermosUsoHandler.AcceptTerms)
		empTermosUso.GET("/:cpf/details", empTermosUsoHandler.GetDetails)
	}

	// Endpoints públicos para empregabilidade (apenas vagas publicadas)
	apiPublicEmpregabilidade := apiPublic.Group("/empregabilidade")
	{
		apiPublicEmpregabilidade.GET("/vagas", empVagaHandler.PublicList)
		apiPublicEmpregabilidade.GET("/vagas/:id", empVagaHandler.PublicGetByID)
	}

	return r
}
