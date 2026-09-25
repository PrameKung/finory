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
	}{
		{name: "connected", wantStatus: http.StatusOK},
		{name: "disconnected", checkError: errors.New("database unavailable"), wantStatus: http.StatusServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := New(func(context.Context) error { return tc.checkError })
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
			if response.Code != tc.wantStatus {
				t.Fatalf("readiness status = %d, want %d", response.Code, tc.wantStatus)
			}
		})
	}
}
