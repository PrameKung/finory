package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"
)

const flowCookieName = "finory_oauth"
const accessCookieName = "finory_access"

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
	e.POST("/auth/logout", h.logout)
}

func (h *Handler) logout(c *echo.Context) error {
	for _, cookie := range []struct{ name, path string }{
		{name: accessCookieName, path: "/api/v1"},
		{name: flowCookieName, path: "/api/v1/auth/google"},
	} {
		c.SetCookie(&http.Cookie{
			Name: cookie.name, Path: cookie.path, MaxAge: -1,
			Expires: time.Unix(0, 0), HttpOnly: true,
			Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
		})
	}
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
	c.SetCookie(&http.Cookie{
		Name: flowCookieName, Path: "/api/v1/auth/google", MaxAge: -1,
		HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
	})
	cookie, err := c.Cookie(flowCookieName)
	if err != nil {
		return invalidCallback(c)
	}
	parts := strings.Split(cookie.Value, ".")
	query := c.Request().URL.Query()
	states, codes := query["state"], query["code"]
	_, hasProviderError := query["error"]
	if len(parts) != 3 || len(states) != 1 || len(codes) != 1 || states[0] == "" || codes[0] == "" ||
		len(parts[0]) < 40 || len(parts[1]) < 40 || len(parts[2]) < 40 ||
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
	accessToken, err := h.service.SignIn(ctx, GoogleUser{
		Subject: idToken.Subject, Email: claims.Email,
		DisplayName: claims.Name, AvatarURL: claims.Picture,
	}, claims.EmailVerified)
	if errors.Is(err, ErrInvalidIdentity) {
		return invalidCallback(c)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "sign_in_failed"})
	}
	c.SetCookie(&http.Cookie{
		Name: accessCookieName, Value: accessToken, Path: "/api/v1",
		MaxAge: 3600, HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
	})
	return c.Redirect(http.StatusSeeOther, h.appRedirectURL)
}

func randomValue() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), nil
}

func invalidCallback(c *echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_oauth_callback"})
}
