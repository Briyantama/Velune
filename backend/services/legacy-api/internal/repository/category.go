package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *domain.Category) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error)
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Category, int64, error)
	Update(ctx context.Context, c *domain.Category) error
	SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error
}
