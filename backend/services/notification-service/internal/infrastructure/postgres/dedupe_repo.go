package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type DedupeRepo struct{ s *Store }

func NewDedupeRepo(s *Store) *DedupeRepo { return &DedupeRepo{s: s} }

func (r *DedupeRepo) SeenOrMark(ctx context.Context, key string, eventID uuid.UUID) (bool, error) {
	_, err := r.s.Pool.Exec(ctx, `
		INSERT INTO event_dedupe (idempotency_key, event_id, processed_at)
		VALUES ($1, $2, now())
	`, key, eventID)
	if err == nil {
		return false, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true, nil
	}
	return false, err
}
