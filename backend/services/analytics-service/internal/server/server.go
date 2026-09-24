package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// New returns the HTTP handler for this service.
func New() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
	})
	return r
}
