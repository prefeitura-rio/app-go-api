package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
)

func TestRegisterMetricsRoute_ExposesRMIResilienceMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerMetricsRoute(r)

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
