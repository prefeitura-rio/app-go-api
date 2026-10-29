package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// syntheticCPF é um CPF fictício usado só nos testes (nunca um dado real).
const syntheticCPF = "12345678909"

// tracedGinEngine monta um engine Gin com o otelgin ligado a um TracerProvider
// que tem o redator de url.path e um gravador de spans em memória.
func tracedGinEngine(t *testing.T) (*gin.Engine, *tracetest.SpanRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(newURLPathRedactor()),
		sdktrace.WithSpanProcessor(recorder),
	)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	r := gin.New()
	r.Use(otelgin.Middleware("app-go-api-test", otelgin.WithTracerProvider(tp)))
	r.GET("/api/v1/enrollments/user/:cpf", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/v1/empregabilidade/curriculo/:cpf/experiencias", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r, recorder
}

func spanAttr(s sdktrace.ReadOnlySpan, key attribute.Key) (attribute.Value, bool) {
	for _, kv := range s.Attributes() {
		if kv.Key == key {
			return kv.Value, true
		}
	}
	return attribute.Value{}, false
}

func assertNoAttributeContains(t *testing.T, s sdktrace.ReadOnlySpan, needle string) {
	t.Helper()
	assert.NotContains(t, s.Name(), needle, "span name must not carry the CPF")
	for _, kv := range s.Attributes() {
		assert.NotContains(t, kv.Value.Emit(), needle, "attribute %q must not carry the CPF", kv.Key)
	}
}

func TestURLPathRedactor_OtelginRouteWithCPF_RecordsRouteTemplate(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "enrollments by user cpf",
			path:     "/api/v1/enrollments/user/" + syntheticCPF,
			wantPath: "/api/v1/enrollments/user/:cpf",
		},
		{
			name:     "curriculo nested route",
			path:     "/api/v1/empregabilidade/curriculo/" + syntheticCPF + "/experiencias",
			wantPath: "/api/v1/empregabilidade/curriculo/:cpf/experiencias",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, recorder := tracedGinEngine(t)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))
			require.Equal(t, http.StatusOK, w.Code)

			spans := recorder.Ended()
			require.Len(t, spans, 1)

			got, ok := spanAttr(spans[0], "url.path")
			require.True(t, ok, "url.path should still be recorded")
			assert.Equal(t, tt.wantPath, got.AsString())

			route, ok := spanAttr(spans[0], "http.route")
			require.True(t, ok)
			assert.Equal(t, tt.wantPath, route.AsString())

			assertNoAttributeContains(t, spans[0], syntheticCPF)
		})
	}
}

func TestURLPathRedactor_OtelginUnmatchedRoute_RecordsPlaceholder(t *testing.T) {
	r, recorder := tracedGinEngine(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/rota-inexistente/"+syntheticCPF, nil))
	require.Equal(t, http.StatusNotFound, w.Code)

	spans := recorder.Ended()
	require.Len(t, spans, 1)

	got, ok := spanAttr(spans[0], "url.path")
	require.True(t, ok)
	assert.Equal(t, unmatchedRoutePath, got.AsString())
	assertNoAttributeContains(t, spans[0], syntheticCPF)
}

func TestURLPathRedactor_SpanWithoutURLPath_IsUntouched(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(newURLPathRedactor()),
		sdktrace.WithSpanProcessor(recorder),
	)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	_, span := tp.Tracer("test").Start(context.Background(), "gorm.query",
		trace.WithAttributes(attribute.String("db.system", "postgresql")))
	span.End()

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	_, ok := spanAttr(spans[0], "url.path")
	assert.False(t, ok, "redactor must not add url.path to spans that do not have it")
	dbSystem, ok := spanAttr(spans[0], "db.system")
	require.True(t, ok)
	assert.Equal(t, "postgresql", dbSystem.AsString())
}

func TestInitTracer_Enabled_ExportedSpansDoNotCarryRawURLPath(t *testing.T) {
	collector, endpoint := startFakeOTLPCollector(t)

	tracerProvider = nil
	t.Cleanup(func() {
		ShutdownTracer()
		tracerProvider = nil
	})

	cfg := &config.AppConfig{
		Tracing: config.TracingSettings{
			Enabled:        true,
			Endpoint:       endpoint,
			ServiceName:    "app-go-api-test",
			ServiceVersion: "v9.9.9",
		},
	}
	require.NoError(t, InitTracer(cfg))

	_, span := otel.Tracer("redaction-export-test").Start(context.Background(), "GET /api/v1/enrollments/user/:cpf",
		trace.WithAttributes(
			attribute.String("url.path", "/api/v1/enrollments/user/"+syntheticCPF),
			attribute.String("http.route", "/api/v1/enrollments/user/:cpf"),
		))
	span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, tracerProvider.ForceFlush(ctx))

	var urlPaths []string
	for _, req := range collector.received() {
		for _, rs := range req.GetResourceSpans() {
			for _, ss := range rs.GetScopeSpans() {
				for _, s := range ss.GetSpans() {
					for _, attr := range s.GetAttributes() {
						assert.False(t, strings.Contains(attr.GetValue().GetStringValue(), syntheticCPF),
							"exported attribute %q must not carry the CPF", attr.GetKey())
						if attr.GetKey() == "url.path" {
							urlPaths = append(urlPaths, attr.GetValue().GetStringValue())
						}
					}
				}
			}
		}
	}

	assert.Equal(t, []string{"/api/v1/enrollments/user/:cpf"}, urlPaths)
}
