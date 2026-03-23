package usecase

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/pagination"
	"go.uber.org/zap"
)

type TransactionService struct {
	Ledger        repository.Ledger
	Transactions  repository.TransactionRepository
	Accounts      repository.AccountRepository
	Categories    repository.CategoryRepository
	Logger        *zap.Logger
}

type CreateTransactionInput struct {
	AccountID               uuid.UUID  `validate:"required"`
	CategoryID              *uuid.UUID
	CounterpartyAccountID   *uuid.UUID
	AmountMinor             int64      `validate:"required"`
	Currency                string     `validate:"required,len=3"`
	Type                    string     `validate:"required,oneof=income expense transfer adjustment"`
	Description             string     `validate:"max=2000"`
	OccurredAt              time.Time  `validate:"required"`
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
		Currency:              strings.ToUpper(in.Currency),
		Type:                  tt,
		Description:           strings.TrimSpace(in.Description),
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
			return nil, errs.New("NOT_FOUND", "account not found", http.StatusNotFound)
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
			return errs.New("VALIDATION_ERROR", "amount must be positive", http.StatusBadRequest)
		}
	case domain.TransactionTransfer:
		if in.AmountMinor <= 0 || in.CounterpartyAccountID == nil {
			return errs.New("VALIDATION_ERROR", "transfer requires positive amount and counterparty account", http.StatusBadRequest)
		}
		if *in.CounterpartyAccountID == in.AccountID {
			return errs.New("VALIDATION_ERROR", "cannot transfer to the same account", http.StatusBadRequest)
		}
	case domain.TransactionAdjustment:
		if in.AmountMinor == 0 {
			return errs.New("VALIDATION_ERROR", "adjustment amount cannot be zero", http.StatusBadRequest)
		}
	default:
		return errs.New("VALIDATION_ERROR", "invalid transaction type", http.StatusBadRequest)
	}
	if in.CategoryID != nil {
		cat, err := s.Categories.GetByID(ctx, userID, *in.CategoryID)
		if err != nil {
			return err
		}
		if cat == nil {
			return errs.New("NOT_FOUND", "category not found", http.StatusNotFound)
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

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, in ListTransactionsInput) ([]domain.Transaction, int64, error) {
	p := pagination.Normalize(in.Page, in.Limit)
	f := repository.TransactionFilter{Currency: strings.ToUpper(in.Currency)}
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
		return errs.New("CONFLICT", "version conflict", http.StatusConflict)
	}
	return err
}
