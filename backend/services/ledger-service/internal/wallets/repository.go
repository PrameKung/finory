package wallets

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

// CreateDefaultCash creates a user's initial THB cash wallet with a zero
// balance. It is safe to call more than once because an existing name is kept.
func (r *Repository) CreateDefaultCash(ctx context.Context, userID string) (int64, error) {
	var id pgtype.UUID
	if err := id.Scan(userID); err != nil {
		return 0, err
	}
	return r.queries.CreateDefaultCashWallet(ctx, id)
}
