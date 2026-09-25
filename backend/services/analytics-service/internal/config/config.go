package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	Port             string
	LedgerServiceURL *url.URL
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number from 1 to 65535")
	}

	rawURL := os.Getenv("LEDGER_SERVICE_URL")
	if rawURL == "" {
		rawURL = "http://localhost:8082"
	}
	ledgerURL, err := url.Parse(rawURL)
	if err != nil || (ledgerURL.Scheme != "http" && ledgerURL.Scheme != "https") || ledgerURL.Host == "" || ledgerURL.User != nil || ledgerURL.RawQuery != "" || ledgerURL.Fragment != "" || (ledgerURL.Path != "" && ledgerURL.Path != "/") {
		return Config{}, fmt.Errorf("LEDGER_SERVICE_URL must be an HTTP(S) service base URL")
	}
	return Config{Port: port, LedgerServiceURL: ledgerURL}, nil
}
