package reconciliation

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	txclient "github.com/moon-eye/velune/services/budget-service/internal/infrastructure/transactions"
	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/helper"
	"github.com/moon-eye/velune/shared/metrics"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
	"go.uber.org/zap"
)

// alertImpliedSpent derives cents/minor units last recorded for alert state from last_usage_percent.
func alertImpliedSpent(limit int64, lastUsagePercent float64) int64 {
	if limit <= 0 {
		return 0
	}
	return int64(math.Round(lastUsagePercent / 100.0 * float64(limit)))
}

func spentMinorExceedsTolerance(a, b int64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d > 1
}

// VerifyBudgetAlertDrift compares fresh transaction-summary spend to spend implied by budget_alert_state.
// Budgets without alert state are skipped (nothing stored to compare).
// Returns budgetsChecked (rows considered) and mismatchCount.
func VerifyBudgetAlertDrift(ctx context.Context, pool *pgxpool.Pool, txc *txclient.Client, log *zap.Logger) (budgetsChecked int, mismatchCount int, err error) {
	if txc == nil {
		return 0, 0, nil
	}
	q := db.New(pool)
	candidates, err := q.BudgetAlertDriftCandidates(ctx)
	if err != nil {
		return 0, 0, err
	}

	for _, row := range candidates {
		id := helper.FromPgUUID(row.ID)
		userID := helper.FromPgUUID(row.UserID)
		start := helper.FromPgDate(row.StartDate)
		end := helper.FromPgDate(row.EndDate)
		currency := row.Currency
		categoryID := helper.FromPgUUIDPtr(row.CategoryID)
		limitMinor := row.LimitAmountMinor
		lastPct := row.LastUsagePercent
		budgetsChecked++
		to := end.Add(24 * time.Hour)
		var spent int64
		if categoryID == nil {
			_, expense, err := txc.Summary(ctx, userID, start, to, currency)
			if err != nil {
				continue
			}
			spent = expense
		} else {
			byCat, err := txc.SummaryByCategory(ctx, userID, start, to, currency)
			if err != nil {
				continue
			}
			spent = byCat[*categoryID]
		}
		implied := alertImpliedSpent(limitMinor, lastPct)
		if !spentMinorExceedsTolerance(spent, implied) {
			continue
		}
		mismatchCount++
		metrics.ReconciliationMismatchTotal.WithLabelValues("budget_usage").Inc()
		details, _ := json.Marshal(map[string]any{
			"budgetId":                   id.String(),
			"userId":                     userID.String(),
			"spentFromTransactionsMinor": spent,
			"alertImpliedSpentMinor":     implied,
			"lastUsagePercent":           lastPct,
			"currency":                   currency,
		})
		if err := q.AuditLogInsert(ctx, db.AuditLogInsertParams{
			Type:    "budget_usage",
			Status:  "mismatch",
			Details: details,
		}); err != nil {
			log.Error("budget_reconcile_audit_insert", zap.Error(err))
			continue
		}
		payload, err := helper.ToJSONMarshal(contracts.BudgetMismatchDetected{
			BudgetID:                   id,
			UserID:                     userID,
			SpentFromTransactionsMinor: spent,
			AlertImpliedSpentMinor:     implied,
			LastUsagePercent:           lastPct,
			Currency:                   currency,
		})
		if err != nil {
			continue
		}
		eid := uuid.New()
		env := contracts.EventEnvelope{
			EventID:     eid,
			EventType:   contracts.EventBudgetMismatchDetected,
			Source:      "budget-service",
			OccurredAt:  time.Now().UTC(),
			UserID:      &userID,
			Idempotency: "reconcile:budget:" + id.String() + ":" + time.Now().UTC().Format(time.RFC3339Nano),
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
			log.Error("budget_reconcile_outbox", zap.Error(err))
			continue
		}
		log.Error("budget_mismatch_detected",
			zap.String("budget_id", id.String()),
			zap.Int64("spent_from_tx", spent),
			zap.Int64("alert_implied_spent", implied),
			zap.Float64("last_usage_percent", lastPct),
		)
	}
	return budgetsChecked, mismatchCount, nil
}
