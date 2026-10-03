package server

import (
	"net/http"

	"finory/backend/services/analytics-service/internal/summary"

	"github.com/labstack/echo/v5"
)

// New returns the HTTP handler for this service.
func New(summaryHandler *summary.Handler) http.Handler {
	e := echo.New()
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	if summaryHandler != nil {
		summaryHandler.Register(e)
	}
	return e
}
