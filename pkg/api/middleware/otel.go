package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// OtelHTTP returns a middleware that adds OpenTelemetry tracing to HTTP requests.
func OtelHTTP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "grpc-gateway")
	}
}
