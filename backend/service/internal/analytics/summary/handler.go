package summary

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"finory/backend/service/internal/routes"
	"github.com/labstack/echo/v5"
)

const userIDHeader = "X-User-ID"

type summaryService interface {
	Monthly(context.Context, string, string, time.Time) (MonthlySummary, error)
	Categories(context.Context, string, string, time.Time) (CategoryDistribution, error)
	Trends(context.Context, string, string, time.Time) (TrendSeries, error)
	MonthlyComparison(context.Context, string, string, time.Time) (MonthlyComparison, error)
}

type Handler struct {
	service summaryService
}

func NewHandler(service summaryService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(e routes.Router) {
	e.GET("/analytics/summary", h.monthly)
	e.GET("/analytics/categories", h.categories)
	e.GET("/analytics/trends", h.trends)
	e.GET("/analytics/monthly", h.monthlyComparison)
}

func (h *Handler) monthly(c *echo.Context) error {
	userID, month, errorCode := requestScope(c)
	if errorCode != "" {
		return requestError(c, errorCode)
	}

	result, err := h.service.Monthly(
		c.Request().Context(), userID, c.Request().Header.Get(echo.HeaderXRequestID), month,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "analytics_failed"})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) categories(c *echo.Context) error {
	userID, month, errorCode := requestScope(c)
	if errorCode != "" {
		return requestError(c, errorCode)
	}

	result, err := h.service.Categories(
		c.Request().Context(), userID, c.Request().Header.Get(echo.HeaderXRequestID), month,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "analytics_failed"})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) trends(c *echo.Context) error {
	userID, month, errorCode := requestScope(c)
	if errorCode != "" {
		return requestError(c, errorCode)
	}

	result, err := h.service.Trends(
		c.Request().Context(), userID, c.Request().Header.Get(echo.HeaderXRequestID), month,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "analytics_failed"})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) monthlyComparison(c *echo.Context) error {
	userID, month, errorCode := requestScope(c)
	if errorCode != "" {
		return requestError(c, errorCode)
	}

	result, err := h.service.MonthlyComparison(
		c.Request().Context(), userID, c.Request().Header.Get(echo.HeaderXRequestID), month,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "analytics_failed"})
	}
	return c.JSON(http.StatusOK, result)
}

func requestScope(c *echo.Context) (string, time.Time, string) {
	userID := c.Request().Header.Get(userIDHeader)
	if !validUUID(userID) {
		return "", time.Time{}, "unauthorized"
	}

	month := time.Now().UTC()
	if value := strings.TrimSpace(c.QueryParam("month")); value != "" {
		parsed, err := time.Parse("2006-01", value)
		if err != nil {
			return "", time.Time{}, "invalid_request"
		}
		month = parsed
	}
	return userID, month, ""
}

func requestError(c *echo.Context, code string) error {
	status := http.StatusBadRequest
	if code == "unauthorized" {
		status = http.StatusUnauthorized
	}
	return c.JSON(status, map[string]string{"error": code})
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	compact := strings.ReplaceAll(value, "-", "")
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != 16 {
		return false
	}
	for _, part := range decoded {
		if part != 0 {
			return true
		}
	}
	return false
}
