package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type ModeloTrabalhoService struct {
	repo *repository.ModeloTrabalhoRepository
}

func NewModeloTrabalhoService(repo *repository.ModeloTrabalhoRepository) *ModeloTrabalhoService {
	return &ModeloTrabalhoService{repo: repo}
}

func (s *ModeloTrabalhoService) Create(ctx context.Context, entity *empregabilidade.ModeloTrabalho) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *ModeloTrabalhoService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.ModeloTrabalho, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ModeloTrabalhoService) Update(ctx context.Context, entity *empregabilidade.ModeloTrabalho) error {
	return s.repo.Update(ctx, entity)
}

func (s *ModeloTrabalhoService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ModeloTrabalhoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.ModeloTrabalho, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
