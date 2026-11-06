package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// importJobs reads goapi_import_jobs_total{result} from the default registry
// served at /metrics.
func importJobs(t *testing.T, result string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	for _, mf := range families {
		if mf.GetName() != "goapi_import_jobs_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "result" && lp.GetValue() == result {
					return m.GetCounter().GetValue()
				}
			}
		}
	}
	return 0
}

func TestProcessJob_Completed_CountsCompletedImportJob(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	p.processors[models.JobTypeEnrollmentImport] = func(context.Context, *models.Job) error { return nil }
	job := createTestJob(models.JobTypeEnrollmentImport, nil)
	require.NoError(t, db.Create(job).Error)

	before := importJobs(t, "completed")
	beforeFailed := importJobs(t, "failed")

	require.NoError(t, p.ProcessJob(job.ID))

	assert.Equal(t, before+1, importJobs(t, "completed"))
	assert.Equal(t, beforeFailed, importJobs(t, "failed"))
}

func TestProcessJob_ProcessorError_CountsFailedImportJob(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	p.processors[models.JobTypeEnrollmentImport] = func(context.Context, *models.Job) error {
		return errors.New("arquivo inválido")
	}
	job := createTestJob(models.JobTypeEnrollmentImport, nil)
	require.NoError(t, db.Create(job).Error)

	before := importJobs(t, "failed")

	require.Error(t, p.ProcessJob(job.ID))

	assert.Equal(t, before+1, importJobs(t, "failed"))
}

func TestReapStaleJobs_CountsReapedImportJobs(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	longAgo := time.Now().Add(-1 * time.Hour)
	insertJobUpdatedAt(t, db, models.JobStatusProcessing, longAgo)
	insertJobUpdatedAt(t, db, models.JobStatusPending, longAgo)
	insertJobUpdatedAt(t, db, models.JobStatusProcessing, time.Now())

	before := importJobs(t, "reaped")

	n, err := p.ReapStaleJobs(context.Background(), 10*time.Minute)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)

	assert.Equal(t, before+2, importJobs(t, "reaped"))
}

func TestReapStaleJobs_NothingStale_CountsNothing(t *testing.T) {
	p, db := newRecoveryProcessor(t)
	insertJobUpdatedAt(t, db, models.JobStatusProcessing, time.Now())

	before := importJobs(t, "reaped")

	n, err := p.ReapStaleJobs(context.Background(), 10*time.Minute)
	require.NoError(t, err)
	require.Zero(t, n)

	assert.Equal(t, before, importJobs(t, "reaped"))
}
