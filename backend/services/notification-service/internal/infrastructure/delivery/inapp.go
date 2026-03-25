package delivery

import (
	"context"

	"github.com/moon-eye/velune/services/notification-service/internal/domain"
	"go.uber.org/zap"
)

type InAppChannel struct {
	Log *zap.Logger
}

func (c *InAppChannel) Name() string { return "inapp" }

func (c *InAppChannel) Deliver(_ context.Context, alert domain.OverspendAlert) error {
	c.Log.Info("overspend_inapp",
		zap.String("budget_id", alert.BudgetID.String()),
		zap.String("user_id", alert.UserID.String()),
		zap.Float64("usage_percent", alert.UsagePercent),
		zap.Int64("spent_minor", alert.SpentMinor),
		zap.Int64("limit_minor", alert.LimitAmountMinor),
		zap.String("currency", alert.Currency),
	)
	return nil
}
