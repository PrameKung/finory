package auth

import (
	"context"

	"finory/backend/services/auth-service/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
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

func (r *Repository) GetUserByID(ctx context.Context, id string) (User, error) {
	var userID pgtype.UUID
	if err := userID.Scan(id); err != nil {
		return User{}, err
	}
	row, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	return User{ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, AvatarURL: row.AvatarUrl}, nil
}

func (r *Repository) CreateRefreshSession(ctx context.Context, userID string, tokenHash []byte) error {
	var id pgtype.UUID
	if err := id.Scan(userID); err != nil {
		return err
	}
	return r.queries.CreateRefreshSession(ctx, database.CreateRefreshSessionParams{
		UserID: id, TokenHash: tokenHash,
	})
}

func (r *Repository) RotateRefreshSession(ctx context.Context, oldHash, newHash []byte) (string, error) {
	return r.queries.RotateRefreshSession(ctx, database.RotateRefreshSessionParams{
		OldTokenHash: oldHash, NewTokenHash: newHash,
	})
}

func (r *Repository) DeleteRefreshSession(ctx context.Context, tokenHash []byte) error {
	return r.queries.DeleteRefreshSession(ctx, tokenHash)
}
