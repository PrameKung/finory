package server

import (
	"context"
	"net/http"
	"slices"
	"time"

	"finory/backend/service/internal/analytics/summary"
	"finory/backend/service/internal/auth"
	"finory/backend/service/internal/config"
	"finory/backend/service/internal/httpapi/middleware"
	"finory/backend/service/internal/ledger/budgets"
	"finory/backend/service/internal/ledger/categories"
	"finory/backend/service/internal/ledger/transactions"
	"finory/backend/service/internal/ledger/wallets"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

// New mounts all application features in one HTTP process under /api/v1.
func New(
	cfg config.Config,
	checkDatabase func(context.Context) error,
	authHandler *auth.Handler,
	categoryHandler *categories.Handler,
	walletHandler *wallets.Handler,
	transactionHandler *transactions.Handler,
	budgetHandler *budgets.Handler,
	summaryHandler *summary.Handler,
) http.Handler {
	e := echo.New()
	registerAPIDocs(e)
	e.Use(echomiddleware.RequestIDWithConfig(echomiddleware.RequestIDConfig{
		RequestIDHandler: func(c *echo.Context, requestID string) {
			c.Request().Header.Set(echo.HeaderXRequestID, requestID)
		},
	}))
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, "X-Request-ID"},
		ExposeHeaders:    []string{echo.HeaderXRequestID},
	}))
	registerHealthRoutes(e, checkDatabase)

	requireAccessToken := middleware.RequireAccessToken(cfg.JWTAccessSecret, cfg.CORSAllowedOrigins)
	v1 := e.Group("/api/v1")
	if authHandler != nil {
		authHandler.Register(v1, auth.RouteMiddleware{
			RequireAccessToken:   requireAccessToken,
			RequireRefreshOrigin: requireAllowedOrigin(cfg.CORSAllowedOrigins),
			RequireLogoutAuth:    conditionalAccessToken(requireAccessToken, cfg.CORSAllowedOrigins),
		})
	}
	protected := v1.Group("", requireAccessToken)
	if categoryHandler != nil {
		categoryHandler.Register(protected)
	}
	if walletHandler != nil {
		walletHandler.Register(protected)
	}
	if transactionHandler != nil {
		transactionHandler.Register(protected)
	}
	if budgetHandler != nil {
		budgetHandler.Register(protected)
	}
	if summaryHandler != nil {
		summaryHandler.Register(protected)
	}
	return e
}

func registerHealthRoutes(e *echo.Echo, checkDatabase func(context.Context) error) {
	health := func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
	e.GET("/health", health)
	e.GET("/ready", func(c *echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		if checkDatabase == nil || checkDatabase(ctx) != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		}
		return health(c)
	})
}

func requireAllowedOrigin(origins []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !slices.Contains(origins, c.Request().Header.Get("Origin")) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden_origin"})
			}
			return next(c)
		}
	}
}

func conditionalAccessToken(requireToken echo.MiddlewareFunc, origins []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		withToken := requireToken(next)
		return func(c *echo.Context) error {
			if slices.Contains(origins, c.Request().Header.Get("Origin")) {
				return next(c)
			}
			return withToken(c)
		}
	}
}
