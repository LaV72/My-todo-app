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
	storage   storage.Storage
	validator *validator.Validate
	idGen     IDGenerator
	clock     Clock
	config    *Config
}

// NewObjectiveService creates a new ObjectiveService
func NewObjectiveService(storage storage.Storage, clock Clock, idGen IDGenerator, validate *validator.Validate, config *Config) ObjectiveService {
	return &ObjectiveServiceImpl{
		storage:   storage,
		validator: validate,
		idGen:     idGen,
		clock:     clock,
		config:    config,
	}
}

// CreateObjective creates a new objective for a task
func (s *ObjectiveServiceImpl) CreateObjective(ctx context.Context, taskID string, req models.ObjectiveRequest) (*models.Objective, error) {
	// 1. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, ErrInvalidInput
	}

	// 2. Verify task exists
	task, err := s.storage.GetTask(ctx, taskID)
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

	// 5. Recalculate task progress
	task, err = s.storage.GetTask(ctx, taskID) // Reload with new objective
	if err != nil {
		return nil, fmt.Errorf("reload task: %w", err)
	}

	if err := s.recalculateTaskProgress(ctx, task); err != nil {
		return nil, fmt.Errorf("recalculate progress: %w", err)
	}

	return objective, nil
}

// UpdateObjective updates an objective
// Note: This requires finding the objective by searching through tasks
func (s *ObjectiveServiceImpl) UpdateObjective(ctx context.Context, id string, req models.ObjectiveUpdateRequest) (*models.Objective, error) {
	// 1. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, ErrInvalidInput
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

	// 5. Recalculate task progress if completion changed
	if completedChanged {
		task, err = s.storage.GetTask(ctx, task.ID) // Reload
		if err != nil {
			return nil, fmt.Errorf("reload task: %w", err)
		}

		if err := s.recalculateTaskProgress(ctx, task); err != nil {
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

	// 3. Recalculate task progress
	task, err = s.storage.GetTask(ctx, task.ID) // Reload
	if err != nil {
		return fmt.Errorf("reload task: %w", err)
	}

	if err := s.recalculateTaskProgress(ctx, task); err != nil {
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

	// 4. Reload task and recalculate progress
	task, err = s.storage.GetTask(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("reload task: %w", err)
	}

	if err := s.recalculateTaskProgress(ctx, task); err != nil {
		return nil, fmt.Errorf("recalculate progress: %w", err)
	}

	// 5. Optional: Auto-complete task if all objectives done
	if s.config.AutoCompleteOnFullProgress && task.Progress == 100 && task.Status == models.StatusActive {
		now := s.clock.Now()
		task.Status = models.StatusComplete
		task.CompletedAt = &now
		task.UpdatedAt = now

		if err := s.storage.UpdateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("auto-complete task: %w", err)
		}
	}

	return objective, nil
}

// Helper functions

func (s *ObjectiveServiceImpl) recalculateTaskProgress(ctx context.Context, task *models.Task) error {
	// Recalculate progress from objectives
	task.Progress = s.calculateProgress(task.Objectives)
	task.UpdatedAt = s.clock.Now()

	// Update task
	if err := s.storage.UpdateTask(ctx, task); err != nil {
		return err
	}

	return nil
}

func (s *ObjectiveServiceImpl) calculateProgress(objectives []models.Objective) float64 {
	if len(objectives) == 0 {
		return 0.0
	}

	completed := 0
	for _, obj := range objectives {
		if obj.Completed {
			completed++
		}
	}

	return float64(completed) / float64(len(objectives)) * 100.0
}

// findObjective finds an objective by ID across all tasks
// This is a workaround since storage doesn't provide GetObjective
func (s *ObjectiveServiceImpl) findObjective(ctx context.Context, objectiveID string) (*models.Task, *models.Objective, error) {
	// Get all tasks (not ideal for large datasets, but works for now)
	tasks, err := s.storage.ListTasks(ctx, models.TaskFilter{
		IncludeCompleted: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("list tasks: %w", err)
	}

	// Search through tasks for the objective
	for _, task := range tasks {
		for i := range task.Objectives {
			if task.Objectives[i].ID == objectiveID {
				return task, &task.Objectives[i], nil
			}
		}
	}

	return nil, nil, ErrObjectiveNotFound
}
