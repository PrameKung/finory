package budgets

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"finory/backend/service/internal/routes"
	"github.com/labstack/echo/v5"
)

const userIDHeader = "X-User-ID"

type budgetService interface {
	Create(context.Context, string, CreateInput) (Budget, error)
	List(context.Context, string, ListFilter) ([]Budget, error)
	Update(context.Context, string, string, UpdateInput) (Budget, error)
	Delete(context.Context, string, string) error
}

type Handler struct {
	service budgetService
}

func NewHandler(service budgetService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(e routes.Router) {
	e.GET("/budgets", h.list)
	e.POST("/budgets", h.create)
	e.PATCH("/budgets/:id", h.update)
	e.DELETE("/budgets/:id", h.delete)
}

func (h *Handler) list(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	items, err := h.service.List(c.Request().Context(), userID, ListFilter{Month: c.QueryParam("month")})
	if err != nil {
		return budgetError(c, err)
	}
	response := make([]budgetResponse, 0, len(items))
	for _, item := range items {
		response = append(response, newBudgetResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) create(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	var request CreateInput
	if err := decodeJSON(c, &request); err != nil {
		return invalidRequest(c)
	}
	budget, err := h.service.Create(c.Request().Context(), userID, request)
	if err != nil {
		return budgetError(c, err)
	}
	return c.JSON(http.StatusCreated, newBudgetResponse(budget))
}

func (h *Handler) update(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	if !validUUID(c.Param("id")) {
		return invalidRequest(c)
	}
	var request UpdateInput
	if err := decodeJSON(c, &request); err != nil {
		return invalidRequest(c)
	}
	budget, err := h.service.Update(c.Request().Context(), userID, c.Param("id"), request)
	if err != nil {
		return budgetError(c, err)
	}
	return c.JSON(http.StatusOK, newBudgetResponse(budget))
}

func (h *Handler) delete(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	if !validUUID(c.Param("id")) {
		return invalidRequest(c)
	}
	if err := h.service.Delete(c.Request().Context(), userID, c.Param("id")); err != nil {
		return budgetError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func authenticatedUserID(c *echo.Context) (string, bool) {
	value := c.Request().Header.Get(userIDHeader)
	return value, validUUID(value)
}

func decodeJSON(c *echo.Context, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(c.Request().Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func budgetError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidBudget):
		return invalidRequest(c)
	case errors.Is(err, ErrInvalidBudgetReference):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_budget_reference"})
	case errors.Is(err, ErrBudgetNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "budget_not_found"})
	case errors.Is(err, ErrBudgetConflict):
		return c.JSON(http.StatusConflict, map[string]string{"error": "budget_already_exists"})
	default:
		return operationFailed(c)
	}
}

func newBudgetResponse(budget Budget) budgetResponse {
	return budgetResponse{
		ID: budget.ID, CategoryID: budget.CategoryID, Amount: budget.Amount,
		Month:     budget.MonthStart.Format("2006-01"),
		CreatedAt: budget.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: budget.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func unauthorized(c *echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func invalidRequest(c *echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
}

func operationFailed(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "budget_operation_failed"})
}
