package workers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// gaugeOrCounter reads one series of the default registry (served at /metrics).
// It returns 0 when the series does not exist.
func gaugeOrCounter(t *testing.T, name string, labels map[string]string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	for _, mf := range families {
		if mf.GetName() != name {
			continue
		}
	next:
		for _, m := range mf.GetMetric() {
			got := map[string]string{}
			for _, lp := range m.GetLabel() {
				got[lp.GetName()] = lp.GetValue()
			}
			for k, v := range labels {
				if got[k] != v {
					continue next
				}
			}
			if m.GetCounter() != nil {
				return m.GetCounter().GetValue()
			}
			return m.GetGauge().GetValue()
		}
	}
	return 0
}

func discardLogger() *log.Logger { return log.New(io.Discard, "", 0) }

func syncRuns(t *testing.T, worker, result string) float64 {
	return gaugeOrCounter(t, "goapi_sync_runs_total", map[string]string{"worker": worker, "result": result})
}

func syncLastSuccess(t *testing.T, worker string) float64 {
	return gaugeOrCounter(t, "goapi_sync_last_success_timestamp_seconds", map[string]string{"worker": worker})
}

func deadLetters(t *testing.T, reason string) float64 {
	return gaugeOrCounter(t, "goapi_email_dead_letter_total", map[string]string{"reason": reason})
}

func newOrgaoWorkerForMetrics(t *testing.T, redisClient *redis.Client) *OrgaoSyncWorker {
	t.Helper()
	db := setupTestDB(t)
	return NewOrgaoSyncWorker(db, nil, redisClient, repository.NewOrgaoSnapshotRepository(db), nil, nil, nil,
		&config.OrgaoSyncSettings{SyncInterval: time.Minute, StaleThreshold: time.Hour, FailedRetryThreshold: time.Hour, BatchSize: 10, MaxRetries: 1})
}

// --- orgao sync ---

func TestOrgaoSyncWorker_runSync_CompletedCycle_RecordsSuccessAndLastSuccess(t *testing.T) {
	w := newOrgaoWorkerForMetrics(t, nil)
	before := syncRuns(t, "orgao_sync", "success")
	start := time.Now().Unix()

	require.NoError(t, w.runSync(context.Background()))

	assert.Equal(t, before+1, syncRuns(t, "orgao_sync", "success"))
	assert.GreaterOrEqual(t, syncLastSuccess(t, "orgao_sync"), float64(start))
}

func TestOrgaoSyncWorker_runSync_LockHeldByAnotherPod_RecordsSkipped(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, mr.Set(orgaoSyncLockKey, "outro-pod"))

	w := newOrgaoWorkerForMetrics(t, client)
	beforeSkipped := syncRuns(t, "orgao_sync", "skipped")
	beforeSuccess := syncRuns(t, "orgao_sync", "success")

	require.NoError(t, w.runSync(context.Background()))

	assert.Equal(t, beforeSkipped+1, syncRuns(t, "orgao_sync", "skipped"))
	assert.Equal(t, beforeSuccess, syncRuns(t, "orgao_sync", "success"), "skipped cycle is not a success")
}

func TestOrgaoSyncWorker_runSync_DatabaseError_RecordsFailureKeepsLastSuccess(t *testing.T) {
	w := newOrgaoWorkerForMetrics(t, nil)
	sqlDB, err := w.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	lastBefore := syncLastSuccess(t, "orgao_sync")
	before := syncRuns(t, "orgao_sync", "failure")

	require.Error(t, w.runSync(context.Background()))

	assert.Equal(t, before+1, syncRuns(t, "orgao_sync", "failure"))
	assert.Equal(t, lastBefore, syncLastSuccess(t, "orgao_sync"))
}

func TestOrgaoSyncWorker_runSync_RedisDown_RecordsFailure(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = client.Close() })
	mr.Close()

	w := newOrgaoWorkerForMetrics(t, client)
	before := syncRuns(t, "orgao_sync", "failure")

	require.Error(t, w.runSync(context.Background()))

	assert.Equal(t, before+1, syncRuns(t, "orgao_sync", "failure"))
}

// --- citizen sync ---

func TestCitizenSyncWorker_runSync_DatabaseError_RecordsFailureKeepsLastSuccess(t *testing.T) {
	db := setupTestDB(t) // no citizen tables: the query fails
	w := NewCitizenSyncWorker(nil, repository.NewCitizenSnapshotRepository(db), nil,
		&config.CitizenSyncSettings{SyncInterval: time.Minute, StaleThreshold: time.Hour, BatchSize: 10})
	w.infoLog = discardLogger()
	w.errorLog = discardLogger()

	lastBefore := syncLastSuccess(t, "citizen_sync")
	before := syncRuns(t, "citizen_sync", "failure")

	require.Error(t, w.runSync(context.Background()))

	assert.Equal(t, before+1, syncRuns(t, "citizen_sync", "failure"))
	assert.Equal(t, lastBefore, syncLastSuccess(t, "citizen_sync"))
}

// --- e-mail dead-letter queue ---

func TestEmailWorker_consume_ExhaustedRetries_CountsDeadLetter(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()
	before := deadLetters(t, "retries_exhausted")

	pushTask(t, worker.redis, emailQueueKey, emailTask{
		Type:     "bad.type",
		Payload:  json.RawMessage(`{}`),
		Attempts: emailMaxRetries - 1,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.consume(ctx)

	assert.Eventually(t, func() bool {
		return queueLen(worker.redis, emailDeadLetterKey) == 1
	}, 2*time.Second, 20*time.Millisecond)
	assert.Eventually(t, func() bool {
		return deadLetters(t, "retries_exhausted") == before+1
	}, time.Second, 10*time.Millisecond)
}

func TestEmailWorker_consume_PanickingTask_CountsDeadLetter(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()
	worker.emailService = &panickingNotifier{}
	before := deadLetters(t, "panic")

	pushTask(t, worker.redis, emailQueueKey, buildCandidaturaSemVagaTask(t, taskCandidaturaReprovada))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.consume(ctx)

	assert.Eventually(t, func() bool {
		return queueLen(worker.redis, emailDeadLetterKey) == 1
	}, 3*time.Second, 20*time.Millisecond)
	assert.Eventually(t, func() bool {
		return deadLetters(t, "panic") == before+1
	}, time.Second, 10*time.Millisecond)
}

// --- e-mail queue lengths ---

const emailQueueMetricsText = `
# HELP goapi_email_queue_length Tarefas de e-mail paradas em cada fila do Redis: pending (email:queue), delayed (email:queue:delayed, aguardando nova tentativa) e dead_letter (email:queue:failed).
# TYPE goapi_email_queue_length gauge
goapi_email_queue_length{queue="dead_letter"} 1
goapi_email_queue_length{queue="delayed"} 2
goapi_email_queue_length{queue="pending"} 3
# HELP goapi_email_queue_up 1 quando a última leitura das filas de e-mail no Redis funcionou, 0 quando falhou.
# TYPE goapi_email_queue_up gauge
goapi_email_queue_up 1
`

func TestEmailQueueCollector_ReportsLengthOfEachQueue(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, client.LPush(ctx, emailQueueKey, "t").Err())
	}
	require.NoError(t, client.ZAdd(ctx, emailDelayedQueueKey, redis.Z{Score: 1, Member: "a"}, redis.Z{Score: 2, Member: "b"}).Err())
	require.NoError(t, client.LPush(ctx, emailDeadLetterKey, "dead").Err())

	err := testutil.CollectAndCompare(NewEmailQueueCollector(client), strings.NewReader(emailQueueMetricsText),
		"goapi_email_queue_length", "goapi_email_queue_up")
	require.NoError(t, err)
}

func TestEmailQueueCollector_RedisDown_ReportsUpZeroWithoutLengths(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = client.Close() })
	mr.Close()

	expected := `
# HELP goapi_email_queue_up 1 quando a última leitura das filas de e-mail no Redis funcionou, 0 quando falhou.
# TYPE goapi_email_queue_up gauge
goapi_email_queue_up 0
`
	c := NewEmailQueueCollector(client)
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected), "goapi_email_queue_up"))
	assert.Equal(t, 1, testutil.CollectAndCount(c), "only the up series, no stale lengths")
}

func TestRegisterEmailQueueMetrics_IsIdempotent(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	require.NoError(t, RegisterEmailQueueMetrics(client))
	require.NoError(t, RegisterEmailQueueMetrics(client), "a second router setup must not fail")
	assert.Equal(t, float64(1), gaugeOrCounter(t, "goapi_email_queue_up", nil))
}

func TestRegisterEmailQueueMetrics_NilClient_NoOp(t *testing.T) {
	assert.NoError(t, RegisterEmailQueueMetrics(nil))
}
