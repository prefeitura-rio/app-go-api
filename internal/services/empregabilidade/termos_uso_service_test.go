package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockTermosUsoRepository struct{}

func (m *mockTermosUsoRepository) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.TermosUso, error) {
	return nil, nil
}

func (m *mockTermosUsoRepository) Upsert(ctx context.Context, entity *empregabilidade.TermosUso) error {
	return nil
}

func (m *mockTermosUsoRepository) AcceptTerms(ctx context.Context, cpf string) error {
	return nil
}

func TestNewTermosUsoService(t *testing.T) {
	mockRepo := &mockTermosUsoRepository{}
	service := services.NewTermosUsoServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}
