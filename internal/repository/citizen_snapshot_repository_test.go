package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestCitizenSnapshotRepository_GetByCPF(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		cpf := "12345678901"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"cpf", "nome", "nome_social", "email", "celular", "data_nascimento",
			"endereco", "raca", "genero", "renda_familiar", "escolaridade",
			"deficiencia", "last_synced_at", "created_at", "updated_at",
		}).AddRow(
			cpf, "John Doe", "Johnny", "john@example.com", "11999999999", now,
			nil, "Branca", "Masculino", "1-3 SM", "Superior", "", now, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_snapshots" WHERE cpf = $1 ORDER BY "citizen_snapshots"."cpf" LIMIT $2`)).
			WithArgs(cpf, 1).
			WillReturnRows(rows)

		snapshot, err := repo.GetByCPF(ctx, cpf)

		require.NoError(t, err)
		require.NotNil(t, snapshot)
		assert.Equal(t, cpf, snapshot.CPF)
		assert.Equal(t, "John Doe", snapshot.Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		cpf := "99999999999"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_snapshots" WHERE cpf = $1 ORDER BY "citizen_snapshots"."cpf" LIMIT $2`)).
			WithArgs(cpf, 1).
			WillReturnRows(sqlmock.NewRows([]string{"cpf"}))

		snapshot, err := repo.GetByCPF(ctx, cpf)

		require.NoError(t, err)
		assert.Nil(t, snapshot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		cpf := "12345678901"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_snapshots" WHERE cpf = $1 ORDER BY "citizen_snapshots"."cpf" LIMIT $2`)).
			WithArgs(cpf, 1).
			WillReturnError(sql.ErrConnDone)

		snapshot, err := repo.GetByCPF(ctx, cpf)

		require.Error(t, err)
		assert.Nil(t, snapshot)
		assert.Contains(t, err.Error(), "failed to get citizen snapshot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCitizenSnapshotRepository_GetByCPFs(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		cpfs := []string{"11111111111", "22222222222"}
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"cpf", "nome", "nome_social", "email", "celular", "data_nascimento",
			"endereco", "raca", "genero", "renda_familiar", "escolaridade",
			"deficiencia", "last_synced_at", "created_at", "updated_at",
		}).AddRow(
			"11111111111", "Alice", "", "alice@example.com", "", now,
			nil, "", "", "", "", "", now, now, now,
		).AddRow(
			"22222222222", "Bob", "", "bob@example.com", "", now,
			nil, "", "", "", "", "", now, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_snapshots" WHERE cpf IN ($1,$2)`)).
			WithArgs("11111111111", "22222222222").
			WillReturnRows(rows)

		snapshotMap, err := repo.GetByCPFs(ctx, cpfs)

		require.NoError(t, err)
		require.NotNil(t, snapshotMap)
		assert.Len(t, snapshotMap, 2)
		assert.Equal(t, "Alice", snapshotMap["11111111111"].Nome)
		assert.Equal(t, "Bob", snapshotMap["22222222222"].Nome)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptyCPFs", func(t *testing.T) {
		cpfs := []string{}

		snapshotMap, err := repo.GetByCPFs(ctx, cpfs)

		require.NoError(t, err)
		assert.NotNil(t, snapshotMap)
		assert.Empty(t, snapshotMap)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		cpfs := []string{"11111111111"}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_snapshots" WHERE cpf IN ($1)`)).
			WithArgs("11111111111").
			WillReturnError(sql.ErrConnDone)

		snapshotMap, err := repo.GetByCPFs(ctx, cpfs)

		require.Error(t, err)
		assert.Nil(t, snapshotMap)
		assert.Contains(t, err.Error(), "failed to get citizen snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCitizenSnapshotRepository_Upsert(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		snapshot := &models.CitizenSnapshot{
			CPF:   "12345678901",
			Nome:  "John Doe",
			Email: "john@example.com",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "citizen_snapshots"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Upsert(ctx, snapshot)

		require.NoError(t, err)
		assert.False(t, snapshot.LastSyncedAt.IsZero())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		snapshot := &models.CitizenSnapshot{
			CPF:  "12345678901",
			Nome: "John Doe",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "citizen_snapshots"`)).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Upsert(ctx, snapshot)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upsert citizen snapshot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCitizenSnapshotRepository_BatchUpsert(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		snapshots := []*models.CitizenSnapshot{
			{CPF: "11111111111", Nome: "Alice"},
			{CPF: "22222222222", Nome: "Bob"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "citizen_snapshots"`)).
			WillReturnResult(sqlmock.NewResult(1, 2))
		mock.ExpectCommit()

		err := repo.BatchUpsert(ctx, snapshots)

		require.NoError(t, err)
		assert.False(t, snapshots[0].LastSyncedAt.IsZero())
		assert.False(t, snapshots[1].LastSyncedAt.IsZero())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptySnapshots", func(t *testing.T) {
		snapshots := []*models.CitizenSnapshot{}

		err := repo.BatchUpsert(ctx, snapshots)

		require.NoError(t, err)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		snapshots := []*models.CitizenSnapshot{
			{CPF: "11111111111", Nome: "Alice"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "citizen_snapshots"`)).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.BatchUpsert(ctx, snapshots)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to batch upsert citizen snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCitizenSnapshotRepository_GetStaleSnapshots(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		staleThreshold := 24 * time.Hour
		limit := 10
		oldTime := time.Now().Add(-48 * time.Hour)

		rows := sqlmock.NewRows([]string{
			"cpf", "nome", "nome_social", "email", "celular", "data_nascimento",
			"endereco", "raca", "genero", "renda_familiar", "escolaridade",
			"deficiencia", "last_synced_at", "created_at", "updated_at",
		}).AddRow(
			"12345678901", "Stale User", "", "stale@example.com", "", oldTime,
			nil, "", "", "", "", "", oldTime, oldTime, oldTime,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_snapshots" WHERE last_synced_at < $1 ORDER BY last_synced_at ASC LIMIT $2`)).
			WithArgs(sqlmock.AnyArg(), limit).
			WillReturnRows(rows)

		snapshots, err := repo.GetStaleSnapshots(ctx, staleThreshold, limit)

		require.NoError(t, err)
		require.NotNil(t, snapshots)
		assert.Len(t, snapshots, 1)
		assert.Equal(t, "12345678901", snapshots[0].CPF)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "citizen_snapshots" WHERE last_synced_at < $1 ORDER BY last_synced_at ASC LIMIT $2`)).
			WithArgs(sqlmock.AnyArg(), 10).
			WillReturnError(sql.ErrConnDone)

		snapshots, err := repo.GetStaleSnapshots(ctx, 24*time.Hour, 10)

		require.Error(t, err)
		assert.Nil(t, snapshots)
		assert.Contains(t, err.Error(), "failed to get stale snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCitizenSnapshotRepository_GetCPFsWithEnrollments(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		staleThreshold := 24 * time.Hour
		limit := 5

		rows := sqlmock.NewRows([]string{"cpf"}).
			AddRow("11111111111").
			AddRow("22222222222")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT e.cpf FROM ( SELECT i.cpf AS cpf FROM inscricoes i UNION SELECT c.cpf AS cpf FROM emp_candidaturas c WHERE c.deleted_at IS NULL ) e LEFT JOIN citizen_snapshots cs ON e.cpf = cs.cpf WHERE cs.cpf IS NULL OR cs.last_synced_at < $1 LIMIT $2`)).
			WithArgs(sqlmock.AnyArg(), limit).
			WillReturnRows(rows)

		cpfs, err := repo.GetCPFsWithEnrollments(ctx, staleThreshold, limit)

		require.NoError(t, err)
		require.NotNil(t, cpfs)
		assert.Len(t, cpfs, 2)
		assert.Contains(t, cpfs, "11111111111")
		assert.Contains(t, cpfs, "22222222222")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT e.cpf FROM ( SELECT i.cpf AS cpf FROM inscricoes i UNION SELECT c.cpf AS cpf FROM emp_candidaturas c WHERE c.deleted_at IS NULL ) e LEFT JOIN citizen_snapshots cs ON e.cpf = cs.cpf WHERE cs.cpf IS NULL OR cs.last_synced_at < $1 LIMIT $2`)).
			WillReturnError(sql.ErrConnDone)

		cpfs, err := repo.GetCPFsWithEnrollments(ctx, 24*time.Hour, 5)

		require.Error(t, err)
		assert.Nil(t, cpfs)
		assert.Contains(t, err.Error(), "failed to get CPFs with enrollments")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCitizenSnapshotRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		cpf := "12345678901"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "citizen_snapshots" WHERE cpf = $1`)).
			WithArgs(cpf).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, cpf)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		cpf := "12345678901"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "citizen_snapshots" WHERE cpf = $1`)).
			WithArgs(cpf).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Delete(ctx, cpf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete citizen snapshot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCitizenSnapshotRepository_Count(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewCitizenSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expectedCount := int64(42)

		rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "citizen_snapshots"`)).
			WillReturnRows(rows)

		count, err := repo.Count(ctx)

		require.NoError(t, err)
		assert.Equal(t, expectedCount, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "citizen_snapshots"`)).
			WillReturnError(sql.ErrConnDone)

		count, err := repo.Count(ctx)

		require.Error(t, err)
		assert.Equal(t, int64(0), count)
		assert.Contains(t, err.Error(), "failed to count citizen snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
