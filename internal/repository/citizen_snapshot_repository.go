package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// CitizenSnapshotRepository handles database operations for citizen snapshots
type CitizenSnapshotRepository struct {
	db *gorm.DB
}

// NewCitizenSnapshotRepository creates a new repository instance
func NewCitizenSnapshotRepository(db *gorm.DB) *CitizenSnapshotRepository {
	return &CitizenSnapshotRepository{
		db: db,
	}
}

// GetByCPF fetches a snapshot by CPF
func (r *CitizenSnapshotRepository) GetByCPF(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	var snapshot models.CitizenSnapshot

	result := r.db.WithContext(ctx).Where("cpf = ?", cpf).First(&snapshot)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get citizen snapshot: %w", result.Error)
	}

	return &snapshot, nil
}

// GetByCPFs fetches multiple snapshots by CPFs (batch query)
func (r *CitizenSnapshotRepository) GetByCPFs(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
	if len(cpfs) == 0 {
		return make(map[string]*models.CitizenSnapshot), nil
	}

	var snapshots []models.CitizenSnapshot
	result := r.db.WithContext(ctx).Where("cpf IN ?", cpfs).Find(&snapshots)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get citizen snapshots: %w", result.Error)
	}

	// Build map for easy lookup
	snapshotMap := make(map[string]*models.CitizenSnapshot)
	for i := range snapshots {
		snapshotMap[snapshots[i].CPF] = &snapshots[i]
	}

	return snapshotMap, nil
}

// Upsert inserts or updates a snapshot (UPSERT operation)
func (r *CitizenSnapshotRepository) Upsert(ctx context.Context, snapshot *models.CitizenSnapshot) error {
	snapshot.LastSyncedAt = time.Now()

	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cpf"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"nome", "email", "celular", "data_nascimento", "endereco",
			"raca", "genero", "renda_familiar", "escolaridade", "deficiencia",
			"last_synced_at", "updated_at",
		}),
	}).Create(snapshot)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert citizen snapshot: %w", result.Error)
	}

	return nil
}

// BatchUpsert performs batch upsert for multiple snapshots
func (r *CitizenSnapshotRepository) BatchUpsert(ctx context.Context, snapshots []*models.CitizenSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	now := time.Now()
	for i := range snapshots {
		snapshots[i].LastSyncedAt = now
	}

	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cpf"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"nome", "email", "celular", "data_nascimento", "endereco",
			"raca", "genero", "renda_familiar", "escolaridade", "deficiencia",
			"last_synced_at", "updated_at",
		}),
	}).Create(&snapshots)

	if result.Error != nil {
		return fmt.Errorf("failed to batch upsert citizen snapshots: %w", result.Error)
	}

	return nil
}

// GetStaleSnapshots fetches snapshots older than the given threshold
func (r *CitizenSnapshotRepository) GetStaleSnapshots(ctx context.Context, staleThreshold time.Duration, limit int) ([]*models.CitizenSnapshot, error) {
	var snapshots []*models.CitizenSnapshot
	staleTime := time.Now().Add(-staleThreshold)

	result := r.db.WithContext(ctx).
		Where("last_synced_at < ?", staleTime).
		Order("last_synced_at ASC").
		Limit(limit).
		Find(&snapshots)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get stale snapshots: %w", result.Error)
	}

	return snapshots, nil
}

// GetCPFsWithEnrollments returns CPFs that have enrollments but no snapshot or stale snapshot
func (r *CitizenSnapshotRepository) GetCPFsWithEnrollments(ctx context.Context, staleThreshold time.Duration, limit int) ([]string, error) {
	var cpfs []string
	staleTime := time.Now().Add(-staleThreshold)

	// Find CPFs from inscricoes that either:
	// 1. Don't have a citizen_snapshot record
	// 2. Have a stale citizen_snapshot record
	query := `
		SELECT DISTINCT i.cpf 
		FROM inscricoes i
		LEFT JOIN citizen_snapshots cs ON i.cpf = cs.cpf
		WHERE cs.cpf IS NULL OR cs.last_synced_at < $1
		LIMIT $2
	`

	result := r.db.WithContext(ctx).Raw(query, staleTime, limit).Scan(&cpfs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get CPFs with enrollments: %w", result.Error)
	}

	return cpfs, nil
}

// Delete removes a snapshot by CPF
func (r *CitizenSnapshotRepository) Delete(ctx context.Context, cpf string) error {
	result := r.db.WithContext(ctx).Delete(&models.CitizenSnapshot{}, "cpf = ?", cpf)
	if result.Error != nil {
		return fmt.Errorf("failed to delete citizen snapshot: %w", result.Error)
	}
	return nil
}

// Count returns the total number of snapshots
func (r *CitizenSnapshotRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.CitizenSnapshot{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count citizen snapshots: %w", result.Error)
	}
	return count, nil
}
