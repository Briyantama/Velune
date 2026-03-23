package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
)

type mockTxClient struct {
	summary       *contracts.TransactionSummary
	byCategory    *contracts.TransactionCategoryTotalsResponse
	summaryErr    error
	byCategoryErr error
}

func (m *mockTxClient) Summary(_ context.Context, _ string, _ contracts.TransactionAnalyticsQuery) (*contracts.TransactionSummary, error) {
	if m.summaryErr != nil {
		return nil, m.summaryErr
	}
	return m.summary, nil
}

func (m *mockTxClient) SummaryByCategory(_ context.Context, _ string, _ contracts.TransactionAnalyticsQuery) (*contracts.TransactionCategoryTotalsResponse, error) {
	if m.byCategoryErr != nil {
		return nil, m.byCategoryErr
	}
	return m.byCategory, nil
}

func TestMonthly_Success(t *testing.T) {
	cid := uuid.New()
	s := &ReportService{
		Transactions: &mockTxClient{
			summary: &contracts.TransactionSummary{IncomeMinor: 20000, ExpenseMinor: 5000},
			byCategory: &contracts.TransactionCategoryTotalsResponse{
				Breakdown: []contracts.TransactionCategorySummary{
					{CategoryID: &cid, TotalMinor: 5000},
				},
			},
		},
	}
	out, err := s.Monthly(context.Background(), uuid.New(), MonthlyInput{Year: 2026, Month: 3, Currency: "USD"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.IncomeMinor != 20000 || out.ExpenseMinor != 5000 {
		t.Fatalf("unexpected totals: %+v", out)
	}
	if len(out.ByCategory) != 1 || out.ByCategory[0].TotalMinor != 5000 {
		t.Fatalf("unexpected category breakdown: %+v", out.ByCategory)
	}
	if out.GeneratedAt.Before(time.Now().Add(-5 * time.Second)) {
		t.Fatalf("generatedAt too old: %v", out.GeneratedAt)
	}
}

func TestMonthly_ValidationError(t *testing.T) {
	s := &ReportService{Transactions: &mockTxClient{}}
	_, err := s.Monthly(context.Background(), uuid.New(), MonthlyInput{Year: 2026, Month: 13})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestMonthly_UpstreamError(t *testing.T) {
	s := &ReportService{
		Transactions: &mockTxClient{
			summaryErr: errs.New("UPSTREAM_ERROR", "boom", 502),
		},
	}
	_, err := s.Monthly(context.Background(), uuid.New(), MonthlyInput{Year: 2026, Month: 3})
	if err == nil {
		t.Fatal("expected upstream error")
	}
}
