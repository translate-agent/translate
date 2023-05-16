package testutil

import (
	"context"
	"strings"
	"testing"

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

		// // If the test failed, store the error in the global map.
		// if err, ok := errMsgs.Load(t.Name()); ok {

		// 	// TODO temporal solution to handle different error types.
		// 	switch e := err.(type) {
		// 	case traceErr:
		// 		span.SetStatus(otelcodes.Error, e.description)
		// 		span.RecordError(e.err)
		// 	case error:
		// 		span.SetStatus(otelcodes.Error, "")
		// 		span.RecordError(err.(error))
		// 	default:
		// 		t.Fatalf("unknown error type: %T", err)
		// 	}

		// 	// If the test is a sub-test, then parent span should also be marked as failed.
		// 	if len(testNameParts) > 1 {
		// 		errMsgs.Store(testNameParts[0], errors.New("one or more sub-tests failed"))
		// 	}

		// }
	}
	return newCtx, spanEnd
}
