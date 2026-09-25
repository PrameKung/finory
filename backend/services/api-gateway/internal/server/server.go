package server

import (
	"net/http"
	"net/http/httputil"
	"slices"

	"finory/backend/services/api-gateway/internal/config"
	gatewaymiddleware "finory/backend/services/api-gateway/internal/middleware"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// New returns the HTTP handler for this service.
func New(cfg config.Config) http.Handler {
	e := echo.New()
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c *echo.Context, requestID string) {
			c.Request().Header.Set(echo.HeaderXRequestID, requestID)
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, "X-Request-ID"},
		ExposeHeaders:    []string{echo.HeaderXRequestID},
	}))
	requireAccessToken := gatewaymiddleware.RequireAccessToken(cfg.JWTAccessSecret, cfg.CORSAllowedOrigins)
	authService := httputil.NewSingleHostReverseProxy(cfg.AuthServiceURL)
	authService.ModifyResponse = removeUpstreamRequestID
	e.GET("/api/v1/auth/health", func(c *echo.Context) error {
		http.StripPrefix("/api/v1/auth", authService).ServeHTTP(c.Response(), c.Request())
		return nil
	})
	authProxy := http.StripPrefix("/api/v1", authService)
	proxyAuth := func(c *echo.Context) error {
		authProxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	for _, path := range []string{"/api/v1/auth/google", "/api/v1/auth/google/callback"} {
		e.GET(path, proxyAuth)
	}
	e.POST("/api/v1/auth/logout", proxyAuth, func(next echo.HandlerFunc) echo.HandlerFunc {
		requireToken := requireAccessToken(next)
		return func(c *echo.Context) error {
			if slices.Contains(cfg.CORSAllowedOrigins, c.Request().Header.Get("Origin")) {
				return next(c)
			}
			return requireToken(c)
		}
	})
	e.Any("/api/v1/auth", proxyAuth, requireAccessToken)
	e.Any("/api/v1/auth/*", proxyAuth, requireAccessToken)
	ledgerService := httputil.NewSingleHostReverseProxy(cfg.LedgerServiceURL)
	ledgerService.ModifyResponse = removeUpstreamRequestID
	ledgerProxy := http.StripPrefix("/api/v1", ledgerService)
	proxyLedger := func(c *echo.Context) error {
		ledgerProxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	for _, resource := range []string{"transactions", "categories", "wallets", "budgets"} {
		path := "/api/v1/" + resource
		e.Any(path, proxyLedger, requireAccessToken)
		e.Any(path+"/*", proxyLedger, requireAccessToken)
	}
	analyticsService := httputil.NewSingleHostReverseProxy(cfg.AnalyticsServiceURL)
	analyticsService.ModifyResponse = removeUpstreamRequestID
	e.GET("/api/v1/analytics/health", func(c *echo.Context) error {
		http.StripPrefix("/api/v1/analytics", analyticsService).ServeHTTP(c.Response(), c.Request())
		return nil
	})
	analyticsProxy := http.StripPrefix("/api/v1", analyticsService)
	proxyAnalytics := func(c *echo.Context) error {
		analyticsProxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	e.Any("/api/v1/analytics", proxyAnalytics, requireAccessToken)
	e.Any("/api/v1/analytics/*", proxyAnalytics, requireAccessToken)
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})
	return e
}

func removeUpstreamRequestID(response *http.Response) error {
	response.Header.Del(echo.HeaderXRequestID)
	return nil
}
