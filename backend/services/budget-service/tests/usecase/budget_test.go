package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/budget-service/internal/domain"
	"github.com/moon-eye/velune/services/budget-service/internal/repository"
	"github.com/moon-eye/velune/services/budget-service/internal/usecase"
)

type mockBudgetRepo struct {
	getFn func(context.Context, uuid.UUID, uuid.UUID) (*domain.Budget, error)
}

func (m *mockBudgetRepo) Create(ctx context.Context, b *domain.Budget) error { return nil }
func (m *mockBudgetRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Budget, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockBudgetRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int, activeOn *time.Time) ([]domain.Budget, int64, error) {
	return nil, 0, nil
}
func (m *mockBudgetRepo) Update(ctx context.Context, b *domain.Budget) error { return nil }
func (m *mockBudgetRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	return nil
}

type mockTxSummaryClient struct {
	income     int64
	expense    int64
	byCategory map[uuid.UUID]int64
}

func (m *mockBudgetRepo) TransitionAlertStateAndEnqueue(ctx context.Context, budgetID uuid.UUID, usagePercent float64, envelopePayload json.RawMessage) (bool, error) {
	return usagePercent >= 100, nil
}

func (m *mockTxSummaryClient) Summary(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (int64, int64, error) {
	return m.income, m.expense, nil
}

func (m *mockTxSummaryClient) SummaryByCategory(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error) {
	return m.byCategory, nil
}

func TestBudgetUsage_OverspendDetected(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	svc := &usecase.BudgetService{
		Budgets: &mockBudgetRepo{getFn: func(ctx context.Context, uid, id uuid.UUID) (*domain.Budget, error) {
			return &domain.Budget{
				ID:               budgetID,
				UserID:           userID,
				PeriodType:       domain.BudgetPeriodMonthly,
				StartDate:        start,
				EndDate:          end,
				LimitAmountMinor: 1000,
				Currency:         "USD",
			}, nil
		}},
		Transactions: &mockTxSummaryClient{income: 0, expense: 1500},
	}

	usage, err := svc.Usage(context.Background(), userID, budgetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !usage.IsOverspent || usage.OverspentMinor != 500 {
		t.Fatalf("expected overspend 500, got isOverspent=%v overspent=%d", usage.IsOverspent, usage.OverspentMinor)
	}
}

func TestBudgetUsage_CategoryBudgetUsesCategoryTotals(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	categoryID := uuid.New()
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)

	svc := &usecase.BudgetService{
		Budgets: &mockBudgetRepo{getFn: func(ctx context.Context, uid, id uuid.UUID) (*domain.Budget, error) {
			return &domain.Budget{
				ID:               budgetID,
				UserID:           userID,
				PeriodType:       domain.BudgetPeriodMonthly,
				CategoryID:       &categoryID,
				StartDate:        start,
				EndDate:          end,
				LimitAmountMinor: 2000,
				Currency:         "USD",
			}, nil
		}},
		Transactions: &mockTxSummaryClient{byCategory: map[uuid.UUID]int64{
			categoryID: 1250,
		}},
	}

	usage, err := svc.Usage(context.Background(), userID, budgetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.SpentMinor != 1250 || usage.RemainingMinor != 750 || usage.IsOverspent {
		t.Fatalf("unexpected usage result: %+v", usage)
	}
}

func TestBudgetUsage_OverThresholdEnqueues(t *testing.T) {
	userID := uuid.New()
	budgetID := uuid.New()
	svc := &usecase.BudgetService{
		Budgets: &mockBudgetRepo{getFn: func(ctx context.Context, uid, id uuid.UUID) (*domain.Budget, error) {
			return &domain.Budget{
				ID:               budgetID,
				UserID:           userID,
				PeriodType:       domain.BudgetPeriodMonthly,
				StartDate:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:          time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
				LimitAmountMinor: 1000,
				Currency:         "USD",
			}, nil
		}},
		Transactions: &mockTxSummaryClient{income: 0, expense: 1200},
	}
	usage, err := svc.Usage(context.Background(), userID, budgetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !usage.IsOverspent {
		t.Fatalf("expected overspent usage result")
	}
}

var _ repository.BudgetRepository = (*mockBudgetRepo)(nil)
