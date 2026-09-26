package summary

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"

	"finory/backend/services/analytics-service/internal/ledger"
)

var (
	ErrInvalidLedgerData = errors.New("invalid ledger data")
	amountPattern        = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,4})?$`)
)

type ledgerReader interface {
	ListTransactions(context.Context, string, string, time.Time) ([]ledger.Transaction, error)
	GetMonthlyData(context.Context, string, string, time.Time) (ledger.MonthlyData, error)
}

type Service struct {
	ledger ledgerReader
}

func NewService(ledgerClient ledgerReader) *Service {
	return &Service{ledger: ledgerClient}
}

func (s *Service) Categories(ctx context.Context, userID, requestID string, month time.Time) (CategoryDistribution, error) {
	data, err := s.ledger.GetMonthlyData(ctx, userID, requestID, month)
	if err != nil {
		return CategoryDistribution{}, err
	}

	categories := make(map[string]ledger.Category, len(data.Categories))
	for _, category := range data.Categories {
		if category.ID == "" || strings.TrimSpace(category.Name) == "" ||
			(category.Type != "income" && category.Type != "expense") {
			return CategoryDistribution{}, ErrInvalidLedgerData
		}
		if _, exists := categories[category.ID]; exists {
			return CategoryDistribution{}, ErrInvalidLedgerData
		}
		categories[category.ID] = category
	}

	type categoryTotal struct {
		category ledger.Category
		amount   *big.Rat
	}
	totals := make(map[string]*categoryTotal)
	totalExpense := new(big.Rat)
	for _, transaction := range data.Transactions {
		amount, ok := parseAmount(transaction.Amount)
		if !ok {
			return CategoryDistribution{}, ErrInvalidLedgerData
		}
		switch transaction.Type {
		case "income":
			continue
		case "expense":
			category, exists := categories[transaction.CategoryID]
			if !exists || category.Type != "expense" {
				return CategoryDistribution{}, ErrInvalidLedgerData
			}
			total := totals[category.ID]
			if total == nil {
				total = &categoryTotal{category: category, amount: new(big.Rat)}
				totals[category.ID] = total
			}
			total.amount.Add(total.amount, amount)
			totalExpense.Add(totalExpense, amount)
		default:
			return CategoryDistribution{}, ErrInvalidLedgerData
		}
	}

	ranked := make([]*categoryTotal, 0, len(totals))
	for _, total := range totals {
		ranked = append(ranked, total)
	}
	sort.Slice(ranked, func(i, j int) bool {
		if comparison := ranked[i].amount.Cmp(ranked[j].amount); comparison != 0 {
			return comparison > 0
		}
		if ranked[i].category.Name != ranked[j].category.Name {
			return ranked[i].category.Name < ranked[j].category.Name
		}
		return ranked[i].category.ID < ranked[j].category.ID
	})

	result := CategoryDistribution{
		Month:        month.Format("2006-01"),
		TotalExpense: totalExpense.FloatString(4),
		Categories:   make([]CategoryRank, 0, len(ranked)),
	}
	for index, total := range ranked {
		percentage := new(big.Rat).Mul(total.amount, big.NewRat(100, 1))
		percentage.Quo(percentage, totalExpense)
		result.Categories = append(result.Categories, CategoryRank{
			Rank: index + 1, CategoryID: total.category.ID, Name: total.category.Name,
			Icon: total.category.Icon, Color: total.category.Color,
			Amount: total.amount.FloatString(4), Percentage: percentage.FloatString(2),
		})
	}
	return result, nil
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
