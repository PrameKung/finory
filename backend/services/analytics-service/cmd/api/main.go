package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/services/analytics-service/internal/config"
	"finory/backend/services/analytics-service/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid service configuration", "error", err)
		os.Exit(1)
	}
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "analytics-service", "port", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
