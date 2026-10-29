package empregabilidade

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O SQL de verdade é exercitado em participacao_banco_integration_test.go.

func TestCurriculoRepository_GetParticipacaoBanco(t *testing.T) {
	ctx := context.Background()
	query := regexp.QuoteMeta(`SELECT disponivel_banco, disponivel_banco_alterado_em, origem_cadastro FROM emp_curriculos WHERE cpf = $1`)
	cols := []string{"disponivel_banco", "disponivel_banco_alterado_em", "origem_cadastro"}

	t.Run("sem currículo", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(query).WithArgs("11144477735").WillReturnRows(sqlmock.NewRows(cols))

		participacao, err := NewCurriculoRepository(db).GetParticipacaoBanco(ctx, "111.444.777-35")

		require.NoError(t, err)
		assert.Nil(t, participacao)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ligado com origem", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		alterado := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
		mock.ExpectQuery(query).WithArgs("11144477735").
			WillReturnRows(sqlmock.NewRows(cols).AddRow(true, alterado, "banco_curriculos"))

		participacao, err := NewCurriculoRepository(db).GetParticipacaoBanco(ctx, "11144477735")

		require.NoError(t, err)
		require.NotNil(t, participacao)
		assert.True(t, participacao.Disponivel)
		require.NotNil(t, participacao.AlteradoEm)
		assert.True(t, alterado.Equal(*participacao.AlteradoEm))
		require.NotNil(t, participacao.OrigemCadastro)
		assert.Equal(t, "banco_curriculos", *participacao.OrigemCadastro)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("desligado, nunca mexeu", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(query).WithArgs("11144477735").
			WillReturnRows(sqlmock.NewRows(cols).AddRow(false, nil, nil))

		participacao, err := NewCurriculoRepository(db).GetParticipacaoBanco(ctx, "11144477735")

		require.NoError(t, err)
		require.NotNil(t, participacao)
		assert.False(t, participacao.Disponivel)
		assert.Nil(t, participacao.AlteradoEm)
		assert.Nil(t, participacao.OrigemCadastro)
	})

	t.Run("erro", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(query).WillReturnError(sql.ErrConnDone)

		_, err := NewCurriculoRepository(db).GetParticipacaoBanco(ctx, "11144477735")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "participação no banco de currículos")
	})
}

func TestCurriculoRepository_AtualizarParticipacaoBanco(t *testing.T) {
	ctx := context.Background()
	quando := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)

	query := regexp.QuoteMeta(`UPDATE emp_curriculos SET disponivel_banco = $1, disponivel_banco_alterado_em = $2 WHERE cpf = $3 AND disponivel_banco <> $4`)

	for _, disponivel := range []bool{true, false} {
		t.Run("só muda quando o estado muda", func(t *testing.T) {
			db, mock, cleanup := repository.SetupMockDB(t)
			defer cleanup()
			mock.ExpectExec(query).
				WithArgs(disponivel, quando, "11144477735", disponivel).
				WillReturnResult(sqlmock.NewResult(0, 1))

			require.NoError(t, NewCurriculoRepository(db).AtualizarParticipacaoBanco(ctx, "111.444.777-35", disponivel, quando))
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("erro", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectExec(`UPDATE emp_curriculos`).WillReturnError(sql.ErrConnDone)

		err := NewCurriculoRepository(db).AtualizarParticipacaoBanco(ctx, "11144477735", false, quando)

		require.Error(t, err)
	})
}

func TestCurriculoRepository_RegistrarOrigemCadastro(t *testing.T) {
	ctx := context.Background()
	query := `UPDATE emp_curriculos SET origem_cadastro = CASE\s+WHEN created_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours' THEN 'vaga'\s+ELSE 'banco_curriculos'\s+END\s+WHERE cpf = \$1 AND origem_cadastro IS NULL`

	t.Run("normaliza o CPF", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectExec(query).WithArgs("11144477735").WillReturnResult(sqlmock.NewResult(0, 1))

		require.NoError(t, NewCurriculoRepository(db).RegistrarOrigemCadastro(ctx, "111.444.777-35"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("erro", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectExec(query).WillReturnError(sql.ErrConnDone)

		require.Error(t, NewCurriculoRepository(db).RegistrarOrigemCadastro(ctx, "11144477735"))
	})
}

func TestCurriculoRepository_ListBancoCurriculos_SomenteDisponiveis(t *testing.T) {
	ctx := context.Background()

	t.Run("sozinho", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(regexp.QuoteMeta(bancoCurriculosCountSQL + ` WHERE c.disponivel_banco`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		_, total, err := NewCurriculoRepository(db).ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{SomenteDisponiveis: true}, 1, 10)

		require.NoError(t, err)
		assert.Zero(t, total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("junto com a busca", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(regexp.QuoteMeta(`unaccent(cs.nome_social) ILIKE unaccent($2)) AND c.disponivel_banco`)).
			WithArgs("%ana%", "%ana%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		_, _, err := NewCurriculoRepository(db).ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "ana", SomenteDisponiveis: true}, 1, 10)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
