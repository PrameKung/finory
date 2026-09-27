package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"finory/backend/services/auth-service/internal/auth"
	"finory/backend/services/auth-service/internal/config"
	"finory/backend/services/auth-service/internal/server"
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
	userRepository := auth.NewRepository(pool)
	authService := auth.NewService(userRepository, cfg.JWTAccessSecret)
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
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.New(pool.Ping, authHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service listening", "service", "auth-service", "port", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("service stopped", "error", err)
		os.Exit(1)
	}
}
