package empregabilidade

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

const MaxInformacoesComplementaresPorVaga = 5

type InformacaoComplementarService struct {
	repo *repository.InformacaoComplementarRepository
}

func NewInformacaoComplementarService(repo *repository.InformacaoComplementarRepository) *InformacaoComplementarService {
	return &InformacaoComplementarService{repo: repo}
}

func (s *InformacaoComplementarService) Create(ctx context.Context, entity *empregabilidade.InformacaoComplementar) (uuid.UUID, error) {
	existing, err := s.repo.ListByVaga(ctx, entity.IDVaga)
	if err != nil {
		return uuid.Nil, err
	}

	if len(existing) >= MaxInformacoesComplementaresPorVaga {
		return uuid.Nil, errors.New("limite máximo de 5 informações complementares por vaga atingido")
	}

	return s.repo.Create(ctx, entity)
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
