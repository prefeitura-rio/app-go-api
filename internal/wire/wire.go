//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/wire/providers"
)

// ApplicationContainer holds all wired components for the full application
type ApplicationContainer struct {
	DB     *gorm.DB
	Config *config.AppConfig

	// Infrastructure
	RedisClient     *redis.Client
	RMIClient       *clients.RMIClient
	TokenManager    *auth.ServiceAccountTokenManager
	DataRelayClient *clients.DataRelayClient

	// Caches
	ReferenceCaches    *providers.ReferenceCaches
	LegalEntitiesCache *cache.LegalEntitiesCache
	CourseCache        *cache.CourseCache

	// Core Repositories
	CursoRepo           *repository.CursoRepository
	InscricaoRepo       *repository.InscricaoRepository
	EmpregoRepo         *repository.EmpregoRepository
	AcessibilidadeRepo  *repository.AcessibilidadeRepository
	CategoriaRepo       repository.CategoriaRepositoryInterface
	EscolaridadeRepo    *repository.EscolaridadeRepository
	EmpresaRepo         *repository.EmpresaRepository
	InstituicaoRepo     *repository.InstituicaoRepository
	JobRepo             *repository.JobRepository
	OportunidadeMEIRepo *repository.OportunidadeMEIRepository
	PropostaMEIRepo     *repository.PropostaMEIRepository
	OrgaoSnapshotRepo   *repository.OrgaoSnapshotRepository
	CitizenSnapshotRepo *repository.CitizenSnapshotRepository

	// Empregabilidade Repositories
	EmpRegimeContratacaoRepo *empRepository.RegimeContratacaoRepository
	EmpModeloTrabalhoRepo    *empRepository.ModeloTrabalhoRepository
	EmpTipoPCDRepo           *empRepository.TipoPCDRepository
	EmpIdiomaRepo            *empRepository.IdiomaRepository
	EmpNivelIdiomaRepo       *empRepository.NivelIdiomaRepository
	EmpEscolaridadeRepo      *empRepository.EscolaridadeRepository
	EmpTipoConquistaRepo     *empRepository.TipoConquistaRepository
	EmpSituacaoAtualRepo     *empRepository.SituacaoAtualRepository
	EmpDisponibilidadeRepo   *empRepository.DisponibilidadeRepository
	EmpEmpresaRepo           *empRepository.EmpresaRepository
	EmpVagaRepo              *empRepository.VagaRepository
	EmpEtapaRepo             *empRepository.EtapaRepository
	EmpCandidaturaRepo       *empRepository.CandidaturaRepository
	EmpCurriculoRepo         *empRepository.CurriculoRepository
	EmpOnboardingRepo        *empRepository.OnboardingRepository
	EmpTermosUsoRepo         *empRepository.TermosUsoRepository

	// Core Services
	CursoService             *services.CursoService
	InscricaoService         *services.InscricaoService
	EmpregoService           *services.EmpregoService
	AcessibilidadeService    *services.AcessibilidadeService
	CategoriaService         *services.CategoriaService
	EscolaridadeService      *services.EscolaridadeService
	EmpresaService           *services.EmpresaService
	InstituicaoService       *services.InstituicaoService
	JobService               *services.JobService
	OportunidadeMEIService   *services.OportunidadeMEIService
	PropostaMEIService       *services.PropostaMEIService
	CNAEValidationService    *services.CNAEValidationService
	ContactInfoService       *services.ContactInfoService
	EmailNotificationService *services.EmailNotificationService

	// Empregabilidade Services
	EmpRegimeContratacaoService *empServices.RegimeContratacaoService
	EmpModeloTrabalhoService    *empServices.ModeloTrabalhoService
	EmpTipoPCDService           *empServices.TipoPCDService
	EmpIdiomaService            *empServices.IdiomaService
	EmpNivelIdiomaService       *empServices.NivelIdiomaService
	EmpEscolaridadeService      *empServices.EscolaridadeService
	EmpTipoConquistaService     *empServices.TipoConquistaService
	EmpSituacaoAtualService     *empServices.SituacaoAtualService
	EmpDisponibilidadeService   *empServices.DisponibilidadeService
	EmpEmpresaService           *empServices.EmpresaService
	EmpVagaService              *empServices.VagaService
	EmpEtapaService             *empServices.EtapaService
	EmpCurriculoService         *empServices.CurriculoService
	EmpCandidaturaService       *empServices.CandidaturaService
	EmpOnboardingService        *empServices.OnboardingService
	EmpTermosUsoService         *empServices.TermosUsoService
	EmpCNPJConsultaService      *empServices.CNPJConsultaService

	// Core Handlers
	EmpregoHandler         *v1.EmpregoHandler
	AcessibilidadeHandler  *v1.AcessibilidadeHandler
	CategoriaHandler       *v1.CategoriaHandler
	EmpresaHandler         *v1.EmpresaHandler
	EscolaridadeHandler    *v1.EscolaridadeHandler
	InstituicaoHandler     *v1.InstituicaoHandler
	InscricaoHandler       *v1.InscricaoHandler
	CourseHandler          *v1.CourseHandler
	JobHandler             *v1.JobHandler
	OportunidadeMEIHandler *v1.OportunidadeMEIHandler
	PropostaMEIHandler     *v1.PropostaMEIHandler
	TypesenseHandler       *v1.TypesenseHandler

	// Empregabilidade Handlers
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

// CategoriaContainer holds wired components for Categorias proof-of-concept
type CategoriaContainer struct {
	DB               *gorm.DB
	Config           *config.AppConfig
	CategoriaHandler *v1.CategoriaHandler
}

// Provider sets

var DatabaseSet = wire.NewSet(
	providers.ProvideDatabase,
)

var InfrastructureSet = wire.NewSet(
	providers.ProvideDatabase,
	providers.ProvideRedisClient,
	providers.ProvideRMIClient,
	providers.ProvideDataRelayClient,
	providers.ProvideServiceAccountTokenManager,
)

var CacheSet = wire.NewSet(
	providers.ProvideLegalEntitiesCache,
	providers.ProvideReferenceCaches,
	providers.ProvideCourseCache,
)

var CoreRepositorySet = wire.NewSet(
	providers.ProvideCategoriaRepository,
	providers.ProvideCursoRepository,
	providers.ProvideInscricaoRepository,
	providers.ProvideEmpregoRepository,
	providers.ProvideAcessibilidadeRepository,
	providers.ProvideEscolaridadeRepository,
	providers.ProvideEmpresaRepository,
	providers.ProvideInstituicaoRepository,
	providers.ProvideJobRepository,
	providers.ProvideOportunidadeMEIRepository,
	providers.ProvidePropostaMEIRepository,
	providers.ProvideOrgaoSnapshotRepository,
	providers.ProvideCitizenSnapshotRepository,
)

var EmpRepositorySet = wire.NewSet(
	providers.ProvideEmpRegimeContratacaoRepository,
	providers.ProvideEmpModeloTrabalhoRepository,
	providers.ProvideEmpTipoPCDRepository,
	providers.ProvideEmpIdiomaRepository,
	providers.ProvideEmpNivelIdiomaRepository,
	providers.ProvideEmpEscolaridadeRepository,
	providers.ProvideEmpTipoConquistaRepository,
	providers.ProvideEmpSituacaoAtualRepository,
	providers.ProvideEmpDisponibilidadeRepository,
	providers.ProvideEmpEmpresaRepository,
	providers.ProvideEmpVagaRepository,
	providers.ProvideEmpEtapaRepository,
	providers.ProvideEmpCandidaturaRepository,
	providers.ProvideEmpCurriculoRepository,
	providers.ProvideEmpOnboardingRepository,
	providers.ProvideEmpTermosUsoRepository,
)

var CoreServiceSet = wire.NewSet(
	providers.ProvideCategoriaService,
	providers.ProvideCursoService,
	providers.ProvideEmpregoService,
	providers.ProvideAcessibilidadeService,
	providers.ProvideEscolaridadeService,
	providers.ProvideEmpresaService,
	providers.ProvideInstituicaoService,
	providers.ProvideJobService,
	providers.ProvideOportunidadeMEIService,
	providers.ProvideCNAEValidationService,
	providers.ProvideContactInfoService,
	providers.ProvideEmailNotificationService,
	providers.ProvideInscricaoService,
	providers.ProvidePropostaMEIService,
)

var EmpServiceSet = wire.NewSet(
	providers.ProvideEmpRegimeContratacaoService,
	providers.ProvideEmpModeloTrabalhoService,
	providers.ProvideEmpTipoPCDService,
	providers.ProvideEmpIdiomaService,
	providers.ProvideEmpNivelIdiomaService,
	providers.ProvideEmpEscolaridadeService,
	providers.ProvideEmpTipoConquistaService,
	providers.ProvideEmpSituacaoAtualService,
	providers.ProvideEmpDisponibilidadeService,
	providers.ProvideEmpEmpresaService,
	providers.ProvideEmpVagaService,
	providers.ProvideEmpEtapaService,
	providers.ProvideEmpCurriculoService,
	providers.ProvideEmpCandidaturaService,
	providers.ProvideEmpOnboardingService,
	providers.ProvideEmpTermosUsoService,
	providers.ProvideEmpCNPJConsultaService,
)

var CoreHandlerSet = wire.NewSet(
	providers.ProvideEmpregoHandler,
	providers.ProvideAcessibilidadeHandler,
	providers.ProvideCategoriaHandlerWithCache,
	providers.ProvideEmpresaHandler,
	providers.ProvideEscolaridadeHandler,
	providers.ProvideInstituicaoHandler,
	providers.ProvideInscricaoHandler,
	providers.ProvideCourseHandler,
	providers.ProvideJobHandler,
	providers.ProvideOportunidadeMEIHandler,
	providers.ProvidePropostaMEIHandler,
	providers.ProvideTypesenseHandler,
)

var EmpHandlerSet = wire.NewSet(
	providers.ProvideEmpRegimeContratacaoHandler,
	providers.ProvideEmpModeloTrabalhoHandler,
	providers.ProvideEmpTipoPCDHandler,
	providers.ProvideEmpIdiomaHandler,
	providers.ProvideEmpNivelIdiomaHandler,
	providers.ProvideEmpEscolaridadeHandler,
	providers.ProvideEmpTipoConquistaHandler,
	providers.ProvideEmpSituacaoAtualHandler,
	providers.ProvideEmpDisponibilidadeHandler,
	providers.ProvideEmpEmpresaHandler,
	providers.ProvideEmpVagaHandler,
	providers.ProvideEmpEtapaHandler,
	providers.ProvideEmpCandidaturaHandler,
	providers.ProvideEmpCurriculoHandler,
	providers.ProvideEmpOnboardingHandler,
	providers.ProvideEmpTermosUsoHandler,
)

// Legacy set kept for backward compatibility with the Categorias POC
var CategoriaSet = wire.NewSet(
	providers.ProvideCategoriaRepository,
	providers.ProvideCategoriaService,
	v1.NewCategoriaHandler,
)

// InitializeCategoriaContainer wires up Categorias domain for proof-of-concept
func InitializeCategoriaContainer(cfg *config.AppConfig) (*CategoriaContainer, error) {
	wire.Build(
		DatabaseSet,
		CategoriaSet,
		wire.Struct(new(CategoriaContainer), "*"),
	)
	return &CategoriaContainer{}, nil
}

// InitializeApplication wires up the full application container
func InitializeApplication(cfg *config.AppConfig) (*ApplicationContainer, error) {
	wire.Build(
		InfrastructureSet,
		CacheSet,
		CoreRepositorySet,
		EmpRepositorySet,
		CoreServiceSet,
		EmpServiceSet,
		CoreHandlerSet,
		EmpHandlerSet,
		wire.Struct(new(ApplicationContainer), "*"),
	)
	return &ApplicationContainer{}, nil
}
