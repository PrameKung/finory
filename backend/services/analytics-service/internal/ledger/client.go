package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	userIDHeader    = "X-User-ID"
	requestIDHeader = "X-Request-ID"
	maxResponseSize = 4 << 20
)

var ErrInvalidRequest = errors.New("invalid ledger request")

// UpstreamError reports a non-success response from Ledger Service without
// exposing its response body to callers.
type UpstreamError struct {
	StatusCode int
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("ledger service returned status %d", e.StatusCode)
}

type Transaction struct {
	ID              string  `json:"id"`
	CategoryID      string  `json:"categoryId"`
	WalletID        string  `json:"walletId"`
	Type            string  `json:"type"`
	Amount          string  `json:"amount"`
	Description     *string `json:"description"`
	TransactionDate string  `json:"transactionDate"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type Category struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Icon      *string `json:"icon"`
	Color     *string `json:"color"`
	IsDefault bool    `json:"isDefault"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// MonthlyData is the Ledger-owned data needed to build the MVP analytics
// views. Monetary values remain decimal strings so they are never converted
// through floating-point numbers at the service boundary.
type MonthlyData struct {
	Transactions []Transaction
	Categories   []Category
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewClient(baseURL *url.URL, httpClient *http.Client) *Client {
	return &Client{baseURL: cloneURL(baseURL), httpClient: httpClient}
}

func (c *Client) ListTransactions(ctx context.Context, userID, requestID string, month time.Time) ([]Transaction, error) {
	if err := c.validateRequest(userID); err != nil {
		return nil, err
	}
	if month.IsZero() {
		return nil, ErrInvalidRequest
	}

	transactionsURL := c.endpoint("/transactions")
	query := transactionsURL.Query()
	query.Set("month", month.Format("2006-01"))
	transactionsURL.RawQuery = query.Encode()

	var transactions []Transaction
	if err := c.get(ctx, transactionsURL, userID, requestID, &transactions); err != nil {
		return nil, fmt.Errorf("get monthly transactions: %w", err)
	}
	return transactions, nil
}

func (c *Client) ListCategories(ctx context.Context, userID, requestID string) ([]Category, error) {
	if err := c.validateRequest(userID); err != nil {
		return nil, err
	}

	var categories []Category
	if err := c.get(ctx, c.endpoint("/categories"), userID, requestID, &categories); err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	return categories, nil
}

// GetMonthlyData obtains user-scoped transactions and category metadata only
// through Ledger Service's HTTP API.
func (c *Client) GetMonthlyData(ctx context.Context, userID, requestID string, month time.Time) (MonthlyData, error) {
	transactions, err := c.ListTransactions(ctx, userID, requestID, month)
	if err != nil {
		return MonthlyData{}, err
	}

	categories, err := c.ListCategories(ctx, userID, requestID)
	if err != nil {
		return MonthlyData{}, err
	}

	return MonthlyData{Transactions: transactions, Categories: categories}, nil
}

func (c *Client) validateRequest(userID string) error {
	if c == nil || c.baseURL == nil || c.httpClient == nil || strings.TrimSpace(userID) == "" {
		return ErrInvalidRequest
	}
	return nil
}

func (c *Client) get(ctx context.Context, endpoint *url.URL, userID, requestID string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Set(userIDHeader, userID)
	if requestID != "" {
		request.Header.Set(requestIDHeader, requestID)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseSize))
		return &UpstreamError{StatusCode: response.StatusCode}
	}

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize))
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("decode response: expected one JSON value")
	}
	return nil
}

func (c *Client) endpoint(resourcePath string) *url.URL {
	endpoint := cloneURL(c.baseURL)
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + resourcePath
	endpoint.RawPath = ""
	endpoint.RawQuery = ""
	endpoint.Fragment = ""
	return endpoint
}

func cloneURL(value *url.URL) *url.URL {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
