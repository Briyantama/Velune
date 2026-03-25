package transactions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/otelx"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    otelx.TracedHTTPClient(&http.Client{Timeout: 15 * time.Second}),
	}
}

func (c *Client) Summary(ctx context.Context, userID string, q contracts.TransactionAnalyticsQuery) (*contracts.TransactionSummary, error) {
	var out contracts.TransactionSummary
	if err := c.getJSON(ctx, userID, "/api/v1/transactions/summary", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SummaryByCategory(ctx context.Context, userID string, q contracts.TransactionAnalyticsQuery) (*contracts.TransactionCategoryTotalsResponse, error) {
	var out contracts.TransactionCategoryTotalsResponse
	if err := c.getJSON(ctx, userID, "/api/v1/transactions/summary/categories", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, userID, path string, q contracts.TransactionAnalyticsQuery, out any) error {
	if c.BaseURL == "" {
		return errs.New("UPSTREAM_UNAVAILABLE", "transaction service URL is not configured", constx.StatusBadGateway)
	}
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return errs.New("INTERNAL_ERROR", "invalid transaction service URL", constx.StatusInternalServerError)
	}
	query := u.Query()
	query.Set("from", q.From.Format(time.RFC3339))
	query.Set("to", q.To.Format(time.RFC3339))
	query.Set("currency", q.Currency)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, constx.MethodGet, u.String(), nil)
	if err != nil {
		return errs.ErrInternal
	}
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return errs.New("UPSTREAM_UNAVAILABLE", "transaction service is unavailable", constx.StatusBadGateway)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return errs.New("UPSTREAM_ERROR", fmt.Sprintf("transaction service returned status %d", resp.StatusCode), constx.StatusBadGateway)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return errs.New("UPSTREAM_BAD_RESPONSE", "failed to decode transaction service response", constx.StatusBadGateway)
	}
	return nil
}
