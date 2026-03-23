package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
	"github.com/moon-eye/velune/services/transaction-service/internal/repository"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

type AccountRepo struct{ s *Store }

func NewAccountRepo(s *Store) repository.AccountRepository {
	return &AccountRepo{s: s}
}

func (r *AccountRepo) Create(ctx context.Context, a *domain.Account) error {
	return r.s.Queries.AccountCreate(ctx, db.AccountCreateParams{
		ID:           helper.ToPgUUID(a.ID),
		UserID:       helper.ToPgUUID(a.UserID),
		Name:         a.Name,
		Type:         string(a.Type),
		Currency:     a.Currency,
		BalanceMinor: a.BalanceMinor,
		Version:      a.Version,
		CreatedAt:    helper.ToPgTS(a.CreatedAt),
		UpdatedAt:    helper.ToPgTS(a.UpdatedAt),
	})
}

func (r *AccountRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error) {
	row, err := r.s.Queries.AccountGetByID(ctx, db.AccountGetByIDParams{
		ID:     helper.ToPgUUID(id),
		UserID: helper.ToPgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return accountFromModel(row), nil
}

func (r *AccountRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Account, int64, error) {
	total, err := r.s.Queries.AccountCountList(ctx, helper.ToPgUUID(userID))
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.s.Queries.AccountList(ctx, db.AccountListParams{
		UserID: helper.ToPgUUID(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		out = append(out, *accountFromModel(row))
	}
	return out, total, nil
}

func (r *AccountRepo) Update(ctx context.Context, a *domain.Account) error {
	tag, err := r.s.Queries.AccountUpdate(ctx, db.AccountUpdateParams{
		Name:      a.Name,
		Type:      string(a.Type),
		Version:   a.Version,
		UpdatedAt: helper.ToPgTS(a.UpdatedAt),
		ID:        helper.ToPgUUID(a.ID),
		UserID:    helper.ToPgUUID(a.UserID),
		Version_2: a.Version - 1,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *AccountRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	tag, err := r.s.Queries.AccountSoftDelete(ctx, db.AccountSoftDeleteParams{
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

func accountFromModel(m db.Account) *domain.Account {
	return &domain.Account{
		ID:           helper.FromPgUUID(m.ID),
		UserID:       helper.FromPgUUID(m.UserID),
		Name:         m.Name,
		Type:         domain.AccountType(m.Type),
		Currency:     m.Currency,
		BalanceMinor: m.BalanceMinor,
		Version:      m.Version,
		CreatedAt:    m.CreatedAt.Time,
		UpdatedAt:    m.UpdatedAt.Time,
		DeletedAt:    helper.FromPgTSPtr(m.DeletedAt),
	}
}
