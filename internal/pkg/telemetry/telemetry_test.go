package telemetry_test

import (
	"context"
	"net"
	"testing"
	"time"

	"go-restful-api/internal/pkg/telemetry"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Helper function to simulate calling StartSpan from a method/function
func dummyServiceMethod(ctx context.Context) {
	_, end := telemetry.StartSpan(ctx)
	defer end()
}

func TestStartSpan_AutoCapturesFunctionName(t *testing.T) {
	// 1. Setup in-memory exporter (captures spans in memory instead of sending over network)
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// 2. Call the function that invokes StartSpan
	ctx := context.Background()
	dummyServiceMethod(ctx)

	// 3. Retrieve captured spans
	spans := exporter.GetSpans()

	// Assertions
	if len(spans) != 1 {
		t.Fatalf("expected 1 captured span, got %d", len(spans))
	}

	span := spans[0]
	expectedName := "telemetry_test.dummyServiceMethod"

	t.Logf("Captured Span Name: %s", span.Name)

	if span.Name != expectedName {
		t.Errorf("expected span name %q, got %q", expectedName, span.Name)
	}
}

func TestStartSpan_NestedSpans(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)

	ctx := context.Background()

	// Create a parent span, then a child span via StartSpan.
	// Must end parent before GetSpans (defer is too late).
	ctx, end := telemetry.StartSpan(ctx)
	dummyServiceMethod(ctx)
	end()

	spans := exporter.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("expected 2 captured spans, got %d", len(spans))
	}

	// Find parent and child by name (order is not guaranteed).
	var parent, child tracetest.SpanStub
	for _, s := range spans {
		if s.Name == "telemetry_test.dummyServiceMethod" {
			child = s
		} else {
			parent = s
		}
	}

	if child.Parent.SpanID() != parent.SpanContext.SpanID() {
		t.Errorf("child.Parent.SpanID() = %s, parent.SpanContext.SpanID() = %s",
			child.Parent.SpanID(), parent.SpanContext.SpanID())
	}
}

func TestInit_Success(t *testing.T) {
	// Start a dummy TCP listener that accepts connections (OTLP will
	// connect lazily, so this ensures Init doesn't fail at creation time).
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	shutdown, err := telemetry.Init(context.Background(), "test-svc", ln.Addr().String())
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if shutdown == nil {
		t.Fatal("Init() returned nil shutdown")
	}

	// Verify the global tracer provider was set (not nil).
	if tp := otel.GetTracerProvider(); tp == nil {
		t.Error("global TracerProvider is nil after Init")
	}

	// Shutdown should succeed (even with a dummy endpoint that didn't
	// receive data).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Fatalf("shutdown() error = %v", err)
	}
}

func TestInit_InvalidEndpoint(t *testing.T) {
	// gRPC connections are lazy; Init with a garbage endpoint still
	// succeeds and returns a valid shutdown function.
	shutdown, err := telemetry.Init(context.Background(), "test-svc", "127.0.0.1:1")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if shutdown == nil {
		t.Fatal("Init() returned nil shutdown")
	}

	// Shutdown should not panic.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = shutdown(ctx) // may error, which is fine
}

func TestStartSpan_FallbackName(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)

	_, end := telemetry.StartSpan(context.Background())
	end() // must end before GetSpans

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	// Verify the span name is not "unknown" — it should be the test
	// function name.
	if spans[0].Name == "unknown" {
		t.Error("span name is 'unknown', expected caller function name")
	}
}
