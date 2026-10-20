package empregabilidade_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	emprepo "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// Pessoas fictícias do teste dos filtros avançados.
const (
	cpfAna    = "90000000001" // mulher cis, 30 anos, Andaraí, médio completo, PCD visual, 3 anos (2 em carteira), inglês avançado
	cpfBruno  = "90000000002" // homem cis, 45 anos, Madureira, superior completo, sem deficiência, 11 anos sem carteira, inglês básico
	cpfCarla  = "90000000003" // mulher trans, 19 anos, andarai (sem acento), médio completo, reabilitado INSS, sem experiência
	cpfDiego  = "90000000004" // sem cadastro sincronizado, 4 meses em carteira
	cpfSemPCD = "90000000005" // não binário, deficiência não informada, 5 meses de experiência
)

func nascidoHa(anos int) *time.Time {
	d := time.Now().AddDate(-anos, 0, -1)
	return &d
}

func meses(v int) *int { return &v }

type idsReferencia struct {
	ingles, espanhol           uuid.UUID
	basico, intermed, avancado uuid.UUID
}

func referenciasIdioma(t *testing.T, tx *gorm.DB) idsReferencia {
	t.Helper()
	var ids idsReferencia
	buscar := func(tabela, descricao string) uuid.UUID {
		var texto string
		require.NoError(t, tx.Raw("SELECT id::text FROM "+tabela+" WHERE descricao = ?", descricao).Row().Scan(&texto),
			"%s %s não encontrado no seed", tabela, descricao)
		return uuid.MustParse(texto)
	}
	ids.ingles = buscar("emp_idiomas", "Inglês")
	ids.espanhol = buscar("emp_idiomas", "Espanhol")
	ids.basico = buscar("emp_niveis_idioma", "Básico")
	ids.intermed = buscar("emp_niveis_idioma", "Intermediário")
	ids.avancado = buscar("emp_niveis_idioma", "Avançado")
	return ids
}

func semearCandidatosFiltro(t *testing.T, tx *gorm.DB, ids idsReferencia) {
	t.Helper()
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)

	require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfAna, []*empregabilidade.CurriculoExperiencia{
		{Cargo: "Auxiliar", Empresa: "Empresa A", TempoExperienciaMeses: meses(24), ExperienciaComprovadaCT: true},
		{Cargo: "Atendente", Empresa: "Empresa B", TempoExperienciaMeses: meses(12)},
	}))
	require.NoError(t, repo.ReplaceAllIdiomasByCPF(ctx, cpfAna, []*empregabilidade.CurriculoIdioma{{IDIdioma: ids.ingles, IDNivel: ids.avancado}}))

	require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfBruno, []*empregabilidade.CurriculoExperiencia{
		{Cargo: "Motorista", Empresa: "Empresa C", TempoExperienciaMeses: meses(132)},
	}))
	require.NoError(t, repo.ReplaceAllIdiomasByCPF(ctx, cpfBruno, []*empregabilidade.CurriculoIdioma{{IDIdioma: ids.ingles, IDNivel: ids.basico}}))

	require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{CPF: cpfCarla}))

	require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfDiego, []*empregabilidade.CurriculoExperiencia{
		{Cargo: "Estoquista", Empresa: "Empresa D", TempoExperienciaMeses: meses(4), ExperienciaComprovadaCT: true},
	}))
	require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfSemPCD, []*empregabilidade.CurriculoExperiencia{
		{Cargo: "Vendedor", Empresa: "Empresa E", TempoExperienciaMeses: meses(5)},
	}))

	for _, s := range []*models.CitizenSnapshot{
		{CPF: cpfAna, Nome: "Ana Filtro", Genero: "Mulher cisgênero", DataNascimento: nascidoHa(30), Escolaridade: "Médio completo",
			Deficiencia: "Visual", Endereco: &models.CitizenEndereco{Bairro: "Andaraí"}, LastSyncedAt: time.Now()},
		{CPF: cpfBruno, Nome: "Bruno Filtro", Genero: "Homem cisgênero", DataNascimento: nascidoHa(45), Escolaridade: "Superior Completo",
			Deficiencia: "Não sou pessoa com deficiência", Endereco: &models.CitizenEndereco{Bairro: "Madureira"}, LastSyncedAt: time.Now()},
		{CPF: cpfCarla, Nome: "Carla Filtro", Genero: "Mulher transgênero", DataNascimento: nascidoHa(19), Escolaridade: "médio completo",
			Deficiencia: "Reabilitado do INSS", Endereco: &models.CitizenEndereco{Bairro: "andarai"}, LastSyncedAt: time.Now()},
		{CPF: cpfSemPCD, Nome: "Sam Filtro", Genero: "Não binário", LastSyncedAt: time.Now()},
	} {
		require.NoError(t, tx.Create(s).Error)
	}
}

func cpfsDoBanco(t *testing.T, repo *emprepo.CurriculoRepository, f empregabilidade.FiltroCandidato) []string {
	t.Helper()
	items, total, err := repo.ListBancoCurriculos(context.Background(), empregabilidade.BancoCurriculoFilter{Search: "filtro", Candidato: f}, 1, 100)
	require.NoError(t, err)
	cpfs := []string{}
	for _, item := range items {
		cpfs = append(cpfs, item.CPF)
	}
	sort.Strings(cpfs)
	assert.Equal(t, int64(len(cpfs)), total, "o total acompanha o filtro")
	return cpfs
}

func TestFiltroCandidatoIntegration_BancoDeCurriculos(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ids := referenciasIdioma(t, tx)
	semearCandidatosFiltro(t, tx, ids)
	repo := emprepo.NewCurriculoRepository(tx)

	sim, nao := true, false
	cases := []struct {
		name   string
		filtro empregabilidade.FiltroCandidato
		cpfs   []string
	}{
		// A busca "filtro" pega quem tem cadastro; Diego (sem cadastro) só entra sem busca.
		{"sem filtro", empregabilidade.FiltroCandidato{}, []string{cpfAna, cpfBruno, cpfCarla, cpfSemPCD}},
		{"gêneros selecionados, sem diferenciar maiúsculas", empregabilidade.FiltroCandidato{Generos: []string{"mulher cisgênero", "Mulher transgênero"}}, []string{cpfAna, cpfCarla}},
		{"idade mínima", empregabilidade.FiltroCandidato{IdadeMinima: meses(30)}, []string{cpfAna, cpfBruno}},
		{"faixa de idade", empregabilidade.FiltroCandidato{IdadeMinima: meses(18), IdadeMaxima: meses(30)}, []string{cpfAna, cpfCarla}},
		{"sem experiência prévia", empregabilidade.FiltroCandidato{TempoExperiencia: empregabilidade.FaixaSemExperiencia}, []string{cpfCarla}},
		{"até 2 anos", empregabilidade.FiltroCandidato{TempoExperiencia: empregabilidade.FaixaAte2Anos}, []string{cpfSemPCD}},
		{"2 a 4 anos soma as experiências", empregabilidade.FiltroCandidato{TempoExperiencia: empregabilidade.Faixa2a4Anos}, []string{cpfAna}},
		{"mais de 10 anos", empregabilidade.FiltroCandidato{TempoExperiencia: empregabilidade.FaixaMaisDe10Anos}, []string{cpfBruno}},
		{"carteira exigida", empregabilidade.FiltroCandidato{Carteira: empregabilidade.CarteiraExigida}, []string{cpfAna}},
		{"6 meses em carteira", empregabilidade.FiltroCandidato{Carteira: empregabilidade.CarteiraSeisMeses}, []string{cpfAna}},
		{"escolaridade exata, sem diferenciar maiúsculas", empregabilidade.FiltroCandidato{Escolaridade: "Médio completo"}, []string{cpfAna, cpfCarla}},
		{"escolaridade exata não pega quem estudou mais", empregabilidade.FiltroCandidato{Escolaridade: "Superior completo"}, []string{cpfBruno}},
		{"idioma em qualquer nível", empregabilidade.FiltroCandidato{IDIdioma: &ids.ingles}, []string{cpfAna, cpfBruno}},
		{"idioma no nível ou acima", empregabilidade.FiltroCandidato{IDIdioma: &ids.ingles, IDNivelIdiomaMinimo: &ids.intermed}, []string{cpfAna}},
		{"idioma que ninguém tem", empregabilidade.FiltroCandidato{IDIdioma: &ids.espanhol}, []string{}},
		{"bairro sem acento e sem diferenciar maiúsculas", empregabilidade.FiltroCandidato{Bairro: "ANDARAÍ"}, []string{cpfAna, cpfCarla}},
		{"PCD sim inclui reabilitado do INSS", empregabilidade.FiltroCandidato{PCD: &sim}, []string{cpfAna, cpfCarla}},
		{"PCD não exclui quem não informou", empregabilidade.FiltroCandidato{PCD: &nao}, []string{cpfBruno}},
		{"filtros combinados", empregabilidade.FiltroCandidato{Generos: []string{"Mulher cisgênero", "Mulher transgênero"}, PCD: &sim, Carteira: empregabilidade.CarteiraExigida}, []string{cpfAna}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.cpfs, cpfsDoBanco(t, repo, tc.filtro))
		})
	}

	t.Run("filtro do currículo vale para quem não tem cadastro", func(t *testing.T) {
		items, _, err := repo.ListBancoCurriculos(context.Background(), empregabilidade.BancoCurriculoFilter{
			Search:    cpfDiego,
			Candidato: empregabilidade.FiltroCandidato{TempoExperiencia: empregabilidade.FaixaAte2Anos},
		}, 1, 10)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, cpfDiego, items[0].CPF)
	})
}

func TestFiltroCandidatoIntegration_CandidatosDaVaga(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ids := referenciasIdioma(t, tx)
	semearCandidatosFiltro(t, tx, ids)
	ctx := context.Background()

	var regime empregabilidade.RegimeContratacao
	require.NoError(t, tx.First(&regime).Error)
	var modelo empregabilidade.ModeloTrabalho
	require.NoError(t, tx.First(&modelo).Error)
	require.NoError(t, tx.Create(&empregabilidade.Empresa{CNPJ: "11222333000182", NomeFantasia: "Empresa Filtro"}).Error)
	vaga := &empregabilidade.Vaga{
		Titulo: "Vaga Filtro", Descricao: "x", IDContratante: "11222333000182",
		IDRegimeContratacao: regime.ID, IDModeloTrabalho: modelo.ID, Status: empregabilidade.StatusVagaPublicadoAtivo,
	}
	require.NoError(t, tx.Create(vaga).Error)
	for _, cand := range []*empregabilidade.Candidatura{
		{CPF: cpfAna, IDVaga: vaga.ID, Status: empregabilidade.StatusCandidaturaEnviada},
		// Gravada com máscara, como o backoffice às vezes grava.
		{CPF: "900.000.000-03", IDVaga: vaga.ID, Status: empregabilidade.StatusCandidaturaAprovada},
		{CPF: cpfBruno, IDVaga: vaga.ID, Status: empregabilidade.StatusCandidaturaReprovada},
	} {
		require.NoError(t, tx.Create(cand).Error)
	}

	repo := emprepo.NewCandidaturaRepository(tx)
	filtro := empregabilidade.CandidaturaFilter{
		VagaID:    &vaga.ID,
		Candidato: empregabilidade.FiltroCandidato{Generos: []string{"Mulher cisgênero", "Mulher transgênero"}},
	}

	candidaturas, total, err := repo.List(ctx, filtro, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	cpfs := []string{}
	for _, c := range candidaturas {
		cpfs = append(cpfs, c.CPF)
	}
	sort.Strings(cpfs)
	assert.Equal(t, []string{"900.000.000-03", cpfAna}, cpfs, "acha a candidatura gravada com máscara")

	contagem, err := repo.CountByStatus(ctx, filtro)
	require.NoError(t, err)
	assert.Equal(t, int64(1), contagem[empregabilidade.StatusCandidaturaEnviada])
	assert.Equal(t, int64(1), contagem[empregabilidade.StatusCandidaturaAprovada])
	assert.Equal(t, int64(0), contagem[empregabilidade.StatusCandidaturaReprovada], "o resumo por status respeita o filtro")

	t.Run("filtro combinado com busca e status", func(t *testing.T) {
		comStatus := filtro
		comStatus.Status = string(empregabilidade.StatusCandidaturaAprovada)
		candidaturas, total, err := repo.List(ctx, comStatus, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.Len(t, candidaturas, 1)
		assert.Equal(t, "900.000.000-03", candidaturas[0].CPF)
	})
}
