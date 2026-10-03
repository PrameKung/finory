package budgets

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidBudget          = errors.New("invalid budget")
	ErrInvalidBudgetReference = errors.New("invalid budget reference")
	ErrBudgetNotFound         = errors.New("budget not found")
	ErrBudgetConflict         = errors.New("budget already exists")
)

var amountPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,14})(\.[0-9]{1,4})?$`)

type budgetRepository interface {
	Create(context.Context, string, CreateParams) (Budget, error)
	List(context.Context, string, ListParams) ([]Budget, error)
	Get(context.Context, string, string) (Budget, error)
	Update(context.Context, string, string, UpdateParams) (Budget, error)
	Delete(context.Context, string, string) (bool, error)
}

type Service struct {
	repository budgetRepository
}

func NewService(repository budgetRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Budget, error) {
	input.CategoryID = strings.TrimSpace(input.CategoryID)
	input.Amount = strings.TrimSpace(input.Amount)
	month, monthOK := parseMonth(input.Month)
	if !validUUID(input.CategoryID) || !validAmount(input.Amount) || !monthOK {
		return Budget{}, ErrInvalidBudget
	}
	budget, err := s.repository.Create(ctx, userID, CreateParams{
		CategoryID: input.CategoryID, Amount: input.Amount, MonthStart: month,
	})
	if errors.Is(err, pgx.ErrNoRows) || isForeignKeyViolation(err) {
		return Budget{}, ErrInvalidBudgetReference
	}
	if isUniqueViolation(err) {
		return Budget{}, ErrBudgetConflict
	}
	return budget, err
}

func (s *Service) List(ctx context.Context, userID string, filter ListFilter) ([]Budget, error) {
	params := ListParams{}
	filter.Month = strings.TrimSpace(filter.Month)
	if filter.Month != "" {
		month, ok := parseMonth(filter.Month)
		if !ok {
			return nil, ErrInvalidBudget
		}
		params.MonthStart = &month
	}
	return s.repository.List(ctx, userID, params)
}

func (s *Service) Update(ctx context.Context, userID, budgetID string, input UpdateInput) (Budget, error) {
	if input.CategoryID == nil && input.Amount == nil && input.Month == nil {
		return Budget{}, ErrInvalidBudget
	}
	params := UpdateParams{}
	if input.CategoryID != nil {
		value := strings.TrimSpace(*input.CategoryID)
		if !validUUID(value) {
			return Budget{}, ErrInvalidBudget
		}
		params.CategoryID = &value
	}
	if input.Amount != nil {
		value := strings.TrimSpace(*input.Amount)
		if !validAmount(value) {
			return Budget{}, ErrInvalidBudget
		}
		params.Amount = &value
	}
	if input.Month != nil {
		value, ok := parseMonth(*input.Month)
		if !ok {
			return Budget{}, ErrInvalidBudget
		}
		params.MonthStart = &value
	}

	budget, err := s.repository.Update(ctx, userID, budgetID, params)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, getErr := s.repository.Get(ctx, userID, budgetID); errors.Is(getErr, pgx.ErrNoRows) {
			return Budget{}, ErrBudgetNotFound
		} else if getErr != nil {
			return Budget{}, getErr
		}
		return Budget{}, ErrInvalidBudgetReference
	}
	if isForeignKeyViolation(err) {
		return Budget{}, ErrInvalidBudgetReference
	}
	if isUniqueViolation(err) {
		return Budget{}, ErrBudgetConflict
	}
	return budget, err
}

func (s *Service) Delete(ctx context.Context, userID, budgetID string) error {
	deleted, err := s.repository.Delete(ctx, userID, budgetID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrBudgetNotFound
	}
	return nil
}

func parseMonth(value string) (time.Time, bool) {
	month, err := time.Parse("2006-01", strings.TrimSpace(value))
	return month, err == nil
}

func validAmount(value string) bool {
	if !amountPattern.MatchString(value) {
		return false
	}
	return strings.ContainsAny(value, "123456789")
}

func validUUID(value string) bool {
	var id pgtype.UUID
	return id.Scan(value) == nil && id.Valid && id.Bytes != [16]byte{}
}

func isForeignKeyViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23503"
}

func isUniqueViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23505"
}
