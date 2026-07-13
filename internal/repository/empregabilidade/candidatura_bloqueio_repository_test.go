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
)

func TestCandidaturaBloqueioRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.CandidaturaBloqueio{
		ID:                    uuid.New(),
		CPF:                   "12345678900",
		IDVaga:                uuid.New(),
		CriteriosNaoAtendidos: []string{"idade_minima"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidatura_bloqueios"`)).
		WithArgs(entity.CPF, entity.IDVaga, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(entity.ID))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, entity)

	assert.NoError(t, err)
	assert.Equal(t, entity.ID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_Create_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.CandidaturaBloqueio{
		ID:                    uuid.New(),
		CPF:                   "12345678900",
		IDVaga:                uuid.New(),
		CriteriosNaoAtendidos: []string{"idade_minima"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_candidatura_bloqueios"`)).
		WithArgs(entity.CPF, entity.IDVaga, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.ID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	id, err := repo.Create(ctx, entity)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "erro ao criar bloqueio de candidatura")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_List(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)
	ctx := context.Background()

	expectedEntities := []*empregabilidade.CandidaturaBloqueio{
		{ID: uuid.New(), CPF: "12345678900", IDVaga: uuid.New()},
		{ID: uuid.New(), CPF: "98765432100", IDVaga: uuid.New()},
	}

	rows := sqlmock.NewRows([]string{"id", "cpf", "id_vaga"})
	for _, e := range expectedEntities {
		rows.AddRow(e.ID, e.CPF, e.IDVaga)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidatura_bloqueios"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidatura_bloqueios"`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_List_Empty(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidatura_bloqueios"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidatura_bloqueios"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "cpf", "id_vaga"}))

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, results, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_List_WithFilter(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	filter := map[string]interface{}{
		"id_vaga": vagaID,
		"cpf":     "12345678900",
	}

	expectedEntity := &empregabilidade.CandidaturaBloqueio{
		ID:     uuid.New(),
		CPF:    "12345678900",
		IDVaga: vagaID,
	}

	rows := sqlmock.NewRows([]string{"id", "cpf", "id_vaga"}).
		AddRow(expectedEntity.ID, expectedEntity.CPF, expectedEntity.IDVaga)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidatura_bloqueios" WHERE id_vaga = $1 AND cpf = $2`)).
		WithArgs(vagaID, "12345678900").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidatura_bloqueios" WHERE id_vaga = $1 AND cpf = $2`)).
		WillReturnRows(rows)

	results, total, err := repo.List(ctx, filter, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, results, 1)
	assert.Equal(t, expectedEntity.CPF, results[0].CPF)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCandidaturaBloqueioRepository_List_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaBloqueioRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_candidatura_bloqueios"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_candidatura_bloqueios"`)).
		WillReturnError(errors.New("database error"))

	results, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao listar bloqueios de candidatura")
	assert.Equal(t, 0, total)
	assert.Nil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}
