package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
	"go.uber.org/zap"
)

type ReportService struct {
	Transactions repository.TransactionRepository
	Budgets      repository.BudgetRepository
	Categories   repository.CategoryRepository
	Logger       *zap.Logger
}

// MonthlyInput defines calendar month boundaries in UTC.
type MonthlyInput struct {
	Year     int
	Month    int
	Currency string
}

func (s *ReportService) Monthly(ctx context.Context, userID uuid.UUID, in MonthlyInput) (*domain.MonthlyReport, error) {
	from := time.Date(in.Year, time.Month(in.Month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	cur := in.Currency
	if cur == "" {
		cur = "USD"
	}
	income, expense, err := s.Transactions.SumIncomeExpenseInRange(ctx, userID, from, to, cur)
	if err != nil {
		return nil, err
	}
	byCat, err := s.Transactions.SumExpensesByCategoryInRange(ctx, userID, from, to, cur)
	if err != nil {
		return nil, err
	}
	var breakdown []domain.MonthlyCategoryBreakdown
	for cid, total := range byCat {
		name := "Uncategorized"
		cat, err := s.Categories.GetByID(ctx, userID, cid)
		if err != nil {
			return nil, err
		}
		if cat != nil {
			name = cat.Name
		}
		cidCopy := cid
		breakdown = append(breakdown, domain.MonthlyCategoryBreakdown{
			CategoryID:   &cidCopy,
			CategoryName: name,
			TotalMinor:   total,
			Currency:     cur,
		})
	}
	rep := &domain.MonthlyReport{
		UserID:       userID,
		Year:         in.Year,
		Month:        in.Month,
		IncomeMinor:  income,
		ExpenseMinor: expense,
		Currency:     cur,
		ByCategory:   breakdown,
		GeneratedAt:  time.Now().UTC(),
	}
	s.checkBudgets(ctx, userID, from, to, cur, expense)
	return rep, nil
}

func (s *ReportService) checkBudgets(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string, totalExpense int64) {
	mid := from.AddDate(0, 0, 15)
	budgets, _, err := s.Budgets.List(ctx, userID, 100, 0, &mid)
	if err != nil || len(budgets) == 0 {
		return
	}
	spendByCat, err := s.Transactions.SumExpensesByCategoryInRange(ctx, userID, from, to, currency)
	if err != nil {
		return
	}
	for _, b := range budgets {
		if b.Currency != currency {
			continue
		}
		var spent int64
		if b.CategoryID == nil {
			spent = totalExpense
		} else {
			spent = spendByCat[*b.CategoryID]
		}
		if spent > b.LimitAmountMinor {
			s.Logger.Warn("budget_exceeded",
				zap.String("budget_id", b.ID.String()),
				zap.String("user_id", userID.String()),
				zap.Int64("spent_minor", spent),
				zap.Int64("limit_minor", b.LimitAmountMinor),
			)
		}
	}
}
