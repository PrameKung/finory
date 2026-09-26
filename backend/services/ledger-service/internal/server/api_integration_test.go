package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"finory/backend/services/ledger-service/internal/budgets"
	"finory/backend/services/ledger-service/internal/categories"
	"finory/backend/services/ledger-service/internal/transactions"
	"finory/backend/services/ledger-service/internal/wallets"
	"github.com/labstack/echo/v5"
)

const (
	apiTestUserID        = "4c3b9d7e-91ef-4b2b-9e28-2408153715d9"
	apiTestCategoryID    = "1c3b9d7e-91ef-4b2b-9e28-2408153715d9"
	apiTestWalletID      = "2c3b9d7e-91ef-4b2b-9e28-2408153715d9"
	apiTestTransactionID = "3c3b9d7e-91ef-4b2b-9e28-2408153715d9"
	apiTestBudgetID      = "4c3b9d7e-91ef-4b2b-9e28-2408153715d8"
)

var apiTestTime = time.Date(2026, 9, 27, 8, 30, 0, 0, time.UTC)

func TestLedgerAPIIntegration(t *testing.T) {
	categoryService := &categoryAPIService{t: t}
	walletService := &walletAPIService{t: t}
	transactionService := &transactionAPIService{t: t}
	budgetService := &budgetAPIService{t: t}
	handler := New(
		func(context.Context) error { return nil },
		categories.NewHandler(categoryService),
		wallets.NewHandler(walletService),
		transactions.NewHandler(transactionService),
		budgets.NewHandler(budgetService),
	)

	for _, tc := range []struct {
		name, method, path, body, wantID string
		wantStatus                       int
		listResponse                     bool
	}{
		{name: "list categories", method: http.MethodGet, path: "/categories", wantStatus: http.StatusOK, wantID: apiTestCategoryID, listResponse: true},
		{name: "create category", method: http.MethodPost, path: "/categories", body: `{"name":"Dining","type":"expense"}`, wantStatus: http.StatusCreated, wantID: apiTestCategoryID},
		{name: "update category", method: http.MethodPatch, path: "/categories/" + apiTestCategoryID, body: `{"name":"Food"}`, wantStatus: http.StatusOK, wantID: apiTestCategoryID},
		{name: "delete category", method: http.MethodDelete, path: "/categories/" + apiTestCategoryID, wantStatus: http.StatusNoContent},
		{name: "list wallets", method: http.MethodGet, path: "/wallets", wantStatus: http.StatusOK, wantID: apiTestWalletID, listResponse: true},
		{name: "create wallet", method: http.MethodPost, path: "/wallets", body: `{"name":"Checking","type":"bank","balance":"1000.00","currencyCode":"THB"}`, wantStatus: http.StatusCreated, wantID: apiTestWalletID},
		{name: "update wallet", method: http.MethodPatch, path: "/wallets/" + apiTestWalletID, body: `{"balance":"1250.00"}`, wantStatus: http.StatusOK, wantID: apiTestWalletID},
		{name: "delete wallet", method: http.MethodDelete, path: "/wallets/" + apiTestWalletID, wantStatus: http.StatusNoContent},
		{name: "list transactions", method: http.MethodGet, path: "/transactions?month=2026-09&type=expense", wantStatus: http.StatusOK, wantID: apiTestTransactionID, listResponse: true},
		{name: "create transaction", method: http.MethodPost, path: "/transactions", body: `{"categoryId":"` + apiTestCategoryID + `","walletId":"` + apiTestWalletID + `","type":"expense","amount":"25.50","transactionDate":"2026-09-27"}`, wantStatus: http.StatusCreated, wantID: apiTestTransactionID},
		{name: "get transaction", method: http.MethodGet, path: "/transactions/" + apiTestTransactionID, wantStatus: http.StatusOK, wantID: apiTestTransactionID},
		{name: "update transaction", method: http.MethodPatch, path: "/transactions/" + apiTestTransactionID, body: `{"amount":"30.00"}`, wantStatus: http.StatusOK, wantID: apiTestTransactionID},
		{name: "delete transaction", method: http.MethodDelete, path: "/transactions/" + apiTestTransactionID, wantStatus: http.StatusNoContent},
		{name: "list budgets", method: http.MethodGet, path: "/budgets?month=2026-09", wantStatus: http.StatusOK, wantID: apiTestBudgetID, listResponse: true},
		{name: "create budget", method: http.MethodPost, path: "/budgets", body: `{"categoryId":"` + apiTestCategoryID + `","amount":"2500.00","month":"2026-09"}`, wantStatus: http.StatusCreated, wantID: apiTestBudgetID},
		{name: "update budget", method: http.MethodPatch, path: "/budgets/" + apiTestBudgetID, body: `{"amount":"3000.00"}`, wantStatus: http.StatusOK, wantID: apiTestBudgetID},
		{name: "delete budget", method: http.MethodDelete, path: "/budgets/" + apiTestBudgetID, wantStatus: http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			request.Header.Set("X-User-ID", apiTestUserID)
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body=%q", response.Code, tc.wantStatus, response.Body.String())
			}
			if tc.wantID != "" {
				assertResponseID(t, response.Body.Bytes(), tc.wantID, tc.listResponse)
			}
		})
	}
}

func TestLedgerAPIRejectsRequestsWithoutAuthenticatedUser(t *testing.T) {
	handler := New(
		func(context.Context) error { return nil },
		categories.NewHandler(&categoryAPIService{t: t}),
		wallets.NewHandler(&walletAPIService{t: t}),
		transactions.NewHandler(&transactionAPIService{t: t}),
		budgets.NewHandler(&budgetAPIService{t: t}),
	)

	for _, path := range []string{"/categories", "/wallets", "/transactions", "/budgets"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusUnauthorized || response.Body.String() != "{\"error\":\"unauthorized\"}\n" {
				t.Fatalf("response = %d %q", response.Code, response.Body.String())
			}
		})
	}
}

func assertResponseID(t *testing.T, body []byte, wantID string, list bool) {
	t.Helper()
	if list {
		var response []map[string]any
		if err := json.Unmarshal(body, &response); err != nil || len(response) != 1 || response[0]["id"] != wantID {
			t.Fatalf("response = %s, want one item with id %q; error=%v", body, wantID, err)
		}
		return
	}
	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil || response["id"] != wantID {
		t.Fatalf("response = %s, want id %q; error=%v", body, wantID, err)
	}
}

type categoryAPIService struct{ t *testing.T }

func (s *categoryAPIService) Create(_ context.Context, userID string, input categories.CreateInput) (categories.Category, error) {
	s.user(userID)
	if input.Name != "Dining" || input.Type != "expense" {
		s.t.Errorf("category create input = %+v", input)
	}
	return apiCategory(), nil
}

func (s *categoryAPIService) List(_ context.Context, userID string) ([]categories.Category, error) {
	s.user(userID)
	return []categories.Category{apiCategory()}, nil
}

func (s *categoryAPIService) Update(_ context.Context, userID, id string, input categories.UpdateInput) (categories.Category, error) {
	s.user(userID)
	if id != apiTestCategoryID || input.Name == nil || *input.Name != "Food" {
		s.t.Errorf("category update: id=%q input=%+v", id, input)
	}
	return apiCategory(), nil
}

func (s *categoryAPIService) Delete(_ context.Context, userID, id string) error {
	s.user(userID)
	s.id(id, apiTestCategoryID)
	return nil
}

func (s *categoryAPIService) user(userID string) {
	s.t.Helper()
	if userID != apiTestUserID {
		s.t.Errorf("user ID = %q, want %q", userID, apiTestUserID)
	}
}

func (s *categoryAPIService) id(got, want string) {
	s.t.Helper()
	if got != want {
		s.t.Errorf("resource ID = %q, want %q", got, want)
	}
}

type walletAPIService struct{ t *testing.T }

func (s *walletAPIService) Create(_ context.Context, userID string, input wallets.CreateInput) (wallets.Wallet, error) {
	s.user(userID)
	if input.Name != "Checking" || input.Type != "bank" || input.Balance != "1000.00" || input.CurrencyCode != "THB" {
		s.t.Errorf("wallet create input = %+v", input)
	}
	return apiWallet(), nil
}

func (s *walletAPIService) List(_ context.Context, userID string) ([]wallets.Wallet, error) {
	s.user(userID)
	return []wallets.Wallet{apiWallet()}, nil
}

func (s *walletAPIService) Update(_ context.Context, userID, id string, input wallets.UpdateInput) (wallets.Wallet, error) {
	s.user(userID)
	s.id(id, apiTestWalletID)
	if input.Balance == nil || *input.Balance != "1250.00" {
		s.t.Errorf("wallet update input = %+v", input)
	}
	return apiWallet(), nil
}

func (s *walletAPIService) Delete(_ context.Context, userID, id string) error {
	s.user(userID)
	s.id(id, apiTestWalletID)
	return nil
}

func (s *walletAPIService) user(userID string) {
	s.t.Helper()
	if userID != apiTestUserID {
		s.t.Errorf("user ID = %q, want %q", userID, apiTestUserID)
	}
}

func (s *walletAPIService) id(got, want string) {
	s.t.Helper()
	if got != want {
		s.t.Errorf("resource ID = %q, want %q", got, want)
	}
}

type transactionAPIService struct{ t *testing.T }

func (s *transactionAPIService) Create(_ context.Context, userID string, input transactions.CreateInput) (transactions.Transaction, error) {
	s.user(userID)
	if input.CategoryID != apiTestCategoryID || input.WalletID != apiTestWalletID || input.Amount != "25.50" {
		s.t.Errorf("transaction create input = %+v", input)
	}
	return apiTransaction(), nil
}

func (s *transactionAPIService) List(_ context.Context, userID string, filter transactions.ListFilter) ([]transactions.Transaction, error) {
	s.user(userID)
	if filter.Month != "2026-09" || filter.Type != "expense" {
		s.t.Errorf("transaction filter = %+v", filter)
	}
	return []transactions.Transaction{apiTransaction()}, nil
}

func (s *transactionAPIService) Get(_ context.Context, userID, id string) (transactions.Transaction, error) {
	s.user(userID)
	s.id(id, apiTestTransactionID)
	return apiTransaction(), nil
}

func (s *transactionAPIService) Update(_ context.Context, userID, id string, input transactions.UpdateInput) (transactions.Transaction, error) {
	s.user(userID)
	s.id(id, apiTestTransactionID)
	if input.Amount == nil || *input.Amount != "30.00" {
		s.t.Errorf("transaction update input = %+v", input)
	}
	return apiTransaction(), nil
}

func (s *transactionAPIService) Delete(_ context.Context, userID, id string) error {
	s.user(userID)
	s.id(id, apiTestTransactionID)
	return nil
}

func (s *transactionAPIService) user(userID string) {
	s.t.Helper()
	if userID != apiTestUserID {
		s.t.Errorf("user ID = %q, want %q", userID, apiTestUserID)
	}
}

func (s *transactionAPIService) id(got, want string) {
	s.t.Helper()
	if got != want {
		s.t.Errorf("resource ID = %q, want %q", got, want)
	}
}

type budgetAPIService struct{ t *testing.T }

func (s *budgetAPIService) Create(_ context.Context, userID string, input budgets.CreateInput) (budgets.Budget, error) {
	s.user(userID)
	if input.CategoryID != apiTestCategoryID || input.Amount != "2500.00" || input.Month != "2026-09" {
		s.t.Errorf("budget create input = %+v", input)
	}
	return apiBudget(), nil
}

func (s *budgetAPIService) List(_ context.Context, userID string, filter budgets.ListFilter) ([]budgets.Budget, error) {
	s.user(userID)
	if filter.Month != "2026-09" {
		s.t.Errorf("budget filter = %+v", filter)
	}
	return []budgets.Budget{apiBudget()}, nil
}

func (s *budgetAPIService) Update(_ context.Context, userID, id string, input budgets.UpdateInput) (budgets.Budget, error) {
	s.user(userID)
	s.id(id, apiTestBudgetID)
	if input.Amount == nil || *input.Amount != "3000.00" {
		s.t.Errorf("budget update input = %+v", input)
	}
	return apiBudget(), nil
}

func (s *budgetAPIService) Delete(_ context.Context, userID, id string) error {
	s.user(userID)
	s.id(id, apiTestBudgetID)
	return nil
}

func (s *budgetAPIService) user(userID string) {
	s.t.Helper()
	if userID != apiTestUserID {
		s.t.Errorf("user ID = %q, want %q", userID, apiTestUserID)
	}
}

func (s *budgetAPIService) id(got, want string) {
	s.t.Helper()
	if got != want {
		s.t.Errorf("resource ID = %q, want %q", got, want)
	}
}

func apiCategory() categories.Category {
	return categories.Category{ID: apiTestCategoryID, Name: "Dining", Type: "expense", CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
}

func apiWallet() wallets.Wallet {
	return wallets.Wallet{ID: apiTestWalletID, Name: "Checking", Type: "bank", Balance: "1000.0000", CurrencyCode: "THB", CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
}

func apiTransaction() transactions.Transaction {
	return transactions.Transaction{ID: apiTestTransactionID, CategoryID: apiTestCategoryID, WalletID: apiTestWalletID, Type: "expense", Amount: "25.5000", TransactionDate: apiTestTime, CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
}

func apiBudget() budgets.Budget {
	return budgets.Budget{ID: apiTestBudgetID, CategoryID: apiTestCategoryID, Amount: "2500.0000", MonthStart: apiTestTime, CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
}
