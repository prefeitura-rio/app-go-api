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

func TestOrgaoSnapshotRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		snapshot := &models.OrgaoSnapshot{
			OrgaoID:      "org-123",
			Name:         "Test Organization",
			SyncStatus:   models.SyncStatusSynced,
			LastSyncedAt: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orgao_snapshots"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(ctx, snapshot)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		snapshot := &models.OrgaoSnapshot{
			OrgaoID:      "org-123",
			Name:         "Test Organization",
			SyncStatus:   models.SyncStatusSynced,
			LastSyncedAt: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orgao_snapshots"`)).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Create(ctx, snapshot)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create orgao snapshot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_GetByOrgaoID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		orgaoID := "org-123"
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "orgao_id", "name", "sigla", "metadata", "last_synced_at",
			"sync_status", "sync_error", "created_at", "updated_at",
		}).AddRow(
			1, orgaoID, "Test Org", "TO", nil, now,
			models.SyncStatusSynced, nil, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE orgao_id = $1 ORDER BY "orgao_snapshots"."id" LIMIT $2`)).
			WithArgs(orgaoID, 1).
			WillReturnRows(rows)

		snapshot, err := repo.GetByOrgaoID(ctx, orgaoID)

		require.NoError(t, err)
		require.NotNil(t, snapshot)
		assert.Equal(t, orgaoID, snapshot.OrgaoID)
		assert.Equal(t, "Test Org", snapshot.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		orgaoID := "org-999"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE orgao_id = $1 ORDER BY "orgao_snapshots"."id" LIMIT $2`)).
			WithArgs(orgaoID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		snapshot, err := repo.GetByOrgaoID(ctx, orgaoID)

		require.NoError(t, err)
		assert.Nil(t, snapshot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		orgaoID := "org-123"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE orgao_id = $1 ORDER BY "orgao_snapshots"."id" LIMIT $2`)).
			WithArgs(orgaoID, 1).
			WillReturnError(sql.ErrConnDone)

		snapshot, err := repo.GetByOrgaoID(ctx, orgaoID)

		require.Error(t, err)
		assert.Nil(t, snapshot)
		assert.Contains(t, err.Error(), "failed to get orgao snapshot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_GetByOrgaoIDs(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		orgaoIDs := []string{"org-1", "org-2"}
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "orgao_id", "name", "sigla", "metadata", "last_synced_at",
			"sync_status", "sync_error", "created_at", "updated_at",
		}).AddRow(
			1, "org-1", "Org One", "O1", nil, now,
			models.SyncStatusSynced, nil, now, now,
		).AddRow(
			2, "org-2", "Org Two", "O2", nil, now,
			models.SyncStatusSynced, nil, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE orgao_id IN ($1,$2)`)).
			WithArgs("org-1", "org-2").
			WillReturnRows(rows)

		snapshotMap, err := repo.GetByOrgaoIDs(ctx, orgaoIDs)

		require.NoError(t, err)
		require.NotNil(t, snapshotMap)
		assert.Len(t, snapshotMap, 2)
		assert.Equal(t, "Org One", snapshotMap["org-1"].Name)
		assert.Equal(t, "Org Two", snapshotMap["org-2"].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		orgaoIDs := []string{"org-1"}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE orgao_id IN ($1)`)).
			WithArgs("org-1").
			WillReturnError(sql.ErrConnDone)

		snapshotMap, err := repo.GetByOrgaoIDs(ctx, orgaoIDs)

		require.Error(t, err)
		assert.Nil(t, snapshotMap)
		assert.Contains(t, err.Error(), "failed to get orgao snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		snapshot := &models.OrgaoSnapshot{
			ID:           1,
			OrgaoID:      "org-123",
			Name:         "Updated Org",
			SyncStatus:   models.SyncStatusSynced,
			LastSyncedAt: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orgao_snapshots"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, snapshot)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		snapshot := &models.OrgaoSnapshot{
			ID:      1,
			OrgaoID: "org-123",
			Name:    "Updated Org",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orgao_snapshots"`)).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Update(ctx, snapshot)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update orgao snapshot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_Upsert(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		snapshot := &models.OrgaoSnapshot{
			OrgaoID:      "org-123",
			Name:         "Test Org",
			SyncStatus:   models.SyncStatusSynced,
			LastSyncedAt: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orgao_snapshots"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Upsert(ctx, snapshot)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		snapshot := &models.OrgaoSnapshot{
			OrgaoID:    "org-123",
			Name:       "Test Org",
			SyncStatus: models.SyncStatusSynced,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orgao_snapshots"`)).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Upsert(ctx, snapshot)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upsert orgao snapshot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_BatchUpsert(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		snapshots := []*models.OrgaoSnapshot{
			{OrgaoID: "org-1", Name: "Org One", SyncStatus: models.SyncStatusSynced, LastSyncedAt: now},
			{OrgaoID: "org-2", Name: "Org Two", SyncStatus: models.SyncStatusSynced, LastSyncedAt: now},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orgao_snapshots"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
		mock.ExpectCommit()

		err := repo.BatchUpsert(ctx, snapshots)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptySnapshots", func(t *testing.T) {
		snapshots := []*models.OrgaoSnapshot{}

		err := repo.BatchUpsert(ctx, snapshots)

		require.NoError(t, err)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		snapshots := []*models.OrgaoSnapshot{
			{OrgaoID: "org-1", Name: "Org One", SyncStatus: models.SyncStatusSynced},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orgao_snapshots"`)).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.BatchUpsert(ctx, snapshots)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to batch upsert orgao snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_GetStaleSnapshots(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		staleThreshold := 24 * time.Hour
		oldTime := time.Now().Add(-48 * time.Hour)

		rows := sqlmock.NewRows([]string{
			"id", "orgao_id", "name", "sigla", "metadata", "last_synced_at",
			"sync_status", "sync_error", "created_at", "updated_at",
		}).AddRow(
			1, "org-stale", "Stale Org", "SO", nil, oldTime,
			models.SyncStatusSynced, nil, oldTime, oldTime,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE last_synced_at < $1`)).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		snapshots, err := repo.GetStaleSnapshots(ctx, staleThreshold)

		require.NoError(t, err)
		require.NotNil(t, snapshots)
		assert.Len(t, snapshots, 1)
		assert.Equal(t, "org-stale", snapshots[0].OrgaoID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE last_synced_at < $1`)).
			WillReturnError(sql.ErrConnDone)

		snapshots, err := repo.GetStaleSnapshots(ctx, 24*time.Hour)

		require.Error(t, err)
		assert.Nil(t, snapshots)
		assert.Contains(t, err.Error(), "failed to get stale snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_GetFailedSnapshots(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		errMsg := "sync error"

		rows := sqlmock.NewRows([]string{
			"id", "orgao_id", "name", "sigla", "metadata", "last_synced_at",
			"sync_status", "sync_error", "created_at", "updated_at",
		}).AddRow(
			1, "org-failed", "Failed Org", "FO", nil, now,
			models.SyncStatusFailed, &errMsg, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusFailed).
			WillReturnRows(rows)

		snapshots, err := repo.GetFailedSnapshots(ctx)

		require.NoError(t, err)
		require.NotNil(t, snapshots)
		assert.Len(t, snapshots, 1)
		assert.Equal(t, "org-failed", snapshots[0].OrgaoID)
		assert.Equal(t, models.SyncStatusFailed, snapshots[0].SyncStatus)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusFailed).
			WillReturnError(sql.ErrConnDone)

		snapshots, err := repo.GetFailedSnapshots(ctx)

		require.Error(t, err)
		assert.Nil(t, snapshots)
		assert.Contains(t, err.Error(), "failed to get failed snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_GetPendingSnapshots(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "orgao_id", "name", "sigla", "metadata", "last_synced_at",
			"sync_status", "sync_error", "created_at", "updated_at",
		}).AddRow(
			1, "org-pending", "Pending Org", "PO", nil, now,
			models.SyncStatusPending, nil, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusPending).
			WillReturnRows(rows)

		snapshots, err := repo.GetPendingSnapshots(ctx)

		require.NoError(t, err)
		require.NotNil(t, snapshots)
		assert.Len(t, snapshots, 1)
		assert.Equal(t, "org-pending", snapshots[0].OrgaoID)
		assert.Equal(t, models.SyncStatusPending, snapshots[0].SyncStatus)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusPending).
			WillReturnError(sql.ErrConnDone)

		snapshots, err := repo.GetPendingSnapshots(ctx)

		require.Error(t, err)
		assert.Nil(t, snapshots)
		assert.Contains(t, err.Error(), "failed to get pending snapshots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOrgaoSnapshotRepository_CountByStatus(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOrgaoSnapshotRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		syncedRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
		failedRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		pendingRows := sqlmock.NewRows([]string{"count"}).AddRow(5)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusSynced).
			WillReturnRows(syncedRows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusFailed).
			WillReturnRows(failedRows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusPending).
			WillReturnRows(pendingRows)

		counts, err := repo.CountByStatus(ctx)

		require.NoError(t, err)
		require.NotNil(t, counts)
		assert.Equal(t, int64(10), counts[models.SyncStatusSynced])
		assert.Equal(t, int64(3), counts[models.SyncStatusFailed])
		assert.Equal(t, int64(5), counts[models.SyncStatusPending])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orgao_snapshots" WHERE sync_status = $1`)).
			WithArgs(models.SyncStatusSynced).
			WillReturnError(sql.ErrConnDone)

		counts, err := repo.CountByStatus(ctx)

		require.Error(t, err)
		assert.Nil(t, counts)
		assert.Contains(t, err.Error(), "failed to count snapshots by status")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
