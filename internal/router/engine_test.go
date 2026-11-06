package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewEngine_AccessLog_DoesNotLogCPF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var out bytes.Buffer
	previous := gin.DefaultWriter
	gin.DefaultWriter = &out
	t.Cleanup(func() { gin.DefaultWriter = previous })

	const syntheticCPF = "00011122233"
	r := newEngine()
	r.GET("/api/v1/courses/:courseId/enrollments", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/enrollments/user/:cpf", func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, target := range []string{
		"/api/v1/courses/2882/enrollments?search=" + syntheticCPF + "&limit=1",
		"/api/v1/enrollments/user/" + syntheticCPF,
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		assert.Equal(t, http.StatusOK, w.Code)
	}

	log := out.String()
	assert.Contains(t, log, "[GIN]")
	assert.Contains(t, log, "/api/v1/courses/:courseId/enrollments?limit=1&search=[omitido]")
	assert.Contains(t, log, "/api/v1/enrollments/user/:cpf")
	assert.NotContains(t, log, syntheticCPF)
}

func TestNewEngine_RecoversFromPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousOut, previousErr := gin.DefaultWriter, gin.DefaultErrorWriter
	gin.DefaultWriter, gin.DefaultErrorWriter = &bytes.Buffer{}, &bytes.Buffer{}
	t.Cleanup(func() { gin.DefaultWriter, gin.DefaultErrorWriter = previousOut, previousErr })

	r := newEngine()
	r.GET("/panic", func(c *gin.Context) { panic("falha simulada") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/panic", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Every request, panics included, feeds the per-endpoint latency histogram the
// production latency alert reads.
func TestNewEngine_RecordsRequestLatencyByRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousOut, previousErr := gin.DefaultWriter, gin.DefaultErrorWriter
	gin.DefaultWriter, gin.DefaultErrorWriter = &bytes.Buffer{}, &bytes.Buffer{}
	t.Cleanup(func() { gin.DefaultWriter, gin.DefaultErrorWriter = previousOut, previousErr })

	r := newEngine()
	r.GET("/engine-metrics/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/engine-metrics-panic", func(c *gin.Context) { panic("falha simulada") })
	registerMetricsRoute(r, nil)

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/engine-metrics/1", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/engine-metrics-panic", nil))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := w.Body.String()
	assert.Contains(t, body, `goapi_http_request_duration_seconds_count{method="GET",route="/engine-metrics/:id"} 1`)
	assert.Contains(t, body, `goapi_http_request_duration_seconds_count{method="GET",route="/engine-metrics-panic"} 1`)
}
