package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("AUTH_SERVICE_URL", "")
	t.Setenv("LEDGER_SERVICE_URL", "")
	t.Setenv("ANALYTICS_SERVICE_URL", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("JWT_ACCESS_SECRET", "test-secret-with-at-least-32-bytes-long")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.AuthServiceURL.Host != "localhost:8081" || cfg.LedgerServiceURL.Host != "localhost:8082" || cfg.AnalyticsServiceURL.Host != "localhost:8083" || len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "http://localhost:3000" || cfg.CORSAllowedOrigins[1] != "http://127.0.0.1:3000" || string(cfg.JWTAccessSecret) != "test-secret-with-at-least-32-bytes-long" {
		t.Fatalf("unexpected gateway defaults: %+v", cfg)
	}

	t.Setenv("AUTH_SERVICE_URL", "http://auth-service:8081")
	cfg, err = Load()
	if err != nil || cfg.AuthServiceURL.Host != "auth-service:8081" {
		t.Fatalf("gateway service URL was not loaded: config=%+v, error=%v", cfg, err)
	}

	t.Setenv("AUTH_SERVICE_URL", "auth-service:10000")
	cfg, err = Load()
	if err != nil || cfg.AuthServiceURL.String() != "http://auth-service:10000" {
		t.Fatalf("Render host and port were not normalized: config=%+v, error=%v", cfg, err)
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

func TestLoadRequiresJWTAccessSecret(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("AUTH_SERVICE_URL", "")
	t.Setenv("LEDGER_SERVICE_URL", "")
	t.Setenv("ANALYTICS_SERVICE_URL", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	for _, secret := range []string{"", "short-secret"} {
		t.Setenv("JWT_ACCESS_SECRET", secret)
		if _, err := Load(); err == nil {
			t.Fatalf("expected JWT secret %q to fail", secret)
		}
	}
}

func TestCORSAllowedOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	origins, err := corsAllowedOrigins()
	if err != nil || len(origins) != 2 || origins[0] != "https://app.example.com" || origins[1] != "https://admin.example.com" {
		t.Fatalf("unexpected CORS origins: %v, error: %v", origins, err)
	}

	for _, invalid := range []string{"*", "https://app.example.com/path", "https://app.example.com,", "ftp://app.example.com"} {
		t.Setenv("CORS_ALLOWED_ORIGINS", invalid)
		if _, err := corsAllowedOrigins(); err == nil {
			t.Errorf("expected invalid CORS origin %q to fail", invalid)
		}
	}
}
