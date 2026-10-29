package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOrigemCadastro struct {
	cpfs []string
	err  error
}

func (f *fakeOrigemCadastro) RegistrarOrigemCadastro(_ context.Context, cpf string) error {
	f.cpfs = append(f.cpfs, cpf)
	return f.err
}

func candidaturaServiceComOrigem(t *testing.T, registrar *fakeOrigemCadastro) (*services.CandidaturaService, uuid.UUID) {
	t.Helper()
	vagaRepo := NewMockVagaRepo()
	vagaID := uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{ID: vagaID, Status: empregabilidade.StatusVagaPublicadoAtivo}
	s := services.NewCandidaturaService(NewMockCandidaturaRepo(), vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())
	s.SetOrigemCadastroRegistrar(registrar)
	return s, vagaID
}

func TestCandidaturaService_Create_RegistraOrigemCadastro(t *testing.T) {
	for _, origem := range []empregabilidade.OrigemCandidatura{
		empregabilidade.OrigemCandidaturaPortalCidadao,
		empregabilidade.OrigemCandidaturaSecretaria,
	} {
		t.Run(string(origem), func(t *testing.T) {
			registrar := &fakeOrigemCadastro{}
			s, vagaID := candidaturaServiceComOrigem(t, registrar)

			_, err := s.Create(context.Background(), &empregabilidade.Candidatura{CPF: "12345678901", IDVaga: vagaID, Origem: &origem})

			require.NoError(t, err)
			assert.Equal(t, []string{"12345678901"}, registrar.cpfs)
		})
	}
}

func TestCandidaturaService_Create_AssociacaoNaoRegistraOrigemCadastro(t *testing.T) {
	registrar := &fakeOrigemCadastro{}
	s, vagaID := candidaturaServiceComOrigem(t, registrar)
	origem := empregabilidade.OrigemCandidaturaBancoCurriculos

	_, err := s.CreateSemNotificacao(context.Background(), &empregabilidade.Candidatura{CPF: "12345678901", IDVaga: vagaID, Origem: &origem})

	require.NoError(t, err)
	assert.Empty(t, registrar.cpfs, "quem foi associado pelo recrutador não se candidatou")
}

func TestCandidaturaService_Create_FalhaNaOrigemNaoBloqueiaCandidatura(t *testing.T) {
	registrar := &fakeOrigemCadastro{err: errors.New("banco fora")}
	s, vagaID := candidaturaServiceComOrigem(t, registrar)

	id, err := s.Create(context.Background(), &empregabilidade.Candidatura{CPF: "12345678901", IDVaga: vagaID})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestCandidaturaService_Update_PreservaOrigem(t *testing.T) {
	repo := NewMockCandidaturaRepo()
	id := uuid.New()
	origem := empregabilidade.OrigemCandidaturaBancoCurriculos
	operador := "22222222222"
	repo.candidaturas[id] = &empregabilidade.Candidatura{
		ID:              id,
		CPF:             "12345678901",
		IDVaga:          uuid.New(),
		Origem:          &origem,
		AssociadoPorCPF: &operador,
	}
	s := services.NewCandidaturaService(repo, NewMockVagaRepo(), NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

	outraOrigem := empregabilidade.OrigemCandidaturaPortalCidadao
	err := s.Update(context.Background(), &empregabilidade.Candidatura{ID: id, Origem: &outraOrigem})

	require.NoError(t, err)
	resultado := repo.candidaturas[id]
	require.NotNil(t, resultado.Origem)
	assert.Equal(t, origem, *resultado.Origem)
	require.NotNil(t, resultado.AssociadoPorCPF)
	assert.Equal(t, operador, *resultado.AssociadoPorCPF)
}
