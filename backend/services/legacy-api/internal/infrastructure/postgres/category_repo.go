package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
)

type CategoryRepo struct{ s *Store }

func NewCategoryRepo(s *Store) repository.CategoryRepository {
	return &CategoryRepo{s: s}
}

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	const q = `
INSERT INTO categories (id, user_id, name, parent_id, version, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.s.Pool.Exec(ctx, q,
		c.ID, c.UserID, c.Name, c.ParentID, c.Version, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *CategoryRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error) {
	const q = `
SELECT id, user_id, name, parent_id, version, created_at, updated_at, deleted_at
FROM categories WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	row := r.s.Pool.QueryRow(ctx, q, id, userID)
	return scanCategory(row)
}

func (r *CategoryRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Category, int64, error) {
	const countQ = `SELECT COUNT(*) FROM categories WHERE user_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.s.Pool.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	const q = `
SELECT id, user_id, name, parent_id, version, created_at, updated_at, deleted_at
FROM categories WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name ASC
LIMIT $2 OFFSET $3`
	rows, err := r.s.Pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *c)
	}
	return out, total, rows.Err()
}

func scanCategory(row pgx.Row) (*domain.Category, error) {
	var c domain.Category
	err := row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.ParentID, &c.Version,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	const q = `
UPDATE categories SET name = $1, parent_id = $2, version = $3, updated_at = $4
WHERE id = $5 AND user_id = $6 AND version = $7 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q,
		c.Name, c.ParentID, c.Version, c.UpdatedAt, c.ID, c.UserID, c.Version-1,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *CategoryRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	const q = `
UPDATE categories SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL`
	tag, err := r.s.Pool.Exec(ctx, q, id, userID, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}
