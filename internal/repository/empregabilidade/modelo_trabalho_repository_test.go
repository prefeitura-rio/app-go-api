package empregabilidade

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestModeloTrabalhoRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.ModeloTrabalho{
		ID:        uuid.New(),
		Descricao: "Presencial",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_modelos_trabalho"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(entity.ID))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, entity)

	assert.NoError(t, err)
	assert.Equal(t, entity.ID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_GetByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	id := uuid.New()
	expectedEntity := &empregabilidade.ModeloTrabalho{
		ID:        id,
		Descricao: "Híbrido",
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"}).
		AddRow(expectedEntity.ID, expectedEntity.Descricao)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_modelos_trabalho"`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEntity.ID, result.ID)
	assert.Equal(t, expectedEntity.Descricao, result.Descricao)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_modelos_trabalho"`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_Update(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.ModeloTrabalho{
		ID:        uuid.New(),
		Descricao: "Remoto",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_modelos_trabalho"`)).
		WithArgs(entity.Descricao, sqlmock.AnyArg(), entity.ID, entity.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, entity)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_modelos_trabalho"`)).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, id)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_List(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	expectedEntities := []*empregabilidade.ModeloTrabalho{
		{ID: uuid.New(), Descricao: "Híbrido"},
		{ID: uuid.New(), Descricao: "Presencial"},
	}

	rows := sqlmock.NewRows([]string{"id", "descricao"})
	for _, e := range expectedEntities {
		rows.AddRow(e.ID, e.Descricao)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_modelos_trabalho"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_modelos_trabalho"`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Error path tests

func TestModeloTrabalhoRepository_Create_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.ModeloTrabalho{
		ID:        uuid.New(),
		Descricao: "Presencial",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_modelos_trabalho"`)).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err := repo.Create(ctx, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao criar modelo de trabalho")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()
	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_modelos_trabalho"`)).
		WillReturnError(assert.AnError)

	result, err := repo.GetByID(ctx, id)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "erro ao buscar modelo de trabalho")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_Update_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.ModeloTrabalho{
		ID:        uuid.New(),
		Descricao: "Remoto",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_modelos_trabalho"`)).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Update(ctx, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar modelo de trabalho")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()
	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_modelos_trabalho"`)).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Delete(ctx, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir modelo de trabalho")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModeloTrabalhoRepository_List_FindError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewModeloTrabalhoRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_modelos_trabalho"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_modelos_trabalho"`)).
		WillReturnError(assert.AnError)

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Equal(t, 0, total)
	assert.Contains(t, err.Error(), "erro ao listar modelos de trabalho")
	assert.NoError(t, mock.ExpectationsWereMet())
}
