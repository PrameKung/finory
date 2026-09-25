package auth

import (
	"context"

	"finory/backend/services/auth-service/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GoogleUser struct {
	Subject     string
	Email       string
	DisplayName string
	AvatarURL   string
}

type Repository struct {
	queries *database.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: database.New(db)}
}

// UpsertGoogleUser uses Google's stable subject as the account key, never email.
func (r *Repository) UpsertGoogleUser(ctx context.Context, user GoogleUser) (string, error) {
	return r.queries.UpsertGoogleUser(ctx, database.UpsertGoogleUserParams{
		GoogleSubject: user.Subject,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		AvatarUrl:     user.AvatarURL,
	})
}
