package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type ZonaService struct {
	repo ZonaRepositoryInterface
}

func NewZonaService(repo *repository.ZonaRepository) *ZonaService {
	return &ZonaService{repo: repo}
}

func NewZonaServiceWithInterface(repo ZonaRepositoryInterface) *ZonaService {
	return &ZonaService{repo: repo}
}

func (s *ZonaService) Create(ctx context.Context, entity *empregabilidade.Zona) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *ZonaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Zona, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ZonaService) Update(ctx context.Context, entity *empregabilidade.Zona) error {
	return s.repo.Update(ctx, entity)
}

func (s *ZonaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ZonaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Zona, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
