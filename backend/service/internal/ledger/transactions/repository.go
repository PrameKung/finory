package transactions

import (
	"context"
	"time"

	"finory/backend/service/internal/ledger/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *database.Queries
}

type CreateParams struct {
	CategoryID      string
	WalletID        string
	Type            string
	Amount          string
	Description     string
	TransactionDate time.Time
}

type ListParams struct {
	MonthStart *time.Time
	Type       *string
}

type UpdateParams struct {
	CategoryID      *string
	WalletID        *string
	Type            *string
	Amount          *string
	Description     *string
	TransactionDate *time.Time
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: database.New(db)}
}

func (r *Repository) Create(ctx context.Context, userID string, params CreateParams) (Transaction, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Transaction{}, err
	}
	categoryUUID, err := parseUUID(params.CategoryID)
	if err != nil {
		return Transaction{}, err
	}
	walletUUID, err := parseUUID(params.WalletID)
	if err != nil {
		return Transaction{}, err
	}
	row, err := r.queries.CreateTransaction(ctx, database.CreateTransactionParams{
		UserID: userUUID, CategoryID: categoryUUID, WalletID: walletUUID,
		Type: params.Type, Amount: params.Amount, Description: params.Description,
		TransactionDate: pgtype.Date{Time: params.TransactionDate, Valid: true},
	})
	if err != nil {
		return Transaction{}, err
	}
	return transactionFromCreateRow(row), nil
}

func (r *Repository) List(ctx context.Context, userID string, params ListParams) ([]Transaction, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTransactions(ctx, database.ListTransactionsParams{
		UserID: userUUID, MonthStart: nullableDate(params.MonthStart), Type: nullableText(params.Type),
	})
	if err != nil {
		return nil, err
	}
	transactions := make([]Transaction, 0, len(rows))
	for _, row := range rows {
		transactions = append(transactions, transactionFromListRow(row))
	}
	return transactions, nil
}

func (r *Repository) Get(ctx context.Context, userID, transactionID string) (Transaction, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Transaction{}, err
	}
	transactionUUID, err := parseUUID(transactionID)
	if err != nil {
		return Transaction{}, err
	}
	row, err := r.queries.GetTransaction(ctx, database.GetTransactionParams{ID: transactionUUID, UserID: userUUID})
	if err != nil {
		return Transaction{}, err
	}
	return transactionFromGetRow(row), nil
}

func (r *Repository) Update(ctx context.Context, userID, transactionID string, params UpdateParams) (Transaction, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Transaction{}, err
	}
	transactionUUID, err := parseUUID(transactionID)
	if err != nil {
		return Transaction{}, err
	}
	categoryUUID, err := nullableUUID(params.CategoryID)
	if err != nil {
		return Transaction{}, err
	}
	walletUUID, err := nullableUUID(params.WalletID)
	if err != nil {
		return Transaction{}, err
	}
	row, err := r.queries.UpdateTransaction(ctx, database.UpdateTransactionParams{
		CategoryID: categoryUUID, WalletID: walletUUID, Type: nullableText(params.Type),
		Amount: nullableText(params.Amount), SetDescription: params.Description != nil,
		Description: nullableText(params.Description), TransactionDate: nullableDate(params.TransactionDate),
		ID: transactionUUID, UserID: userUUID,
	})
	if err != nil {
		return Transaction{}, err
	}
	return transactionFromUpdateRow(row), nil
}

func (r *Repository) Delete(ctx context.Context, userID, transactionID string) (bool, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return false, err
	}
	transactionUUID, err := parseUUID(transactionID)
	if err != nil {
		return false, err
	}
	rows, err := r.queries.DeleteTransaction(ctx, database.DeleteTransactionParams{
		ID: transactionUUID, UserID: userUUID,
	})
	return rows > 0, err
}

func parseUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}

func nullableUUID(value *string) (pgtype.UUID, error) {
	if value == nil {
		return pgtype.UUID{}, nil
	}
	return parseUUID(*value)
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func nullableDate(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *value, Valid: true}
}

func optionalText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func transactionFromCreateRow(row database.CreateTransactionRow) Transaction {
	return Transaction{
		ID: row.ID, CategoryID: row.CategoryID, WalletID: row.WalletID,
		Type: row.Type, Amount: row.Amount, Description: optionalText(row.Description),
		TransactionDate: row.TransactionDate.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func transactionFromListRow(row database.ListTransactionsRow) Transaction {
	return Transaction{
		ID: row.ID, CategoryID: row.CategoryID, WalletID: row.WalletID,
		Type: row.Type, Amount: row.Amount, Description: optionalText(row.Description),
		TransactionDate: row.TransactionDate.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func transactionFromGetRow(row database.GetTransactionRow) Transaction {
	return Transaction{
		ID: row.ID, CategoryID: row.CategoryID, WalletID: row.WalletID,
		Type: row.Type, Amount: row.Amount, Description: optionalText(row.Description),
		TransactionDate: row.TransactionDate.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func transactionFromUpdateRow(row database.UpdateTransactionRow) Transaction {
	return Transaction{
		ID: row.ID, CategoryID: row.CategoryID, WalletID: row.WalletID,
		Type: row.Type, Amount: row.Amount, Description: optionalText(row.Description),
		TransactionDate: row.TransactionDate.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}
