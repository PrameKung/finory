package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	Port                string
	AuthServiceURL      *url.URL
	LedgerServiceURL    *url.URL
	AnalyticsServiceURL *url.URL
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

	return Config{
		Port:                port,
		AuthServiceURL:      authURL,
		LedgerServiceURL:    ledgerURL,
		AnalyticsServiceURL: analyticsURL,
	}, nil
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
