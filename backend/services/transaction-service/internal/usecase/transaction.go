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
	"go.uber.org/zap"
)

type TransactionService struct {
	Ledger       repository.Ledger
	Transactions repository.TransactionRepository
	Accounts     repository.AccountRepository
	Categories   repository.CategoryRepository
	Logger       *zap.Logger
}

type CreateTransactionInput struct {
	AccountID             uuid.UUID `validate:"required"`
	CategoryID            *uuid.UUID
	CounterpartyAccountID *uuid.UUID
	AmountMinor           int64     `validate:"required"`
	Currency              string    `validate:"required,len=3"`
	Type                  string    `validate:"required,oneof=income expense transfer adjustment"`
	Description           string    `validate:"max=2000"`
	OccurredAt            time.Time `validate:"required"`
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, in CreateTransactionInput) (*domain.Transaction, error) {
	tt := domain.TransactionType(in.Type)
	if err := s.validateCreate(ctx, userID, &in, tt); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	t := &domain.Transaction{
		ID:                    uuid.New(),
		UserID:                userID,
		AccountID:             in.AccountID,
		CategoryID:            in.CategoryID,
		CounterpartyAccountID: in.CounterpartyAccountID,
		AmountMinor:           in.AmountMinor,
		Currency:              stringx.Upper(in.Currency),
		Type:                  tt,
		Description:           stringx.TrimSpace(in.Description),
		OccurredAt:            in.OccurredAt.UTC(),
		Version:               1,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := s.Ledger.CreateTransaction(ctx, t); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return nil, errs.ErrInsufficient
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.New("NOT_FOUND", "account not found", constx.StatusNotFound)
		}
		return nil, err
	}
	s.Logger.Info("transaction_created",
		zap.String("transaction_id", t.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("type", string(t.Type)),
	)
	return t, nil
}

func (s *TransactionService) validateCreate(ctx context.Context, userID uuid.UUID, in *CreateTransactionInput, tt domain.TransactionType) error {
	switch tt {
	case domain.TransactionIncome, domain.TransactionExpense:
		if in.AmountMinor <= 0 {
			return errs.New("VALIDATION_ERROR", "amount must be positive", constx.StatusBadRequest)
		}
	case domain.TransactionTransfer:
		if in.AmountMinor <= 0 || in.CounterpartyAccountID == nil {
			return errs.New("VALIDATION_ERROR", "transfer requires positive amount and counterparty account", constx.StatusBadRequest)
		}
		if *in.CounterpartyAccountID == in.AccountID {
			return errs.New("VALIDATION_ERROR", "cannot transfer to the same account", constx.StatusBadRequest)
		}
	case domain.TransactionAdjustment:
		if in.AmountMinor == 0 {
			return errs.New("VALIDATION_ERROR", "adjustment amount cannot be zero", constx.StatusBadRequest)
		}
	default:
		return errs.New("VALIDATION_ERROR", "invalid transaction type", constx.StatusBadRequest)
	}
	if in.CategoryID != nil {
		cat, err := s.Categories.GetByID(ctx, userID, *in.CategoryID)
		if err != nil {
			return err
		}
		if cat == nil {
			return errs.New("NOT_FOUND", "category not found", constx.StatusNotFound)
		}
	}
	return nil
}

type ListTransactionsInput struct {
	Page       int
	Limit      int
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	Type       *string
	From       *time.Time
	To         *time.Time
	Currency   string
}

type UpdateTransactionInput struct {
	AccountID             uuid.UUID
	CategoryID            *uuid.UUID
	CounterpartyAccountID *uuid.UUID
	AmountMinor           int64
	Currency              string
	Type                  string
	Description           string
	OccurredAt            time.Time
}

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, in ListTransactionsInput) ([]domain.Transaction, int64, error) {
	p := pagination.Normalize(in.Page, in.Limit)
	f := repository.TransactionFilter{Currency: stringx.Upper(in.Currency)}
	if in.AccountID != nil {
		f.AccountID = in.AccountID
	}
	if in.CategoryID != nil {
		f.CategoryID = in.CategoryID
	}
	if in.Type != nil && *in.Type != "" {
		tt := domain.TransactionType(*in.Type)
		f.Type = &tt
	}
	if in.From != nil {
		t := in.From.UTC()
		f.From = &t
	}
	if in.To != nil {
		t := in.To.UTC()
		f.To = &t
	}
	return s.Transactions.List(ctx, userID, f, p.Limit, p.Offset)
}

func (s *TransactionService) Delete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	err := s.Ledger.SoftDeleteTransaction(ctx, userID, id, version)
	if errors.Is(err, repository.ErrNotFound) {
		return errs.ErrNotFound
	}
	if errors.Is(err, repository.ErrOptimisticLock) {
		return errs.New("CONFLICT", "version conflict", constx.StatusConflict)
	}
	return err
}

func (s *TransactionService) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error) {
	t, err := s.Transactions.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errs.ErrNotFound
	}
	return t, nil
}

func (s *TransactionService) Update(ctx context.Context, userID, id uuid.UUID, version int64, in UpdateTransactionInput) (*domain.Transaction, error) {
	tt := domain.TransactionType(in.Type)
	createLike := CreateTransactionInput{
		AccountID:             in.AccountID,
		CategoryID:            in.CategoryID,
		CounterpartyAccountID: in.CounterpartyAccountID,
		AmountMinor:           in.AmountMinor,
		Currency:              in.Currency,
		Type:                  in.Type,
		Description:           in.Description,
		OccurredAt:            in.OccurredAt,
	}
	if err := s.validateCreate(ctx, userID, &createLike, tt); err != nil {
		return nil, err
	}

	current, err := s.Transactions.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errs.ErrNotFound
	}
	if current.Version != version {
		return nil, errs.New("CONFLICT", "version conflict", constx.StatusConflict)
	}

	next := *current
	next.AccountID = in.AccountID
	next.CategoryID = in.CategoryID
	next.CounterpartyAccountID = in.CounterpartyAccountID
	next.AmountMinor = in.AmountMinor
	next.Currency = stringx.Upper(in.Currency)
	next.Type = tt
	next.Description = stringx.TrimSpace(in.Description)
	next.OccurredAt = in.OccurredAt.UTC()
	next.Version = current.Version + 1
	next.UpdatedAt = time.Now().UTC()

	if err := s.Ledger.UpdateTransaction(ctx, userID, &next, version); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return nil, errs.ErrInsufficient
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.ErrNotFound
		}
		if errors.Is(err, repository.ErrOptimisticLock) {
			return nil, errs.New("CONFLICT", "version conflict", constx.StatusConflict)
		}
		return nil, err
	}
	return &next, nil
}

func (s *TransactionService) Summary(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (int64, int64, error) {
	return s.Transactions.SumIncomeExpenseInRange(ctx, userID, from.UTC(), to.UTC(), stringx.Upper(currency))
}

func (s *TransactionService) SummaryByCategory(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error) {
	return s.Transactions.SumExpensesByCategoryInRange(ctx, userID, from.UTC(), to.UTC(), stringx.Upper(currency))
}
