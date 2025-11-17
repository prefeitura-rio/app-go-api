package observability

import (
	"context"
	"log"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	tracerProvider *sdktrace.TracerProvider
)

// InitTracer initializes the OpenTelemetry tracer with gRPC OTLP exporter
func InitTracer(cfg *config.AppConfig) error {
	if !cfg.Tracing.Enabled {
		log.Println("OpenTelemetry tracing is disabled")
		return nil
	}

	ctx := context.Background()

	// Create OTLP gRPC exporter
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Tracing.Endpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return err
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.Tracing.ServiceName),
			semconv.ServiceVersionKey.String(cfg.Tracing.ServiceVersion),
		),
	)
	if err != nil {
		return err
	}

	// Create trace provider with batching
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(time.Second*10),
			sdktrace.WithMaxQueueSize(2048),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Printf("OpenTelemetry tracer initialized (endpoint: %s, service: %s)",
		cfg.Tracing.Endpoint, cfg.Tracing.ServiceName)

	return nil
}

// ShutdownTracer gracefully shuts down the tracer provider
func ShutdownTracer() {
	if tracerProvider == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := tracerProvider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	} else {
		log.Println("Tracer provider shut down successfully")
	}
}
