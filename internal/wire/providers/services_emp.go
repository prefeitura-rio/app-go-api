package providers

import (
	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ProvideEmpRegimeContratacaoService creates empregabilidade RegimeContratacaoService
func ProvideEmpRegimeContratacaoService(repo *empRepository.RegimeContratacaoRepository) *empServices.RegimeContratacaoService {
	return empServices.NewRegimeContratacaoService(repo)
}

// ProvideEmpModeloTrabalhoService creates empregabilidade ModeloTrabalhoService
func ProvideEmpModeloTrabalhoService(repo *empRepository.ModeloTrabalhoRepository) *empServices.ModeloTrabalhoService {
	return empServices.NewModeloTrabalhoService(repo)
}

// ProvideEmpTipoPCDService creates empregabilidade TipoPCDService
func ProvideEmpTipoPCDService(repo *empRepository.TipoPCDRepository) *empServices.TipoPCDService {
	return empServices.NewTipoPCDService(repo)
}

// ProvideEmpIdiomaService creates empregabilidade IdiomaService
func ProvideEmpIdiomaService(repo *empRepository.IdiomaRepository) *empServices.IdiomaService {
	return empServices.NewIdiomaService(repo)
}

// ProvideEmpNivelIdiomaService creates empregabilidade NivelIdiomaService
func ProvideEmpNivelIdiomaService(repo *empRepository.NivelIdiomaRepository) *empServices.NivelIdiomaService {
	return empServices.NewNivelIdiomaService(repo)
}

// ProvideEmpEscolaridadeService creates empregabilidade EscolaridadeService
func ProvideEmpEscolaridadeService(repo *empRepository.EscolaridadeRepository) *empServices.EscolaridadeService {
	return empServices.NewEscolaridadeService(repo)
}

// ProvideEmpTipoConquistaService creates empregabilidade TipoConquistaService
func ProvideEmpTipoConquistaService(repo *empRepository.TipoConquistaRepository) *empServices.TipoConquistaService {
	return empServices.NewTipoConquistaService(repo)
}

// ProvideEmpSituacaoAtualService creates empregabilidade SituacaoAtualService
func ProvideEmpSituacaoAtualService(repo *empRepository.SituacaoAtualRepository) *empServices.SituacaoAtualService {
	return empServices.NewSituacaoAtualService(repo)
}

// ProvideEmpDisponibilidadeService creates empregabilidade DisponibilidadeService
func ProvideEmpDisponibilidadeService(repo *empRepository.DisponibilidadeRepository) *empServices.DisponibilidadeService {
	return empServices.NewDisponibilidadeService(repo)
}

// ProvideEmpEmpresaService creates empregabilidade EmpresaService
func ProvideEmpEmpresaService(repo *empRepository.EmpresaRepository) *empServices.EmpresaService {
	return empServices.NewEmpresaService(repo)
}

// ProvideEmpVagaService creates empregabilidade VagaService
func ProvideEmpVagaService(
	vagaRepo *empRepository.VagaRepository,
	empresaRepo *empRepository.EmpresaRepository,
	candidaturaRepo *empRepository.CandidaturaRepository,
) *empServices.VagaService {
	return empServices.NewVagaService(vagaRepo, empresaRepo, candidaturaRepo)
}

// ProvideEmpEtapaService creates empregabilidade EtapaService
func ProvideEmpEtapaService(repo *empRepository.EtapaRepository) *empServices.EtapaService {
	return empServices.NewEtapaService(repo)
}

// ProvideEmpCurriculoService creates empregabilidade CurriculoService
func ProvideEmpCurriculoService(repo *empRepository.CurriculoRepository) *empServices.CurriculoService {
	return empServices.NewCurriculoService(repo)
}

// ProvideEmpCandidaturaService creates empregabilidade CandidaturaService.
// citizenSnapshotRepo and citizenDataFetcher are passed from the core domain.
// citizenDataFetcher may be nil when citizen sync is disabled.
func ProvideEmpCandidaturaService(
	candidaturaRepo *empRepository.CandidaturaRepository,
	vagaRepo *empRepository.VagaRepository,
	curriculoService *empServices.CurriculoService,
	citizenSnapshotRepo *repository.CitizenSnapshotRepository,
	emailNotificationService *services.EmailNotificationService,
) *empServices.CandidaturaService {
	// citizenDataFetcher is nil; citizen sync worker is started separately in the router
	return empServices.NewCandidaturaService(candidaturaRepo, vagaRepo, curriculoService, citizenSnapshotRepo, nil, emailNotificationService)
}

// ProvideEmpOnboardingService creates empregabilidade OnboardingService
func ProvideEmpOnboardingService(repo *empRepository.OnboardingRepository) *empServices.OnboardingService {
	return empServices.NewOnboardingService(repo)
}

// ProvideEmpTermosUsoService creates empregabilidade TermosUsoService
func ProvideEmpTermosUsoService(repo *empRepository.TermosUsoRepository) *empServices.TermosUsoService {
	return empServices.NewTermosUsoService(repo)
}

// ProvideEmpCNPJConsultaService creates empregabilidade CNPJConsultaService.
// Returns nil when the token manager is not configured (Keycloak not set up).
func ProvideEmpCNPJConsultaService(rmiClient *clients.RMIClient, tokenManager *auth.ServiceAccountTokenManager) *empServices.CNPJConsultaService {
	if tokenManager == nil {
		return nil
	}
	return empServices.NewCNPJConsultaService(rmiClient, tokenManager)
}
