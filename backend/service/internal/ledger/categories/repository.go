package categories

import (
	"context"

	"finory/backend/service/internal/ledger/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *database.Queries
}

type CreateParams struct {
	Name  string
	Type  string
	Icon  string
	Color string
}

type UpdateParams struct {
	Name  *string
	Type  *string
	Icon  *string
	Color *string
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: database.New(db)}
}

// CreateDefaults creates the standard categories for a user. It is safe to
// call more than once because existing category names are left unchanged.
func (r *Repository) CreateDefaults(ctx context.Context, userID string) (int64, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return 0, err
	}
	return r.queries.CreateDefaultCategories(ctx, id)
}

func (r *Repository) List(ctx context.Context, userID string) ([]Category, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListCategories(ctx, id)
	if err != nil {
		return nil, err
	}
	categories := make([]Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, categoryFromListRow(row))
	}
	return categories, nil
}

func (r *Repository) Create(ctx context.Context, userID string, params CreateParams) (Category, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return Category{}, err
	}
	row, err := r.queries.CreateCategory(ctx, database.CreateCategoryParams{
		UserID: id, Name: params.Name, Type: params.Type, Icon: params.Icon, Color: params.Color,
	})
	if err != nil {
		return Category{}, err
	}
	return categoryFromCreateRow(row), nil
}

func (r *Repository) Update(ctx context.Context, userID, categoryID string, params UpdateParams) (Category, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return Category{}, err
	}
	categoryUUID, err := parseUUID(categoryID)
	if err != nil {
		return Category{}, err
	}
	row, err := r.queries.UpdateCategory(ctx, database.UpdateCategoryParams{
		Name: nullableText(params.Name), Type: nullableText(params.Type),
		Icon: nullableText(params.Icon), Color: nullableText(params.Color),
		ID: categoryUUID, UserID: userUUID,
	})
	if err != nil {
		return Category{}, err
	}
	return categoryFromUpdateRow(row), nil
}

func (r *Repository) Delete(ctx context.Context, userID, categoryID string) (bool, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return false, err
	}
	categoryUUID, err := parseUUID(categoryID)
	if err != nil {
		return false, err
	}
	rows, err := r.queries.DeleteCategory(ctx, database.DeleteCategoryParams{ID: categoryUUID, UserID: userUUID})
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

func optionalText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func categoryFromListRow(row database.ListCategoriesRow) Category {
	return Category{
		ID: row.ID, Name: row.Name, Type: row.Type,
		Icon: optionalText(row.Icon), Color: optionalText(row.Color), IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func categoryFromCreateRow(row database.CreateCategoryRow) Category {
	return Category{
		ID: row.ID, Name: row.Name, Type: row.Type,
		Icon: optionalText(row.Icon), Color: optionalText(row.Color), IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func categoryFromUpdateRow(row database.UpdateCategoryRow) Category {
	return Category{
		ID: row.ID, Name: row.Name, Type: row.Type,
		Icon: optionalText(row.Icon), Color: optionalText(row.Color), IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}
