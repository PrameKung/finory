package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	response := httptest.NewRecorder()
	New(nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK || response.Body.String() != "{\"status\":\"ok\"}\n" || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected health response: status=%d, body=%q, headers=%v", response.Code, response.Body.String(), response.Header())
	}
}
