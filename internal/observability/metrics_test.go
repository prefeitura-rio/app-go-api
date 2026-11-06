package observability

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObserveSyncRun_Success_CountsRunAndMovesLastSuccess(t *testing.T) {
	before := testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerOrgao, syncRunSuccess))
	start := time.Now()

	ObserveSyncRun(SyncWorkerOrgao, nil)

	assert.Equal(t, before+1, testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerOrgao, syncRunSuccess)))
	last := testutil.ToFloat64(syncLastSuccess.WithLabelValues(SyncWorkerOrgao))
	assert.GreaterOrEqual(t, last, float64(start.Unix()))
	assert.LessOrEqual(t, last, float64(time.Now().Unix())+1)
}

func TestObserveSyncRun_Failure_CountsFailureAndKeepsLastSuccess(t *testing.T) {
	syncLastSuccess.WithLabelValues(SyncWorkerCitizen).Set(42)
	before := testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerCitizen, syncRunFailure))

	ObserveSyncRun(SyncWorkerCitizen, errors.New("rmi fora"))

	assert.Equal(t, before+1, testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerCitizen, syncRunFailure)))
	assert.Equal(t, float64(42), testutil.ToFloat64(syncLastSuccess.WithLabelValues(SyncWorkerCitizen)))
}

func TestObserveSyncSkipped_CountsSkippedAndKeepsLastSuccess(t *testing.T) {
	syncLastSuccess.WithLabelValues(SyncWorkerOrgao).Set(7)
	beforeSkipped := testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerOrgao, syncRunSkipped))
	beforeSuccess := testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerOrgao, syncRunSuccess))

	ObserveSyncSkipped(SyncWorkerOrgao)

	assert.Equal(t, beforeSkipped+1, testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerOrgao, syncRunSkipped)))
	assert.Equal(t, beforeSuccess, testutil.ToFloat64(syncRunsTotal.WithLabelValues(SyncWorkerOrgao, syncRunSuccess)))
	assert.Equal(t, float64(7), testutil.ToFloat64(syncLastSuccess.WithLabelValues(SyncWorkerOrgao)))
}

func TestAddSyncItems_AddsPositiveCountsOnly(t *testing.T) {
	tests := []struct {
		name  string
		n     int
		delta float64
	}{
		{"positive", 3, 3},
		{"zero", 0, 0},
		{"negative is ignored", -2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := testutil.ToFloat64(syncItemsTotal.WithLabelValues(SyncWorkerCitizen, SyncItemFailed))
			AddSyncItems(SyncWorkerCitizen, SyncItemFailed, tt.n)
			assert.Equal(t, before+tt.delta, testutil.ToFloat64(syncItemsTotal.WithLabelValues(SyncWorkerCitizen, SyncItemFailed)))
		})
	}
}

func TestObserveImportJobs_CountsByResult(t *testing.T) {
	tests := []struct {
		result string
		n      int64
		delta  float64
	}{
		{ImportJobCompleted, 1, 1},
		{ImportJobFailed, 1, 1},
		{ImportJobReaped, 4, 4},
		{ImportJobReaped, 0, 0},
		{ImportJobReaped, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.result, func(t *testing.T) {
			before := testutil.ToFloat64(importJobsTotal.WithLabelValues(tt.result))
			ObserveImportJobs(tt.result, tt.n)
			assert.Equal(t, before+tt.delta, testutil.ToFloat64(importJobsTotal.WithLabelValues(tt.result)))
		})
	}
}

func TestObserveEmailDeadLetter_CountsByReason(t *testing.T) {
	for _, reason := range []string{EmailDeadLetterPanic, EmailDeadLetterRetriesExhausted} {
		t.Run(reason, func(t *testing.T) {
			before := testutil.ToFloat64(emailDeadLetterTotal.WithLabelValues(reason))
			ObserveEmailDeadLetter(reason)
			assert.Equal(t, before+1, testutil.ToFloat64(emailDeadLetterTotal.WithLabelValues(reason)))
		})
	}
}

// The alert rules read these series from the default registry served at
// /metrics. Every known label combination must exist from startup (value 0),
// otherwise increase() over a counter born after the first failure misses it.
func TestMetrics_RegisteredWithKnownLabelsFromStartup(t *testing.T) {
	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	series := map[string]int{}
	for _, mf := range families {
		if strings.HasPrefix(mf.GetName(), "goapi_") {
			series[mf.GetName()] = len(mf.GetMetric())
		}
	}

	assert.GreaterOrEqual(t, series["goapi_sync_last_success_timestamp_seconds"], 2, "orgao_sync and citizen_sync")
	assert.GreaterOrEqual(t, series["goapi_sync_runs_total"], 6, "2 workers x success/failure/skipped")
	assert.GreaterOrEqual(t, series["goapi_sync_items_total"], 6, "2 workers x synced/skipped/failed")
	assert.GreaterOrEqual(t, series["goapi_import_jobs_total"], 3, "completed/failed/reaped")
	assert.GreaterOrEqual(t, series["goapi_email_dead_letter_total"], 2, "panic/retries_exhausted")
}
