package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                string
	AuthServiceURL      *url.URL
	LedgerServiceURL    *url.URL
	AnalyticsServiceURL *url.URL
	CORSAllowedOrigins  []string
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := validatePort(port); err != nil {
		return Config{}, err
	}

	authURL, err := serviceURL("AUTH_SERVICE_URL", "http://localhost:8081")
	if err != nil {
		return Config{}, err
	}
	ledgerURL, err := serviceURL("LEDGER_SERVICE_URL", "http://localhost:8082")
	if err != nil {
		return Config{}, err
	}
	analyticsURL, err := serviceURL("ANALYTICS_SERVICE_URL", "http://localhost:8083")
	if err != nil {
		return Config{}, err
	}
	allowedOrigins, err := corsAllowedOrigins()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:                port,
		AuthServiceURL:      authURL,
		LedgerServiceURL:    ledgerURL,
		AnalyticsServiceURL: analyticsURL,
		CORSAllowedOrigins:  allowedOrigins,
	}, nil
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

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("PORT must be a number from 1 to 65535")
	}
	return nil
}

func serviceURL(name, fallback string) (*url.URL, error) {
	raw := os.Getenv(name)
	if raw == "" {
		raw = fallback
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return nil, fmt.Errorf("%s must be an HTTP(S) service base URL", name)
	}
	return parsed, nil
}
