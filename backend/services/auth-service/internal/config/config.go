package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	DatabaseURL       string
	GoogleClientID    string
	GoogleSecret      string
	GoogleRedirectURL string
	AppRedirectURL    string
	JWTAccessSecret   []byte
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number from 1 to 65535")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
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
	secret := os.Getenv("JWT_ACCESS_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 bytes")
	}
	return Config{
		Port: port, DatabaseURL: databaseURL, GoogleClientID: clientID,
		GoogleSecret: clientSecret, GoogleRedirectURL: redirectURL,
		AppRedirectURL: appRedirectURL, JWTAccessSecret: []byte(secret),
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
