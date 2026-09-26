package categories

import (
	"context"

	"finory/backend/services/ledger-service/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *database.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: database.New(db)}
}

// CreateDefaults creates the standard categories for a user. It is safe to
// call more than once because existing category names are left unchanged.
func (r *Repository) CreateDefaults(ctx context.Context, userID string) (int64, error) {
	var id pgtype.UUID
	if err := id.Scan(userID); err != nil {
		return 0, err
	}
	return r.queries.CreateDefaultCategories(ctx, id)
}
