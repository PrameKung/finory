package transactions

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

type fakeTransactionService struct {
	lastFilter ListFilter
}

func (s *fakeTransactionService) Create(context.Context, string, CreateInput) (Transaction, error) {
	return testTransaction(), nil
}

func (s *fakeTransactionService) List(_ context.Context, _ string, filter ListFilter) ([]Transaction, error) {
	s.lastFilter = filter
	return []Transaction{testTransaction()}, nil
}

func (s *fakeTransactionService) Get(context.Context, string, string) (Transaction, error) {
	return testTransaction(), nil
}

func (s *fakeTransactionService) Update(context.Context, string, string, UpdateInput) (Transaction, error) {
	return testTransaction(), nil
}

func (s *fakeTransactionService) Delete(context.Context, string, string) error {
	return nil
}

func TestHandlerRegistersTransactionCRUD(t *testing.T) {
	service := &fakeTransactionService{}
	e := echo.New()
	NewHandler(service).Register(e)
	for _, tc := range []struct {
		name, method, path, body string
		want                     int
	}{
		{name: "list", method: http.MethodGet, path: "/transactions?month=2026-09&type=expense", want: http.StatusOK},
		{name: "create", method: http.MethodPost, path: "/transactions", body: `{"categoryId":"` + testCategoryID + `","walletId":"` + testWalletID + `","type":"expense","amount":"25.50","transactionDate":"2026-09-26"}`, want: http.StatusCreated},
		{name: "read", method: http.MethodGet, path: "/transactions/" + testTransactionID, want: http.StatusOK},
		{name: "update", method: http.MethodPatch, path: "/transactions/" + testTransactionID, body: `{"amount":"30.00"}`, want: http.StatusOK},
		{name: "delete", method: http.MethodDelete, path: "/transactions/" + testTransactionID, want: http.StatusNoContent},
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
	if service.lastFilter != (ListFilter{Month: "2026-09", Type: "expense"}) {
		t.Fatalf("list filter = %+v", service.lastFilter)
	}
}

func TestHandlerRequiresAuthenticatedUser(t *testing.T) {
	e := echo.New()
	NewHandler(&fakeTransactionService{}).Register(e)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/transactions", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func testTransaction() Transaction {
	return Transaction{
		ID: testTransactionID, CategoryID: testCategoryID, WalletID: testWalletID,
		Type: "expense", Amount: "25.5000", TransactionDate: time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC),
		CreatedAt: time.Date(2026, 9, 26, 1, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 9, 26, 1, 0, 0, 0, time.UTC),
	}
}
