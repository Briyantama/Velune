package reconciliation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/helper"
	"github.com/moon-eye/velune/shared/metrics"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
	"go.uber.org/zap"
)

// ReconcileAccountBalances compares accounts.balance_minor to net ledger sum per account.
// Returns accountsChecked and mismatchCount (audit/outbox rows attempted per mismatch).
func ReconcileAccountBalances(ctx context.Context, pool *pgxpool.Pool, log *zap.Logger) (accountsChecked int, mismatchCount int, err error) {
	q := db.New(pool)
	ledgerRows, err := q.ReconcileAccountStoredVsLedger(ctx)
	if err != nil {
		return 0, 0, err
	}

	for _, row := range ledgerRows {
		id := helper.FromPgUUID(row.ID)
		userID := helper.FromPgUUID(row.UserID)
		stored := row.BalanceMinor
		ledgerSum := row.LedgerSum
		currency := row.Currency
		accountsChecked++
		if stored == ledgerSum {
			continue
		}
		mismatchCount++
		metrics.ReconciliationMismatchTotal.WithLabelValues("balance").Inc()
		details, _ := json.Marshal(map[string]any{
			"accountId":          id.String(),
			"userId":             userID.String(),
			"storedBalanceMinor": stored,
			"ledgerSumMinor":     ledgerSum,
			"currency":           currency,
		})
		if err := q.AuditLogInsert(ctx, db.AuditLogInsertParams{
			Type:    "balance",
			Status:  "mismatch",
			Details: details,
		}); err != nil {
			log.Error("reconcile_audit_insert", zap.Error(err))
			continue
		}
		payload, err := helper.ToJSONMarshal(contracts.BalanceMismatchDetected{
			AccountID:          id,
			UserID:             userID,
			StoredBalanceMinor: stored,
			LedgerSumMinor:     ledgerSum,
			Currency:           currency,
		})
		if err != nil {
			log.Error("reconcile_payload", zap.Error(err))
			continue
		}
		eid := uuid.New()
		env := contracts.EventEnvelope{
			EventID:     eid,
			EventType:   contracts.EventBalanceMismatchDetected,
			Source:      "transaction-service",
			OccurredAt:  time.Now().UTC(),
			UserID:      &userID,
			Idempotency: "reconcile:balance:" + id.String() + ":" + time.Now().UTC().Format(time.RFC3339Nano),
			Payload:     payload,
		}
		body, err := helper.ToJSONMarshal(env)
		if err != nil {
			continue
		}
		if err := q.OutboxInsert(ctx, db.OutboxInsertParams{
			ID:        helper.ToPgUUID(eid),
			EventType: env.EventType,
			Payload:   body,
		}); err != nil {
			log.Error("reconcile_outbox", zap.Error(err))
			continue
		}
		log.Error("balance_mismatch_detected",
			zap.String("account_id", id.String()),
			zap.String("user_id", userID.String()),
			zap.Int64("stored", stored),
			zap.Int64("ledger_sum", ledgerSum),
		)
	}
	return accountsChecked, mismatchCount, nil
}
