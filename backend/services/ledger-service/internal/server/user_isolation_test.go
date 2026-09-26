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
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
)

const (
	isolationUserA        = "10000000-0000-4000-8000-000000000001"
	isolationUserB        = "10000000-0000-4000-8000-000000000002"
	isolationCategoryA    = "20000000-0000-4000-8000-000000000001"
	isolationCategoryB    = "20000000-0000-4000-8000-000000000002"
	isolationWalletA      = "30000000-0000-4000-8000-000000000001"
	isolationWalletB      = "30000000-0000-4000-8000-000000000002"
	isolationTransactionA = "40000000-0000-4000-8000-000000000001"
	isolationTransactionB = "40000000-0000-4000-8000-000000000002"
	isolationBudgetA      = "50000000-0000-4000-8000-000000000001"
	isolationBudgetB      = "50000000-0000-4000-8000-000000000002"
)

type ownedCategory struct {
	owner string
	value categories.Category
}

type isolationCategoryRepository struct {
	items map[string]ownedCategory
}

func (r *isolationCategoryRepository) CreateDefaults(context.Context, string) (int64, error) {
	return 0, nil
}

func (r *isolationCategoryRepository) List(_ context.Context, userID string) ([]categories.Category, error) {
	items := make([]categories.Category, 0)
	for _, item := range r.items {
		if item.owner == userID {
			items = append(items, item.value)
		}
	}
	return items, nil
}

func (r *isolationCategoryRepository) Create(_ context.Context, userID string, params categories.CreateParams) (categories.Category, error) {
	item := categories.Category{ID: isolationCategoryA, Name: params.Name, Type: params.Type, CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
	r.items[item.ID] = ownedCategory{owner: userID, value: item}
	return item, nil
}

func (r *isolationCategoryRepository) Update(_ context.Context, userID, id string, _ categories.UpdateParams) (categories.Category, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return categories.Category{}, pgx.ErrNoRows
	}
	return item.value, nil
}

func (r *isolationCategoryRepository) Delete(_ context.Context, userID, id string) (bool, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

type ownedWallet struct {
	owner string
	value wallets.Wallet
}

type isolationWalletRepository struct {
	items map[string]ownedWallet
}

func (r *isolationWalletRepository) CreateDefaultCash(context.Context, string) (int64, error) {
	return 0, nil
}

func (r *isolationWalletRepository) List(_ context.Context, userID string) ([]wallets.Wallet, error) {
	items := make([]wallets.Wallet, 0)
	for _, item := range r.items {
		if item.owner == userID {
			items = append(items, item.value)
		}
	}
	return items, nil
}

func (r *isolationWalletRepository) Create(_ context.Context, userID string, params wallets.CreateParams) (wallets.Wallet, error) {
	item := wallets.Wallet{ID: isolationWalletA, Name: params.Name, Type: params.Type, Balance: params.Balance, CurrencyCode: params.CurrencyCode, CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
	r.items[item.ID] = ownedWallet{owner: userID, value: item}
	return item, nil
}

func (r *isolationWalletRepository) Update(_ context.Context, userID, id string, _ wallets.UpdateParams) (wallets.Wallet, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return wallets.Wallet{}, pgx.ErrNoRows
	}
	return item.value, nil
}

func (r *isolationWalletRepository) Delete(_ context.Context, userID, id string) (bool, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

type ownedTransaction struct {
	owner string
	value transactions.Transaction
}

type isolationTransactionRepository struct {
	items          map[string]ownedTransaction
	categoryOwners map[string]string
	walletOwners   map[string]string
}

func (r *isolationTransactionRepository) Create(_ context.Context, userID string, params transactions.CreateParams) (transactions.Transaction, error) {
	if r.categoryOwners[params.CategoryID] != userID || r.walletOwners[params.WalletID] != userID {
		return transactions.Transaction{}, pgx.ErrNoRows
	}
	item := transactions.Transaction{ID: isolationTransactionA, CategoryID: params.CategoryID, WalletID: params.WalletID, Type: params.Type, Amount: params.Amount, TransactionDate: params.TransactionDate, CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
	r.items[item.ID] = ownedTransaction{owner: userID, value: item}
	return item, nil
}

func (r *isolationTransactionRepository) List(_ context.Context, userID string, _ transactions.ListParams) ([]transactions.Transaction, error) {
	items := make([]transactions.Transaction, 0)
	for _, item := range r.items {
		if item.owner == userID {
			items = append(items, item.value)
		}
	}
	return items, nil
}

func (r *isolationTransactionRepository) Get(_ context.Context, userID, id string) (transactions.Transaction, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return transactions.Transaction{}, pgx.ErrNoRows
	}
	return item.value, nil
}

func (r *isolationTransactionRepository) Update(_ context.Context, userID, id string, _ transactions.UpdateParams) (transactions.Transaction, error) {
	return r.Get(context.Background(), userID, id)
}

func (r *isolationTransactionRepository) Delete(_ context.Context, userID, id string) (bool, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

type ownedBudget struct {
	owner string
	value budgets.Budget
}

type isolationBudgetRepository struct {
	items          map[string]ownedBudget
	categoryOwners map[string]string
}

func (r *isolationBudgetRepository) Create(_ context.Context, userID string, params budgets.CreateParams) (budgets.Budget, error) {
	if r.categoryOwners[params.CategoryID] != userID {
		return budgets.Budget{}, pgx.ErrNoRows
	}
	item := budgets.Budget{ID: isolationBudgetA, CategoryID: params.CategoryID, Amount: params.Amount, MonthStart: params.MonthStart, CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
	r.items[item.ID] = ownedBudget{owner: userID, value: item}
	return item, nil
}

func (r *isolationBudgetRepository) List(_ context.Context, userID string, _ budgets.ListParams) ([]budgets.Budget, error) {
	items := make([]budgets.Budget, 0)
	for _, item := range r.items {
		if item.owner == userID {
			items = append(items, item.value)
		}
	}
	return items, nil
}

func (r *isolationBudgetRepository) Get(_ context.Context, userID, id string) (budgets.Budget, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return budgets.Budget{}, pgx.ErrNoRows
	}
	return item.value, nil
}

func (r *isolationBudgetRepository) Update(_ context.Context, userID, id string, _ budgets.UpdateParams) (budgets.Budget, error) {
	return r.Get(context.Background(), userID, id)
}

func (r *isolationBudgetRepository) Delete(_ context.Context, userID, id string) (bool, error) {
	item, ok := r.items[id]
	if !ok || item.owner != userID {
		return false, nil
	}
	delete(r.items, id)
	return true, nil
}

func TestLedgerAPIIsolatesUserData(t *testing.T) {
	handler := newIsolationHandler()

	for _, tc := range []struct {
		path, wantID string
	}{
		{path: "/categories", wantID: isolationCategoryA},
		{path: "/wallets", wantID: isolationWalletA},
		{path: "/transactions", wantID: isolationTransactionA},
		{path: "/budgets", wantID: isolationBudgetA},
	} {
		t.Run("list "+tc.path, func(t *testing.T) {
			response := isolationRequest(handler, http.MethodGet, tc.path, "")
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d; body=%q", response.Code, response.Body.String())
			}
			var items []map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &items); err != nil || len(items) != 1 {
				t.Fatalf("response = %q; error=%v", response.Body.String(), err)
			}
			if items[0]["id"] != tc.wantID {
				t.Fatalf("response = %s, want only id %q", response.Body.String(), tc.wantID)
			}
		})
	}
}

func TestLedgerAPIRejectsCrossUserResourceAccess(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body, wantError string
		wantStatus                          int
	}{
		{name: "update category", method: http.MethodPatch, path: "/categories/" + isolationCategoryB, body: `{"name":"Stolen"}`, wantStatus: http.StatusNotFound, wantError: "category_not_found"},
		{name: "delete category", method: http.MethodDelete, path: "/categories/" + isolationCategoryB, wantStatus: http.StatusNotFound, wantError: "category_not_found"},
		{name: "update wallet", method: http.MethodPatch, path: "/wallets/" + isolationWalletB, body: `{"name":"Stolen"}`, wantStatus: http.StatusNotFound, wantError: "wallet_not_found"},
		{name: "delete wallet", method: http.MethodDelete, path: "/wallets/" + isolationWalletB, wantStatus: http.StatusNotFound, wantError: "wallet_not_found"},
		{name: "get transaction", method: http.MethodGet, path: "/transactions/" + isolationTransactionB, wantStatus: http.StatusNotFound, wantError: "transaction_not_found"},
		{name: "update transaction", method: http.MethodPatch, path: "/transactions/" + isolationTransactionB, body: `{"amount":"99.00"}`, wantStatus: http.StatusNotFound, wantError: "transaction_not_found"},
		{name: "delete transaction", method: http.MethodDelete, path: "/transactions/" + isolationTransactionB, wantStatus: http.StatusNotFound, wantError: "transaction_not_found"},
		{name: "create transaction with foreign references", method: http.MethodPost, path: "/transactions", body: `{"categoryId":"` + isolationCategoryB + `","walletId":"` + isolationWalletB + `","type":"expense","amount":"10.00","transactionDate":"2026-09-27"}`, wantStatus: http.StatusBadRequest, wantError: "invalid_transaction_reference"},
		{name: "update budget", method: http.MethodPatch, path: "/budgets/" + isolationBudgetB, body: `{"amount":"99.00"}`, wantStatus: http.StatusNotFound, wantError: "budget_not_found"},
		{name: "delete budget", method: http.MethodDelete, path: "/budgets/" + isolationBudgetB, wantStatus: http.StatusNotFound, wantError: "budget_not_found"},
		{name: "create budget with foreign category", method: http.MethodPost, path: "/budgets", body: `{"categoryId":"` + isolationCategoryB + `","amount":"100.00","month":"2026-09"}`, wantStatus: http.StatusBadRequest, wantError: "invalid_budget_reference"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := isolationRequest(newIsolationHandler(), tc.method, tc.path, tc.body)
			if response.Code != tc.wantStatus || response.Body.String() != `{"error":"`+tc.wantError+`"}`+"\n" {
				t.Fatalf("response = %d %q", response.Code, response.Body.String())
			}
		})
	}
}

func TestLedgerAPIRejectsInvalidUserIdentity(t *testing.T) {
	handler := newIsolationHandler()
	for _, path := range []string{"/categories", "/wallets", "/transactions", "/budgets"} {
		for _, userID := range []string{"", "not-a-uuid", "00000000-0000-0000-0000-000000000000"} {
			t.Run(path+"/"+userID, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, path, nil)
				if userID != "" {
					request.Header.Set("X-User-ID", userID)
				}
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, request)
				if response.Code != http.StatusUnauthorized {
					t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
				}
			})
		}
	}
}

func newIsolationHandler() http.Handler {
	categoriesByID := map[string]ownedCategory{
		isolationCategoryA: {owner: isolationUserA, value: categories.Category{ID: isolationCategoryA, Name: "User A expense", Type: "expense", CreatedAt: apiTestTime, UpdatedAt: apiTestTime}},
		isolationCategoryB: {owner: isolationUserB, value: categories.Category{ID: isolationCategoryB, Name: "User B expense", Type: "expense", CreatedAt: apiTestTime, UpdatedAt: apiTestTime}},
	}
	walletsByID := map[string]ownedWallet{
		isolationWalletA: {owner: isolationUserA, value: wallets.Wallet{ID: isolationWalletA, Name: "User A wallet", Type: "cash", Balance: "100.0000", CurrencyCode: "THB", CreatedAt: apiTestTime, UpdatedAt: apiTestTime}},
		isolationWalletB: {owner: isolationUserB, value: wallets.Wallet{ID: isolationWalletB, Name: "User B wallet", Type: "cash", Balance: "200.0000", CurrencyCode: "THB", CreatedAt: apiTestTime, UpdatedAt: apiTestTime}},
	}
	categoryOwners := map[string]string{isolationCategoryA: isolationUserA, isolationCategoryB: isolationUserB}
	walletOwners := map[string]string{isolationWalletA: isolationUserA, isolationWalletB: isolationUserB}
	transactionRepository := &isolationTransactionRepository{
		categoryOwners: categoryOwners,
		walletOwners:   walletOwners,
		items: map[string]ownedTransaction{
			isolationTransactionA: {owner: isolationUserA, value: isolationTransaction(isolationTransactionA, isolationCategoryA, isolationWalletA)},
			isolationTransactionB: {owner: isolationUserB, value: isolationTransaction(isolationTransactionB, isolationCategoryB, isolationWalletB)},
		},
	}
	budgetRepository := &isolationBudgetRepository{
		categoryOwners: categoryOwners,
		items: map[string]ownedBudget{
			isolationBudgetA: {owner: isolationUserA, value: isolationBudget(isolationBudgetA, isolationCategoryA)},
			isolationBudgetB: {owner: isolationUserB, value: isolationBudget(isolationBudgetB, isolationCategoryB)},
		},
	}
	return New(
		func(context.Context) error { return nil },
		categories.NewHandler(categories.NewService(&isolationCategoryRepository{items: categoriesByID})),
		wallets.NewHandler(wallets.NewService(&isolationWalletRepository{items: walletsByID})),
		transactions.NewHandler(transactions.NewService(transactionRepository)),
		budgets.NewHandler(budgets.NewService(budgetRepository)),
	)
}

func isolationRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("X-User-ID", isolationUserA)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func isolationTransaction(id, categoryID, walletID string) transactions.Transaction {
	return transactions.Transaction{ID: id, CategoryID: categoryID, WalletID: walletID, Type: "expense", Amount: "10.0000", TransactionDate: apiTestTime, CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
}

func isolationBudget(id, categoryID string) budgets.Budget {
	return budgets.Budget{ID: id, CategoryID: categoryID, Amount: "100.0000", MonthStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), CreatedAt: apiTestTime, UpdatedAt: apiTestTime}
}
