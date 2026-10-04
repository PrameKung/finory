package transactions

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
	ErrInvalidTransaction          = errors.New("invalid transaction")
	ErrInvalidTransactionReference = errors.New("invalid transaction reference")
	ErrTransactionNotFound         = errors.New("transaction not found")
)

var amountPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,14})(\.[0-9]{1,4})?$`)

type transactionRepository interface {
	Create(context.Context, string, CreateParams) (Transaction, error)
	List(context.Context, string, ListParams) ([]Transaction, error)
	Get(context.Context, string, string) (Transaction, error)
	Update(context.Context, string, string, UpdateParams) (Transaction, error)
	Delete(context.Context, string, string) (bool, error)
}

type Service struct {
	repository transactionRepository
}

func NewService(repository transactionRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Transaction, error) {
	input.CategoryID = strings.TrimSpace(input.CategoryID)
	input.WalletID = strings.TrimSpace(input.WalletID)
	input.Type = normalizeType(input.Type)
	input.Amount = strings.TrimSpace(input.Amount)
	input.Description = strings.TrimSpace(input.Description)
	date, dateOK := parseDate(input.TransactionDate)
	if !validUUID(input.CategoryID) || !validUUID(input.WalletID) || !validType(input.Type) ||
		!validAmount(input.Amount) || !validDescription(input.Description) || !dateOK {
		return Transaction{}, ErrInvalidTransaction
	}
	transaction, err := s.repository.Create(ctx, userID, CreateParams{
		CategoryID: input.CategoryID, WalletID: input.WalletID, Type: input.Type,
		Amount: input.Amount, Description: input.Description, TransactionDate: date,
	})
	if errors.Is(err, pgx.ErrNoRows) || isForeignKeyViolation(err) {
		return Transaction{}, ErrInvalidTransactionReference
	}
	return transaction, err
}

func (s *Service) List(ctx context.Context, userID string, filter ListFilter) ([]Transaction, error) {
	params := ListParams{}
	filter.Month = strings.TrimSpace(filter.Month)
	filter.Type = normalizeType(filter.Type)
	if filter.Month != "" {
		month, err := time.Parse("2006-01", filter.Month)
		if err != nil {
			return nil, ErrInvalidTransaction
		}
		params.MonthStart = &month
	}
	if filter.Type != "" {
		if !validType(filter.Type) {
			return nil, ErrInvalidTransaction
		}
		params.Type = &filter.Type
	}
	return s.repository.List(ctx, userID, params)
}

func (s *Service) Get(ctx context.Context, userID, transactionID string) (Transaction, error) {
	transaction, err := s.repository.Get(ctx, userID, transactionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrTransactionNotFound
	}
	return transaction, err
}

func (s *Service) Update(ctx context.Context, userID, transactionID string, input UpdateInput) (Transaction, error) {
	if input.CategoryID == nil && input.WalletID == nil && input.Type == nil && input.Amount == nil &&
		input.Description == nil && input.TransactionDate == nil {
		return Transaction{}, ErrInvalidTransaction
	}
	params := UpdateParams{}
	if input.CategoryID != nil {
		value := strings.TrimSpace(*input.CategoryID)
		if !validUUID(value) {
			return Transaction{}, ErrInvalidTransaction
		}
		params.CategoryID = &value
	}
	if input.WalletID != nil {
		value := strings.TrimSpace(*input.WalletID)
		if !validUUID(value) {
			return Transaction{}, ErrInvalidTransaction
		}
		params.WalletID = &value
	}
	if input.Type != nil {
		value := normalizeType(*input.Type)
		if !validType(value) {
			return Transaction{}, ErrInvalidTransaction
		}
		params.Type = &value
	}
	if input.Amount != nil {
		value := strings.TrimSpace(*input.Amount)
		if !validAmount(value) {
			return Transaction{}, ErrInvalidTransaction
		}
		params.Amount = &value
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		if !validDescription(value) {
			return Transaction{}, ErrInvalidTransaction
		}
		params.Description = &value
	}
	if input.TransactionDate != nil {
		value, ok := parseDate(*input.TransactionDate)
		if !ok {
			return Transaction{}, ErrInvalidTransaction
		}
		params.TransactionDate = &value
	}

	transaction, err := s.repository.Update(ctx, userID, transactionID, params)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, getErr := s.repository.Get(ctx, userID, transactionID); errors.Is(getErr, pgx.ErrNoRows) {
			return Transaction{}, ErrTransactionNotFound
		} else if getErr != nil {
			return Transaction{}, getErr
		}
		return Transaction{}, ErrInvalidTransactionReference
	}
	if isForeignKeyViolation(err) {
		return Transaction{}, ErrInvalidTransactionReference
	}
	return transaction, err
}

func (s *Service) Delete(ctx context.Context, userID, transactionID string) error {
	deleted, err := s.repository.Delete(ctx, userID, transactionID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrTransactionNotFound
	}
	return nil
}

func normalizeType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validType(value string) bool {
	return value == "income" || value == "expense"
}

func validAmount(value string) bool {
	if !amountPattern.MatchString(value) {
		return false
	}
	return strings.ContainsAny(value, "123456789")
}

func validDescription(value string) bool {
	return len([]rune(value)) <= 500
}

func parseDate(value string) (time.Time, bool) {
	date, err := time.Parse(time.DateOnly, strings.TrimSpace(value))
	return date, err == nil
}

func validUUID(value string) bool {
	var id pgtype.UUID
	return id.Scan(value) == nil && id.Valid && id.Bytes != [16]byte{}
}

func isForeignKeyViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23503"
}
