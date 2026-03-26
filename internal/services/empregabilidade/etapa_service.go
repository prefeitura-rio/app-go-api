package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type EtapaService struct {
	repo EtapaRepositoryInterface
}

func NewEtapaService(repo *repository.EtapaRepository) *EtapaService {
	return &EtapaService{repo: repo}
}

func NewEtapaServiceWithInterface(repo EtapaRepositoryInterface) *EtapaService {
	return &EtapaService{repo: repo}
}

func (s *EtapaService) Create(ctx context.Context, entity *empregabilidade.Etapa) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *EtapaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EtapaService) Update(ctx context.Context, entity *empregabilidade.Etapa) error {
	return s.repo.Update(ctx, entity)
}

func (s *EtapaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *EtapaService) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error) {
	return s.repo.ListByVaga(ctx, vagaID)
}

func (s *EtapaService) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	return s.repo.DeleteByVaga(ctx, vagaID)
}
