package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moon-eye/velune/shared/otelx"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

// Store holds shared database access for repositories and the ledger.
type Store struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	otelx.InstrumentPoolConfig(cfg)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{
		Pool:    pool,
		Queries: db.New(pool),
	}, nil
}

func (s *Store) Close() {
	s.Pool.Close()
}
