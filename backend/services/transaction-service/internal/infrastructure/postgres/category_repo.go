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

type CategoryRepo struct{ s *Store }

func NewCategoryRepo(s *Store) repository.CategoryRepository {
	return &CategoryRepo{s: s}
}

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	return r.s.Queries.CategoryCreate(ctx, db.CategoryCreateParams{
		ID:        helper.ToPgUUID(c.ID),
		UserID:    helper.ToPgUUID(c.UserID),
		Name:      c.Name,
		ParentID:  helper.ToPgUUIDPtr(c.ParentID),
		Version:   c.Version,
		CreatedAt: helper.ToPgTS(c.CreatedAt),
		UpdatedAt: helper.ToPgTS(c.UpdatedAt),
	})
}

func (r *CategoryRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error) {
	row, err := r.s.Queries.CategoryGetByID(ctx, db.CategoryGetByIDParams{
		ID:     helper.ToPgUUID(id),
		UserID: helper.ToPgUUID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return categoryFromModel(row), nil
}

func (r *CategoryRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Category, int64, error) {
	total, err := r.s.Queries.CategoryCountList(ctx, helper.ToPgUUID(userID))
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.s.Queries.CategoryList(ctx, db.CategoryListParams{
		UserID: helper.ToPgUUID(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, *categoryFromModel(row))
	}
	return out, total, nil
}

func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	tag, err := r.s.Queries.CategoryUpdate(ctx, db.CategoryUpdateParams{
		Name:      c.Name,
		ParentID:  helper.ToPgUUIDPtr(c.ParentID),
		Version:   c.Version,
		UpdatedAt: helper.ToPgTS(c.UpdatedAt),
		ID:        helper.ToPgUUID(c.ID),
		UserID:    helper.ToPgUUID(c.UserID),
		Version_2: c.Version - 1,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func (r *CategoryRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	tag, err := r.s.Queries.CategorySoftDelete(ctx, db.CategorySoftDeleteParams{
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

func categoryFromModel(m db.Category) *domain.Category {
	return &domain.Category{
		ID:        helper.FromPgUUID(m.ID),
		UserID:    helper.FromPgUUID(m.UserID),
		Name:      m.Name,
		ParentID:  helper.FromPgUUIDPtr(m.ParentID),
		Version:   m.Version,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
		DeletedAt: helper.FromPgTSPtr(m.DeletedAt),
	}
}
