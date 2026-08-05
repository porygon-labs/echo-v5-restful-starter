// Package telemetry provides OpenTelemetry instrumentation helpers.
package telemetry

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

var defaultTracerName = "app"

// Init creates an OTLP gRPC exporter, builds a TracerProvider, and registers
// it as the global tracer provider. It returns a shutdown function that should
// be deferred by the caller.
func Init(ctx context.Context, serviceName, endpoint string) (shutdown func(context.Context) error, err error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("merge resource: %w", err)
	}

	defaultTracerName = serviceName

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown = func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown tracer provider: %w", err)
		}
		return nil
	}

	return shutdown, nil
}

// StartSpan inspects the call stack and uses the caller's function name as the
// span name.
func StartSpan(ctx context.Context) (context.Context, func()) {
	pc, _, _, ok := runtime.Caller(1)
	spanName := "unknown"

	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			spanName = path.Base(fn.Name())
		}
	}

	tracer := otel.Tracer(defaultTracerName)
	ctx, span := tracer.Start(ctx, spanName)

	return ctx, func() {
		span.End()
	}
}
