package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/workers"
)

// registerMetricsRoute exposes the Prometheus default registry at GET /metrics:
// RMI client circuit breaker, retries and secretaria cache metrics, and the
// goapi_* signals the production alerts read (syncs, import jobs and the
// e-mail queues, these read from redisClient on each scrape).
func registerMetricsRoute(r *gin.Engine, redisClient *redis.Client) {
	if err := workers.RegisterEmailQueueMetrics(redisClient); err != nil {
		log.Printf("[Router] failed to register e-mail queue metrics: %v", err)
	}
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
