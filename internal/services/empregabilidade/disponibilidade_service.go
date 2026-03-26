package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type DisponibilidadeService struct {
	repo DisponibilidadeRepositoryInterface
}

func NewDisponibilidadeService(repo *repository.DisponibilidadeRepository) *DisponibilidadeService {
	return &DisponibilidadeService{repo: repo}
}

func NewDisponibilidadeServiceWithInterface(repo DisponibilidadeRepositoryInterface) *DisponibilidadeService {
	return &DisponibilidadeService{repo: repo}
}

func (s *DisponibilidadeService) Create(ctx context.Context, entity *empregabilidade.Disponibilidade) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *DisponibilidadeService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Disponibilidade, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DisponibilidadeService) Update(ctx context.Context, entity *empregabilidade.Disponibilidade) error {
	return s.repo.Update(ctx, entity)
}

func (s *DisponibilidadeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *DisponibilidadeService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Disponibilidade, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
