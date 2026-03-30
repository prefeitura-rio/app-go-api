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

func TestAcessibilidadeRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewAcessibilidadeRepository(db)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		acessibilidade := &models.Acessibilidade{
			Nome: "Acessibilidade para cadeirantes",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "acessibilidades" ("nome") VALUES ($1) RETURNING "id"`)).
			WithArgs(acessibilidade.Nome).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, acessibilidade)

		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.Equal(t, 1, acessibilidade.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create with database error", func(t *testing.T) {
		acessibilidade := &models.Acessibilidade{
			Nome: "Test Acessibilidade",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "acessibilidades"`)).
			WithArgs(acessibilidade.Nome).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		id, err := repo.Create(ctx, acessibilidade)

		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "erro ao criar acessibilidade")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAcessibilidadeRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewAcessibilidadeRepository(db)
	ctx := context.Background()

	t.Run("successful get by id", func(t *testing.T) {
		expectedAcessibilidade := &models.Acessibilidade{
			ID:   1,
			Nome: "Libras disponível",
		}

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(expectedAcessibilidade.ID, expectedAcessibilidade.Nome)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "acessibilidades" WHERE "acessibilidades"."id" = $1 ORDER BY "acessibilidades"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnRows(rows)

		acessibilidade, err := repo.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, acessibilidade)
		assert.Equal(t, expectedAcessibilidade.ID, acessibilidade.ID)
		assert.Equal(t, expectedAcessibilidade.Nome, acessibilidade.Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("acessibilidade not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "acessibilidades" WHERE "acessibilidades"."id" = $1 ORDER BY "acessibilidades"."id" LIMIT $2`)).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}))

		acessibilidade, err := repo.GetByID(ctx, 999)

		assert.NoError(t, err)
		assert.Nil(t, acessibilidade)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "acessibilidades" WHERE "acessibilidades"."id" = $1 ORDER BY "acessibilidades"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		acessibilidade, err := repo.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, acessibilidade)
		assert.Contains(t, err.Error(), "erro ao buscar acessibilidade por ID")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAcessibilidadeRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewAcessibilidadeRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		acessibilidade := &models.Acessibilidade{
			ID:   1,
			Nome: "Acessibilidade atualizada",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "acessibilidades" SET "nome"=$1 WHERE id = $2 AND "id" = $3`)).
			WithArgs(acessibilidade.Nome, acessibilidade.ID, acessibilidade.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, acessibilidade)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update with database error", func(t *testing.T) {
		acessibilidade := &models.Acessibilidade{
			ID:   1,
			Nome: "Test",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "acessibilidades"`)).
			WithArgs(acessibilidade.Nome, acessibilidade.ID, acessibilidade.ID).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Update(ctx, acessibilidade)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar acessibilidade")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAcessibilidadeRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewAcessibilidadeRepository(db)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "acessibilidades" WHERE "acessibilidades"."id" = $1`)).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, 1)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete with database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "acessibilidades"`)).
			WithArgs(1).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Delete(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir acessibilidade")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAcessibilidadeRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewAcessibilidadeRepository(db)
	ctx := context.Background()

	t.Run("successful list without filter", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "acessibilidades"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(2, "Audiodescrição").
			AddRow(1, "Libras")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "acessibilidades" ORDER BY id DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		acessibilidades, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, acessibilidades, 2)
		assert.Equal(t, "Audiodescrição", acessibilidades[0].Nome)
		assert.Equal(t, "Libras", acessibilidades[1].Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successful list with filter", func(t *testing.T) {
		filter := map[string]interface{}{
			"nome": "Libras",
		}

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "acessibilidades" WHERE nome = $1`)).
			WithArgs("Libras").
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Libras")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "acessibilidades" WHERE nome = $1 ORDER BY id DESC LIMIT $2`)).
			WithArgs("Libras", 10).
			WillReturnRows(rows)

		acessibilidades, total, err := repo.List(ctx, filter, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, acessibilidades, 1)
		assert.Equal(t, "Libras", acessibilidades[0].Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with pagination", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "acessibilidades"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(3, "Item 3").
			AddRow(2, "Item 2")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "acessibilidades" ORDER BY id DESC LIMIT $1 OFFSET $2`)).
			WithArgs(2, 2).
			WillReturnRows(rows)

		acessibilidades, total, err := repo.List(ctx, map[string]interface{}{}, 2, 2)

		assert.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, acessibilidades, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with database error on count", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "acessibilidades"`)).
			WillReturnError(errors.New("database error"))

		acessibilidades, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, acessibilidades)
		assert.Contains(t, err.Error(), "erro ao listar acessibilidades")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with database error on query", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "acessibilidades"`)).
			WillReturnRows(countRows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "acessibilidades"`)).
			WillReturnError(errors.New("database error"))

		acessibilidades, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, acessibilidades)
		assert.Contains(t, err.Error(), "erro ao listar acessibilidades")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
