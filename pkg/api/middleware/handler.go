package middleware

import (
	"net/http"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Handler struct {
	RestHandler http.Handler // Handler for REST requests
	GrpcHandler http.Handler // Handler for gRPC requests
}

// HandlerType is the type of handler to apply the middleware to.
type HandlerType int

const (
	REST HandlerType = iota
	GRPC
	BOTH
)

// Add adds the given middleware to the specific handler.
func (h *Handler) Add(to HandlerType, mw func(http.Handler) http.Handler) {
	if to == REST || to == BOTH {
		h.RestHandler = mw(h.RestHandler)
	}

	if to == GRPC || to == BOTH {
		h.GrpcHandler = mw(h.GrpcHandler)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// routes gRPC and REST requests to the appropriate handler.
	h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			h.GrpcHandler.ServeHTTP(w, r)
		} else {
			h.RestHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{}).ServeHTTP(w, r)
}
