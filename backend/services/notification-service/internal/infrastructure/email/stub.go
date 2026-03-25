package email

import (
	"context"

	"github.com/moon-eye/velune/services/notification-service/internal/domain"
	"go.uber.org/zap"
)

type StubSender struct {
	Log  *zap.Logger
	From string
}

func (s *StubSender) Name() string { return "email" }

func (s *StubSender) Deliver(_ context.Context, alert domain.OverspendAlert) error {
	s.Log.Info("overspend_email_stub",
		zap.String("from", s.From),
		zap.String("budget_id", alert.BudgetID.String()),
		zap.String("user_id", alert.UserID.String()),
		zap.Float64("usage_percent", alert.UsagePercent),
		zap.Int64("spent_minor", alert.SpentMinor),
		zap.Int64("limit_minor", alert.LimitAmountMinor),
		zap.String("currency", alert.Currency),
	)
	return nil
}
