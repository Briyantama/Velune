package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
)

type UserRepo struct{ s *Store }

func NewUserRepo(s *Store) repository.UserRepository {
	return &UserRepo{s: s}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
INSERT INTO users (id, email, password_hash, base_currency, version, created_at, updated_at)
VALUES ($1, lower($2), $3, $4, $5, $6, $7)`
	_, err := r.s.Pool.Exec(ctx, q,
		u.ID, u.Email, u.PasswordHash, u.BaseCurrency, u.Version, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, base_currency, version, created_at, updated_at, deleted_at
FROM users WHERE id = $1 AND deleted_at IS NULL`
	row := r.s.Pool.QueryRow(ctx, q, id)
	return scanUser(row)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, base_currency, version, created_at, updated_at, deleted_at
FROM users WHERE lower(email) = lower($1) AND deleted_at IS NULL`
	row := r.s.Pool.QueryRow(ctx, q, email)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.BaseCurrency, &u.Version,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
