package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// RegimeContratacaoRepositoryInterface defines the interface for RegimeContratacao repository.
type RegimeContratacaoRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.RegimeContratacao) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.RegimeContratacao, error)
	Update(ctx context.Context, entity *empregabilidade.RegimeContratacao) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.RegimeContratacao, int, error)
}

// ModeloTrabalhoRepositoryInterface defines the interface for ModeloTrabalho repository.
type ModeloTrabalhoRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.ModeloTrabalho) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.ModeloTrabalho, error)
	Update(ctx context.Context, entity *empregabilidade.ModeloTrabalho) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.ModeloTrabalho, int, error)
}

// TipoPCDRepositoryInterface defines the interface for TipoPCD repository.
type TipoPCDRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.TipoPCD) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoPCD, error)
	Update(ctx context.Context, entity *empregabilidade.TipoPCD) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoPCD, int, error)
}

// IdiomaRepositoryInterface defines the interface for Idioma repository.
type IdiomaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Idioma) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Idioma, error)
	Update(ctx context.Context, entity *empregabilidade.Idioma) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Idioma, int, error)
}

// NivelIdiomaRepositoryInterface defines the interface for NivelIdioma repository.
type NivelIdiomaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.NivelIdioma) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.NivelIdioma, error)
	Update(ctx context.Context, entity *empregabilidade.NivelIdioma) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.NivelIdioma, int, error)
}

// EmpEscolaridadeRepositoryInterface defines the interface for empregabilidade Escolaridade repository.
type EmpEscolaridadeRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Escolaridade) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Escolaridade, error)
	Update(ctx context.Context, entity *empregabilidade.Escolaridade) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Escolaridade, int, error)
}

// ZonaRepositoryInterface defines the interface for Zona repository.
type ZonaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Zona) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Zona, error)
	Update(ctx context.Context, entity *empregabilidade.Zona) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Zona, int, error)
}

// CandidaturaBloqueioRepositoryInterface defines the interface for CandidaturaBloqueio repository.
type CandidaturaBloqueioRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.CandidaturaBloqueio) (uuid.UUID, error)
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.CandidaturaBloqueio, int, error)
}

// TipoConquistaRepositoryInterface defines the interface for TipoConquista repository.
type TipoConquistaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.TipoConquista) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoConquista, error)
	Update(ctx context.Context, entity *empregabilidade.TipoConquista) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoConquista, int, error)
}

// SituacaoAtualRepositoryInterface defines the interface for SituacaoAtual repository.
type SituacaoAtualRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.SituacaoAtual) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.SituacaoAtual, error)
	Update(ctx context.Context, entity *empregabilidade.SituacaoAtual) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.SituacaoAtual, int, error)
}

// DisponibilidadeRepositoryInterface defines the interface for Disponibilidade repository.
type DisponibilidadeRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Disponibilidade) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Disponibilidade, error)
	Update(ctx context.Context, entity *empregabilidade.Disponibilidade) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Disponibilidade, int, error)
}

// EmpresaRepositoryInterface defines the interface for empregabilidade Empresa repository.
type EmpresaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Empresa) (string, error)
	GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error)
	Update(ctx context.Context, entity *empregabilidade.Empresa) error
	Delete(ctx context.Context, cnpj string) error
	List(ctx context.Context, filter empregabilidade.EmpresaFilter, limit, offset int) ([]*empregabilidade.Empresa, int, error)
	Upsert(ctx context.Context, entity *empregabilidade.Empresa) error
}

// EtapaRepositoryInterface defines the interface for Etapa repository.
type EtapaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Etapa) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error)
	Update(ctx context.Context, entity *empregabilidade.Etapa) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error)
	DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error
}

// OnboardingRepositoryInterface defines the interface for Onboarding repository.
type OnboardingRepositoryInterface interface {
	GetByCPF(ctx context.Context, cpf string) (*empregabilidade.Onboarding, error)
	Upsert(ctx context.Context, entity *empregabilidade.Onboarding) error
	MarkFirstLoginCompleted(ctx context.Context, cpf string) error
}

// TermosUsoRepositoryInterface defines the interface for TermosUso repository.
type TermosUsoRepositoryInterface interface {
	GetByCPF(ctx context.Context, cpf string) (*empregabilidade.TermosUso, error)
	Upsert(ctx context.Context, entity *empregabilidade.TermosUso) error
	AcceptTerms(ctx context.Context, cpf string) error
}

// CurriculoRepositoryInterface defines the interface for Curriculo repository.
type CurriculoRepositoryInterface interface {
	CreateFormacao(ctx context.Context, entity *empregabilidade.CurriculoFormacao) (uuid.UUID, error)
	GetFormacaoByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoFormacao, error)
	UpdateFormacao(ctx context.Context, entity *empregabilidade.CurriculoFormacao) error
	DeleteFormacao(ctx context.Context, id uuid.UUID) error
	ListFormacoesByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoFormacao, error)

	CreateIdioma(ctx context.Context, entity *empregabilidade.CurriculoIdioma) (uuid.UUID, error)
	GetIdiomaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoIdioma, error)
	UpdateIdioma(ctx context.Context, entity *empregabilidade.CurriculoIdioma) error
	DeleteIdioma(ctx context.Context, id uuid.UUID) error
	ListIdiomasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoIdioma, error)

	CreateCursoComplementar(ctx context.Context, entity *empregabilidade.CurriculoCursoComplementar) (uuid.UUID, error)
	GetCursoComplementarByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoCursoComplementar, error)
	UpdateCursoComplementar(ctx context.Context, entity *empregabilidade.CurriculoCursoComplementar) error
	DeleteCursoComplementar(ctx context.Context, id uuid.UUID) error
	ListCursosComplementaresByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoCursoComplementar, error)

	CreateExperiencia(ctx context.Context, entity *empregabilidade.CurriculoExperiencia) (uuid.UUID, error)
	GetExperienciaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoExperiencia, error)
	UpdateExperiencia(ctx context.Context, entity *empregabilidade.CurriculoExperiencia) error
	DeleteExperiencia(ctx context.Context, id uuid.UUID) error
	ListExperienciasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoExperiencia, error)

	CreateConquista(ctx context.Context, entity *empregabilidade.CurriculoConquista) (uuid.UUID, error)
	GetConquistaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoConquista, error)
	UpdateConquista(ctx context.Context, entity *empregabilidade.CurriculoConquista) error
	DeleteConquista(ctx context.Context, id uuid.UUID) error
	ListConquistasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoConquista, error)

	ReplaceAllFormacoesByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoFormacao) error
	ReplaceAllFormacaoAccordionByCPF(ctx context.Context, cpf string, formacoes []*empregabilidade.CurriculoFormacao, idiomas []*empregabilidade.CurriculoIdioma) error
	ReplaceAllExperienciasByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoExperiencia) error
	ReplaceAllExperienciaProfissionalAccordionByCPF(ctx context.Context, cpf string, experiencias []*empregabilidade.CurriculoExperiencia, conquistas []*empregabilidade.CurriculoConquista) error
	ReplaceAllConquistasByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoConquista) error
	ReplaceAllIdiomasByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoIdioma) error
	ReplaceAllCursosComplementaresByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoCursoComplementar) error

	UpsertSituacaoInteresses(ctx context.Context, entity *empregabilidade.CurriculoSituacaoInteresses) error
	GetSituacaoInteressesByCPF(ctx context.Context, cpf string) (*empregabilidade.CurriculoSituacaoInteresses, error)
}
