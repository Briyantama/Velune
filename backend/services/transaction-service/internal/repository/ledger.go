package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
)

// Ledger applies transaction rows and keeps account balances consistent within DB transactions.
type Ledger interface {
	CreateTransaction(ctx context.Context, t *domain.Transaction) error
	UpdateTransaction(ctx context.Context, userID uuid.UUID, next *domain.Transaction, prevVersion int64) error
	SoftDeleteTransaction(ctx context.Context, userID, id uuid.UUID, version int64) error
}
