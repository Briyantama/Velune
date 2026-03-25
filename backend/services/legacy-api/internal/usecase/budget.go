package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/pagination"
)

type BudgetService struct {
	Budgets      repository.BudgetRepository
	Categories   repository.CategoryRepository
	Transactions repository.TransactionRepository
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
		return nil, errs.New("VALIDATION_ERROR", "end date must be on or after start date",constx.StatusBadRequest)
	}
	if in.CategoryID != nil {
		c, err := s.Categories.GetByID(ctx, userID, *in.CategoryID)
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, errs.New("NOT_FOUND", "category not found",constx.StatusNotFound)
		}
	}
	now := time.Now().UTC()
	b := &domain.Budget{
		ID:               uuid.New(),
		UserID:           userID,
		Name:             strings.TrimSpace(in.Name),
		PeriodType:       domain.BudgetPeriod(in.PeriodType),
		CategoryID:       in.CategoryID,
		StartDate:        truncateDate(in.StartDate),
		EndDate:          truncateDate(in.EndDate),
		LimitAmountMinor: in.LimitAmountMinor,
		Currency:         strings.ToUpper(in.Currency),
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
		return nil, errs.New("VALIDATION_ERROR", "end date must be on or after start date",constx.StatusBadRequest)
	}
	b, err := s.Budgets.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, errs.ErrNotFound
	}
	if in.CategoryID != nil {
		c, err := s.Categories.GetByID(ctx, userID, *in.CategoryID)
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, errs.New("NOT_FOUND", "category not found",constx.StatusNotFound)
		}
	}
	now := time.Now().UTC()
	b.Name = strings.TrimSpace(in.Name)
	b.PeriodType = domain.BudgetPeriod(in.PeriodType)
	b.CategoryID = in.CategoryID
	b.StartDate = truncateDate(in.StartDate)
	b.EndDate = truncateDate(in.EndDate)
	b.LimitAmountMinor = in.LimitAmountMinor
	b.Currency = strings.ToUpper(in.Currency)
	b.Version = version + 1
	b.UpdatedAt = now
	if err := s.Budgets.Update(ctx, b); err != nil {
		if errors.Is(err, repository.ErrOptimisticLock) {
			return nil, errs.New("CONFLICT", "version conflict",constx.StatusConflict)
		}
		return nil, err
	}
	return b, nil
}

func (s *BudgetService) Delete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	err := s.Budgets.SoftDelete(ctx, userID, id, version)
	if errors.Is(err, repository.ErrOptimisticLock) {
		return errs.New("CONFLICT", "version conflict",constx.StatusConflict)
	}
	return err
}
