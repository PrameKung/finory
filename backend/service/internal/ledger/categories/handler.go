package categories

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

type categoryService interface {
	Create(context.Context, string, CreateInput) (Category, error)
	List(context.Context, string) ([]Category, error)
	Update(context.Context, string, string, UpdateInput) (Category, error)
	Delete(context.Context, string, string) error
}

type Handler struct {
	service categoryService
}

func NewHandler(service categoryService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(e routes.Router) {
	e.GET("/categories", h.list)
	e.POST("/categories", h.create)
	e.PATCH("/categories/:id", h.update)
	e.DELETE("/categories/:id", h.delete)
}

func (h *Handler) list(c *echo.Context) error {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return unauthorized(c)
	}
	categories, err := h.service.List(c.Request().Context(), userID)
	if err != nil {
		return operationFailed(c)
	}
	response := make([]categoryResponse, 0, len(categories))
	for _, category := range categories {
		response = append(response, newCategoryResponse(category))
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
	category, err := h.service.Create(c.Request().Context(), userID, request)
	if err != nil {
		return categoryError(c, err)
	}
	return c.JSON(http.StatusCreated, newCategoryResponse(category))
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
	category, err := h.service.Update(c.Request().Context(), userID, c.Param("id"), request)
	if err != nil {
		return categoryError(c, err)
	}
	return c.JSON(http.StatusOK, newCategoryResponse(category))
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
		return categoryError(c, err)
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

func categoryError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrInvalidCategory):
		return invalidRequest(c)
	case errors.Is(err, ErrCategoryNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "category_not_found"})
	case errors.Is(err, ErrCategoryConflict):
		return c.JSON(http.StatusConflict, map[string]string{"error": "category_already_exists"})
	default:
		return operationFailed(c)
	}
}

func newCategoryResponse(category Category) categoryResponse {
	return categoryResponse{
		ID: category.ID, Name: category.Name, Type: category.Type,
		Icon: category.Icon, Color: category.Color, IsDefault: category.IsDefault,
		CreatedAt: category.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: category.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func unauthorized(c *echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func invalidRequest(c *echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
}

func operationFailed(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "category_operation_failed"})
}
