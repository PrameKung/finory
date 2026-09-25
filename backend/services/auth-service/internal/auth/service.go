package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidIdentity = errors.New("invalid Google identity")

type UserStore interface {
	UpsertGoogleUser(context.Context, GoogleUser) (string, error)
}

type Service struct {
	users        UserStore
	accessSecret []byte
}

func NewService(users UserStore, accessSecret []byte) *Service {
	return &Service{users: users, accessSecret: accessSecret}
}

func (s *Service) SignIn(ctx context.Context, user GoogleUser, emailVerified bool) (string, error) {
	if strings.TrimSpace(user.Subject) == "" || strings.TrimSpace(user.Email) == "" || !emailVerified {
		return "", ErrInvalidIdentity
	}
	id, err := s.users.UpsertGoogleUser(ctx, user)
	if err != nil {
		return "", err
	}
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   id,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	}).SignedString(s.accessSecret)
}
