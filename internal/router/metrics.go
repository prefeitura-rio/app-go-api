package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// registerMetricsRoute exposes the Prometheus default registry (RMI client
// circuit breaker, retries and secretaria cache metrics) at GET /metrics.
func registerMetricsRoute(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
