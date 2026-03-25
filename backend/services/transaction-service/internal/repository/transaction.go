package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
)

type TransactionFilter struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	Type       *domain.TransactionType
	From       *time.Time
	To         *time.Time
	Currency   string
}

type TransactionRepository interface {
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error)
	List(ctx context.Context, userID uuid.UUID, f TransactionFilter, limit, offset int) ([]domain.Transaction, int64, error)
	SumIncomeExpenseInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (incomeMinor, expenseMinor int64, err error)
	SumExpensesByCategoryInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error)
}
