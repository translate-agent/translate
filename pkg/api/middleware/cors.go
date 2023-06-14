package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type cors struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func (c *cors) addHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", strings.Join(c.AllowedHeaders, ","))
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.AllowedMethods, ","))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.AllowedHeaders, ","))
	w.Header().Set("Access-Control-Expose-Headers", strings.Join(c.ExposedHeaders, ","))
	w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprint(c.AllowCredentials))
	w.Header().Set("Access-Control-Max-Age", fmt.Sprint(c.MaxAge))
}

func newCORS() cors {
	return cors{
		AllowedOrigins:   viper.GetStringSlice("service.api.cors.allowed_origins"),
		AllowedMethods:   viper.GetStringSlice("service.api.cors.allowed_methods"),
		AllowedHeaders:   viper.GetStringSlice("service.api.cors.allowed_headers"),
		ExposedHeaders:   viper.GetStringSlice("service.api.cors.exposed_headers"),
		AllowCredentials: viper.GetBool("service.api.cors.allow_credentials"),
		MaxAge:           viper.GetInt("service.api.cors.max_age"),
	}
}

// CORS returns a middleware that adds CORS headers and handles OPTIONS requests.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cors := newCORS()
			cors.addHeaders(w)

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
