package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	result := r.db.WithContext(ctx).Create(job)
	if result.Error != nil {
		return fmt.Errorf("erro ao criar job: %w", result.Error)
	}
	return nil
}

func (r *JobRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	var job models.Job

	result := r.db.WithContext(ctx).First(&job, "id = ?", id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar job por ID: %w", result.Error)
	}

	return &job, nil
}

func (r *JobRepository) Update(ctx context.Context, job *models.Job) error {
	job.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).Save(job)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar job: %w", result.Error)
	}
	return nil
}

func (r *JobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == models.JobStatusCompleted || status == models.JobStatusFailed {
		now := time.Now()
		updates["completed_at"] = &now
	}

	result := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar status do job: %w", result.Error)
	}

	return nil
}

func (r *JobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error {
	updates := map[string]interface{}{
		"progress":      progress,
		"success_count": successCount,
		"error_count":   errorCount,
		"updated_at":    time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar progresso do job: %w", result.Error)
	}

	return nil
}

// Touch renews updated_at of a job that is still processing. It is the
// heartbeat that tells FailStale the job is alive in some pod.
func (r *JobRepository) Touch(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("id = ? AND status = ?", id, models.JobStatusProcessing).
		UpdateColumn("updated_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("erro ao renovar job: %w", result.Error)
	}

	return nil
}

// FailStale marks as failed, in a single conditional UPDATE, every pending or
// processing job not updated since cutoff (its pod stopped before finishing),
// recording jobErrors as the reason. It is safe to run from several pods at
// once and returns how many jobs it closed.
func (r *JobRepository) FailStale(ctx context.Context, cutoff time.Time, jobErrors datatypes.JSON) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Job{}).
		Where("status IN ?", []models.JobStatus{models.JobStatusPending, models.JobStatusProcessing}).
		Where("updated_at < ?", cutoff).
		Updates(map[string]interface{}{
			"status":       models.JobStatusFailed,
			"errors":       jobErrors,
			"completed_at": &now,
			"updated_at":   now,
		})

	if result.Error != nil {
		return 0, fmt.Errorf("erro ao encerrar jobs interrompidos: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func (r *JobRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error) {
	var jobs []*models.Job
	var total int64

	baseQuery := r.db.WithContext(ctx).Model(&models.Job{})

	// Apply filters
	for key, value := range filter {
		switch key {
		case "status":
			baseQuery = baseQuery.Where("status = ?", value)
		case "type":
			baseQuery = baseQuery.Where("type = ?", value)
		}
	}

	// Count total records
	baseQuery.Count(&total)

	// Get paginated results
	result := baseQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar jobs: %w", result.Error)
	}

	return jobs, int(total), nil
}

func (r *JobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Job{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao deletar job: %w", result.Error)
	}
	return nil
}
