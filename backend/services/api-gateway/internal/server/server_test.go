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
