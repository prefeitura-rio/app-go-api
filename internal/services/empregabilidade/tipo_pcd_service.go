package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type TipoPCDService struct {
	repo *repository.TipoPCDRepository
}

func NewTipoPCDService(repo *repository.TipoPCDRepository) *TipoPCDService {
	return &TipoPCDService{repo: repo}
}

func (s *TipoPCDService) Create(ctx context.Context, entity *empregabilidade.TipoPCD) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *TipoPCDService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoPCD, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TipoPCDService) Update(ctx context.Context, entity *empregabilidade.TipoPCD) error {
	return s.repo.Update(ctx, entity)
}

func (s *TipoPCDService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *TipoPCDService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.TipoPCD, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
