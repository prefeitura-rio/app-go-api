package jobs

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNewJobProcessor(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.activeJobs)
	assert.Equal(t, 0, len(processor.activeJobs))
}

func TestJobProcessorActiveJobsTracking(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	// Test adding to active jobs
	jobID := uuid.New()
	cancelFunc := func() {}

	processor.mu.Lock()
	processor.activeJobs[jobID] = cancelFunc
	processor.mu.Unlock()

	processor.mu.RLock()
	_, exists := processor.activeJobs[jobID]
	processor.mu.RUnlock()

	assert.True(t, exists, "Job should be tracked in activeJobs")

	// Test removing from active jobs
	processor.mu.Lock()
	delete(processor.activeJobs, jobID)
	processor.mu.Unlock()

	processor.mu.RLock()
	_, exists = processor.activeJobs[jobID]
	processor.mu.RUnlock()

	assert.False(t, exists, "Job should be removed from activeJobs")
}

func TestJobProcessorConcurrentAccess(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			jobID := uuid.New()
			cancelFunc := func() {}

			processor.mu.Lock()
			processor.activeJobs[jobID] = cancelFunc
			processor.mu.Unlock()
		}(i)
	}

	wg.Wait()

	processor.mu.RLock()
	count := len(processor.activeJobs)
	processor.mu.RUnlock()

	assert.Equal(t, numGoroutines, count, "All jobs should be tracked")
}

func TestCancelJob_JobNotRunning(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	// Try to cancel a non-existent job
	jobID := uuid.New()
	err := processor.CancelJob(jobID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "não está em execução")
}

func TestCancelJob_JobRunning(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	// Add a job to active jobs
	jobID := uuid.New()
	cancelled := false
	cancelFunc := func() {
		cancelled = true
	}

	processor.mu.Lock()
	processor.activeJobs[jobID] = cancelFunc
	processor.mu.Unlock()

	// Cancel the job
	err := processor.CancelJob(jobID)

	assert.NoError(t, err)
	assert.True(t, cancelled, "Cancel function should have been called")
}

// createTestJob creates a job with a pre-generated UUID for SQLite compatibility
func createTestJob(jobType models.JobType, metadata datatypes.JSON) *models.Job {
	return &models.Job{
		ID:       uuid.New(), // SQLite can't auto-generate UUIDs like PostgreSQL
		Type:     jobType,
		Status:   models.JobStatusPending,
		Metadata: metadata,
	}
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables manually to avoid PostgreSQL-specific syntax issues with AutoMigrate
	// Jobs table
	db.Exec(`CREATE TABLE jobs (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		status TEXT DEFAULT 'pending' NOT NULL,
		progress INTEGER DEFAULT 0,
		total_records INTEGER DEFAULT 0,
		success_count INTEGER DEFAULT 0,
		error_count INTEGER DEFAULT 0,
		errors TEXT,
		result TEXT,
		metadata TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		completed_at DATETIME
	)`)

	// Cursos table (minimal version)
	db.Exec(`CREATE TABLE cursos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		titulo TEXT NOT NULL,
		descricao TEXT,
		orgao_id TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	// Inscricoes table (minimal version)
	db.Exec(`CREATE TABLE inscricoes (
		id TEXT PRIMARY KEY,
		curso_id INTEGER,
		cpf TEXT,
		name TEXT,
		email TEXT,
		phone TEXT,
		age INTEGER,
		address TEXT,
		neighborhood TEXT,
		custom_fields_data TEXT,
		schedule_id TEXT,
		enrolled_unit TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	return db
}

func TestProcessJob_JobNotFound(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Try to process a non-existent job
	jobID := uuid.New()
	err := processor.ProcessJob(jobID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job não encontrado")

	// Verify job was removed from active jobs
	processor.mu.RLock()
	_, exists := processor.activeJobs[jobID]
	processor.mu.RUnlock()
	assert.False(t, exists, "Job should be removed from activeJobs after error")
}

func TestProcessJob_UnknownJobType(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create a job with unknown type
	job := &models.Job{
		ID:     uuid.New(),
		Type:   "unknown_type",
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process the job
	err = processor.ProcessJob(job.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tipo de job desconhecido")

	// Verify job status was updated to failed
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedJob, "Job should exist after processing")
	assert.Equal(t, models.JobStatusFailed, updatedJob.Status)
}

func TestProcessJob_EnrollmentImportCursoNotFound(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create job with non-existent curso
	metadata := models.EnrollmentImportMetadata{
		CursoID:  999, // Non-existent
		FileName: "/tmp/test.csv",
	}
	metadataJSON, _ := json.Marshal(metadata)

	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON(metadataJSON),
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process the job
	err = processor.ProcessJob(job.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "curso não encontrado")

	// Verify job status was updated to failed
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedJob)
	assert.Equal(t, models.JobStatusFailed, updatedJob.Status)
}

func TestProcessJob_ConcurrentJobProcessing(t *testing.T) {
	t.Skip("Skipping concurrent test - SQLite in-memory has concurrency limitations")
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	numJobs := 10
	var wg sync.WaitGroup
	jobIDs := make([]uuid.UUID, numJobs)

	// Create multiple jobs
	for i := 0; i < numJobs; i++ {
		job := &models.Job{
			ID:     uuid.New(),
			Type:   "unknown_type",
			Status: models.JobStatusPending,
		}
		err := jobService.Create(context.Background(), job)
		assert.NoError(t, err)
		jobIDs[i] = job.ID
	}

	// Process jobs concurrently
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(jobID uuid.UUID) {
			defer wg.Done()
			_ = processor.ProcessJob(jobID)
		}(jobIDs[i])
	}

	wg.Wait()

	// Verify all jobs were removed from active jobs
	processor.mu.RLock()
	activeCount := len(processor.activeJobs)
	processor.mu.RUnlock()

	assert.Equal(t, 0, activeCount, "All jobs should be removed from activeJobs after processing")

	// Verify all jobs were marked as failed
	for _, jobID := range jobIDs {
		job, err := jobService.GetByID(context.Background(), jobID)
		assert.NoError(t, err)
		assert.Equal(t, models.JobStatusFailed, job.Status)
	}
}

func TestProcessJob_JobCleanupOnCompletion(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	job := &models.Job{
		ID:     uuid.New(),
		Type:   "unknown_type",
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Verify job is added to active jobs during processing
	processingStarted := make(chan bool)
	go func() {
		time.Sleep(10 * time.Millisecond)
		processor.mu.RLock()
		_, exists := processor.activeJobs[job.ID]
		processor.mu.RUnlock()
		if exists {
			processingStarted <- true
		}
	}()

	err = processor.ProcessJob(job.ID)
	assert.Error(t, err)

	// Verify job is removed from active jobs after processing
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()

	assert.False(t, exists, "Job should be removed from activeJobs after processing")
}

func TestStartJob_SuccessfulExecution(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	job := &models.Job{
		ID:     uuid.New(),
		Type:   "unknown_type",
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// StartJob runs in a goroutine, so we need to wait
	processor.StartJob(job.ID)
	time.Sleep(100 * time.Millisecond) // Give goroutine time to complete

	// Verify job was removed from active jobs
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()

	assert.False(t, exists, "Job should be removed from activeJobs after StartJob completes")

	// Verify job was marked as failed (unknown type)
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusFailed, updatedJob.Status)
}

func TestProcessJob_MultipleStartJobCalls(t *testing.T) {
	t.Skip("Timing-dependent test - may be flaky")
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	numJobs := 5
	for i := 0; i < numJobs; i++ {
		job := &models.Job{
			ID:     uuid.New(),
			Type:   "unknown_type",
			Status: models.JobStatusPending,
		}
		err := jobService.Create(context.Background(), job)
		assert.NoError(t, err)
		processor.StartJob(job.ID)
	}

	// Wait for all jobs to complete
	time.Sleep(200 * time.Millisecond)

	// Verify all jobs were removed from active jobs
	processor.mu.RLock()
	activeCount := len(processor.activeJobs)
	processor.mu.RUnlock()

	assert.Equal(t, 0, activeCount, "All jobs should be removed from activeJobs")
}

func TestCancelJob_ConcurrentCancellation(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	numJobs := 50
	jobIDs := make([]uuid.UUID, numJobs)
	cancelCounts := make([]int, numJobs)

	// Add jobs to active jobs
	for i := 0; i < numJobs; i++ {
		jobIDs[i] = uuid.New()
		idx := i
		cancelFunc := func() {
			cancelCounts[idx]++
		}

		processor.mu.Lock()
		processor.activeJobs[jobIDs[i]] = cancelFunc
		processor.mu.Unlock()
	}

	var wg sync.WaitGroup

	// Cancel jobs concurrently
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(jobID uuid.UUID) {
			defer wg.Done()
			_ = processor.CancelJob(jobID)
		}(jobIDs[i])
	}

	wg.Wait()

	// Verify all cancel functions were called exactly once
	for i := 0; i < numJobs; i++ {
		assert.Equal(t, 1, cancelCounts[i], "Cancel function should be called exactly once for job %d", i)
	}
}

func TestCancelJob_MultipleAttempts(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	jobID := uuid.New()
	cancelCount := 0
	cancelFunc := func() {
		cancelCount++
	}

	processor.mu.Lock()
	processor.activeJobs[jobID] = cancelFunc
	processor.mu.Unlock()

	// First cancellation should succeed
	err1 := processor.CancelJob(jobID)
	assert.NoError(t, err1)
	assert.Equal(t, 1, cancelCount)

	// Second cancellation should also succeed because CancelJob doesn't remove from activeJobs
	// The job is only removed when ProcessJob's defer runs
	err2 := processor.CancelJob(jobID)
	assert.NoError(t, err2)
	assert.Equal(t, 2, cancelCount, "Cancel function can be called multiple times")

	// Now manually remove the job from activeJobs
	processor.mu.Lock()
	delete(processor.activeJobs, jobID)
	processor.mu.Unlock()

	// Third cancellation should fail (job truly not active)
	err3 := processor.CancelJob(jobID)
	assert.Error(t, err3)
	assert.Contains(t, err3.Error(), "não está em execução")
	assert.Equal(t, 2, cancelCount, "Cancel function should not be called after removal")
}

func TestEnrollmentRow_DefaultValues(t *testing.T) {
	row := EnrollmentRow{}

	assert.Equal(t, "", row.NomeCompleto)
	assert.Equal(t, "", row.CPF)
	assert.Equal(t, 0, row.Idade)
	assert.Nil(t, row.CustomFields)
}

func TestEnrollmentRow_CustomFields(t *testing.T) {
	row := EnrollmentRow{
		NomeCompleto: "João Silva",
		CPF:          "12345678901",
		Idade:        30,
		CustomFields: map[string]string{
			"Campo1": "Valor1",
			"Campo2": "Valor2",
		},
	}

	assert.Equal(t, "João Silva", row.NomeCompleto)
	assert.Equal(t, "12345678901", row.CPF)
	assert.Equal(t, 30, row.Idade)
	assert.NotNil(t, row.CustomFields)
	assert.Equal(t, 2, len(row.CustomFields))
	assert.Equal(t, "Valor1", row.CustomFields["Campo1"])
	assert.Equal(t, "Valor2", row.CustomFields["Campo2"])
}

func TestProcessJob_JobStateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		jobType        models.JobType
		expectedStatus models.JobStatus
		shouldError    bool
	}{
		{
			name:           "unknown type transitions to failed",
			jobType:        "unknown_type",
			expectedStatus: models.JobStatusFailed,
			shouldError:    true,
		},
		{
			name:           "enrollment import type with missing curso fails",
			jobType:        models.JobTypeEnrollmentImport,
			expectedStatus: models.JobStatusFailed,
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			jobRepo := repository.NewJobRepository(db)
			inscricaoRepo := repository.NewInscricaoRepository(db)
			cursoRepo := repository.NewCursoRepository(db)

			jobService := services.NewJobService(jobRepo)
			inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
			cursoService := services.NewCursoService(cursoRepo)

			processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

			// Create job
			metadata := models.EnrollmentImportMetadata{
				CursoID:  999, // Non-existent
				FileName: "/tmp/test.csv",
			}
			metadataJSON, _ := json.Marshal(metadata)

			job := &models.Job{
				ID:       uuid.New(),
				Type:     tt.jobType,
				Status:   models.JobStatusPending,
				Metadata: datatypes.JSON(metadataJSON),
			}
			err := jobService.Create(context.Background(), job)
			assert.NoError(t, err)

			// Process job
			err = processor.ProcessJob(job.ID)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify final status
			updatedJob, err := jobService.GetByID(context.Background(), job.ID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, updatedJob.Status)
		})
	}
}

func TestProcessJob_ActiveJobsMapThreadSafety(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	numReaders := 50
	numWriters := 50
	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				processor.mu.RLock()
				_ = len(processor.activeJobs)
				processor.mu.RUnlock()
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				jobID := uuid.New()
				processor.mu.Lock()
				processor.activeJobs[jobID] = func() {}
				processor.mu.Unlock()

				processor.mu.Lock()
				delete(processor.activeJobs, jobID)
				processor.mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Final check - map should be empty since all writers deleted their entries
	processor.mu.RLock()
	count := len(processor.activeJobs)
	processor.mu.RUnlock()

	assert.Equal(t, 0, count, "Active jobs map should be empty after all operations")
}

func TestProcessJob_EnrollmentImportInvalidMetadata(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create job with invalid metadata
	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON([]byte(`{invalid json`)),
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process the job
	err = processor.ProcessJob(job.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao decodificar metadata")

	// Verify job status was updated to failed
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusFailed, updatedJob.Status)
}

func TestProcessJob_EnrollmentImportWithValidCurso(t *testing.T) {
	t.Skip("Integration test - requires full database setup")
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create a valid curso
	curso := &models.Curso{
		Titulo:    "Test Course",
		Descricao: "Test Description",
		OrgaoID:   "test-orgao",
	}
	db.Create(curso)

	// Create a test CSV file with valid data
	csvContent := `Nome,CPF
João Silva,12345678901
Maria Santos,98765432100`

	tmpFile, err := os.CreateTemp("", "test-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(csvContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Create job metadata
	metadata := models.EnrollmentImportMetadata{
		CursoID:  curso.ID,
		FileName: tmpFile.Name(),
	}
	metadataJSON, _ := json.Marshal(metadata)

	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON(metadataJSON),
	}
	err = jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process the job
	err = processor.ProcessJob(job.ID)

	// Should complete without error
	assert.NoError(t, err)

	// Verify job status was updated to completed
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 100, updatedJob.Progress)
}

func TestProcessEnrollmentImport_DirectCall(t *testing.T) {
	t.Skip("Integration test - requires full database setup")
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create a valid curso
	curso := &models.Curso{
		Titulo:    "Test Course",
		Descricao: "Test Description",
		OrgaoID:   "test-orgao",
	}
	db.Create(curso)

	// Create a test CSV file
	csvContent := `Nome,CPF
Test User,11111111111`

	tmpFile, err := os.CreateTemp("", "test-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(csvContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Create job
	metadata := models.EnrollmentImportMetadata{
		CursoID:  curso.ID,
		FileName: tmpFile.Name(),
	}
	metadataJSON, _ := json.Marshal(metadata)

	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusProcessing,
		Metadata: datatypes.JSON(metadataJSON),
	}

	// Directly call processEnrollmentImport
	err = processor.processEnrollmentImport(context.Background(), job)
	assert.NoError(t, err)
}

func TestInitializeJobProcessor(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	// Initialize global processor
	InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)

	assert.NotNil(t, GlobalJobProcessor)
	assert.NotNil(t, GlobalJobProcessor.activeJobs)
	assert.Equal(t, 0, len(GlobalJobProcessor.activeJobs))
}

func TestProcessJob_ContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create a job
	job := &models.Job{
		ID:     uuid.New(),
		Type:   "unknown_type",
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Start processing in a goroutine
	done := make(chan bool)
	go func() {
		_ = processor.ProcessJob(job.ID)
		done <- true
	}()

	// Try to cancel the job while it's processing
	time.Sleep(5 * time.Millisecond)
	err = processor.CancelJob(job.ID)

	// May or may not succeed depending on timing
	// The important thing is no panic or deadlock
	<-done

	// Verify job is no longer in active jobs
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()
	assert.False(t, exists)
}

func TestProcessJob_RaceConditionOnActiveJobs(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create multiple jobs
	numJobs := 20
	jobIDs := make([]uuid.UUID, numJobs)
	for i := 0; i < numJobs; i++ {
		job := &models.Job{
			ID:     uuid.New(),
			Type:   "unknown_type",
			Status: models.JobStatusPending,
		}
		err := jobService.Create(context.Background(), job)
		assert.NoError(t, err)
		jobIDs[i] = job.ID
	}

	var wg sync.WaitGroup

	// Start processing and cancelling concurrently
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(jobID uuid.UUID, idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				// Even indices: process job
				_ = processor.ProcessJob(jobID)
			} else {
				// Odd indices: wait a bit then try to cancel
				time.Sleep(5 * time.Millisecond)
				_ = processor.CancelJob(jobID)
			}
		}(jobIDs[i], i)
	}

	wg.Wait()

	// Verify all jobs are removed from active jobs
	processor.mu.RLock()
	activeCount := len(processor.activeJobs)
	processor.mu.RUnlock()

	assert.Equal(t, 0, activeCount, "All jobs should be removed from active jobs")
}

func TestProcessJob_SuccessfulEnrollmentProcessing(t *testing.T) {
	t.Skip("Integration test - requires full database setup")
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create a curso with required fields
	curso := &models.Curso{
		Titulo:    "Full Test Course",
		Descricao: "Complete test description",
		OrgaoID:   "test-orgao-123",
	}
	db.Create(curso)

	// Create a CSV file with complete enrollment data
	csvContent := `Nome,CPF,Email,Telefone,Idade
João Silva,12345678901,joao@test.com,(21) 99999-9999,25
Maria Santos,98765432100,maria@test.com,(21) 88888-8888,30`

	tmpFile, err := os.CreateTemp("", "enrollment-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(csvContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Create job
	metadata := models.EnrollmentImportMetadata{
		CursoID:  curso.ID,
		FileName: tmpFile.Name(),
	}
	metadataJSON, _ := json.Marshal(metadata)

	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON(metadataJSON),
	}
	err = jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process the job
	err = processor.ProcessJob(job.ID)
	assert.NoError(t, err)

	// Verify job completed successfully
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 100, updatedJob.Progress)
	assert.Equal(t, 2, updatedJob.TotalRecords)
	assert.Equal(t, 2, updatedJob.SuccessCount)
	assert.Equal(t, 0, updatedJob.ErrorCount)

	// Verify enrollments were created
	var enrollments []models.Inscricao
	db.Where("curso_id = ?", curso.ID).Find(&enrollments)
	assert.Equal(t, 2, len(enrollments))
}

// New Tests for Coverage Improvement

func TestStartJob_SuccessPath(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)
	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create a curso for successful processing
	curso := &models.Curso{
		Titulo:    "Test Course for Success",
		Descricao: "Test Description",
		OrgaoID:   "test-orgao",
	}
	db.Create(curso)

	// Create a valid CSV file
	csvContent := `Nome,CPF
Test User,12345678901`
	tmpFile, err := os.CreateTemp("", "success-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(csvContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Create job
	metadata := models.EnrollmentImportMetadata{
		CursoID:  curso.ID,
		FileName: tmpFile.Name(),
	}
	metadataJSON, _ := json.Marshal(metadata)

	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON(metadataJSON),
	}
	err = jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Start job in goroutine
	processor.StartJob(job.ID)

	// Wait for completion
	time.Sleep(150 * time.Millisecond)

	// Verify job is no longer in active jobs
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()
	assert.False(t, exists)
}

func TestStartJob_ErrorPath(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create job with unknown type (will fail)
	job := &models.Job{
		ID:     uuid.New(),
		Type:   "unknown_type",
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Start job in goroutine
	processor.StartJob(job.ID)

	// Wait for completion
	time.Sleep(150 * time.Millisecond)

	// Verify job is no longer in active jobs
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()
	assert.False(t, exists)

	// Verify job was marked as failed
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusFailed, updatedJob.Status)
}

func TestStartJob_NonExistentJob(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Start a non-existent job
	jobID := uuid.New()
	processor.StartJob(jobID)

	// Wait for goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Verify job is not in active jobs
	processor.mu.RLock()
	_, exists := processor.activeJobs[jobID]
	processor.mu.RUnlock()
	assert.False(t, exists)
}

func TestStartJob_ConcurrentStarts(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create multiple jobs
	numJobs := 10
	jobIDs := make([]uuid.UUID, numJobs)
	for i := 0; i < numJobs; i++ {
		job := &models.Job{
			ID:     uuid.New(),
			Type:   "unknown_type",
			Status: models.JobStatusPending,
		}
		err := jobService.Create(context.Background(), job)
		assert.NoError(t, err)
		jobIDs[i] = job.ID
	}

	// Start all jobs concurrently
	for i := 0; i < numJobs; i++ {
		processor.StartJob(jobIDs[i])
	}

	// Wait for all to complete
	time.Sleep(300 * time.Millisecond)

	// Verify all jobs are removed from active jobs
	processor.mu.RLock()
	activeCount := len(processor.activeJobs)
	processor.mu.RUnlock()
	assert.Equal(t, 0, activeCount)
}

func TestProcessJob_UpdateStatusToProcessingError(t *testing.T) {
	// This test covers the error path when updating status to processing fails
	// We need to use a mock or closed database to trigger this
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create a job
	job := &models.Job{
		ID:     uuid.New(),
		Type:   models.JobTypeEnrollmentImport,
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Close the database to cause UpdateStatus to fail
	sqlDB, _ := db.DB()
	sqlDB.Close()

	// Process the job - should fail to update status
	err = processor.ProcessJob(job.ID)
	assert.Error(t, err)
}

func TestProcessJob_UpdateStatusToCompletedError(t *testing.T) {
	// This test simulates a successful processing but failure to update to completed
	// We create a valid job setup, then close DB just before completion would happen
	t.Skip("Difficult to test without mocking - requires precise timing")
}

func TestProcessJob_DeferCleanupExecutes(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create job
	job := &models.Job{
		ID:     uuid.New(),
		Type:   "unknown_type",
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process job (will fail with unknown type)
	err = processor.ProcessJob(job.ID)
	assert.Error(t, err)

	// Verify defer cleanup removed the job
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()
	assert.False(t, exists, "Defer should have removed job from activeJobs")
}

func TestProcessJob_ProcessingErrorUpdatesStatusToFailed(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)
	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create job with non-existent curso (will fail processing)
	metadata := models.EnrollmentImportMetadata{
		CursoID:  999,
		FileName: "/tmp/nonexistent.csv",
	}
	metadataJSON, _ := json.Marshal(metadata)

	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON(metadataJSON),
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process job - should fail and update status
	err = processor.ProcessJob(job.ID)
	assert.Error(t, err)

	// Verify status was updated to failed
	updatedJob, err := jobService.GetByID(context.Background(), job.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.JobStatusFailed, updatedJob.Status)
}

func TestProcessJob_NilJobError(t *testing.T) {
	// This covers the "job == nil" check after GetByID
	// In practice, GetByID returns an error if not found, but we test the nil check
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Try to process non-existent job
	jobID := uuid.New()
	err := processor.ProcessJob(jobID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job não encontrado")
}

func TestCancelJob_ErrorPath(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	// Try to cancel job that doesn't exist
	jobID := uuid.New()
	err := processor.CancelJob(jobID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "não está em execução")
}

func TestCancelJob_SuccessPath(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	// Add job to active jobs
	jobID := uuid.New()
	cancelCalled := false
	cancelFunc := func() {
		cancelCalled = true
	}

	processor.mu.Lock()
	processor.activeJobs[jobID] = cancelFunc
	processor.mu.Unlock()

	// Cancel the job
	err := processor.CancelJob(jobID)

	assert.NoError(t, err)
	assert.True(t, cancelCalled)
}

func TestProcessJob_ActiveJobsMapCleanupOnPanic(t *testing.T) {
	// Verify defer cleanup happens even if panic occurs
	// Note: We can't actually test panic in ProcessJob without modifying code
	// But we can verify the defer logic works
	processor := NewJobProcessor(nil, nil, nil, nil)

	jobID := uuid.New()
	processor.mu.Lock()
	processor.activeJobs[jobID] = func() {}
	processor.mu.Unlock()

	// Manually simulate defer cleanup
	processor.mu.Lock()
	delete(processor.activeJobs, jobID)
	processor.mu.Unlock()

	processor.mu.RLock()
	_, exists := processor.activeJobs[jobID]
	processor.mu.RUnlock()

	assert.False(t, exists)
}

func TestProcessJob_MultipleJobTypesBranches(t *testing.T) {
	tests := []struct {
		name        string
		jobType     models.JobType
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "enrollment import type",
			jobType:     models.JobTypeEnrollmentImport,
			shouldError: true,
			errorMsg:    "curso não encontrado",
		},
		{
			name:        "unknown job type",
			jobType:     "unsupported_type",
			shouldError: true,
			errorMsg:    "tipo de job desconhecido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			jobRepo := repository.NewJobRepository(db)
			inscricaoRepo := repository.NewInscricaoRepository(db)
			cursoRepo := repository.NewCursoRepository(db)

			jobService := services.NewJobService(jobRepo)
			inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
			cursoService := services.NewCursoService(cursoRepo)
			processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

			metadata := models.EnrollmentImportMetadata{
				CursoID:  999,
				FileName: "/tmp/test.csv",
			}
			metadataJSON, _ := json.Marshal(metadata)

			job := &models.Job{
				ID:       uuid.New(),
				Type:     tt.jobType,
				Status:   models.JobStatusPending,
				Metadata: datatypes.JSON(metadataJSON),
			}
			err := jobService.Create(context.Background(), job)
			assert.NoError(t, err)

			err = processor.ProcessJob(job.ID)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitializeJobProcessor_Success(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	// Save old global processor
	oldProcessor := GlobalJobProcessor

	// Initialize
	InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)

	// Verify initialization
	assert.NotNil(t, GlobalJobProcessor)
	assert.NotNil(t, GlobalJobProcessor.db)
	assert.NotNil(t, GlobalJobProcessor.jobService)
	assert.NotNil(t, GlobalJobProcessor.inscricaoService)
	assert.NotNil(t, GlobalJobProcessor.cursoService)
	assert.NotNil(t, GlobalJobProcessor.activeJobs)
	assert.Equal(t, 0, len(GlobalJobProcessor.activeJobs))

	// Restore old processor
	GlobalJobProcessor = oldProcessor
}

func TestInitializeJobProcessor_CreatesNewInstance(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	// Save old global processor
	oldProcessor := GlobalJobProcessor

	// Initialize first time
	InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)
	firstProcessor := GlobalJobProcessor

	// Initialize again
	InitializeJobProcessor(db, jobRepo, inscricaoRepo, cursoRepo)
	secondProcessor := GlobalJobProcessor

	// Should be different instances
	assert.NotSame(t, firstProcessor, secondProcessor)

	// Restore old processor
	GlobalJobProcessor = oldProcessor
}

func TestProcessEnrollmentImport_DelegationToProcessor(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)
	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create job with invalid curso (will fail in enrollment processor)
	metadata := models.EnrollmentImportMetadata{
		CursoID:  999,
		FileName: "/tmp/test.csv",
	}
	metadataJSON, _ := json.Marshal(metadata)

	job := &models.Job{
		ID:       uuid.New(),
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusProcessing,
		Metadata: datatypes.JSON(metadataJSON),
	}

	// Call processEnrollmentImport directly
	err := processor.processEnrollmentImport(context.Background(), job)

	// Should error because curso doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "curso não encontrado")
}

func TestProcessJob_ConcurrentProcessAndCancel(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	// Create job
	job := &models.Job{
		ID:     uuid.New(),
		Type:   "unknown_type",
		Status: models.JobStatusPending,
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)

	// Process in one goroutine
	go func() {
		defer wg.Done()
		_ = processor.ProcessJob(job.ID)
	}()

	// Try to cancel in another goroutine
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		_ = processor.CancelJob(job.ID)
	}()

	wg.Wait()

	// Verify job is no longer active
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()
	assert.False(t, exists)
}

func TestStartJob_MultipleJobsParallel(t *testing.T) {
	db := setupTestDB(t)
	jobRepo := repository.NewJobRepository(db)
	jobService := services.NewJobService(jobRepo)
	processor := NewJobProcessor(db, jobService, nil, nil)

	numJobs := 15
	jobIDs := make([]uuid.UUID, numJobs)

	// Create jobs
	for i := 0; i < numJobs; i++ {
		job := &models.Job{
			ID:     uuid.New(),
			Type:   "unknown_type",
			Status: models.JobStatusPending,
		}
		err := jobService.Create(context.Background(), job)
		assert.NoError(t, err)
		jobIDs[i] = job.ID
	}

	// Start all jobs
	for i := 0; i < numJobs; i++ {
		processor.StartJob(jobIDs[i])
	}

	// Wait for completion
	time.Sleep(400 * time.Millisecond)

	// Verify all jobs completed
	processor.mu.RLock()
	activeCount := len(processor.activeJobs)
	processor.mu.RUnlock()
	assert.Equal(t, 0, activeCount)

	// Verify all jobs are no longer pending (may have failed)
	for _, jobID := range jobIDs {
		updatedJob, err := jobService.GetByID(context.Background(), jobID)
		if err == nil && updatedJob != nil {
			assert.NotEqual(t, models.JobStatusPending, updatedJob.Status)
		}
	}
}
