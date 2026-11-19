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
	// Argumentos recebidos na última chamada a List.
	gotOrgaoIDs []string
	gotLimit    int
	gotOffset   int
}

func (m *mockCandidaturaBloqueioRepository) Create(ctx context.Context, entity *empregabilidade.CandidaturaBloqueio) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.Nil, m.createErr
	}
	return uuid.New(), nil
}

func (m *mockCandidaturaBloqueioRepository) List(ctx context.Context, filter map[string]interface{}, orgaoParceiroIDs []string, limit, offset int) ([]*empregabilidade.CandidaturaBloqueio, int, error) {
	m.gotOrgaoIDs = orgaoParceiroIDs
	m.gotLimit = limit
	m.gotOffset = offset
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.items, m.total, nil
}

type mockVagaRepoForBloqueio struct {
	exists bool
	err    error
}

func (m *mockVagaRepoForBloqueio) ExistsByID(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.exists, m.err
}

func vagaExistente() *mockVagaRepoForBloqueio {
	return &mockVagaRepoForBloqueio{exists: true}
}

func TestNewCandidaturaBloqueioService(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, vagaExistente())
	assert.NotNil(t, service)
}

func TestCandidaturaBloqueioService_Create(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, vagaExistente())

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
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, vagaExistente())

	id, err := service.Create(context.Background(), &empregabilidade.CandidaturaBloqueio{CPF: "12345678900"})

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}

func TestCandidaturaBloqueioService_Create_VagaInexistente(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, &mockVagaRepoForBloqueio{})

	id, err := service.Create(context.Background(), &empregabilidade.CandidaturaBloqueio{CPF: "12345678900", IDVaga: uuid.New()})

	assert.ErrorIs(t, err, services.ErrBloqueioVagaNaoEncontrada)
	assert.Equal(t, uuid.Nil, id)
}

func TestCandidaturaBloqueioService_Create_ErroAoBuscarVaga(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{}
	dbErr := errors.New("db down")
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, &mockVagaRepoForBloqueio{err: dbErr})

	id, err := service.Create(context.Background(), &empregabilidade.CandidaturaBloqueio{CPF: "12345678900", IDVaga: uuid.New()})

	assert.ErrorIs(t, err, dbErr)
	assert.Equal(t, uuid.Nil, id)
}

func TestCandidaturaBloqueioService_List(t *testing.T) {
	vagaID := uuid.New()
	mockItems := []*empregabilidade.CandidaturaBloqueio{
		{ID: uuid.New(), CPF: "12345678900", IDVaga: vagaID},
		{ID: uuid.New(), CPF: "98765432100", IDVaga: vagaID},
	}
	mockRepo := &mockCandidaturaBloqueioRepository{items: mockItems, total: 2}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, vagaExistente())

	items, total, err := service.List(context.Background(), nil, nil, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)
}

func TestCandidaturaBloqueioService_List_RepoError(t *testing.T) {
	mockRepo := &mockCandidaturaBloqueioRepository{listErr: errors.New("db error")}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, vagaExistente())

	items, total, err := service.List(context.Background(), nil, nil, 1, 10)

	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, items)
}

func TestCandidaturaBloqueioService_List_RepassaEscopoDeOrgao(t *testing.T) {
	// DEV-518: o escopo de órgão parceiro do editor chega intacto ao
	// repositório, com a paginação convertida em limit/offset.
	mockRepo := &mockCandidaturaBloqueioRepository{}
	service := services.NewCandidaturaBloqueioServiceWithInterface(mockRepo, vagaExistente())

	_, _, err := service.List(context.Background(), nil, []string{"1", "2600"}, 3, 20)

	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2600"}, mockRepo.gotOrgaoIDs)
	assert.Equal(t, 20, mockRepo.gotLimit)
	assert.Equal(t, 40, mockRepo.gotOffset)
}
