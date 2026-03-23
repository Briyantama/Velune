package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionIncome    TransactionType = "income"
	TransactionExpense   TransactionType = "expense"
	TransactionTransfer  TransactionType = "transfer"
	TransactionAdjustment TransactionType = "adjustment"
)

type Transaction struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	AccountID       uuid.UUID
	CategoryID      *uuid.UUID
	CounterpartyAccountID *uuid.UUID // transfers
	AmountMinor     int64
	Currency        string // ISO 4217
	Type            TransactionType
	Description     string
	OccurredAt      time.Time
	Version         int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
