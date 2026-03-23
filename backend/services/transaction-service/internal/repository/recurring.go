package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
)

type RecurringRepository interface {
	Create(ctx context.Context, r *domain.RecurringRule) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.RecurringRule, error)
	List(ctx context.Context, userID uuid.UUID, limit, offset int, activeOnly bool) ([]domain.RecurringRule, int64, error)
	Update(ctx context.Context, r *domain.RecurringRule) error
	SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error
	ListDue(ctx context.Context, before time.Time, limit int) ([]domain.RecurringRule, error)
}
