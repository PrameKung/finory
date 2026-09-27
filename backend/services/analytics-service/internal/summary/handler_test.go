package summary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

const testUserID = "11111111-1111-4111-8111-111111111111"

type fakeSummaryService struct {
	result           MonthlySummary
	categoryResult   CategoryDistribution
	trendResult      TrendSeries
	comparisonResult MonthlyComparison
	err              error
	userID           string
	requestID        string
	month            time.Time
}

func (f *fakeSummaryService) Categories(_ context.Context, userID, requestID string, month time.Time) (CategoryDistribution, error) {
	f.userID = userID
	f.requestID = requestID
	f.month = month
	return f.categoryResult, f.err
}

func (f *fakeSummaryService) Trends(_ context.Context, userID, requestID string, month time.Time) (TrendSeries, error) {
	f.userID = userID
	f.requestID = requestID
	f.month = month
	return f.trendResult, f.err
}

func (f *fakeSummaryService) MonthlyComparison(_ context.Context, userID, requestID string, month time.Time) (MonthlyComparison, error) {
	f.userID = userID
	f.requestID = requestID
	f.month = month
	return f.comparisonResult, f.err
}

func (f *fakeSummaryService) Monthly(_ context.Context, userID, requestID string, month time.Time) (MonthlySummary, error) {
	f.userID = userID
	f.requestID = requestID
	f.month = month
	return f.result, f.err
}

func TestMonthlySummaryHandler(t *testing.T) {
	service := &fakeSummaryService{result: MonthlySummary{
		Month: "2026-09", Income: "100.0000", Expense: "40.0000", NetBalance: "60.0000",
	}}
	e := echo.New()
	NewHandler(service).Register(e)
	request := httptest.NewRequest(http.MethodGet, "/analytics/summary?month=2026-09", nil)
	request.Header.Set(userIDHeader, testUserID)
	request.Header.Set(echo.HeaderXRequestID, "request-123")
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != "{\"month\":\"2026-09\",\"income\":\"100.0000\",\"expense\":\"40.0000\",\"netBalance\":\"60.0000\"}\n" {
		t.Fatalf("unexpected response: status=%d body=%q", response.Code, response.Body.String())
	}
	if service.userID != testUserID || service.requestID != "request-123" || service.month.Format("2006-01") != "2026-09" {
		t.Fatalf("request scope was not forwarded: %+v", service)
	}
}

func TestMonthlySummaryHandlerValidatesRequest(t *testing.T) {
	for name, testCase := range map[string]struct {
		userID string
		path   string
		status int
	}{
		"missing user":  {path: "/analytics/summary?month=2026-09", status: http.StatusUnauthorized},
		"invalid user":  {userID: "not-a-uuid", path: "/analytics/summary?month=2026-09", status: http.StatusUnauthorized},
		"invalid month": {userID: testUserID, path: "/analytics/summary?month=2026-13", status: http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			NewHandler(&fakeSummaryService{}).Register(e)
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request.Header.Set(userIDHeader, testCase.userID)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			if response.Code != testCase.status {
				t.Fatalf("status=%d, want %d; body=%q", response.Code, testCase.status, response.Body.String())
			}
		})
	}
}

func TestMonthlySummaryHandlerReportsLedgerFailure(t *testing.T) {
	e := echo.New()
	NewHandler(&fakeSummaryService{err: errors.New("unavailable")}).Register(e)
	request := httptest.NewRequest(http.MethodGet, "/analytics/summary?month=2026-09", nil)
	request.Header.Set(userIDHeader, testUserID)
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway || response.Body.String() != "{\"error\":\"ledger_service_unavailable\"}\n" {
		t.Fatalf("unexpected response: status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestCategoryDistributionHandler(t *testing.T) {
	service := &fakeSummaryService{categoryResult: CategoryDistribution{
		Month: "2026-09", TotalExpense: "40.0000",
		Categories: []CategoryRank{{Rank: 1, CategoryID: "food", Name: "Food", Amount: "40.0000", Percentage: "100.00"}},
	}}
	e := echo.New()
	NewHandler(service).Register(e)
	request := httptest.NewRequest(http.MethodGet, "/analytics/categories?month=2026-09", nil)
	request.Header.Set(userIDHeader, testUserID)
	request.Header.Set(echo.HeaderXRequestID, "request-123")
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	want := "{\"month\":\"2026-09\",\"totalExpense\":\"40.0000\",\"categories\":[{\"rank\":1,\"categoryId\":\"food\",\"name\":\"Food\",\"icon\":null,\"color\":null,\"amount\":\"40.0000\",\"percentage\":\"100.00\"}]}\n"
	if response.Code != http.StatusOK || response.Body.String() != want {
		t.Fatalf("unexpected response: status=%d body=%q", response.Code, response.Body.String())
	}
	if service.userID != testUserID || service.requestID != "request-123" || service.month.Format("2006-01") != "2026-09" {
		t.Fatalf("request scope was not forwarded: %+v", service)
	}
}

func TestIncomeExpenseTrendsHandler(t *testing.T) {
	service := &fakeSummaryService{trendResult: TrendSeries{
		Month: "2026-09", Granularity: "day",
		Points: []TrendPoint{{Date: "2026-09-01", Income: "10.0000", Expense: "5.0000"}},
	}}
	e := echo.New()
	NewHandler(service).Register(e)
	request := httptest.NewRequest(http.MethodGet, "/analytics/trends?month=2026-09", nil)
	request.Header.Set(userIDHeader, testUserID)
	request.Header.Set(echo.HeaderXRequestID, "request-123")
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	var result TrendSeries
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected response: status=%d body=%q", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Month != "2026-09" || result.Granularity != "day" || len(result.Points) != 1 || result.Points[0].Income != "10.0000" {
		t.Fatalf("unexpected trend response: %+v", result)
	}
	if service.userID != testUserID || service.requestID != "request-123" || service.month.Format("2006-01") != "2026-09" {
		t.Fatalf("request scope was not forwarded: %+v", service)
	}
}

func TestMonthOverMonthComparisonHandler(t *testing.T) {
	percentage := "25.00"
	service := &fakeSummaryService{comparisonResult: MonthlyComparison{
		Month: "2026-09", PreviousMonth: "2026-08",
		Current:  MonthlyTotals{Income: "125.0000", Expense: "0.0000"},
		Previous: MonthlyTotals{Income: "100.0000", Expense: "0.0000"},
		Changes: MonthlyChanges{
			Income:  AmountChange{Amount: "25.0000", Percentage: &percentage},
			Expense: AmountChange{Amount: "0.0000"},
		},
	}}
	e := echo.New()
	NewHandler(service).Register(e)
	request := httptest.NewRequest(http.MethodGet, "/analytics/monthly?month=2026-09", nil)
	request.Header.Set(userIDHeader, testUserID)
	request.Header.Set(echo.HeaderXRequestID, "request-123")
	response := httptest.NewRecorder()

	e.ServeHTTP(response, request)

	var result MonthlyComparison
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected response: status=%d body=%q", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Month != "2026-09" || result.PreviousMonth != "2026-08" ||
		result.Changes.Income.Percentage == nil || *result.Changes.Income.Percentage != "25.00" ||
		result.Changes.Expense.Percentage != nil {
		t.Fatalf("unexpected comparison response: %+v", result)
	}
	if service.userID != testUserID || service.requestID != "request-123" || service.month.Format("2006-01") != "2026-09" {
		t.Fatalf("request scope was not forwarded: %+v", service)
	}
}
