package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("AUTH_SERVICE_URL", "")
	t.Setenv("LEDGER_SERVICE_URL", "")
	t.Setenv("ANALYTICS_SERVICE_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.AuthServiceURL.Host != "localhost:8081" || cfg.LedgerServiceURL.Host != "localhost:8082" || cfg.AnalyticsServiceURL.Host != "localhost:8083" {
		t.Fatalf("unexpected gateway defaults: %+v", cfg)
	}

	t.Setenv("AUTH_SERVICE_URL", "http://auth-service:8081")
	cfg, err = Load()
	if err != nil || cfg.AuthServiceURL.Host != "auth-service:8081" {
		t.Fatalf("gateway service URL was not loaded: config=%+v, error=%v", cfg, err)
	}

	t.Setenv("AUTH_SERVICE_URL", "not-a-url")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid service URL to fail")
	}

	t.Setenv("AUTH_SERVICE_URL", "http://localhost:8081")
	t.Setenv("PORT", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid port to fail")
	}
}
