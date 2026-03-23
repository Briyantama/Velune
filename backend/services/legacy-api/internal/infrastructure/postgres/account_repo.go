package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
)

type AccountRepo struct{ s *Store }

func NewAccountRepo(s *Store) repository.AccountRepository {
	return &AccountRepo{s: s}
}

func (r *AccountRepo) Create(ctx context.Context, a *domain.Account) error {
	const q = `
INSERT INTO accounts (id, user_id, name, type, currency, balance_minor, version, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.s.Pool.Exec(ctx, q,
		a.ID, a.UserID, a.Name, string(a.Type), a.Currency, a.BalanceMinor, a.Version,
		a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *AccountRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error) {
	const q = `
SELECT id, user_id, name, type, currency, balance_minor, version, created_at, updated_at, deleted_at
FROM accounts WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	row := r.s.Pool.QueryRow(ctx, q, id, userID)
	return scanAccount(row)
}

func (r *AccountRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Account, int64, error) {
	const countQ = `SELECT COUNT(*) FROM accounts WHERE user_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.s.Pool.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	const q = `
SELECT id, user_id, name, type, currency, balance_minor, version, created_at, updated_at, deleted_at
FROM accounts WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`
	rows, err := r.s.Pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Account
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, rows.Err()
}

func scanAccount(row pgx.Row) (*domain.Account, error) {
	var a domain.Account
	var typ string
	err := row.Scan(
		&a.ID, &a.UserID, &a.Name, &typ, &a.Currency, &a.BalanceMinor, &a.Version,
		&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.Type = domain.AccountType(typ)
	return &a, nil
}

func (r *AccountRepo) Update(ctx context.Context, a *domain.Account) error {
	const q = `
UPDATE accounts SET name = $1, type = $2, version = $3, updated_at = $4
WHERE id = $5 AND user_id = $6 AND version = $7 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q,
		a.Name, string(a.Type), a.Version, a.UpdatedAt, a.ID, a.UserID, a.Version-1,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *AccountRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	const q = `
UPDATE accounts SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q, id, userID, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}
