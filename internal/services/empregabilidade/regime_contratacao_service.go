package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type RegimeContratacaoService struct {
	repo *repository.RegimeContratacaoRepository
}

func NewRegimeContratacaoService(repo *repository.RegimeContratacaoRepository) *RegimeContratacaoService {
	return &RegimeContratacaoService{repo: repo}
}

func (s *RegimeContratacaoService) Create(ctx context.Context, entity *empregabilidade.RegimeContratacao) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *RegimeContratacaoService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.RegimeContratacao, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RegimeContratacaoService) Update(ctx context.Context, entity *empregabilidade.RegimeContratacao) error {
	return s.repo.Update(ctx, entity)
}

func (s *RegimeContratacaoService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *RegimeContratacaoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.RegimeContratacao, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
