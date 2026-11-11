package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// Origem dos eventos do histórico de candidaturas (DEV-532): o serviço diz ao
// repositório quem provocou cada transição.

type repoCandidaturaCapturaOrigem struct {
	*MockCandidaturaRepo
	origem    empregabilidade.OrigemEventoCandidatura
	temOrigem bool
}

func (m *repoCandidaturaCapturaOrigem) Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	m.origem, m.temOrigem = empregabilidade.OrigemEventoDoContexto(ctx)
	return m.MockCandidaturaRepo.Create(ctx, entity)
}

func TestCandidaturaService_Create_OrigemDoEvento(t *testing.T) {
	origem := func(o empregabilidade.OrigemCandidatura) *empregabilidade.OrigemCandidatura { return &o }

	tests := []struct {
		name     string
		origem   *empregabilidade.OrigemCandidatura
		semAviso bool
		ctx      context.Context
		esperado empregabilidade.AtorEventoCandidatura
	}{
		{"candidatura pelo app é do candidato", origem(empregabilidade.OrigemCandidaturaPortalCidadao), false, context.Background(), empregabilidade.AtorCandidato},
		{"cadastro manual pela secretaria é do operador", origem(empregabilidade.OrigemCandidaturaSecretaria), false, context.Background(), empregabilidade.AtorOperador},
		{"associação pelo banco de currículos é do operador", origem(empregabilidade.OrigemCandidaturaBancoCurriculos), true, context.Background(), empregabilidade.AtorOperador},
		{"origem já declarada no contexto é mantida", origem(empregabilidade.OrigemCandidaturaPortalCidadao), false,
			empregabilidade.ComOrigemEvento(context.Background(), empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorAutomacao}), empregabilidade.AtorAutomacao},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repoCandidaturaCapturaOrigem{MockCandidaturaRepo: NewMockCandidaturaRepo()}
			vagaRepo := NewMockVagaRepo()
			vagaID := uuid.New()
			vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{ID: vagaID, Status: empregabilidade.StatusVagaPublicadoAtivo}
			svc := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

			candidatura := &empregabilidade.Candidatura{CPF: "12345678901", IDVaga: vagaID, Origem: tt.origem}
			var err error
			if tt.semAviso {
				_, err = svc.CreateSemNotificacao(tt.ctx, candidatura)
			} else {
				_, err = svc.Create(tt.ctx, candidatura)
			}
			require.NoError(t, err)
			require.True(t, repo.temOrigem, "o repositório precisa receber a origem do evento")
			assert.Equal(t, tt.esperado, repo.origem.Ator)
		})
	}
}

type repoVagaCandidaturasCapturaOrigem struct {
	origens []empregabilidade.OrigemEventoCandidatura
}

func (m *repoVagaCandidaturasCapturaOrigem) BulkSaveAndUpdateStatusByVagaID(ctx context.Context, _ uuid.UUID, _ empregabilidade.StatusCandidatura) error {
	origem, _ := empregabilidade.OrigemEventoDoContexto(ctx)
	m.origens = append(m.origens, origem)
	return nil
}

func (m *repoVagaCandidaturasCapturaOrigem) BulkRestoreStatusByVagaID(ctx context.Context, _ uuid.UUID) error {
	origem, _ := empregabilidade.OrigemEventoDoContexto(ctx)
	m.origens = append(m.origens, origem)
	return nil
}

func TestVagaService_MudancaDeEstado_OrigemDoEvento(t *testing.T) {
	tests := []struct {
		name       string
		statusVaga empregabilidade.StatusVaga
		acao       func(*services.VagaService, context.Context, uuid.UUID) error
		motivo     string
	}{
		{"congelar", empregabilidade.StatusVagaPublicadoAtivo, (*services.VagaService).FreezeVaga, empregabilidade.MotivoVagaCongelada},
		{"descongelar", empregabilidade.StatusVagaCongelada, (*services.VagaService).UnfreezeVaga, empregabilidade.MotivoVagaDescongelada},
		{"descontinuar", empregabilidade.StatusVagaPublicadoAtivo, (*services.VagaService).DiscontinueVaga, empregabilidade.MotivoVagaDescontinuada},
		{"reativar", empregabilidade.StatusVagaDescontinuada, (*services.VagaService).ReactivateVaga, empregabilidade.MotivoVagaReativada},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vagaRepo := NewMockVagaRepoForService()
			candRepo := &repoVagaCandidaturasCapturaOrigem{}
			svc := services.NewVagaServiceWithInterfaces(vagaRepo, NewMockEmpresaRepoForService(), candRepo)
			vagaID := uuid.New()
			vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{ID: vagaID, Status: tt.statusVaga}

			require.NoError(t, tt.acao(svc, context.Background(), vagaID))
			require.Len(t, candRepo.origens, 1)
			assert.Equal(t, empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorAutomacao, Motivo: tt.motivo}, candRepo.origens[0])
		})
	}
}
