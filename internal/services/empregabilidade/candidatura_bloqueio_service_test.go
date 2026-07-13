package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

type mockCandidaturaBloqueioRepository struct {
	items     []*empregabilidade.CandidaturaBloqueio
	total     int
	createErr error
	listErr   error
}

func (m *mockCandidaturaBloqueioRepository) Create(ctx context.Context, entity *empregabilidade.CandidaturaBloqueio) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.Nil, m.createErr
	}
	return uuid.New(), nil
}

func (m *mockCandidaturaBloqueioRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.CandidaturaBloqueio, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.items, m.total, nil
}

func TestNewCandidaturaBloqueioService(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo)
	assert.NotNil(t, service)
}

func TestCandidaturaBloqueioService_Create(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo)

	id, err := service.Create(context.Background(), &empregabilidade.CandidaturaBloqueio{
		CPF:                   "12345678900",
		IDVaga:                uuid.New(),
		CriteriosNaoAtendidos: []string{"idade_minima"},
	})

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestCandidaturaBloqueioService_Create_RepoError(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{createErr: errors.New("db error")}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo)

	id, err := service.Create(context.Background(), &empregabilidade.CandidaturaBloqueio{CPF: "12345678900"})

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}

func TestCandidaturaBloqueioService_List(t *testing.T) {
	vagaID := uuid.New()
	mockItems := []*empregabilidade.CandidaturaBloqueio{
		{ID: uuid.New(), CPF: "12345678900", IDVaga: vagaID},
		{ID: uuid.New(), CPF: "98765432100", IDVaga: vagaID},
	}
	mockRepo := &mockCandidaturaBloqueioRepository{items: mockItems, total: 2}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
}

func TestCandidaturaBloqueioService_List_RepoError(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{listErr: errors.New("db error")}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo)

	items, total, err := service.List(context.Background(), nil, 1, 10)

	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, items)
}
