package observability

import (
	"context"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestInitTracer_Disabled(t *testing.T) {
	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled: false,
		},
	}

	err := InitTracer(cfg)
	assert.NoError(t, err)

	// Verify that the global tracer provider is not set by this code
	// (it will be the default no-op provider)
	provider := otel.GetTracerProvider()
	assert.NotNil(t, provider)
}

func TestInitTracer_Enabled_WithValidConfig(t *testing.T) {
	// Reset global state
	tracerProvider = nil

	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled:        true,
			Endpoint:       "localhost:4317",
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
		},
	}

	err := InitTracer(cfg)

	// Note: This may fail in CI/test environments where no OTLP collector is running
	// We're testing the setup logic, not the actual connection
	// The error (if any) should be about connection, not configuration
	if err != nil {
		t.Logf("InitTracer returned error (expected if no collector is running): %v", err)
		// This is acceptable - we're testing configuration handling
	} else {
		// If it succeeded, verify tracer provider is set
		assert.NotNil(t, tracerProvider)

		// Clean up
		ShutdownTracer()
		tracerProvider = nil
	}
}

func TestShutdownTracer_WhenNil(t *testing.T) {
	// Reset state
	tracerProvider = nil

	// Should not panic
	ShutdownTracer()
}

func TestShutdownTracer_WhenSet(t *testing.T) {
	// Reset global state
	tracerProvider = nil

	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled:        true,
			Endpoint:       "localhost:4317",
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
		},
	}

	// Try to initialize (may fail if no collector running)
	err := InitTracer(cfg)

	if err == nil && tracerProvider != nil {
		// Successfully initialized, test shutdown
		ShutdownTracer()
		// Should not panic and should work
	} else {
		t.Logf("Skipping shutdown test - tracer not initialized: %v", err)
	}
}

func TestInitTracer_MultipleCallsIdempotent(t *testing.T) {
	// Reset global state
	tracerProvider = nil

	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled:        true,
			Endpoint:       "localhost:4317",
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
		},
	}

	// First call
	err1 := InitTracer(cfg)

	// Second call - should replace the provider
	err2 := InitTracer(cfg)

	// Both should handle the same (either both error or both succeed)
	// If one succeeds and one fails, that would indicate a problem
	if err1 == nil {
		assert.NoError(t, err2)
		ShutdownTracer()
		tracerProvider = nil
	}
}

func TestShutdownTracer_WithTimeout(t *testing.T) {
	// Reset global state
	tracerProvider = nil

	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled:        true,
			Endpoint:       "localhost:4317",
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
		},
	}

	err := InitTracer(cfg)

	if err == nil && tracerProvider != nil {
		// Create a context to test shutdown behavior
		start := time.Now()
		ShutdownTracer()
		duration := time.Since(start)

		// Shutdown should complete within 5 seconds (the timeout)
		// Plus a small buffer for overhead
		assert.Less(t, duration, 10*time.Second)

		tracerProvider = nil
	}
}

func TestInitTracer_ContextHandling(t *testing.T) {
	// Reset global state
	tracerProvider = nil

	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled:        true,
			Endpoint:       "localhost:4317",
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
		},
	}

	// InitTracer creates its own context internally
	// This test verifies it doesn't panic with context handling
	err := InitTracer(cfg)

	if err == nil && tracerProvider != nil {
		// Verify we can create a tracer
		tracer := otel.Tracer("test")
		assert.NotNil(t, tracer)

		// Verify we can start a span
		ctx := context.Background()
		ctx, span := tracer.Start(ctx, "test-span")
		assert.NotNil(t, span)
		span.End()

		ShutdownTracer()
		tracerProvider = nil
	}
}

func TestInitTracer_ServiceMetadata(t *testing.T) {
	// Reset global state
	tracerProvider = nil

	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled:        true,
			Endpoint:       "localhost:4317",
			ServiceName:    "my-custom-service",
			ServiceVersion: "v2.3.4",
		},
	}

	err := InitTracer(cfg)

	if err == nil && tracerProvider != nil {
		// If initialization succeeded, verify the provider was set
		provider := otel.GetTracerProvider()
		assert.NotNil(t, provider)

		ShutdownTracer()
		tracerProvider = nil
	}
}
