package server

import (
	"net/http"
	"net/http/httputil"

	"finory/backend/services/api-gateway/internal/config"

	"github.com/labstack/echo/v5"
)

// New returns the HTTP handler for this service.
func New(cfg config.Config) http.Handler {
	e := echo.New()
	authService := httputil.NewSingleHostReverseProxy(cfg.AuthServiceURL)
	e.GET("/api/v1/auth/health", func(c *echo.Context) error {
		http.StripPrefix("/api/v1/auth", authService).ServeHTTP(c.Response(), c.Request())
		return nil
	})
	authProxy := http.StripPrefix("/api/v1", authService)
	proxyAuth := func(c *echo.Context) error {
		authProxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	e.Any("/api/v1/auth", proxyAuth)
	e.Any("/api/v1/auth/*", proxyAuth)
	ledgerProxy := http.StripPrefix("/api/v1", httputil.NewSingleHostReverseProxy(cfg.LedgerServiceURL))
	proxyLedger := func(c *echo.Context) error {
		ledgerProxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	for _, resource := range []string{"transactions", "categories", "wallets", "budgets"} {
		path := "/api/v1/" + resource
		e.Any(path, proxyLedger)
		e.Any(path+"/*", proxyLedger)
	}
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})
	return e
}
