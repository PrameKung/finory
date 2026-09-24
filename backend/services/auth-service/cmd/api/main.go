package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/services/auth-service/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "auth-service", "port", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
