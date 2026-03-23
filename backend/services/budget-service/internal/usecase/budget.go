package usecase

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/budget-service/internal/domain"
	"github.com/moon-eye/velune/services/budget-service/internal/repository"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/pagination"
	"github.com/moon-eye/velune/shared/stringx"
)

type BudgetService struct {
	Budgets       repository.BudgetRepository
	Transactions  TransactionSummaryClient
}

type TransactionSummaryClient interface {
	Summary(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (incomeMinor, expenseMinor int64, err error)
	SummaryByCategory(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error)
}

type CreateBudgetInput struct {
	Name             string     `validate:"required,min=1,max=200"`
	PeriodType       string     `validate:"required,oneof=monthly weekly custom"`
	CategoryID       *uuid.UUID
	StartDate        time.Time  `validate:"required"`
	EndDate          time.Time  `validate:"required"`
	LimitAmountMinor int64      `validate:"min=0"`
	Currency         string     `validate:"required,len=3"`
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, in CreateBudgetInput) (*domain.Budget, error) {
	if in.EndDate.Before(in.StartDate) {
		return nil, errs.New("VALIDATION_ERROR", "end date must be on or after start date", http.StatusBadRequest)
	}
	now := time.Now().UTC()
	b := &domain.Budget{
		ID:               uuid.New(),
		UserID:           userID,
		Name:             stringx.TrimSpace(in.Name),
		PeriodType:       domain.BudgetPeriod(in.PeriodType),
		CategoryID:       in.CategoryID,
		StartDate:        truncateDate(in.StartDate),
		EndDate:          truncateDate(in.EndDate),
		LimitAmountMinor: in.LimitAmountMinor,
		Currency:         stringx.Upper(in.Currency),
		Version:          1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.Budgets.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func truncateDate(t time.Time) time.Time {
	tt := t.UTC()
	return time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID, page, limit int, activeOn *time.Time) ([]domain.Budget, int64, error) {
	p := pagination.Normalize(page, limit)
	return s.Budgets.List(ctx, userID, p.Limit, p.Offset, activeOn)
}

type UpdateBudgetInput struct {
	Name             string    `validate:"required,min=1,max=200"`
	PeriodType       string    `validate:"required,oneof=monthly weekly custom"`
	CategoryID       *uuid.UUID
	StartDate        time.Time `validate:"required"`
	EndDate          time.Time `validate:"required"`
	LimitAmountMinor int64     `validate:"min=0"`
	Currency         string    `validate:"required,len=3"`
}

func (s *BudgetService) Update(ctx context.Context, userID, id uuid.UUID, version int64, in UpdateBudgetInput) (*domain.Budget, error) {
	if in.EndDate.Before(in.StartDate) {
		return nil, errs.New("VALIDATION_ERROR", "end date must be on or after start date", http.StatusBadRequest)
	}
	b, err := s.Budgets.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, errs.ErrNotFound
	}
	now := time.Now().UTC()
	b.Name = stringx.TrimSpace(in.Name)
	b.PeriodType = domain.BudgetPeriod(in.PeriodType)
	b.CategoryID = in.CategoryID
	b.StartDate = truncateDate(in.StartDate)
	b.EndDate = truncateDate(in.EndDate)
	b.LimitAmountMinor = in.LimitAmountMinor
	b.Currency = stringx.Upper(in.Currency)
	b.Version = version + 1
	b.UpdatedAt = now
	if err := s.Budgets.Update(ctx, b); err != nil {
		if errors.Is(err, repository.ErrOptimisticLock) {
			return nil, errs.New("CONFLICT", "version conflict", http.StatusConflict)
		}
		return nil, err
	}
	return b, nil
}

func (s *BudgetService) Delete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	err := s.Budgets.SoftDelete(ctx, userID, id, version)
	if errors.Is(err, repository.ErrOptimisticLock) {
		return errs.New("CONFLICT", "version conflict", http.StatusConflict)
	}
	return err
}

type BudgetUsage struct {
	BudgetID          uuid.UUID `json:"budgetId"`
	From              time.Time `json:"from"`
	To                time.Time `json:"to"`
	Currency          string    `json:"currency"`
	LimitAmountMinor  int64     `json:"limitAmountMinor"`
	SpentMinor        int64     `json:"spentMinor"`
	RemainingMinor    int64     `json:"remainingMinor"`
	OverspentMinor    int64     `json:"overspentMinor"`
	IsOverspent       bool      `json:"isOverspent"`
}

func (s *BudgetService) Usage(ctx context.Context, userID, budgetID uuid.UUID) (*BudgetUsage, error) {
	b, err := s.Budgets.GetByID(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, errs.ErrNotFound
	}
	if s.Transactions == nil {
		return nil, errs.New("UPSTREAM_UNAVAILABLE", "transaction summary client not configured", http.StatusServiceUnavailable)
	}

	var spent int64
	if b.CategoryID == nil {
		_, expense, err := s.Transactions.Summary(ctx, userID, b.StartDate, b.EndDate.Add(24*time.Hour), b.Currency)
		if err != nil {
			return nil, err
		}
		spent = expense
	} else {
		byCat, err := s.Transactions.SummaryByCategory(ctx, userID, b.StartDate, b.EndDate.Add(24*time.Hour), b.Currency)
		if err != nil {
			return nil, err
		}
		spent = byCat[*b.CategoryID]
	}

	remaining := b.LimitAmountMinor - spent
	overspent := int64(0)
	if remaining < 0 {
		overspent = -remaining
	}
	return &BudgetUsage{
		BudgetID:         b.ID,
		From:             b.StartDate,
		To:               b.EndDate,
		Currency:         b.Currency,
		LimitAmountMinor: b.LimitAmountMinor,
		SpentMinor:       spent,
		RemainingMinor:   remaining,
		OverspentMinor:   overspent,
		IsOverspent:      overspent > 0,
	}, nil
}
