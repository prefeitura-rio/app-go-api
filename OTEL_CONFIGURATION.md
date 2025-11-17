# OpenTelemetry Configuration Guide

This API now has full OpenTelemetry instrumentation for distributed tracing.

## Architecture

The implementation follows the same pattern as `app-rmi`:
- **gRPC OTLP Exporter** - Sends traces to OpenTelemetry Collector
- **Automatic Instrumentation** for:
  - HTTP requests (Gin middleware)
  - Database queries (GORM/PostgreSQL)
  - Redis operations
  - HTTP client calls

## Configuration

### Environment Variables

Add these to your `.env` file or Kubernetes secrets:

```bash
# Enable/disable tracing
TRACING_ENABLED=true

# OTLP gRPC endpoint (no http:// prefix)
TRACING_ENDPOINT=localhost:4317

# Service identification
TRACING_SERVICE_NAME=app-go-api
TRACING_SERVICE_VERSION=v1.0.0
```

### Local Development

For local development with Docker Compose:

```yaml
services:
  otel-collector:
    image: otel/opentelemetry-collector:latest
    ports:
      - "4317:4317"  # OTLP gRPC receiver
      - "4318:4318"  # OTLP HTTP receiver
    volumes:
      - ./otel-collector-config.yaml:/etc/otel/config.yaml
    command: ["--config", "/etc/otel/config.yaml"]
```

Example `otel-collector-config.yaml`:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:

exporters:
  logging:
    loglevel: debug
  # Add your backend exporter (Jaeger, Tempo, etc.)

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
```

### Kubernetes

The K8s manifests (`k8s/staging/resources.yaml` and `k8s/prod/resources.yaml`) are already configured with:

```yaml
env:
  - name: TRACING_ENABLED
    value: "true"
  - name: TRACING_ENDPOINT
    value: "opentelemetry-collector.observability.svc.cluster.local:4317"
  - name: TRACING_SERVICE_NAME
    value: "app-go-api"
  - name: TRACING_SERVICE_VERSION
    value: "v1.0.0"
```

**Important:** Make sure your Kubernetes cluster has an OpenTelemetry Collector deployed in the `observability` namespace at the service name `opentelemetry-collector`.

## What Gets Traced

### HTTP Requests
- Route path, method, status code
- Request duration
- Request/response headers (configurable)
- Span name: `{HTTP_METHOD} {route_pattern}`

### Database Queries
- SQL statements
- Query duration
- Database name, table name
- Span name: `{operation} {table}`

### Redis Operations
- Command name (GET, SET, etc.)
- Key names
- Operation duration
- Span name: `redis.{command}`

### External HTTP Calls
If you use instrumented HTTP clients, outgoing requests to external APIs (like RMI) will also be traced.

## Trace Context Propagation

The implementation uses W3C Trace Context propagation, which means:
- Traces are propagated across service boundaries via HTTP headers
- Compatible with Istio, Envoy, and other service meshes
- Supports distributed tracing across microservices

## Performance Impact

- **Minimal overhead**: Batching and async export
- **Configurable sampling**: Currently set to `AlwaysSample()` (trace everything)
- **Graceful degradation**: If collector is unavailable, traces are buffered and retried

## Disabling Tracing

Set `TRACING_ENABLED=false` to disable all tracing overhead.

## Troubleshooting

### Traces not appearing

1. Check that `TRACING_ENABLED=true`
2. Verify OTLP collector is reachable at `TRACING_ENDPOINT`
3. Check application logs for tracer initialization messages:
   ```
   OpenTelemetry tracer initialized (endpoint: ..., service: ...)
   GORM OpenTelemetry instrumentation enabled
   ```

### Connection refused errors

- Ensure collector endpoint doesn't include `http://` prefix
- Default port is `4317` for gRPC
- Use `localhost:4317` for local dev, not `0.0.0.0:4317`

### High cardinality warnings

If you see warnings about high cardinality attributes, consider:
- Adding sampling (replace `AlwaysSample()` with `TraceIDRatioBased()`)
- Filtering sensitive attributes
- Limiting span attributes

## Next Steps

1. **Configure your backend**: Send traces to Jaeger, Tempo, Honeycomb, etc.
2. **Add custom spans**: Use `otel.Tracer()` to create manual spans for business logic
3. **Add metrics**: Extend with OpenTelemetry metrics if needed
4. **Set up alerts**: Configure alerts based on trace data (error rates, latency, etc.)
