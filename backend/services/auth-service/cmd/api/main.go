package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/services/auth-service/internal/config"
	"finory/backend/services/auth-service/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid service configuration", "error", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
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
	if databaseName != "auth_db" {
		logger.Error("unexpected database", "expected", "auth_db", "actual", databaseName)
		os.Exit(1)
	}
	logger.Info("database connected", "database", databaseName)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(pool.Ping),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "auth-service", "port", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
