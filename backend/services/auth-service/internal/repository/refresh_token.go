package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
)

// RefreshTokenRepository persists opaque refresh token references hashed at rest.
type RefreshTokenRepository interface {
	Store(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Rotate(ctx context.Context, tokenID uuid.UUID, newTokenHash string, newExpiresAt time.Time) error
	SoftDelete(ctx context.Context, tokenID uuid.UUID) error
}