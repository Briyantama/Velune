package email

import (
	"context"
	"fmt"

	"github.com/moon-eye/velune/services/notification-service/internal/domain"
	"github.com/moon-eye/velune/services/notification-service/internal/repository"
	"github.com/moon-eye/velune/shared/sim"
	"go.uber.org/zap"
)

// ChaosSender wraps a DeliveryChannel and randomly fails Deliver when SimulateEmailFailure triggers.
type ChaosSender struct {
	Inner repository.DeliveryChannel
	Sim   *sim.Config
	Log   *zap.Logger
}

func (c *ChaosSender) Name() string {
	if c.Inner != nil {
		return c.Inner.Name()
	}
	return "email"
}

func (c *ChaosSender) Deliver(ctx context.Context, alert domain.OverspendAlert) error {
	if c.Sim != nil && c.Sim.SimulateEmailFailure() {
		if c.Log != nil {
			c.Log.Warn("sim_email_fail",
				zap.String("budget_id", alert.BudgetID.String()),
				zap.String("user_id", alert.UserID.String()),
				zap.String("channel", "email"),
			)
		}
		return fmt.Errorf("simulated email failure")
	}
	if c.Inner == nil {
		return fmt.Errorf("inner email sender not configured")
	}
	return c.Inner.Deliver(ctx, alert)
}
