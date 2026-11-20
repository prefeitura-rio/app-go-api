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

The `go` Service (staging and production; the active/stable one, not `go-preview`/`go-canary`)
carries the `prometheus.io/scrape|port|path` annotations (port `8080`, path `/metrics`), so the
`kubernetes-service-endpoints` job of the Prometheus in `istio-system` scrapes every ready pod behind
it, with the `namespace`, `service`, `app` and `instance` (pod `IP:8080`) target labels; that job
sets no `pod` label. The annotations are not on the Rollout pod template on purpose: the Argo CD
Application `go` (`superapp/modules/k8s/yamls/argocd/applications/go.yaml` in `prefeitura-rio/infra`)
ignores `/spec/template/metadata/annotations` of the Rollout with `RespectIgnoreDifferences`, so a
change there never reaches the pods (and, on a fresh Rollout, the Istio metrics merge would scrape
the same series a second time through the `kubernetes-pods` job). `internal/router/scrape_manifests_test.go`
keeps it that way. From the public entry point `/metrics` answers 403 (`ext_authz`).

## Production alerts

The production alert rules live in [`monitoring/prometheus/alerting-rules.yaml`](../monitoring/prometheus/alerting-rules.yaml).
This repository only versions them; installing them in the monitoring stack and routing the
notifications belongs to the Infra lane (`prefeitura-rio/infra`).

### Format

A plain Prometheus [rule file](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/):
a top-level `groups:` list with one group, `app-go-api`, of alerting rules (`alert`, `expr`, `for`,
`labels`, `annotations`). It is **not** a `PrometheusRule` resource and is not listed in any
`k8s/*/kustomization.yaml`: the clusters run the `prometheus` community Helm chart, without the
Prometheus Operator CRDs, so Argo CD must not apply it.

- Labels on every alert: `severity` (`critical` or `warning`) and `service: app-go-api`, for routing.
- Annotations: `summary`, `description` (with the value and the `route`, `pod` (kube-state-metrics), `instance` or `client` label) and
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
  `goapi_*` and `rmi_client_*` series exist once the `go` Service carries the scrape annotations
  (job `kubernetes-service-endpoints`, already in the chart's default scrape config).

### Changing the rules

Validate with `promtool check rules monitoring/prometheus/alerting-rules.yaml` and
`promtool test rules monitoring/prometheus/alerting-rules.test.yaml` (synthetic series that must,
and must not, fire each alert; promtool ships with the Prometheus release). The test
`internal/router/alerting_rules_test.go` (`just test`) checks the file format, the four alerts and
their thresholds, and that every `goapi_*`/`rmi_client_*` metric a rule reads is served at
`/metrics`. A new alert also gets a row in `docs/runbook.md` section 3.
