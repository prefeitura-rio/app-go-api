package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// novaVagaComEtapa registra no mock uma vaga ativa com uma etapa própria.
func novaVagaComEtapa(vagaRepo *MockVagaRepo) (vagaID, etapaID uuid.UUID) {
	vagaID = uuid.New()
	etapaID = uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
		ID:     vagaID,
		Status: empregabilidade.StatusVagaPublicadoAtivo,
		Etapas: []empregabilidade.Etapa{{ID: etapaID, IDVaga: vagaID, Titulo: "Análise de currículo", Ordem: 1}},
	}
	return vagaID, etapaID
}

// A etapa de uma candidatura nova, quando informada, precisa ser da mesma
// vaga: etapa de outra vaga é recusada com 400 e nada é gravado (DEV-767).
func TestCandidaturaService_Create_EtapaDeOutraVaga_Recusa(t *testing.T) {
	for _, tc := range []struct {
		name  string
		criar func(*services.CandidaturaService, *empregabilidade.Candidatura) (uuid.UUID, error)
	}{
		{"Create", func(s *services.CandidaturaService, c *empregabilidade.Candidatura) (uuid.UUID, error) {
			return s.Create(context.Background(), c)
		}},
		{"CreateSemNotificacao", func(s *services.CandidaturaService, c *empregabilidade.Candidatura) (uuid.UUID, error) {
			return s.CreateSemNotificacao(context.Background(), c)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockCandidaturaRepo()
			vagaRepo := NewMockVagaRepo()
			vagaID, _ := novaVagaComEtapa(vagaRepo)
			_, etapaOutraVaga := novaVagaComEtapa(vagaRepo)
			service := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

			id, err := tc.criar(service, &empregabilidade.Candidatura{
				CPF:          "12345678900",
				IDVaga:       vagaID,
				IDEtapaAtual: &etapaOutraVaga,
			})

			require.Error(t, err)
			assert.True(t, apperrors.IsKind(err, apperrors.KindValidation), "esperado erro de validação, veio %v", err)
			assert.Equal(t, uuid.Nil, id)
			assert.Empty(t, repo.candidaturas, "nenhuma candidatura deveria ser gravada")
		})
	}
}

// Etapa inexistente (nenhuma vaga a tem) também é recusada.
func TestCandidaturaService_Create_EtapaInexistente_Recusa(t *testing.T) {
	repo := NewMockCandidaturaRepo()
	vagaRepo := NewMockVagaRepo()
	vagaID, _ := novaVagaComEtapa(vagaRepo)
	service := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

	etapa := uuid.New()
	_, err := service.Create(context.Background(), &empregabilidade.Candidatura{
		CPF:          "12345678900",
		IDVaga:       vagaID,
		IDEtapaAtual: &etapa,
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsKind(err, apperrors.KindValidation), "esperado erro de validação, veio %v", err)
	assert.Empty(t, repo.candidaturas)
}

// Etapa da própria vaga continua aceita (cadastro pela secretaria).
func TestCandidaturaService_Create_EtapaDaVaga_Mantem(t *testing.T) {
	repo := NewMockCandidaturaRepo()
	vagaRepo := NewMockVagaRepo()
	vagaID, etapaID := novaVagaComEtapa(vagaRepo)
	service := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

	etapa := etapaID
	id, err := service.Create(context.Background(), &empregabilidade.Candidatura{
		CPF:          "12345678900",
		IDVaga:       vagaID,
		IDEtapaAtual: &etapa,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.candidaturas[id].IDEtapaAtual)
	assert.Equal(t, etapaID, *repo.candidaturas[id].IDEtapaAtual)
}

// Sem etapa, a candidatura nasce sem etapa, como antes.
func TestCandidaturaService_Create_SemEtapa_Aceita(t *testing.T) {
	repo := NewMockCandidaturaRepo()
	vagaRepo := NewMockVagaRepo()
	vagaID, _ := novaVagaComEtapa(vagaRepo)
	service := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

	id, err := service.Create(context.Background(), &empregabilidade.Candidatura{CPF: "12345678900", IDVaga: vagaID})

	require.NoError(t, err)
	assert.Nil(t, repo.candidaturas[id].IDEtapaAtual)
}

// O status de uma candidatura nova é sempre candidatura_enviada: não há status
// anterior, então o que vier do corpo é descartado (senão o descongelamento da
// vaga o restauraria).
func TestCandidaturaService_Create_DescartaStatusAnterior(t *testing.T) {
	repo := NewMockCandidaturaRepo()
	vagaRepo := NewMockVagaRepo()
	vagaID, _ := novaVagaComEtapa(vagaRepo)
	service := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

	aprovada := empregabilidade.StatusCandidaturaAprovada
	id, err := service.Create(context.Background(), &empregabilidade.Candidatura{
		CPF:            "12345678900",
		IDVaga:         vagaID,
		StatusAnterior: &aprovada,
	})

	require.NoError(t, err)
	assert.Nil(t, repo.candidaturas[id].StatusAnterior)
	assert.Equal(t, empregabilidade.StatusCandidaturaEnviada, repo.candidaturas[id].Status)
}
