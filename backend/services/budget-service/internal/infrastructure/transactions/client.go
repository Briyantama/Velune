package transactions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Summary(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (int64, int64, error) {
	u, err := url.Parse(c.BaseURL + "/api/v1/transactions/summary")
	if err != nil {
		return 0, 0, err
	}
	q := u.Query()
	q.Set("from", from.UTC().Format(time.RFC3339))
	q.Set("to", to.UTC().Format(time.RFC3339))
	q.Set("currency", currency)
	u.RawQuery = q.Encode()

	var out contracts.TransactionSummary
	if err := c.get(ctx, u.String(), userID, &out); err != nil {
		return 0, 0, err
	}
	return out.IncomeMinor, out.ExpenseMinor, nil
}

func (c *Client) SummaryByCategory(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error) {
	u, err := url.Parse(c.BaseURL + "/api/v1/transactions/summary/categories")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("from", from.UTC().Format(time.RFC3339))
	q.Set("to", to.UTC().Format(time.RFC3339))
	q.Set("currency", currency)
	u.RawQuery = q.Encode()

	var out struct {
		Totals map[uuid.UUID]int64 `json:"totals"`
	}
	if err := c.get(ctx, u.String(), userID, &out); err != nil {
		return nil, err
	}
	return out.Totals, nil
}

func (c *Client) get(ctx context.Context, uri string, userID uuid.UUID, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return err
	}
	// budget-service forwards user id through internal trusted header.
	req.Header.Set("X-User-ID", userID.String())
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return errs.New("UPSTREAM_ERROR", fmt.Sprintf("transaction-service returned %d", resp.StatusCode), http.StatusBadGateway)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

