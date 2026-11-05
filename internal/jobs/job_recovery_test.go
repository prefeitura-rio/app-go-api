package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// newRecoveryProcessor wires a JobProcessor over an in-memory SQLite jobs table.
func newRecoveryProcessor(t *testing.T) (*JobProcessor, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	jobService := services.NewJobService(repository.NewJobRepository(db))
	return NewJobProcessor(db, jobService, nil, nil), db
}

// insertJobUpdatedAt stores a job with the given status and last update time,
// simulating a job left behind by a pod that stopped renewing it.
func insertJobUpdatedAt(t *testing.T, db *gorm.DB, status models.JobStatus, updatedAt time.Time) uuid.UUID {
	t.Helper()
	job := createTestJob(models.JobTypeEnrollmentImport, nil)
	job.Status = status
	require.NoError(t, db.Create(job).Error)
	require.NoError(t, db.Model(&models.Job{}).Where("id = ?", job.ID).UpdateColumn("updated_at", updatedAt).Error)
	return job.ID
}

func loadJob(t *testing.T, db *gorm.DB, id uuid.UUID) models.Job {
	t.Helper()
	var job models.Job
	require.NoError(t, db.First(&job, "id = ?", id).Error)
	return job
}

func TestReapStaleJobs_InterruptedJobsEndFailedWithReason(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	longAgo := time.Now().Add(-1 * time.Hour)

	orphanProcessing := insertJobUpdatedAt(t, db, models.JobStatusProcessing, longAgo)
	orphanPending := insertJobUpdatedAt(t, db, models.JobStatusPending, longAgo)
	aliveProcessing := insertJobUpdatedAt(t, db, models.JobStatusProcessing, time.Now())
	oldCompleted := insertJobUpdatedAt(t, db, models.JobStatusCompleted, longAgo)
	oldFailed := insertJobUpdatedAt(t, db, models.JobStatusFailed, longAgo)

	n, err := p.ReapStaleJobs(context.Background(), 10*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	for _, id := range []uuid.UUID{orphanProcessing, orphanPending} {
		job := loadJob(t, db, id)
		assert.Equal(t, models.JobStatusFailed, job.Status, "interrupted job must end failed")
		assert.NotNil(t, job.CompletedAt, "interrupted job must carry completed_at")

		var jobErrors []models.JobError
		require.NoError(t, json.Unmarshal(job.Errors, &jobErrors))
		require.Len(t, jobErrors, 1)
		assert.Equal(t, services.JobInterruptedMessage, jobErrors[0].Message)
	}

	assert.Equal(t, models.JobStatusProcessing, loadJob(t, db, aliveProcessing).Status, "job renewed recently must keep running")
	assert.Equal(t, models.JobStatusCompleted, loadJob(t, db, oldCompleted).Status, "finished job must not change")
	assert.Nil(t, loadJob(t, db, oldCompleted).Errors)
	assert.Equal(t, models.JobStatusFailed, loadJob(t, db, oldFailed).Status)
	assert.Nil(t, loadJob(t, db, oldFailed).Errors, "already failed job must not be rewritten")
}

func TestReapStaleJobs_IsIdempotent(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	insertJobUpdatedAt(t, db, models.JobStatusProcessing, time.Now().Add(-1*time.Hour))

	n, err := p.ReapStaleJobs(context.Background(), 10*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	n, err = p.ReapStaleJobs(context.Background(), 10*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n, "second sweep must not touch the already failed job")
}

func TestRunStaleJobReaper_SweepsOnStartupAndPeriodically(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	first := insertJobUpdatedAt(t, db, models.JobStatusProcessing, time.Now().Add(-1*time.Hour))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.RunStaleJobReaper(ctx, 20*time.Millisecond, 10*time.Minute)
		close(done)
	}()

	assert.Eventually(t, func() bool {
		return loadJob(t, db, first).Status == models.JobStatusFailed
	}, 2*time.Second, 5*time.Millisecond, "startup sweep must fail the orphaned job")

	second := insertJobUpdatedAt(t, db, models.JobStatusProcessing, time.Now().Add(-1*time.Hour))
	assert.Eventually(t, func() bool {
		return loadJob(t, db, second).Status == models.JobStatusFailed
	}, 2*time.Second, 5*time.Millisecond, "periodic sweep must fail jobs orphaned later")

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reaper did not stop after context cancellation")
	}
}

func TestRunStaleJobReaper_InvalidSettingsDoNotRun(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	id := insertJobUpdatedAt(t, db, models.JobStatusProcessing, time.Now().Add(-1*time.Hour))

	done := make(chan struct{})
	go func() {
		p.RunStaleJobReaper(context.Background(), 0, 10*time.Minute)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reaper with invalid interval must return immediately")
	}
	assert.Equal(t, models.JobStatusProcessing, loadJob(t, db, id).Status)
}

func TestHeartbeat_RenewsRunningJobUntilStopped(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	longAgo := time.Now().Add(-1 * time.Hour)
	id := insertJobUpdatedAt(t, db, models.JobStatusProcessing, longAgo)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.heartbeat(ctx, id, 10*time.Millisecond)
		close(done)
	}()

	assert.Eventually(t, func() bool {
		return loadJob(t, db, id).UpdatedAt.After(time.Now().Add(-1 * time.Minute))
	}, 2*time.Second, 5*time.Millisecond, "heartbeat must renew updated_at")

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("heartbeat did not stop after context cancellation")
	}

	// A renewed job is not reaped.
	n, err := p.ReapStaleJobs(context.Background(), 10*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestProcessJob_RunsHeartbeatWhileProcessing(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	p.SetHeartbeatInterval(5 * time.Millisecond)

	id := insertJobUpdatedAt(t, db, models.JobStatusPending, time.Now().Add(-1*time.Hour))

	started := make(chan struct{})
	release := make(chan struct{})
	p.processors = map[models.JobType]func(context.Context, *models.Job) error{
		models.JobTypeEnrollmentImport: func(ctx context.Context, job *models.Job) error {
			// Make the job look abandoned while it is still running.
			assert.NoError(t, db.Model(&models.Job{}).Where("id = ?", job.ID).UpdateColumn("updated_at", time.Now().Add(-1*time.Hour)).Error)
			close(started)
			<-release
			return nil
		},
	}

	done := make(chan error, 1)
	go func() { done <- p.ProcessJob(id) }()
	<-started

	assert.Eventually(t, func() bool {
		return loadJob(t, db, id).UpdatedAt.After(time.Now().Add(-1 * time.Minute))
	}, 2*time.Second, 5*time.Millisecond, "running job must be renewed by the heartbeat")

	n, err := p.ReapStaleJobs(context.Background(), 10*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n, "a job that is still running must not be reaped")

	close(release)
	require.NoError(t, <-done)
	assert.Equal(t, models.JobStatusCompleted, loadJob(t, db, id).Status)
}

func TestSetHeartbeatInterval_IgnoresNonPositive(t *testing.T) {
	p := NewJobProcessor(nil, nil, nil, nil)
	before := p.heartbeatInterval
	p.SetHeartbeatInterval(0)
	assert.Equal(t, before, p.heartbeatInterval)
	p.SetHeartbeatInterval(-time.Second)
	assert.Equal(t, before, p.heartbeatInterval)
	p.SetHeartbeatInterval(3 * time.Second)
	assert.Equal(t, 3*time.Second, p.heartbeatInterval)
}
