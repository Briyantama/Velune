package contracts

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventBudgetUsageEvaluated    = "budget.usage.evaluated"
	EventOverspendAlertRequested = "notification.overspend.requested"
	EventNotificationDispatched  = "notification.dispatched"
)

type BudgetUsageEvaluated struct {
	BudgetID         uuid.UUID  `json:"budgetId"`
	UserID           uuid.UUID  `json:"userId"`
	CategoryID       *uuid.UUID `json:"categoryId,omitempty"`
	From             time.Time  `json:"from"`
	To               time.Time  `json:"to"`
	Currency         string     `json:"currency"`
	LimitAmountMinor int64      `json:"limitAmountMinor"`
	SpentMinor       int64      `json:"spentMinor"`
	RemainingMinor   int64      `json:"remainingMinor"`
	OverspentMinor   int64      `json:"overspentMinor"`
	UsagePercent     float64    `json:"usagePercent"`
}

type OverspendAlertRequested struct {
	BudgetID         uuid.UUID  `json:"budgetId"`
	UserID           uuid.UUID  `json:"userId"`
	CategoryID       *uuid.UUID `json:"categoryId,omitempty"`
	Currency         string     `json:"currency"`
	LimitAmountMinor int64      `json:"limitAmountMinor"`
	SpentMinor       int64      `json:"spentMinor"`
	UsagePercent     float64    `json:"usagePercent"`
	IsOverspent      bool       `json:"isOverspent"`
}

type NotificationDispatched struct {
	EventID      uuid.UUID `json:"eventId"`
	Kind         string    `json:"kind"`
	Channel      string    `json:"channel"`
	Status       string    `json:"status"`
	Reason       string    `json:"reason,omitempty"`
	DispatchedAt time.Time `json:"dispatchedAt"`
}
