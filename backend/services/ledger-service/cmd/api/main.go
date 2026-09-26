package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/services/ledger-service/internal/budgets"
	"finory/backend/services/ledger-service/internal/categories"
	"finory/backend/services/ledger-service/internal/config"
	"finory/backend/services/ledger-service/internal/server"
	"finory/backend/services/ledger-service/internal/transactions"
	"finory/backend/services/ledger-service/internal/wallets"
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
	if databaseName != "ledger_db" {
		logger.Error("unexpected database", "expected", "ledger_db", "actual", databaseName)
		os.Exit(1)
	}
	logger.Info("database connected", "database", databaseName)
	categoryRepository := categories.NewRepository(pool)
	categoryService := categories.NewService(categoryRepository)
	categoryHandler := categories.NewHandler(categoryService)
	walletRepository := wallets.NewRepository(pool)
	walletService := wallets.NewService(walletRepository)
	walletHandler := wallets.NewHandler(walletService)
	transactionRepository := transactions.NewRepository(pool)
	transactionService := transactions.NewService(transactionRepository)
	transactionHandler := transactions.NewHandler(transactionService)
	budgetRepository := budgets.NewRepository(pool)
	budgetService := budgets.NewService(budgetRepository)
	budgetHandler := budgets.NewHandler(budgetService)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(pool.Ping, categoryHandler, walletHandler, transactionHandler, budgetHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "ledger-service", "port", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
