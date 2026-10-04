package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	AuthDatabaseURL    string
	LedgerDatabaseURL  string
	GoogleClientID     string
	GoogleSecret       string
	GoogleRedirectURL  string
	AppRedirectURL     string
	AppLoginURL        string
	CORSAllowedOrigins []string
	JWTAccessSecret    []byte
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number from 1 to 65535")
	}
	authDatabaseURL := strings.TrimSpace(os.Getenv("AUTH_DATABASE_URL"))
	ledgerDatabaseURL := strings.TrimSpace(os.Getenv("LEDGER_DATABASE_URL"))
	if authDatabaseURL == "" || ledgerDatabaseURL == "" {
		return Config{}, fmt.Errorf("AUTH_DATABASE_URL and LEDGER_DATABASE_URL are required")
	}
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return Config{}, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required")
	}
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if err := validateURL(redirectURL, "/api/v1/auth/google/callback"); err != nil {
		return Config{}, fmt.Errorf("GOOGLE_REDIRECT_URL: %w", err)
	}
	appRedirectURL := os.Getenv("APP_REDIRECT_URL")
	if err := validateURL(appRedirectURL, ""); err != nil {
		return Config{}, fmt.Errorf("APP_REDIRECT_URL: %w", err)
	}
	appURL, _ := url.Parse(appRedirectURL)
	appURL.Path = "/login"
	appURL.RawPath = ""
	secret := os.Getenv("JWT_ACCESS_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 bytes")
	}
	origins, err := corsAllowedOrigins()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Port: port, AuthDatabaseURL: authDatabaseURL, LedgerDatabaseURL: ledgerDatabaseURL,
		GoogleClientID: clientID,
		GoogleSecret:   clientSecret, GoogleRedirectURL: redirectURL,
		AppRedirectURL: appRedirectURL, AppLoginURL: appURL.String(),
		CORSAllowedOrigins: origins, JWTAccessSecret: []byte(secret),
	}, nil
}

func validateURL(raw, requiredPath string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		(u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) ||
		(requiredPath != "" && u.Path != requiredPath) {
		return fmt.Errorf("must be an HTTPS URL (HTTP allowed for localhost) with the expected path and no query or fragment")
	}
	return nil
}

func corsAllowedOrigins() ([]string, error) {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:3000,http://127.0.0.1:3000"
	}
	origins := strings.Split(raw, ",")
	for i, origin := range origins {
		origin = strings.TrimSpace(origin)
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS must contain only HTTP(S) origins")
		}
		origins[i] = origin
	}
	return origins, nil
}
