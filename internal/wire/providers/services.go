package providers

import (
	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// ProvideCategoriaService creates CategoriaService
func ProvideCategoriaService(repo repository.CategoriaRepositoryInterface) *services.CategoriaService {
	return services.NewCategoriaService(repo)
}

// ProvideCursoService creates CursoService
func ProvideCursoService(repo *repository.CursoRepository) *services.CursoService {
	return services.NewCursoService(repo)
}

// ProvideEmpregoService creates EmpregoService
func ProvideEmpregoService(repo *repository.EmpregoRepository) *services.EmpregoService {
	return services.NewEmpregoService(repo)
}

// ProvideAcessibilidadeService creates AcessibilidadeService
func ProvideAcessibilidadeService(repo *repository.AcessibilidadeRepository) *services.AcessibilidadeService {
	// AcessibilidadeRepository implements AcessibilidadeRepositoryInterface
	return services.NewAcessibilidadeService(repo)
}

// ProvideEscolaridadeService creates EscolaridadeService
func ProvideEscolaridadeService(repo *repository.EscolaridadeRepository) *services.EscolaridadeService {
	// EscolaridadeRepository implements EscolaridadeRepositoryInterface
	return services.NewEscolaridadeService(repo)
}

// ProvideEmpresaService creates the core EmpresaService
func ProvideEmpresaService(repo *repository.EmpresaRepository) *services.EmpresaService {
	return services.NewEmpresaService(repo)
}

// ProvideInstituicaoService creates InstituicaoService
func ProvideInstituicaoService(repo *repository.InstituicaoRepository) *services.InstituicaoService {
	return services.NewInstituicaoService(repo)
}

// ProvideJobService creates JobService
func ProvideJobService(repo *repository.JobRepository) *services.JobService {
	return services.NewJobService(repo)
}

// ProvideOportunidadeMEIService creates OportunidadeMEIService
func ProvideOportunidadeMEIService(repo *repository.OportunidadeMEIRepository) *services.OportunidadeMEIService {
	return services.NewOportunidadeMEIService(repo)
}

// ProvideCNAEValidationService creates CNAEValidationService using the RMI client and legal entities cache
func ProvideCNAEValidationService(rmiClient *clients.RMIClient, legalEntitiesCache *cache.LegalEntitiesCache) *services.CNAEValidationService {
	return services.NewCNAEValidationService(rmiClient, legalEntitiesCache)
}

// ProvideContactInfoService creates ContactInfoService.
// Returns nil if the token manager is not configured (Keycloak not set up).
func ProvideContactInfoService(rmiClient *clients.RMIClient, tokenManager *auth.ServiceAccountTokenManager, redisClient *redis.Client, cfg *config.AppConfig) *services.ContactInfoService {
	if tokenManager == nil {
		return nil
	}
	return services.NewContactInfoService(rmiClient, tokenManager, redisClient, cfg)
}

// ProvideEmailNotificationService creates EmailNotificationService.
// Email sending is enabled only when DataRelay base URL and API key are both configured.
func ProvideEmailNotificationService(
	dataRelayClient *clients.DataRelayClient,
	cursoRepo *repository.CursoRepository,
	orgaoSnapshotRepo *repository.OrgaoSnapshotRepository,
	citizenSnapshotRepo *repository.CitizenSnapshotRepository,
	cfg *config.AppConfig,
) *services.EmailNotificationService {
	enabled := cfg.DataRelay.BaseURL != "" && cfg.DataRelay.APIKey != ""
	return services.NewEmailNotificationService(
		dataRelayClient,
		cursoRepo,
		orgaoSnapshotRepo,
		citizenSnapshotRepo,
		enabled,
		cfg.PrefRio.Domain,
	)
}

// ProvideInscricaoService creates InscricaoService.
// citizenDataFetcher may be nil when citizen sync is disabled.
func ProvideInscricaoService(
	inscricaoRepo *repository.InscricaoRepository,
	cursoRepo *repository.CursoRepository,
	citizenSnapshotRepo *repository.CitizenSnapshotRepository,
	emailNotificationService *services.EmailNotificationService,
	cfg *config.AppConfig,
) *services.InscricaoService {
	// citizenDataFetcher is left nil; the citizen sync worker is started separately
	// in the router and implements the CitizenDataFetcher interface
	return services.NewInscricaoService(
		inscricaoRepo,
		cursoRepo,
		citizenSnapshotRepo,
		nil,
		emailNotificationService,
		cfg,
	)
}

// ProvidePropostaMEIService creates PropostaMEIService
func ProvidePropostaMEIService(
	propostaRepo *repository.PropostaMEIRepository,
	oportunidadeRepo *repository.OportunidadeMEIRepository,
	cnaeValidationService *services.CNAEValidationService,
	contactInfoService *services.ContactInfoService,
) *services.PropostaMEIService {
	return services.NewPropostaMEIService(propostaRepo, oportunidadeRepo, cnaeValidationService, contactInfoService)
}
