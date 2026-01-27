package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type CurriculoService struct {
	repo *repository.CurriculoRepository
}

func NewCurriculoService(repo *repository.CurriculoRepository) *CurriculoService {
	return &CurriculoService{repo: repo}
}

// Formação Acadêmica

func (s *CurriculoService) CreateFormacao(ctx context.Context, entity *empregabilidade.CurriculoFormacao) (uuid.UUID, error) {
	return s.repo.CreateFormacao(ctx, entity)
}

func (s *CurriculoService) GetFormacaoByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoFormacao, error) {
	return s.repo.GetFormacaoByID(ctx, id)
}

func (s *CurriculoService) UpdateFormacao(ctx context.Context, entity *empregabilidade.CurriculoFormacao) error {
	return s.repo.UpdateFormacao(ctx, entity)
}

func (s *CurriculoService) DeleteFormacao(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteFormacao(ctx, id)
}

func (s *CurriculoService) ListFormacoesByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoFormacao, error) {
	return s.repo.ListFormacoesByCPF(ctx, cpf)
}

// Idiomas

func (s *CurriculoService) CreateIdioma(ctx context.Context, entity *empregabilidade.CurriculoIdioma) (uuid.UUID, error) {
	return s.repo.CreateIdioma(ctx, entity)
}

func (s *CurriculoService) GetIdiomaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoIdioma, error) {
	return s.repo.GetIdiomaByID(ctx, id)
}

func (s *CurriculoService) UpdateIdioma(ctx context.Context, entity *empregabilidade.CurriculoIdioma) error {
	return s.repo.UpdateIdioma(ctx, entity)
}

func (s *CurriculoService) DeleteIdioma(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteIdioma(ctx, id)
}

func (s *CurriculoService) ListIdiomasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoIdioma, error) {
	return s.repo.ListIdiomasByCPF(ctx, cpf)
}

// Cursos Complementares

func (s *CurriculoService) CreateCursoComplementar(ctx context.Context, entity *empregabilidade.CurriculoCursoComplementar) (uuid.UUID, error) {
	return s.repo.CreateCursoComplementar(ctx, entity)
}

func (s *CurriculoService) GetCursoComplementarByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoCursoComplementar, error) {
	return s.repo.GetCursoComplementarByID(ctx, id)
}

func (s *CurriculoService) UpdateCursoComplementar(ctx context.Context, entity *empregabilidade.CurriculoCursoComplementar) error {
	return s.repo.UpdateCursoComplementar(ctx, entity)
}

func (s *CurriculoService) DeleteCursoComplementar(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCursoComplementar(ctx, id)
}

func (s *CurriculoService) ListCursosComplementaresByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoCursoComplementar, error) {
	return s.repo.ListCursosComplementaresByCPF(ctx, cpf)
}

// Experiências

func (s *CurriculoService) CreateExperiencia(ctx context.Context, entity *empregabilidade.CurriculoExperiencia) (uuid.UUID, error) {
	return s.repo.CreateExperiencia(ctx, entity)
}

func (s *CurriculoService) GetExperienciaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoExperiencia, error) {
	return s.repo.GetExperienciaByID(ctx, id)
}

func (s *CurriculoService) UpdateExperiencia(ctx context.Context, entity *empregabilidade.CurriculoExperiencia) error {
	return s.repo.UpdateExperiencia(ctx, entity)
}

func (s *CurriculoService) DeleteExperiencia(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteExperiencia(ctx, id)
}

func (s *CurriculoService) ListExperienciasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoExperiencia, error) {
	return s.repo.ListExperienciasByCPF(ctx, cpf)
}

// Conquistas

func (s *CurriculoService) CreateConquista(ctx context.Context, entity *empregabilidade.CurriculoConquista) (uuid.UUID, error) {
	return s.repo.CreateConquista(ctx, entity)
}

func (s *CurriculoService) GetConquistaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoConquista, error) {
	return s.repo.GetConquistaByID(ctx, id)
}

func (s *CurriculoService) UpdateConquista(ctx context.Context, entity *empregabilidade.CurriculoConquista) error {
	return s.repo.UpdateConquista(ctx, entity)
}

func (s *CurriculoService) DeleteConquista(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteConquista(ctx, id)
}

func (s *CurriculoService) ListConquistasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoConquista, error) {
	return s.repo.ListConquistasByCPF(ctx, cpf)
}

// Situação e Interesses

func (s *CurriculoService) UpsertSituacaoInteresses(ctx context.Context, entity *empregabilidade.CurriculoSituacaoInteresses) error {
	return s.repo.UpsertSituacaoInteresses(ctx, entity)
}

func (s *CurriculoService) GetSituacaoInteressesByCPF(ctx context.Context, cpf string) (*empregabilidade.CurriculoSituacaoInteresses, error) {
	return s.repo.GetSituacaoInteressesByCPF(ctx, cpf)
}

// Currículo Completo

type CurriculoCompleto struct {
	Formacoes           []*empregabilidade.CurriculoFormacao           `json:"formacoes"`
	Idiomas             []*empregabilidade.CurriculoIdioma             `json:"idiomas"`
	CursosComplementares []*empregabilidade.CurriculoCursoComplementar `json:"cursos_complementares"`
	Experiencias        []*empregabilidade.CurriculoExperiencia        `json:"experiencias"`
	Conquistas          []*empregabilidade.CurriculoConquista          `json:"conquistas"`
	SituacaoInteresses  *empregabilidade.CurriculoSituacaoInteresses   `json:"situacao_interesses"`
}

func (s *CurriculoService) GetCurriculoCompleto(ctx context.Context, cpf string) (*CurriculoCompleto, error) {
	formacoes, err := s.repo.ListFormacoesByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	idiomas, err := s.repo.ListIdiomasByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	cursos, err := s.repo.ListCursosComplementaresByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	experiencias, err := s.repo.ListExperienciasByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	conquistas, err := s.repo.ListConquistasByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	situacao, err := s.repo.GetSituacaoInteressesByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	return &CurriculoCompleto{
		Formacoes:            formacoes,
		Idiomas:              idiomas,
		CursosComplementares: cursos,
		Experiencias:         experiencias,
		Conquistas:           conquistas,
		SituacaoInteresses:   situacao,
	}, nil
}
