package telemetry

import (
	"context"
	"path"
	"runtime"

	"go.opentelemetry.io/otel"
)

const defaultTracerName = "go-restful-api"

// StartSpan automatically inspects the call stack and uses the caller's function name for the span.
func StartSpan(ctx context.Context) (context.Context, func()) {
	// Skip 1 stack frame to get the caller of StartSpan
	pc, _, _, ok := runtime.Caller(1)
	spanName := "unknown"

	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			// fn.Name() returns e.g. "myproject/internal/modules/book.(*service).GetByID"
			// path.Base trims package path, returning "book.(*service).GetByID"
			spanName = path.Base(fn.Name())
		}
	}

	tracer := otel.Tracer(defaultTracerName)
	ctx, span := tracer.Start(ctx, spanName)

	return ctx, func() {
		span.End()
	}
}
