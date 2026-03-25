package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
	"github.com/moon-eye/velune/services/transaction-service/internal/repository"
	"github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/pagination"
	"github.com/moon-eye/velune/shared/stringx"
)

type AccountService struct {
	Accounts repository.AccountRepository
}

type CreateAccountInput struct {
	Name     string `validate:"required,min=1,max=200"`
	Type     string `validate:"required,oneof=wallet bank e_money cash card"`
	Currency string `validate:"required,len=3"`
}

func (s *AccountService) Create(ctx context.Context, userID uuid.UUID, in CreateAccountInput) (*domain.Account, error) {
	now := time.Now().UTC()
	a := &domain.Account{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         stringx.TrimSpace(in.Name),
		Type:         domain.AccountType(in.Type),
		Currency:     stringx.Upper(in.Currency),
		BalanceMinor: 0,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.Accounts.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AccountService) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error) {
	a, err := s.Accounts.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errs.ErrNotFound
	}
	return a, nil
}

func (s *AccountService) List(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Account, int64, error) {
	p := pagination.Normalize(page, limit)
	return s.Accounts.List(ctx, userID, p.Limit, p.Offset)
}

type UpdateAccountInput struct {
	Name string `validate:"required,min=1,max=200"`
	Type string `validate:"required,oneof=wallet bank e_money cash card"`
}

func (s *AccountService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, version int64, in UpdateAccountInput) (*domain.Account, error) {
	a, err := s.Accounts.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errs.ErrNotFound
	}
	now := time.Now().UTC()
	a.Name = stringx.TrimSpace(in.Name)
	a.Type = domain.AccountType(in.Type)
	a.Version = version + 1
	a.UpdatedAt = now
	if err := s.Accounts.Update(ctx, a); err != nil {
		if errors.Is(err, repository.ErrOptimisticLock) {
			return nil, errs.New("CONFLICT", "version conflict", constx.StatusConflict)
		}
		return nil, err
	}
	return a, nil
}

func (s *AccountService) Delete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	err := s.Accounts.SoftDelete(ctx, userID, id, version)
	if errors.Is(err, repository.ErrOptimisticLock) {
		return errs.New("CONFLICT", "version conflict", constx.StatusConflict)
	}
	return err
}
