package repository

import (
	"context"

	"github.com/moon-eye/velune/shared/contracts"
)

type TransactionAnalyticsRepository interface {
	Summary(ctx context.Context, userID string, q contracts.TransactionAnalyticsQuery) (*contracts.TransactionSummary, error)
	SummaryByCategory(ctx context.Context, userID string, q contracts.TransactionAnalyticsQuery) (*contracts.TransactionCategoryTotalsResponse, error)
}
