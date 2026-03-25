package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

type DedupeRepo struct{ s *Store }

func NewDedupeRepo(s *Store) *DedupeRepo { return &DedupeRepo{s: s} }

func (r *DedupeRepo) SeenOrMark(ctx context.Context, key string, eventID uuid.UUID) (bool, error) {
	err := db.New(r.s.Pool).EventDedupeInsert(ctx, db.EventDedupeInsertParams{
		IdempotencyKey: key,
		EventID:        helper.ToPgUUID(eventID),
	})
	if err == nil {
		return false, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true, nil
	}
	return false, err
}
