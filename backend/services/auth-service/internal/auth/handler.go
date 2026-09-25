package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"
)

const flowCookieName = "finory_oauth"
const accessCookieName = "finory_access"
const refreshCookieName = "finory_refresh"

type TokenVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type Handler struct {
	oauth          oauth2.Config
	verifier       TokenVerifier
	service        *Service
	appRedirectURL string
	secureCookies  bool
}

func NewHandler(oauth oauth2.Config, verifier TokenVerifier, service *Service, appRedirectURL string) *Handler {
	return &Handler{
		oauth: oauth, verifier: verifier, service: service,
		appRedirectURL: appRedirectURL,
		secureCookies:  strings.HasPrefix(oauth.RedirectURL, "https://"),
	}
}

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/auth/google", h.authorize)
	e.GET("/auth/google/callback", h.callback)
	e.GET("/auth/me", h.me)
	e.POST("/auth/refresh", h.refresh)
	e.POST("/auth/logout", h.logout)
}

func (h *Handler) me(c *echo.Context) error {
	var rawToken string
	authorizations := c.Request().Header.Values(echo.HeaderAuthorization)
	switch len(authorizations) {
	case 0:
		cookie, err := c.Cookie(accessCookieName)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		rawToken = cookie.Value
	case 1:
		parts := strings.Fields(authorizations[0])
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		rawToken = parts[1]
	default:
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()
	user, err := h.service.CurrentUser(ctx, rawToken)
	if errors.Is(err, ErrInvalidAccessToken) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "user_lookup_failed"})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) logout(c *echo.Context) error {
	if cookie, err := c.Cookie(refreshCookieName); err == nil {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()
		if err := h.service.Logout(ctx, cookie.Value); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "logout_failed"})
		}
	}
	for _, cookie := range []struct{ name, path string }{
		{name: accessCookieName, path: "/api/v1"},
		{name: refreshCookieName, path: "/api/v1/auth"},
		{name: flowCookieName, path: "/api/v1/auth/google"},
	} {
		h.clearCookie(c, cookie.name, cookie.path)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) refresh(c *echo.Context) error {
	cookie, err := c.Cookie(refreshCookieName)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid_session"})
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()
	tokens, err := h.service.Renew(ctx, cookie.Value)
	if errors.Is(err, ErrInvalidSession) {
		h.clearCookie(c, refreshCookieName, "/api/v1/auth")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid_session"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "refresh_failed"})
	}
	h.setSessionCookies(c, tokens)
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) authorize(c *echo.Context) error {
	state, err := randomValue()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "oauth_unavailable"})
	}
	nonce, err := randomValue()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "oauth_unavailable"})
	}
	verifier := oauth2.GenerateVerifier()
	c.SetCookie(&http.Cookie{
		Name: flowCookieName, Value: state + "." + nonce + "." + verifier,
		Path: "/api/v1/auth/google", MaxAge: 600, HttpOnly: true,
		Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
	})
	url := h.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("nonce", nonce))
	return c.Redirect(http.StatusFound, url)
}

func (h *Handler) callback(c *echo.Context) error {
	h.clearCookie(c, flowCookieName, "/api/v1/auth/google")
	var flowCookie *http.Cookie
	for _, cookie := range c.Request().Cookies() {
		if cookie.Name == flowCookieName {
			if flowCookie != nil {
				return invalidCallback(c)
			}
			flowCookie = cookie
		}
	}
	if flowCookie == nil {
		return invalidCallback(c)
	}
	parts := strings.Split(flowCookie.Value, ".")
	query, err := url.ParseQuery(c.Request().URL.RawQuery)
	if err != nil {
		return invalidCallback(c)
	}
	states, codes := query["state"], query["code"]
	_, hasProviderError := query["error"]
	if len(parts) != 3 || len(states) != 1 || len(codes) != 1 || states[0] == "" || codes[0] == "" ||
		!validRandomValue(parts[0]) || !validRandomValue(parts[1]) || !validRandomValue(parts[2]) ||
		subtle.ConstantTimeCompare([]byte(states[0]), []byte(parts[0])) != 1 || hasProviderError {
		return invalidCallback(c)
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	token, err := h.oauth.Exchange(ctx, codes[0], oauth2.VerifierOption(parts[2]))
	if err != nil {
		return invalidCallback(c)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return invalidCallback(c)
	}
	idToken, err := h.verifier.Verify(ctx, rawIDToken)
	if err != nil || idToken == nil || subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(parts[1])) != 1 {
		return invalidCallback(c)
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return invalidCallback(c)
	}
	tokens, err := h.service.SignIn(ctx, GoogleUser{
		Subject: idToken.Subject, Email: claims.Email,
		DisplayName: claims.Name, AvatarURL: claims.Picture,
	}, claims.EmailVerified)
	if errors.Is(err, ErrInvalidIdentity) {
		return invalidCallback(c)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "sign_in_failed"})
	}
	h.setSessionCookies(c, tokens)
	return c.Redirect(http.StatusSeeOther, h.appRedirectURL)
}

func (h *Handler) setSessionCookies(c *echo.Context, tokens SessionTokens) {
	for _, cookie := range []struct {
		name, value, path string
		maxAge            int
	}{
		{name: accessCookieName, value: tokens.AccessToken, path: "/api/v1", maxAge: int(accessTokenTTL.Seconds())},
		{name: refreshCookieName, value: tokens.RefreshToken, path: "/api/v1/auth", maxAge: int(refreshTokenTTL.Seconds())},
	} {
		c.SetCookie(&http.Cookie{
			Name: cookie.name, Value: cookie.value, Path: cookie.path,
			MaxAge: cookie.maxAge, HttpOnly: true,
			Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
		})
	}
}

func (h *Handler) clearCookie(c *echo.Context, name, path string) {
	c.SetCookie(&http.Cookie{
		Name: name, Path: path, MaxAge: -1,
		Expires: time.Unix(0, 0), HttpOnly: true,
		Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
	})
}

func randomValue() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), nil
}

func validRandomValue(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == 32 && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func invalidCallback(c *echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_oauth_callback"})
}
