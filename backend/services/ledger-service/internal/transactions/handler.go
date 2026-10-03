package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

const userIDHeader = "X-User-ID"

type transactionService interface {
	Create(context.Context, string, CreateInput) (Transaction, error)
	List(context.Context, string, ListFilter) ([]Transaction, error)
	Get(context.Context, string, string) (Transaction, error)
	Update(context.Context, string, string, UpdateInput) (Transaction, error)
	Delete(context.Context, string, string) error
}

type Handler struct {
	service transactionService
}

func NewHandler(service transactionService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/transactions", h.list)
	e.POST("/transactions", h.create)
	e.GET("/transactions/:id", h.get)
	e.PATCH("/transactions/:id", h.update)
	e.DELETE("/transactions/:id", h.delete)
}

func (h *Handler) list(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	items, err := h.service.List(c.Request().Context(), userID, ListFilter{
		Month: c.QueryParam("month"), Type: c.QueryParam("type"),
	})
	if err != nil {
		return transactionError(c, err)
	}
	response := make([]transactionResponse, 0, len(items))
	for _, item := range items {
		response = append(response, newTransactionResponse(item))
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
	transaction, err := h.service.Create(c.Request().Context(), userID, request)
	if err != nil {
		return transactionError(c, err)
	}
	return c.JSON(http.StatusCreated, newTransactionResponse(transaction))
}

func (h *Handler) get(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	if !validUUID(c.Param("id")) {
		return invalidRequest(c)
	}
	transaction, err := h.service.Get(c.Request().Context(), userID, c.Param("id"))
	if err != nil {
		return transactionError(c, err)
	}
	return c.JSON(http.StatusOK, newTransactionResponse(transaction))
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
	transaction, err := h.service.Update(c.Request().Context(), userID, c.Param("id"), request)
	if err != nil {
		return transactionError(c, err)
	}
	return c.JSON(http.StatusOK, newTransactionResponse(transaction))
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
		return transactionError(c, err)
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

func transactionError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidTransaction):
		return invalidRequest(c)
	case errors.Is(err, ErrInvalidTransactionReference):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_transaction_reference"})
	case errors.Is(err, ErrTransactionNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "transaction_not_found"})
	default:
		return operationFailed(c)
	}
}

func newTransactionResponse(transaction Transaction) transactionResponse {
	return transactionResponse{
		ID: transaction.ID, CategoryID: transaction.CategoryID, WalletID: transaction.WalletID,
		Type: transaction.Type, Amount: transaction.Amount, Description: transaction.Description,
		TransactionDate: transaction.TransactionDate.Format(time.DateOnly),
		CreatedAt:       transaction.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:       transaction.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func unauthorized(c *echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func invalidRequest(c *echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
}

func operationFailed(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "transaction_operation_failed"})
}
