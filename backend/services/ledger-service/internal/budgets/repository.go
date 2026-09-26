package budgets

import (
	"context"
	"time"

	"finory/backend/services/ledger-service/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *database.Queries
}

type CreateParams struct {
	CategoryID string
	Amount     string
	MonthStart time.Time
}

type ListParams struct {
	MonthStart *time.Time
}

type UpdateParams struct {
	CategoryID *string
	Amount     *string
	MonthStart *time.Time
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: database.New(db)}
}

func (r *Repository) Create(ctx context.Context, userID string, params CreateParams) (Budget, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Budget{}, err
	}
	categoryUUID, err := parseUUID(params.CategoryID)
	if err != nil {
		return Budget{}, err
	}
	row, err := r.queries.CreateBudget(ctx, database.CreateBudgetParams{
		UserID: userUUID, CategoryID: categoryUUID, Amount: params.Amount,
		MonthStart: pgtype.Date{Time: params.MonthStart, Valid: true},
	})
	if err != nil {
		return Budget{}, err
	}
	return budgetFromCreateRow(row), nil
}

func (r *Repository) List(ctx context.Context, userID string, params ListParams) ([]Budget, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListBudgets(ctx, database.ListBudgetsParams{
		UserID: userUUID, MonthStart: nullableDate(params.MonthStart),
	})
	if err != nil {
		return nil, err
	}
	items := make([]Budget, 0, len(rows))
	for _, row := range rows {
		items = append(items, budgetFromListRow(row))
	}
	return items, nil
}

func (r *Repository) Get(ctx context.Context, userID, budgetID string) (Budget, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Budget{}, err
	}
	budgetUUID, err := parseUUID(budgetID)
	if err != nil {
		return Budget{}, err
	}
	row, err := r.queries.GetBudget(ctx, database.GetBudgetParams{ID: budgetUUID, UserID: userUUID})
	if err != nil {
		return Budget{}, err
	}
	return budgetFromGetRow(row), nil
}

func (r *Repository) Update(ctx context.Context, userID, budgetID string, params UpdateParams) (Budget, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Budget{}, err
	}
	budgetUUID, err := parseUUID(budgetID)
	if err != nil {
		return Budget{}, err
	}
	categoryUUID, err := nullableUUID(params.CategoryID)
	if err != nil {
		return Budget{}, err
	}
	row, err := r.queries.UpdateBudget(ctx, database.UpdateBudgetParams{
		CategoryID: categoryUUID, Amount: nullableText(params.Amount),
		MonthStart: nullableDate(params.MonthStart), ID: budgetUUID, UserID: userUUID,
	})
	if err != nil {
		return Budget{}, err
	}
	return budgetFromUpdateRow(row), nil
}

func (r *Repository) Delete(ctx context.Context, userID, budgetID string) (bool, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return false, err
	}
	budgetUUID, err := parseUUID(budgetID)
	if err != nil {
		return false, err
	}
	rows, err := r.queries.DeleteBudget(ctx, database.DeleteBudgetParams{ID: budgetUUID, UserID: userUUID})
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

func budgetFromCreateRow(row database.CreateBudgetRow) Budget {
	return Budget{
		ID: row.ID, CategoryID: row.CategoryID, Amount: row.Amount, MonthStart: row.MonthStart.Time,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func budgetFromListRow(row database.ListBudgetsRow) Budget {
	return Budget{
		ID: row.ID, CategoryID: row.CategoryID, Amount: row.Amount, MonthStart: row.MonthStart.Time,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func budgetFromGetRow(row database.GetBudgetRow) Budget {
	return Budget{
		ID: row.ID, CategoryID: row.CategoryID, Amount: row.Amount, MonthStart: row.MonthStart.Time,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func budgetFromUpdateRow(row database.UpdateBudgetRow) Budget {
	return Budget{
		ID: row.ID, CategoryID: row.CategoryID, Amount: row.Amount, MonthStart: row.MonthStart.Time,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}
