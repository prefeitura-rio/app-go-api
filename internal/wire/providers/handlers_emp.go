package providers

import (
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ProvideEmpRegimeContratacaoHandler creates empregabilidade RegimeContratacaoHandler
func ProvideEmpRegimeContratacaoHandler(service *empServices.RegimeContratacaoService) *empHandlers.RegimeContratacaoHandler {
	return empHandlers.NewRegimeContratacaoHandler(service)
}

// ProvideEmpModeloTrabalhoHandler creates empregabilidade ModeloTrabalhoHandler
func ProvideEmpModeloTrabalhoHandler(service *empServices.ModeloTrabalhoService) *empHandlers.ModeloTrabalhoHandler {
	return empHandlers.NewModeloTrabalhoHandler(service)
}

// ProvideEmpTipoPCDHandler creates empregabilidade TipoPCDHandler
func ProvideEmpTipoPCDHandler(service *empServices.TipoPCDService) *empHandlers.TipoPCDHandler {
	return empHandlers.NewTipoPCDHandler(service)
}

// ProvideEmpIdiomaHandler creates empregabilidade IdiomaHandler
func ProvideEmpIdiomaHandler(service *empServices.IdiomaService) *empHandlers.IdiomaHandler {
	return empHandlers.NewIdiomaHandler(service)
}

// ProvideEmpNivelIdiomaHandler creates empregabilidade NivelIdiomaHandler
func ProvideEmpNivelIdiomaHandler(service *empServices.NivelIdiomaService) *empHandlers.NivelIdiomaHandler {
	return empHandlers.NewNivelIdiomaHandler(service)
}

// ProvideEmpEscolaridadeHandler creates empregabilidade EscolaridadeHandler
func ProvideEmpEscolaridadeHandler(service *empServices.EscolaridadeService) *empHandlers.EscolaridadeHandler {
	return empHandlers.NewEscolaridadeHandler(service)
}

// ProvideEmpTipoConquistaHandler creates empregabilidade TipoConquistaHandler
func ProvideEmpTipoConquistaHandler(service *empServices.TipoConquistaService) *empHandlers.TipoConquistaHandler {
	return empHandlers.NewTipoConquistaHandler(service)
}

// ProvideEmpSituacaoAtualHandler creates empregabilidade SituacaoAtualHandler
func ProvideEmpSituacaoAtualHandler(service *empServices.SituacaoAtualService) *empHandlers.SituacaoAtualHandler {
	return empHandlers.NewSituacaoAtualHandler(service)
}

// ProvideEmpDisponibilidadeHandler creates empregabilidade DisponibilidadeHandler
func ProvideEmpDisponibilidadeHandler(service *empServices.DisponibilidadeService) *empHandlers.DisponibilidadeHandler {
	return empHandlers.NewDisponibilidadeHandler(service)
}

// ProvideEmpEmpresaHandler creates empregabilidade EmpresaHandler with optional CNPJ consulta support
func ProvideEmpEmpresaHandler(service *empServices.EmpresaService, cnpjConsultaService *empServices.CNPJConsultaService) *empHandlers.EmpresaHandler {
	return empHandlers.NewEmpresaHandler(service).WithCNPJConsulta(cnpjConsultaService)
}

// ProvideEmpVagaHandler creates empregabilidade VagaHandler
func ProvideEmpVagaHandler(service *empServices.VagaService) *empHandlers.VagaHandler {
	return empHandlers.NewVagaHandler(service)
}

// ProvideEmpEtapaHandler creates empregabilidade EtapaHandler
func ProvideEmpEtapaHandler(service *empServices.EtapaService) *empHandlers.EtapaHandler {
	return empHandlers.NewEtapaHandler(service)
}

// ProvideEmpCandidaturaHandler creates empregabilidade CandidaturaHandler
func ProvideEmpCandidaturaHandler(service *empServices.CandidaturaService) *empHandlers.CandidaturaHandler {
	return empHandlers.NewCandidaturaHandler(service)
}

// ProvideEmpCurriculoHandler creates empregabilidade CurriculoHandler
func ProvideEmpCurriculoHandler(service *empServices.CurriculoService) *empHandlers.CurriculoHandler {
	return empHandlers.NewCurriculoHandler(service)
}

// ProvideEmpOnboardingHandler creates empregabilidade OnboardingHandler
func ProvideEmpOnboardingHandler(service *empServices.OnboardingService) *empHandlers.OnboardingHandler {
	return empHandlers.NewOnboardingHandler(service)
}

// ProvideEmpTermosUsoHandler creates empregabilidade TermosUsoHandler
func ProvideEmpTermosUsoHandler(service *empServices.TermosUsoService) *empHandlers.TermosUsoHandler {
	return empHandlers.NewTermosUsoHandler(service)
}

// ProvideEmpZonaHandler creates empregabilidade ZonaHandler
func ProvideEmpZonaHandler(service *empServices.ZonaService) *empHandlers.ZonaHandler {
	return empHandlers.NewZonaHandler(service)
}

// ProvideEmpCandidaturaBloqueioHandler creates empregabilidade CandidaturaBloqueioHandler
func ProvideEmpCandidaturaBloqueioHandler(service *empServices.CandidaturaBloqueioService) *empHandlers.CandidaturaBloqueioHandler {
	return empHandlers.NewCandidaturaBloqueioHandler(service)
}
