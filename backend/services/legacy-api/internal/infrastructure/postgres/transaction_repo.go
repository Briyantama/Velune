package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
)

type TransactionRepo struct{ s *Store }

func NewTransactionRepo(s *Store) repository.TransactionRepository {
	return &TransactionRepo{s: s}
}

func (r *TransactionRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error) {
	const q = `
SELECT id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type,
       description, occurred_at, version, created_at, updated_at, deleted_at
FROM transactions WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	row := r.s.Pool.QueryRow(ctx, q, id, userID)
	return scanTransaction(row)
}

func (r *TransactionRepo) List(ctx context.Context, userID uuid.UUID, f repository.TransactionFilter, limit, offset int) ([]domain.Transaction, int64, error) {
	var conds []string
	var args []interface{}
	args = append(args, userID)
	argPos := 2
	conds = append(conds, "user_id = $1", "deleted_at IS NULL")

	if f.AccountID != nil {
		conds = append(conds, fmt.Sprintf("account_id = $%d", argPos))
		args = append(args, *f.AccountID)
		argPos++
	}
	if f.CategoryID != nil {
		conds = append(conds, fmt.Sprintf("category_id = $%d", argPos))
		args = append(args, *f.CategoryID)
		argPos++
	}
	if f.Type != nil {
		conds = append(conds, fmt.Sprintf("type = $%d", argPos))
		args = append(args, string(*f.Type))
		argPos++
	}
	if f.From != nil {
		conds = append(conds, fmt.Sprintf("occurred_at >= $%d", argPos))
		args = append(args, *f.From)
		argPos++
	}
	if f.To != nil {
		conds = append(conds, fmt.Sprintf("occurred_at < $%d", argPos))
		args = append(args, *f.To)
		argPos++
	}
	if f.Currency != "" {
		conds = append(conds, fmt.Sprintf("currency = $%d", argPos))
		args = append(args, f.Currency)
		argPos++
	}

	where := strings.Join(conds, " AND ")
	countQ := "SELECT COUNT(*) FROM transactions WHERE " + where
	var total int64
	if err := r.s.Pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := argPos
	offsetArg := argPos + 1
	args = append(args, limit, offset)

	q := fmt.Sprintf(`
SELECT id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type,
       description, occurred_at, version, created_at, updated_at, deleted_at
FROM transactions WHERE %s
ORDER BY occurred_at DESC, id DESC
LIMIT $%d OFFSET $%d`, where, limitArg, offsetArg)

	rows, err := r.s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Transaction
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *t)
	}
	return out, total, rows.Err()
}

func scanTransaction(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Transaction, error) {
	var t domain.Transaction
	var typ string
	err := row.Scan(
		&t.ID, &t.UserID, &t.AccountID, &t.CategoryID, &t.CounterpartyAccountID, &t.AmountMinor, &t.Currency,
		&typ, &t.Description, &t.OccurredAt, &t.Version, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Type = domain.TransactionType(typ)
	return &t, nil
}

func (r *TransactionRepo) SumExpensesByCategoryInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error) {
	const q = `
SELECT category_id, COALESCE(SUM(amount_minor), 0)
FROM transactions
WHERE user_id = $1 AND deleted_at IS NULL AND type = 'expense'
  AND occurred_at >= $2 AND occurred_at < $3 AND currency = $4
  AND category_id IS NOT NULL
GROUP BY category_id`
	rows, err := r.s.Pool.Query(ctx, q, userID, from, to, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]int64)
	for rows.Next() {
		var cid uuid.UUID
		var sum int64
		if err := rows.Scan(&cid, &sum); err != nil {
			return nil, err
		}
		out[cid] = sum
	}
	return out, rows.Err()
}
