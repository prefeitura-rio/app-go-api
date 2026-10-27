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
