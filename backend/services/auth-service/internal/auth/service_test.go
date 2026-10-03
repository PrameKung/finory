package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthenticationServiceFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	users := &memoryUsers{}
	service := NewService(users, []byte(testSecret))
	googleUser := GoogleUser{
		Subject:     "stable-google-subject",
		Email:       "user@example.com",
		DisplayName: "Example User",
		AvatarURL:   "https://example.com/avatar.png",
	}

	session, err := service.SignIn(ctx, googleUser, true)
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	if session.AccessToken == "" || session.RefreshToken == "" || len(users.sessions) != 1 {
		t.Fatalf("sign in did not create a complete session: tokens=%+v, sessions=%d", session, len(users.sessions))
	}

	user, err := service.CurrentUser(ctx, session.AccessToken)
	if err != nil {
		t.Fatalf("get current user: %v", err)
	}
	if user.ID != testUserID || user.Email != googleUser.Email || user.DisplayName != googleUser.DisplayName || user.AvatarURL != googleUser.AvatarURL {
		t.Fatalf("current user = %+v", user)
	}

	renewed, err := service.Renew(ctx, session.RefreshToken)
	if err != nil {
		t.Fatalf("renew session: %v", err)
	}
	if renewed.AccessToken == "" || renewed.RefreshToken == "" || renewed.RefreshToken == session.RefreshToken || len(users.sessions) != 1 {
		t.Fatalf("refresh token was not rotated: old=%+v, new=%+v, sessions=%d", session, renewed, len(users.sessions))
	}
	if _, err := service.Renew(ctx, session.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("replayed refresh token error = %v, want %v", err, ErrInvalidSession)
	}
	if _, err := service.CurrentUser(ctx, renewed.AccessToken); err != nil {
		t.Fatalf("renewed access token is not usable: %v", err)
	}

	if err := service.Logout(ctx, renewed.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if len(users.sessions) != 0 {
		t.Fatalf("logout left %d refresh sessions", len(users.sessions))
	}
	if _, err := service.Renew(ctx, renewed.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("revoked refresh token error = %v, want %v", err, ErrInvalidSession)
	}
}

func TestSignInRejectsInvalidGoogleIdentity(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		user          GoogleUser
		emailVerified bool
	}{
		{
			name:          "missing subject",
			user:          GoogleUser{Email: "user@example.com"},
			emailVerified: true,
		},
		{
			name:          "blank subject",
			user:          GoogleUser{Subject: "  ", Email: "user@example.com"},
			emailVerified: true,
		},
		{
			name:          "missing email",
			user:          GoogleUser{Subject: "google-subject"},
			emailVerified: true,
		},
		{
			name:          "unverified email",
			user:          GoogleUser{Subject: "google-subject", Email: "user@example.com"},
			emailVerified: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			users := &memoryUsers{}
			_, err := NewService(users, []byte(testSecret)).SignIn(context.Background(), tc.user, tc.emailVerified)
			if !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("sign in error = %v, want %v", err, ErrInvalidIdentity)
			}
			if users.upserts != 0 || len(users.sessions) != 0 {
				t.Fatalf("invalid identity changed auth data: upserts=%d, sessions=%d", users.upserts, len(users.sessions))
			}
		})
	}
}

func TestAuthenticationServiceRejectsInvalidCredentials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	users := &memoryUsers{}
	service := NewService(users, []byte(testSecret))

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   testUserID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	wrongSignature, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   testUserID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}).SignedString([]byte("another-test-secret-with-at-least-32-bytes"))
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name  string
		token string
	}{
		{name: "malformed access token", token: "not-a-jwt"},
		{name: "expired access token", token: expiredToken},
		{name: "wrong access-token signature", token: wrongSignature},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := service.CurrentUser(ctx, tc.token); !errors.Is(err, ErrInvalidAccessToken) {
				t.Fatalf("current user error = %v, want %v", err, ErrInvalidAccessToken)
			}
		})
	}

	validButUnknownRefreshToken, _, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"not-a-refresh-token", validButUnknownRefreshToken} {
		if _, err := service.Renew(ctx, token); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("renew with %q error = %v, want %v", token, err, ErrInvalidSession)
		}
	}
}
