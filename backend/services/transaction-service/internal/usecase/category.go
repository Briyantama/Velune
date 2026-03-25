package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
	"github.com/moon-eye/velune/services/transaction-service/internal/repository"
	"github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/pagination"
	"github.com/moon-eye/velune/shared/stringx"
)

type CategoryService struct {
	Categories repository.CategoryRepository
}

type CreateCategoryInput struct {
	Name     string     `validate:"required,min=1,max=200"`
	ParentID *uuid.UUID
}

func (s *CategoryService) Create(ctx context.Context, userID uuid.UUID, in CreateCategoryInput) (*domain.Category, error) {
	if in.ParentID != nil {
		p, err := s.Categories.GetByID(ctx, userID, *in.ParentID)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, errs.New("NOT_FOUND", "parent category not found", constx.StatusNotFound)
		}
	}
	now := time.Now().UTC()
	c := &domain.Category{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      stringx.TrimSpace(in.Name),
		ParentID:  in.ParentID,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.Categories.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Category, int64, error) {
	p := pagination.Normalize(page, limit)
	return s.Categories.List(ctx, userID, p.Limit, p.Offset)
}

type UpdateCategoryInput struct {
	Name     string     `validate:"required,min=1,max=200"`
	ParentID *uuid.UUID
}

func (s *CategoryService) Update(ctx context.Context, userID, id uuid.UUID, version int64, in UpdateCategoryInput) (*domain.Category, error) {
	c, err := s.Categories.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errs.ErrNotFound
	}
	if in.ParentID != nil && *in.ParentID == id {
		return nil, errs.New("VALIDATION_ERROR", "category cannot be its own parent", constx.StatusBadRequest)
	}
	now := time.Now().UTC()
	c.Name = stringx.TrimSpace(in.Name)
	c.ParentID = in.ParentID
	c.Version = version + 1
	c.UpdatedAt = now
	if err := s.Categories.Update(ctx, c); err != nil {
		if errors.Is(err, repository.ErrOptimisticLock) {
			return nil, errs.New("CONFLICT", "version conflict", constx.StatusConflict)
		}
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Delete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	err := s.Categories.SoftDelete(ctx, userID, id, version)
	if errors.Is(err, repository.ErrOptimisticLock) {
		return errs.New("CONFLICT", "version conflict", constx.StatusConflict)
	}
	return err
}
