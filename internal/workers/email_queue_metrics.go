package workers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// emailQueueReadTimeout bounds the Redis reads of one /metrics scrape.
const emailQueueReadTimeout = 2 * time.Second

var (
	emailQueueLengthDesc = prometheus.NewDesc(
		"goapi_email_queue_length",
		"Tarefas de e-mail paradas em cada fila do Redis: pending (email:queue), delayed (email:queue:delayed, aguardando nova tentativa) e dead_letter (email:queue:failed).",
		[]string{"queue"}, nil,
	)
	emailQueueUpDesc = prometheus.NewDesc(
		"goapi_email_queue_up",
		"1 quando a última leitura das filas de e-mail no Redis funcionou, 0 quando falhou.",
		nil, nil,
	)
)

// EmailQueueCollector reports the length of the e-mail queues on each scrape.
// The queues are shared in Redis, so every pod reports the same values:
// aggregate with max() across pods.
type EmailQueueCollector struct {
	redis *redis.Client
}

// NewEmailQueueCollector builds a collector over the e-mail queues in redisClient.
func NewEmailQueueCollector(redisClient *redis.Client) *EmailQueueCollector {
	return &EmailQueueCollector{redis: redisClient}
}

// Describe implements prometheus.Collector.
func (c *EmailQueueCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- emailQueueLengthDesc
	ch <- emailQueueUpDesc
}

// Collect implements prometheus.Collector. When Redis cannot be read it
// reports goapi_email_queue_up 0 and no lengths, never stale values.
func (c *EmailQueueCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), emailQueueReadTimeout)
	defer cancel()

	pipe := c.redis.Pipeline()
	pending := pipe.LLen(ctx, emailQueueKey)
	delayed := pipe.ZCard(ctx, emailDelayedQueueKey)
	dead := pipe.LLen(ctx, emailDeadLetterKey)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("[EmailWorker] failed to read queue lengths for metrics: %v", err)
		ch <- prometheus.MustNewConstMetric(emailQueueUpDesc, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(emailQueueUpDesc, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(emailQueueLengthDesc, prometheus.GaugeValue, float64(pending.Val()), "pending")
	ch <- prometheus.MustNewConstMetric(emailQueueLengthDesc, prometheus.GaugeValue, float64(delayed.Val()), "delayed")
	ch <- prometheus.MustNewConstMetric(emailQueueLengthDesc, prometheus.GaugeValue, float64(dead.Val()), "dead_letter")
}

// RegisterEmailQueueMetrics registers the e-mail queue collector in the
// default registry served at /metrics. Registering again is a no-op, and a nil
// client registers nothing.
func RegisterEmailQueueMetrics(redisClient *redis.Client) error {
	if redisClient == nil {
		return nil
	}
	err := prometheus.Register(NewEmailQueueCollector(redisClient))
	var already prometheus.AlreadyRegisteredError
	if errors.As(err, &already) {
		return nil
	}
	return err
}
