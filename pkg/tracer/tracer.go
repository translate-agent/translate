package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func TracerProvider() (func(context.Context) error, error) {
	//nolint:lll
	// OpenTelemetry SDK environment variables docs: https://opentelemetry.io/docs/reference/specification/sdk-environment-variables/
	// OpenTelemetry Protocol Exporter (OTLP) docs: https://opentelemetry.io/docs/reference/specification/protocol/exporter/
	exp, err := otlptracehttp.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
		)),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}
