package testutil

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Trace starts a new span using the provided Tracer.
// The span name is derived from the provided testing.T.
func Trace(ctx context.Context, t *testing.T, tracer trace.Tracer) context.Context {
	var span trace.Span

	ctx, span = tracer.Start(ctx, t.Name())

	t.Cleanup(func() {
		if t.Failed() {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	})

	return ctx
}
