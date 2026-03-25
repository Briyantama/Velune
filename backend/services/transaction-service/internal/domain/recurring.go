package domain

import (
	"time"

	"github.com/google/uuid"
)

type RecurringFrequency string

const (
	RecurringDaily   RecurringFrequency = "daily"
	RecurringWeekly  RecurringFrequency = "weekly"
	RecurringMonthly RecurringFrequency = "monthly"
	RecurringYearly  RecurringFrequency = "yearly"
)

type RecurringRule struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	AccountID   uuid.UUID
	CategoryID  *uuid.UUID
	AmountMinor int64
	Currency    string
	Type        TransactionType
	Frequency   RecurringFrequency
	NextRunAt   time.Time
	LastRunAt   *time.Time
	IsActive    bool
	Description string
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
