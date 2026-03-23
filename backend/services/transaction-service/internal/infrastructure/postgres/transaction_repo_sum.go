package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

func (r *TransactionRepo) SumIncomeExpenseInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (incomeMinor, expenseMinor int64, err error) {
	row, err := r.s.Queries.TransactionSumIncomeExpenseInRange(ctx, db.TransactionSumIncomeExpenseInRangeParams{
		UserID:       helper.ToPgUUID(userID),
		Currency:     currency,
		OccurredAt:   helper.ToPgTS(from),
		OccurredAt_2: helper.ToPgTS(to),
	})
	if err != nil {
		return 0, 0, err
	}
	income, err := helper.ToInt64(row.IncomeMinor)
	if err != nil {
		return 0, 0, err
	}
	expense, err := helper.ToInt64(row.ExpenseMinor)
	if err != nil {
		return 0, 0, err
	}
	return income, expense, nil
}
