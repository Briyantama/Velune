package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store bundles shared Postgres access for repositories.
type Store struct {
	Pool *pgxpool.Pool
}

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{Pool: pool}, nil
}

func (s *Store) Close() {
	if s == nil || s.Pool == nil {
		return
	}
	s.Pool.Close()
}

