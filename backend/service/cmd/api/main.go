package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/service/internal/analytics"
	"finory/backend/service/internal/analytics/summary"
	"finory/backend/service/internal/auth"
	"finory/backend/service/internal/config"
	"finory/backend/service/internal/httpapi/server"
	"finory/backend/service/internal/ledger/budgets"
	"finory/backend/service/internal/ledger/categories"
	"finory/backend/service/internal/ledger/transactions"
	"finory/backend/service/internal/ledger/wallets"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid service configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	authPool, err := connectDatabase(ctx, cfg.AuthDatabaseURL)
	if err != nil {
		cancel()
		logger.Error("auth database connection failed", "error", err)
		os.Exit(1)
	}
	ledgerPool, err := connectDatabase(ctx, cfg.LedgerDatabaseURL)
	cancel()
	if err != nil {
		authPool.Close()
		logger.Error("ledger database connection failed", "error", err)
		os.Exit(1)
	}
	defer authPool.Close()
	defer ledgerPool.Close()
	logger.Info("database connections established")

	authService := auth.NewService(auth.NewRepository(authPool), cfg.JWTAccessSecret)
	oauthConfig := oauth2.Config{
		ClientID: cfg.GoogleClientID, ClientSecret: cfg.GoogleSecret,
		RedirectURL: cfg.GoogleRedirectURL,
		Scopes:      []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:  "https://oauth2.googleapis.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
	keySet := oidc.NewRemoteKeySet(context.Background(), "https://www.googleapis.com/oauth2/v3/certs")
	verifier := oidc.NewVerifier("https://accounts.google.com", keySet, &oidc.Config{ClientID: cfg.GoogleClientID})
	authHandler := auth.NewHandler(oauthConfig, verifier, authService, cfg.AppRedirectURL, cfg.AppLoginURL)

	categoryService := categories.NewService(categories.NewRepository(ledgerPool))
	walletService := wallets.NewService(wallets.NewRepository(ledgerPool))
	transactionService := transactions.NewService(transactions.NewRepository(ledgerPool))
	budgetService := budgets.NewService(budgets.NewRepository(ledgerPool))
	summaryService := summary.NewService(analytics.NewLedgerReader(transactionService, categoryService))

	httpServer := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: server.New(
			cfg, checkDatabases(authPool, ledgerPool), authHandler,
			categories.NewHandler(categoryService), wallets.NewHandler(walletService),
			transactions.NewHandler(transactionService), budgets.NewHandler(budgetService),
			summary.NewHandler(summaryService),
		),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "api", "port", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func connectDatabase(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func checkDatabases(authPool, ledgerPool *pgxpool.Pool) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := authPool.Ping(ctx); err != nil {
			return err
		}
		return ledgerPool.Ping(ctx)
	}
}
