package empregabilidade

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type HabilidadeService struct {
	repo HabilidadeRepositoryInterface
}

func NewHabilidadeService(repo *repository.HabilidadeRepository) *HabilidadeService {
	return &HabilidadeService{repo: repo}
}

func NewHabilidadeServiceWithInterface(repo HabilidadeRepositoryInterface) *HabilidadeService {
	return &HabilidadeService{repo: repo}
}

// CreateHabilidade cria uma nova habilidade e retorna o ID gerado (int64)
func (s *HabilidadeService) CreateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) (int64, error) {
	return s.repo.CreateHabilidade(ctx, entity)
}

// GetHabilidadeByID busca uma habilidade pelo seu ID (int64)
func (s *HabilidadeService) GetHabilidadeByID(ctx context.Context, id int64) (*empregabilidade.Habilidade, error) {
	return s.repo.GetHabilidadeByID(ctx, id)
}

// UpdateHabilidade atualiza os dados de uma habilidade
func (s *HabilidadeService) UpdateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) error {
	return s.repo.UpdateHabilidade(ctx, entity)
}

// DeleteHabilidade remove uma habilidade pelo seu ID (int64)
func (s *HabilidadeService) DeleteHabilidade(ctx context.Context, id int64) error {
	return s.repo.DeleteHabilidade(ctx, id)
}

// ListHabilidades retorna uma lista paginada de habilidades com suporte a filtros
func (s *HabilidadeService) ListHabilidades(ctx context.Context, filter empregabilidade.HabilidadeFilter, page, pageSize int) ([]*empregabilidade.Habilidade, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListHabilidades(ctx, filter, pageSize, offset)
}

// CreateAreaAtuacao cria uma nova área de atuação e retorna o ID gerado (int64)
func (s *HabilidadeService) CreateAreaAtuacao(ctx context.Context, entity *empregabilidade.AreaAtuacao) (int64, error) {
	return s.repo.CreateAreaAtuacao(ctx, entity)
}

// GetAreaAtuacaoByID busca uma área de atuação pelo seu ID (int64)
func (s *HabilidadeService) GetAreaAtuacaoByID(ctx context.Context, id int64) (*empregabilidade.AreaAtuacao, error) {
	return s.repo.GetAreaAtuacaoByID(ctx, id)
}

// UpdateAreaAtuacao atualiza os dados de uma área de atuação
func (s *HabilidadeService) UpdateAreaAtuacao(ctx context.Context, entity *empregabilidade.AreaAtuacao) error {
	return s.repo.UpdateAreaAtuacao(ctx, entity)
}

// DeleteAreaAtuacao remove uma área de atuação pelo seu ID (int64)
func (s *HabilidadeService) DeleteAreaAtuacao(ctx context.Context, id int64) error {
	return s.repo.DeleteAreaAtuacao(ctx, id)
}

// ListAreasAtuacao retorna uma lista paginada das áreas de atuação vinculadas às habilidades
func (s *HabilidadeService) ListAreasAtuacao(ctx context.Context, filter empregabilidade.AreaAtuacaoFilter, page, pageSize int) ([]*empregabilidade.AreaAtuacao, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListAreasAtuacao(ctx, filter, pageSize, offset)
}

// AddHabilidadeAoCurriculo vincula uma habilidade ao currículo do candidato
func (s *HabilidadeService) AddHabilidadeAoCurriculo(ctx context.Context, vinculo *empregabilidade.CurriculoHabilidade) error {
	return s.repo.AddHabilidadeAoCurriculo(ctx, vinculo)
}

// ListHabilidadesPorCPF busca todas as habilidades associadas ao CPF do candidato
func (s *HabilidadeService) ListHabilidadesPorCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoHabilidade, error) {
	return s.repo.ListHabilidadesPorCPF(ctx, cpf)
}
