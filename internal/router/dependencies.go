package router

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/authorization"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/jobs"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/workers"
)

type Dependencies struct {
	// Handlers - Cursos
	CourseHandler    *v1.CourseHandler
	InscricaoHandler *v1.InscricaoHandler
	JobHandler       *v1.JobHandler

	// Handlers - Legacy
	EmpregoHandler       *v1.EmpregoHandler
	AcessibilidadeHandler *v1.AcessibilidadeHandler
	CategoriaHandler     *v1.CategoriaHandler
	EmpresaHandler       *v1.EmpresaHandler
	EscolaridadeHandler  *v1.EscolaridadeHandler
	InstituicaoHandler   *v1.InstituicaoHandler

	// Handlers - MEI
	OportunidadeMEIHandler *v1.OportunidadeMEIHandler
	PropostaMEIHandler     *v1.PropostaMEIHandler

	// Handlers - Typesense
	TypesenseHandler *v1.TypesenseHandler

	// Handlers - Empregabilidade
	EmpRegimeContratacaoHandler *empHandlers.RegimeContratacaoHandler
	EmpModeloTrabalhoHandler    *empHandlers.ModeloTrabalhoHandler
	EmpTipoPCDHandler           *empHandlers.TipoPCDHandler
	EmpIdiomaHandler            *empHandlers.IdiomaHandler
	EmpNivelIdiomaHandler       *empHandlers.NivelIdiomaHandler
	EmpEscolaridadeHandler      *empHandlers.EscolaridadeHandler
	EmpTipoConquistaHandler     *empHandlers.TipoConquistaHandler
	EmpSituacaoAtualHandler     *empHandlers.SituacaoAtualHandler
	EmpDisponibilidadeHandler   *empHandlers.DisponibilidadeHandler
	EmpEmpresaHandler           *empHandlers.EmpresaHandler
	EmpVagaHandler              *empHandlers.VagaHandler
	EmpEtapaHandler             *empHandlers.EtapaHandler
	EmpCandidaturaHandler       *empHandlers.CandidaturaHandler
	EmpCurriculoHandler         *empHandlers.CurriculoHandler
	EmpOnboardingHandler        *empHandlers.OnboardingHandler
	EmpTermosUsoHandler         *empHandlers.TermosUsoHandler
}

func InitDependencies(db *gorm.DB, cfg *config.AppConfig) *Dependencies {
	// Repositories
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
	empInformacaoComplementarRepo := empRepository.NewInformacaoComplementarRepository(db)
	empCandidaturaRepo := empRepository.NewCandidaturaRepository(db)
	empCurriculoRepo := empRepository.NewCurriculoRepository(db)
	empOnboardingRepo := empRepository.NewOnboardingRepository(db)
	empTermosUsoRepo := empRepository.NewTermosUsoRepository(db)

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	if cfg.Tracing.Enabled {
		if err := redisotel.InstrumentTracing(redisClient); err != nil {
			panic(fmt.Sprintf("Error instrumenting Redis with OTEL: %v", err))
		}
	}

	// External clients
	rmiClient := clients.NewRMIClient(cfg.RMI.BaseURL, 15*time.Second)
	dataRelayClient := clients.NewDataRelayClient(cfg.DataRelay.BaseURL, cfg.DataRelay.APIKey, 30*time.Second)

	// Caches
	legalEntitiesCache := cache.NewLegalEntitiesCache(redisClient, 30*time.Minute)
	categoriasCache := cache.NewReferenceDataCache(redisClient, 1*time.Hour, "categorias")
	acessibilidadesCache := cache.NewReferenceDataCache(redisClient, 1*time.Hour, "acessibilidades")
	escolaridadesCache := cache.NewReferenceDataCache(redisClient, 1*time.Hour, "escolaridades")
	courseCache := cache.NewCourseCache(redisClient, 5*time.Minute)

	// Auth
	var tokenManager *auth.ServiceAccountTokenManager
	if cfg.Keycloak.URL != "" && cfg.Keycloak.ClientID != "" && cfg.Keycloak.ClientSecret != "" {
		tokenManager = auth.NewServiceAccountTokenManager(
			cfg.Keycloak.URL,
			cfg.Keycloak.Realm,
			cfg.Keycloak.ClientID,
			cfg.Keycloak.ClientSecret,
		)
	}

	// Services
	cnaeValidationService := services.NewCNAEValidationService(rmiClient, legalEntitiesCache)

	var contactInfoService *services.ContactInfoService
	if tokenManager != nil {
		contactInfoService = services.NewContactInfoService(rmiClient, tokenManager, redisClient, cfg)
	}

	emailNotificationEnabled := cfg.DataRelay.BaseURL != "" && cfg.DataRelay.APIKey != ""
	emailNotificationService := services.NewEmailNotificationService(dataRelayClient, cursoRepo, orgaoSnapshotRepo, emailNotificationEnabled, cfg.PrefRio.Domain)

	// Citizen sync worker
	var citizenDataFetcher services.CitizenDataFetcher
	if cfg.CitizenSync.Enabled && tokenManager != nil {
		citizenSyncWorker := workers.NewCitizenSyncWorker(
			rmiClient, citizenSnapshotRepo, tokenManager, &cfg.CitizenSync,
		)
		citizenDataFetcher = citizenSyncWorker
		go func() {
			if err := citizenSyncWorker.Start(context.Background()); err != nil {
				log.Printf("[Router] Citizen sync worker stopped: %v", err)
			}
		}()
		log.Println("[Router] Citizen sync worker started")
	}

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
	empVagaService := empServices.NewVagaService(empVagaRepo, empEmpresaRepo, empEtapaRepo, empInformacaoComplementarRepo)
	empEtapaService := empServices.NewEtapaService(empEtapaRepo)
	empCurriculoService := empServices.NewCurriculoService(empCurriculoRepo)
	empCandidaturaService := empServices.NewCandidaturaService(empCandidaturaRepo, empVagaRepo, empCurriculoService)
	empOnboardingService := empServices.NewOnboardingService(empOnboardingRepo)
	empTermosUsoService := empServices.NewTermosUsoService(empTermosUsoRepo)

	var empCNPJConsultaService *empServices.CNPJConsultaService
	if tokenManager != nil {
		empCNPJConsultaService = empServices.NewCNPJConsultaService(rmiClient, tokenManager)
	}

	// Job processor
	jobs.InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)

	// Authorization
	var authChecker *authorization.Checker
	if cfg.Cerbos.Enabled {
		authChecker = authorization.NewChecker(
			cfg.Cerbos.Endpoint,
			time.Duration(cfg.Cerbos.Timeout)*time.Second,
		)
	}

	// Typesense
	typesenseHandler, err := v1.NewTypesenseHandler()
	if err != nil {
		log.Printf("Erro ao inicializar o Typesense: %v", err)
	}

	return &Dependencies{
		CourseHandler:    v1.NewCourseHandler(cursoService, inscricaoService, cursoRepo).WithCache(courseCache),
		InscricaoHandler: v1.NewInscricaoHandler(inscricaoService, jobService, cursoRepo),
		JobHandler:       v1.NewJobHandler(jobService),

		EmpregoHandler:       v1.NewEmpregoHandler(empregoService),
		AcessibilidadeHandler: v1.NewAcessibilidadeHandler(acessibilidadeService).WithCache(acessibilidadesCache),
		CategoriaHandler:     v1.NewCategoriaHandler(categoriaService).WithCache(categoriasCache),
		EmpresaHandler:       v1.NewEmpresaHandler(empresaService),
		EscolaridadeHandler:  v1.NewEscolaridadeHandler(escolaridadeService).WithCache(escolaridadesCache),
		InstituicaoHandler:   v1.NewInstituicaoHandler(instituicaoService),

		OportunidadeMEIHandler: v1.NewOportunidadeMEIHandler(oportunidadeMEIService),
		PropostaMEIHandler:     v1.NewPropostaMEIHandler(propostaMEIService, cnaeValidationService, authChecker, cfg),

		TypesenseHandler: typesenseHandler,

		EmpRegimeContratacaoHandler: empHandlers.NewRegimeContratacaoHandler(empRegimeContratacaoService),
		EmpModeloTrabalhoHandler:    empHandlers.NewModeloTrabalhoHandler(empModeloTrabalhoService),
		EmpTipoPCDHandler:           empHandlers.NewTipoPCDHandler(empTipoPCDService),
		EmpIdiomaHandler:            empHandlers.NewIdiomaHandler(empIdiomaService),
		EmpNivelIdiomaHandler:       empHandlers.NewNivelIdiomaHandler(empNivelIdiomaService),
		EmpEscolaridadeHandler:      empHandlers.NewEscolaridadeHandler(empEscolaridadeService),
		EmpTipoConquistaHandler:     empHandlers.NewTipoConquistaHandler(empTipoConquistaService),
		EmpSituacaoAtualHandler:     empHandlers.NewSituacaoAtualHandler(empSituacaoAtualService),
		EmpDisponibilidadeHandler:   empHandlers.NewDisponibilidadeHandler(empDisponibilidadeService),
		EmpEmpresaHandler:           empHandlers.NewEmpresaHandler(empEmpresaService).WithCNPJConsulta(empCNPJConsultaService),
		EmpVagaHandler:              empHandlers.NewVagaHandler(empVagaService),
		EmpEtapaHandler:             empHandlers.NewEtapaHandler(empEtapaService),
		EmpCandidaturaHandler:       empHandlers.NewCandidaturaHandler(empCandidaturaService),
		EmpCurriculoHandler:         empHandlers.NewCurriculoHandler(empCurriculoService),
		EmpOnboardingHandler:        empHandlers.NewOnboardingHandler(empOnboardingService),
		EmpTermosUsoHandler:         empHandlers.NewTermosUsoHandler(empTermosUsoService),
	}
}
