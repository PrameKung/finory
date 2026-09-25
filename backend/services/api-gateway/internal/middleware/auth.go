package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// RequireAccessToken verifies access tokens before a request reaches a service.
func RequireAccessToken(secret []byte, allowedOrigins []string) echo.MiddlewareFunc {
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
			var rawToken string
			if len(authorization) == 0 {
				cookie, err := c.Cookie("finory_access")
				if err != nil {
					return unauthorized()
				}
				if c.Request().Method != http.MethodGet && c.Request().Method != http.MethodHead && c.Request().Method != http.MethodOptions {
					origin := c.Request().Header.Get("Origin")
					allowed := false
					for _, candidate := range allowedOrigins {
						if origin == candidate {
							allowed = true
							break
						}
					}
					if !allowed {
						return unauthorized()
					}
				}
				rawToken = cookie.Value
			} else if len(authorization) == 1 {
				parts := strings.Fields(authorization[0])
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
					return unauthorized()
				}
				rawToken = parts[1]
			} else {
				return unauthorized()
			}

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(rawToken, claims, func(*jwt.Token) (any, error) {
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
