package observability

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

// fakeOTLPCollector is an in-process OTLP/gRPC trace collector that records
// every export request, standing in for the SigNoz collector.
type fakeOTLPCollector struct {
	coltracepb.UnimplementedTraceServiceServer

	mu       sync.Mutex
	requests []*coltracepb.ExportTraceServiceRequest
}

func (c *fakeOTLPCollector) Export(_ context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests = append(c.requests, req)
	return &coltracepb.ExportTraceServiceResponse{}, nil
}

func (c *fakeOTLPCollector) received() []*coltracepb.ExportTraceServiceRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*coltracepb.ExportTraceServiceRequest(nil), c.requests...)
}

func startFakeOTLPCollector(t *testing.T) (*fakeOTLPCollector, string) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	collector := &fakeOTLPCollector{}
	server := grpc.NewServer()
	coltracepb.RegisterTraceServiceServer(server, collector)

	go func() { _ = server.Serve(lis) }()
	t.Cleanup(server.Stop)

	return collector, lis.Addr().String()
}

func TestInitTracer_Enabled_ExportsSpansToOTLPCollector(t *testing.T) {
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
	require.NotNil(t, tracerProvider)

	_, span := otel.Tracer("tracing-export-test").Start(context.Background(), "export-check")
	span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, tracerProvider.ForceFlush(ctx))

	requests := collector.received()
	require.NotEmpty(t, requests, "collector should receive at least one export request")

	var (
		spanNames   []string
		serviceName string
		serviceVer  string
	)
	for _, req := range requests {
		for _, rs := range req.GetResourceSpans() {
			for _, attr := range rs.GetResource().GetAttributes() {
				switch attr.GetKey() {
				case "service.name":
					serviceName = attr.GetValue().GetStringValue()
				case "service.version":
					serviceVer = attr.GetValue().GetStringValue()
				}
			}
			for _, ss := range rs.GetScopeSpans() {
				for _, s := range ss.GetSpans() {
					spanNames = append(spanNames, s.GetName())
				}
			}
		}
	}

	assert.Contains(t, spanNames, "export-check")
	assert.Equal(t, "app-go-api-test", serviceName)
	assert.Equal(t, "v9.9.9", serviceVer)
}
