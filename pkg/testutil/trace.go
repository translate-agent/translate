package testutil

import (
	"context"
	"log"
	"testing"

	"go.expect.digital/translate/pkg/tracer"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	tp         *sdktrace.TracerProvider
	testTracer trace.Tracer
)

func init() {
	ctx := context.Background()

	var err error

	tp, err = tracer.Provider(ctx)
	if err != nil {
		log.Panicf("set tracer provider: %v", err)
	}

	testTracer = tp.Tracer("go.expect.digital/translate-agent/translate")
}

// Tracer returns test tracer.
func Tracer() trace.Tracer { //nolint:ireturn
	return testTracer
}

// Trace starts a new root span using the provided Tracer.
// The span name is derived from the provided testing.T.
//
//nolint:spancheck
func Trace(t *testing.T) (context.Context, SubtestFn) {
	t.Helper()

	ctx, span := testTracer.Start(t.Context(), t.Name())

	t.Cleanup(func() {
		if t.Failed() {
			span.SetStatus(codes.Error, "")
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()

		tp.ForceFlush(ctx)
	})

	return ctx, Subtest(t)
}

// SubtestFn is a function that runs parallel subtest with a trace instrumentation.
type SubtestFn = func(name string, f func(context.Context, *testing.T))

// Subtest returns SubtestFn that runs parallel subtest with a trace instrumentation.
func Subtest(t *testing.T) SubtestFn {
	t.Helper()

	return func(name string, f func(ctx context.Context, t *testing.T)) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, _ := Trace(t)

			f(ctx, t)
		})
	}
}
