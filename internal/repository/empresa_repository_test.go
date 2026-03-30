package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestEmpresaRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &models.Empresa{
		Nome: "Empresa Teste LTDA",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "empresas"`).
		WithArgs(
			sqlmock.AnyArg(), // nome
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, empresa)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Create_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &models.Empresa{
		Nome: "Empresa Teste",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "empresas"`).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	id, err := repo.Create(ctx, empresa)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "erro ao criar empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	expectedEmpresa := &models.Empresa{
		ID:   1,
		Nome: "Empresa Teste LTDA",
	}

	mock.ExpectQuery(`SELECT \* FROM "empresas" WHERE "empresas"\."id" = \$1 ORDER BY "empresas"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Empresa Teste LTDA"))

	empresa, err := repo.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, empresa)
	assert.Equal(t, expectedEmpresa.ID, empresa.ID)
	assert.Equal(t, expectedEmpresa.Nome, empresa.Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "empresas" WHERE "empresas"\."id" = \$1 ORDER BY "empresas"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}))

	empresa, err := repo.GetByID(ctx, 999)

	assert.NoError(t, err)
	assert.Nil(t, empresa)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_GetByID_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "empresas" WHERE "empresas"\."id" = \$1 ORDER BY "empresas"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(errors.New("database error"))

	empresa, err := repo.GetByID(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, empresa)
	assert.Contains(t, err.Error(), "erro ao buscar empresa por ID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &models.Empresa{
		ID:   1,
		Nome: "Empresa Teste Atualizada LTDA",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "empresas" SET`).
		WithArgs(
			sqlmock.AnyArg(), // nome
			1,                // id in WHERE clause (from .Where("id = ?"))
			1,                // id in WHERE clause (from GORM's model ID)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, empresa)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Update_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	empresa := &models.Empresa{
		ID:   1,
		Nome: "Empresa Teste",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "empresas" SET`).
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	err := repo.Update(ctx, empresa)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "empresas" WHERE "empresas"\."id" = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_Delete_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "empresas"`).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir empresa")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 10
	offset := 0

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock list query
	mock.ExpectQuery(`SELECT \* FROM "empresas" ORDER BY id DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Empresa A").
			AddRow(2, "Empresa B").
			AddRow(3, "Empresa C"))

	empresas, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, empresas, 3)
	assert.Equal(t, "Empresa A", empresas[0].Nome)
	assert.Equal(t, "Empresa B", empresas[1].Nome)
	assert.Equal(t, "Empresa C", empresas[2].Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_WithFilter(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{
		"nome": "Empresa Filtrada",
	}
	limit := 10
	offset := 0

	// Mock count query with filter
	mock.ExpectQuery(`SELECT count\(\*\) FROM "empresas" WHERE nome = \$1`).
		WithArgs("Empresa Filtrada").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock list query with filter
	mock.ExpectQuery(`SELECT \* FROM "empresas" WHERE nome = \$1 ORDER BY id DESC LIMIT \$2`).
		WithArgs("Empresa Filtrada", limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(5, "Empresa Filtrada"))

	empresas, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, empresas, 1)
	assert.Equal(t, "Empresa Filtrada", empresas[0].Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_WithPagination(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 2
	offset := 1

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock list query with pagination
	mock.ExpectQuery(`SELECT \* FROM "empresas" ORDER BY id DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(limit, offset).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(2, "Empresa B").
			AddRow(3, "Empresa C"))

	empresas, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, empresas, 2)
	assert.Equal(t, "Empresa B", empresas[0].Nome)
	assert.Equal(t, "Empresa C", empresas[1].Nome)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpresaRepository_List_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpresaRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 10
	offset := 0

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "empresas"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock list query error
	mock.ExpectQuery(`SELECT \* FROM "empresas"`).
		WillReturnError(errors.New("query failed"))

	empresas, total, err := repo.List(ctx, filter, limit, offset)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao listar empresas")
	assert.Nil(t, empresas)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}
