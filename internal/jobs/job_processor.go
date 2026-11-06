package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/observability"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// defaultHeartbeatInterval is how often a running job renews updated_at when
// SetHeartbeatInterval is not called.
const defaultHeartbeatInterval = time.Minute

type JobProcessor struct {
	db                *gorm.DB
	jobService        *services.JobService
	inscricaoService  *services.InscricaoService
	cursoService      *services.CursoService
	activeJobs        map[uuid.UUID]context.CancelFunc
	mu                sync.RWMutex
	heartbeatInterval time.Duration
	// processors runs each job type; tests replace an entry to control a run.
	processors map[models.JobType]func(context.Context, *models.Job) error
}

func NewJobProcessor(
	db *gorm.DB,
	jobService *services.JobService,
	inscricaoService *services.InscricaoService,
	cursoService *services.CursoService,
) *JobProcessor {
	p := &JobProcessor{
		db:                db,
		jobService:        jobService,
		inscricaoService:  inscricaoService,
		cursoService:      cursoService,
		activeJobs:        make(map[uuid.UUID]context.CancelFunc),
		heartbeatInterval: defaultHeartbeatInterval,
	}
	p.processors = map[models.JobType]func(context.Context, *models.Job) error{
		models.JobTypeEnrollmentImport: p.processEnrollmentImport,
	}
	return p
}

// SetHeartbeatInterval sets how often a running job renews updated_at. Call it
// before starting jobs; a non-positive value keeps the current interval.
func (p *JobProcessor) SetHeartbeatInterval(interval time.Duration) {
	if interval > 0 {
		p.heartbeatInterval = interval
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
	job.Status = models.JobStatusProcessing

	// Renew the job while it runs, so the stale job reaper of any pod can tell
	// it apart from a job whose pod stopped in the middle.
	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		p.heartbeat(heartbeatCtx, jobID, p.heartbeatInterval)
	}()
	defer func() {
		stopHeartbeat()
		<-heartbeatDone
	}()

	var processingErr error

	if process, ok := p.processors[job.Type]; ok {
		processingErr = process(ctx, job)
	} else {
		processingErr = fmt.Errorf("tipo de job desconhecido: %s", job.Type)
	}

	if processingErr != nil {
		log.Printf("Erro ao processar job %s: %v", jobID, processingErr)
		observability.ObserveImportJobs(observability.ImportJobFailed, 1)
		if err := p.jobService.UpdateStatus(ctx, jobID, models.JobStatusFailed); err != nil {
			log.Printf("Erro ao atualizar status do job %s para failed: %v", jobID, err)
		}
		return processingErr
	}

	if err := p.jobService.UpdateStatus(ctx, jobID, models.JobStatusCompleted); err != nil {
		log.Printf("Erro ao atualizar status do job %s para completed: %v", jobID, err)
		return err
	}
	observability.ObserveImportJobs(observability.ImportJobCompleted, 1)

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

// heartbeat renews the job every interval until ctx is cancelled.
func (p *JobProcessor) heartbeat(ctx context.Context, jobID uuid.UUID, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.jobService.Heartbeat(ctx, jobID); err != nil && ctx.Err() == nil {
				log.Printf("Erro ao renovar job %s: %v", jobID, err)
			}
		}
	}
}

// ReapStaleJobs closes as failed the pending or processing jobs not renewed
// within staleAfter: their pod stopped before finishing (deploy, restart, OOM,
// scale down), so they would otherwise stay "processing" forever.
func (p *JobProcessor) ReapStaleJobs(ctx context.Context, staleAfter time.Duration) (int64, error) {
	n, err := p.jobService.FailStaleJobs(ctx, staleAfter)
	if err == nil {
		observability.ObserveImportJobs(observability.ImportJobReaped, n)
	}
	return n, err
}

// RunStaleJobReaper runs ReapStaleJobs at startup and then every interval
// until ctx is cancelled. Every pod may run it: the update is conditional.
func (p *JobProcessor) RunStaleJobReaper(ctx context.Context, interval, staleAfter time.Duration) {
	if interval <= 0 || staleAfter <= 0 {
		log.Printf("[JobReaper] desativado: intervalo (%s) e limite (%s) devem ser positivos", interval, staleAfter)
		return
	}

	reap := func() {
		n, err := p.ReapStaleJobs(ctx, staleAfter)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("[JobReaper] erro ao encerrar jobs interrompidos: %v", err)
			}
			return
		}
		if n > 0 {
			log.Printf("[JobReaper] %d job(s) interrompido(s) marcado(s) como failed", n)
		}
	}

	reap()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reap()
		}
	}
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
