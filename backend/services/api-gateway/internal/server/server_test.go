package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"finory/backend/services/api-gateway/internal/config"
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
	gateway := New(config.Config{AuthServiceURL: authURL})
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
	gateway := New(config.Config{AuthServiceURL: authURL})
	response := httptest.NewRecorder()
	gateway.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/auth", nil))
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
	gateway := New(config.Config{AuthServiceURL: authURL})
	response := httptest.NewRecorder()
	gateway.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/auth/health", nil))
	if response.Code != http.StatusOK || response.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected auth health response: status=%d, body=%q", response.Code, response.Body.String())
	}
}

func TestGatewayHealth(t *testing.T) {
	response := httptest.NewRecorder()
	New(config.Config{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
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
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch || r.URL.Path != tc.want || r.URL.RawQuery != "page=2" {
					t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.String())
				}
				if r.Header.Get("Authorization") != "Bearer example" {
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
			gateway := New(config.Config{LedgerServiceURL: ledgerURL})
			request := httptest.NewRequest(http.MethodPatch, tc.path+"?page=2", strings.NewReader(`{"name":"updated"}`))
			request.Header.Set("Authorization", "Bearer example")
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
	New(config.Config{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/transactions-extra", nil))
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
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != tc.want || r.URL.RawQuery != "month=09" {
					t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.String())
				}
				if r.Header.Get("Authorization") != "Bearer example" {
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
			gateway := New(config.Config{AnalyticsServiceURL: analyticsURL})
			request := httptest.NewRequest(http.MethodGet, tc.path+"?month=09", nil)
			request.Header.Set("Authorization", "Bearer example")
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
	New(config.Config{AnalyticsServiceURL: analyticsURL}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/analytics/health", nil))
	if response.Code != http.StatusOK || response.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected analytics health response: status=%d, body=%q", response.Code, response.Body.String())
	}
}

func TestAnalyticsRouteBoundary(t *testing.T) {
	response := httptest.NewRecorder()
	New(config.Config{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/analytics-extra", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("unrelated route should not be proxied: %d", response.Code)
	}
}
