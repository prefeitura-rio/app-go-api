package empregabilidade_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	emprepo "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// vagaDeTesteParticipacao cria uma vaga só para pendurar candidaturas.
func vagaDeTesteParticipacao(t *testing.T, tx *gorm.DB) string {
	t.Helper()
	var regime, modelo, vaga string
	require.NoError(t, tx.Exec(`INSERT INTO emp_empresas (cnpj) VALUES ('99999999000272') ON CONFLICT DO NOTHING`).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_regimes_contratacao LIMIT 1`).Scan(&regime).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_modelos_trabalho LIMIT 1`).Scan(&modelo).Error)
	require.NoError(t, tx.Raw(`INSERT INTO emp_vagas (titulo, descricao, id_contratante, id_regime_contratacao, id_modelo_trabalho)
		VALUES ('Vaga de teste', 'Descrição', '99999999000272', ?, ?) RETURNING id`, regime, modelo).Scan(&vaga).Error)
	return vaga
}

func origemDoCurriculo(t *testing.T, repo *emprepo.CurriculoRepository, cpf string) *string {
	t.Helper()
	participacao, err := repo.GetParticipacaoBanco(context.Background(), cpf)
	require.NoError(t, err)
	require.NotNil(t, participacao)
	return participacao.OrigemCadastro
}

func TestParticipacaoBancoIntegration_OcultarEMostrar(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)

	const (
		cpfVisivel      = "90000000001"
		cpfOculto       = "90000000002"
		cpfSemCurriculo = "90000000004"
	)
	for _, cpf := range []string{cpfVisivel, cpfOculto} {
		require.NoError(t, tx.Exec(`INSERT INTO emp_curriculos (cpf) VALUES (?)`, cpf).Error)
	}

	t.Run("sem currículo não há participação", func(t *testing.T) {
		participacao, err := repo.GetParticipacaoBanco(ctx, cpfSemCurriculo)
		require.NoError(t, err)
		assert.Nil(t, participacao)
	})

	t.Run("currículo novo já nasce visível", func(t *testing.T) {
		participacao, err := repo.GetParticipacaoBanco(ctx, cpfVisivel)
		require.NoError(t, err)
		require.NotNil(t, participacao)
		assert.True(t, participacao.Disponivel)
		assert.Nil(t, participacao.AlteradoEm)
		assert.Nil(t, participacao.OrigemCadastro)
	})

	ocultou := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)

	t.Run("ocultar registra a data", func(t *testing.T) {
		require.NoError(t, repo.AtualizarParticipacaoBanco(ctx, cpfOculto, false, ocultou))

		participacao, err := repo.GetParticipacaoBanco(ctx, cpfOculto)
		require.NoError(t, err)
		assert.False(t, participacao.Disponivel)
		require.NotNil(t, participacao.AlteradoEm)
		assert.True(t, ocultou.Equal(*participacao.AlteradoEm))
	})

	t.Run("pedir o mesmo estado não muda a data", func(t *testing.T) {
		require.NoError(t, repo.AtualizarParticipacaoBanco(ctx, cpfOculto, false, ocultou.Add(time.Hour)))

		participacao, err := repo.GetParticipacaoBanco(ctx, cpfOculto)
		require.NoError(t, err)
		assert.True(t, ocultou.Equal(*participacao.AlteradoEm))
	})

	t.Run("a listagem deixa de fora quem se ocultou", func(t *testing.T) {
		items, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{SomenteDisponiveis: true}, 1, 1000)
		require.NoError(t, err)

		cpfs := map[string]bool{}
		for _, item := range items {
			cpfs[item.CPF] = true
		}
		assert.True(t, cpfs[cpfVisivel])
		assert.False(t, cpfs[cpfOculto])
	})

	t.Run("mostrar de novo volta para a listagem", func(t *testing.T) {
		voltou := ocultou.Add(2 * time.Hour)
		require.NoError(t, repo.AtualizarParticipacaoBanco(ctx, cpfOculto, true, voltou))

		participacao, err := repo.GetParticipacaoBanco(ctx, cpfOculto)
		require.NoError(t, err)
		assert.True(t, participacao.Disponivel)
		assert.True(t, voltou.Equal(*participacao.AlteradoEm))
	})
}

func TestParticipacaoBancoIntegration_RegistrarOrigemCadastro(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	repo := emprepo.NewCurriculoRepository(tx)

	const (
		cpfMontouNaVaga = "90000000005"
		cpfJaTinhaNoApp = "90000000006"
		cpfJaRegistrado = "90000000007"
	)
	require.NoError(t, tx.Exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, NOW() - INTERVAL '10 minutes')`, cpfMontouNaVaga).Error)
	require.NoError(t, tx.Exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, NOW() - INTERVAL '3 days')`, cpfJaTinhaNoApp).Error)
	require.NoError(t, tx.Exec(`INSERT INTO emp_curriculos (cpf, origem_cadastro) VALUES (?, 'banco_curriculos')`, cpfJaRegistrado).Error)

	// O CPF pode chegar com máscara (candidatura criada pelo backoffice).
	require.NoError(t, repo.RegistrarOrigemCadastro(ctx, "900.000.000-05"))
	require.NoError(t, repo.RegistrarOrigemCadastro(ctx, cpfJaTinhaNoApp))
	require.NoError(t, repo.RegistrarOrigemCadastro(ctx, cpfJaRegistrado))

	assert.Equal(t, string(empregabilidade.OrigemCadastroVaga), *origemDoCurriculo(t, repo, cpfMontouNaVaga))
	assert.Equal(t, string(empregabilidade.OrigemCadastroBancoCurriculos), *origemDoCurriculo(t, repo, cpfJaTinhaNoApp))
	assert.Equal(t, string(empregabilidade.OrigemCadastroBancoCurriculos), *origemDoCurriculo(t, repo, cpfJaRegistrado), "não muda depois de gravada")

	// Sem currículo, não faz nada (e não dá erro).
	require.NoError(t, repo.RegistrarOrigemCadastro(ctx, "90000000008"))
}

func TestParticipacaoBancoIntegration_BackfillDaMigration(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	exec := func(sql string, args ...interface{}) {
		t.Helper()
		require.NoError(t, tx.Exec(sql, args...).Error)
	}

	const (
		cpfMontouNaVaga = "90000000011"
		cpfJaTinhaNoApp = "90000000012"
		cpfAssociado    = "90000000013"
		cpfSoCurriculo  = "90000000014"
	)
	exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, '2026-03-01 08:00:00-03')`, cpfMontouNaVaga)
	exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, '2026-01-01 08:00:00-03')`, cpfJaTinhaNoApp)
	exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, '2026-01-01 08:00:00-03')`, cpfAssociado)
	exec(`INSERT INTO emp_curriculos (cpf, created_at) VALUES (?, '2026-01-01 08:00:00-03')`, cpfSoCurriculo)
	vaga := vagaDeTesteParticipacao(t, tx)
	// Colunas sem fuso, no horário de Brasília; o backoffice grava o CPF com máscara.
	exec(`INSERT INTO emp_candidaturas (cpf, id_vaga, origem, created_at) VALUES ('900.000.000-11', ?, 'secretaria', '2026-03-01 08:20')`, vaga)
	exec(`INSERT INTO emp_candidaturas (cpf, id_vaga, origem, created_at) VALUES (?, ?, 'portal_cidadao', '2026-02-10 09:00')`, cpfJaTinhaNoApp, vaga)
	exec(`INSERT INTO emp_candidaturas (cpf, id_vaga, origem, created_at) VALUES (?, ?, 'banco_curriculos', '2026-01-01 08:30')`, cpfAssociado, vaga)

	exec(`UPDATE emp_curriculos SET origem_cadastro = NULL WHERE cpf IN ?`, []string{cpfMontouNaVaga, cpfJaTinhaNoApp, cpfAssociado, cpfSoCurriculo})
	exec(backfillOrigemDaMigration(t))

	repo := emprepo.NewCurriculoRepository(tx)
	assert.Equal(t, string(empregabilidade.OrigemCadastroVaga), *origemDoCurriculo(t, repo, cpfMontouNaVaga))
	assert.Equal(t, string(empregabilidade.OrigemCadastroBancoCurriculos), *origemDoCurriculo(t, repo, cpfJaTinhaNoApp))
	assert.Nil(t, origemDoCurriculo(t, repo, cpfAssociado), "associação não é candidatura própria")
	assert.Nil(t, origemDoCurriculo(t, repo, cpfSoCurriculo))
}

// backfillOrigemDaMigration extrai o UPDATE de origem da migration, para testar
// o SQL que vai rodar de verdade.
func backfillOrigemDaMigration(t *testing.T) string {
	t.Helper()
	conteudo, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20261022120000_add_participacao_banco_emp_curriculos.sql"))
	require.NoError(t, err)
	sql := string(conteudo)
	inicio := strings.Index(sql, "UPDATE emp_curriculos c")
	require.GreaterOrEqual(t, inicio, 0)
	fim := strings.Index(sql[inicio:], "WHERE pc.cpf = c.cpf;")
	require.GreaterOrEqual(t, fim, 0)
	return sql[inicio : inicio+fim+len("WHERE pc.cpf = c.cpf")]
}
