package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
	"github.com/moon-eye/velune/services/transaction-service/internal/repository"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/pagination"
	"github.com/moon-eye/velune/shared/stringx"
)

type RecurringService struct {
	Recurring  repository.RecurringRepository
	Accounts   repository.AccountRepository
	Categories repository.CategoryRepository
}

type CreateRecurringInput struct {
	AccountID   uuid.UUID `validate:"required"`
	CategoryID  *uuid.UUID
	AmountMinor int64     `validate:"min=1"`
	Currency    string    `validate:"required,len=3"`
	Type        string    `validate:"required,oneof=income expense"`
	Frequency   string    `validate:"required,oneof=daily weekly monthly yearly"`
	NextRunAt   time.Time `validate:"required"`
	Description string    `validate:"max=2000"`
}

func (s *RecurringService) Create(ctx context.Context, userID uuid.UUID, in CreateRecurringInput) (*domain.RecurringRule, error) {
	acct, err := s.Accounts.GetByID(ctx, userID, in.AccountID)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, errs.New("NOT_FOUND", "account not found", constx.StatusNotFound)
	}
	if in.CategoryID != nil {
		c, err := s.Categories.GetByID(ctx, userID, *in.CategoryID)
		if err != nil {
			return nil, err
		}
		if c == nil {
			return nil, errs.New("NOT_FOUND", "category not found", constx.StatusNotFound)
		}
	}
	now := time.Now().UTC()
	rr := &domain.RecurringRule{
		ID:          uuid.New(),
		UserID:      userID,
		AccountID:   in.AccountID,
		CategoryID:  in.CategoryID,
		AmountMinor: in.AmountMinor,
		Currency:    stringx.Upper(in.Currency),
		Type:        domain.TransactionType(in.Type),
		Frequency:   domain.RecurringFrequency(in.Frequency),
		NextRunAt:   in.NextRunAt.UTC(),
		IsActive:    true,
		Description: stringx.TrimSpace(in.Description),
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.Recurring.Create(ctx, rr); err != nil {
		return nil, err
	}
	return rr, nil
}

func (s *RecurringService) List(ctx context.Context, userID uuid.UUID, page, limit int, activeOnly bool) ([]domain.RecurringRule, int64, error) {
	p := pagination.Normalize(page, limit)
	return s.Recurring.List(ctx, userID, p.Limit, p.Offset, activeOnly)
}

func (s *RecurringService) Delete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	err := s.Recurring.SoftDelete(ctx, userID, id, version)
	if errors.Is(err, repository.ErrOptimisticLock) {
		return errs.New("CONFLICT", "version conflict", constx.StatusConflict)
	}
	return err
}
