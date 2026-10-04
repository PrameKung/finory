package wallets

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"finory/backend/service/internal/routes"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

const userIDHeader = "X-User-ID"

type walletService interface {
	Create(context.Context, string, CreateInput) (Wallet, error)
	List(context.Context, string) ([]Wallet, error)
	Update(context.Context, string, string, UpdateInput) (Wallet, error)
	Delete(context.Context, string, string) error
}

type Handler struct {
	service walletService
}

func NewHandler(service walletService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(e routes.Router) {
	e.GET("/wallets", h.list)
	e.POST("/wallets", h.create)
	e.PATCH("/wallets/:id", h.update)
	e.DELETE("/wallets/:id", h.delete)
}

func (h *Handler) list(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	wallets, err := h.service.List(c.Request().Context(), userID)
	if err != nil {
		return operationFailed(c)
	}
	response := make([]walletResponse, 0, len(wallets))
	for _, wallet := range wallets {
		response = append(response, newWalletResponse(wallet))
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
	wallet, err := h.service.Create(c.Request().Context(), userID, request)
	if err != nil {
		return walletError(c, err)
	}
	return c.JSON(http.StatusCreated, newWalletResponse(wallet))
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
	wallet, err := h.service.Update(c.Request().Context(), userID, c.Param("id"), request)
	if err != nil {
		return walletError(c, err)
	}
	return c.JSON(http.StatusOK, newWalletResponse(wallet))
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
		return walletError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func authenticatedUserID(c *echo.Context) (string, bool) {
	value := c.Request().Header.Get(userIDHeader)
	return value, validUUID(value)
}

func validUUID(value string) bool {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil || !id.Valid || id.Bytes == [16]byte{} {
		return false
	}
	return true
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

func walletError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidWallet):
		return invalidRequest(c)
	case errors.Is(err, ErrWalletNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "wallet_not_found"})
	case errors.Is(err, ErrWalletConflict):
		return c.JSON(http.StatusConflict, map[string]string{"error": "wallet_already_exists"})
	default:
		return operationFailed(c)
	}
}

func newWalletResponse(wallet Wallet) walletResponse {
	return walletResponse{
		ID: wallet.ID, Name: wallet.Name, Type: wallet.Type, Balance: wallet.Balance,
		CurrencyCode: wallet.CurrencyCode, IsDefault: wallet.IsDefault,
		CreatedAt: wallet.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: wallet.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func unauthorized(c *echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func invalidRequest(c *echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
}

func operationFailed(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "wallet_operation_failed"})
}
