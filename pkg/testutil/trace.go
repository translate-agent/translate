package testutil

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Trace(ctx context.Context, t *testing.T, tracer trace.Tracer) (context.Context, func()) {
	testNameParts := strings.Split(t.Name(), "/")

	// Get the sub-test name if it exists.
	spanName := testNameParts[0]
	if isSubTest := len(testNameParts) > 1; isSubTest {
		spanName = testNameParts[len(testNameParts)-1]
	}

	newCtx, span := tracer.Start(ctx, spanName)

	spanEnd := func() {
		defer span.End()

		if t.Failed() {
			span.SetStatus(codes.Error, "")
		}
	}

	return newCtx, spanEnd
}
