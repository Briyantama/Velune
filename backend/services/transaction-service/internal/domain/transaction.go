package domain

import (
	"time"

	"github.com/google/uuid"
)

// Transaction is owned only by transaction-service (integrate via IDs with other services).
type Transaction struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	AccountID             uuid.UUID
	CategoryID            *uuid.UUID
	CounterpartyAccountID *uuid.UUID
	AmountMinor           int64
	Currency              string
	Type                  string
	Description           string
	OccurredAt            time.Time
	Version               int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
