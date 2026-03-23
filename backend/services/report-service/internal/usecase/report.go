package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/report-service/internal/repository"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
)

type ReportService struct {
	Transactions repository.TransactionAnalyticsRepository
}

type MonthlyInput struct {
	Year     int
	Month    int
	Currency string
}

func (s *ReportService) Monthly(ctx context.Context, userID uuid.UUID, in MonthlyInput) (*contracts.MonthlyReportDTO, error) {
	if in.Year < 1900 || in.Year > 3000 {
		return nil, errs.New("VALIDATION_ERROR", "valid year is required", 400)
	}
	if in.Month < 1 || in.Month > 12 {
		return nil, errs.New("VALIDATION_ERROR", "valid month 1-12 is required", 400)
	}
	from := time.Date(in.Year, time.Month(in.Month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	cur := in.Currency
	if cur == "" {
		cur = "USD"
	}

	q := contracts.TransactionAnalyticsQuery{From: from, To: to, Currency: cur}
	sum, err := s.Transactions.Summary(ctx, userID.String(), q)
	if err != nil {
		return nil, err
	}
	byCat, err := s.Transactions.SummaryByCategory(ctx, userID.String(), q)
	if err != nil {
		return nil, err
	}

	breakdown := make([]contracts.MonthlyCategoryBreakdownDTO, 0, len(byCat.Breakdown))
	for _, item := range byCat.Breakdown {
		name := "Uncategorized"
		if item.CategoryID != nil {
			name = item.CategoryID.String()
		}
		breakdown = append(breakdown, contracts.MonthlyCategoryBreakdownDTO{
			CategoryID:   item.CategoryID,
			CategoryName: name,
			TotalMinor:   item.TotalMinor,
			Currency:     cur,
		})
	}

	return &contracts.MonthlyReportDTO{
		UserID:       userID,
		Year:         in.Year,
		Month:        in.Month,
		IncomeMinor:  sum.IncomeMinor,
		ExpenseMinor: sum.ExpenseMinor,
		Currency:     cur,
		ByCategory:   breakdown,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}
