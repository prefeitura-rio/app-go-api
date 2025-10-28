package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type JobService struct {
	repo *repository.JobRepository
}

func NewJobService(repo *repository.JobRepository) *JobService {
	return &JobService{
		repo: repo,
	}
}

func (s *JobService) Create(ctx context.Context, job *models.Job) error {
	if job.Type == "" {
		return fmt.Errorf("tipo de job é obrigatório")
	}

	job.Status = models.JobStatusPending
	job.Progress = 0
	job.SuccessCount = 0
	job.ErrorCount = 0

	return s.repo.Create(ctx, job)
}

func (s *JobService) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *JobService) Update(ctx context.Context, job *models.Job) error {
	existingJob, err := s.repo.GetByID(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("erro ao verificar job: %w", err)
	}
	if existingJob == nil {
		return fmt.Errorf("job não encontrado")
	}

	return s.repo.Update(ctx, job)
}

func (s *JobService) UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job não encontrado")
	}

	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *JobService) UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job não encontrado")
	}

	return s.repo.UpdateProgress(ctx, id, progress, successCount, errorCount)
}

func (s *JobService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Job, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}

func (s *JobService) Delete(ctx context.Context, id uuid.UUID) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job não encontrado")
	}

	return s.repo.Delete(ctx, id)
}
