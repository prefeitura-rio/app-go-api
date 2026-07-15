package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type CandidaturaBloqueioService struct {
	repo CandidaturaBloqueioRepositoryInterface
}

func NewCandidaturaBloqueioService(repo *repository.CandidaturaBloqueioRepository) *CandidaturaBloqueioService {
	return &CandidaturaBloqueioService{repo: repo}
}

func NewCandidaturaBloqueioServiceWithInterface(repo CandidaturaBloqueioRepositoryInterface) *CandidaturaBloqueioService {
	return &CandidaturaBloqueioService{repo: repo}
}

func (s *CandidaturaBloqueioService) Create(ctx context.Context, entity *empregabilidade.CandidaturaBloqueio) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *CandidaturaBloqueioService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.CandidaturaBloqueio, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
