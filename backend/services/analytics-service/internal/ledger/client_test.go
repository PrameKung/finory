package ledger

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

const testUserID = "11111111-1111-4111-8111-111111111111"

func TestGetMonthlyData(t *testing.T) {
	var requests int
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.Header.Get(userIDHeader) != testUserID || request.Header.Get(requestIDHeader) != "request-123" {
			t.Fatalf("identity headers were not propagated: %v", request.Header)
		}

		var body string
		switch request.URL.Path {
		case "/transactions":
			if request.URL.Query().Get("month") != "2026-09" {
				t.Fatalf("unexpected month: %q", request.URL.Query().Get("month"))
			}
			body = `[{"id":"transaction-1","categoryId":"category-1","walletId":"wallet-1","type":"expense","amount":"42.50","description":null,"transactionDate":"2026-09-12","createdAt":"2026-09-12T00:00:00Z","updatedAt":"2026-09-12T00:00:00Z"}]`
		case "/categories":
			body = `[{"id":"category-1","name":"Food","type":"expense","icon":null,"color":"#ffffff","isDefault":true,"createdAt":"2026-09-01T00:00:00Z","updatedAt":"2026-09-01T00:00:00Z"}]`
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
		return response(http.StatusOK, body), nil
	})}

	baseURL, err := url.Parse("http://ledger-service:8082")
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, httpClient)
	data, err := client.GetMonthlyData(context.Background(), testUserID, "request-123", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	if requests != 2 || len(data.Transactions) != 1 || len(data.Categories) != 1 {
		t.Fatalf("unexpected monthly data: requests=%d data=%+v", requests, data)
	}
	if data.Transactions[0].Amount != "42.50" || data.Categories[0].Name != "Food" {
		t.Fatalf("ledger values were not decoded: %+v", data)
	}
}

func TestGetMonthlyDataStopsOnTransactionFailure(t *testing.T) {
	var requests int
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		requests++
		return response(http.StatusBadGateway, `{}`), nil
	})}

	baseURL, err := url.Parse("http://ledger-service:8082")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewClient(baseURL, httpClient).GetMonthlyData(
		context.Background(), testUserID, "", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
	)
	var upstreamError *UpstreamError
	if !errors.As(err, &upstreamError) || upstreamError.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected upstream error, got %v", err)
	}
	if requests != 1 {
		t.Fatalf("categories should not be fetched after transaction failure; requests=%d", requests)
	}
}

func TestGetMonthlyDataRejectsInvalidInput(t *testing.T) {
	baseURL, err := url.Parse("http://ledger-service:8082")
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, http.DefaultClient)

	for name, testCase := range map[string]struct {
		userID string
		month  time.Time
	}{
		"missing user":  {month: time.Now()},
		"missing month": {userID: testUserID},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := client.GetMonthlyData(context.Background(), testCase.userID, "", testCase.month)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("expected invalid request, got %v", err)
			}
		})
	}
}

func TestGetMonthlyDataRejectsMalformedResponse(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return response(http.StatusOK, `not-json`), nil
	})}

	baseURL, err := url.Parse("http://ledger-service:8082")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewClient(baseURL, httpClient).GetMonthlyData(
		context.Background(), testUserID, "", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatal("expected malformed response to fail")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
