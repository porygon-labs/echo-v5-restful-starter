package telemetry_test

import (
	"context"
	"testing"

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
