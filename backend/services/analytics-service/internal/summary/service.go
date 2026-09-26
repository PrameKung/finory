package summary

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"time"

	"finory/backend/services/analytics-service/internal/ledger"
)

var (
	ErrInvalidLedgerData = errors.New("invalid ledger data")
	amountPattern        = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,4})?$`)
)

type transactionReader interface {
	ListTransactions(context.Context, string, string, time.Time) ([]ledger.Transaction, error)
}

type Service struct {
	ledger transactionReader
}

func NewService(ledgerClient transactionReader) *Service {
	return &Service{ledger: ledgerClient}
}

func (s *Service) Monthly(ctx context.Context, userID, requestID string, month time.Time) (MonthlySummary, error) {
	transactions, err := s.ledger.ListTransactions(ctx, userID, requestID, month)
	if err != nil {
		return MonthlySummary{}, err
	}

	income := new(big.Rat)
	expense := new(big.Rat)
	for _, transaction := range transactions {
		amount, ok := parseAmount(transaction.Amount)
		if !ok {
			return MonthlySummary{}, ErrInvalidLedgerData
		}
		switch transaction.Type {
		case "income":
			income.Add(income, amount)
		case "expense":
			expense.Add(expense, amount)
		default:
			return MonthlySummary{}, ErrInvalidLedgerData
		}
	}

	return MonthlySummary{
		Month:   month.Format("2006-01"),
		Income:  income.FloatString(4),
		Expense: expense.FloatString(4),
	}, nil
}

func parseAmount(value string) (*big.Rat, bool) {
	if !amountPattern.MatchString(value) {
		return nil, false
	}
	amount, ok := new(big.Rat).SetString(value)
	return amount, ok && amount.Sign() > 0
}
