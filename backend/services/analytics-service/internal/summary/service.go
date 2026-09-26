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

func (s *Service) MonthlyComparison(ctx context.Context, userID, requestID string, month time.Time) (MonthlyComparison, error) {
	previousMonth := time.Date(month.Year(), month.Month()-1, 1, 0, 0, 0, 0, time.UTC)

	currentTransactions, err := s.ledger.ListTransactions(ctx, userID, requestID, month)
	if err != nil {
		return MonthlyComparison{}, err
	}
	previousTransactions, err := s.ledger.ListTransactions(ctx, userID, requestID, previousMonth)
	if err != nil {
		return MonthlyComparison{}, err
	}

	current, err := calculateTotals(currentTransactions)
	if err != nil {
		return MonthlyComparison{}, err
	}
	previous, err := calculateTotals(previousTransactions)
	if err != nil {
		return MonthlyComparison{}, err
	}

	return MonthlyComparison{
		Month: month.Format("2006-01"), PreviousMonth: previousMonth.Format("2006-01"),
		Current:  totalsResponse(current),
		Previous: totalsResponse(previous),
		Changes: MonthlyChanges{
			Income:  amountChange(current.income, previous.income),
			Expense: amountChange(current.expense, previous.expense),
		},
	}, nil
}

func (s *Service) Trends(ctx context.Context, userID, requestID string, month time.Time) (TrendSeries, error) {
	transactions, err := s.ledger.ListTransactions(ctx, userID, requestID, month)
	if err != nil {
		return TrendSeries{}, err
	}

	type dailyTotal struct {
		income  *big.Rat
		expense *big.Rat
	}
	daysInMonth := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	daily := make([]dailyTotal, daysInMonth+1)
	for day := 1; day <= daysInMonth; day++ {
		daily[day] = dailyTotal{income: new(big.Rat), expense: new(big.Rat)}
	}

	for _, transaction := range transactions {
		date, err := time.Parse(time.DateOnly, transaction.TransactionDate)
		if err != nil || date.Year() != month.Year() || date.Month() != month.Month() {
			return TrendSeries{}, ErrInvalidLedgerData
		}
		amount, ok := parseAmount(transaction.Amount)
		if !ok {
			return TrendSeries{}, ErrInvalidLedgerData
		}
		switch transaction.Type {
		case "income":
			daily[date.Day()].income.Add(daily[date.Day()].income, amount)
		case "expense":
			daily[date.Day()].expense.Add(daily[date.Day()].expense, amount)
		default:
			return TrendSeries{}, ErrInvalidLedgerData
		}
	}

	result := TrendSeries{
		Month: month.Format("2006-01"), Granularity: "day",
		Points: make([]TrendPoint, 0, daysInMonth),
	}
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.UTC)
		result.Points = append(result.Points, TrendPoint{
			Date: date.Format(time.DateOnly), Income: daily[day].income.FloatString(4),
			Expense: daily[day].expense.FloatString(4),
		})
	}
	return result, nil
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
	totals, err := calculateTotals(transactions)
	if err != nil {
		return MonthlySummary{}, err
	}

	return MonthlySummary{
		Month:   month.Format("2006-01"),
		Income:  totals.income.FloatString(4),
		Expense: totals.expense.FloatString(4),
	}, nil
}

type transactionTotals struct {
	income  *big.Rat
	expense *big.Rat
}

func calculateTotals(transactions []ledger.Transaction) (transactionTotals, error) {
	totals := transactionTotals{income: new(big.Rat), expense: new(big.Rat)}
	for _, transaction := range transactions {
		amount, ok := parseAmount(transaction.Amount)
		if !ok {
			return transactionTotals{}, ErrInvalidLedgerData
		}
		switch transaction.Type {
		case "income":
			totals.income.Add(totals.income, amount)
		case "expense":
			totals.expense.Add(totals.expense, amount)
		default:
			return transactionTotals{}, ErrInvalidLedgerData
		}
	}
	return totals, nil
}

func totalsResponse(totals transactionTotals) MonthlyTotals {
	return MonthlyTotals{Income: totals.income.FloatString(4), Expense: totals.expense.FloatString(4)}
}

func amountChange(current, previous *big.Rat) AmountChange {
	change := new(big.Rat).Sub(current, previous)
	result := AmountChange{Amount: change.FloatString(4)}
	if previous.Sign() == 0 {
		return result
	}
	percentage := new(big.Rat).Mul(change, big.NewRat(100, 1))
	percentage.Quo(percentage, previous)
	formatted := percentage.FloatString(2)
	result.Percentage = &formatted
	return result
}

func parseAmount(value string) (*big.Rat, bool) {
	if !amountPattern.MatchString(value) {
		return nil, false
	}
	amount, ok := new(big.Rat).SetString(value)
	return amount, ok && amount.Sign() > 0
}
