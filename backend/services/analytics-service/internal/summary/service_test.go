package summary

import (
	"context"
	"errors"
	"testing"
	"time"

	"finory/backend/services/analytics-service/internal/ledger"
)

type fakeLedger struct {
	transactions []ledger.Transaction
	monthlyData  ledger.MonthlyData
	err          error
	userID       string
	requestID    string
	month        time.Time
}

func (f *fakeLedger) ListTransactions(_ context.Context, userID, requestID string, month time.Time) ([]ledger.Transaction, error) {
	f.userID = userID
	f.requestID = requestID
	f.month = month
	return f.transactions, f.err
}

func (f *fakeLedger) GetMonthlyData(_ context.Context, userID, requestID string, month time.Time) (ledger.MonthlyData, error) {
	f.userID = userID
	f.requestID = requestID
	f.month = month
	return f.monthlyData, f.err
}

func TestMonthlySummary(t *testing.T) {
	month := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	ledgerClient := &fakeLedger{transactions: []ledger.Transaction{
		{Type: "income", Amount: "0.1000"},
		{Type: "income", Amount: "0.2000"},
		{Type: "income", Amount: "1000"},
		{Type: "expense", Amount: "25.5050"},
		{Type: "expense", Amount: "4.495"},
	}}

	result, err := NewService(ledgerClient).Monthly(context.Background(), "user-1", "request-1", month)
	if err != nil {
		t.Fatal(err)
	}
	if result != (MonthlySummary{Month: "2026-09", Income: "1000.3000", Expense: "30.0000"}) {
		t.Fatalf("unexpected summary: %+v", result)
	}
	if ledgerClient.userID != "user-1" || ledgerClient.requestID != "request-1" || !ledgerClient.month.Equal(month) {
		t.Fatalf("ledger scope was not forwarded: %+v", ledgerClient)
	}
}

func TestMonthlySummaryRejectsInvalidLedgerData(t *testing.T) {
	for name, transaction := range map[string]ledger.Transaction{
		"invalid amount":  {Type: "income", Amount: "12.34567"},
		"negative amount": {Type: "expense", Amount: "-1.00"},
		"invalid type":    {Type: "transfer", Amount: "1.00"},
	} {
		t.Run(name, func(t *testing.T) {
			service := NewService(&fakeLedger{transactions: []ledger.Transaction{transaction}})
			_, err := service.Monthly(context.Background(), "user-1", "", time.Now())
			if !errors.Is(err, ErrInvalidLedgerData) {
				t.Fatalf("expected invalid ledger data, got %v", err)
			}
		})
	}
}

func TestMonthlySummaryPropagatesLedgerError(t *testing.T) {
	want := errors.New("ledger unavailable")
	_, err := NewService(&fakeLedger{err: want}).Monthly(context.Background(), "user-1", "", time.Now())
	if !errors.Is(err, want) {
		t.Fatalf("expected ledger error, got %v", err)
	}
}

func TestCategoryDistributionAndRanking(t *testing.T) {
	month := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	foodIcon, foodColor := "utensils", "#ff0000"
	ledgerClient := &fakeLedger{monthlyData: ledger.MonthlyData{
		Categories: []ledger.Category{
			{ID: "food", Name: "Food", Type: "expense", Icon: &foodIcon, Color: &foodColor},
			{ID: "rent", Name: "Rent", Type: "expense"},
			{ID: "salary", Name: "Salary", Type: "income"},
			{ID: "unused", Name: "Unused", Type: "expense"},
		},
		Transactions: []ledger.Transaction{
			{CategoryID: "food", Type: "expense", Amount: "20.0000"},
			{CategoryID: "food", Type: "expense", Amount: "40"},
			{CategoryID: "rent", Type: "expense", Amount: "30.0000"},
			{CategoryID: "salary", Type: "income", Amount: "1000.0000"},
		},
	}}

	result, err := NewService(ledgerClient).Categories(context.Background(), "user-1", "request-1", month)
	if err != nil {
		t.Fatal(err)
	}
	if result.Month != "2026-09" || result.TotalExpense != "90.0000" || len(result.Categories) != 2 {
		t.Fatalf("unexpected distribution: %+v", result)
	}
	if result.Categories[0] != (CategoryRank{
		Rank: 1, CategoryID: "food", Name: "Food", Icon: &foodIcon, Color: &foodColor,
		Amount: "60.0000", Percentage: "66.67",
	}) {
		t.Fatalf("unexpected first rank: %+v", result.Categories[0])
	}
	if result.Categories[1].Rank != 2 || result.Categories[1].CategoryID != "rent" ||
		result.Categories[1].Amount != "30.0000" || result.Categories[1].Percentage != "33.33" {
		t.Fatalf("unexpected second rank: %+v", result.Categories[1])
	}
	if ledgerClient.userID != "user-1" || ledgerClient.requestID != "request-1" || !ledgerClient.month.Equal(month) {
		t.Fatalf("ledger scope was not forwarded: %+v", ledgerClient)
	}
}

func TestCategoryDistributionReturnsEmptyListWithoutExpenses(t *testing.T) {
	service := NewService(&fakeLedger{monthlyData: ledger.MonthlyData{
		Categories:   []ledger.Category{{ID: "salary", Name: "Salary", Type: "income"}},
		Transactions: []ledger.Transaction{{CategoryID: "salary", Type: "income", Amount: "100.00"}},
	}})
	result, err := service.Categories(context.Background(), "user-1", "", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalExpense != "0.0000" || result.Categories == nil || len(result.Categories) != 0 {
		t.Fatalf("unexpected empty distribution: %+v", result)
	}
}

func TestCategoryDistributionRejectsInvalidLedgerData(t *testing.T) {
	for name, data := range map[string]ledger.MonthlyData{
		"missing category": {
			Transactions: []ledger.Transaction{{CategoryID: "missing", Type: "expense", Amount: "1.00"}},
		},
		"wrong category type": {
			Categories:   []ledger.Category{{ID: "salary", Name: "Salary", Type: "income"}},
			Transactions: []ledger.Transaction{{CategoryID: "salary", Type: "expense", Amount: "1.00"}},
		},
		"duplicate category": {
			Categories: []ledger.Category{{ID: "food", Name: "Food", Type: "expense"}, {ID: "food", Name: "Food", Type: "expense"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewService(&fakeLedger{monthlyData: data}).Categories(context.Background(), "user-1", "", time.Now())
			if !errors.Is(err, ErrInvalidLedgerData) {
				t.Fatalf("expected invalid ledger data, got %v", err)
			}
		})
	}
}

func TestDailyIncomeExpenseTrends(t *testing.T) {
	month := time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)
	ledgerClient := &fakeLedger{transactions: []ledger.Transaction{
		{Type: "income", Amount: "100.0000", TransactionDate: "2024-02-01"},
		{Type: "income", Amount: "0.2500", TransactionDate: "2024-02-01"},
		{Type: "expense", Amount: "40.5000", TransactionDate: "2024-02-01"},
		{Type: "expense", Amount: "10", TransactionDate: "2024-02-29"},
	}}

	result, err := NewService(ledgerClient).Trends(context.Background(), "user-1", "request-1", month)
	if err != nil {
		t.Fatal(err)
	}
	if result.Month != "2024-02" || result.Granularity != "day" || len(result.Points) != 29 {
		t.Fatalf("unexpected trend series: %+v", result)
	}
	if result.Points[0] != (TrendPoint{Date: "2024-02-01", Income: "100.2500", Expense: "40.5000"}) {
		t.Fatalf("unexpected first point: %+v", result.Points[0])
	}
	if result.Points[1] != (TrendPoint{Date: "2024-02-02", Income: "0.0000", Expense: "0.0000"}) {
		t.Fatalf("unexpected zero point: %+v", result.Points[1])
	}
	if result.Points[28] != (TrendPoint{Date: "2024-02-29", Income: "0.0000", Expense: "10.0000"}) {
		t.Fatalf("unexpected last point: %+v", result.Points[28])
	}
	if ledgerClient.userID != "user-1" || ledgerClient.requestID != "request-1" || !ledgerClient.month.Equal(month) {
		t.Fatalf("ledger scope was not forwarded: %+v", ledgerClient)
	}
}

func TestDailyTrendsRejectInvalidLedgerData(t *testing.T) {
	month := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	for name, transaction := range map[string]ledger.Transaction{
		"invalid date":       {Type: "income", Amount: "1.00", TransactionDate: "not-a-date"},
		"date outside month": {Type: "expense", Amount: "1.00", TransactionDate: "2026-08-31"},
		"invalid type":       {Type: "transfer", Amount: "1.00", TransactionDate: "2026-09-01"},
	} {
		t.Run(name, func(t *testing.T) {
			service := NewService(&fakeLedger{transactions: []ledger.Transaction{transaction}})
			_, err := service.Trends(context.Background(), "user-1", "", month)
			if !errors.Is(err, ErrInvalidLedgerData) {
				t.Fatalf("expected invalid ledger data, got %v", err)
			}
		})
	}
}
