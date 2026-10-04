package wallets

import (
	"context"

	"finory/backend/service/internal/ledger/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *database.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: database.New(db)}
}

type CreateParams struct {
	Name         string
	Type         string
	Balance      string
	CurrencyCode string
}

type UpdateParams struct {
	Name         *string
	Type         *string
	Balance      *string
	CurrencyCode *string
}

// CreateDefaultCash creates a user's initial THB cash wallet with a zero
// balance. It is safe to call more than once because an existing name is kept.
func (r *Repository) CreateDefaultCash(ctx context.Context, userID string) (int64, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return 0, err
	}
	return r.queries.CreateDefaultCashWallet(ctx, id)
}

func (r *Repository) List(ctx context.Context, userID string) ([]Wallet, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListWallets(ctx, id)
	if err != nil {
		return nil, err
	}
	wallets := make([]Wallet, 0, len(rows))
	for _, row := range rows {
		wallets = append(wallets, walletFromListRow(row))
	}
	return wallets, nil
}

func (r *Repository) Create(ctx context.Context, userID string, params CreateParams) (Wallet, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return Wallet{}, err
	}
	row, err := r.queries.CreateWallet(ctx, database.CreateWalletParams{
		UserID: id, Name: params.Name, Type: params.Type,
		Balance: params.Balance, CurrencyCode: params.CurrencyCode,
	})
	if err != nil {
		return Wallet{}, err
	}
	return walletFromCreateRow(row), nil
}

func (r *Repository) Update(ctx context.Context, userID, walletID string, params UpdateParams) (Wallet, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Wallet{}, err
	}
	walletUUID, err := parseUUID(walletID)
	if err != nil {
		return Wallet{}, err
	}
	row, err := r.queries.UpdateWallet(ctx, database.UpdateWalletParams{
		Name: nullableText(params.Name), Type: nullableText(params.Type),
		Balance: nullableText(params.Balance), CurrencyCode: nullableText(params.CurrencyCode),
		ID: walletUUID, UserID: userUUID,
	})
	if err != nil {
		return Wallet{}, err
	}
	return walletFromUpdateRow(row), nil
}

func (r *Repository) Delete(ctx context.Context, userID, walletID string) (bool, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return false, err
	}
	walletUUID, err := parseUUID(walletID)
	if err != nil {
		return false, err
	}
	rows, err := r.queries.DeleteWallet(ctx, database.DeleteWalletParams{ID: walletUUID, UserID: userUUID})
	return rows > 0, err
}

func parseUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func walletFromListRow(row database.ListWalletsRow) Wallet {
	return Wallet{
		ID: row.ID, Name: row.Name, Type: row.Type, Balance: row.Balance,
		CurrencyCode: row.CurrencyCode, IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func walletFromCreateRow(row database.CreateWalletRow) Wallet {
	return Wallet{
		ID: row.ID, Name: row.Name, Type: row.Type, Balance: row.Balance,
		CurrencyCode: row.CurrencyCode, IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func walletFromUpdateRow(row database.UpdateWalletRow) Wallet {
	return Wallet{
		ID: row.ID, Name: row.Name, Type: row.Type, Balance: row.Balance,
		CurrencyCode: row.CurrencyCode, IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}
