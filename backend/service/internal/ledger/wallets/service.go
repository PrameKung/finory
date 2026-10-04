package wallets

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidWallet  = errors.New("invalid wallet")
	ErrWalletNotFound = errors.New("wallet not found")
	ErrWalletConflict = errors.New("wallet already exists")
)

var balancePattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,14})(\.[0-9]{1,4})?$`)
var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

type walletRepository interface {
	CreateDefaultCash(context.Context, string) (int64, error)
	List(context.Context, string) ([]Wallet, error)
	Create(context.Context, string, CreateParams) (Wallet, error)
	Update(context.Context, string, string, UpdateParams) (Wallet, error)
	Delete(context.Context, string, string) (bool, error)
}

type Service struct {
	repository walletRepository
}

func NewService(repository walletRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, userID string) ([]Wallet, error) {
	if _, err := s.repository.CreateDefaultCash(ctx, userID); err != nil {
		return nil, err
	}
	return s.repository.List(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Wallet, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	input.Balance = strings.TrimSpace(input.Balance)
	input.CurrencyCode = strings.ToUpper(strings.TrimSpace(input.CurrencyCode))
	if input.Balance == "" {
		input.Balance = "0"
	}
	if !validName(input.Name) || !validType(input.Type) || !validBalance(input.Balance) || !validCurrencyCode(input.CurrencyCode) {
		return Wallet{}, ErrInvalidWallet
	}
	if _, err := s.repository.CreateDefaultCash(ctx, userID); err != nil {
		return Wallet{}, err
	}
	wallet, err := s.repository.Create(ctx, userID, CreateParams(input))
	if isUniqueViolation(err) {
		return Wallet{}, ErrWalletConflict
	}
	return wallet, err
}

func (s *Service) Update(ctx context.Context, userID, walletID string, input UpdateInput) (Wallet, error) {
	if input.Name == nil && input.Type == nil && input.Balance == nil && input.CurrencyCode == nil {
		return Wallet{}, ErrInvalidWallet
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !validName(value) {
			return Wallet{}, ErrInvalidWallet
		}
		input.Name = &value
	}
	if input.Type != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Type))
		if !validType(value) {
			return Wallet{}, ErrInvalidWallet
		}
		input.Type = &value
	}
	if input.Balance != nil {
		value := strings.TrimSpace(*input.Balance)
		if !validBalance(value) {
			return Wallet{}, ErrInvalidWallet
		}
		input.Balance = &value
	}
	if input.CurrencyCode != nil {
		value := strings.ToUpper(strings.TrimSpace(*input.CurrencyCode))
		if !validCurrencyCode(value) {
			return Wallet{}, ErrInvalidWallet
		}
		input.CurrencyCode = &value
	}
	wallet, err := s.repository.Update(ctx, userID, walletID, UpdateParams(input))
	if errors.Is(err, pgx.ErrNoRows) {
		return Wallet{}, ErrWalletNotFound
	}
	if isUniqueViolation(err) {
		return Wallet{}, ErrWalletConflict
	}
	return wallet, err
}

func (s *Service) Delete(ctx context.Context, userID, walletID string) error {
	deleted, err := s.repository.Delete(ctx, userID, walletID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrWalletNotFound
	}
	return nil
}

func validName(value string) bool {
	return value != "" && len([]rune(value)) <= 100
}

func validType(value string) bool {
	switch value {
	case "cash", "bank", "e_wallet", "other":
		return true
	default:
		return false
	}
}

func validBalance(value string) bool {
	return balancePattern.MatchString(value)
}

func validCurrencyCode(value string) bool {
	return currencyCodePattern.MatchString(value)
}

func isUniqueViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23505"
}
