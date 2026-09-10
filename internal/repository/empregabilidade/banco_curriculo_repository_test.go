package empregabilidade

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bancoCurriculosCountSQL = `SELECT COUNT(*) FROM emp_curriculos c LEFT JOIN citizen_snapshots cs ON cs.cpf = c.cpf`

var bancoCurriculosListCols = []string{"cpf", "data_inclusao", "nome", "nome_social", "escolaridade", "profissao"}

func TestCurriculoRepository_ListBancoCurriculos(t *testing.T) {
	ctx := context.Background()
	inclusao := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

	t.Run("sem busca pagina pela inclusão mais recente", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectQuery(regexp.QuoteMeta(bancoCurriculosCountSQL)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))
		mock.ExpectQuery(`ORDER BY c\.created_at DESC, c\.cpf ASC LIMIT \$1 OFFSET \$2`).
			WithArgs(10, 10).
			WillReturnRows(sqlmock.NewRows(bancoCurriculosListCols).
				AddRow("11111111111", inclusao, "Ana Silva", nil, "Médio completo", "Redatora").
				AddRow("22222222222", inclusao, nil, nil, nil, nil))

		items, total, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{}, 2, 10)

		require.NoError(t, err)
		assert.Equal(t, int64(12), total)
		require.Len(t, items, 2)
		assert.Equal(t, "11111111111", items[0].CPF)
		require.NotNil(t, items[0].Nome)
		assert.Equal(t, "Ana Silva", *items[0].Nome)
		require.NotNil(t, items[0].Profissao)
		assert.Equal(t, "Redatora", *items[0].Profissao)
		require.NotNil(t, items[0].Escolaridade)
		assert.Equal(t, "Médio completo", *items[0].Escolaridade)
		assert.True(t, inclusao.Equal(items[0].DataInclusao))
		assert.Nil(t, items[1].Nome)
		assert.Nil(t, items[1].Profissao)
		assert.Nil(t, items[1].Escolaridade)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("busca por nome ou nome social ignorando acento", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		where := ` WHERE (unaccent(cs.nome) ILIKE unaccent($1) OR unaccent(cs.nome_social) ILIKE unaccent($2))`
		mock.ExpectQuery(regexp.QuoteMeta(bancoCurriculosCountSQL+where)).
			WithArgs("%ana%", "%ana%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(where+` ORDER BY c.created_at DESC, c.cpf ASC LIMIT $3 OFFSET $4`)).
			WithArgs("%ana%", "%ana%", 10, 0).
			WillReturnRows(sqlmock.NewRows(bancoCurriculosListCols).
				AddRow("11111111111", inclusao, "Ana Silva", nil, nil, nil))

		items, total, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{Search: "  ana "}, 1, 10)

		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, items, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("profissão vem do emprego atual ou da maior experiência", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectQuery(regexp.QuoteMeta(bancoCurriculosCountSQL)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM emp_curriculo_experiencias e WHERE e.cpf = c.cpf ORDER BY e.eh_trabalho_atual DESC, e.tempo_experiencia_meses DESC NULLS LAST, e.cargo ASC LIMIT 1`)).
			WillReturnRows(sqlmock.NewRows(bancoCurriculosListCols).
				AddRow("11111111111", inclusao, nil, nil, nil, "Redatora"))

		_, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{}, 1, 10)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sem resultados não consulta a página", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectQuery(regexp.QuoteMeta(bancoCurriculosCountSQL)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		items, total, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{}, 1, 10)

		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.NotNil(t, items)
		assert.Empty(t, items)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("erro na contagem", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectQuery(regexp.QuoteMeta(bancoCurriculosCountSQL)).
			WillReturnError(sql.ErrConnDone)

		items, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{}, 1, 10)

		require.Error(t, err)
		assert.Nil(t, items)
		assert.Contains(t, err.Error(), "erro ao contar currículos")
	})

	t.Run("erro na listagem", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectQuery(regexp.QuoteMeta(bancoCurriculosCountSQL)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
		mock.ExpectQuery(`ORDER BY c\.created_at DESC`).
			WillReturnError(sql.ErrConnDone)

		items, _, err := repo.ListBancoCurriculos(ctx, empregabilidade.BancoCurriculoFilter{}, 1, 10)

		require.Error(t, err)
		assert.Nil(t, items)
		assert.Contains(t, err.Error(), "erro ao listar currículos")
	})
}

func TestCurriculoRepository_RegistroDoCurriculo(t *testing.T) {
	ctx := context.Background()
	experiencia := func() *empregabilidade.CurriculoExperiencia {
		return &empregabilidade.CurriculoExperiencia{CPF: "11111111111", Cargo: "Redatora", Empresa: "BR Transportadora"}
	}

	t.Run("não sobrescreve a data de quem já está no banco", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "emp_curriculos" ("cpf","created_at") VALUES ($1,$2) ON CONFLICT ("cpf") DO NOTHING`)).
			WithArgs("11111111111", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		_, err := repo.CreateExperiencia(ctx, experiencia())

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("falha ao registrar desfaz a escrita da seção", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "emp_curriculos"`)).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		id, err := repo.CreateExperiencia(ctx, experiencia())

		require.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
		assert.Contains(t, err.Error(), "erro ao registrar currículo")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCurriculoRepository_GetCurriculoByCPF(t *testing.T) {
	ctx := context.Background()
	cpf := "11111111111"
	query := regexp.QuoteMeta(`SELECT * FROM "emp_curriculos" WHERE cpf = $1 ORDER BY "emp_curriculos"."cpf" LIMIT $2`)

	t.Run("encontrado", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)
		inclusao := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

		mock.ExpectQuery(query).
			WithArgs(cpf, 1).
			WillReturnRows(sqlmock.NewRows([]string{"cpf", "created_at"}).AddRow(cpf, inclusao))

		curriculo, err := repo.GetCurriculoByCPF(ctx, cpf)

		require.NoError(t, err)
		require.NotNil(t, curriculo)
		assert.Equal(t, cpf, curriculo.CPF)
		assert.True(t, inclusao.Equal(curriculo.CreatedAt))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("não encontrado", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectQuery(query).
			WithArgs(cpf, 1).
			WillReturnRows(sqlmock.NewRows([]string{"cpf", "created_at"}))

		curriculo, err := repo.GetCurriculoByCPF(ctx, cpf)

		require.NoError(t, err)
		assert.Nil(t, curriculo)
	})

	t.Run("erro", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewCurriculoRepository(db)

		mock.ExpectQuery(query).WillReturnError(sql.ErrConnDone)

		curriculo, err := repo.GetCurriculoByCPF(ctx, cpf)

		require.Error(t, err)
		assert.Nil(t, curriculo)
		assert.Contains(t, err.Error(), "erro ao buscar currículo")
	})
}
