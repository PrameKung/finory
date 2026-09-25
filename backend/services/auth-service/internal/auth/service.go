package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

var ErrInvalidIdentity = errors.New("invalid Google identity")
var ErrInvalidSession = errors.New("invalid refresh session")

const accessTokenTTL = 15 * time.Minute
const refreshTokenTTL = 30 * 24 * time.Hour

type Store interface {
	UpsertGoogleUser(context.Context, GoogleUser) (string, error)
	CreateRefreshSession(context.Context, string, []byte) error
	RotateRefreshSession(context.Context, []byte, []byte) (string, error)
	DeleteRefreshSession(context.Context, []byte) error
}

type SessionTokens struct {
	AccessToken  string
	RefreshToken string
}

type Service struct {
	store        Store
	accessSecret []byte
}

func NewService(store Store, accessSecret []byte) *Service {
	return &Service{store: store, accessSecret: accessSecret}
}

func (s *Service) SignIn(ctx context.Context, user GoogleUser, emailVerified bool) (SessionTokens, error) {
	if strings.TrimSpace(user.Subject) == "" || strings.TrimSpace(user.Email) == "" || !emailVerified {
		return SessionTokens{}, ErrInvalidIdentity
	}
	id, err := s.store.UpsertGoogleUser(ctx, user)
	if err != nil {
		return SessionTokens{}, err
	}
	accessToken, err := s.issueAccessToken(id)
	if err != nil {
		return SessionTokens{}, err
	}
	refreshToken, hash, err := newRefreshToken()
	if err != nil {
		return SessionTokens{}, err
	}
	if err := s.store.CreateRefreshSession(ctx, id, hash[:]); err != nil {
		return SessionTokens{}, err
	}
	return SessionTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) Renew(ctx context.Context, oldRefreshToken string) (SessionTokens, error) {
	oldHash, err := refreshTokenHash(oldRefreshToken)
	if err != nil {
		return SessionTokens{}, ErrInvalidSession
	}
	newToken, newHash, err := newRefreshToken()
	if err != nil {
		return SessionTokens{}, err
	}
	id, err := s.store.RotateRefreshSession(ctx, oldHash[:], newHash[:])
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionTokens{}, ErrInvalidSession
	}
	if err != nil {
		return SessionTokens{}, err
	}
	accessToken, err := s.issueAccessToken(id)
	if err != nil {
		return SessionTokens{}, err
	}
	return SessionTokens{AccessToken: accessToken, RefreshToken: newToken}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash, err := refreshTokenHash(refreshToken)
	if err != nil {
		return nil
	}
	return s.store.DeleteRefreshSession(ctx, hash[:])
}

func (s *Service) issueAccessToken(id string) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   id,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
	}).SignedString(s.accessSecret)
}

func newRefreshToken() (string, [32]byte, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", [32]byte{}, err
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), sha256.Sum256(value[:]), nil
}

func refreshTokenHash(token string) ([32]byte, error) {
	value, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(value) != 32 {
		return [32]byte{}, ErrInvalidSession
	}
	return sha256.Sum256(value), nil
}
