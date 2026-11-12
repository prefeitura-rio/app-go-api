package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O PUT da candidatura é do candidato; o último evento visto é da API: o
// Save da edição não pode apagá-lo (as atualizações voltariam a aparecer).
func TestCandidaturaService_Update_PreservaUltimoEventoVisto(t *testing.T) {
	repo := NewMockCandidaturaRepo()
	id := uuid.New()
	visto := int64(42)
	repo.candidaturas[id] = &empregabilidade.Candidatura{
		ID:                id,
		CPF:               "12345678900",
		UltimoEventoVisto: &visto,
	}
	service := services.NewCandidaturaService(repo, NewMockVagaRepo(), NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

	require.NoError(t, service.Update(context.Background(), &empregabilidade.Candidatura{ID: id}))

	result := repo.candidaturas[id]
	require.NotNil(t, result.UltimoEventoVisto)
	assert.Equal(t, visto, *result.UltimoEventoVisto)
}

// Candidatura nova começa sem evento visto.
func TestCandidaturaService_Create_ComecaSemEventoVisto(t *testing.T) {
	repo := NewMockCandidaturaRepo()
	vagaRepo := NewMockVagaRepo()
	vagaID := uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{ID: vagaID, Status: empregabilidade.StatusVagaPublicadoAtivo}
	service := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

	visto := int64(7)
	id, err := service.Create(context.Background(), &empregabilidade.Candidatura{
		CPF:               "12345678900",
		IDVaga:            vagaID,
		UltimoEventoVisto: &visto,
	})
	require.NoError(t, err)

	assert.Nil(t, repo.candidaturas[id].UltimoEventoVisto)
}
