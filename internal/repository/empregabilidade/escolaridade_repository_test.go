package empregabilidade

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEscolaridadeRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Escolaridade{
		ID:        uuid.New(),
		Descricao: "Superior Completo",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_escolaridades"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(entity.ID))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, entity)

	assert.NoError(t, err)
	assert.Equal(t, entity.ID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_Create_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Escolaridade{
		ID:        uuid.New(),
		Descricao: "Ensino Médio",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_escolaridades"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	id, err := repo.Create(ctx, entity)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "erro ao criar escolaridade")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_GetByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	id := uuid.New()
	expectedEntity := &empregabilidade.Escolaridade{
		ID:        id,
		Descricao: "Ensino Fundamental",
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"}).
		AddRow(expectedEntity.ID, expectedEntity.Descricao)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEntity.ID, result.ID)
	assert.Equal(t, expectedEntity.Descricao, result.Descricao)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_GetByID_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
		WithArgs(id, 1).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetByID(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "erro ao buscar escolaridade")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_Update(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Escolaridade{
		ID:        uuid.New(),
		Descricao: "Pós-Graduação",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_escolaridades"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), entity.ID, entity.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, entity)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_Update_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Escolaridade{
		ID:        uuid.New(),
		Descricao: "Mestrado",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_escolaridades"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), entity.ID, entity.ID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Update(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar escolaridade")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_escolaridades"`)).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, id)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_Delete_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_escolaridades"`)).
		WithArgs(id).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir escolaridade")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_List(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	expectedEntities := []*empregabilidade.Escolaridade{
		{ID: uuid.New(), Descricao: "Ensino Fundamental"},
		{ID: uuid.New(), Descricao: "Ensino Médio"},
		{ID: uuid.New(), Descricao: "Superior Completo"},
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"})
	for _, e := range expectedEntities {
		rows.AddRow(e.ID, e.Descricao)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_escolaridades"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, results, 3)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_List_Empty(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_escolaridades"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "descricao"}))

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, results, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_List_WithFilter(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{
		"descricao": "Superior Completo",
	}

	expectedEntity := &empregabilidade.Escolaridade{
		ID:        uuid.New(),
		Descricao: "Superior Completo",
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"}).
		AddRow(expectedEntity.ID, expectedEntity.Descricao)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_escolaridades" WHERE descricao = $1`)).
		WithArgs("Superior Completo").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades" WHERE descricao = $1 ORDER BY descricao ASC LIMIT`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, filter, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, results, 1)
	assert.Equal(t, expectedEntity.Descricao, results[0].Descricao)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_List_WithPagination(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	limit := 2
	offset := 1

	allEntities := []*empregabilidade.Escolaridade{
		{ID: uuid.New(), Descricao: "Ensino Fundamental"},
		{ID: uuid.New(), Descricao: "Ensino Médio"},
		{ID: uuid.New(), Descricao: "Superior Completo"},
	}

	// Return page 2 (offset 1, limit 2)
	rows := sqlmock.NewRows([]string{"id", "descricao"}).
		AddRow(allEntities[1].ID, allEntities[1].Descricao).
		AddRow(allEntities[2].ID, allEntities[2].Descricao)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_escolaridades"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades" ORDER BY descricao ASC LIMIT`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, map[string]interface{}{}, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEscolaridadeRepository_List_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEscolaridadeRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_escolaridades"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
		WillReturnError(errors.New("database error"))

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao listar escolaridades")
	assert.Equal(t, 0, total)
	assert.Nil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}
