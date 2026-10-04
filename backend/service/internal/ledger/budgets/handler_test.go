package budgets

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

type fakeBudgetService struct {
	lastFilter ListFilter
}

func (s *fakeBudgetService) Create(context.Context, string, CreateInput) (Budget, error) {
	return testBudget(), nil
}

func (s *fakeBudgetService) List(_ context.Context, _ string, filter ListFilter) ([]Budget, error) {
	s.lastFilter = filter
	return []Budget{testBudget()}, nil
}

func (s *fakeBudgetService) Update(context.Context, string, string, UpdateInput) (Budget, error) {
	return testBudget(), nil
}

func (s *fakeBudgetService) Delete(context.Context, string, string) error {
	return nil
}

func TestHandlerRegistersBudgetCRUD(t *testing.T) {
	service := &fakeBudgetService{}
	e := echo.New()
	NewHandler(service).Register(e)
	for _, tc := range []struct {
		name, method, path, body string
		want                     int
	}{
		{name: "list", method: http.MethodGet, path: "/budgets?month=2026-09", want: http.StatusOK},
		{name: "create", method: http.MethodPost, path: "/budgets", body: `{"categoryId":"` + testCategoryID + `","amount":"2500.00","month":"2026-09"}`, want: http.StatusCreated},
		{name: "update", method: http.MethodPatch, path: "/budgets/" + testBudgetID, body: `{"amount":"3000.00"}`, want: http.StatusOK},
		{name: "delete", method: http.MethodDelete, path: "/budgets/" + testBudgetID, want: http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			request.Header.Set(userIDHeader, testUserID)
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			if response.Code != tc.want {
				t.Fatalf("status = %d, want %d; body=%q", response.Code, tc.want, response.Body.String())
			}
		})
	}
	if service.lastFilter != (ListFilter{Month: "2026-09"}) {
		t.Fatalf("list filter = %+v", service.lastFilter)
	}
}

func TestHandlerRequiresAuthenticatedUser(t *testing.T) {
	e := echo.New()
	NewHandler(&fakeBudgetService{}).Register(e)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/budgets", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func testBudget() Budget {
	return Budget{
		ID: testBudgetID, CategoryID: testCategoryID, Amount: "2500.0000",
		MonthStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:  time.Date(2026, 9, 26, 1, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 9, 26, 1, 0, 0, 0, time.UTC),
	}
}
