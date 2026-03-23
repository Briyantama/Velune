package domain

import (
	"time"

	"github.com/google/uuid"
)

type BudgetPeriod string

const (
	BudgetPeriodMonthly BudgetPeriod = "monthly"
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodCustom  BudgetPeriod = "custom"
)

type Budget struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Name            string
	PeriodType      BudgetPeriod
	CategoryID      *uuid.UUID
	StartDate       time.Time
	EndDate         time.Time
	LimitAmountMinor int64
	Currency        string
	Version         int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
