package testutil

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Trace starts a new span using the provided Tracer.
// The span name is derived from the provided testing.T object.
func Trace(ctx context.Context, t *testing.T, tracer trace.Tracer) (context.Context, func()) {
	// Split test name into sub-tests names.
	testNameParts := strings.Split(t.Name(), "/")

	spanName := testNameParts[0]
	// If the test name contains sub-tests, use the sub-test name as span name.
	if len(testNameParts) > 1 {
		spanName = testNameParts[len(testNameParts)-1]
	}

	newCtx, span := tracer.Start(ctx, spanName)
	spanEnd := func() {
		defer span.End()

		// If the test failed, set the span status to Error
		if t.Failed() {
			span.SetStatus(codes.Error, "")
		}
	}

	return newCtx, spanEnd
}
