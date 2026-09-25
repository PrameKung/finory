package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "postgresql://example/auth_db")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8081" || cfg.DatabaseURL != "postgresql://example/auth_db" {
		t.Fatalf("unexpected auth configuration: %+v", cfg)
	}

	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing database URL to fail")
	}

	t.Setenv("DATABASE_URL", "postgresql://example/auth_db")
	t.Setenv("PORT", "65536")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid port to fail")
	}
}
