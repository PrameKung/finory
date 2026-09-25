package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// RequireAccessToken verifies access tokens before a request reaches a service.
func RequireAccessToken(secret []byte) echo.MiddlewareFunc {
	if len(secret) < 32 {
		panic("JWT access secret must be at least 32 bytes")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			unauthorized := func() error {
				c.Response().Header().Set("WWW-Authenticate", "Bearer")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			authorization := c.Request().Header.Values(echo.HeaderAuthorization)
			if len(authorization) != 1 {
				return unauthorized()
			}
			parts := strings.Fields(authorization[0])
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return unauthorized()
			}

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(*jwt.Token) (any, error) {
				return secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithStrictDecoding())
			if err != nil || !token.Valid {
				return unauthorized()
			}
			userID, err := uuid.Parse(claims.Subject)
			if err != nil || userID == uuid.Nil || userID.String() != claims.Subject {
				return unauthorized()
			}

			return next(c)
		}
	}
}
