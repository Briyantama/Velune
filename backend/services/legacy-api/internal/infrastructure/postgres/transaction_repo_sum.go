package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (r *TransactionRepo) SumIncomeExpenseInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (incomeMinor, expenseMinor int64, err error) {
	const q = `
SELECT
  COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0)
FROM transactions
WHERE user_id = $1 AND deleted_at IS NULL AND currency = $2
  AND occurred_at >= $3 AND occurred_at < $4`
	err = r.s.Pool.QueryRow(ctx, q, userID, currency, from, to).Scan(&incomeMinor, &expenseMinor)
	return incomeMinor, expenseMinor, err
}
