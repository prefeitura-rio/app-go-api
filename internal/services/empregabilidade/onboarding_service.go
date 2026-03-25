package empregabilidade

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type OnboardingService struct {
	repo OnboardingRepositoryInterface
}

func NewOnboardingService(repo *repository.OnboardingRepository) *OnboardingService {
	return &OnboardingService{repo: repo}
}

func NewOnboardingServiceWithInterface(repo OnboardingRepositoryInterface) *OnboardingService {
	return &OnboardingService{repo: repo}
}

func (s *OnboardingService) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.Onboarding, error) {
	return s.repo.GetByCPF(ctx, cpf)
}

func (s *OnboardingService) Upsert(ctx context.Context, entity *empregabilidade.Onboarding) error {
	return s.repo.Upsert(ctx, entity)
}

func (s *OnboardingService) MarkFirstLoginCompleted(ctx context.Context, cpf string) error {
	return s.repo.MarkFirstLoginCompleted(ctx, cpf)
}

func (s *OnboardingService) IsFirstLogin(ctx context.Context, cpf string) (bool, error) {
	onboarding, err := s.repo.GetByCPF(ctx, cpf)
	if err != nil {
		return false, err
	}
	if onboarding == nil {
		return true, nil
	}
	return onboarding.IsEmpregabilidadeFirstLogin, nil
}
