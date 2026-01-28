package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type IdiomaService struct {
	repo *repository.IdiomaRepository
}

func NewIdiomaService(repo *repository.IdiomaRepository) *IdiomaService {
	return &IdiomaService{repo: repo}
}

func (s *IdiomaService) Create(ctx context.Context, entity *empregabilidade.Idioma) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *IdiomaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Idioma, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *IdiomaService) Update(ctx context.Context, entity *empregabilidade.Idioma) error {
	return s.repo.Update(ctx, entity)
}

func (s *IdiomaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *IdiomaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Idioma, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
