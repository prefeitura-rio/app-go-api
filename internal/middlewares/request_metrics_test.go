package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requestCount reads the sample count of goapi_http_request_duration_seconds
// for one method and route from the default registry served at /metrics.
func requestCount(t *testing.T, method, route string) uint64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	for _, mf := range families {
		if mf.GetName() != "goapi_http_request_duration_seconds" {
			continue
		}
		for _, m := range mf.GetMetric() {
			labels := map[string]string{}
			for _, lp := range m.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}
			if labels["method"] == method && labels["route"] == route {
				return m.GetHistogram().GetSampleCount()
			}
		}
	}
	return 0
}

func newMetricsEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestMetrics())
	r.GET("/api/v1/enrollments/user/:cpf", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/api/v1/courses", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })
	return r
}

func TestRequestMetrics_RecordsLatencyByRouteTemplate(t *testing.T) {
	r := newMetricsEngine()
	route := "/api/v1/enrollments/user/:cpf"
	before := requestCount(t, http.MethodGet, route)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/enrollments/user/"+accessLogCPF, nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, before+1, requestCount(t, http.MethodGet, route))
	assert.Zero(t, requestCount(t, http.MethodGet, "/api/v1/enrollments/user/"+accessLogCPF), "the raw path (with the CPF) never becomes a label")
}

func TestRequestMetrics_ErrorResponsesAreMeasuredToo(t *testing.T) {
	r := newMetricsEngine()
	before := requestCount(t, http.MethodPost, "/api/v1/courses")

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/courses", nil))

	assert.Equal(t, before+1, requestCount(t, http.MethodPost, "/api/v1/courses"))
}

func TestRequestMetrics_UnmatchedPathsShareOneRouteLabel(t *testing.T) {
	r := newMetricsEngine()
	before := requestCount(t, http.MethodGet, unmatchedRoute)

	for _, path := range []string{"/nao-existe", "/outro/" + accessLogCPF} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	assert.Equal(t, before+2, requestCount(t, http.MethodGet, unmatchedRoute))
	assert.Zero(t, requestCount(t, http.MethodGet, "/nao-existe"))
}

func TestRequestMetrics_UnknownMethodsShareOneLabel(t *testing.T) {
	r := newMetricsEngine()
	before := requestCount(t, "OTHER", unmatchedRoute)

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("XYZQ", "/nao-existe", nil))

	assert.Equal(t, before+1, requestCount(t, "OTHER", unmatchedRoute))
}
