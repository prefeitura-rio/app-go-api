# OpenTelemetry / Observability

Distributed tracing via OpenTelemetry (OTLP/gRPC), following the same pattern as `app-rmi`.

## What is instrumented

- HTTP requests (Gin middleware)
- Database queries (GORM/PostgreSQL)
- Redis operations
- HTTP client calls (when using instrumented clients)

## Configuration

Environment variables (see `.env.example`):

```bash
TRACING_ENABLED=true
TRACING_ENDPOINT=localhost:4317          # host:port only — no http:// prefix
TRACING_SERVICE_NAME=app-go-api
TRACING_SERVICE_VERSION=v1.0.0
```

Set `TRACING_ENABLED=false` to disable tracing.

### Kubernetes

Staging and production manifests set tracing against SigNoz:

```yaml
TRACING_ENABLED: "true"
TRACING_ENDPOINT: "signoz-otel-collector.signoz.svc.cluster.local:4317"
TRACING_SERVICE_NAME: "app-go-api"
TRACING_SERVICE_VERSION: "v1.0.0"
```

### Local development

Point `TRACING_ENDPOINT` at a local OTLP collector (default gRPC port `4317`). Example Compose service:

```yaml
services:
  otel-collector:
    image: otel/opentelemetry-collector:latest
    ports:
      - "4317:4317"
      - "4318:4318"
    volumes:
      - ./otel-collector-config.yaml:/etc/otel/config.yaml
    command: ["--config", "/etc/otel/config.yaml"]
```

## Trace context

Uses W3C Trace Context headers, compatible with Istio/Envoy and cross-service propagation.

## Troubleshooting

1. Confirm `TRACING_ENABLED=true` and that `TRACING_ENDPOINT` is reachable.
2. Look for startup logs such as `OpenTelemetry tracer initialized` / `GORM OpenTelemetry instrumentation enabled`.
3. Connection refused: endpoint must be `host:port` (no `http://`); use `localhost:4317` locally, not `0.0.0.0:4317`.

## Metrics (`/metrics`)

`GET /metrics` serves the Prometheus default registry: Go/process metrics, the RMI client resilience
metrics (`rmi_client_*`, see `docs/runbook.md` section 2) and the per-route request latency:

| Metric | Labels | Meaning |
|---|---|---|
| `goapi_http_request_duration_seconds` (histogram, buckets 0.05 … 10 s) | `method`, `route` | Duration of every request. `route` is the Gin route template (`/api/v1/enrollments/user/:cpf`), never the raw path, which may carry a CPF; a path with no route is `(rota não encontrada)` and a non-standard method is `OTHER` |
| `rmi_client_circuit_breaker_state` (gauge) | `client` (`api`, `orgao_sync`) | `0` closed, `1` half-open, `2` open |
| `goapi_sync_last_success_timestamp_seconds` (gauge) | `worker` (`orgao_sync`, `citizen_sync`) | Unix time of the last successful sync cycle in this pod (`0` = none yet): no error, and some item synced or none failed; aggregate with `max` across pods |
| `goapi_sync_runs_total` (counter) | `worker`, `result` (`success`, `failure`, `skipped`, `backoff`) | Sync cycles by outcome. `failure` = an error, or every item tried failed (RMI down, circuit open); `skipped` = another pod held the lock; `backoff` = nothing to try because the orgaos that failed in an RMI outage wait for their retry (`ORGAO_SYNC_FAILED_RETRY_THRESHOLD`). Neither `failure` nor `backoff` moves the last success timestamp |
| `goapi_sync_items_total` (counter) | `worker`, `result` (`synced`, `skipped`, `failed`) | Items handled by the sync cycles; `skipped` = the RMI answered about the item itself (citizen or orgao unknown, invalid id), not an outage |
| `goapi_import_jobs_total` (counter) | `result` (`completed`, `failed`, `reaped`) | Finished import jobs; `reaped` = closed as failed by the stale job reaper because its pod stopped |
| `goapi_email_queue_length` (gauge, read from Redis on each scrape) | `queue` (`pending`, `delayed`, `dead_letter`) | Tasks in `email:queue`, `email:queue:delayed` and `email:queue:failed`; the same in every pod, aggregate with `max` |
| `goapi_email_queue_up` (gauge) | — | `1` when the last read of the e-mail queues worked, `0` when Redis could not be read (no lengths are reported then) |
| `goapi_email_dead_letter_total` (counter) | `reason` (`panic`, `retries_exhausted`) | E-mail tasks moved to the dead-letter queue |

The `go` Rollout (staging and production) carries the `prometheus.io/scrape|port|path` annotations
(port `8080`, path `/metrics`). The Istio sidecar injector rewrites them to the agent port
(`15020/stats/prometheus`) and merges the application metrics with the Envoy ones, so the
`kubernetes-pods` job of the Prometheus in `istio-system` collects both, with the `namespace`, `pod`
and `app` target labels. From the public entry point `/metrics` answers 403 (`ext_authz`).

## Production alerts

The production alert rules live in [`monitoring/prometheus/alerting-rules.yaml`](../monitoring/prometheus/alerting-rules.yaml).
This repository only versions them; installing them in the monitoring stack and routing the
notifications belongs to the Infra lane (`prefeitura-rio/infra`).

The rules that run today in the Prometheus of `istio-system` (staging and production) are a separate
set, versioned in `prefeitura-rio/infra` at `superapp/modules/k8s/yamls/istio/prometheus-alerting-rules.yaml`
(OPS-151): 5xx and p95 of the service (`istio_*`), p95 per route, sync stale or failing, import jobs
reaped, scrape target down, and the e-mail queue stuck, unreadable or with a growing dead-letter
queue, over the `goapi_*` series above. They show in `/api/v1/rules` and `/alerts`; the Alertmanager
is still off, so they notify no one yet. `docs/runbook.md` section 3 lists both sets.

### Format

A plain Prometheus [rule file](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/):
a top-level `groups:` list with one group, `app-go-api`, of alerting rules (`alert`, `expr`, `for`,
`labels`, `annotations`). It is **not** a `PrometheusRule` resource and is not listed in any
`k8s/*/kustomization.yaml`: the clusters run the `prometheus` community Helm chart, without the
Prometheus Operator CRDs, so Argo CD must not apply it.

- Labels on every alert: `severity` (`critical` or `warning`) and `service: app-go-api`, for routing.
- Annotations: `summary`, `description` (with the value and the `route`/`pod`/`client` label) and
  `runbook_url` (a section of `docs/runbook.md`).
- The same file serves staging and production: the namespace and the workload are both `go` in each
  cluster, and each cluster has its own Prometheus.

| Alert | Fires when | Source |
|---|---|---|
| `GoApiHighErrorRate` (critical) | 5xx above 2% of the responses of the `go` workload for 5 min, with at least 0.1 req/s | `istio_requests_total` (destination sidecar) |
| `GoApiHighLatencyP95` (warning) | p95 of one `/api/v1/…` or `/api/public/…` route above 1 s for 10 min, with at least 0.05 req/s on that route | `goapi_http_request_duration_seconds` |
| `GoApiPodRestarting` (critical) | more than 3 restarts of the `go` container of one pod in 15 min (CrashLoopBackOff, OOMKilled, liveness) | `kube_pod_container_status_restarts_total` (kube-state-metrics) |
| `GoApiRMICircuitBreakerOpen` (warning) | the RMI client circuit of one pod is not closed (open or half-open) for 5 min, i.e. no successful RMI call in that time | `rmi_client_circuit_breaker_state` |

### Installing (Infra lane)

With the `prometheus` chart (`helm_release.prometheus` in `superapp/modules/k8s/istio.tf`), the
server already reads `/etc/config/alerting_rules.yml`; the file content goes, as is, under
`serverFiles."alerting_rules.yml"` (a `groups:` map), for example
`serverFiles = { "alerting_rules.yml" = yamldecode(file("<copy of alerting-rules.yaml>")) }`.
With the Prometheus Operator, the same `groups:` list is the `spec.groups` of a `PrometheusRule`.
After loading, the rules show in `/api/v1/rules`.

Prerequisites on the stack side:

- **kube-state-metrics** for `GoApiPodRestarting`: the chart runs today with
  `kube-state-metrics = { enabled = false }`; without it the alert has no data and never fires.
- **Alertmanager** (or another receiver) for the notifications: the chart runs with
  `alertmanager = { enabled = false }`; without it the alerts only show in `/alerts`.
- `istio_requests_total` is already collected (KEDA and the canary analysis read it); the
  `goapi_*` and `rmi_client_*` series exist after the deploy that adds the scrape annotations.

### Changing the rules

Validate with `promtool check rules monitoring/prometheus/alerting-rules.yaml` and
`promtool test rules monitoring/prometheus/alerting-rules.test.yaml` (synthetic series that must,
and must not, fire each alert; promtool ships with the Prometheus release). The test
`internal/router/alerting_rules_test.go` (`just test`) checks the file format, the four alerts and
their thresholds, and that every `goapi_*`/`rmi_client_*` metric a rule reads is served at
`/metrics`. A new alert also gets a row in `docs/runbook.md` section 3.
