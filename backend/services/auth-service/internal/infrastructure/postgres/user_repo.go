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
	"github.com/moon-eye/velune/shared/sqlc/generated"
	"github.com/moon-eye/velune/shared/stringx"
)

type UserRepo struct {
	s *Store
}

func NewUserRepo(s *Store) repository.UserRepository {
	return &UserRepo{s: s}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	var emailVerifiedAt pgtype.Timestamptz
	if u.EmailVerifiedAt != nil {
		emailVerifiedAt = pgtype.Timestamptz{
			Time:  *u.EmailVerifiedAt,
			Valid: true,
		}
	} else {
		emailVerifiedAt = pgtype.Timestamptz{Valid: false}
	}
	return r.s.Queries.CreateUser(ctx, db.CreateUserParams{
		ID: pgtype.UUID{
			Bytes: u.ID,
			Valid: true,
		},
		Email:           stringx.Lower(u.Email),
		PasswordHash:    u.PasswordHash,
		BaseCurrency:    stringx.Upper(u.BaseCurrency),
		Status:          u.Status,
		EmailVerifiedAt: emailVerifiedAt,
		DisplayName:     u.DisplayName,
		Version:         u.Version,
		CreatedAt: pgtype.Timestamptz{
			Time:  u.CreatedAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  u.UpdatedAt,
			Valid: true,
		},
	})
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := r.s.Queries.GetUserByID(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		t := u.DeletedAt.Time
		deletedAt = &t
	}

	return &domain.User{
		ID:           uuid.UUID(u.ID.Bytes),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		BaseCurrency: u.BaseCurrency,
		Status:       u.Status,
		EmailVerifiedAt: func() *time.Time {
			if u.EmailVerifiedAt.Valid {
				t := u.EmailVerifiedAt.Time
				return &t
			}
			return nil
		}(),
		DisplayName: u.DisplayName,
		Version:     u.Version,
		CreatedAt:   u.CreatedAt.Time,
		UpdatedAt:   u.UpdatedAt.Time,
		DeletedAt:   deletedAt,
	}, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.s.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		t := u.DeletedAt.Time
		deletedAt = &t
	}

	return &domain.User{
		ID:           uuid.UUID(u.ID.Bytes),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		BaseCurrency: u.BaseCurrency,
		Status:       u.Status,
		EmailVerifiedAt: func() *time.Time {
			if u.EmailVerifiedAt.Valid {
				t := u.EmailVerifiedAt.Time
				return &t
			}
			return nil
		}(),
		DisplayName: u.DisplayName,
		Version:     u.Version,
		CreatedAt:   u.CreatedAt.Time,
		UpdatedAt:   u.UpdatedAt.Time,
		DeletedAt:   deletedAt,
	}, nil
}

func (r *UserRepo) ActivateAfterOTP(ctx context.Context, userID uuid.UUID) error {
	return r.s.Queries.UserActivateAfterOTP(ctx, pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
}
