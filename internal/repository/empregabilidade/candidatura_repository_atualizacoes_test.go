package empregabilidade

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// atualizacaoNaoVistaSQL é o filtro das candidaturas com evento do histórico
// que o candidato ainda não viu: mudança feita pelo recrutador ou pela
// automação, menos a criação (a associação pelo banco não conta) e o
// cancelamento. O cursor é o id do evento, que cresce na ordem de gravação.
const atualizacaoNaoVistaSQL = `cpf = \$1 AND EXISTS \(SELECT 1 FROM emp_candidatura_historico h WHERE h\.id_candidatura = emp_candidaturas\.id AND h\.ator <> 'candidato' AND h\.tipo_evento NOT IN \('candidatura_criada', 'candidatura_cancelada'\) AND h\.id > COALESCE\(emp_candidaturas\.ultimo_evento_visto, 0\)\)`

func TestCandidaturaRepository_CountAtualizacoesNaoVistas(t *testing.T) {
	query := `SELECT count\(\*\) FROM "emp_candidaturas" WHERE \(` + atualizacaoNaoVistaSQL + `\) AND "emp_candidaturas"\."deleted_at" IS NULL`

	t.Run("conta as candidaturas com evento depois da última visualização", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCandidaturaRepository(db)

		mock.ExpectQuery(query).
			WithArgs("12345678900").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		total, err := repo.CountAtualizacoesNaoVistas(context.Background(), "12345678900")
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("erro do banco", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCandidaturaRepository(db)

		mock.ExpectQuery(query).WillReturnError(assert.AnError)

		_, err := repo.CountAtualizacoesNaoVistas(context.Background(), "12345678900")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao contar atualizações das candidaturas")
	})
}

func TestCandidaturaRepository_MarcarAtualizacoesComoVistas(t *testing.T) {
	// Só o cursor muda, para o último evento da candidatura: ver a lista não é
	// atualização da candidatura, então updated_at fica como está.
	exec := `UPDATE "emp_candidaturas" SET "ultimo_evento_visto"=\(SELECT MAX\(h\.id\) FROM emp_candidatura_historico h WHERE h\.id_candidatura = emp_candidaturas\.id\) WHERE \(` + atualizacaoNaoVistaSQL + `\) AND "emp_candidaturas"\."deleted_at" IS NULL`

	t.Run("marca só as candidaturas com atualização não vista", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCandidaturaRepository(db)

		mock.ExpectBegin()
		mock.ExpectExec(exec).
			WithArgs("12345678900").
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		require.NoError(t, repo.MarcarAtualizacoesComoVistas(context.Background(), "12345678900"))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("erro do banco", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCandidaturaRepository(db)

		mock.ExpectBegin()
		mock.ExpectExec(exec).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.MarcarAtualizacoesComoVistas(context.Background(), "12345678900")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao marcar atualizações das candidaturas como vistas")
	})
}
