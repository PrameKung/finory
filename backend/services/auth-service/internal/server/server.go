package server

import (
	"context"
	"net/http"
	"time"

	"finory/backend/services/auth-service/internal/auth"

	"github.com/labstack/echo/v5"
)

// New returns the HTTP handler for this service.
func New(checkDatabase func(context.Context) error, authHandler *auth.Handler) http.Handler {
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
	if authHandler != nil {
		authHandler.Register(e)
	}
	return e
}
