package contracts

import (
	"time"

	"github.com/google/uuid"
)

// TransactionCreated is an integration-safe event payload shape.
type TransactionCreated struct {
	TransactionID uuid.UUID  `json:"transactionId"`
	UserID        uuid.UUID  `json:"userId"`
	AccountID     uuid.UUID  `json:"accountId"`
	CategoryID    *uuid.UUID `json:"categoryId,omitempty"`
	AmountMinor   int64      `json:"amountMinor"`
	Currency      string     `json:"currency"`
	Type          string     `json:"type"`
	OccurredAt    time.Time  `json:"occurredAt"`
	Version       int64      `json:"version"`
}

type TransactionUpdated struct {
	TransactionID uuid.UUID `json:"transactionId"`
	UserID        uuid.UUID `json:"userId"`
	AmountMinor   int64     `json:"amountMinor"`
	Currency      string    `json:"currency"`
	Type          string    `json:"type"`
	OccurredAt    time.Time `json:"occurredAt"`
	Version       int64     `json:"version"`
}

type TransactionDeleted struct {
	TransactionID uuid.UUID `json:"transactionId"`
	UserID        uuid.UUID `json:"userId"`
	Version       int64     `json:"version"`
	DeletedAt     time.Time `json:"deletedAt"`
}

type TransactionSummary struct {
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	Currency     string    `json:"currency"`
	IncomeMinor  int64     `json:"incomeMinor"`
	ExpenseMinor int64     `json:"expenseMinor"`
	NetMinor     int64     `json:"netMinor"`
}

type BudgetRelevantTransaction struct {
	CategoryID *uuid.UUID `json:"categoryId,omitempty"`
	SpentMinor int64      `json:"spentMinor"`
	Currency   string     `json:"currency"`
}

type BalanceSnapshot struct {
	UserID       uuid.UUID `json:"userId"`
	AccountID    uuid.UUID `json:"accountId"`
	BalanceMinor int64     `json:"balanceMinor"`
	Currency     string    `json:"currency"`
	AsOf         time.Time `json:"asOf"`
}
