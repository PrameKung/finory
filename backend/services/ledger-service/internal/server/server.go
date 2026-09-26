package server

import (
	"context"
	"net/http"
	"time"

	"finory/backend/services/ledger-service/internal/categories"
	"finory/backend/services/ledger-service/internal/transactions"
	"finory/backend/services/ledger-service/internal/wallets"

	"github.com/labstack/echo/v5"
)

// New returns the HTTP handler for this service.
func New(
	checkDatabase func(context.Context) error,
	categoryHandler *categories.Handler,
	walletHandler *wallets.Handler,
	transactionHandler *transactions.Handler,
) http.Handler {
	e := echo.New()
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/ready", func(c *echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		if err := checkDatabase(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	if categoryHandler != nil {
		categoryHandler.Register(e)
	}
	if walletHandler != nil {
		walletHandler.Register(e)
	}
	if transactionHandler != nil {
		transactionHandler.Register(e)
	}
	return e
}
