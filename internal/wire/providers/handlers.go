package providers

import (
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/authorization"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

func ProvideEmpregoHandler(service *services.EmpregoService, rmiClient *clients.RMIClient, tokenManager *auth.ServiceAccountTokenManager) *v1.EmpregoHandler {
	return v1.NewEmpregoHandler(service, rmiClient, tokenManager)
}

// ProvideAcessibilidadeHandler creates AcessibilidadeHandler with cache
func ProvideAcessibilidadeHandler(service *services.AcessibilidadeService, caches *ReferenceCaches) *v1.AcessibilidadeHandler {
	return v1.NewAcessibilidadeHandler(service).WithCache(caches.Acessibilidades)
}

// ProvideCategoriaHandlerWithCache creates CategoriaHandler with cache
func ProvideCategoriaHandlerWithCache(service *services.CategoriaService, caches *ReferenceCaches) *v1.CategoriaHandler {
	return v1.NewCategoriaHandler(service).WithCache(caches.Categorias)
}

// ProvideEmpresaHandler creates the core EmpresaHandler
func ProvideEmpresaHandler(service *services.EmpresaService) *v1.EmpresaHandler {
	return v1.NewEmpresaHandler(service)
}

// ProvideEscolaridadeHandler creates EscolaridadeHandler with cache
func ProvideEscolaridadeHandler(service *services.EscolaridadeService, caches *ReferenceCaches) *v1.EscolaridadeHandler {
	return v1.NewEscolaridadeHandler(service).WithCache(caches.Escolaridades)
}

// ProvideInstituicaoHandler creates InstituicaoHandler
func ProvideInstituicaoHandler(service *services.InstituicaoService) *v1.InstituicaoHandler {
	return v1.NewInstituicaoHandler(service)
}

// ProvideInscricaoHandler creates InscricaoHandler
func ProvideInscricaoHandler(service *services.InscricaoService, jobService *services.JobService, cursoRepo *repository.CursoRepository) *v1.InscricaoHandler {
	return v1.NewInscricaoHandler(service, jobService, cursoRepo)
}

func ProvideCourseHandler(cursoService *services.CursoService, inscricaoService *services.InscricaoService, cursoRepo *repository.CursoRepository, courseCache *cache.CourseCache, rmiClient *clients.RMIClient, tokenManager *auth.ServiceAccountTokenManager) *v1.CourseHandler {
	return v1.NewCourseHandler(cursoService, inscricaoService, cursoRepo, rmiClient, tokenManager).WithCache(courseCache)
}

// ProvideJobHandler creates JobHandler
func ProvideJobHandler(service *services.JobService) *v1.JobHandler {
	return v1.NewJobHandler(service)
}

// ProvideOportunidadeMEIHandler creates OportunidadeMEIHandler
func ProvideOportunidadeMEIHandler(service *services.OportunidadeMEIService) *v1.OportunidadeMEIHandler {
	return v1.NewOportunidadeMEIHandler(service)
}

// ProvidePropostaMEIHandler creates PropostaMEIHandler with optional Cerbos authorization
func ProvidePropostaMEIHandler(
	service *services.PropostaMEIService,
	cnaeValidationService *services.CNAEValidationService,
	cfg *config.AppConfig,
) *v1.PropostaMEIHandler {
	var authChecker *authorization.Checker
	if cfg.Cerbos.Enabled {
		authChecker = authorization.NewChecker(
			cfg.Cerbos.Endpoint,
			time.Duration(cfg.Cerbos.Timeout)*time.Second,
		)
	}
	return v1.NewPropostaMEIHandler(service, cnaeValidationService, authChecker, cfg)
}

// ProvideTypesenseHandler creates TypesenseHandler. Returns nil when
// Typesense is not configured; callers must guard against a nil handler.
func ProvideTypesenseHandler() *v1.TypesenseHandler {
	handler, err := v1.NewTypesenseHandler()
	if err != nil {
		return nil
	}
	return handler
}
