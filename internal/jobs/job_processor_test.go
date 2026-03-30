package jobs

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

// Note: ProcessJob error path testing requires integration tests with real services
// because JobProcessor uses concrete types instead of interfaces.
// The following error paths in ProcessJob are covered by integration tests:
// - GetByID error
// - Job not found
// - UpdateStatus to processing error
// - Unknown job type
// - UpdateStatus to completed error
// - UpdateStatus to failed error
