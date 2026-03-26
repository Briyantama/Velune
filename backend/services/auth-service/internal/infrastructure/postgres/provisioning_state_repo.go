package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/moon-eye/velune/services/auth-service/internal/repository"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

type ProvisioningStateRepo struct {
	s *Store
}

func NewProvisioningStateRepo(s *Store) repository.ProvisioningStateRepository {
	return &ProvisioningStateRepo{s: s}
}

func (r *ProvisioningStateRepo) GetAccountProvisionedAt(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	ts, err := r.s.Queries.UserGetProvisioningAccountProvisionedAt(ctx, helper.ToPgUUID(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if !ts.Valid {
		return nil, nil
	}
	t := ts.Time
	return &t, nil
}

func (r *ProvisioningStateRepo) MarkAccountProvisionedAt(ctx context.Context, userID uuid.UUID, at time.Time) error {
	now := time.Now().UTC()
	return r.s.Queries.ProvisioningStateUpsertMarkAccountProvisioned(ctx, db.ProvisioningStateUpsertMarkAccountProvisionedParams{
		ID:                   pgtype.UUID{Bytes: uuid.New(), Valid: true},
		UserID:               helper.ToPgUUID(userID),
		AccountProvisionedAt: helper.ToPgTS(at),
		Version:              1,
		CreatedAt:            helper.ToPgTS(now),
		UpdatedAt:            helper.ToPgTS(now),
	})
}
