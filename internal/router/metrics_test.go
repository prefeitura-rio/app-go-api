package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
)

func TestRegisterMetricsRoute_ExposesRMIResilienceMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerMetricsRoute(r, nil)

	_ = clients.NewRMIClient("http://rmi.invalid", time.Second, clients.WithName("router-metrics-test"))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `rmi_client_circuit_breaker_state{client="router-metrics-test"} 0`)
	assert.Contains(t, body, "rmi_client_circuit_breaker_rejections_total")
	assert.Contains(t, body, "rmi_client_secretaria_cache_total")
}

// The production alerts read these series (docs/runbook.md, section 3): they
// must be served at /metrics from startup, with the e-mail queues read from Redis.
func TestRegisterMetricsRoute_ExposesAlertSignals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	_, err := mr.Lpush("email:queue:failed", "tarefa")
	assert.NoError(t, err)

	r := gin.New()
	registerMetricsRoute(r, client)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `goapi_sync_last_success_timestamp_seconds{worker="orgao_sync"}`)
	assert.Contains(t, body, `goapi_sync_last_success_timestamp_seconds{worker="citizen_sync"}`)
	assert.Contains(t, body, `goapi_sync_runs_total{result="failure",worker="orgao_sync"}`)
	assert.Contains(t, body, `goapi_import_jobs_total{result="reaped"}`)
	assert.Contains(t, body, `goapi_email_dead_letter_total{reason="retries_exhausted"}`)
	assert.Contains(t, body, "goapi_email_queue_up 1")
	assert.Contains(t, body, `goapi_email_queue_length{queue="dead_letter"} 1`)
}
