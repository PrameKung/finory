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

const integrationUserID = "4c3b9d7e-91ef-4b2b-9e28-2408153715d9"

func TestGatewayIntegrationRoutesVerifiedIdentityToServices(t *testing.T) {
	t.Parallel()

	authServer, authURL := newGatewayUpstream(t, gatewayUpstreamExpectation{
		name: "auth", method: http.MethodGet, path: "/auth/me",
	})
	defer authServer.Close()
	ledgerServer, ledgerURL := newGatewayUpstream(t, gatewayUpstreamExpectation{
		name: "ledger", method: http.MethodPatch, path: "/transactions/transaction-id",
		query: "month=2026-09", body: `{"amount":"25.50"}`,
	})
	defer ledgerServer.Close()
	analyticsServer, analyticsURL := newGatewayUpstream(t, gatewayUpstreamExpectation{
		name: "analytics", method: http.MethodGet, path: "/analytics/summary",
		query: "month=2026-09",
	})
	defer analyticsServer.Close()

	gateway := newTestGateway(config.Config{
		AuthServiceURL: authURL, LedgerServiceURL: ledgerURL, AnalyticsServiceURL: analyticsURL,
	})
	token := testAccessToken(t)

	for _, tc := range []struct {
		name, method, path, body, wantService string
	}{
		{name: "auth", method: http.MethodGet, path: "/api/v1/auth/me", wantService: "auth"},
		{name: "ledger", method: http.MethodPatch, path: "/api/v1/transactions/transaction-id?month=2026-09", body: `{"amount":"25.50"}`, wantService: "ledger"},
		{name: "analytics", method: http.MethodGet, path: "/api/v1/analytics/summary?month=2026-09", wantService: "analytics"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}
			request := httptest.NewRequest(tc.method, tc.path, body)
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set("X-Request-ID", "gateway-integration-request")
			request.Header.Set("X-User-ID", "client-supplied-user-id")
			response := httptest.NewRecorder()

			gateway.ServeHTTP(response, request)

			if response.Code != http.StatusOK || response.Body.String() != `{"service":"`+tc.wantService+`"}` {
				t.Fatalf("response = %d %q", response.Code, response.Body.String())
			}
			if response.Header().Get("X-Upstream-Service") != tc.wantService {
				t.Fatalf("upstream response header = %q, want %q", response.Header().Get("X-Upstream-Service"), tc.wantService)
			}
			if response.Header().Get("X-Request-ID") != "gateway-integration-request" {
				t.Fatalf("request ID = %q", response.Header().Get("X-Request-ID"))
			}
		})
	}
}

func TestGatewayIntegrationPreservesOAuthRedirectAndCookie(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/auth/google" {
			t.Errorf("upstream request = %s %s", r.Method, r.URL.Path)
		}
		http.SetCookie(w, &http.Cookie{
			Name: "finory_oauth", Value: "state.nonce.verifier",
			Path: "/api/v1/auth/google", MaxAge: 600, HttpOnly: true, SameSite: http.SameSiteLaxMode,
		})
		w.Header().Set("Location", "https://accounts.google.com/o/oauth2/v2/auth?state=example")
		w.WriteHeader(http.StatusFound)
	}))
	defer upstream.Close()

	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
	response := httptest.NewRecorder()

	newTestGateway(config.Config{AuthServiceURL: upstreamURL}).ServeHTTP(response, request)

	if response.Code != http.StatusFound || response.Header().Get("Location") != "https://accounts.google.com/o/oauth2/v2/auth?state=example" {
		t.Fatalf("OAuth response = %d, location=%q", response.Code, response.Header().Get("Location"))
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "finory_oauth" || cookies[0].Value != "state.nonce.verifier" ||
		cookies[0].Path != "/api/v1/auth/google" || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("OAuth flow cookie was not preserved: %+v", cookies)
	}
}

type gatewayUpstreamExpectation struct {
	name, method, path, query, body string
}

func newGatewayUpstream(t *testing.T, want gatewayUpstreamExpectation) (*httptest.Server, *url.URL) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != want.method || r.URL.Path != want.path || r.URL.RawQuery != want.query {
			t.Errorf("%s upstream request = %s %s, want %s %s", want.name, r.Method, r.URL.String(), want.method, want.path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != want.body {
			t.Errorf("%s upstream body = %q, want %q; error=%v", want.name, body, want.body, err)
		}
		if r.Header.Get("Authorization") == "" {
			t.Errorf("%s upstream is missing Authorization", want.name)
		}
		if r.Header.Get("X-User-ID") != integrationUserID {
			t.Errorf("%s upstream user ID = %q, want verified subject %q", want.name, r.Header.Get("X-User-ID"), integrationUserID)
		}
		if r.Header.Get("X-Request-ID") != "gateway-integration-request" {
			t.Errorf("%s upstream request ID = %q", want.name, r.Header.Get("X-Request-ID"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Upstream-Service", want.name)
		_, _ = w.Write([]byte(`{"service":"` + want.name + `"}`))
	}))
	t.Cleanup(server.Close)
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return server, serverURL
}
