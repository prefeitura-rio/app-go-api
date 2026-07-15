package providers

import (
	"gorm.io/gorm"

	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// ProvideEmpRegimeContratacaoRepository creates empregabilidade RegimeContratacaoRepository
func ProvideEmpRegimeContratacaoRepository(db *gorm.DB) *empRepository.RegimeContratacaoRepository {
	return empRepository.NewRegimeContratacaoRepository(db)
}

// ProvideEmpModeloTrabalhoRepository creates empregabilidade ModeloTrabalhoRepository
func ProvideEmpModeloTrabalhoRepository(db *gorm.DB) *empRepository.ModeloTrabalhoRepository {
	return empRepository.NewModeloTrabalhoRepository(db)
}

// ProvideEmpTipoPCDRepository creates empregabilidade TipoPCDRepository
func ProvideEmpTipoPCDRepository(db *gorm.DB) *empRepository.TipoPCDRepository {
	return empRepository.NewTipoPCDRepository(db)
}

// ProvideEmpIdiomaRepository creates empregabilidade IdiomaRepository
func ProvideEmpIdiomaRepository(db *gorm.DB) *empRepository.IdiomaRepository {
	return empRepository.NewIdiomaRepository(db)
}

// ProvideEmpNivelIdiomaRepository creates empregabilidade NivelIdiomaRepository
func ProvideEmpNivelIdiomaRepository(db *gorm.DB) *empRepository.NivelIdiomaRepository {
	return empRepository.NewNivelIdiomaRepository(db)
}

// ProvideEmpEscolaridadeRepository creates empregabilidade EscolaridadeRepository
func ProvideEmpEscolaridadeRepository(db *gorm.DB) *empRepository.EscolaridadeRepository {
	return empRepository.NewEscolaridadeRepository(db)
}

// ProvideEmpTipoConquistaRepository creates empregabilidade TipoConquistaRepository
func ProvideEmpTipoConquistaRepository(db *gorm.DB) *empRepository.TipoConquistaRepository {
	return empRepository.NewTipoConquistaRepository(db)
}

// ProvideEmpSituacaoAtualRepository creates empregabilidade SituacaoAtualRepository
func ProvideEmpSituacaoAtualRepository(db *gorm.DB) *empRepository.SituacaoAtualRepository {
	return empRepository.NewSituacaoAtualRepository(db)
}

// ProvideEmpDisponibilidadeRepository creates empregabilidade DisponibilidadeRepository
func ProvideEmpDisponibilidadeRepository(db *gorm.DB) *empRepository.DisponibilidadeRepository {
	return empRepository.NewDisponibilidadeRepository(db)
}

// ProvideEmpEmpresaRepository creates empregabilidade EmpresaRepository
func ProvideEmpEmpresaRepository(db *gorm.DB) *empRepository.EmpresaRepository {
	return empRepository.NewEmpresaRepository(db)
}

// ProvideEmpVagaRepository creates empregabilidade VagaRepository
func ProvideEmpVagaRepository(db *gorm.DB) *empRepository.VagaRepository {
	return empRepository.NewVagaRepository(db)
}

// ProvideEmpEtapaRepository creates empregabilidade EtapaRepository
func ProvideEmpEtapaRepository(db *gorm.DB) *empRepository.EtapaRepository {
	return empRepository.NewEtapaRepository(db)
}

// ProvideEmpCandidaturaRepository creates empregabilidade CandidaturaRepository
func ProvideEmpCandidaturaRepository(db *gorm.DB) *empRepository.CandidaturaRepository {
	return empRepository.NewCandidaturaRepository(db)
}

// ProvideEmpCurriculoRepository creates empregabilidade CurriculoRepository
func ProvideEmpCurriculoRepository(db *gorm.DB) *empRepository.CurriculoRepository {
	return empRepository.NewCurriculoRepository(db)
}

// ProvideEmpOnboardingRepository creates empregabilidade OnboardingRepository
func ProvideEmpOnboardingRepository(db *gorm.DB) *empRepository.OnboardingRepository {
	return empRepository.NewOnboardingRepository(db)
}

// ProvideEmpTermosUsoRepository creates empregabilidade TermosUsoRepository
func ProvideEmpTermosUsoRepository(db *gorm.DB) *empRepository.TermosUsoRepository {
	return empRepository.NewTermosUsoRepository(db)
}

// ProvideEmpInformacaoComplementarRepository creates empregabilidade InformacaoComplementarRepository
func ProvideEmpInformacaoComplementarRepository(db *gorm.DB) *empRepository.InformacaoComplementarRepository {
	return empRepository.NewInformacaoComplementarRepository(db)
}

// ProvideEmpZonaRepository creates empregabilidade ZonaRepository
func ProvideEmpZonaRepository(db *gorm.DB) *empRepository.ZonaRepository {
	return empRepository.NewZonaRepository(db)
}

// ProvideEmpCandidaturaBloqueioRepository creates empregabilidade CandidaturaBloqueioRepository
func ProvideEmpCandidaturaBloqueioRepository(db *gorm.DB) *empRepository.CandidaturaBloqueioRepository {
	return empRepository.NewCandidaturaBloqueioRepository(db)
}
