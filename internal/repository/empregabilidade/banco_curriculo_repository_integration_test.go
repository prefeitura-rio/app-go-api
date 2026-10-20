package empregabilidade_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	emprepo "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// Os testes de integração do banco de currículos executam o SQL de verdade — os
// testes com sqlmock só comparam o texto. Exigem um banco com as migrations
// aplicadas (just migrate-up) e rodam numa transação desfeita ao fim, então não
// deixam dados para trás.
func bancoCurriculosIntegrationTx(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("RUN_REPOSITORY_INTEGRATION") == "" && os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: set RUN_REPOSITORY_INTEGRATION=1 or DATABASE_URL to run")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		cfg, err := config.Load()
		if err != nil {
			t.Skipf("Skipping integration test: config load failed: %v", err)
		}
		if cfg.Database.Host == "" {
			t.Skip("Skipping integration test: DB_HOST not set")
		}
		dsn = cfg.Database.DSN()
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to database: %v", err)
	}

	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { tx.Rollback() })
	return tx
}

func mesesIntegracao(v int) *int { return &v }

func TestBancoCurriculosIntegration_EscritaListagemEBusca(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)

	const (
		cpfEmpregoAtual = "99999999901"
		cpfSemAtual     = "99999999902"
		cpfSemCadastro  = "99999999903"
	)
	empregosAtual := func() []*empregabilidade.CurriculoExperiencia {
		return []*empregabilidade.CurriculoExperiencia{
			{Cargo: "Auxiliar administrativa", Empresa: "Empresa A", TempoExperienciaMeses: mesesIntegracao(60)},
			{Cargo: "Redatora Sênior", Empresa: "BR Transportadora", EhTrabalhoAtual: true, TempoExperienciaMeses: mesesIntegracao(36)},
		}
	}

	_, totalAntes, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{}, 1, 1)
	require.NoError(t, err)

	// Grava pelos mesmos caminhos que o formulário do cidadão usa.
	require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfEmpregoAtual, empregosAtual()))
	require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfSemAtual, []*empregabilidade.CurriculoExperiencia{
		{Cargo: "Estagiária", Empresa: "Empresa B"},
		{Cargo: "Contadora", Empresa: "Empresa C", TempoExperienciaMeses: mesesIntegracao(60)},
		{Cargo: "Auxiliar", Empresa: "Empresa D", TempoExperienciaMeses: mesesIntegracao(12)},
	}))
	require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{CPF: cpfSemCadastro}))

	require.NoError(t, tx.Create(&models.CitizenSnapshot{
		CPF: cpfEmpregoAtual, Nome: "Beatriz Integração", Escolaridade: "Médio completo", LastSyncedAt: time.Now(),
	}).Error)
	require.NoError(t, tx.Create(&models.CitizenSnapshot{
		CPF: cpfSemAtual, Nome: "Ána Cláudia Integração", NomeSocial: "Aninha Integração", LastSyncedAt: time.Now(),
	}).Error)

	t.Run("toda escrita registra o currículo e editar não move a data de inclusão", func(t *testing.T) {
		antes, err := repo.GetCurriculoByCPF(ctx, cpfEmpregoAtual)
		require.NoError(t, err)
		require.NotNil(t, antes)

		// O formulário apaga e recria a seção a cada salvamento.
		require.NoError(t, repo.ReplaceAllExperienciasByCPF(ctx, cpfEmpregoAtual, empregosAtual()))

		depois, err := repo.GetCurriculoByCPF(ctx, cpfEmpregoAtual)
		require.NoError(t, err)
		assert.True(t, antes.CreatedAt.Equal(depois.CreatedAt), "antes %v, depois %v", antes.CreatedAt, depois.CreatedAt)

		soSituacao, err := repo.GetCurriculoByCPF(ctx, cpfSemCadastro)
		require.NoError(t, err)
		assert.NotNil(t, soSituacao, "situação e interesses também registra o currículo")
	})

	t.Run("lista todos, inclusive quem não tem cadastro sincronizado", func(t *testing.T) {
		_, total, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{}, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, totalAntes+3, total)
	})

	t.Run("busca ignora acento e a profissão vem do emprego atual", func(t *testing.T) {
		items, total, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "beatriz integracao"}, 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, items, 1)
		assert.Equal(t, cpfEmpregoAtual, items[0].CPF)
		require.NotNil(t, items[0].Profissao)
		assert.Equal(t, "Redatora Sênior", *items[0].Profissao)
		require.NotNil(t, items[0].Escolaridade)
		assert.Equal(t, "Médio completo", *items[0].Escolaridade)
	})

	t.Run("busca pelo nome social; sem emprego atual vale a maior experiência; vazio vira nulo", func(t *testing.T) {
		items, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "aninha integracao"}, 1, 10)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, cpfSemAtual, items[0].CPF)
		require.NotNil(t, items[0].Nome)
		assert.Equal(t, "Ána Cláudia Integração", *items[0].Nome)
		require.NotNil(t, items[0].Profissao)
		assert.Equal(t, "Contadora", *items[0].Profissao)
		assert.Nil(t, items[0].Escolaridade)
	})

	t.Run("ordena da inclusão mais recente para a mais antiga", func(t *testing.T) {
		items, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "integracao"}, 1, 10)
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, cpfSemAtual, items[0].CPF)
		assert.Equal(t, cpfEmpregoAtual, items[1].CPF)
	})
}

func TestBancoCurriculosIntegration_BuscaPorCPF(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)

	// O CPF é gravado só com dígitos; o operador digita com ou sem máscara.
	const (
		cpfProcurado = "98765432100"
		cpfOutro     = "12345678909"
	)

	require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{CPF: cpfProcurado}))
	require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{CPF: cpfOutro}))
	require.NoError(t, tx.Create(&models.CitizenSnapshot{
		CPF: cpfProcurado, Nome: "Carlos Procurado Integração", LastSyncedAt: time.Now(),
	}).Error)
	require.NoError(t, tx.Create(&models.CitizenSnapshot{
		CPF: cpfOutro, Nome: "Diego Outro Integração", LastSyncedAt: time.Now(),
	}).Error)

	// A base local pode ter outros currículos, então o teste olha quem veio, não
	// o total. A página é grande para o esperado não ficar de fora.
	cpfsEncontrados := func(t *testing.T, busca string) []string {
		t.Helper()
		items, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: busca}, 1, 100)
		require.NoError(t, err)
		cpfs := make([]string, 0, len(items))
		for _, item := range items {
			cpfs = append(cpfs, item.CPF)
		}
		return cpfs
	}

	t.Run("acha pelo CPF inteiro", func(t *testing.T) {
		cpfs := cpfsEncontrados(t, cpfProcurado)
		assert.Contains(t, cpfs, cpfProcurado)
		assert.NotContains(t, cpfs, cpfOutro)
	})

	t.Run("acha pelo CPF com máscara", func(t *testing.T) {
		cpfs := cpfsEncontrados(t, " 987.654.321-00 ")
		assert.Contains(t, cpfs, cpfProcurado)
		assert.NotContains(t, cpfs, cpfOutro)
	})

	t.Run("acha por trecho do CPF", func(t *testing.T) {
		cpfs := cpfsEncontrados(t, "76543210")
		assert.Contains(t, cpfs, cpfProcurado)
		assert.NotContains(t, cpfs, cpfOutro)
	})

	t.Run("busca por nome continua funcionando e não mistura os dois", func(t *testing.T) {
		cpfs := cpfsEncontrados(t, "diego outro integracao")
		assert.Contains(t, cpfs, cpfOutro)
		assert.NotContains(t, cpfs, cpfProcurado)
	})

	t.Run("CPF que não existe não traz ninguém", func(t *testing.T) {
		items, total, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "55555555555"}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, items)
	})
}

func TestBancoCurriculosIntegration_GetUltimaAtualizacao(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	require.NoError(t, err)
	exec := func(sql string, args ...interface{}) {
		t.Helper()
		require.NoError(t, tx.Exec(sql, args...).Error)
	}

	const (
		cpfVariasSecoes = "99999999907"
		cpfUmaSecao     = "99999999908"
		cpfSemSecao     = "99999999909"
	)

	// updated_at das seções não tem fuso e guarda horário de Brasília.
	exec(`INSERT INTO emp_curriculo_experiencias (cpf, cargo, empresa, updated_at) VALUES (?, 'Redatora', 'Empresa A', '2026-05-10 09:00')`, cpfVariasSecoes)
	exec(`INSERT INTO emp_curriculo_formacoes (cpf, nome_curso, updated_at) VALUES (?, 'Letras', '2026-06-20 16:45')`, cpfVariasSecoes)
	exec(`INSERT INTO emp_curriculo_perfil (cpf, resumo_profissional, updated_at) VALUES (?, 'Resumo', '2026-04-01 08:00')`, cpfVariasSecoes)
	exec(`INSERT INTO emp_curriculo_situacao_interesses (cpf, updated_at) VALUES (?, '2026-07-05 18:30')`, cpfUmaSecao)

	t.Run("vale a seção salva por último, em horário de Brasília", func(t *testing.T) {
		quando, err := repo.GetUltimaAtualizacao(ctx, cpfVariasSecoes)
		require.NoError(t, err)
		require.NotNil(t, quando)
		assert.True(t, quando.Equal(time.Date(2026, 6, 20, 16, 45, 0, 0, saoPaulo)), "obtido %v", quando)
	})

	t.Run("a data de outro cidadão não vaza", func(t *testing.T) {
		quando, err := repo.GetUltimaAtualizacao(ctx, cpfUmaSecao)
		require.NoError(t, err)
		require.NotNil(t, quando)
		assert.True(t, quando.Equal(time.Date(2026, 7, 5, 18, 30, 0, 0, saoPaulo)), "obtido %v", quando)
	})

	t.Run("sem nenhuma seção não tem data de atualização", func(t *testing.T) {
		quando, err := repo.GetUltimaAtualizacao(ctx, cpfSemSecao)
		require.NoError(t, err)
		assert.Nil(t, quando)
	})
}

func TestBancoCurriculosIntegration_BackfillDaMigration(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	require.NoError(t, err)
	exec := func(sql string, args ...interface{}) {
		t.Helper()
		require.NoError(t, tx.Exec(sql, args...).Error)
	}

	const (
		cpfEditouDepois  = "99999999904" // candidatou-se em março, editou o currículo em maio
		cpfSoCurriculo   = "99999999905" // nunca se candidatou
		cpfSoCandidatura = "99999999906" // candidatura sem currículo: não entra no banco
	)

	exec(`INSERT INTO emp_curriculo_experiencias (cpf, cargo, empresa, created_at) VALUES (?, 'Redatora', 'Empresa A', '2026-05-10 09:00')`, cpfEditouDepois)
	exec(`INSERT INTO emp_curriculo_experiencias (cpf, cargo, empresa, created_at) VALUES (?, 'Contadora', 'Empresa B', '2026-06-15 10:00')`, cpfSoCurriculo)
	exec(`INSERT INTO emp_curriculo_perfil (cpf, resumo_profissional, created_at) VALUES (?, '', '2026-06-20 10:00')`, cpfSoCurriculo)

	var regime, modelo, vaga string
	exec(`INSERT INTO emp_empresas (cnpj) VALUES ('99999999000191')`)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_regimes_contratacao LIMIT 1`).Scan(&regime).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_modelos_trabalho LIMIT 1`).Scan(&modelo).Error)
	require.NoError(t, tx.Raw(`INSERT INTO emp_vagas (titulo, descricao, id_contratante, id_regime_contratacao, id_modelo_trabalho)
		VALUES ('Vaga de teste', 'Descrição', '99999999000191', ?, ?) RETURNING id`, regime, modelo).Scan(&vaga).Error)
	exec(`INSERT INTO emp_candidaturas (cpf, id_vaga, created_at) VALUES (?, ?, '2026-03-01 08:00')`, cpfEditouDepois, vaga)
	exec(`INSERT INTO emp_candidaturas (cpf, id_vaga, created_at) VALUES (?, ?, '2026-02-01 08:00')`, cpfSoCandidatura, vaga)

	// Refaz o backfill do zero, como na primeira execução da migration.
	exec(`DELETE FROM emp_curriculos`)
	exec(backfillDaMigration(t))

	var linhas []struct {
		CPF       string
		CreatedAt time.Time
	}
	require.NoError(t, tx.Raw(`SELECT cpf, created_at FROM emp_curriculos WHERE cpf IN ?`,
		[]string{cpfEditouDepois, cpfSoCurriculo, cpfSoCandidatura}).Scan(&linhas).Error)
	datas := map[string]time.Time{}
	for _, l := range linhas {
		datas[l.CPF] = l.CreatedAt
	}

	require.Contains(t, datas, cpfEditouDepois)
	assert.True(t, datas[cpfEditouDepois].Equal(time.Date(2026, 3, 1, 8, 0, 0, 0, saoPaulo)),
		"a candidatura é anterior à última edição: %v", datas[cpfEditouDepois])
	require.Contains(t, datas, cpfSoCurriculo)
	assert.True(t, datas[cpfSoCurriculo].Equal(time.Date(2026, 6, 15, 10, 0, 0, 0, saoPaulo)),
		"vale a seção mais antiga: %v", datas[cpfSoCurriculo])
	assert.NotContains(t, datas, cpfSoCandidatura)
}

// backfillDaMigration extrai o INSERT de backfill da migration, para testar o
// SQL exato que vai rodar em produção.
func backfillDaMigration(t *testing.T) string {
	t.Helper()
	conteudo, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260910120000_create_emp_curriculos.sql"))
	require.NoError(t, err)

	up := strings.Split(string(conteudo), "-- +goose Down")[0]
	blocos := strings.Split(up, "-- +goose StatementBegin")
	backfill := strings.Split(blocos[len(blocos)-1], "-- +goose StatementEnd")[0]
	require.Contains(t, backfill, "INSERT INTO emp_curriculos")
	return backfill
}

func TestBancoCurriculosIntegration_SalvarDeNovoPreservaCreatedAt(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)
	const cpf = "99999999907"

	criadoEm := func(tabela string) time.Time {
		t.Helper()
		var criado time.Time
		require.NoError(t, tx.Raw(`SELECT created_at FROM `+tabela+` WHERE cpf = ?`, cpf).Scan(&criado).Error)
		require.True(t, criado.Year() > 1900, "%s gravou created_at zerado: %v", tabela, criado)
		return criado
	}

	t.Run("situação e interesses", func(t *testing.T) {
		require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{
			CPF: cpf, TempoProcurandoEmprego: "UP_TO_6",
		}))
		antes := criadoEm("emp_curriculo_situacao_interesses")

		// Segundo salvamento, como o handler monta: entidade nova, CreatedAt zerado.
		require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{
			CPF: cpf, TempoProcurandoEmprego: "OVER_24",
		}))

		assert.True(t, criadoEm("emp_curriculo_situacao_interesses").Equal(antes))
		salvo, err := repo.GetSituacaoInteressesByCPF(ctx, cpf)
		require.NoError(t, err)
		assert.Equal(t, "OVER_24", salvo.TempoProcurandoEmprego)
	})

	t.Run("formação", func(t *testing.T) {
		id, err := repo.CreateFormacao(ctx, &empregabilidade.CurriculoFormacao{
			CPF: cpf, NomeInstituicao: "Instituição Integração", NomeCurso: "Curso A",
		})
		require.NoError(t, err)
		antes := criadoEm("emp_curriculo_formacoes")

		require.NoError(t, repo.UpdateFormacao(ctx, &empregabilidade.CurriculoFormacao{
			ID: id, CPF: cpf, NomeInstituicao: "Instituição Integração", NomeCurso: "Curso B",
		}))

		assert.True(t, criadoEm("emp_curriculo_formacoes").Equal(antes))
		salva, err := repo.GetFormacaoByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "Curso B", salva.NomeCurso)
	})
}

func TestBancoCurriculosIntegration_BuscaTrataCuringasComoTexto(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)
	const (
		cpfComPorcento = "99999999908"
		cpfSemPorcento = "99999999909"
	)

	for cpf, nome := range map[string]string{
		cpfComPorcento: "Promo 100% Integração",
		cpfSemPorcento: "Beatriz Integração",
	} {
		require.NoError(t, repo.UpsertSituacaoInteresses(ctx, &empregabilidade.CurriculoSituacaoInteresses{CPF: cpf}))
		require.NoError(t, tx.Create(&models.CitizenSnapshot{CPF: cpf, Nome: nome, LastSyncedAt: time.Now()}).Error)
	}

	items, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "%"}, 1, 100)
	require.NoError(t, err)
	cpfs := map[string]bool{}
	for _, item := range items {
		cpfs[item.CPF] = true
	}
	assert.True(t, cpfs[cpfComPorcento], "o nome com %% deve casar")
	assert.False(t, cpfs[cpfSemPorcento], "%% não pode funcionar como curinga")
}

func TestBancoCurriculosIntegration_CorrecaoDaDataZerada(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	require.NoError(t, err)
	exec := func(sql string, args ...interface{}) {
		t.Helper()
		require.NoError(t, tx.Exec(sql, args...).Error)
	}

	const (
		cpfZerado = "99999999910" // salvou situação e interesses duas vezes
		cpfCerto  = "99999999911" // data de inclusão correta: não pode mudar
	)
	// O estado que o Save deixava: seção com created_at no zero do Go.
	exec(`INSERT INTO emp_curriculo_situacao_interesses (cpf, created_at, updated_at) VALUES (?, '0001-01-01 00:00', '2026-05-20 16:07')`, cpfZerado)
	exec(`INSERT INTO emp_curriculo_experiencias (cpf, cargo, empresa, created_at, updated_at) VALUES (?, 'Redatora', 'Empresa A', '2026-06-01 09:00', '2026-06-01 09:00')`, cpfZerado)
	exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, '0001-01-01 00:00:00-03:06')`, cpfZerado)
	exec(`INSERT INTO emp_curriculo_experiencias (cpf, cargo, empresa, created_at, updated_at) VALUES (?, 'Contadora', 'Empresa B', '2026-04-01 09:00', '2026-04-01 09:00')`, cpfCerto)
	exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, '2026-03-01 08:00:00-03')`, cpfCerto)

	for _, comando := range comandosDaMigration(t, "20260924120000_fix_zeroed_curriculo_created_at.sql") {
		exec(comando)
	}

	var secao time.Time
	require.NoError(t, tx.Raw(`SELECT created_at FROM emp_curriculo_situacao_interesses WHERE cpf = ?`, cpfZerado).Scan(&secao).Error)
	assert.True(t, secao.Equal(time.Date(2026, 5, 20, 16, 7, 0, 0, time.UTC)), "a seção herda o updated_at: %v", secao)

	datas := map[string]time.Time{}
	var linhas []struct {
		CPF       string
		CreatedAt time.Time
	}
	require.NoError(t, tx.Raw(`SELECT cpf, created_at FROM emp_curriculos WHERE cpf IN ?`, []string{cpfZerado, cpfCerto}).Scan(&linhas).Error)
	for _, l := range linhas {
		datas[l.CPF] = l.CreatedAt
	}
	assert.True(t, datas[cpfZerado].Equal(time.Date(2026, 5, 20, 16, 7, 0, 0, saoPaulo)),
		"vale a seção mais antiga depois da correção: %v", datas[cpfZerado])
	assert.True(t, datas[cpfCerto].Equal(time.Date(2026, 3, 1, 8, 0, 0, 0, saoPaulo)),
		"data correta não muda: %v", datas[cpfCerto])
}

// comandosDaMigration separa os comandos do Up de uma migration, para testar o
// SQL exato que vai rodar em produção.
func comandosDaMigration(t *testing.T, arquivo string) []string {
	t.Helper()
	conteudo, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", arquivo))
	require.NoError(t, err)

	// Tira os comentários antes de separar: eles podem ter ";".
	up := strings.Split(string(conteudo), "-- +goose Down")[0]
	var linhas []string
	for _, linha := range strings.Split(up, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(linha), "--") {
			linhas = append(linhas, linha)
		}
	}
	var comandos []string
	for _, comando := range strings.Split(strings.Join(linhas, "\n"), ";") {
		if sql := strings.TrimSpace(comando); sql != "" {
			comandos = append(comandos, sql)
		}
	}
	require.NotEmpty(t, comandos)
	return comandos
}

func TestBancoCurriculosIntegration_HistoricoDeCandidaturas(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)

	const (
		cpfCandidato = "98765432100"
		cpfOutro     = "12345678909"
	)

	var regime empregabilidade.RegimeContratacao
	require.NoError(t, tx.First(&regime).Error)
	var modelo empregabilidade.ModeloTrabalho
	require.NoError(t, tx.First(&modelo).Error)
	empresa := &empregabilidade.Empresa{CNPJ: "11222333000181", NomeFantasia: "SMPD Integração"}
	require.NoError(t, tx.Create(empresa).Error)

	novaVaga := func(titulo string) *empregabilidade.Vaga {
		t.Helper()
		vaga := &empregabilidade.Vaga{
			Titulo:              titulo,
			Descricao:           "Descrição da vaga",
			IDContratante:       empresa.CNPJ,
			IDRegimeContratacao: regime.ID,
			IDModeloTrabalho:    modelo.ID,
			Status:              empregabilidade.StatusVagaPublicadoAtivo,
		}
		require.NoError(t, tx.Create(vaga).Error)
		return vaga
	}
	novaCandidatura := func(cpf string, vaga *empregabilidade.Vaga, criada time.Time) *empregabilidade.Candidatura {
		t.Helper()
		candidatura := &empregabilidade.Candidatura{
			CPF:       cpf,
			IDVaga:    vaga.ID,
			Status:    empregabilidade.StatusCandidaturaEnviada,
			CreatedAt: criada,
		}
		require.NoError(t, tx.Create(candidatura).Error)
		return candidatura
	}

	base := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	escritorio := novaVaga("Auxiliar de escritório Integração")
	administracao := novaVaga("Auxiliar de administração Integração")
	excluida := novaVaga("Recepcionista Integração")
	cancelada := novaVaga("Motorista Integração")

	maisAntiga := novaCandidatura(cpfCandidato, escritorio, base)
	origem := empregabilidade.OrigemCandidaturaBancoCurriculos
	operador := "22222222222"
	require.NoError(t, tx.Model(maisAntiga).Updates(map[string]interface{}{"origem": origem, "associado_por_cpf": operador}).Error)
	// O backoffice grava o CPF como veio na requisição, às vezes com máscara.
	comMascara := novaCandidatura("987.654.321-00", administracao, base.Add(time.Hour))
	daVagaExcluida := novaCandidatura(cpfCandidato, excluida, base.Add(2*time.Hour))
	desfeita := novaCandidatura(cpfCandidato, cancelada, base.Add(3*time.Hour))
	novaCandidatura(cpfOutro, escritorio, base.Add(4*time.Hour))

	require.NoError(t, tx.Delete(excluida).Error)
	require.NoError(t, tx.Delete(desfeita).Error)

	t.Run("lista as candidaturas do CPF, da mais recente para a mais antiga", func(t *testing.T) {
		historico, total, err := repo.ListCandidaturasByCPF(ctx, cpfCandidato, empregabilidade.BancoCurriculoCandidaturaFilter{}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total, "ignora a candidatura excluída e a de outro CPF")
		require.Len(t, historico, 3)
		assert.Equal(t, daVagaExcluida.ID, historico[0].ID)
		assert.Equal(t, comMascara.ID, historico[1].ID)
		assert.Equal(t, maisAntiga.ID, historico[2].ID)
	})

	t.Run("carrega a vaga com contratante, regime e modelo, mesmo excluída", func(t *testing.T) {
		historico, _, err := repo.ListCandidaturasByCPF(ctx, cpfCandidato, empregabilidade.BancoCurriculoCandidaturaFilter{}, 1, 10)
		require.NoError(t, err)
		require.Len(t, historico, 3)

		require.NotNil(t, historico[0].Vaga)
		assert.Equal(t, "Recepcionista Integração", historico[0].Vaga.Titulo)

		vaga := historico[2].Vaga
		require.NotNil(t, vaga)
		assert.Equal(t, escritorio.ID, vaga.ID)
		require.NotNil(t, vaga.Contratante)
		assert.Equal(t, "SMPD Integração", vaga.Contratante.NomeFantasia)
		require.NotNil(t, vaga.RegimeContratacao)
		assert.Equal(t, regime.Descricao, vaga.RegimeContratacao.Descricao)
		require.NotNil(t, vaga.ModeloTrabalho)
		assert.Equal(t, modelo.Descricao, vaga.ModeloTrabalho.Descricao)
		assert.Nil(t, historico[2].CurriculoSnapshot, "o snapshot do currículo não entra no histórico")
	})

	t.Run("traz a origem da candidatura", func(t *testing.T) {
		historico, _, err := repo.ListCandidaturasByCPF(ctx, cpfCandidato, empregabilidade.BancoCurriculoCandidaturaFilter{}, 1, 10)
		require.NoError(t, err)
		require.Len(t, historico, 3)
		require.NotNil(t, historico[2].Origem)
		assert.Equal(t, empregabilidade.OrigemCandidaturaBancoCurriculos, *historico[2].Origem)
		assert.Nil(t, historico[1].Origem, "candidatura antiga fica sem origem")
	})

	t.Run("diz em quais vagas o CPF já participa, com ou sem máscara e sem as excluídas", func(t *testing.T) {
		outraVaga := novaVaga("Vaga sem candidatura Integração")
		participa, err := repo.VagasComCandidatura(ctx, cpfCandidato, []uuid.UUID{escritorio.ID, administracao.ID, cancelada.ID, outraVaga.ID})
		require.NoError(t, err)
		assert.True(t, participa[escritorio.ID])
		assert.True(t, participa[administracao.ID], "candidatura gravada com máscara")
		assert.False(t, participa[cancelada.ID], "candidatura excluída não conta")
		assert.False(t, participa[outraVaga.ID])

		vazio, err := repo.VagasComCandidatura(ctx, cpfCandidato, nil)
		require.NoError(t, err)
		assert.Empty(t, vazio)
	})

	t.Run("busca pelo título da vaga ignorando acento", func(t *testing.T) {
		historico, total, err := repo.ListCandidaturasByCPF(ctx, cpfCandidato, empregabilidade.BancoCurriculoCandidaturaFilter{Search: "escritorio integracao"}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, historico, 1)
		assert.Equal(t, maisAntiga.ID, historico[0].ID)
	})

	t.Run("pagina", func(t *testing.T) {
		historico, total, err := repo.ListCandidaturasByCPF(ctx, cpfCandidato, empregabilidade.BancoCurriculoCandidaturaFilter{}, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		require.Len(t, historico, 1)
		assert.Equal(t, maisAntiga.ID, historico[0].ID)
	})

	t.Run("CPF sem candidatura devolve lista vazia", func(t *testing.T) {
		historico, total, err := repo.ListCandidaturasByCPF(ctx, "11111111111", empregabilidade.BancoCurriculoCandidaturaFilter{}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.NotNil(t, historico)
		assert.Empty(t, historico)
	})
}
