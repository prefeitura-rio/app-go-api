package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type SituacaoAtualService struct {
	repo *repository.SituacaoAtualRepository
}

func NewSituacaoAtualService(repo *repository.SituacaoAtualRepository) *SituacaoAtualService {
	return &SituacaoAtualService{repo: repo}
}

func (s *SituacaoAtualService) Create(ctx context.Context, entity *empregabilidade.SituacaoAtual) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *SituacaoAtualService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.SituacaoAtual, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SituacaoAtualService) Update(ctx context.Context, entity *empregabilidade.SituacaoAtual) error {
	return s.repo.Update(ctx, entity)
}

func (s *SituacaoAtualService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SituacaoAtualService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.SituacaoAtual, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
