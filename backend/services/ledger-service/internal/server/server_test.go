package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadinessChecksDatabase(t *testing.T) {
	for _, tc := range []struct {
		name       string
		checkError error
		wantStatus int
		wantBody   string
	}{
		{name: "connected", wantStatus: http.StatusOK, wantBody: "{\"status\":\"ok\"}\n"},
		{name: "disconnected", checkError: errors.New("database unavailable"), wantStatus: http.StatusServiceUnavailable, wantBody: "{\"status\":\"unavailable\"}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := New(func(context.Context) error { return tc.checkError }, nil, nil, nil, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
			if response.Code != tc.wantStatus || response.Body.String() != tc.wantBody {
				t.Fatalf("readiness response = %d %q, want %d %q", response.Code, response.Body.String(), tc.wantStatus, tc.wantBody)
			}
		})
	}
}
