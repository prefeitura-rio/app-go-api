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

func TestInscricaoRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		inscricao := &models.Inscricao{
			ID:      uuid.New(),
			CursoID: 1,
			CPF:     "12345678901",
			Name:    "João Silva",
			Email:   "joao@example.com",
			Status:  models.StatusInscricaoPending,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inscricoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(inscricao.ID))
		mock.ExpectCommit()

		err := repo.Create(ctx, inscricao)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		inscricao := &models.Inscricao{
			ID:      uuid.New(),
			CursoID: 1,
			CPF:     "12345678901",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inscricoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Create(ctx, inscricao)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar inscrição")
	})
}

func TestInscricaoRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "curso_id", "cpf", "name", "email", "status"}).
			AddRow(id, 1, "12345678901", "João Silva", "joao@example.com", "pending")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "inscricoes"`)).
			WillReturnRows(rows)

		// Expect curso preload query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		inscricao, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, inscricao)
		assert.Equal(t, id, inscricao.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "inscricoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		inscricao, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, inscricao)
	})
}

func TestInscricaoRepository_UpdateStatus(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("update status success", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inscricoes"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStatus(ctx, id, models.StatusInscricaoApproved, "", "")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update status with reason", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inscricoes"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStatus(ctx, id, models.StatusInscricaoRejected, "Não atende requisitos", "")
		assert.NoError(t, err)
	})

	t.Run("update status not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inscricoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateStatus(ctx, id, models.StatusInscricaoApproved, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "não encontrada")
	})
}

func TestInscricaoRepository_GetByCursoID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("list by curso with filters", func(t *testing.T) {
		cursoID := 1
		filter := map[string]interface{}{
			"status": models.StatusInscricaoPending,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inscricoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		rows := sqlmock.NewRows([]string{"id", "curso_id", "cpf", "status"}).
			AddRow(uuid.New(), cursoID, "12345678901", models.StatusInscricaoPending).
			AddRow(uuid.New(), cursoID, "98765432109", models.StatusInscricaoPending)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "inscricoes"`)).
			WillReturnRows(rows)

		inscricoes, total, err := repo.GetByCursoID(ctx, cursoID, filter, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, inscricoes, 2)
	})
}

func TestInscricaoRepository_ExistsByCPFAndCurso(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("exists returns true", func(t *testing.T) {
		cpf := "12345678901"
		cursoID := 1

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inscricoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		exists, err := repo.ExistsByCPFAndCurso(ctx, cpf, cursoID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("exists returns false", func(t *testing.T) {
		cpf := "12345678901"
		cursoID := 1

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inscricoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := repo.ExistsByCPFAndCurso(ctx, cpf, cursoID)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestInscricaoRepository_UpdateCertificate(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("update certificate success", func(t *testing.T) {
		id := uuid.New()
		certURL := "https://example.com/cert.pdf"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inscricoes"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateCertificate(ctx, id, certURL)
		assert.NoError(t, err)
	})

	t.Run("update certificate not found", func(t *testing.T) {
		id := uuid.New()
		certURL := "https://example.com/cert.pdf"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inscricoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateCertificate(ctx, id, certURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "não encontrada")
	})
}

func TestInscricaoRepository_GetSummaryByCursoID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("get summary", func(t *testing.T) {
		cursoID := 1

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inscricoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		statusRows := sqlmock.NewRows([]string{"status", "count"}).
			AddRow("pending", 3).
			AddRow("approved", 5).
			AddRow("rejected", 2)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT status, count(*) as count FROM "inscricoes"`)).
			WillReturnRows(statusRows)

		summary, err := repo.GetSummaryByCursoID(ctx, cursoID)
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 10, summary.Total)
		assert.Equal(t, 3, summary.Pending)
		assert.Equal(t, 5, summary.Approved)
		assert.Equal(t, 2, summary.Rejected)
	})
}

func TestInscricaoRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		inscricao := &models.Inscricao{
			ID:      uuid.New(),
			CursoID: 1,
			CPF:     "12345678901",
			Name:    "João Updated",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inscricoes"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, inscricao)
		assert.NoError(t, err)
	})
}

func TestInscricaoRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewInscricaoRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "inscricoes"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
	})
}
