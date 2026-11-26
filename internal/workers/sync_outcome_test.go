package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// The sync alerts (GoApiSyncStale, GoApiSyncFailing) read the cycle outcome:
// a cycle in which the RMI answered nothing usable must not count as a success,
// and the failed-retry backoff must not hide it in the following cycles.

// fakeRMIDepartments answers /v1/departments/<id>: "ok-*" with an orgao,
// "missing-*" with 404 and anything else with 503 (RMI unavailable).
func fakeRMIDepartments(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		switch {
		case strings.HasPrefix(id, "ok-"):
			_ = json.NewEncoder(w).Encode(models.Orgao{ID: id, CdUA: id, NomeUA: "Orgao Sintetico " + id})
		case strings.HasPrefix(id, "missing-"):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newOrgaoWorkerWithRMI(t *testing.T, db *gorm.DB, redisClient *redis.Client, orgaoIDs ...string) *OrgaoSyncWorker {
	t.Helper()
	for _, id := range orgaoIDs {
		require.NoError(t, db.Exec(`INSERT INTO oportunidades_mei (titulo, orgao_id) VALUES (?, ?)`, "Oportunidade sintetica", id).Error)
	}
	rmi := clients.NewRMIClient(fakeRMIDepartments(t).URL, 5*time.Second,
		clients.WithName(fmt.Sprintf("orgao-sync-%s-%d", t.Name(), time.Now().UnixNano())),
		clients.WithRetryBackoff(time.Millisecond, 2*time.Millisecond))
	return NewOrgaoSyncWorker(db, rmi, redisClient, repository.NewOrgaoSnapshotRepository(db), nil, nil, nil,
		&config.OrgaoSyncSettings{SyncInterval: time.Minute, StaleThreshold: time.Hour, FailedRetryThreshold: time.Hour, BatchSize: 10, MaxRetries: 1})
}

func orgaoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&models.OrgaoSnapshot{}))
	return db
}

type syncCounters struct{ success, failure, backoff, last float64 }

func readSyncCounters(t *testing.T, worker string) syncCounters {
	return syncCounters{
		success: syncRuns(t, worker, "success"),
		failure: syncRuns(t, worker, "failure"),
		backoff: syncRuns(t, worker, "backoff"),
		last:    syncLastSuccess(t, worker),
	}
}

func TestOrgaoSyncWorker_runSync_RMIUnavailable_RecordsFailureKeepsLastSuccess(t *testing.T) {
	// 7 orgaos all answered 503: the RMI client circuit opens after 5 failures,
	// so the last ones fail with the circuit open, which is also a failure.
	ids := []string{"down-1", "down-2", "down-3", "down-4", "down-5", "down-6", "down-7"}
	w := newOrgaoWorkerWithRMI(t, orgaoTestDB(t), nil, ids...)
	before := readSyncCounters(t, "orgao_sync")

	err := w.runSync(context.Background())

	require.Error(t, err, "a cycle in which every orgao failed is a failed cycle")
	after := readSyncCounters(t, "orgao_sync")
	assert.Equal(t, before.failure+1, after.failure)
	assert.Equal(t, before.success, after.success)
	assert.Equal(t, before.last, after.last, "the last success timestamp must not move")
}

func TestOrgaoSyncWorker_runSync_BackoffAfterRMIUnavailable_IsNotASuccess(t *testing.T) {
	w := newOrgaoWorkerWithRMI(t, orgaoTestDB(t), nil, "down-1", "down-2")
	require.Error(t, w.runSync(context.Background()))
	before := readSyncCounters(t, "orgao_sync")

	// Every orgao is now in the failed-retry backoff: nothing to try this cycle.
	require.NoError(t, w.runSync(context.Background()))

	after := readSyncCounters(t, "orgao_sync")
	assert.Equal(t, before.backoff+1, after.backoff)
	assert.Equal(t, before.success, after.success, "a cycle held back by the backoff after an RMI outage is not a success")
	assert.Equal(t, before.last, after.last)
}

func TestOrgaoSyncWorker_runSync_BackoffAfterRMIUnavailable_SharedAcrossPods(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	db := orgaoTestDB(t)

	podA := newOrgaoWorkerWithRMI(t, db, client, "down-1")
	require.Error(t, podA.runSync(context.Background()))
	before := readSyncCounters(t, "orgao_sync")

	// Another pod takes the next cycle (the lock is per cycle).
	podB := newOrgaoWorkerWithRMI(t, db, client)
	require.NoError(t, podB.runSync(context.Background()))

	after := readSyncCounters(t, "orgao_sync")
	assert.Equal(t, before.backoff+1, after.backoff)
	assert.Equal(t, before.success, after.success)
	assert.Equal(t, before.last, after.last)
}

func TestOrgaoSyncWorker_runSync_PartialFailure_IsASuccess(t *testing.T) {
	w := newOrgaoWorkerWithRMI(t, orgaoTestDB(t), nil, "ok-1", "down-1")
	before := readSyncCounters(t, "orgao_sync")
	start := float64(time.Now().Unix())

	require.NoError(t, w.runSync(context.Background()))

	after := readSyncCounters(t, "orgao_sync")
	assert.Equal(t, before.success+1, after.success)
	assert.GreaterOrEqual(t, after.last, start)
}

func TestOrgaoSyncWorker_runSync_SuccessClearsTheOutage(t *testing.T) {
	db := orgaoTestDB(t)
	w := newOrgaoWorkerWithRMI(t, db, nil, "down-1")
	require.Error(t, w.runSync(context.Background()))

	// A new orgao syncs fine: the RMI is back, and the next cycles, with
	// down-1 still in its backoff, are successes again.
	require.NoError(t, db.Exec(`INSERT INTO oportunidades_mei (titulo, orgao_id) VALUES (?, ?)`, "Oportunidade sintetica", "ok-1").Error)
	require.NoError(t, w.runSync(context.Background()))
	before := readSyncCounters(t, "orgao_sync")

	require.NoError(t, w.runSync(context.Background()))

	after := readSyncCounters(t, "orgao_sync")
	assert.Equal(t, before.success+1, after.success)
	assert.Equal(t, before.backoff, after.backoff)
}

func TestOrgaoSyncWorker_runSync_OrgaoNotFound_IsNotAnOutage(t *testing.T) {
	// An orgao the RMI does not know (404) is a data problem of one item, not
	// an RMI outage: the cycle and the backoff cycles after it stay successes.
	w := newOrgaoWorkerWithRMI(t, orgaoTestDB(t), nil, "missing-1")
	before := readSyncCounters(t, "orgao_sync")
	skippedItems := func() float64 {
		return gaugeOrCounter(t, "goapi_sync_items_total", map[string]string{"worker": "orgao_sync", "result": "skipped"})
	}
	skippedBefore := skippedItems()

	require.NoError(t, w.runSync(context.Background()))
	require.NoError(t, w.runSync(context.Background()))

	after := readSyncCounters(t, "orgao_sync")
	assert.Equal(t, before.success+2, after.success)
	assert.Equal(t, before.failure, after.failure)
	assert.Equal(t, before.backoff, after.backoff)
	assert.Equal(t, skippedBefore+1, skippedItems(), "the 404 is counted as a skipped item")
}

// --- citizen sync ---

func newCitizenWorkerForRunSync(t *testing.T, cpfs ...string) *CitizenSyncWorker {
	t.Helper()
	w, db, _, _ := newCitizenSyncTestWorkerWithDB(t)
	for _, stmt := range []string{
		`CREATE TABLE inscricoes (cpf VARCHAR(11))`,
		`CREATE TABLE emp_candidaturas (cpf VARCHAR(11), deleted_at DATETIME)`,
		`CREATE TABLE emp_curriculos (cpf VARCHAR(11))`,
	} {
		require.NoError(t, db.Exec(stmt).Error)
	}
	for _, cpf := range cpfs {
		require.NoError(t, db.Exec(`INSERT INTO inscricoes (cpf) VALUES (?)`, cpf).Error)
	}
	keycloak := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(rw).Encode(auth.TokenResponse{AccessToken: "token-sintetico", ExpiresIn: 300})
	}))
	t.Cleanup(keycloak.Close)
	w.tokenManager = auth.NewServiceAccountTokenManager(keycloak.URL, "realm", "client", "secret")
	w.staleThreshold = time.Hour
	w.batchSize = 10
	return w
}

func TestCitizenSyncWorker_runSync_RMIUnavailable_RecordsFailureKeepsLastSuccess(t *testing.T) {
	w := newCitizenWorkerForRunSync(t, syncTestCPFUnavailable)
	before := readSyncCounters(t, "citizen_sync")

	require.Error(t, w.runSync(context.Background()))

	after := readSyncCounters(t, "citizen_sync")
	assert.Equal(t, before.failure+1, after.failure)
	assert.Equal(t, before.success, after.success)
	assert.Equal(t, before.last, after.last)
}

func TestCitizenSyncWorker_runSync_PartialFailure_IsASuccess(t *testing.T) {
	w := newCitizenWorkerForRunSync(t, syncTestCPFOK, syncTestCPFUnavailable)
	before := readSyncCounters(t, "citizen_sync")
	start := float64(time.Now().Unix())

	require.NoError(t, w.runSync(context.Background()))

	after := readSyncCounters(t, "citizen_sync")
	assert.Equal(t, before.success+1, after.success)
	assert.GreaterOrEqual(t, after.last, start)
}

func TestCitizenSyncWorker_runSync_OnlyExpectedResults_IsASuccess(t *testing.T) {
	w := newCitizenWorkerForRunSync(t, syncTestCPFNotFound, syncTestCPFInvalid)
	before := readSyncCounters(t, "citizen_sync")

	require.NoError(t, w.runSync(context.Background()))

	after := readSyncCounters(t, "citizen_sync")
	assert.Equal(t, before.success+1, after.success)
	assert.Equal(t, before.failure, after.failure)
}
