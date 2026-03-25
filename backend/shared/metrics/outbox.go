package metrics

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

// RefreshOutboxPending updates outbox_pending_total for rows eligible for dispatch.
func RefreshOutboxPending(ctx context.Context, pool *pgxpool.Pool, maxRetry int) {
	if pool == nil || maxRetry <= 0 {
		return
	}
	q := db.New(pool)
	n, err := q.OutboxCountRetryEligible(ctx, int32(maxRetry))
	if err != nil {
		return
	}
	OutboxPending.Set(float64(n))
}
