package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
	"github.com/moon-eye/velune/services/auth-service/internal/repository"
)

type RefreshTokenRepo struct {
	s *Store
}

func NewRefreshTokenRepo(s *Store) repository.RefreshTokenRepository {
	return &RefreshTokenRepo{s: s}
}

func (r *RefreshTokenRepo) Store(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	const q = `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, version, created_at, updated_at)
VALUES ($1,$2,$3,$4,1,now(),now())`
	_, err := r.s.Pool.Exec(ctx, q, uuid.New(), userID, tokenHash, expiresAt.UTC())
	return err
}

func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	const q = `
SELECT id, user_id, token_hash, expires_at, version, created_at, updated_at, deleted_at
FROM refresh_tokens
WHERE token_hash = $1 AND deleted_at IS NULL AND expires_at > now()`
	row := r.s.Pool.QueryRow(ctx, q, tokenHash)
	var rr domain.RefreshToken
	err := row.Scan(
		&rr.ID,
		&rr.UserID,
		&rr.TokenHash,
		&rr.ExpiresAt,
		&rr.Version,
		&rr.CreatedAt,
		&rr.UpdatedAt,
		&rr.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rr, nil
}

func (r *RefreshTokenRepo) Rotate(ctx context.Context, tokenID uuid.UUID, newTokenHash string, newExpiresAt time.Time) error {
	const q = `
UPDATE refresh_tokens
SET token_hash = $1,
    expires_at = $2,
    version = version + 1,
    updated_at = now()
WHERE id = $3 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q, newTokenHash, newExpiresAt.UTC(), tokenID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *RefreshTokenRepo) SoftDelete(ctx context.Context, tokenID uuid.UUID) error {
	const q = `
UPDATE refresh_tokens
SET deleted_at = now(),
    version = version + 1,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q, tokenID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
