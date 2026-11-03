package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

const scheduleLockSQL = `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`

// TestInscricaoRepository_WithScheduleLock_CountAndCreateInsideLockedTransaction
// proves the vacancy count and the insert of an auto-approved enrollment run in
// the same transaction that holds the turma lock (DEV-199): one BEGIN, the
// advisory lock, the COUNT, the INSERT without a transaction of its own, one
// COMMIT.
func TestInscricaoRepository_WithScheduleLock_CountAndCreateInsideLockedTransaction(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	inscricaoRepo := NewInscricaoRepository(db)
	cursoRepo := NewCursoRepository(db)
	scheduleID := uuid.New()
	inscricao := &models.Inscricao{
		ID:         uuid.New(),
		CursoID:    1,
		CPF:        "00000000191",
		ScheduleID: &scheduleID,
		Status:     models.StatusInscricaoApproved,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(scheduleLockSQL)).
		WithArgs("inscricoes:schedule:" + scheduleID.String()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inscricoes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inscricoes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(inscricao.ID))
	mock.ExpectCommit()

	err := inscricaoRepo.WithScheduleLock(context.Background(), scheduleID, func(ctx context.Context) error {
		count, err := cursoRepo.CountEnrollmentsByScheduleID(ctx, scheduleID)
		if err != nil {
			return err
		}
		assert.Equal(t, int64(2), count)
		return inscricaoRepo.Create(ctx, inscricao)
	})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestInscricaoRepository_WithScheduleLock_UpdateInsideLockedTransaction covers
// the turma change: the COUNT of the new turma and the UPDATE share the locked
// transaction.
func TestInscricaoRepository_WithScheduleLock_UpdateInsideLockedTransaction(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	inscricaoRepo := NewInscricaoRepository(db)
	cursoRepo := NewCursoRepository(db)
	scheduleID := uuid.New()
	inscricao := &models.Inscricao{ID: uuid.New(), CursoID: 1, ScheduleID: &scheduleID}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(scheduleLockSQL)).
		WithArgs("inscricoes:schedule:" + scheduleID.String()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT schedule_id, COUNT(*) as count FROM "inscricoes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"schedule_id", "count"}).AddRow(scheduleID, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inscricoes"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := inscricaoRepo.WithScheduleLock(context.Background(), scheduleID, func(ctx context.Context) error {
		counts, err := cursoRepo.CountEnrollmentsByScheduleIDs(ctx, []uuid.UUID{scheduleID})
		if err != nil {
			return err
		}
		assert.Equal(t, int64(1), counts[scheduleID])
		return inscricaoRepo.Update(ctx, inscricao)
	})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestInscricaoRepository_WithScheduleLock_LockError_RollsBackWithoutCallingFn
// makes sure nothing is written when the lock cannot be taken.
func TestInscricaoRepository_WithScheduleLock_LockError_RollsBackWithoutCallingFn(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	scheduleID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(scheduleLockSQL)).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	called := false
	err := repo.WithScheduleLock(context.Background(), scheduleID, func(context.Context) error {
		called = true
		return nil
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Contains(t, err.Error(), "erro ao bloquear turma")
	assert.False(t, called, "fn must not run without the lock")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestInscricaoRepository_WithScheduleLock_FnError_RollsBack makes sure an error
// returned by fn (e.g. turma full) undoes everything done under the lock.
func TestInscricaoRepository_WithScheduleLock_FnError_RollsBack(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	scheduleID := uuid.New()
	errFull := errors.New("não há vagas disponíveis nesta turma")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(scheduleLockSQL)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.WithScheduleLock(context.Background(), scheduleID, func(context.Context) error {
		return errFull
	})

	assert.ErrorIs(t, err, errFull)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestInscricaoRepository_Create_OutsideLock_KeepsOwnTransaction makes sure the
// context plumbing does not change Create when no lock is held.
func TestInscricaoRepository_Create_OutsideLock_KeepsOwnTransaction(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	inscricao := &models.Inscricao{ID: uuid.New(), CursoID: 1, CPF: "00000000191"}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inscricoes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(inscricao.ID))
	mock.ExpectCommit()

	require.NoError(t, repo.Create(context.Background(), inscricao))
	assert.NoError(t, mock.ExpectationsWereMet())
}
