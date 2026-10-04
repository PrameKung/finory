package analytics

import (
	"context"
	"time"

	"finory/backend/service/internal/analytics/ledger"
	"finory/backend/service/internal/ledger/categories"
	"finory/backend/service/internal/ledger/transactions"
)

type transactionLister interface {
	List(context.Context, string, transactions.ListFilter) ([]transactions.Transaction, error)
}

type categoryLister interface {
	List(context.Context, string) ([]categories.Category, error)
}

// LedgerReader adapts the local ledger application services for analytics.
// Analytics reads through those services in-process and makes no HTTP calls.
type LedgerReader struct {
	transactions transactionLister
	categories   categoryLister
}

func NewLedgerReader(transactionService transactionLister, categoryService categoryLister) *LedgerReader {
	return &LedgerReader{transactions: transactionService, categories: categoryService}
}

func (r *LedgerReader) ListTransactions(ctx context.Context, userID, _ string, month time.Time) ([]ledger.Transaction, error) {
	items, err := r.transactions.List(ctx, userID, transactions.ListFilter{Month: month.Format("2006-01")})
	if err != nil {
		return nil, err
	}
	result := make([]ledger.Transaction, 0, len(items))
	for _, item := range items {
		result = append(result, ledger.Transaction{
			ID: item.ID, CategoryID: item.CategoryID, WalletID: item.WalletID,
			Type: item.Type, Amount: item.Amount, Description: item.Description,
			TransactionDate: item.TransactionDate.Format(time.DateOnly),
			CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return result, nil
}

func (r *LedgerReader) GetMonthlyData(ctx context.Context, userID, requestID string, month time.Time) (ledger.MonthlyData, error) {
	transactions, err := r.ListTransactions(ctx, userID, requestID, month)
	if err != nil {
		return ledger.MonthlyData{}, err
	}
	items, err := r.categories.List(ctx, userID)
	if err != nil {
		return ledger.MonthlyData{}, err
	}
	categories := make([]ledger.Category, 0, len(items))
	for _, item := range items {
		categories = append(categories, ledger.Category{
			ID: item.ID, Name: item.Name, Type: item.Type, Icon: item.Icon,
			Color: item.Color, IsDefault: item.IsDefault,
			CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return ledger.MonthlyData{Transactions: transactions, Categories: categories}, nil
}
