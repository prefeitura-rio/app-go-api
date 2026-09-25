package empregabilidade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBancoRepo struct {
	items             []*empregabilidade.BancoCurriculoItem
	total             int64
	listErr           error
	gotFilter         empregabilidade.BancoCurriculoFilter
	gotPage           int
	gotSize           int
	curriculo         *empregabilidade.Curriculo
	getErr            error
	atualizacao       *time.Time
	atualizacaoErr    error
	gotCPFAtualizacao string
}

func (f *fakeBancoRepo) ListBancoCurriculos(_ context.Context, filter empregabilidade.BancoCurriculoFilter, page, pageSize int) ([]*empregabilidade.BancoCurriculoItem, int64, error) {
	f.gotFilter, f.gotPage, f.gotSize = filter, page, pageSize
	return f.items, f.total, f.listErr
}

func (f *fakeBancoRepo) GetCurriculoByCPF(_ context.Context, _ string) (*empregabilidade.Curriculo, error) {
	return f.curriculo, f.getErr
}

func (f *fakeBancoRepo) GetUltimaAtualizacao(_ context.Context, cpf string) (*time.Time, error) {
	f.gotCPFAtualizacao = cpf
	return f.atualizacao, f.atualizacaoErr
}

type fakeBancoCurriculoCompleto struct {
	curriculo *empregabilidade.CurriculoCompleto
	err       error
	called    bool
}

func (f *fakeBancoCurriculoCompleto) GetCurriculoCompleto(_ context.Context, _ string) (*empregabilidade.CurriculoCompleto, error) {
	f.called = true
	return f.curriculo, f.err
}

type fakeBancoSnapshotRepo struct {
	snapshot *models.CitizenSnapshot
	err      error
}

func (f *fakeBancoSnapshotRepo) GetByCPF(_ context.Context, _ string) (*models.CitizenSnapshot, error) {
	return f.snapshot, f.err
}

func (f *fakeBancoSnapshotRepo) GetByCPFs(_ context.Context, _ []string) (map[string]*models.CitizenSnapshot, error) {
	return nil, nil
}

type fakeBancoFetcher struct {
	snapshot    *models.CitizenSnapshot
	err         error
	calls       int
	forcedCalls int
}

func (f *fakeBancoFetcher) SyncCitizenOnDemand(_ context.Context, _ string) (*models.CitizenSnapshot, error) {
	f.calls++
	return f.snapshot, f.err
}

// SyncCitizenForced existe para satisfazer a interface: a ficha do banco de
// currículos usa só o sync sob demanda, então contamos as chamadas para provar
// que ela não força atualização do cadastro.
func (f *fakeBancoFetcher) SyncCitizenForced(_ context.Context, _ string) (*models.CitizenSnapshot, error) {
	f.forcedCalls++
	return f.snapshot, f.err
}

func (f *fakeBancoFetcher) StaleThreshold() time.Duration { return 24 * time.Hour }

func bancoInt(v int) *int { return &v }

const bancoCPF = "11111111111"

var bancoInclusao = time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

func newBancoService(snapshot *models.CitizenSnapshot, curriculo *empregabilidade.CurriculoCompleto) (*services.BancoCurriculoService, *fakeBancoCurriculoCompleto) {
	completo := &fakeBancoCurriculoCompleto{curriculo: curriculo}
	svc := services.NewBancoCurriculoService(
		&fakeBancoRepo{curriculo: &empregabilidade.Curriculo{CPF: bancoCPF, CreatedAt: bancoInclusao}},
		completo,
		&fakeBancoSnapshotRepo{snapshot: snapshot},
	)
	return svc, completo
}

// newBancoServiceComRepo monta o service deixando o repositório do banco de
// currículos à mão, para os testes que mexem na data de atualização.
func newBancoServiceComRepo(repo *fakeBancoRepo, snapshot *models.CitizenSnapshot) *services.BancoCurriculoService {
	repo.curriculo = &empregabilidade.Curriculo{CPF: bancoCPF, CreatedAt: bancoInclusao}
	return services.NewBancoCurriculoService(
		repo,
		&fakeBancoCurriculoCompleto{curriculo: &empregabilidade.CurriculoCompleto{}},
		&fakeBancoSnapshotRepo{snapshot: snapshot},
	)
}

func TestBancoCurriculoService_List_RepassaFiltroEPaginacao(t *testing.T) {
	nome := "Ana"
	repo := &fakeBancoRepo{
		items: []*empregabilidade.BancoCurriculoItem{{CPF: bancoCPF, Nome: &nome}},
		total: 31,
	}
	svc := services.NewBancoCurriculoService(repo, &fakeBancoCurriculoCompleto{}, &fakeBancoSnapshotRepo{})

	items, total, err := svc.List(context.Background(), empregabilidade.BancoCurriculoFilter{Search: "ana"}, 3, 10)

	require.NoError(t, err)
	assert.Equal(t, int64(31), total)
	assert.Len(t, items, 1)
	assert.Equal(t, "ana", repo.gotFilter.Search)
	assert.Equal(t, 3, repo.gotPage)
	assert.Equal(t, 10, repo.gotSize)
}

func TestBancoCurriculoService_GetDetalhe_NaoEncontrado(t *testing.T) {
	completo := &fakeBancoCurriculoCompleto{}
	svc := services.NewBancoCurriculoService(&fakeBancoRepo{}, completo, &fakeBancoSnapshotRepo{})

	detalhe, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Nil(t, detalhe)
	assert.False(t, completo.called, "não deve carregar o currículo de quem não está no banco")
}

func TestBancoCurriculoService_GetDetalhe_MontaFicha(t *testing.T) {
	nascimento := time.Now().AddDate(-32, 0, -1) // fez 32 anos ontem
	snapshot := &models.CitizenSnapshot{
		CPF:            bancoCPF,
		Nome:           "Ana Claudia Silva",
		Celular:        "21982780000",
		DataNascimento: &nascimento,
		Endereco:       &models.CitizenEndereco{Bairro: "Andaraí"},
		Genero:         "Mulher cisgênero",
		Escolaridade:   "Médio completo",
		LastSyncedAt:   time.Now(),
	}
	curriculo := &empregabilidade.CurriculoCompleto{
		Experiencias: []*empregabilidade.CurriculoExperiencia{
			{Cargo: "Auxiliar administrativa", TempoExperienciaMeses: bancoInt(60)},
			{Cargo: "Redatora Sênior", EhTrabalhoAtual: true, TempoExperienciaMeses: bancoInt(36)},
		},
	}
	svc, _ := newBancoService(snapshot, curriculo)

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, bancoCPF, d.CPF)
	assert.True(t, bancoInclusao.Equal(d.DataInclusao))
	assert.Same(t, curriculo, d.Curriculo)
	require.NotNil(t, d.Nome)
	assert.Equal(t, "Ana Claudia Silva", *d.Nome)
	assert.Nil(t, d.NomeSocial)
	require.NotNil(t, d.Profissao)
	assert.Equal(t, "Redatora Sênior", *d.Profissao, "emprego atual tem prioridade")
	require.NotNil(t, d.Escolaridade)
	assert.Equal(t, "Médio completo", *d.Escolaridade)
	require.NotNil(t, d.Bairro)
	assert.Equal(t, "Andaraí", *d.Bairro)
	require.NotNil(t, d.Celular)
	assert.Equal(t, "21982780000", *d.Celular)
	require.NotNil(t, d.Genero)
	assert.Equal(t, "Mulher cisgênero", *d.Genero)
	require.NotNil(t, d.Idade)
	assert.Equal(t, 32, *d.Idade)
}

func TestBancoCurriculoService_GetDetalhe_Idade(t *testing.T) {
	aniversarioAmanha := time.Now().AddDate(-32, 0, 1)
	svc, _ := newBancoService(&models.CitizenSnapshot{DataNascimento: &aniversarioAmanha, LastSyncedAt: time.Now()}, &empregabilidade.CurriculoCompleto{})

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	require.NotNil(t, d.Idade)
	assert.Equal(t, 31, *d.Idade, "ainda não fez aniversário este ano")
}

func TestBancoCurriculoService_GetDetalhe_ProfissaoSemEmpregoAtualUsaMaiorExperiencia(t *testing.T) {
	curriculo := &empregabilidade.CurriculoCompleto{
		Experiencias: []*empregabilidade.CurriculoExperiencia{
			{Cargo: "Auxiliar", TempoExperienciaMeses: bancoInt(12)},
			{Cargo: "Estagiária"},
			{Cargo: "Contadora", TempoExperienciaMeses: bancoInt(60)},
		},
	}
	svc, _ := newBancoService(nil, curriculo)

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	require.NotNil(t, d.Profissao)
	assert.Equal(t, "Contadora", *d.Profissao)
}

func TestBancoCurriculoService_GetDetalhe_SemExperienciaProfissaoNula(t *testing.T) {
	svc, _ := newBancoService(nil, &empregabilidade.CurriculoCompleto{})

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Nil(t, d.Profissao)
}

func TestBancoCurriculoService_GetDetalhe_DadosVaziosViramNulos(t *testing.T) {
	snapshot := &models.CitizenSnapshot{
		CPF:          bancoCPF,
		Nome:         "Ana",
		Escolaridade: "  ",
		Email:        "   ",
		Raca:         "",
		Deficiencia:  " ",
		LastSyncedAt: time.Now(),
	}
	svc, _ := newBancoService(snapshot, &empregabilidade.CurriculoCompleto{})

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Nil(t, d.NomeSocial)
	assert.Nil(t, d.Escolaridade)
	assert.Nil(t, d.Bairro)
	assert.Nil(t, d.Celular)
	assert.Nil(t, d.Genero)
	assert.Nil(t, d.Idade)
	assert.Nil(t, d.Email)
	assert.Nil(t, d.Raca)
	assert.Nil(t, d.Deficiencia)
}

// Email, raça e deficiência saem do cadastro do cidadão, como os demais dados
// pessoais da ficha.
func TestBancoCurriculoService_GetDetalhe_EmailRacaEDeficienciaDoCadastro(t *testing.T) {
	snapshot := &models.CitizenSnapshot{
		CPF:          bancoCPF,
		Nome:         "Ana Claudia Silva",
		Email:        " ana@exemplo.com ",
		Raca:         "Parda",
		Deficiencia:  "Deficiência auditiva",
		LastSyncedAt: time.Now(),
	}
	svc, _ := newBancoService(snapshot, &empregabilidade.CurriculoCompleto{})

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	require.NotNil(t, d.Email)
	assert.Equal(t, "ana@exemplo.com", *d.Email, "espaços das pontas são descartados")
	require.NotNil(t, d.Raca)
	assert.Equal(t, "Parda", *d.Raca)
	require.NotNil(t, d.Deficiencia)
	assert.Equal(t, "Deficiência auditiva", *d.Deficiencia)
}

func TestBancoCurriculoService_GetDetalhe_DataAtualizacao(t *testing.T) {
	t.Run("repassa a data do repositório", func(t *testing.T) {
		atualizacao := time.Date(2026, 9, 18, 14, 30, 0, 0, time.UTC)
		repo := &fakeBancoRepo{atualizacao: &atualizacao}
		svc := newBancoServiceComRepo(repo, &models.CitizenSnapshot{LastSyncedAt: time.Now()})

		d, err := svc.GetDetalhe(context.Background(), bancoCPF)

		require.NoError(t, err)
		require.NotNil(t, d.DataAtualizacao)
		assert.True(t, atualizacao.Equal(*d.DataAtualizacao))
		assert.Equal(t, bancoCPF, repo.gotCPFAtualizacao)
	})

	t.Run("nula quando o currículo não tem nenhuma seção", func(t *testing.T) {
		repo := &fakeBancoRepo{}
		svc := newBancoServiceComRepo(repo, &models.CitizenSnapshot{LastSyncedAt: time.Now()})

		d, err := svc.GetDetalhe(context.Background(), bancoCPF)

		require.NoError(t, err)
		assert.Nil(t, d.DataAtualizacao)
	})
}

func TestBancoCurriculoService_GetDetalhe_GeneroSoAutodeclarado(t *testing.T) {
	cases := []struct {
		name  string
		valor string
		want  *string
	}{
		{"sexo cadastral M não vira identidade", "M", nil},
		{"sexo cadastral F não vira identidade", "F", nil},
		{"sexo cadastral por extenso não vira identidade", "Feminino", nil},
		{"opção autodeclarada", "Mulher transgênero", strPtrBanco("Mulher transgênero")},
		{"normaliza caixa e espaços", "  mulher cisgênero ", strPtrBanco("Mulher cisgênero")},
		{"prefiro não informar é resposta válida", "Prefiro não informar", strPtrBanco("Prefiro não informar")},
		{"outro", "Outro", strPtrBanco("Outro")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newBancoService(&models.CitizenSnapshot{Genero: tc.valor, LastSyncedAt: time.Now()}, &empregabilidade.CurriculoCompleto{})

			d, err := svc.GetDetalhe(context.Background(), bancoCPF)

			require.NoError(t, err)
			assert.Equal(t, tc.want, d.Genero)
		})
	}
}

func strPtrBanco(s string) *string { return &s }

func TestBancoCurriculoService_GetDetalhe_SemCadastroSincronizaSobDemanda(t *testing.T) {
	svc, _ := newBancoService(nil, &empregabilidade.CurriculoCompleto{})
	fetcher := &fakeBancoFetcher{snapshot: &models.CitizenSnapshot{Nome: "Ana Sincronizada"}}
	svc.SetCitizenDataFetcher(fetcher)

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Equal(t, 1, fetcher.calls)
	assert.Equal(t, 0, fetcher.forcedCalls, "a ficha usa o sync sob demanda, não o forçado")
	require.NotNil(t, d.Nome)
	assert.Equal(t, "Ana Sincronizada", *d.Nome)
}

func TestBancoCurriculoService_GetDetalhe_CadastroDesatualizadoMantemAntigoSeSincronizacaoFalha(t *testing.T) {
	antigo := &models.CitizenSnapshot{Nome: "Ana Antiga", LastSyncedAt: time.Now().Add(-48 * time.Hour)}
	svc, _ := newBancoService(antigo, &empregabilidade.CurriculoCompleto{})
	fetcher := &fakeBancoFetcher{err: errors.New("rmi fora do ar")}
	svc.SetCitizenDataFetcher(fetcher)

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Equal(t, 1, fetcher.calls)
	require.NotNil(t, d.Nome)
	assert.Equal(t, "Ana Antiga", *d.Nome)
}

func TestBancoCurriculoService_GetDetalhe_CadastroRecenteNaoSincroniza(t *testing.T) {
	svc, _ := newBancoService(&models.CitizenSnapshot{Nome: "Ana", LastSyncedAt: time.Now()}, &empregabilidade.CurriculoCompleto{})
	fetcher := &fakeBancoFetcher{}
	svc.SetCitizenDataFetcher(fetcher)

	_, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Equal(t, 0, fetcher.calls)
}

func TestBancoCurriculoService_GetDetalhe_ErroNoCadastroNaoDerrubaFicha(t *testing.T) {
	svc := services.NewBancoCurriculoService(
		&fakeBancoRepo{curriculo: &empregabilidade.Curriculo{CPF: bancoCPF, CreatedAt: bancoInclusao}},
		&fakeBancoCurriculoCompleto{curriculo: &empregabilidade.CurriculoCompleto{}},
		&fakeBancoSnapshotRepo{err: errors.New("db")},
	)

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Nil(t, d.Nome)
}

func TestBancoCurriculoService_GetDetalhe_ProfissaoEmpateUsaOrdemAlfabetica(t *testing.T) {
	curriculo := &empregabilidade.CurriculoCompleto{
		Experiencias: []*empregabilidade.CurriculoExperiencia{
			{Cargo: "Contadora", TempoExperienciaMeses: bancoInt(60)},
			{Cargo: "Administradora", TempoExperienciaMeses: bancoInt(60)},
		},
	}
	svc, _ := newBancoService(nil, curriculo)

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	require.NotNil(t, d.Profissao)
	assert.Equal(t, "Administradora", *d.Profissao, "mesmo desempate da listagem (cargo ASC)")
}

func TestBancoCurriculoService_GetDetalhe_NascimentoNoFuturoSemIdade(t *testing.T) {
	nascimento := time.Now().AddDate(1, 0, 0)
	svc, _ := newBancoService(&models.CitizenSnapshot{DataNascimento: &nascimento, LastSyncedAt: time.Now()}, &empregabilidade.CurriculoCompleto{})

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Nil(t, d.Idade)
}

func TestBancoCurriculoService_GetDetalhe_SincronizacaoSemResultadoMantemCadastro(t *testing.T) {
	antigo := &models.CitizenSnapshot{Nome: "Ana Antiga", LastSyncedAt: time.Now().Add(-48 * time.Hour)}
	svc, _ := newBancoService(antigo, &empregabilidade.CurriculoCompleto{})
	fetcher := &fakeBancoFetcher{}
	svc.SetCitizenDataFetcher(fetcher)

	d, err := svc.GetDetalhe(context.Background(), bancoCPF)

	require.NoError(t, err)
	assert.Equal(t, 1, fetcher.calls)
	require.NotNil(t, d.Nome)
	assert.Equal(t, "Ana Antiga", *d.Nome)
}

func TestBancoCurriculoService_GetDetalhe_Erros(t *testing.T) {
	t.Run("ao buscar o currículo raiz", func(t *testing.T) {
		svc := services.NewBancoCurriculoService(&fakeBancoRepo{getErr: errors.New("db")}, &fakeBancoCurriculoCompleto{}, &fakeBancoSnapshotRepo{})

		d, err := svc.GetDetalhe(context.Background(), bancoCPF)

		require.Error(t, err)
		assert.Nil(t, d)
	})

	t.Run("ao carregar o currículo completo", func(t *testing.T) {
		svc := services.NewBancoCurriculoService(
			&fakeBancoRepo{curriculo: &empregabilidade.Curriculo{CPF: bancoCPF}},
			&fakeBancoCurriculoCompleto{err: errors.New("db")},
			&fakeBancoSnapshotRepo{},
		)

		d, err := svc.GetDetalhe(context.Background(), bancoCPF)

		require.Error(t, err)
		assert.Nil(t, d)
	})

	t.Run("ao buscar a última atualização", func(t *testing.T) {
		repo := &fakeBancoRepo{atualizacaoErr: errors.New("db")}
		svc := newBancoServiceComRepo(repo, nil)

		d, err := svc.GetDetalhe(context.Background(), bancoCPF)

		require.Error(t, err)
		assert.Nil(t, d)
	})
}
