package empregabilidade

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type TermosUsoService struct {
	repo TermosUsoRepositoryInterface
}

func NewTermosUsoService(repo *repository.TermosUsoRepository) *TermosUsoService {
	return &TermosUsoService{repo: repo}
}

func NewTermosUsoServiceWithInterface(repo TermosUsoRepositoryInterface) *TermosUsoService {
	return &TermosUsoService{repo: repo}
}

func (s *TermosUsoService) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.TermosUso, error) {
	return s.repo.GetByCPF(ctx, cpf)
}

func (s *TermosUsoService) Upsert(ctx context.Context, entity *empregabilidade.TermosUso) error {
	return s.repo.Upsert(ctx, entity)
}

func (s *TermosUsoService) AcceptTerms(ctx context.Context, cpf string) error {
	return s.repo.AcceptTerms(ctx, cpf)
}

func (s *TermosUsoService) HasAcceptedTerms(ctx context.Context, cpf string) (bool, error) {
	termos, err := s.repo.GetByCPF(ctx, cpf)
	if err != nil {
		return false, err
	}
	if termos == nil {
		return false, nil
	}
	return termos.UserConsent, nil
}
