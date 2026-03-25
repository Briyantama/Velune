package domain

import "github.com/google/uuid"

type OverspendAlert struct {
	BudgetID         uuid.UUID
	UserID           uuid.UUID
	Currency         string
	LimitAmountMinor int64
	SpentMinor       int64
	UsagePercent     float64
	IsOverspent      bool
}
