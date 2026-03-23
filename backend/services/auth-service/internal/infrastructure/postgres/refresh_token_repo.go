package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
	"github.com/moon-eye/velune/services/auth-service/internal/repository"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/sqlc/generated"
)

type RefreshTokenRepo struct {
	s *Store
}

func NewRefreshTokenRepo(s *Store) repository.RefreshTokenRepository {
	return &RefreshTokenRepo{s: s}
}

func (r *RefreshTokenRepo) Store(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return r.s.Queries.StoreRefreshToken(ctx, db.StoreRefreshTokenParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt.UTC(),
			Valid: true,
		},
	})
}

func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	rr, err := r.s.Queries.GetRefreshTokenByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var deletedAt *time.Time
	if rr.DeletedAt.Valid {
		t := rr.DeletedAt.Time
		deletedAt = &t
	}

	return &domain.RefreshToken{
		ID:         uuid.UUID(rr.ID.Bytes),
		UserID:     uuid.UUID(rr.UserID.Bytes),
		TokenHash:  rr.TokenHash,
		ExpiresAt:  rr.ExpiresAt.Time,
		Version:    rr.Version,
		CreatedAt:  rr.CreatedAt.Time,
		UpdatedAt:  rr.UpdatedAt.Time,
		DeletedAt:  deletedAt,
	}, nil
}

func (r *RefreshTokenRepo) Rotate(ctx context.Context, tokenID uuid.UUID, newTokenHash string, newExpiresAt time.Time) error {
	rows, err := r.s.Queries.RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		TokenHash: newTokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  newExpiresAt.UTC(),
			Valid: true,
		},
		ID: pgtype.UUID{
			Bytes: tokenID,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return errs.ErrRefreshToken
	}
	return nil
}

func (r *RefreshTokenRepo) SoftDelete(ctx context.Context, tokenID uuid.UUID) error {
	rows, err := r.s.Queries.SoftDeleteRefreshToken(ctx, pgtype.UUID{
		Bytes: tokenID,
		Valid: true,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return errs.ErrRefreshToken
	}
	return nil
}
