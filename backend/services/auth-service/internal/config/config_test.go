package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "postgresql://example/auth_db")
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	t.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback")
	t.Setenv("APP_REDIRECT_URL", "http://localhost:3000/dashboard")
	t.Setenv("JWT_ACCESS_SECRET", "test-secret-with-at-least-32-bytes")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8081" || cfg.DatabaseURL != "postgresql://example/auth_db" ||
		cfg.AppRedirectURL != "http://localhost:3000/dashboard" || cfg.AppLoginURL != "http://localhost:3000/login" {
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

	t.Setenv("PORT", "8081")
	t.Setenv("GOOGLE_REDIRECT_URL", "https://evil.example.com/callback")
	if _, err := Load(); err == nil {
		t.Fatal("expected unexpected Google callback path to fail")
	}
}
