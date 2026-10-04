package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestOpenAPIDocument(t *testing.T) {
	response := httptest.NewRecorder()
	e := echo.New()
	registerAPIDocs(e)
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", contentType)
	}
	var document struct {
		OpenAPI string                     `json:"openapi"`
		Paths   map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatalf("invalid OpenAPI JSON: %v", err)
	}
	if document.OpenAPI != "3.1.0" {
		t.Fatalf("OpenAPI version = %q", document.OpenAPI)
	}
	for _, path := range []string{
		"/api/v1/auth/google", "/api/v1/auth/me", "/health", "/api/v1/transactions",
		"/api/v1/categories", "/api/v1/wallets", "/api/v1/budgets",
		"/api/v1/analytics/summary",
	} {
		if _, ok := document.Paths[path]; !ok {
			t.Errorf("OpenAPI document is missing %s", path)
		}
	}
}

func TestScalarAPIReference(t *testing.T) {
	for _, path := range []string{"/docs", "/docs/"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			e := echo.New()
			registerAPIDocs(e)
			e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if !strings.Contains(response.Body.String(), "@scalar/api-reference@1.70.0") ||
				!strings.Contains(response.Body.String(), "url: '/openapi.json'") {
				t.Fatalf("response does not configure Scalar with the OpenAPI document: %q", response.Body.String())
			}
		})
	}
}
