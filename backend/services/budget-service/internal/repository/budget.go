package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/budget-service/internal/domain"
)

type BudgetRepository interface {
	Create(ctx context.Context, b *domain.Budget) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Budget, error)
	List(ctx context.Context, userID uuid.UUID, limit, offset int, activeOn *time.Time) ([]domain.Budget, int64, error)
	Update(ctx context.Context, b *domain.Budget) error
	SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error
}
