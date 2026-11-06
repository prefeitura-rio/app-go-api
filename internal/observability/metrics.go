package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Operational metrics read by the production alerts (docs/runbook.md, section
// 3). They live in the Prometheus default registry served at GET /metrics.

// Sync workers, the "worker" label.
const (
	SyncWorkerOrgao   = "orgao_sync"
	SyncWorkerCitizen = "citizen_sync"
)

// Outcome of one sync cycle, the "result" label of goapi_sync_runs_total.
const (
	syncRunSuccess = "success"
	syncRunFailure = "failure"
	syncRunSkipped = "skipped"
)

// Outcome of one synced item, the "result" label of goapi_sync_items_total.
const (
	SyncItemSynced  = "synced"
	SyncItemSkipped = "skipped"
	SyncItemFailed  = "failed"
)

// Outcome of one import job, the "result" label of goapi_import_jobs_total.
const (
	ImportJobCompleted = "completed"
	ImportJobFailed    = "failed"
	// ImportJobReaped is a job the stale job reaper closed as failed because
	// its pod stopped renewing it.
	ImportJobReaped = "reaped"
)

// Why an e-mail task went to the dead-letter queue, the "reason" label.
const (
	EmailDeadLetterPanic            = "panic"
	EmailDeadLetterRetriesExhausted = "retries_exhausted"
)

var (
	syncLastSuccess = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "goapi_sync_last_success_timestamp_seconds",
		Help: "Horário Unix do último ciclo de sync concluído sem erro neste pod (0 = nenhum ainda). Agregue com max() entre os pods.",
	}, []string{"worker"})
	syncRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goapi_sync_runs_total",
		Help: "Ciclos de sync por resultado: success, failure ou skipped (outro pod segurava o lock).",
	}, []string{"worker", "result"})
	syncItemsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goapi_sync_items_total",
		Help: "Itens tratados pelos ciclos de sync por resultado: synced, skipped ou failed.",
	}, []string{"worker", "result"})
	importJobsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goapi_import_jobs_total",
		Help: "Jobs de importação encerrados por resultado: completed, failed ou reaped (interrompidos, fechados pelo reaper).",
	}, []string{"result"})
	emailDeadLetterTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goapi_email_dead_letter_total",
		Help: "Tarefas de e-mail movidas para a fila morta (email:queue:failed) por motivo.",
	}, []string{"reason"})
)

func init() {
	// Create every known series at 0, so increase() sees the first event and
	// the absence of a sync is visible as a 0 timestamp.
	for _, worker := range []string{SyncWorkerOrgao, SyncWorkerCitizen} {
		syncLastSuccess.WithLabelValues(worker)
		for _, result := range []string{syncRunSuccess, syncRunFailure, syncRunSkipped} {
			syncRunsTotal.WithLabelValues(worker, result)
		}
		for _, result := range []string{SyncItemSynced, SyncItemSkipped, SyncItemFailed} {
			syncItemsTotal.WithLabelValues(worker, result)
		}
	}
	for _, result := range []string{ImportJobCompleted, ImportJobFailed, ImportJobReaped} {
		importJobsTotal.WithLabelValues(result)
	}
	for _, reason := range []string{EmailDeadLetterPanic, EmailDeadLetterRetriesExhausted} {
		emailDeadLetterTotal.WithLabelValues(reason)
	}
}

// ObserveSyncRun records a sync cycle that ran. A nil err is a success and
// moves the last success timestamp; otherwise it counts as a failure.
func ObserveSyncRun(worker string, err error) {
	if err != nil {
		syncRunsTotal.WithLabelValues(worker, syncRunFailure).Inc()
		return
	}
	syncRunsTotal.WithLabelValues(worker, syncRunSuccess).Inc()
	syncLastSuccess.WithLabelValues(worker).Set(float64(time.Now().Unix()))
}

// ObserveSyncSkipped records a sync cycle this pod did not run because another
// pod held the lock. It does not move the last success timestamp.
func ObserveSyncSkipped(worker string) {
	syncRunsTotal.WithLabelValues(worker, syncRunSkipped).Inc()
}

// AddSyncItems adds n items with the given result to the worker's count.
func AddSyncItems(worker, result string, n int) {
	if n > 0 {
		syncItemsTotal.WithLabelValues(worker, result).Add(float64(n))
	}
}

// ObserveImportJobs adds n import jobs that finished with the given result.
func ObserveImportJobs(result string, n int64) {
	if n > 0 {
		importJobsTotal.WithLabelValues(result).Add(float64(n))
	}
}

// ObserveEmailDeadLetter records an e-mail task moved to the dead-letter queue.
func ObserveEmailDeadLetter(reason string) {
	emailDeadLetterTotal.WithLabelValues(reason).Inc()
}
