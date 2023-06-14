package tracer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func getExporter(ctx context.Context, s string) (exp tracesdk.SpanExporter, err error) {
	trimmed := strings.TrimSpace(strings.ToLower(s))
	switch trimmed {
	case "stdout":
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlp":
		//nolint:lll
		// OpenTelemetry Protocol Exporter (OTLP) docs: https://opentelemetry.io/docs/reference/specification/protocol/exporter/
		exp, err = otlptracehttp.New(ctx)
	default:
		log.Printf("Not using traces, exporter is either is set to empty or not supported: %s\n", s)
		return
	}

	if err != nil {
		return nil, fmt.Errorf("new %s exporter: %w", s, err)
	}

	log.Println("Using exporter:", s)

	return
}

func TracerProvider(ctx context.Context, exporter string) (*tracesdk.TracerProvider, error) {
	exp, err := getExporter(ctx, exporter)
	if err != nil {
		return nil, fmt.Errorf("create exporter: %w", err)
	}

	//nolint:lll
	// OpenTelemetry SDK environment variables docs: https://opentelemetry.io/docs/reference/specification/sdk-environment-variables/
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}
