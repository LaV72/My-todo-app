package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
	"github.com/go-playground/validator/v10"
)

// ObjectiveServiceImpl implements ObjectiveService
type ObjectiveServiceImpl struct {
	storage     storage.Storage
	taskService TaskService
	validator   *validator.Validate
	idGen       IDGenerator
	clock       Clock
	config      *Config
}

// NewObjectiveService creates a new ObjectiveService
func NewObjectiveService(storage storage.Storage, taskService TaskService, clock Clock, idGen IDGenerator, validate *validator.Validate, config *Config) ObjectiveService {
	return &ObjectiveServiceImpl{
		storage:     storage,
		taskService: taskService,
		validator:   validate,
		idGen:       idGen,
		clock:       clock,
		config:      config,
	}
}

// CreateObjective creates a new objective for a task
func (s *ObjectiveServiceImpl) CreateObjective(ctx context.Context, taskID string, req models.ObjectiveRequest) (*models.Objective, error) {
	// 1. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, wrapValidationError(err)
	}

	// 2. Verify task exists
	_, err := s.storage.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	// 3. Build objective
	now := s.clock.Now()
	objective := &models.Objective{
		ID:        s.idGen.Generate(),
		TaskID:    taskID,
		Text:      strings.TrimSpace(req.Text),
		Completed: false,
		Order:     req.Order,
		CreatedAt: now,
	}

	// 4. Persist
	if err := s.storage.AddObjective(ctx, taskID, objective); err != nil {
		return nil, fmt.Errorf("create objective: %w", err)
	}

	// 5. Recalculate task progress and auto-complete if needed
	if err := s.taskService.RecalculateProgressAndAutoComplete(ctx, taskID); err != nil {
		return nil, fmt.Errorf("recalculate progress: %w", err)
	}

	return objective, nil
}

// UpdateObjective updates an objective
// Note: This requires finding the objective by searching through tasks
func (s *ObjectiveServiceImpl) UpdateObjective(ctx context.Context, id string, req models.ObjectiveUpdateRequest) (*models.Objective, error) {
	// 1. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, wrapValidationError(err)
	}

	// 2. Find objective by searching tasks (storage doesn't have GetObjective)
	task, objective, err := s.findObjective(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Apply updates
	completedChanged := false
	if req.Text != nil {
		objective.Text = strings.TrimSpace(*req.Text)
	}
	if req.Completed != nil {
		if objective.Completed != *req.Completed {
			completedChanged = true
		}
		objective.Completed = *req.Completed
	}

	// 4. Persist
	if err := s.storage.UpdateObjective(ctx, task.ID, objective.ID, objective); err != nil {
		return nil, fmt.Errorf("update objective: %w", err)
	}

	// 5. Recalculate task progress and auto-complete if completion changed
	if completedChanged {
		if err := s.taskService.RecalculateProgressAndAutoComplete(ctx, task.ID); err != nil {
			return nil, fmt.Errorf("recalculate progress: %w", err)
		}
	}

	return objective, nil
}

// DeleteObjective deletes an objective
func (s *ObjectiveServiceImpl) DeleteObjective(ctx context.Context, id string) error {
	// 1. Find objective
	task, objective, err := s.findObjective(ctx, id)
	if err != nil {
		return err
	}

	// 2. Delete objective
	if err := s.storage.DeleteObjective(ctx, task.ID, objective.ID); err != nil {
		return fmt.Errorf("delete objective: %w", err)
	}

	// 3. Recalculate task progress and auto-complete
	if err := s.taskService.RecalculateProgressAndAutoComplete(ctx, task.ID); err != nil {
		return fmt.Errorf("recalculate progress: %w", err)
	}

	return nil
}

// ToggleObjective toggles objective completion status
func (s *ObjectiveServiceImpl) ToggleObjective(ctx context.Context, id string) (*models.Objective, error) {
	// 1. Find objective
	task, objective, err := s.findObjective(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Toggle completion
	objective.Completed = !objective.Completed

	// 3. Update objective
	if err := s.storage.UpdateObjective(ctx, task.ID, objective.ID, objective); err != nil {
		return nil, fmt.Errorf("toggle objective: %w", err)
	}

	// 4. Recalculate progress and auto-complete if needed
	if err := s.taskService.RecalculateProgressAndAutoComplete(ctx, task.ID); err != nil {
		return nil, fmt.Errorf("recalculate progress: %w", err)
	}

	return objective, nil
}

// Helper functions

// findObjective finds an objective by ID and its parent task
func (s *ObjectiveServiceImpl) findObjective(ctx context.Context, objectiveID string) (*models.Task, *models.Objective, error) {
	// Get the objective and its task ID directly from storage
	objective, taskID, err := s.storage.GetObjective(ctx, objectiveID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, nil, ErrObjectiveNotFound
		}
		return nil, nil, fmt.Errorf("get objective: %w", err)
	}

	// Get the task
	task, err := s.storage.GetTask(ctx, taskID)
	if err != nil {
		return nil, nil, fmt.Errorf("get task: %w", err)
	}

	return task, objective, nil
}
