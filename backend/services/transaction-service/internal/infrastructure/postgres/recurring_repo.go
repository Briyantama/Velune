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

type RecurringRepo struct{ s *Store }

func NewRecurringRepo(s *Store) repository.RecurringRepository {
	return &RecurringRepo{s: s}
}

func (r *RecurringRepo) Create(ctx context.Context, rr *domain.RecurringRule) error {
	return r.s.Queries.RecurringCreate(ctx, db.RecurringCreateParams{
		ID:          helper.ToPgUUID(rr.ID),
		UserID:      helper.ToPgUUID(rr.UserID),
		AccountID:   helper.ToPgUUID(rr.AccountID),
		CategoryID:  helper.ToPgUUIDPtr(rr.CategoryID),
		AmountMinor: rr.AmountMinor,
		Currency:    rr.Currency,
		Type:        string(rr.Type),
		Frequency:   string(rr.Frequency),
		NextRunAt:   helper.ToPgTS(rr.NextRunAt),
		LastRunAt:   helper.ToPgTSPtr(rr.LastRunAt),
		IsActive:    rr.IsActive,
		Description: rr.Description,
		Version:     rr.Version,
		CreatedAt:   helper.ToPgTS(rr.CreatedAt),
		UpdatedAt:   helper.ToPgTS(rr.UpdatedAt),
	})
}

func (r *RecurringRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.RecurringRule, error) {
	row, err := r.s.Queries.RecurringGetByID(ctx, db.RecurringGetByIDParams{
		ID:     helper.ToPgUUID(id),
		UserID: helper.ToPgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return recurringFromModel(row), nil
}

func (r *RecurringRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int, activeOnly bool) ([]domain.RecurringRule, int64, error) {
	total, err := r.s.Queries.RecurringCountList(ctx, db.RecurringCountListParams{
		UserID:     helper.ToPgUUID(userID),
		ActiveOnly: activeOnly,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.s.Queries.RecurringList(ctx, db.RecurringListParams{
		UserID:     helper.ToPgUUID(userID),
		Limit:      int32(limit),
		Offset:     int32(offset),
		ActiveOnly: activeOnly,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.RecurringRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, *recurringFromModel(row))
	}
	return out, total, nil
}

func (r *RecurringRepo) ListDue(ctx context.Context, before time.Time, limit int) ([]domain.RecurringRule, error) {
	rows, err := r.s.Queries.RecurringListDue(ctx, db.RecurringListDueParams{
		NextRunAt: helper.ToPgTS(before),
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.RecurringRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, *recurringFromModel(row))
	}
	return out, nil
}

func (r *RecurringRepo) Update(ctx context.Context, rr *domain.RecurringRule) error {
	tag, err := r.s.Queries.RecurringUpdate(ctx, db.RecurringUpdateParams{
		AccountID:   helper.ToPgUUID(rr.AccountID),
		CategoryID:  helper.ToPgUUIDPtr(rr.CategoryID),
		AmountMinor: rr.AmountMinor,
		Currency:    rr.Currency,
		Type:        string(rr.Type),
		Frequency:   string(rr.Frequency),
		NextRunAt:   helper.ToPgTS(rr.NextRunAt),
		LastRunAt:   helper.ToPgTSPtr(rr.LastRunAt),
		IsActive:    rr.IsActive,
		Description: rr.Description,
		Version:     rr.Version,
		UpdatedAt:   helper.ToPgTS(rr.UpdatedAt),
		ID:          helper.ToPgUUID(rr.ID),
		UserID:      helper.ToPgUUID(rr.UserID),
		Version_2:   rr.Version - 1,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *RecurringRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	tag, err := r.s.Queries.RecurringSoftDelete(ctx, db.RecurringSoftDeleteParams{
		ID:      helper.ToPgUUID(id),
		UserID:  helper.ToPgUUID(userID),
		Version: version,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func recurringFromModel(m db.RecurringRule) *domain.RecurringRule {
	return &domain.RecurringRule{
		ID:          helper.FromPgUUID(m.ID),
		UserID:      helper.FromPgUUID(m.UserID),
		AccountID:   helper.FromPgUUID(m.AccountID),
		CategoryID:  helper.FromPgUUIDPtr(m.CategoryID),
		AmountMinor: m.AmountMinor,
		Currency:    m.Currency,
		Type:        domain.TransactionType(m.Type),
		Frequency:   domain.RecurringFrequency(m.Frequency),
		NextRunAt:   m.NextRunAt.Time,
		LastRunAt:   helper.FromPgTSPtr(m.LastRunAt),
		IsActive:    m.IsActive,
		Description: m.Description,
		Version:     m.Version,
		CreatedAt:   m.CreatedAt.Time,
		UpdatedAt:   m.UpdatedAt.Time,
		DeletedAt:   helper.FromPgTSPtr(m.DeletedAt),
	}
}
