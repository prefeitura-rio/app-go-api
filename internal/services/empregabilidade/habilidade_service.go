package empregabilidade

import (
	"context"

	"github.com/google/uuid"
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

// Create cria uma nova habilidade e retorna o ID gerado (UUID)
func (s *HabilidadeService) CreateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) (uuid.UUID, error) {
	return s.repo.CreateHabilidade(ctx, entity)
}

// GetByID busca uma habilidade pelo seu UUID
func (s *HabilidadeService) GetHabilidadeByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Habilidade, error) {
	return s.repo.GetHabilidadeByID(ctx, id)
}

// Update atualiza os dados de uma habilidade
func (s *HabilidadeService) UpdateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) error {
	return s.repo.UpdateHabilidade(ctx, entity)
}

// Delete remove uma habilidade pelo seu UUID
func (s *HabilidadeService) DeleteHabilidade(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteHabilidade(ctx, id)
}

// List retorna uma lista paginada de habilidades com suporte a filtros
func (s *HabilidadeService) ListHabilidades(ctx context.Context, filter empregabilidade.HabilidadeFilter, page, pageSize int) ([]*empregabilidade.Habilidade, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListHabilidades(ctx, filter, pageSize, offset)
}

// ListAreas retorna uma lista paginada das áreas de atuação vinculadas às habilidades
func (s *HabilidadeService) ListAreas(ctx context.Context, filter empregabilidade.AreaAtuacaoFilter, page, pageSize int) ([]*empregabilidade.AreaAtuacao, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListAreas(ctx, filter, pageSize, offset)
}

// AddHabilidadeAoCurriculo vincula uma habilidade ao currículo do candidato
func (s *HabilidadeService) AddHabilidadeAoCurriculo(ctx context.Context, vinculo *empregabilidade.CurriculoHabilidade) error {
	return s.repo.AddHabilidadeAoCurriculo(ctx, vinculo)
}

// ListHabilidadesPorCPF busca todas as habilidades associadas ao CPF do candidato
func (s *HabilidadeService) ListHabilidadesPorCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoHabilidade, error) {
	return s.repo.ListHabilidadesPorCPF(ctx, cpf)
}
