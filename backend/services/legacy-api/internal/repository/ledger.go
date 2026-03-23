package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
)

// Ledger applies transaction rows and keeps account balances consistent within DB transactions.
type Ledger interface {
	CreateTransaction(ctx context.Context, t *domain.Transaction) error
	SoftDeleteTransaction(ctx context.Context, userID, id uuid.UUID, version int64) error
}
