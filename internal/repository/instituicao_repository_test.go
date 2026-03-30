package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestInstituicaoRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	instituicao := &models.InstituicaoEnsino{
		Nome: "Universidade Federal do Rio de Janeiro",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "instituicoes_ensino"`).
		WithArgs(
			sqlmock.AnyArg(), // nome
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, instituicao)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_Create_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	instituicao := &models.InstituicaoEnsino{
		Nome: "Universidade Teste",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "instituicoes_ensino"`).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	id, err := repo.Create(ctx, instituicao)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "erro ao criar instituição de ensino")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	expectedInstituicao := &models.InstituicaoEnsino{
		ID:   1,
		Nome: "Universidade Federal do Rio de Janeiro",
	}

	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino" WHERE "instituicoes_ensino"\."id" = \$1 ORDER BY "instituicoes_ensino"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Universidade Federal do Rio de Janeiro"))

	instituicao, err := repo.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, instituicao)
	assert.Equal(t, expectedInstituicao.ID, instituicao.ID)
	assert.Equal(t, expectedInstituicao.Nome, instituicao.Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino" WHERE "instituicoes_ensino"\."id" = \$1 ORDER BY "instituicoes_ensino"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}))

	instituicao, err := repo.GetByID(ctx, 999)

	assert.NoError(t, err)
	assert.Nil(t, instituicao)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_GetByID_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino" WHERE "instituicoes_ensino"\."id" = \$1 ORDER BY "instituicoes_ensino"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(errors.New("database error"))

	instituicao, err := repo.GetByID(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, instituicao)
	assert.Contains(t, err.Error(), "erro ao buscar instituição de ensino por ID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	instituicao := &models.InstituicaoEnsino{
		ID:   1,
		Nome: "Universidade Federal do Rio de Janeiro - UFRJ",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "instituicoes_ensino" SET`).
		WithArgs(
			sqlmock.AnyArg(), // nome
			1,                // id in WHERE clause (from .Where("id = ?"))
			1,                // id in WHERE clause (from GORM's model ID)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, instituicao)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_Update_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	instituicao := &models.InstituicaoEnsino{
		ID:   1,
		Nome: "Universidade Teste",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "instituicoes_ensino" SET`).
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	err := repo.Update(ctx, instituicao)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar instituição de ensino")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "instituicoes_ensino" WHERE "instituicoes_ensino"\."id" = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_Delete_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "instituicoes_ensino"`).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir instituição de ensino")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 10
	offset := 0

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "instituicoes_ensino"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock list query
	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino" ORDER BY id DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "UFRJ").
			AddRow(2, "UFF").
			AddRow(3, "UERJ"))

	instituicoes, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, instituicoes, 3)
	assert.Equal(t, "UFRJ", instituicoes[0].Nome)
	assert.Equal(t, "UFF", instituicoes[1].Nome)
	assert.Equal(t, "UERJ", instituicoes[2].Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_List_WithFilter(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{
		"nome": "UFRJ",
	}
	limit := 10
	offset := 0

	// Mock count query with filter
	mock.ExpectQuery(`SELECT count\(\*\) FROM "instituicoes_ensino" WHERE nome = \$1`).
		WithArgs("UFRJ").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock list query with filter
	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino" WHERE nome = \$1 ORDER BY id DESC LIMIT \$2`).
		WithArgs("UFRJ", limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "UFRJ"))

	instituicoes, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, instituicoes, 1)
	assert.Equal(t, "UFRJ", instituicoes[0].Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_List_WithPagination(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 2
	offset := 1

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "instituicoes_ensino"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock list query with pagination
	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino" ORDER BY id DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(limit, offset).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(2, "UFF").
			AddRow(3, "UERJ"))

	instituicoes, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, instituicoes, 2)
	assert.Equal(t, "UFF", instituicoes[0].Nome)
	assert.Equal(t, "UERJ", instituicoes[1].Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_List_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 10
	offset := 0

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "instituicoes_ensino"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock list query error
	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino"`).
		WillReturnError(errors.New("query failed"))

	instituicoes, total, err := repo.List(ctx, filter, limit, offset)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao listar instituições de ensino")
	assert.Nil(t, instituicoes)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInstituicaoRepository_List_EmptyResult(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInstituicaoRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 10
	offset := 0

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "instituicoes_ensino"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock list query - empty result
	mock.ExpectQuery(`SELECT \* FROM "instituicoes_ensino" ORDER BY id DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}))

	instituicoes, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, instituicoes, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
