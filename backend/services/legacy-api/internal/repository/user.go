package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}
