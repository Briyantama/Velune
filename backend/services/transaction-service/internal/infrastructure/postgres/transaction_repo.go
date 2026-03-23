package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
	"github.com/moon-eye/velune/services/transaction-service/internal/repository"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

type TransactionRepo struct{ s *Store }

func NewTransactionRepo(s *Store) repository.TransactionRepository {
	return &TransactionRepo{s: s}
}

func (r *TransactionRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error) {
	row, err := r.s.Queries.TransactionGetByID(ctx, db.TransactionGetByIDParams{
		ID:     helper.ToPgUUID(id),
		UserID: helper.ToPgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return transactionFromGetByIDRow(row), nil
}

func (r *TransactionRepo) List(ctx context.Context, userID uuid.UUID, f repository.TransactionFilter, limit, offset int) ([]domain.Transaction, int64, error) {
	arg := db.TransactionListParams{
		UserID:        helper.ToPgUUID(userID),
		Limit:         int32(limit),
		Offset:        int32(offset),
		HasAccountID:  f.AccountID != nil,
		AccountID:     helper.ToPgUUIDPtr(f.AccountID),
		HasCategoryID: f.CategoryID != nil,
		CategoryID:    helper.ToPgUUIDPtr(f.CategoryID),
		HasType:       f.Type != nil,
		HasFrom:       f.From != nil,
		FromAt:        helper.ToPgTSPtr(f.From),
		HasTo:         f.To != nil,
		ToAt:          helper.ToPgTSPtr(f.To),
		Currency:      f.Currency,
	}
	if f.Type != nil {
		arg.Type = string(*f.Type)
	}

	total, err := r.s.Queries.TransactionCountList(ctx, db.TransactionCountListParams{
		UserID:        arg.UserID,
		HasAccountID:  arg.HasAccountID,
		AccountID:     arg.AccountID,
		HasCategoryID: arg.HasCategoryID,
		CategoryID:    arg.CategoryID,
		HasType:       arg.HasType,
		Type:          arg.Type,
		HasFrom:       arg.HasFrom,
		FromAt:        arg.FromAt,
		HasTo:         arg.HasTo,
		ToAt:          arg.ToAt,
		Currency:      arg.Currency,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.s.Queries.TransactionList(ctx, arg)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, *transactionFromListRow(row))
	}
	return out, total, nil
}

func (r *TransactionRepo) SumExpensesByCategoryInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error) {
	rows, err := r.s.Queries.TransactionSumExpensesByCategoryInRange(ctx, db.TransactionSumExpensesByCategoryInRangeParams{
		UserID:       helper.ToPgUUID(userID),
		OccurredAt:   helper.ToPgTS(from),
		OccurredAt_2: helper.ToPgTS(to),
		Currency:     currency,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]int64)
	for _, row := range rows {
		if !row.CategoryID.Valid {
			continue
		}
		sum, err := helper.ToInt64(row.TotalMinor)
		if err != nil {
			return nil, err
		}
		out[helper.FromPgUUID(row.CategoryID)] = sum
	}
	return out, nil
}

func transactionFromGetByIDRow(r db.TransactionGetByIDRow) *domain.Transaction {
	return &domain.Transaction{
		ID:                    helper.FromPgUUID(r.ID),
		UserID:                helper.FromPgUUID(r.UserID),
		AccountID:             helper.FromPgUUID(r.AccountID),
		CategoryID:            helper.FromPgUUIDPtr(r.CategoryID),
		CounterpartyAccountID: helper.FromPgUUIDPtr(r.CounterpartyAccountID),
		AmountMinor:           r.AmountMinor,
		Currency:              r.Currency,
		Type:                  domain.TransactionType(r.Type),
		Description:           r.Description,
		OccurredAt:            r.OccurredAt.Time,
		Version:               r.Version,
		CreatedAt:             r.CreatedAt.Time,
		UpdatedAt:             r.UpdatedAt.Time,
		DeletedAt:             helper.FromPgTSPtr(r.DeletedAt),
	}
}

func transactionFromListRow(r db.TransactionListRow) *domain.Transaction {
	return &domain.Transaction{
		ID:                    helper.FromPgUUID(r.ID),
		UserID:                helper.FromPgUUID(r.UserID),
		AccountID:             helper.FromPgUUID(r.AccountID),
		CategoryID:            helper.FromPgUUIDPtr(r.CategoryID),
		CounterpartyAccountID: helper.FromPgUUIDPtr(r.CounterpartyAccountID),
		AmountMinor:           r.AmountMinor,
		Currency:              r.Currency,
		Type:                  domain.TransactionType(r.Type),
		Description:           r.Description,
		OccurredAt:            r.OccurredAt.Time,
		Version:               r.Version,
		CreatedAt:             r.CreatedAt.Time,
		UpdatedAt:             r.UpdatedAt.Time,
		DeletedAt:             helper.FromPgTSPtr(r.DeletedAt),
	}
}

