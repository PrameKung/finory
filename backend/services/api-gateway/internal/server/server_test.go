package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"finory/backend/services/api-gateway/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/auth/login" || r.URL.RawQuery != "next=dashboard" {
			t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.String())
		}
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != `{"email":"user@example.com"}` {
			t.Errorf("unexpected upstream body: %q, error: %v", body, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":"example"}`))
	}))
	defer upstream.Close()

	authURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := newTestGateway(config.Config{AuthServiceURL: authURL})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login?next=dashboard", strings.NewReader(`{"email":"user@example.com"}`))
	response := httptest.NewRecorder()
	gateway.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || response.Body.String() != `{"token":"example"}` || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected gateway response: status=%d, body=%q, headers=%v", response.Code, response.Body.String(), response.Header())
	}
}

func TestAuthRouteBoundary(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth" {
			t.Errorf("unexpected upstream path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	authURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := newTestGateway(config.Config{AuthServiceURL: authURL})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth", nil)
	request.Header.Set("Authorization", "Bearer "+testAccessToken(t))
	gateway.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("unexpected auth root status: %d", response.Code)
	}

	response = httptest.NewRecorder()
	gateway.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/authentication", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("unrelated route should not be proxied: %d", response.Code)
	}
}

func TestAuthServiceHealthRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/health" {
			t.Errorf("unexpected upstream health request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer upstream.Close()

	authURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := newTestGateway(config.Config{AuthServiceURL: authURL})
	response := httptest.NewRecorder()
	gateway.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/auth/health", nil))
	if response.Code != http.StatusOK || response.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected auth health response: status=%d, body=%q", response.Code, response.Body.String())
	}
}

func TestGatewayHealth(t *testing.T) {
	response := httptest.NewRecorder()
	newTestGateway(config.Config{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK || response.Body.String() != "{\"status\":\"ok\"}\n" || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected gateway health response: status=%d, body=%q, headers=%v", response.Code, response.Body.String(), response.Header())
	}
}

func TestLedgerRoutes(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "transactions root", path: "/api/v1/transactions", want: "/transactions"},
		{name: "transactions trailing slash", path: "/api/v1/transactions/", want: "/transactions/"},
		{name: "transaction ID", path: "/api/v1/transactions/123", want: "/transactions/123"},
		{name: "categories root", path: "/api/v1/categories", want: "/categories"},
		{name: "category ID", path: "/api/v1/categories/123", want: "/categories/123"},
		{name: "wallets root", path: "/api/v1/wallets", want: "/wallets"},
		{name: "wallet ID", path: "/api/v1/wallets/123", want: "/wallets/123"},
		{name: "budgets root", path: "/api/v1/budgets", want: "/budgets"},
		{name: "budget ID", path: "/api/v1/budgets/123", want: "/budgets/123"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token := testAccessToken(t)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch || r.URL.Path != tc.want || r.URL.RawQuery != "page=2" {
					t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.String())
				}
				if r.Header.Get("Authorization") != "Bearer "+token {
					t.Errorf("authorization header was not forwarded: %q", r.Header.Get("Authorization"))
				}
				body, err := io.ReadAll(r.Body)
				if err != nil || string(body) != `{"name":"updated"}` {
					t.Errorf("unexpected upstream body: %q, error: %v", body, err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"saved":true}`))
			}))
			defer upstream.Close()

			ledgerURL, err := url.Parse(upstream.URL)
			if err != nil {
				t.Fatal(err)
			}
			gateway := newTestGateway(config.Config{LedgerServiceURL: ledgerURL})
			request := httptest.NewRequest(http.MethodPatch, tc.path+"?page=2", strings.NewReader(`{"name":"updated"}`))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			gateway.ServeHTTP(response, request)

			if response.Code != http.StatusAccepted || response.Body.String() != `{"saved":true}` || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("unexpected gateway response: status=%d, body=%q, headers=%v", response.Code, response.Body.String(), response.Header())
			}
		})
	}
}

func TestLedgerRouteBoundary(t *testing.T) {
	response := httptest.NewRecorder()
	newTestGateway(config.Config{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/transactions-extra", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("unrelated route should not be proxied: %d", response.Code)
	}
}

func TestAnalyticsRoutes(t *testing.T) {
	for _, tc := range []struct {
		path string
		want string
	}{
		{path: "/api/v1/analytics", want: "/analytics"},
		{path: "/api/v1/analytics/", want: "/analytics/"},
		{path: "/api/v1/analytics/summary", want: "/analytics/summary"},
		{path: "/api/v1/analytics/monthly/2026", want: "/analytics/monthly/2026"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			token := testAccessToken(t)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != tc.want || r.URL.RawQuery != "month=09" {
					t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.String())
				}
				if r.Header.Get("Authorization") != "Bearer "+token {
					t.Errorf("authorization header was not forwarded: %q", r.Header.Get("Authorization"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"total":42}`))
			}))
			defer upstream.Close()

			analyticsURL, err := url.Parse(upstream.URL)
			if err != nil {
				t.Fatal(err)
			}
			gateway := newTestGateway(config.Config{AnalyticsServiceURL: analyticsURL})
			request := httptest.NewRequest(http.MethodGet, tc.path+"?month=09", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			gateway.ServeHTTP(response, request)

			if response.Code != http.StatusAccepted || response.Body.String() != `{"total":42}` || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("unexpected gateway response: status=%d, body=%q, headers=%v", response.Code, response.Body.String(), response.Header())
			}
		})
	}
}

func TestAnalyticsServiceHealthRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/health" {
			t.Errorf("unexpected upstream health request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer upstream.Close()

	analyticsURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	newTestGateway(config.Config{AnalyticsServiceURL: analyticsURL}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/analytics/health", nil))
	if response.Code != http.StatusOK || response.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected analytics health response: status=%d, body=%q", response.Code, response.Body.String())
	}
}

func TestAnalyticsRouteBoundary(t *testing.T) {
	response := httptest.NewRecorder()
	newTestGateway(config.Config{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/analytics-extra", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("unrelated route should not be proxied: %d", response.Code)
	}
}

func newTestGateway(cfg config.Config) http.Handler {
	cfg.CORSAllowedOrigins = []string{"http://localhost:3000"}
	cfg.JWTAccessSecret = []byte(testJWTAccessSecret)
	return New(cfg)
}

const testJWTAccessSecret = "test-access-secret-with-at-least-32-bytes"

func testAccessToken(t *testing.T) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "4c3b9d7e-91ef-4b2b-9e28-2408153715d9",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(testJWTAccessSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestCORSPreflight(t *testing.T) {
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/transactions", "/api/v1/analytics/summary"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodOptions, path, nil)
			request.Header.Set("Origin", "http://localhost:3000")
			request.Header.Set("Access-Control-Request-Method", http.MethodPost)
			request.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
			response := httptest.NewRecorder()
			newTestGateway(config.Config{}).ServeHTTP(response, request)

			if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
				t.Fatalf("unexpected preflight response: status=%d, headers=%v", response.Code, response.Header())
			}
			if !strings.Contains(response.Header().Get("Access-Control-Allow-Methods"), http.MethodPost) || !strings.Contains(response.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
				t.Fatalf("preflight methods or headers missing: %v", response.Header())
			}
		})
	}
}

func TestCORSAllowedAndDisallowedOrigins(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer upstream.Close()

	analyticsURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := newTestGateway(config.Config{AnalyticsServiceURL: analyticsURL})
	token := testAccessToken(t)
	for _, tc := range []struct {
		origin      string
		wantAllowed string
	}{
		{origin: "http://localhost:3000", wantAllowed: "http://localhost:3000"},
		{origin: "https://other.example.com"},
	} {
		t.Run(tc.origin, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/summary", nil)
			request.Header.Set("Origin", tc.origin)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			gateway.ServeHTTP(response, request)
			if response.Code != http.StatusAccepted || response.Header().Get("Access-Control-Allow-Origin") != tc.wantAllowed {
				t.Fatalf("unexpected CORS response: status=%d, headers=%v", response.Code, response.Header())
			}
		})
	}
}

func TestRequestIDPropagatesToServices(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "upstream-id")
		_, _ = w.Write([]byte(r.Header.Get("X-Request-ID")))
	}))
	defer upstream.Close()

	serviceURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := newTestGateway(config.Config{
		AuthServiceURL:      serviceURL,
		LedgerServiceURL:    serviceURL,
		AnalyticsServiceURL: serviceURL,
	})
	token := testAccessToken(t)

	for _, path := range []string{
		"/api/v1/auth/login",
		"/api/v1/auth/health",
		"/api/v1/transactions",
		"/api/v1/analytics/summary",
		"/api/v1/analytics/health",
	} {
		for _, suppliedID := range []string{"", "client-id-123"} {
			t.Run(path+"/"+suppliedID, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, path, nil)
				request.Header.Set("Authorization", "Bearer "+token)
				if suppliedID != "" {
					request.Header.Set("X-Request-ID", suppliedID)
				}
				response := httptest.NewRecorder()
				gateway.ServeHTTP(response, request)

				responseIDs := response.Result().Header.Values("X-Request-ID")
				if response.Code != http.StatusOK || len(responseIDs) != 1 || responseIDs[0] == "" || response.Body.String() != responseIDs[0] {
					t.Fatalf("request ID was not propagated: status=%d, response IDs=%v, upstream ID=%q", response.Code, responseIDs, response.Body.String())
				}
				if suppliedID != "" && responseIDs[0] != suppliedID {
					t.Fatalf("provided request ID changed: got %q, want %q", responseIDs[0], suppliedID)
				}
			})
		}
	}
}

func TestRequestIDOnLocalAndPreflightResponses(t *testing.T) {
	gateway := newTestGateway(config.Config{})
	for _, method := range []string{http.MethodGet, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			request := httptest.NewRequest(method, "/health", nil)
			if method == http.MethodOptions {
				request.Header.Set("Origin", "http://localhost:3000")
				request.Header.Set("Access-Control-Request-Method", http.MethodGet)
			}
			response := httptest.NewRecorder()
			gateway.ServeHTTP(response, request)
			if response.Header().Get("X-Request-ID") == "" {
				t.Fatal("response is missing X-Request-ID")
			}
		})
	}
}

func TestProtectedRoutesRejectMissingAccessToken(t *testing.T) {
	gateway := newTestGateway(config.Config{})
	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/auth/me"},
		{method: http.MethodPost, path: "/api/v1/auth/logout"},
		{method: http.MethodGet, path: "/api/v1/transactions"},
		{method: http.MethodGet, path: "/api/v1/transactions/123"},
		{method: http.MethodGet, path: "/api/v1/categories"},
		{method: http.MethodGet, path: "/api/v1/wallets"},
		{method: http.MethodGet, path: "/api/v1/budgets"},
		{method: http.MethodGet, path: "/api/v1/analytics/summary"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			gateway.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
			if response.Code != http.StatusUnauthorized || response.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("protected route accepted missing token: status=%d, headers=%v", response.Code, response.Header())
			}
		})
	}
}

func TestProtectedRouteRejectsInvalidAccessTokens(t *testing.T) {
	sign := func(t *testing.T, method jwt.SigningMethod, claims jwt.RegisteredClaims, secret string) string {
		t.Helper()
		token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	validClaims := jwt.RegisteredClaims{
		Subject:   "4c3b9d7e-91ef-4b2b-9e28-2408153715d9",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	for _, tc := range []struct {
		name  string
		token string
	}{
		{name: "malformed", token: "not-a-jwt"},
		{name: "wrong signature", token: sign(t, jwt.SigningMethodHS256, validClaims, "other-secret-with-at-least-32-bytes")},
		{name: "wrong algorithm", token: sign(t, jwt.SigningMethodHS384, validClaims, testJWTAccessSecret)},
		{name: "expired", token: sign(t, jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: validClaims.Subject, ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))}, testJWTAccessSecret)},
		{name: "missing expiry", token: sign(t, jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: validClaims.Subject}, testJWTAccessSecret)},
		{name: "missing subject", token: sign(t, jwt.SigningMethodHS256, jwt.RegisteredClaims{ExpiresAt: validClaims.ExpiresAt}, testJWTAccessSecret)},
		{name: "invalid subject", token: sign(t, jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "not-a-uuid", ExpiresAt: validClaims.ExpiresAt}, testJWTAccessSecret)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
			request.Header.Set("Authorization", "Bearer "+tc.token)
			response := httptest.NewRecorder()
			newTestGateway(config.Config{}).ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("invalid token was accepted: status=%d, body=%q", response.Code, response.Body.String())
			}
		})
	}
}

func TestPublicAuthRoutesDoNotRequireAccessToken(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()
	authURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := newTestGateway(config.Config{AuthServiceURL: authURL})
	for _, path := range []string{"/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/refresh"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			gateway.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, nil))
			if response.Code != http.StatusNoContent {
				t.Fatalf("public auth route rejected request: status=%d, body=%q", response.Code, response.Body.String())
			}
		})
	}
}
