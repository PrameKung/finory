package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/services/ledger-service/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("invalid database configuration", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	var databaseName string
	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	if databaseName != "ledger_db" {
		logger.Error("unexpected database", "expected", "ledger_db", "actual", databaseName)
		os.Exit(1)
	}
	logger.Info("database connected", "database", databaseName)
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(pool.Ping),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "ledger-service", "port", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
