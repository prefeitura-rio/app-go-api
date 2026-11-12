package empregabilidade_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Testes de banco das tabelas por vaga de áreas de atuação, conhecimentos de
// cada área e comportamentos/atitudes (DEV-528). Executam o SQL de verdade
// contra um banco com as migrations aplicadas (just migrate-up), numa transação
// desfeita ao fim.

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

// codigoErroPg devolve o SQLSTATE do erro do Postgres, ou "" se não for um.
func codigoErroPg(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// restricaoErroPg devolve o nome da restrição violada, ou "" se não houver.
func restricaoErroPg(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// execDeveFalhar executa o comando dentro de um savepoint e devolve o erro,
// para a transação do teste seguir usável depois da violação.
func execDeveFalhar(t *testing.T, tx *gorm.DB, sql string, args ...any) error {
	t.Helper()
	require.NoError(t, tx.SavePoint("dev528").Error)
	err := tx.Exec(sql, args...).Error
	require.Error(t, err, "o banco deveria recusar: %s", sql)
	require.NoError(t, tx.RollbackTo("dev528").Error)
	return err
}

type parAreaConhecimento struct {
	IDArea       int64
	IDHabilidade int64
}

// paresDoCatalogo devolve dois pares área × conhecimento da mesma área, mais
// um conhecimento que não pertence a essa área no catálogo.
func paresDoCatalogo(t *testing.T, tx *gorm.DB) (par1, par2 parAreaConhecimento, habilidadeDeOutraArea int64) {
	t.Helper()
	var areaID int64
	require.NoError(t, tx.Exec(`INSERT INTO emp_areas_atuacao (nome) VALUES ('Área de teste DEV-528')`).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_areas_atuacao WHERE nome = 'Área de teste DEV-528'`).Scan(&areaID).Error)

	var habs []int64
	for _, nome := range []string{"Conhecimento A DEV-528", "Conhecimento B DEV-528", "Conhecimento fora DEV-528"} {
		var id int64
		require.NoError(t, tx.Raw(`INSERT INTO emp_habilidades (nome) VALUES (?) RETURNING id`, nome).Scan(&id).Error)
		habs = append(habs, id)
	}
	require.NoError(t, tx.Exec(`INSERT INTO area_atuacao_habilidade (id_habilidade, id_area_atuacao) VALUES (?, ?), (?, ?)`,
		habs[0], areaID, habs[1], areaID).Error)

	return parAreaConhecimento{IDArea: areaID, IDHabilidade: habs[0]},
		parAreaConhecimento{IDArea: areaID, IDHabilidade: habs[1]},
		habs[2]
}

func comportamentoDoCatalogo(t *testing.T, tx *gorm.DB) int64 {
	t.Helper()
	var id int64
	require.NoError(t, tx.Raw(`INSERT INTO emp_comportamento_atitudes (nome) VALUES ('Comportamento de teste DEV-528') RETURNING id`).Scan(&id).Error)
	return id
}

func contarPorVaga(t *testing.T, tx *gorm.DB, tabela, vaga string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, tx.Raw(`SELECT COUNT(*) FROM `+tabela+` WHERE id_vaga = ?`, vaga).Scan(&n).Error)
	return n
}

func TestVagaHabilidadesCompetenciasIntegration_Regras(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	vaga := vagaDeTesteParticipacao(t, tx)
	par1, par2, habilidadeDeOutraArea := paresDoCatalogo(t, tx)
	comportamento := comportamentoDoCatalogo(t, tx)

	const (
		insereArea          = `INSERT INTO emp_vagas_areas_atuacao (id_vaga, id_area_atuacao) VALUES (?, ?)`
		insereConhecimento  = `INSERT INTO emp_vagas_conhecimentos (id_vaga, id_area_atuacao, id_habilidade) VALUES (?, ?, ?)`
		insereComportamento = `INSERT INTO emp_vagas_comportamento_atitudes (id_vaga, id_comportamento_atitudes) VALUES (?, ?)`
	)

	t.Run("conhecimento sem a área registrada na vaga é recusado", func(t *testing.T) {
		err := execDeveFalhar(t, tx, insereConhecimento, vaga, par1.IDArea, par1.IDHabilidade)
		assert.Equal(t, pgForeignKeyViolation, codigoErroPg(err), "erro: %v", err)
		assert.Equal(t, "fk_emp_vagas_conhecimentos_vaga_area", restricaoErroPg(err))
	})

	t.Run("grava área, conhecimentos da área e comportamento", func(t *testing.T) {
		require.NoError(t, tx.Exec(insereArea, vaga, par1.IDArea).Error)
		require.NoError(t, tx.Exec(insereConhecimento, vaga, par1.IDArea, par1.IDHabilidade).Error)
		require.NoError(t, tx.Exec(insereConhecimento, vaga, par2.IDArea, par2.IDHabilidade).Error)
		require.NoError(t, tx.Exec(insereComportamento, vaga, comportamento).Error)

		assert.Equal(t, int64(1), contarPorVaga(t, tx, "emp_vagas_areas_atuacao", vaga))
		assert.Equal(t, int64(2), contarPorVaga(t, tx, "emp_vagas_conhecimentos", vaga))
		assert.Equal(t, int64(1), contarPorVaga(t, tx, "emp_vagas_comportamento_atitudes", vaga))
	})

	t.Run("a mesma área duas vezes na vaga é recusada", func(t *testing.T) {
		err := execDeveFalhar(t, tx, insereArea, vaga, par1.IDArea)
		assert.Equal(t, pgUniqueViolation, codigoErroPg(err), "erro: %v", err)
	})

	t.Run("o mesmo conhecimento duas vezes na mesma área da vaga é recusado", func(t *testing.T) {
		err := execDeveFalhar(t, tx, insereConhecimento, vaga, par1.IDArea, par1.IDHabilidade)
		assert.Equal(t, pgUniqueViolation, codigoErroPg(err), "erro: %v", err)
	})

	t.Run("o mesmo comportamento duas vezes na vaga é recusado", func(t *testing.T) {
		err := execDeveFalhar(t, tx, insereComportamento, vaga, comportamento)
		assert.Equal(t, pgUniqueViolation, codigoErroPg(err), "erro: %v", err)
	})

	t.Run("conhecimento que não é da área no catálogo é recusado", func(t *testing.T) {
		err := execDeveFalhar(t, tx, insereConhecimento, vaga, par1.IDArea, habilidadeDeOutraArea)
		assert.Equal(t, pgForeignKeyViolation, codigoErroPg(err), "erro: %v", err)
		assert.Equal(t, "fk_emp_vagas_conhecimentos_area_habilidade", restricaoErroPg(err))
	})

	t.Run("vaga, área ou comportamento inexistentes são recusados", func(t *testing.T) {
		const vagaInexistente = "00000000-0000-0000-0000-000000000528"
		err := execDeveFalhar(t, tx, insereArea, vagaInexistente, par1.IDArea)
		assert.Equal(t, pgForeignKeyViolation, codigoErroPg(err), "erro: %v", err)
		err = execDeveFalhar(t, tx, insereComportamento, vagaInexistente, comportamento)
		assert.Equal(t, pgForeignKeyViolation, codigoErroPg(err), "erro: %v", err)
		err = execDeveFalhar(t, tx, insereArea, vaga, int64(-1))
		assert.Equal(t, pgForeignKeyViolation, codigoErroPg(err), "erro: %v", err)
		err = execDeveFalhar(t, tx, insereComportamento, vaga, int64(-1))
		assert.Equal(t, pgForeignKeyViolation, codigoErroPg(err), "erro: %v", err)
	})

	t.Run("os mesmos itens em outra vaga são aceitos", func(t *testing.T) {
		outra := vagaDeTesteParticipacao(t, tx)
		require.NoError(t, tx.Exec(insereArea, outra, par1.IDArea).Error)
		require.NoError(t, tx.Exec(insereConhecimento, outra, par1.IDArea, par1.IDHabilidade).Error)
		require.NoError(t, tx.Exec(insereComportamento, outra, comportamento).Error)
	})

	t.Run("tirar a área da vaga tira os conhecimentos dela", func(t *testing.T) {
		outra := vagaDeTesteParticipacao(t, tx)
		require.NoError(t, tx.Exec(insereArea, outra, par1.IDArea).Error)
		require.NoError(t, tx.Exec(insereConhecimento, outra, par1.IDArea, par1.IDHabilidade).Error)
		require.NoError(t, tx.Exec(`DELETE FROM emp_vagas_areas_atuacao WHERE id_vaga = ? AND id_area_atuacao = ?`, outra, par1.IDArea).Error)
		assert.Equal(t, int64(0), contarPorVaga(t, tx, "emp_vagas_conhecimentos", outra))
	})

	t.Run("excluir a vaga remove os registros das três tabelas", func(t *testing.T) {
		require.NoError(t, tx.Exec(`DELETE FROM emp_vagas WHERE id = ?`, vaga).Error)
		assert.Equal(t, int64(0), contarPorVaga(t, tx, "emp_vagas_areas_atuacao", vaga))
		assert.Equal(t, int64(0), contarPorVaga(t, tx, "emp_vagas_conhecimentos", vaga))
		assert.Equal(t, int64(0), contarPorVaga(t, tx, "emp_vagas_comportamento_atitudes", vaga))
	})
}

func TestVagaHabilidadesCompetenciasIntegration_IndicesEmIDVaga(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)

	// Cada tabela precisa de um índice cuja primeira coluna é id_vaga: é o que
	// a leitura por vaga e a cascata da exclusão da vaga usam.
	for _, tabela := range []string{"emp_vagas_areas_atuacao", "emp_vagas_conhecimentos", "emp_vagas_comportamento_atitudes"} {
		t.Run(tabela, func(t *testing.T) {
			var n int64
			require.NoError(t, tx.Raw(`
				SELECT COUNT(*)
				FROM pg_index i
				JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = i.indkey[0]
				WHERE i.indrelid = to_regclass(?) AND a.attname = 'id_vaga'`, tabela).Scan(&n).Error)
			assert.Positive(t, n, "%s sem índice começando por id_vaga", tabela)
		})
	}
}
