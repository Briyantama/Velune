package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, a *domain.Account) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error)
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Account, int64, error)
	Update(ctx context.Context, a *domain.Account) error
	SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error
}
