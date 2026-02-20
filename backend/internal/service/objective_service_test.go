package service

import (
	"context"
	"testing"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectiveService_CreateObjective(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		clock := NewFixedClock(fixedTime)
		idGen := NewFixedIDGenerator("obj-1")
		validate := validator.New()
		config := DefaultConfig()

		// Create a task first
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Title:    "Parent Task",
			Status:   models.StatusActive,
			Progress: 0,
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		req := models.ObjectiveRequest{
			Text:  "New Objective",
			Order: 0,
		}

		// Act
		objective, err := service.CreateObjective(context.Background(), "task-1", req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "obj-1", objective.ID)
		assert.Equal(t, "task-1", objective.TaskID)
		assert.Equal(t, "New Objective", objective.Text)
		assert.False(t, objective.Completed)
		assert.Equal(t, fixedTime, objective.CreatedAt)

		// Verify stored
		assert.Contains(t, mockStorage.Objectives, "obj-1")
	})

	t.Run("task not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		req := models.ObjectiveRequest{
			Text: "New Objective",
		}

		// Act
		_, err := service.CreateObjective(context.Background(), "nonexistent", req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTaskNotFound)
	})

	t.Run("validation error - empty text", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		mockStorage.Tasks["task-1"] = &models.Task{ID: "task-1"}
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		req := models.ObjectiveRequest{
			Text: "", // Empty text
		}

		// Act
		_, err := service.CreateObjective(context.Background(), "task-1", req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("progress recalculated after creation", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := NewFixedIDGenerator("obj-2")
		validate := validator.New()
		config := DefaultConfig()

		// Task with one existing completed objective
		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Progress: 100, // Currently 100%
			Objectives: []models.Objective{
				{ID: "obj-1", Completed: true},
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		req := models.ObjectiveRequest{
			Text: "Second Objective",
		}

		// Act
		_, err := service.CreateObjective(context.Background(), "task-1", req)

		// Assert
		require.NoError(t, err)

		// Progress should be recalculated: 1/2 = 50%
		task := mockStorage.Tasks["task-1"]
		assert.Equal(t, 50.0, task.Progress)
	})
}

func TestObjectiveService_UpdateObjective(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Tasks["task-1"] = &models.Task{
			ID: "task-1",
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1", Text: "Original", Completed: false},
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		newText := "Updated Text"
		req := models.ObjectiveUpdateRequest{
			Text: &newText,
		}

		// Act
		objective, err := service.UpdateObjective(context.Background(), "obj-1", req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Updated Text", objective.Text)
	})

	t.Run("objective not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		newText := "Updated"
		req := models.ObjectiveUpdateRequest{
			Text: &newText,
		}

		// Act
		_, err := service.UpdateObjective(context.Background(), "nonexistent", req)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrObjectiveNotFound)
	})

	t.Run("completion change triggers progress recalculation", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Progress: 0, // 0% progress initially
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1", Completed: false},
				{ID: "obj-2", TaskID: "task-1", Completed: false},
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		completed := true
		req := models.ObjectiveUpdateRequest{
			Completed: &completed,
		}

		// Act
		_, err := service.UpdateObjective(context.Background(), "obj-1", req)

		// Assert
		require.NoError(t, err)

		// Progress should be 1/2 = 50%
		task := mockStorage.Tasks["task-1"]
		assert.Equal(t, 50.0, task.Progress)
	})
}

func TestObjectiveService_DeleteObjective(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Tasks["task-1"] = &models.Task{
			ID: "task-1",
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1"},
				{ID: "obj-2", TaskID: "task-1"},
			},
		}
		mockStorage.Objectives["obj-1"] = &models.Objective{ID: "obj-1"}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act
		err := service.DeleteObjective(context.Background(), "obj-1")

		// Assert
		require.NoError(t, err)
		assert.NotContains(t, mockStorage.Objectives, "obj-1")
	})

	t.Run("objective not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act
		err := service.DeleteObjective(context.Background(), "nonexistent")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrObjectiveNotFound)
	})

	t.Run("progress recalculated after deletion", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Progress: 50, // 1/2 = 50%
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1", Completed: true},
				{ID: "obj-2", TaskID: "task-1", Completed: false},
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act - delete the incomplete objective
		err := service.DeleteObjective(context.Background(), "obj-2")

		// Assert
		require.NoError(t, err)

		// Progress should be 1/1 = 100%
		task := mockStorage.Tasks["task-1"]
		assert.Equal(t, 100.0, task.Progress)
	})
}

func TestObjectiveService_ToggleObjective(t *testing.T) {
	t.Run("toggle from incomplete to complete", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Status:   models.StatusActive,
			Progress: 0,
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1", Completed: false},
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act
		objective, err := service.ToggleObjective(context.Background(), "obj-1")

		// Assert
		require.NoError(t, err)
		assert.True(t, objective.Completed)

		// Progress should be 100%
		task := mockStorage.Tasks["task-1"]
		assert.Equal(t, 100.0, task.Progress)
	})

	t.Run("toggle from complete to incomplete", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Progress: 100,
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1", Completed: true},
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act
		objective, err := service.ToggleObjective(context.Background(), "obj-1")

		// Assert
		require.NoError(t, err)
		assert.False(t, objective.Completed)

		// Progress should be 0%
		task := mockStorage.Tasks["task-1"]
		assert.Equal(t, 0.0, task.Progress)
	})

	t.Run("auto-complete task when all objectives done", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		clock := NewFixedClock(fixedTime)
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()
		config.AutoCompleteOnFullProgress = true

		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Status:   models.StatusActive,
			Progress: 50,
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1", Completed: true},
				{ID: "obj-2", TaskID: "task-1", Completed: false}, // Last one
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act - complete the last objective
		_, err := service.ToggleObjective(context.Background(), "obj-2")

		// Assert
		require.NoError(t, err)

		// Task should be auto-completed
		task := mockStorage.Tasks["task-1"]
		assert.Equal(t, models.StatusComplete, task.Status)
		assert.NotNil(t, task.CompletedAt)
		assert.Equal(t, fixedTime, *task.CompletedAt)
		assert.Equal(t, 100.0, task.Progress)
	})

	t.Run("no auto-complete when disabled", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()
		config.AutoCompleteOnFullProgress = false // Disabled

		mockStorage.Tasks["task-1"] = &models.Task{
			ID:       "task-1",
			Status:   models.StatusActive,
			Progress: 50,
			Objectives: []models.Objective{
				{ID: "obj-1", TaskID: "task-1", Completed: true},
				{ID: "obj-2", TaskID: "task-1", Completed: false},
			},
		}

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act
		_, err := service.ToggleObjective(context.Background(), "obj-2")

		// Assert
		require.NoError(t, err)

		// Task should remain active
		task := mockStorage.Tasks["task-1"]
		assert.Equal(t, models.StatusActive, task.Status)
		assert.Nil(t, task.CompletedAt)
		assert.Equal(t, 100.0, task.Progress)
	})

	t.Run("objective not found", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config)

		// Act
		_, err := service.ToggleObjective(context.Background(), "nonexistent")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrObjectiveNotFound)
	})
}

func TestObjectiveService_ProgressCalculation(t *testing.T) {
	t.Run("no objectives returns 0%", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config).(*ObjectiveServiceImpl)

		// Act
		progress := service.calculateProgress([]models.Objective{})

		// Assert
		assert.Equal(t, 0.0, progress)
	})

	t.Run("all completed returns 100%", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config).(*ObjectiveServiceImpl)

		objectives := []models.Objective{
			{Completed: true},
			{Completed: true},
			{Completed: true},
		}

		// Act
		progress := service.calculateProgress(objectives)

		// Assert
		assert.Equal(t, 100.0, progress)
	})

	t.Run("partial completion", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()
		clock := &SystemClock{}
		idGen := &UUIDGenerator{}
		validate := validator.New()
		config := DefaultConfig()

		service := NewObjectiveService(mockStorage, clock, idGen, validate, config).(*ObjectiveServiceImpl)

		objectives := []models.Objective{
			{Completed: true},
			{Completed: false},
			{Completed: true},
			{Completed: false},
		}

		// Act
		progress := service.calculateProgress(objectives)

		// Assert
		assert.Equal(t, 50.0, progress) // 2/4 = 50%
	})
}
