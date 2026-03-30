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

func TestCategoriaRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCategoriaRepository(db)
	ctx := context.Background()

	t.Run("successful create", func(t *testing.T) {
		categoria := &models.Categoria{
			Nome: "Tecnologia",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categorias" ("nome") VALUES ($1) RETURNING "id"`)).
			WithArgs(categoria.Nome).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, categoria)

		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.Equal(t, 1, categoria.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create with database error", func(t *testing.T) {
		categoria := &models.Categoria{
			Nome: "Test Categoria",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categorias"`)).
			WithArgs(categoria.Nome).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		id, err := repo.Create(ctx, categoria)

		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "erro ao criar categoria")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCategoriaRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCategoriaRepository(db)
	ctx := context.Background()

	t.Run("successful get by id", func(t *testing.T) {
		expectedCategoria := &models.Categoria{
			ID:   1,
			Nome: "Desenvolvimento Web",
		}

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(expectedCategoria.ID, expectedCategoria.Nome)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias" WHERE "categorias"."id" = $1 ORDER BY "categorias"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnRows(rows)

		categoria, err := repo.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, categoria)
		assert.Equal(t, expectedCategoria.ID, categoria.ID)
		assert.Equal(t, expectedCategoria.Nome, categoria.Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("categoria not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias" WHERE "categorias"."id" = $1 ORDER BY "categorias"."id" LIMIT $2`)).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}))

		categoria, err := repo.GetByID(ctx, 999)

		assert.NoError(t, err)
		assert.Nil(t, categoria)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias" WHERE "categorias"."id" = $1 ORDER BY "categorias"."id" LIMIT $2`)).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		categoria, err := repo.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, categoria)
		assert.Contains(t, err.Error(), "erro ao buscar categoria por ID")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCategoriaRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCategoriaRepository(db)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		categoria := &models.Categoria{
			ID:   1,
			Nome: "Categoria atualizada",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "categorias" SET "nome"=$1 WHERE id = $2 AND "id" = $3`)).
			WithArgs(categoria.Nome, categoria.ID, categoria.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, categoria)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update with database error", func(t *testing.T) {
		categoria := &models.Categoria{
			ID:   1,
			Nome: "Test",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "categorias"`)).
			WithArgs(categoria.Nome, categoria.ID, categoria.ID).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Update(ctx, categoria)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar categoria")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCategoriaRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCategoriaRepository(db)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "categorias" WHERE "categorias"."id" = $1`)).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, 1)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete with database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "categorias"`)).
			WithArgs(1).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Delete(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir categoria")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCategoriaRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCategoriaRepository(db)
	ctx := context.Background()

	t.Run("successful list without filter", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categorias"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(2, "Design").
			AddRow(1, "Programação")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias" ORDER BY id DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		categorias, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, categorias, 2)
		assert.Equal(t, "Design", categorias[0].Nome)
		assert.Equal(t, "Programação", categorias[1].Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successful list with simple filter", func(t *testing.T) {
		filter := map[string]interface{}{
			"nome": "Tecnologia",
		}

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categorias" WHERE nome = $1`)).
			WithArgs("Tecnologia").
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Tecnologia")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias" WHERE nome = $1 ORDER BY id DESC LIMIT $2`)).
			WithArgs("Tecnologia", 10).
			WillReturnRows(rows)

		categorias, total, err := repo.List(ctx, filter, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, categorias, 1)
		assert.Equal(t, "Tecnologia", categorias[0].Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with empty result", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categorias"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nome"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias" ORDER BY id DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		categorias, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Len(t, categorias, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with pagination", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categorias"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(3, "Item 3").
			AddRow(2, "Item 2")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias" ORDER BY id DESC LIMIT $1 OFFSET $2`)).
			WithArgs(2, 2).
			WillReturnRows(rows)

		categorias, total, err := repo.List(ctx, map[string]interface{}{}, 2, 2)

		assert.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, categorias, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with database error on count", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categorias"`)).
			WillReturnError(errors.New("database error"))

		categorias, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, categorias)
		assert.Contains(t, err.Error(), "erro ao listar categorias")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with database error on query", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "categorias"`)).
			WillReturnRows(countRows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "categorias"`)).
			WillReturnError(errors.New("database error"))

		categorias, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, categorias)
		assert.Contains(t, err.Error(), "erro ao listar categorias")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
