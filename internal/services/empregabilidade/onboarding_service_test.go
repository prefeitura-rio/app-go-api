package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockOnboardingRepository struct{}

func (m *mockOnboardingRepository) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.Onboarding, error) {
	return nil, nil
}

func (m *mockOnboardingRepository) Upsert(ctx context.Context, entity *empregabilidade.Onboarding) error {
	return nil
}

func (m *mockOnboardingRepository) MarkFirstLoginCompleted(ctx context.Context, cpf string) error {
	return nil
}

func TestNewOnboardingService(t *testing.T) {
	mockRepo := &mockOnboardingRepository{}
	service := services.NewOnboardingServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
