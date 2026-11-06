package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// otherMethod groups the HTTP methods outside the standard set, so a client
// cannot create series with arbitrary method names.
const otherMethod = "OTHER"

// httpRequestDuration feeds the per-endpoint latency alert (docs/runbook.md,
// section 3). The route label is the Gin route template, never the raw path,
// which may carry a CPF; unmatched paths share one label.
var httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "goapi_http_request_duration_seconds",
	Help:    "Duração das requisições HTTP por método e template da rota.",
	Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
}, []string{"method", "route"})

var standardMethods = map[string]struct{}{
	http.MethodGet: {}, http.MethodHead: {}, http.MethodPost: {}, http.MethodPut: {},
	http.MethodPatch: {}, http.MethodDelete: {}, http.MethodOptions: {},
}

// RequestMetrics records the latency of every request in
// goapi_http_request_duration_seconds.
func RequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = unmatchedRoute
		}
		method := c.Request.Method
		if _, ok := standardMethods[method]; !ok {
			method = otherMethod
		}
		httpRequestDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
	}
}
