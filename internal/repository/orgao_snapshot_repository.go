package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// OrgaoSnapshotRepository handles database operations for orgao snapshots
type OrgaoSnapshotRepository struct {
	db *gorm.DB
}

// NewOrgaoSnapshotRepository creates a new repository instance
func NewOrgaoSnapshotRepository(db *gorm.DB) *OrgaoSnapshotRepository {
	return &OrgaoSnapshotRepository{
		db: db,
	}
}

// Create creates a new orgao snapshot
func (r *OrgaoSnapshotRepository) Create(ctx context.Context, snapshot *models.OrgaoSnapshot) error {
	result := r.db.WithContext(ctx).Create(snapshot)
	if result.Error != nil {
		return fmt.Errorf("failed to create orgao snapshot: %w", result.Error)
	}
	return nil
}

// GetByOrgaoID fetches a snapshot by orgao_id
func (r *OrgaoSnapshotRepository) GetByOrgaoID(ctx context.Context, orgaoID string) (*models.OrgaoSnapshot, error) {
	var snapshot models.OrgaoSnapshot

	result := r.db.WithContext(ctx).Where("orgao_id = ?", orgaoID).First(&snapshot)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get orgao snapshot: %w", result.Error)
	}

	return &snapshot, nil
}

// GetByOrgaoIDs fetches snapshots for multiple orgao IDs in a single query.
// Returns a map from orgao_id to snapshot; missing IDs are absent from the map.
func (r *OrgaoSnapshotRepository) GetByOrgaoIDs(ctx context.Context, orgaoIDs []string) (map[string]*models.OrgaoSnapshot, error) {
	var snapshots []*models.OrgaoSnapshot

	result := r.db.WithContext(ctx).Where("orgao_id IN ?", orgaoIDs).Find(&snapshots)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get orgao snapshots: %w", result.Error)
	}

	snapshotMap := make(map[string]*models.OrgaoSnapshot, len(snapshots))
	for _, s := range snapshots {
		snapshotMap[s.OrgaoID] = s
	}
	return snapshotMap, nil
}

// Update updates an existing orgao snapshot
func (r *OrgaoSnapshotRepository) Update(ctx context.Context, snapshot *models.OrgaoSnapshot) error {
	result := r.db.WithContext(ctx).Model(snapshot).
		Where("id = ?", snapshot.ID).
		Updates(snapshot)

	if result.Error != nil {
		return fmt.Errorf("failed to update orgao snapshot: %w", result.Error)
	}

	return nil
}

// Upsert inserts or updates a snapshot (UPSERT operation)
func (r *OrgaoSnapshotRepository) Upsert(ctx context.Context, snapshot *models.OrgaoSnapshot) error {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "orgao_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"cd_ua", "name", "sigla", "metadata", "last_synced_at", "sync_status", "sync_error", "updated_at"}),
	}).Create(snapshot)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert orgao snapshot: %w", result.Error)
	}

	return nil
}

// BatchUpsert performs batch upsert for multiple snapshots
func (r *OrgaoSnapshotRepository) BatchUpsert(ctx context.Context, snapshots []*models.OrgaoSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "orgao_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"cd_ua", "name", "sigla", "metadata", "last_synced_at", "sync_status", "sync_error", "updated_at"}),
	}).Create(&snapshots)

	if result.Error != nil {
		return fmt.Errorf("failed to batch upsert orgao snapshots: %w", result.Error)
	}

	return nil
}

// GetOrgaoIDsByCdUAs returns orgao_ids for the given cd_ua values.
// Used to resolve CPF-Secretaria cd_uas into the orgao_ids stored in resources.
func (r *OrgaoSnapshotRepository) GetOrgaoIDsByCdUAs(ctx context.Context, cdUAs []string) ([]string, error) {
	if len(cdUAs) == 0 {
		return []string{}, nil
	}

	var orgaoIDs []string
	result := r.db.WithContext(ctx).
		Model(&models.OrgaoSnapshot{}).
		Where("cd_ua IN ?", cdUAs).
		Pluck("orgao_id", &orgaoIDs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get orgao IDs by cd_uas: %w", result.Error)
	}

	return orgaoIDs, nil
}

// GetStaleSnapshots fetches snapshots older than the given threshold
func (r *OrgaoSnapshotRepository) GetStaleSnapshots(ctx context.Context, staleThreshold time.Duration) ([]*models.OrgaoSnapshot, error) {
	var snapshots []*models.OrgaoSnapshot
	staleTime := time.Now().Add(-staleThreshold)

	result := r.db.WithContext(ctx).
		Where("last_synced_at < ?", staleTime).
		Find(&snapshots)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get stale snapshots: %w", result.Error)
	}

	return snapshots, nil
}

// GetFailedSnapshots fetches all snapshots with failed sync status
func (r *OrgaoSnapshotRepository) GetFailedSnapshots(ctx context.Context) ([]*models.OrgaoSnapshot, error) {
	var snapshots []*models.OrgaoSnapshot

	result := r.db.WithContext(ctx).
		Where("sync_status = ?", models.SyncStatusFailed).
		Find(&snapshots)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get failed snapshots: %w", result.Error)
	}

	return snapshots, nil
}

// GetPendingSnapshots fetches all snapshots with pending sync status
func (r *OrgaoSnapshotRepository) GetPendingSnapshots(ctx context.Context) ([]*models.OrgaoSnapshot, error) {
	var snapshots []*models.OrgaoSnapshot

	result := r.db.WithContext(ctx).
		Where("sync_status = ?", models.SyncStatusPending).
		Find(&snapshots)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get pending snapshots: %w", result.Error)
	}

	return snapshots, nil
}

// CountByStatus returns count of snapshots by status
func (r *OrgaoSnapshotRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	counts := make(map[string]int64)

	statuses := []string{models.SyncStatusSynced, models.SyncStatusFailed, models.SyncStatusPending}

	for _, status := range statuses {
		var count int64
		result := r.db.WithContext(ctx).
			Model(&models.OrgaoSnapshot{}).
			Where("sync_status = ?", status).
			Count(&count)

		if result.Error != nil {
			return nil, fmt.Errorf("failed to count snapshots by status: %w", result.Error)
		}

		counts[status] = count
	}

	return counts, nil
}
