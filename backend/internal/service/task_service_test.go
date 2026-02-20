package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskService_CreateTask(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		clock := NewFixedClock(fixedTime)
		idGen := NewFixedIDGenerator("task-1", "obj-1", "obj-2")
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		req := models.TaskCreateRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Priority:    3,
			Category:    "work",
			Objectives: []models.ObjectiveRequest{
				{Text: "Objective 1", Order: 0},
				{Text: "Objective 2", Order: 1},
			},
			Tags: []string{"urgent", "important"},
		}

		// Act
		task, err := service.CreateTask(context.Background(), req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "task-1", task.ID)
		assert.Equal(t, "Test Task", task.Title)
		assert.Equal(t, "Test Description", task.Description)
		assert.Equal(t, 3, task.Priority)
		assert.Equal(t, "work", task.Category)
		assert.Equal(t, models.StatusActive, task.Status)
		assert.Equal(t, fixedTime, task.CreatedAt)
		assert.Equal(t, fixedTime, task.UpdatedAt)
		assert.Len(t, task.Objectives, 2)
		assert.Equal(t, "obj-1", task.Objectives[0].ID)
		assert.Equal(t, "Objective 1", task.Objectives[0].Text)
		assert.Equal(t, 0.0, task.Progress) // No objectives completed
		assert.Len(t, task.Tags, 2)

		// Verify stored in mock
		stored, err := mockStorage.GetTask(context.Background(), "task-1")
		require.NoError(t, err)
		assert.Equal(t, task.ID, stored.ID)
	})

	t.Run("validation error - empty title", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		req := models.TaskCreateRequest{
			Title:    "", // Empty title
			Priority: 3,
		}

		// Act
		_, err := service.CreateTask(context.Background(), req)

		// Assert
		require.Error(t, err)
		var validationErr *MultiValidationError
		if errors.As(err, &validationErr) {
			assert.Contains(t, validationErr.Errors[0].Message, "Title")
		}
	})

	t.Run("validation error - invalid priority", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		req := models.TaskCreateRequest{
			Title:    "Test",
			Priority: 0, // Invalid priority
		}

		// Act
		_, err := service.CreateTask(context.Background(), req)

		// Assert
		require.Error(t, err)
	})

	t.Run("deadline in past rejected", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		clock := NewFixedClock(fixedTime)
		idGen := NewFixedIDGenerator("task-1")
		validate := validator.New()
		config := DefaultConfig()
		config.AllowPastDeadlines = false

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		pastDate := fixedTime.Add(-24 * time.Hour)
		req := models.TaskCreateRequest{
			Title:    "Test Task",
			Priority: 3,
			Deadline: &models.Deadline{
				Type: "short",
				Date: &pastDate,
			},
		}

		// Act
		_, err := service.CreateTask(context.Background(), req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDeadlineInPast)
	})

	t.Run("progress calculation with objectives", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := NewFixedIDGenerator("task-1", "obj-1", "obj-2", "obj-3")
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		req := models.TaskCreateRequest{
			Title:    "Test Task",
			Priority: 3,
			Objectives: []models.ObjectiveRequest{
				{Text: "Step 1"},
				{Text: "Step 2"},
				{Text: "Step 3"},
			},
		}

		// Act
		task, err := service.CreateTask(context.Background(), req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 0.0, task.Progress) // 0/3 = 0%
	})
}

func TestTaskService_GetTask(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:    "task-1",
			Title: "Existing Task",
		}

		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		task, err := service.GetTask(context.Background(), "task-1")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "task-1", task.ID)
		assert.Equal(t, "Existing Task", task.Title)
	})

	t.Run("task not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		_, err := service.GetTask(context.Background(), "nonexistent")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTaskNotFound)
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:        "task-1",
			Title:     "Original Title",
			Priority:  3,
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		}

		newTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		clock := NewFixedClock(newTime)
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		newTitle := "Updated Title"
		newPriority := 5
		req := models.TaskUpdateRequest{
			Title:    &newTitle,
			Priority: &newPriority,
		}

		// Act
		task, err := service.UpdateTask(context.Background(), "task-1", req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", task.Title)
		assert.Equal(t, 5, task.Priority)
		assert.Equal(t, newTime, task.UpdatedAt)
		assert.Equal(t, fixedTime, task.CreatedAt) // CreatedAt unchanged
	})

	t.Run("partial update", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Title:    "Original",
			Priority: 3,
			Category: "work",
		}

		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		newTitle := "Updated"
		req := models.TaskUpdateRequest{
			Title: &newTitle,
			// Priority and Category not provided
		}

		// Act
		task, err := service.UpdateTask(context.Background(), "task-1", req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Updated", task.Title)
		assert.Equal(t, 3, task.Priority)     // Unchanged
		assert.Equal(t, "work", task.Category) // Unchanged
	})

	t.Run("task not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		newTitle := "Updated"
		req := models.TaskUpdateRequest{
			Title: &newTitle,
		}

		// Act
		_, err := service.UpdateTask(context.Background(), "nonexistent", req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTaskNotFound)
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:    "task-1",
			Title: "To Delete",
		}

		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		err := service.DeleteTask(context.Background(), "task-1")

		// Assert
		require.NoError(t, err)
		_, err = mockStorage.GetTask(context.Background(), "task-1")
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})

	t.Run("task not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		err := service.DeleteTask(context.Background(), "nonexistent")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTaskNotFound)
	})
}

func TestTaskService_CompleteTask(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Title:    "To Complete",
			Status:   models.StatusActive,
			Progress: 100,
		}

		clock := NewFixedClock(fixedTime)
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()
		config.RequireAllObjectives = false

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		task, err := service.CompleteTask(context.Background(), "task-1")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, models.StatusComplete, task.Status)
		assert.NotNil(t, task.CompletedAt)
		assert.Equal(t, fixedTime, *task.CompletedAt)
		assert.Equal(t, fixedTime, task.UpdatedAt)
	})

	t.Run("already completed", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		completedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:          "task-1",
			Status:      models.StatusComplete,
			CompletedAt: &completedTime,
		}

		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		_, err := service.CompleteTask(context.Background(), "task-1")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAlreadyCompleted)
	})

	t.Run("cannot complete failed task", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:     "task-1",
			Status: models.StatusFailed,
		}

		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		_, err := service.CompleteTask(context.Background(), "task-1")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCannotCompleteFailedTask)
	})

	t.Run("objectives incomplete", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Status:   models.StatusActive,
			Progress: 50, // Only 50% complete
			Objectives: []models.Objective{
				{ID: "obj-1", Completed: true},
				{ID: "obj-2", Completed: false},
			},
		}

		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()
		config.RequireAllObjectives = true

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		_, err := service.CompleteTask(context.Background(), "task-1")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrObjectivesIncomplete)
	})
}

func TestTaskService_FailTask(t *testing.T) {
	t.Run("successful fail", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:     "task-1",
			Status: models.StatusActive,
		}

		clock := NewFixedClock(fixedTime)
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		task, err := service.FailTask(context.Background(), "task-1")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, models.StatusFailed, task.Status)
		assert.Equal(t, fixedTime, task.UpdatedAt)
	})
}

func TestTaskService_ReactivateTask(t *testing.T) {
	t.Run("reactivate completed task", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		completedTime := fixedTime.Add(-24 * time.Hour)
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:          "task-1",
			Status:      models.StatusComplete,
			CompletedAt: &completedTime,
		}

		clock := NewFixedClock(fixedTime)
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		// Act
		task, err := service.ReactivateTask(context.Background(), "task-1")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, models.StatusActive, task.Status)
		assert.Nil(t, task.CompletedAt)
		assert.Equal(t, fixedTime, task.UpdatedAt)
	})
}

func TestTaskService_CreateTasksBulk(t *testing.T) {
	t.Run("successful bulk creation", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := NewFixedIDGenerator("task-1", "task-2", "task-3")
		validate := validator.New()
		config := DefaultConfig()

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		reqs := []models.TaskCreateRequest{
			{Title: "Task 1", Priority: 3},
			{Title: "Task 2", Priority: 4},
		}

		// Act
		tasks, err := service.CreateTasksBulk(context.Background(), reqs)

		// Assert
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "task-1", tasks[0].ID)
		assert.Equal(t, "Task 1", tasks[0].Title)
		assert.Equal(t, "task-2", tasks[1].ID)
		assert.Equal(t, "Task 2", tasks[1].Title)
	})

	t.Run("bulk size too large", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()
		config.MaxBulkSize = 2

		service := NewTaskService(mockStorage, clock, idGen, validate, config)

		reqs := make([]models.TaskCreateRequest, 3)
		for i := range reqs {
			reqs[i] = models.TaskCreateRequest{
				Title:    "Task",
				Priority: 3,
			}
		}

		// Act
		_, err := service.CreateTasksBulk(context.Background(), reqs)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBulkSizeTooLarge)
	})
}
