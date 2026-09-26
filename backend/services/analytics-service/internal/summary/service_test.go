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
