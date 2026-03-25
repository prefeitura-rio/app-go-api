package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type EscolaridadeService struct {
	repo EmpEscolaridadeRepositoryInterface
}

func NewEscolaridadeService(repo *repository.EscolaridadeRepository) *EscolaridadeService {
	return &EscolaridadeService{repo: repo}
}

func NewEscolaridadeServiceWithInterface(repo EmpEscolaridadeRepositoryInterface) *EscolaridadeService {
	return &EscolaridadeService{repo: repo}
}

func (s *EscolaridadeService) Create(ctx context.Context, entity *empregabilidade.Escolaridade) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *EscolaridadeService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Escolaridade, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EscolaridadeService) Update(ctx context.Context, entity *empregabilidade.Escolaridade) error {
	return s.repo.Update(ctx, entity)
}

func (s *EscolaridadeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *EscolaridadeService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Escolaridade, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
