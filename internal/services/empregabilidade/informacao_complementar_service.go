package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

const MaxInformacoesComplementaresPorVaga = 5

type InformacaoComplementarRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.InformacaoComplementar) (uuid.UUID, error)
	CreateWithLimitCheck(ctx context.Context, entity *empregabilidade.InformacaoComplementar, maxLimit int) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.InformacaoComplementar, error)
	Update(ctx context.Context, entity *empregabilidade.InformacaoComplementar) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.InformacaoComplementar, error)
	DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error
}

type InformacaoComplementarService struct {
	repo InformacaoComplementarRepositoryInterface
}

func NewInformacaoComplementarService(repo InformacaoComplementarRepositoryInterface) *InformacaoComplementarService {
	return &InformacaoComplementarService{repo: repo}
}

func (s *InformacaoComplementarService) Create(ctx context.Context, entity *empregabilidade.InformacaoComplementar) (uuid.UUID, error) {
	return s.repo.CreateWithLimitCheck(ctx, entity, MaxInformacoesComplementaresPorVaga)
}

func (s *InformacaoComplementarService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.InformacaoComplementar, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *InformacaoComplementarService) Update(ctx context.Context, entity *empregabilidade.InformacaoComplementar) error {
	return s.repo.Update(ctx, entity)
}

func (s *InformacaoComplementarService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *InformacaoComplementarService) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.InformacaoComplementar, error) {
	return s.repo.ListByVaga(ctx, vagaID)
}

func (s *InformacaoComplementarService) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	return s.repo.DeleteByVaga(ctx, vagaID)
}
