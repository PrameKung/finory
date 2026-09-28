package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("LEDGER_SERVICE_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8083" || cfg.LedgerServiceURL.Host != "localhost:8082" {
		t.Fatalf("unexpected analytics defaults: %+v", cfg)
	}

	t.Setenv("LEDGER_SERVICE_URL", "http://ledger-service:8082")
	cfg, err = Load()
	if err != nil || cfg.LedgerServiceURL.Host != "ledger-service:8082" {
		t.Fatalf("analytics service URL was not loaded: config=%+v, error=%v", cfg, err)
	}

	t.Setenv("LEDGER_SERVICE_URL", "ledger-service:10000")
	cfg, err = Load()
	if err != nil || cfg.LedgerServiceURL.String() != "http://ledger-service:10000" {
		t.Fatalf("Render host and port were not normalized: config=%+v, error=%v", cfg, err)
	}

	t.Setenv("LEDGER_SERVICE_URL", "ftp://ledger-service:8082")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid service URL to fail")
	}

	t.Setenv("LEDGER_SERVICE_URL", "http://localhost:8082")
	t.Setenv("PORT", "65536")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid port to fail")
	}
}
