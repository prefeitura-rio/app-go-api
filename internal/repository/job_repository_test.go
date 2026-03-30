package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestJobRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		job := &models.Job{
			ID:           uuid.New(),
			Type:         "enrollment_import",
			Status:       models.JobStatusPending,
			Progress:     0,
			TotalRecords: 100,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(job.ID))
		mock.ExpectCommit()

		err := repo.Create(ctx, job)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		job := &models.Job{
			ID:     uuid.New(),
			Type:   "enrollment_import",
			Status: models.JobStatusPending,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Create(ctx, job)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar job")
	})
}

func TestJobRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "type", "status", "progress", "total_records"}).
			AddRow(id, "enrollment_import", models.JobStatusCompleted, 100, 100)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"`)).
			WillReturnRows(rows)

		job, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, id, job.ID)
		assert.Equal(t, models.JobStatusCompleted, job.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		job, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, job)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"`)).
			WillReturnError(assert.AnError)

		job, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, job)
		assert.Contains(t, err.Error(), "erro ao buscar job por ID")
	})
}

func TestJobRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		job := &models.Job{
			ID:       uuid.New(),
			Type:     "enrollment_import",
			Status:   models.JobStatusProcessing,
			Progress: 50,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, job)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		job := &models.Job{
			ID:     uuid.New(),
			Status: models.JobStatusProcessing,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Update(ctx, job)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar job")
	})
}

func TestJobRepository_UpdateStatus(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	ctx := context.Background()

	t.Run("update status to processing", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStatus(ctx, id, models.JobStatusProcessing)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update status to completed", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStatus(ctx, id, models.JobStatusCompleted)
		assert.NoError(t, err)
	})

	t.Run("update status to failed", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStatus(ctx, id, models.JobStatusFailed)
		assert.NoError(t, err)
	})

	t.Run("update status error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateStatus(ctx, id, models.JobStatusCompleted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar status do job")
	})
}

func TestJobRepository_UpdateProgress(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	ctx := context.Background()

	t.Run("update progress success", func(t *testing.T) {
		id := uuid.New()
		progress := 75
		successCount := 75
		errorCount := 5

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateProgress(ctx, id, progress, successCount, errorCount)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update progress error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateProgress(ctx, id, 50, 50, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar progresso do job")
	})
}

func TestJobRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	ctx := context.Background()

	t.Run("list with no filters", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "jobs"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		rows := sqlmock.NewRows([]string{"id", "type", "status", "progress"}).
			AddRow(uuid.New(), "enrollment_import", models.JobStatusCompleted, 100).
			AddRow(uuid.New(), "enrollment_import", models.JobStatusProcessing, 50).
			AddRow(uuid.New(), "certificate_generation", models.JobStatusPending, 0)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"`)).
			WillReturnRows(rows)

		jobs, total, err := repo.List(ctx, map[string]interface{}{}, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, jobs, 3)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with status filter", func(t *testing.T) {
		filter := map[string]interface{}{
			"status": models.JobStatusCompleted,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "jobs"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "type", "status", "progress"}).
			AddRow(uuid.New(), "enrollment_import", models.JobStatusCompleted, 100)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"`)).
			WillReturnRows(rows)

		jobs, total, err := repo.List(ctx, filter, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, jobs, 1)
	})

	t.Run("list with type filter", func(t *testing.T) {
		filter := map[string]interface{}{
			"type": "enrollment_import",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "jobs"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "type", "status"}).
			AddRow(uuid.New(), "enrollment_import", models.JobStatusCompleted).
			AddRow(uuid.New(), "enrollment_import", models.JobStatusProcessing)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"`)).
			WillReturnRows(rows)

		jobs, total, err := repo.List(ctx, filter, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, jobs, 2)
	})
}

func TestJobRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "jobs"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "jobs"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao deletar job")
	})
}
