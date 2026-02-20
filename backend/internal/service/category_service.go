package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
	"github.com/go-playground/validator/v10"
)

// CategoryServiceImpl implements CategoryService
type CategoryServiceImpl struct {
	storage   storage.Storage
	validator *validator.Validate
	config    *Config
}

// NewCategoryService creates a new CategoryService
func NewCategoryService(storage storage.Storage, validate *validator.Validate, config *Config) CategoryService {
	return &CategoryServiceImpl{
		storage:   storage,
		validator: validate,
		config:    config,
	}
}

// CreateCategory creates a new category
func (s *CategoryServiceImpl) CreateCategory(ctx context.Context, req models.CategoryCreateRequest) (*models.Category, error) {
	// 1. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, ErrInvalidInput
	}

	// 2. Build category
	category := &models.Category{
		ID:    req.Name, // Use name as ID for simplicity
		Name:  req.Name,
		Color: req.Color,
		Icon:  req.Icon,
		Type:  req.Type,
	}

	// 3. Persist
	if err := s.storage.CreateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryServiceImpl) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	category, err := s.storage.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	return category, nil
}

// UpdateCategory updates a category
func (s *CategoryServiceImpl) UpdateCategory(ctx context.Context, id string, req models.CategoryUpdateRequest) (*models.Category, error) {
	// 1. Get existing category
	category, err := s.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, ErrInvalidInput
	}

	// 3. Apply updates
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Color != nil {
		category.Color = *req.Color
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}

	// 4. Persist
	if err := s.storage.UpdateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *CategoryServiceImpl) DeleteCategory(ctx context.Context, id string) error {
	// 1. Business rule: Cannot delete category with active tasks
	if s.config.EnableCategoryRestrictions {
		count, err := s.storage.CountTasks(ctx, models.TaskFilter{
			Categories:       []string{id},
			IncludeCompleted: false,
		})
		if err != nil {
			return fmt.Errorf("count tasks in category: %w", err)
		}

		if count > 0 {
			return ErrCannotDeleteCategory
		}
	}

	// 2. Delete category
	if err := s.storage.DeleteCategory(ctx, id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("delete category: %w", err)
	}

	return nil
}

// ListCategories lists all categories
func (s *CategoryServiceImpl) ListCategories(ctx context.Context) ([]*models.Category, error) {
	categories, err := s.storage.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}
