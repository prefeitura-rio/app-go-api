package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// DEV-119: o PUT do curso não pode apagar turma que tem inscrição, porque
// inscricoes.schedule_id não tem chave estrangeira e a inscrição ficaria órfã.

const enrollmentCountSQL = `SELECT schedule_id, COUNT(*) AS count FROM "inscricoes"`

// expectUpdatePreamble registra as consultas de categorias, acessibilidades e
// custom fields que o Update faz antes das turmas.
func expectUpdatePreamble(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cursos"`)).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "custom_fields"`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
}

func expectNoRemoteClass(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id" FROM "remote_schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "remote_classes"`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
}

func TestCursoRepository_Update_SchedulesWithEnrollments(t *testing.T) {
	ctx := context.Background()

	t.Run("remover turma remota com inscrição devolve 409 sem apagar nada", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)

		remoteID, keepID, enrolledID := uuid.New(), uuid.New(), uuid.New()
		curso := &models.Curso{ID: 1, RemoteClass: &models.RemoteClass{
			ID: remoteID, CursoID: 1,
			Schedules: []models.RemoteSchedule{{ID: keepID, Vacancies: 20}},
		}}

		expectUpdatePreamble(mock)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "curso_id"}).AddRow(remoteID, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(keepID).AddRow(enrolledID))
		mock.ExpectQuery(regexp.QuoteMeta(enrollmentCountSQL)).
			WithArgs(enrolledID).
			WillReturnRows(sqlmock.NewRows([]string{"schedule_id", "count"}).AddRow(enrolledID, 3))
		mock.ExpectRollback()

		err := repo.Update(ctx, curso)

		require.Error(t, err)
		appErr, ok := apperrors.As(err)
		require.True(t, ok, "esperava erro de domínio, veio %v", err)
		assert.Equal(t, apperrors.KindConflict, appErr.Kind)
		assert.Contains(t, appErr.Message, enrolledID.String())
		assert.Contains(t, appErr.Message, "3 inscrição(ões)")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("tirar a turma remota inteira com inscrição devolve 409", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)

		enrolledID := uuid.New()
		curso := &models.Curso{ID: 1, RemoteClass: nil}

		expectUpdatePreamble(mock)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id" FROM "remote_schedules" WHERE remote_class_id IN (SELECT "id" FROM "remote_classes" WHERE curso_id = $1)`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(enrolledID))
		mock.ExpectQuery(regexp.QuoteMeta(enrollmentCountSQL)).
			WithArgs(enrolledID).
			WillReturnRows(sqlmock.NewRows([]string{"schedule_id", "count"}).AddRow(enrolledID, 1))
		mock.ExpectRollback()

		err := repo.Update(ctx, curso)

		assert.True(t, apperrors.IsKind(err, apperrors.KindConflict), "esperava 409, veio %v", err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("remover local cuja turma tem inscrição devolve 409 antes de qualquer escrita", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)

		keepLocationID, removedLocationID := uuid.New(), uuid.New()
		keepScheduleID, enrolledID := uuid.New(), uuid.New()
		curso := &models.Curso{ID: 1, LocationClasses: []models.LocationClass{{
			ID: keepLocationID, CursoID: 1, Address: "Rua Sintética",
			Schedules: []models.CourseSchedule{{ID: keepScheduleID, Vacancies: 25}},
		}}}

		expectUpdatePreamble(mock)
		expectNoRemoteClass(mock)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(keepLocationID).AddRow(removedLocationID))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules" WHERE location_id IN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "location_id"}).
				AddRow(keepScheduleID, keepLocationID).
				AddRow(enrolledID, removedLocationID))
		mock.ExpectQuery(regexp.QuoteMeta(enrollmentCountSQL)).
			WithArgs(enrolledID).
			WillReturnRows(sqlmock.NewRows([]string{"schedule_id", "count"}).AddRow(enrolledID, 2))
		mock.ExpectRollback()

		err := repo.Update(ctx, curso)

		assert.True(t, apperrors.IsKind(err, apperrors.KindConflict), "esperava 409, veio %v", err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("turma que muda de local é atualizada com o novo location_id", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)

		keepLocationID, removedLocationID := uuid.New(), uuid.New()
		movedID := uuid.New()
		curso := &models.Curso{ID: 1, LocationClasses: []models.LocationClass{{
			ID: keepLocationID, CursoID: 1, Address: "Rua Sintética",
			Schedules: []models.CourseSchedule{{ID: movedID, Vacancies: 25}},
		}}}

		expectUpdatePreamble(mock)
		expectNoRemoteClass(mock)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(keepLocationID).AddRow(removedLocationID))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "course_schedules" WHERE location_id IN`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "location_id"}).AddRow(movedID, removedLocationID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "location_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "course_schedules" SET "location_id"=$1`)).
			WithArgs(keepLocationID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), movedID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Only the emptied location is deleted; the moved turma is not.
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "location_classes" WHERE id IN ($1)`)).
			WithArgs(removedLocationID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		require.NoError(t, repo.Update(ctx, curso))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("id de turma que não é do curso vira turma nova, sem alterar a outra", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)

		remoteID, ownID, foreignID := uuid.New(), uuid.New(), uuid.New()
		curso := &models.Curso{ID: 1, RemoteClass: &models.RemoteClass{
			ID: remoteID, CursoID: 1,
			Schedules: []models.RemoteSchedule{{ID: ownID, Vacancies: 20}, {ID: foreignID, Vacancies: 999}},
		}}

		expectUpdatePreamble(mock)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "curso_id"}).AddRow(remoteID, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ownID))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_schedules"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "location_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cursos"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		require.NoError(t, repo.Update(ctx, curso))
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NotEqual(t, foreignID, curso.RemoteClass.Schedules[1].ID, "o id de outra turma não pode ser reaproveitado")
	})

	t.Run("falha ao contar inscrições aborta a transação", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()
		repo := NewCursoRepository(db)

		remoteID, obsoleteID := uuid.New(), uuid.New()
		curso := &models.Curso{ID: 1, RemoteClass: &models.RemoteClass{ID: remoteID, CursoID: 1}}

		expectUpdatePreamble(mock)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_classes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "curso_id"}).AddRow(remoteID, 1))
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "remote_classes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "remote_schedules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(obsoleteID))
		mock.ExpectQuery(regexp.QuoteMeta(enrollmentCountSQL)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Update(ctx, curso)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao verificar inscrições das turmas removidas")
		_, isDomain := apperrors.As(err)
		assert.False(t, isDomain, "falha de banco não é erro de domínio")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
