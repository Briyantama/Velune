package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/budget-service/internal/domain"
	"github.com/moon-eye/velune/services/budget-service/internal/repository"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

type BudgetRepo struct{ s *Store }

func NewBudgetRepo(s *Store) repository.BudgetRepository {
	return &BudgetRepo{s: s}
}

func (r *BudgetRepo) Create(ctx context.Context, b *domain.Budget) error {
	return r.s.Queries.BudgetCreate(ctx, db.BudgetCreateParams{
		ID:               helper.ToPgUUID(b.ID),
		UserID:           helper.ToPgUUID(b.UserID),
		Name:             b.Name,
		PeriodType:       string(b.PeriodType),
		CategoryID:       helper.ToPgUUIDPtr(b.CategoryID),
		Column6:          helper.ToPgDatePtr(&b.StartDate),
		Column7:          helper.ToPgDatePtr(&b.EndDate),
		LimitAmountMinor: b.LimitAmountMinor,
		Currency:         b.Currency,
		Version:          b.Version,
		CreatedAt:        helper.ToPgTS(b.CreatedAt),
		UpdatedAt:        helper.ToPgTS(b.UpdatedAt),
	})
}

func (r *BudgetRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Budget, error) {
	row, err := r.s.Queries.BudgetGetByID(ctx, db.BudgetGetByIDParams{
		ID:     helper.ToPgUUID(id),
		UserID: helper.ToPgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return budgetFromModel(row), nil
}

func (r *BudgetRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int, activeOn *time.Time) ([]domain.Budget, int64, error) {
	arg := db.BudgetListParams{
		UserID:      helper.ToPgUUID(userID),
		Limit:       int32(limit),
		Offset:      int32(offset),
		HasActiveOn: activeOn != nil,
		ActiveOn:    helper.ToPgDatePtr(activeOn),
	}
	total, err := r.s.Queries.BudgetCountList(ctx, db.BudgetCountListParams{
		UserID:      arg.UserID,
		HasActiveOn: arg.HasActiveOn,
		ActiveOn:    arg.ActiveOn,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.s.Queries.BudgetList(ctx, arg)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.Budget, 0, len(rows))
	for _, row := range rows {
		out = append(out, *budgetFromModel(row))
	}
	return out, total, nil
}

func (r *BudgetRepo) Update(ctx context.Context, b *domain.Budget) error {
	tag, err := r.s.Queries.BudgetUpdate(ctx, db.BudgetUpdateParams{
		Name:             b.Name,
		PeriodType:       string(b.PeriodType),
		CategoryID:       helper.ToPgUUIDPtr(b.CategoryID),
		Column4:          helper.ToPgDatePtr(&b.StartDate),
		Column5:          helper.ToPgDatePtr(&b.EndDate),
		LimitAmountMinor: b.LimitAmountMinor,
		Currency:         b.Currency,
		Version:          b.Version,
		UpdatedAt:        helper.ToPgTS(b.UpdatedAt),
		ID:               helper.ToPgUUID(b.ID),
		UserID:           helper.ToPgUUID(b.UserID),
		Version_2:        b.Version - 1,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *BudgetRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	tag, err := r.s.Queries.BudgetSoftDelete(ctx, db.BudgetSoftDeleteParams{
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

func budgetFromModel(m db.Budget) *domain.Budget {
	return &domain.Budget{
		ID:               uuid.UUID(m.ID.Bytes),
		UserID:           uuid.UUID(m.UserID.Bytes),
		Name:             m.Name,
		PeriodType:       domain.BudgetPeriod(m.PeriodType),
		CategoryID:       helper.FromPgUUIDPtr(m.CategoryID),
		StartDate:        m.StartDate.Time,
		EndDate:          m.EndDate.Time,
		LimitAmountMinor: m.LimitAmountMinor,
		Currency:         m.Currency,
		Version:          m.Version,
		CreatedAt:        m.CreatedAt.Time,
		UpdatedAt:        m.UpdatedAt.Time,
		DeletedAt:        helper.FromPgTSPtr(m.DeletedAt),
	}
}
