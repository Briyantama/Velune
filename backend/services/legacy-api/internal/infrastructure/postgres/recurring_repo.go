package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
)

type RecurringRepo struct{ s *Store }

func NewRecurringRepo(s *Store) repository.RecurringRepository {
	return &RecurringRepo{s: s}
}

func (r *RecurringRepo) Create(ctx context.Context, rr *domain.RecurringRule) error {
	const q = `
INSERT INTO recurring_rules (id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
	_, err := r.s.Pool.Exec(ctx, q,
		rr.ID, rr.UserID, rr.AccountID, rr.CategoryID, rr.AmountMinor, rr.Currency,
		string(rr.Type), string(rr.Frequency), rr.NextRunAt, rr.LastRunAt, rr.IsActive, rr.Description,
		rr.Version, rr.CreatedAt, rr.UpdatedAt,
	)
	return err
}

func (r *RecurringRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.RecurringRule, error) {
	const q = `
SELECT id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at, deleted_at
FROM recurring_rules WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	row := r.s.Pool.QueryRow(ctx, q, id, userID)
	return scanRecurring(row)
}

func (r *RecurringRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int, activeOnly bool) ([]domain.RecurringRule, int64, error) {
	var countQ, q string
	var args []interface{}
	if activeOnly {
		countQ = `SELECT COUNT(*) FROM recurring_rules WHERE user_id = $1 AND deleted_at IS NULL AND is_active = true`
		q = `
SELECT id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at, deleted_at
FROM recurring_rules WHERE user_id = $1 AND deleted_at IS NULL AND is_active = true
ORDER BY next_run_at ASC
LIMIT $2 OFFSET $3`
		args = []interface{}{userID, limit, offset}
	} else {
		countQ = `SELECT COUNT(*) FROM recurring_rules WHERE user_id = $1 AND deleted_at IS NULL`
		q = `
SELECT id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at, deleted_at
FROM recurring_rules WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY next_run_at ASC
LIMIT $2 OFFSET $3`
		args = []interface{}{userID, limit, offset}
	}
	var total int64
	if err := r.s.Pool.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.RecurringRule
	for rows.Next() {
		rr, err := scanRecurring(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *rr)
	}
	return out, total, rows.Err()
}

func (r *RecurringRepo) ListDue(ctx context.Context, before time.Time, limit int) ([]domain.RecurringRule, error) {
	const q = `
SELECT id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at, deleted_at
FROM recurring_rules
WHERE deleted_at IS NULL AND is_active = true AND next_run_at <= $1
ORDER BY next_run_at ASC
LIMIT $2`
	rows, err := r.s.Pool.Query(ctx, q, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RecurringRule
	for rows.Next() {
		rr, err := scanRecurring(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rr)
	}
	return out, rows.Err()
}

func scanRecurring(row interface {
	Scan(dest ...interface{}) error
}) (*domain.RecurringRule, error) {
	var rr domain.RecurringRule
	var typ, freq string
	err := row.Scan(
		&rr.ID, &rr.UserID, &rr.AccountID, &rr.CategoryID, &rr.AmountMinor, &rr.Currency,
		&typ, &freq, &rr.NextRunAt, &rr.LastRunAt, &rr.IsActive, &rr.Description,
		&rr.Version, &rr.CreatedAt, &rr.UpdatedAt, &rr.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rr.Type = domain.TransactionType(typ)
	rr.Frequency = domain.RecurringFrequency(freq)
	return &rr, nil
}

func (r *RecurringRepo) Update(ctx context.Context, rr *domain.RecurringRule) error {
	const q = `
UPDATE recurring_rules SET account_id = $1, category_id = $2, amount_minor = $3, currency = $4, type = $5, frequency = $6,
  next_run_at = $7, last_run_at = $8, is_active = $9, description = $10, version = $11, updated_at = $12
WHERE id = $13 AND user_id = $14 AND version = $15 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q,
		rr.AccountID, rr.CategoryID, rr.AmountMinor, rr.Currency, string(rr.Type), string(rr.Frequency),
		rr.NextRunAt, rr.LastRunAt, rr.IsActive, rr.Description, rr.Version, rr.UpdatedAt,
		rr.ID, rr.UserID, rr.Version-1,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *RecurringRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	const q = `
UPDATE recurring_rules SET deleted_at = now(), version = version + 1, updated_at = now()
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
