package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestEscolaridadeRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		escolaridade := &models.Escolaridade{
			Nivel: "Ensino Superior Completo",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "escolaridades" ("nivel") VALUES ($1) RETURNING "id"`)).
			WithArgs(escolaridade.Nivel).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, escolaridade)

		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.Equal(t, 1, escolaridade.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create with database error", func(t *testing.T) {
		escolaridade := &models.Escolaridade{
			Nivel: "Test Escolaridade",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "escolaridades"`)).
			WithArgs(escolaridade.Nivel).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		id, err := repo.Create(ctx, escolaridade)

		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "erro ao criar escolaridade")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEscolaridadeRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	t.Run("successful get by id", func(t *testing.T) {
		expectedEscolaridade := &models.Escolaridade{
			ID:    1,
			Nivel: "Ensino Médio Completo",
		}

		rows := sqlmock.NewRows([]string{"id", "nivel"}).
			AddRow(expectedEscolaridade.ID, expectedEscolaridade.Nivel)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "escolaridades" WHERE "escolaridades"."id" = $1 ORDER BY "escolaridades"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnRows(rows)

		escolaridade, err := repo.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, escolaridade)
		assert.Equal(t, expectedEscolaridade.ID, escolaridade.ID)
		assert.Equal(t, expectedEscolaridade.Nivel, escolaridade.Nivel)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("escolaridade not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "escolaridades" WHERE "escolaridades"."id" = $1 ORDER BY "escolaridades"."id" LIMIT $2`)).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "nivel"}))

		escolaridade, err := repo.GetByID(ctx, 999)

		assert.NoError(t, err)
		assert.Nil(t, escolaridade)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "escolaridades" WHERE "escolaridades"."id" = $1 ORDER BY "escolaridades"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		escolaridade, err := repo.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, escolaridade)
		assert.Contains(t, err.Error(), "erro ao buscar escolaridade por ID")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEscolaridadeRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		escolaridade := &models.Escolaridade{
			ID:    1,
			Nivel: "Escolaridade atualizada",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "escolaridades" SET "nivel"=$1 WHERE id = $2 AND "id" = $3`)).
			WithArgs(escolaridade.Nivel, escolaridade.ID, escolaridade.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, escolaridade)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update with database error", func(t *testing.T) {
		escolaridade := &models.Escolaridade{
			ID:    1,
			Nivel: "Test",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "escolaridades"`)).
			WithArgs(escolaridade.Nivel, escolaridade.ID, escolaridade.ID).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Update(ctx, escolaridade)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar escolaridade")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEscolaridadeRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "escolaridades" WHERE "escolaridades"."id" = $1`)).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, 1)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete with database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "escolaridades"`)).
			WithArgs(1).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Delete(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir escolaridade")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEscolaridadeRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	t.Run("successful list without filter", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "escolaridades"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nivel"}).
			AddRow(2, "Ensino Superior").
			AddRow(1, "Ensino Médio")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "escolaridades" ORDER BY id DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		escolaridades, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, escolaridades, 2)
		assert.Equal(t, "Ensino Superior", escolaridades[0].Nivel)
		assert.Equal(t, "Ensino Médio", escolaridades[1].Nivel)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successful list with filter", func(t *testing.T) {
		filter := map[string]interface{}{
			"nivel": "Ensino Fundamental",
		}

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "escolaridades" WHERE nivel = $1`)).
			WithArgs("Ensino Fundamental").
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nivel"}).
			AddRow(1, "Ensino Fundamental")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "escolaridades" WHERE nivel = $1 ORDER BY id DESC LIMIT $2`)).
			WithArgs("Ensino Fundamental", 10).
			WillReturnRows(rows)

		escolaridades, total, err := repo.List(ctx, filter, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, escolaridades, 1)
		assert.Equal(t, "Ensino Fundamental", escolaridades[0].Nivel)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with pagination", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "escolaridades"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nivel"}).
			AddRow(3, "Nível 3").
			AddRow(2, "Nível 2")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "escolaridades" ORDER BY id DESC LIMIT $1 OFFSET $2`)).
			WithArgs(2, 2).
			WillReturnRows(rows)

		escolaridades, total, err := repo.List(ctx, map[string]interface{}{}, 2, 2)

		assert.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, escolaridades, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with database error on count", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "escolaridades"`)).
			WillReturnError(errors.New("database error"))

		escolaridades, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, escolaridades)
		assert.Contains(t, err.Error(), "erro ao listar escolaridades")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with database error on query", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "escolaridades"`)).
			WillReturnRows(countRows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "escolaridades"`)).
			WillReturnError(errors.New("database error"))

		escolaridades, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, escolaridades)
		assert.Contains(t, err.Error(), "erro ao listar escolaridades")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
