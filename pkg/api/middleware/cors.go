package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
	"github.com/spf13/viper"
)

func CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   viper.GetStringSlice("service.api.cors.allowed_origins"),
		AllowedMethods:   viper.GetStringSlice("service.api.cors.allowed_methods"),
		AllowedHeaders:   viper.GetStringSlice("service.api.cors.allowed_headers"),
		ExposedHeaders:   viper.GetStringSlice("service.api.cors.exposed_headers"),
		AllowCredentials: viper.GetBool("service.api.cors.allow_credentials"),
		MaxAge:           viper.GetInt("service.api.cors.max_age"),
	})
}
