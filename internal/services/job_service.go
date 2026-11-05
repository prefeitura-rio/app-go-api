package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// JobInterruptedMessage is the reason recorded in Job.Errors when a job is
// closed because the pod processing it stopped before finishing (deploy,
// restart, OOM, scale down). Rows already imported are not duplicated on a
// new upload: they come back as "CPF já inscrito neste curso".
const JobInterruptedMessage = "Processamento interrompido: o servidor foi reiniciado antes de concluir a importação. " +
	"Reenvie o arquivo; as linhas já importadas voltam como \"CPF já inscrito neste curso\", sem duplicar."

type JobService struct {
	repo JobRepositoryInterface
}

func NewJobService(repo *repository.JobRepository) *JobService {
	return &JobService{
		repo: repo,
	}
}

func NewJobServiceWithInterface(repo JobRepositoryInterface) *JobService {
	return &JobService{repo: repo}
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

	// Status (and CompletedAt) are managed exclusively via UpdateStatus.
	// Preserve DB values so a stale in-memory Job (e.g. still "pending" after
	// UpdateStatus("processing")) cannot overwrite the current status via Save.
	job.Status = existingJob.Status
	job.CompletedAt = existingJob.CompletedAt

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

// Heartbeat renews a processing job so FailStaleJobs knows it is still alive.
func (s *JobService) Heartbeat(ctx context.Context, id uuid.UUID) error {
	return s.repo.Touch(ctx, id)
}

// FailStaleJobs closes as failed, with JobInterruptedMessage, every pending or
// processing job not renewed within staleAfter, and returns how many it closed.
func (s *JobService) FailStaleJobs(ctx context.Context, staleAfter time.Duration) (int64, error) {
	if staleAfter <= 0 {
		return 0, fmt.Errorf("staleAfter deve ser positivo: %s", staleAfter)
	}

	jobErrors, err := json.Marshal([]models.JobError{{Line: 0, Message: JobInterruptedMessage}})
	if err != nil {
		return 0, fmt.Errorf("erro ao montar motivo da interrupção: %w", err)
	}

	return s.repo.FailStale(ctx, time.Now().Add(-staleAfter), datatypes.JSON(jobErrors))
}
