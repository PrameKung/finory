package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// New returns the HTTP handler for this service.
func New() http.Handler {
	e := echo.New()
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return e
}
