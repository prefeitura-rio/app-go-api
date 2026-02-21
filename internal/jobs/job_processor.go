package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type JobProcessor struct {
	db               *gorm.DB
	jobService       *services.JobService
	inscricaoService *services.InscricaoService
	cursoService     *services.CursoService
	activeJobs       map[uuid.UUID]context.CancelFunc
	mu               sync.RWMutex
}

func NewJobProcessor(
	db *gorm.DB,
	jobService *services.JobService,
	inscricaoService *services.InscricaoService,
	cursoService *services.CursoService,
) *JobProcessor {
	return &JobProcessor{
		db:               db,
		jobService:       jobService,
		inscricaoService: inscricaoService,
		cursoService:     cursoService,
		activeJobs:       make(map[uuid.UUID]context.CancelFunc),
	}
}

func (p *JobProcessor) ProcessJob(jobID uuid.UUID) error {
	ctx, cancel := context.WithCancel(context.Background())

	p.mu.Lock()
	p.activeJobs[jobID] = cancel
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.activeJobs, jobID)
		p.mu.Unlock()
	}()

	job, err := p.jobService.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("erro ao buscar job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job não encontrado")
	}

	// Update status to processing
	if err := p.jobService.UpdateStatus(ctx, jobID, models.JobStatusProcessing); err != nil {
		log.Printf("Erro ao atualizar status do job %s para processing: %v", jobID, err)
		return err
	}

	var processingErr error

	switch job.Type {
	case models.JobTypeEnrollmentImport:
		processingErr = p.processEnrollmentImport(ctx, job)
	default:
		processingErr = fmt.Errorf("tipo de job desconhecido: %s", job.Type)
	}

	if processingErr != nil {
		log.Printf("Erro ao processar job %s: %v", jobID, processingErr)
		if err := p.jobService.UpdateStatus(ctx, jobID, models.JobStatusFailed); err != nil {
			log.Printf("Erro ao atualizar status do job %s para failed: %v", jobID, err)
		}
		return processingErr
	}

	if err := p.jobService.UpdateStatus(ctx, jobID, models.JobStatusCompleted); err != nil {
		log.Printf("Erro ao atualizar status do job %s para completed: %v", jobID, err)
		return err
	}

	return nil
}

func (p *JobProcessor) CancelJob(jobID uuid.UUID) error {
	p.mu.RLock()
	cancel, exists := p.activeJobs[jobID]
	p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job não está em execução")
	}

	cancel()
	return nil
}

func (p *JobProcessor) processEnrollmentImport(ctx context.Context, job *models.Job) error {
	processor := NewEnrollmentImportProcessor(
		p.db,
		p.jobService,
		p.inscricaoService,
		p.cursoService,
	)
	return processor.Process(ctx, job)
}

// StartJob is a helper to start a job in a goroutine
func (p *JobProcessor) StartJob(jobID uuid.UUID) {
	go func() {
		if err := p.ProcessJob(jobID); err != nil {
			log.Printf("Job %s failed with error: %v", jobID, err)
		} else {
			log.Printf("Job %s completed successfully", jobID)
		}
	}()
}

// Global instance to be initialized in main
var GlobalJobProcessor *JobProcessor

func InitializeJobProcessor(
	db *gorm.DB,
	jobRepo *repository.JobRepository,
	inscricaoRepo *repository.InscricaoRepository,
	cursoRepo *repository.CursoRepository,
) {
	jobService := services.NewJobService(jobRepo)
	// Job processor doesn't need email notifications or citizen sync (bulk imports handle their own flow)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	GlobalJobProcessor = NewJobProcessor(db, jobService, inscricaoService, cursoService)
}
