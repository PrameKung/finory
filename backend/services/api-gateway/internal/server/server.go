package server

import (
	"net/http"
	"net/http/httputil"

	"finory/backend/services/api-gateway/internal/config"

	"github.com/go-chi/chi/v5"
)

// New returns the HTTP handler for this service.
func New(cfg config.Config) http.Handler {
	r := chi.NewRouter()
	authService := httputil.NewSingleHostReverseProxy(cfg.AuthServiceURL)
	r.Get("/api/v1/auth/health", http.StripPrefix("/api/v1/auth", authService).ServeHTTP)
	authProxy := http.StripPrefix("/api/v1", authService)
	r.Handle("/api/v1/auth", authProxy)
	r.Handle("/api/v1/auth/*", authProxy)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
	})
	return r
}
