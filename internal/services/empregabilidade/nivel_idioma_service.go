package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type NivelIdiomaService struct {
	repo *repository.NivelIdiomaRepository
}

func NewNivelIdiomaService(repo *repository.NivelIdiomaRepository) *NivelIdiomaService {
	return &NivelIdiomaService{repo: repo}
}

func (s *NivelIdiomaService) Create(ctx context.Context, entity *empregabilidade.NivelIdioma) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *NivelIdiomaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.NivelIdioma, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *NivelIdiomaService) Update(ctx context.Context, entity *empregabilidade.NivelIdioma) error {
	return s.repo.Update(ctx, entity)
}

func (s *NivelIdiomaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *NivelIdiomaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.NivelIdioma, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
