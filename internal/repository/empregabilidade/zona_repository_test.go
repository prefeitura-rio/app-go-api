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

func TestZonaRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Zona{
		ID:        uuid.New(),
		Descricao: "Zona Norte",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_zonas"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(entity.ID))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, entity)

	assert.NoError(t, err)
	assert.Equal(t, entity.ID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_Create_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Zona{
		ID:        uuid.New(),
		Descricao: "Zona Sul",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_zonas"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	id, err := repo.Create(ctx, entity)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "erro ao criar zona")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_GetByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	id := uuid.New()
	expectedEntity := &empregabilidade.Zona{
		ID:        id,
		Descricao: "Zona Oeste",
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"}).
		AddRow(expectedEntity.ID, expectedEntity.Descricao)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas"`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEntity.ID, result.ID)
	assert.Equal(t, expectedEntity.Descricao, result.Descricao)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas"`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_GetByID_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas"`)).
		WithArgs(id, 1).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetByID(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "erro ao buscar zona")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_Update(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Zona{
		ID:        uuid.New(),
		Descricao: "Centro",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_zonas"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), entity.ID, entity.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, entity)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_Update_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Zona{
		ID:        uuid.New(),
		Descricao: "Zona Norte",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_zonas"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), entity.ID, entity.ID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Update(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar zona")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_zonas"`)).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, id)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_Delete_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_zonas"`)).
		WithArgs(id).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir zona")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_List(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	expectedEntities := []*empregabilidade.Zona{
		{ID: uuid.New(), Descricao: "Centro"},
		{ID: uuid.New(), Descricao: "Zona Norte"},
		{ID: uuid.New(), Descricao: "Zona Oeste"},
		{ID: uuid.New(), Descricao: "Zona Sul"},
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"})
	for _, e := range expectedEntities {
		rows.AddRow(e.ID, e.Descricao)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_zonas"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas"`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 4, total)
	assert.Len(t, results, 4)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_List_Empty(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_zonas"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "descricao"}))

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, results, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_List_WithFilter(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{
		"descricao": "Centro",
	}

	expectedEntity := &empregabilidade.Zona{
		ID:        uuid.New(),
		Descricao: "Centro",
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"}).
		AddRow(expectedEntity.ID, expectedEntity.Descricao)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_zonas" WHERE descricao = $1`)).
		WithArgs("Centro").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas" WHERE descricao = $1 ORDER BY descricao ASC LIMIT`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, filter, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, results, 1)
	assert.Equal(t, expectedEntity.Descricao, results[0].Descricao)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_List_WithPagination(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	limit := 2
	offset := 1

	allEntities := []*empregabilidade.Zona{
		{ID: uuid.New(), Descricao: "Centro"},
		{ID: uuid.New(), Descricao: "Zona Norte"},
		{ID: uuid.New(), Descricao: "Zona Oeste"},
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"}).
		AddRow(allEntities[1].ID, allEntities[1].Descricao).
		AddRow(allEntities[2].ID, allEntities[2].Descricao)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_zonas"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas" ORDER BY descricao ASC LIMIT`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, map[string]interface{}{}, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZonaRepository_List_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewZonaRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_zonas"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_zonas"`)).
		WillReturnError(errors.New("database error"))

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao listar zonas")
	assert.Equal(t, 0, total)
	assert.Nil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}
