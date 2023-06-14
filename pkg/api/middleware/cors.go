package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type cors struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int
}

func getCORS() *cors {
	return &cors{
		allowedOrigins:   viper.GetStringSlice("service.api.cors.allowed_origins"),
		allowedMethods:   viper.GetStringSlice("service.api.cors.allowed_methods"),
		allowedHeaders:   viper.GetStringSlice("service.api.cors.allowed_headers"),
		exposedHeaders:   viper.GetStringSlice("service.api.cors.exposed_headers"),
		allowCredentials: viper.GetBool("service.api.cors.allow_credentials"),
		maxAge:           viper.GetInt("service.api.cors.max_age"),
	}
}

// CORS applies Cross-Origin Resource Sharing specification to the request.
//
// TODO: Add preflight (OPTIONS) request handling, now it always returns OK.
// The only header that is checked is the Origin header by the web browser itself.
// All other headers should be checked in preflight request handler.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cors := getCORS()

			w.Header().Set("Access-Control-Allow-Origin", strings.Join(cors.allowedOrigins, ","))
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cors.allowedMethods, ","))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cors.allowedHeaders, ","))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(cors.exposedHeaders, ","))
			w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprint(cors.allowCredentials))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprint(cors.maxAge))

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
