package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/services/ledger-service/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "ledger-service", "port", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
