package empregabilidade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCandidaturasDoCPF struct {
	jaParticipa map[uuid.UUID]bool
	err         error
}

func (f *fakeCandidaturasDoCPF) VagasComCandidatura(_ context.Context, _ string, _ []uuid.UUID) (map[uuid.UUID]bool, error) {
	if f.jaParticipa == nil {
		return map[uuid.UUID]bool{}, f.err
	}
	return f.jaParticipa, f.err
}

type fakeTermos struct {
	aceitos bool
	err     error
}

func (f *fakeTermos) HasAcceptedTerms(_ context.Context, _ string) (bool, error) {
	return f.aceitos, f.err
}

type fakeVagas struct {
	vagas map[uuid.UUID]*empregabilidade.Vaga
	err   error
}

func (f *fakeVagas) GetByID(_ context.Context, id uuid.UUID) (*empregabilidade.Vaga, error) {
	if f.err != nil {
		return nil, f.err
	}
	vaga, ok := f.vagas[id]
	if !ok {
		return nil, nil
	}
	copia := *vaga
	return &copia, nil
}

type fakeCriador struct {
	criadas []*empregabilidade.Candidatura
	err     error
}

func (f *fakeCriador) CreateSemNotificacao(_ context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	if f.err != nil {
		return uuid.Nil, f.err
	}
	f.criadas = append(f.criadas, entity)
	return uuid.New(), nil
}

type fakeEscolaridades struct {
	itens []*empregabilidade.Escolaridade
}

func (f *fakeEscolaridades) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empregabilidade.Escolaridade, int, error) {
	return f.itens, len(f.itens), nil
}

type fakeIdiomas struct{}

func (f *fakeIdiomas) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empregabilidade.Idioma, int, error) {
	return nil, 0, nil
}

type fakeNiveis struct{}

func (f *fakeNiveis) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*empregabilidade.NivelIdioma, int, error) {
	return nil, 0, nil
}

const operadorCPF = "22222222222"

var idSituacaoEmpregado = uuid.New()

func curriculoMinimo() *empregabilidade.CurriculoCompleto {
	return &empregabilidade.CurriculoCompleto{
		SituacaoInteresses: &empregabilidade.CurriculoSituacaoInteresses{
			IDSituacao: &idSituacaoEmpregado,
			Situacao:   &empregabilidade.SituacaoAtual{Descricao: "Empregado(a)"},
		},
	}
}

func vagaAtiva(orgao string) *empregabilidade.Vaga {
	futuro := time.Now().Add(24 * time.Hour)
	return &empregabilidade.Vaga{
		Titulo:          "Auxiliar " + orgao,
		Status:          empregabilidade.StatusVagaPublicadoAtivo,
		DataLimite:      &futuro,
		IDOrgaoParceiro: orgao,
	}
}

type cenarioAssociacao struct {
	curriculo   *empregabilidade.CurriculoCompleto
	termos      *fakeTermos
	vagas       map[uuid.UUID]*empregabilidade.Vaga
	jaParticipa map[uuid.UUID]bool
	criador     *fakeCriador
	snapshot    *models.CitizenSnapshot
}

func (c *cenarioAssociacao) service() *services.AssociacaoCandidatoService {
	if c.curriculo == nil {
		c.curriculo = curriculoMinimo()
	}
	if c.termos == nil {
		c.termos = &fakeTermos{aceitos: true}
	}
	if c.criador == nil {
		c.criador = &fakeCriador{}
	}
	return services.NewAssociacaoCandidatoService(
		&fakeBancoRepo{curriculo: &empregabilidade.Curriculo{CPF: bancoCPF}},
		&fakeBancoCurriculoCompleto{curriculo: c.curriculo},
		&fakeCandidaturasDoCPF{jaParticipa: c.jaParticipa},
		c.termos,
		&fakeVagas{vagas: c.vagas},
		c.criador,
		&fakeEscolaridades{},
		&fakeIdiomas{},
		&fakeNiveis{},
		&fakeBancoSnapshotRepo{snapshot: c.snapshot},
	)
}

func TestAssociacaoCandidatoService_ResultadoPorVaga(t *testing.T) {
	idAtiva, idOutroOrgao, idExpirada, idJaParticipa, idInelegivel, idInexistente := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	passado := time.Now().Add(-time.Hour)
	expirada := vagaAtiva("2600")
	expirada.DataLimite = &passado
	inelegivel := vagaAtiva("2600")
	inelegivel.IdadeMinima = bancoInt(18)

	cenario := &cenarioAssociacao{
		vagas: map[uuid.UUID]*empregabilidade.Vaga{
			idAtiva:       vagaAtiva("2600"),
			idOutroOrgao:  vagaAtiva("9999"),
			idExpirada:    expirada,
			idJaParticipa: vagaAtiva("2600"),
			idInelegivel:  inelegivel,
		},
		jaParticipa: map[uuid.UUID]bool{idJaParticipa: true},
		snapshot:    &models.CitizenSnapshot{CPF: bancoCPF, Nome: "Ana Claudia", Email: "ana@example.com"},
	}

	resultados, err := cenario.service().Associar(context.Background(), services.AssociarParams{
		CPF:              bancoCPF,
		IDsVagas:         []uuid.UUID{idAtiva, idOutroOrgao, idExpirada, idJaParticipa, idInelegivel, idInexistente, idAtiva},
		OperadorCPF:      operadorCPF,
		OrgaosPermitidos: []string{"2600"},
	})

	require.NoError(t, err)
	require.Len(t, resultados, 6, "vaga repetida no pedido conta uma vez")

	esperado := []empregabilidade.ResultadoAssociacao{
		empregabilidade.AssociacaoRealizada,
		empregabilidade.AssociacaoForaDoEscopo,
		empregabilidade.AssociacaoVagaIndisponivel,
		empregabilidade.AssociacaoJaParticipa,
		empregabilidade.AssociacaoNaoElegivel,
		empregabilidade.AssociacaoVagaIndisponivel,
	}
	for i, r := range resultados {
		assert.Equal(t, esperado[i], r.Resultado, "vaga %d", i)
	}
	require.NotNil(t, resultados[0].IDCandidatura)
	assert.Equal(t, "Auxiliar 2600", resultados[0].Titulo)
	assert.Equal(t, []string{"Ter pelo menos 18 anos"}, resultados[4].Motivos, "sem data de nascimento não atende idade mínima")

	require.Len(t, cenario.criador.criadas, 1)
	criada := cenario.criador.criadas[0]
	assert.Equal(t, bancoCPF, criada.CPF)
	assert.Equal(t, idAtiva, criada.IDVaga)
	require.NotNil(t, criada.Origem)
	assert.Equal(t, empregabilidade.OrigemCandidaturaBancoCurriculos, *criada.Origem)
	require.NotNil(t, criada.AssociadoPorCPF)
	assert.Equal(t, operadorCPF, *criada.AssociadoPorCPF)
	require.NotNil(t, criada.Nome)
	assert.Equal(t, "Ana Claudia", *criada.Nome)
	require.NotNil(t, criada.Email)
	assert.Equal(t, "ana@example.com", *criada.Email)
}

func TestAssociacaoCandidatoService_AdminAssociaAQualquerOrgao(t *testing.T) {
	idVaga := uuid.New()
	cenario := &cenarioAssociacao{vagas: map[uuid.UUID]*empregabilidade.Vaga{idVaga: vagaAtiva("9999")}}

	resultados, err := cenario.service().Associar(context.Background(), services.AssociarParams{
		CPF: bancoCPF, IDsVagas: []uuid.UUID{idVaga}, OperadorCPF: operadorCPF,
	})

	require.NoError(t, err)
	require.Len(t, resultados, 1)
	assert.Equal(t, empregabilidade.AssociacaoRealizada, resultados[0].Resultado)
}

func TestAssociacaoCandidatoService_CurriculoForaDoMinimo(t *testing.T) {
	idDesempregado := uuid.New()
	cases := []struct {
		name       string
		curriculo  *empregabilidade.CurriculoCompleto
		aceitos    bool
		pendencias []string
	}{
		{
			name:       "sem situação e sem termos",
			curriculo:  &empregabilidade.CurriculoCompleto{},
			pendencias: []string{"Situação atual não informada", "Termos de uso não aceitos"},
		},
		{
			name: "desempregado sem tempo procurando emprego",
			curriculo: &empregabilidade.CurriculoCompleto{SituacaoInteresses: &empregabilidade.CurriculoSituacaoInteresses{
				IDSituacao: &idDesempregado,
				Situacao:   &empregabilidade.SituacaoAtual{Descricao: "Desempregado(a)"},
			}},
			aceitos:    true,
			pendencias: []string{"Tempo procurando emprego não informado"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cenario := &cenarioAssociacao{curriculo: tc.curriculo, termos: &fakeTermos{aceitos: tc.aceitos}}

			resultados, err := cenario.service().Associar(context.Background(), services.AssociarParams{
				CPF: bancoCPF, IDsVagas: []uuid.UUID{uuid.New()}, OperadorCPF: operadorCPF,
			})

			assert.Nil(t, resultados)
			var naoAssociavel *services.CandidatoNaoAssociavelError
			require.ErrorAs(t, err, &naoAssociavel)
			assert.ErrorIs(t, err, services.ErrCandidatoNaoAssociavel)
			assert.Equal(t, tc.pendencias, naoAssociavel.Pendencias)
			assert.Empty(t, cenario.criador.criadas)
		})
	}
}

func TestAssociacaoCandidatoService_SemCurriculo(t *testing.T) {
	svc := services.NewAssociacaoCandidatoService(
		&fakeBancoRepo{}, &fakeBancoCurriculoCompleto{}, &fakeCandidaturasDoCPF{}, &fakeTermos{aceitos: true},
		&fakeVagas{}, &fakeCriador{}, &fakeEscolaridades{}, &fakeIdiomas{}, &fakeNiveis{}, &fakeBancoSnapshotRepo{},
	)

	resultados, err := svc.Associar(context.Background(), services.AssociarParams{CPF: bancoCPF, IDsVagas: []uuid.UUID{uuid.New()}})

	require.NoError(t, err)
	assert.Nil(t, resultados)
}

func TestAssociacaoCandidatoService_FalhasAoCriar(t *testing.T) {
	idVaga := uuid.New()

	t.Run("candidatura criada por outra associação no meio do caminho", func(t *testing.T) {
		cenario := &cenarioAssociacao{
			vagas:   map[uuid.UUID]*empregabilidade.Vaga{idVaga: vagaAtiva("2600")},
			criador: &fakeCriador{err: errors.New("candidatura já existe para esta vaga")},
		}
		resultados, err := cenario.service().Associar(context.Background(), services.AssociarParams{CPF: bancoCPF, IDsVagas: []uuid.UUID{idVaga}})
		require.NoError(t, err)
		assert.Equal(t, empregabilidade.AssociacaoJaParticipa, resultados[0].Resultado)
	})

	t.Run("erro inesperado fica na vaga, sem derrubar as outras", func(t *testing.T) {
		cenario := &cenarioAssociacao{
			vagas:   map[uuid.UUID]*empregabilidade.Vaga{idVaga: vagaAtiva("2600")},
			criador: &fakeCriador{err: errors.New("pq: connection refused")},
		}
		resultados, err := cenario.service().Associar(context.Background(), services.AssociarParams{CPF: bancoCPF, IDsVagas: []uuid.UUID{idVaga}})
		require.NoError(t, err)
		assert.Equal(t, empregabilidade.AssociacaoErro, resultados[0].Resultado)
		assert.NotContains(t, resultados[0].Motivos[0], "pq:")
	})

	t.Run("erro ao verificar os termos derruba o pedido", func(t *testing.T) {
		cenario := &cenarioAssociacao{termos: &fakeTermos{err: errors.New("db")}}
		resultados, err := cenario.service().Associar(context.Background(), services.AssociarParams{CPF: bancoCPF, IDsVagas: []uuid.UUID{idVaga}})
		require.Error(t, err)
		assert.NotErrorIs(t, err, services.ErrCandidatoNaoAssociavel)
		assert.Nil(t, resultados)
	})
}
