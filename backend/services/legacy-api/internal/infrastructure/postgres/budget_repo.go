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

type BudgetRepo struct{ s *Store }

func NewBudgetRepo(s *Store) repository.BudgetRepository {
	return &BudgetRepo{s: s}
}

func (r *BudgetRepo) Create(ctx context.Context, b *domain.Budget) error {
	const q = `
INSERT INTO budgets (id, user_id, name, period_type, category_id, start_date, end_date, limit_amount_minor, currency, version, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6::date,$7::date,$8,$9,$10,$11,$12)`
	_, err := r.s.Pool.Exec(ctx, q,
		b.ID, b.UserID, b.Name, string(b.PeriodType), b.CategoryID,
		b.StartDate, b.EndDate, b.LimitAmountMinor, b.Currency, b.Version,
		b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *BudgetRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Budget, error) {
	const q = `
SELECT id, user_id, name, period_type, category_id, start_date, end_date, limit_amount_minor, currency, version, created_at, updated_at, deleted_at
FROM budgets WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	row := r.s.Pool.QueryRow(ctx, q, id, userID)
	return scanBudget(row)
}

func (r *BudgetRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int, activeOn *time.Time) ([]domain.Budget, int64, error) {
	if activeOn == nil {
		const countQ = `SELECT COUNT(*) FROM budgets WHERE user_id = $1 AND deleted_at IS NULL`
		var total int64
		if err := r.s.Pool.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
			return nil, 0, err
		}
		const q = `
SELECT id, user_id, name, period_type, category_id, start_date, end_date, limit_amount_minor, currency, version, created_at, updated_at, deleted_at
FROM budgets WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY start_date DESC
LIMIT $2 OFFSET $3`
		rows, err := r.s.Pool.Query(ctx, q, userID, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()
		return scanBudgetRows(rows, total)
	}

	const countQ = `
SELECT COUNT(*) FROM budgets WHERE user_id = $1 AND deleted_at IS NULL
AND start_date <= $2::date AND end_date >= $2::date`
	var total int64
	if err := r.s.Pool.QueryRow(ctx, countQ, userID, *activeOn).Scan(&total); err != nil {
		return nil, 0, err
	}
	const q = `
SELECT id, user_id, name, period_type, category_id, start_date, end_date, limit_amount_minor, currency, version, created_at, updated_at, deleted_at
FROM budgets WHERE user_id = $1 AND deleted_at IS NULL
AND start_date <= $2::date AND end_date >= $2::date
ORDER BY start_date DESC
LIMIT $3 OFFSET $4`
	rows, err := r.s.Pool.Query(ctx, q, userID, *activeOn, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanBudgetRows(rows, total)
}

func scanBudgetRows(rows pgx.Rows, total int64) ([]domain.Budget, int64, error) {
	var out []domain.Budget
	for rows.Next() {
		b, err := scanBudget(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *b)
	}
	return out, total, rows.Err()
}

func scanBudget(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Budget, error) {
	var b domain.Budget
	var pt string
	err := row.Scan(
		&b.ID, &b.UserID, &b.Name, &pt, &b.CategoryID, &b.StartDate, &b.EndDate,
		&b.LimitAmountMinor, &b.Currency, &b.Version, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b.PeriodType = domain.BudgetPeriod(pt)
	return &b, nil
}

func (r *BudgetRepo) Update(ctx context.Context, b *domain.Budget) error {
	const q = `
UPDATE budgets SET name = $1, period_type = $2, category_id = $3, start_date = $4::date, end_date = $5::date,
  limit_amount_minor = $6, currency = $7, version = $8, updated_at = $9
WHERE id = $10 AND user_id = $11 AND version = $12 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q,
		b.Name, string(b.PeriodType), b.CategoryID, b.StartDate, b.EndDate,
		b.LimitAmountMinor, b.Currency, b.Version, b.UpdatedAt, b.ID, b.UserID, b.Version-1,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *BudgetRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	const q = `
UPDATE budgets SET deleted_at = now(), version = version + 1, updated_at = now()
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
