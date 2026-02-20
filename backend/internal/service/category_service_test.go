package service

import (
	"context"
	"testing"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryService_CreateCategory(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		req := models.CategoryCreateRequest{
			Name:  "Work",
			Color: "#FF5733",
			Icon:  "briefcase",
			Type:  "main",
		}

		// Act
		category, err := service.CreateCategory(context.Background(), req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Work", category.ID) // ID is same as name
		assert.Equal(t, "Work", category.Name)
		assert.Equal(t, "#FF5733", category.Color)
		assert.Equal(t, "briefcase", category.Icon)
		assert.Equal(t, "main", category.Type)

		// Verify stored
		stored, err := mockStorage.GetCategory(context.Background(), "Work")
		require.NoError(t, err)
		assert.Equal(t, category.Name, stored.Name)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		req := models.CategoryCreateRequest{
			Name:  "", // Empty
			Color: "#FF5733",
			Type:  "main",
		}

		// Act
		_, err := service.CreateCategory(context.Background(), req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - invalid color", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		req := models.CategoryCreateRequest{
			Name:  "Work",
			Color: "not-a-hex-color", // Invalid
			Type:  "main",
		}

		// Act
		_, err := service.CreateCategory(context.Background(), req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - invalid type", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		req := models.CategoryCreateRequest{
			Name:  "Work",
			Color: "#FF5733",
			Type:  "invalid", // Must be "main" or "side"
		}

		// Act
		_, err := service.CreateCategory(context.Background(), req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestCategoryService_GetCategory(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Categories["work"] = &models.Category{
			ID:    "work",
			Name:  "Work",
			Color: "#FF5733",
			Type:  "main",
		}

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		category, err := service.GetCategory(context.Background(), "work")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "work", category.ID)
		assert.Equal(t, "Work", category.Name)
	})

	t.Run("category not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		_, err := service.GetCategory(context.Background(), "nonexistent")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Categories["work"] = &models.Category{
			ID:    "work",
			Name:  "Work",
			Color: "#FF5733",
			Icon:  "briefcase",
		}

		service := NewCategoryService(mockStorage, validate, config)

		newName := "Professional"
		newColor := "#00FF00"
		req := models.CategoryUpdateRequest{
			Name:  &newName,
			Color: &newColor,
		}

		// Act
		category, err := service.UpdateCategory(context.Background(), "work", req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Professional", category.Name)
		assert.Equal(t, "#00FF00", category.Color)
		assert.Equal(t, "briefcase", category.Icon) // Unchanged
	})

	t.Run("partial update", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Categories["work"] = &models.Category{
			ID:    "work",
			Name:  "Work",
			Color: "#FF5733",
			Icon:  "briefcase",
		}

		service := NewCategoryService(mockStorage, validate, config)

		newColor := "#0000FF"
		req := models.CategoryUpdateRequest{
			Color: &newColor,
			// Name and Icon not provided
		}

		// Act
		category, err := service.UpdateCategory(context.Background(), "work", req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Work", category.Name)       // Unchanged
		assert.Equal(t, "#0000FF", category.Color)   // Changed
		assert.Equal(t, "briefcase", category.Icon) // Unchanged
	})

	t.Run("category not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		newName := "Updated"
		req := models.CategoryUpdateRequest{
			Name: &newName,
		}

		// Act
		_, err := service.UpdateCategory(context.Background(), "nonexistent", req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()
		config.EnableCategoryRestrictions = false // Disable restrictions for this test

		mockStorage.Categories["work"] = &models.Category{
			ID:   "work",
			Name: "Work",
		}

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		err := service.DeleteCategory(context.Background(), "work")

		// Assert
		require.NoError(t, err)
		assert.NotContains(t, mockStorage.Categories, "work")
	})

	t.Run("cannot delete with active tasks", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()
		config.EnableCategoryRestrictions = true

		mockStorage.Categories["work"] = &models.Category{
			ID:   "work",
			Name: "Work",
		}

		// Add an active task in this category
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Category: "work",
			Status:   models.StatusActive,
		}

		// Mock CountTasks to return 1
		mockStorage.CountTasksFunc = func(ctx context.Context, filter models.TaskFilter) (int, error) {
			if len(filter.Categories) > 0 && filter.Categories[0] == "work" && !filter.IncludeCompleted {
				return 1, nil
			}
			return 0, nil
		}

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		err := service.DeleteCategory(context.Background(), "work")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCannotDeleteCategory)

		// Category should still exist
		assert.Contains(t, mockStorage.Categories, "work")
	})

	t.Run("can delete with completed tasks only", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()
		config.EnableCategoryRestrictions = true

		mockStorage.Categories["work"] = &models.Category{
			ID:   "work",
			Name: "Work",
		}

		// Mock CountTasks to return 0 active tasks
		mockStorage.CountTasksFunc = func(ctx context.Context, filter models.TaskFilter) (int, error) {
			if len(filter.Categories) > 0 && filter.Categories[0] == "work" && !filter.IncludeCompleted {
				return 0, nil // No active tasks
			}
			return 0, nil
		}

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		err := service.DeleteCategory(context.Background(), "work")

		// Assert
		require.NoError(t, err)
		assert.NotContains(t, mockStorage.Categories, "work")
	})

	t.Run("category not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		err := service.DeleteCategory(context.Background(), "nonexistent")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}

func TestCategoryService_ListCategories(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Categories["work"] = &models.Category{
			ID:   "work",
			Name: "Work",
			Type: "main",
		}
		mockStorage.Categories["personal"] = &models.Category{
			ID:   "personal",
			Name: "Personal",
			Type: "side",
		}

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		categories, err := service.ListCategories(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Len(t, categories, 2)

		// Check that both categories are returned
		ids := make([]string, len(categories))
		for i, cat := range categories {
			ids[i] = cat.ID
		}
		assert.Contains(t, ids, "work")
		assert.Contains(t, ids, "personal")
	})

	t.Run("empty list", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		validate := validator.New()
		config := DefaultConfig()

		service := NewCategoryService(mockStorage, validate, config)

		// Act
		categories, err := service.ListCategories(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Empty(t, categories)
	})
}
