package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TaskServiceImpl implements TaskService
type TaskServiceImpl struct {
	storage   storage.Storage
	validator *validator.Validate
	idGen     IDGenerator
	clock     Clock
	config    *Config
}

// UUIDGenerator generates UUID v4 IDs
type UUIDGenerator struct{}

func (g *UUIDGenerator) Generate() string {
	return uuid.New().String()
}

// NewTaskService creates a new TaskService
func NewTaskService(storage storage.Storage, clock Clock, idGen IDGenerator, validate *validator.Validate, config *Config) TaskService {
	return &TaskServiceImpl{
		storage:   storage,
		validator: validate,
		idGen:     idGen,
		clock:     clock,
		config:    config,
	}
}

// CreateTask creates a new task with validation and business logic
func (s *TaskServiceImpl) CreateTask(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error) {
	// 1. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, s.wrapValidationError(err)
	}

	// 2. Business rule: deadline must be in future
	if !s.config.AllowPastDeadlines && req.Deadline != nil && req.Deadline.Date != nil {
		if req.Deadline.Date.Before(s.clock.Now()) {
			return nil, ErrDeadlineInPast
		}
	}

	// 3. Get next order value
	order, err := s.getNextOrder(ctx)
	if err != nil {
		return nil, fmt.Errorf("get next order: %w", err)
	}

	// 4. Build task
	now := s.clock.Now()
	task := &models.Task{
		ID:          s.idGen.Generate(),
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Priority:    req.Priority,
		Category:    req.Category,
		Status:      models.StatusActive,
		Notes:       req.Notes,
		Reward:      req.Reward,
		Tags:        req.Tags,
		Order:       order,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Set deadline if provided
	if req.Deadline != nil {
		task.Deadline = *req.Deadline
	}

	// 5. Create objectives
	objectives := make([]models.Objective, len(req.Objectives))
	for i, objReq := range req.Objectives {
		objectives[i] = models.Objective{
			ID:        s.idGen.Generate(),
			TaskID:    task.ID,
			Text:      strings.TrimSpace(objReq.Text),
			Completed: false,
			Order:     objReq.Order,
			CreatedAt: now,
		}
	}
	task.Objectives = objectives

	// 6. Calculate progress
	task.Progress = s.calculateProgress(objectives)

	// 7. Persist to storage
	if err := s.storage.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (s *TaskServiceImpl) GetTask(ctx context.Context, id string) (*models.Task, error) {
	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return task, nil
}

// UpdateTask updates a task with partial updates
func (s *TaskServiceImpl) UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error) {
	// 1. Get existing task
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Validate input
	if err := s.validator.Struct(req); err != nil {
		return nil, s.wrapValidationError(err)
	}

	// 3. Apply updates (only non-nil fields)
	if req.Title != nil {
		task.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		task.Description = strings.TrimSpace(*req.Description)
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Deadline != nil {
		// Business rule: validate future deadline
		if !s.config.AllowPastDeadlines && req.Deadline.Date != nil {
			if req.Deadline.Date.Before(s.clock.Now()) {
				return nil, ErrDeadlineInPast
			}
		}
		task.Deadline = *req.Deadline
	}
	if req.Category != nil {
		task.Category = *req.Category
	}
	if req.Notes != nil {
		task.Notes = *req.Notes
	}
	if req.Reward != nil {
		task.Reward = *req.Reward
	}
	if req.Tags != nil {
		task.Tags = req.Tags
	}

	// 4. Update timestamp
	task.UpdatedAt = s.clock.Now()

	// 5. Persist
	if err := s.storage.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

// DeleteTask deletes a task
func (s *TaskServiceImpl) DeleteTask(ctx context.Context, id string) error {
	// Verify task exists
	if _, err := s.GetTask(ctx, id); err != nil {
		return err
	}

	if err := s.storage.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

// ListTasks lists tasks with filtering
func (s *TaskServiceImpl) ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error) {
	tasks, err := s.storage.ListTasks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return tasks, nil
}

// CountTasks counts tasks matching filter
func (s *TaskServiceImpl) CountTasks(ctx context.Context, filter models.TaskFilter) (int, error) {
	count, err := s.storage.CountTasks(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}
	return count, nil
}

// SearchTasks performs full-text search
func (s *TaskServiceImpl) SearchTasks(ctx context.Context, query string) ([]*models.Task, error) {
	tasks, err := s.storage.SearchTasks(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search tasks: %w", err)
	}
	return tasks, nil
}

// CreateTasksBulk creates multiple tasks
func (s *TaskServiceImpl) CreateTasksBulk(ctx context.Context, reqs []models.TaskCreateRequest) ([]*models.Task, error) {
	// 1. Validate batch size
	if len(reqs) > s.config.MaxBulkSize {
		return nil, ErrBulkSizeTooLarge
	}

	// 2. Validate all requests first
	for i, req := range reqs {
		if err := s.validator.Struct(req); err != nil {
			return nil, fmt.Errorf("request %d: %w", i, s.wrapValidationError(err))
		}
	}

	// 3. Build all tasks
	tasks := make([]*models.Task, len(reqs))
	for i, req := range reqs {
		task, err := s.buildTaskFromRequest(req)
		if err != nil {
			return nil, fmt.Errorf("build task %d: %w", i, err)
		}
		tasks[i] = task
	}

	// 4. Persist all
	if err := s.storage.CreateTasksBulk(ctx, tasks); err != nil {
		return nil, fmt.Errorf("create tasks bulk: %w", err)
	}

	return tasks, nil
}

// UpdateTasksBulk updates multiple tasks
func (s *TaskServiceImpl) UpdateTasksBulk(ctx context.Context, updates map[string]models.TaskUpdateRequest) error {
	// 1. Validate batch size
	if len(updates) > s.config.MaxBulkSize {
		return ErrBulkSizeTooLarge
	}

	// 2. Update each task
	for id, req := range updates {
		if _, err := s.UpdateTask(ctx, id, req); err != nil {
			return fmt.Errorf("update task %s: %w", id, err)
		}
	}

	return nil
}

// DeleteTasksBulk deletes multiple tasks
func (s *TaskServiceImpl) DeleteTasksBulk(ctx context.Context, ids []string) error {
	// 1. Validate batch size
	if len(ids) > s.config.MaxBulkSize {
		return ErrBulkSizeTooLarge
	}

	// 2. Delete all
	if err := s.storage.DeleteTasksBulk(ctx, ids); err != nil {
		return fmt.Errorf("delete tasks bulk: %w", err)
	}

	return nil
}

// CompleteTask marks a task as completed
func (s *TaskServiceImpl) CompleteTask(ctx context.Context, id string) (*models.Task, error) {
	// 1. Get current task
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Validate state transition
	if task.Status == models.StatusComplete {
		return nil, ErrAlreadyCompleted
	}

	if task.Status == models.StatusFailed {
		return nil, ErrCannotCompleteFailedTask
	}

	// 3. Business rule: check if all objectives must be completed
	if s.config.RequireAllObjectives && task.Progress < 100 {
		return nil, ErrObjectivesIncomplete
	}

	// 4. Update task
	now := s.clock.Now()
	task.Status = models.StatusComplete
	task.CompletedAt = &now
	task.UpdatedAt = now

	// 5. Persist
	if err := s.storage.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("complete task: %w", err)
	}

	return task, nil
}

// FailTask marks a task as failed
func (s *TaskServiceImpl) FailTask(ctx context.Context, id string) (*models.Task, error) {
	// 1. Get current task
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Update task
	task.Status = models.StatusFailed
	task.UpdatedAt = s.clock.Now()

	// 3. Persist
	if err := s.storage.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("fail task: %w", err)
	}

	return task, nil
}

// ReactivateTask reactivates a completed or failed task
func (s *TaskServiceImpl) ReactivateTask(ctx context.Context, id string) (*models.Task, error) {
	// 1. Get current task
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Update task
	task.Status = models.StatusActive
	task.CompletedAt = nil
	task.UpdatedAt = s.clock.Now()

	// 3. Persist
	if err := s.storage.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("reactivate task: %w", err)
	}

	return task, nil
}

// ReorderTasks reorders tasks
func (s *TaskServiceImpl) ReorderTasks(ctx context.Context, ids []string) error {
	if err := s.storage.ReorderTasks(ctx, ids); err != nil {
		return fmt.Errorf("reorder tasks: %w", err)
	}
	return nil
}

// Helper functions

func (s *TaskServiceImpl) calculateProgress(objectives []models.Objective) float64 {
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

func (s *TaskServiceImpl) getNextOrder(ctx context.Context) (int, error) {
	// Get count of all tasks for next order value
	count, err := s.storage.CountTasks(ctx, models.TaskFilter{})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *TaskServiceImpl) buildTaskFromRequest(req models.TaskCreateRequest) (*models.Task, error) {
	// Business rule validation
	if !s.config.AllowPastDeadlines && req.Deadline != nil && req.Deadline.Date != nil {
		if req.Deadline.Date.Before(s.clock.Now()) {
			return nil, ErrDeadlineInPast
		}
	}

	now := s.clock.Now()
	task := &models.Task{
		ID:          s.idGen.Generate(),
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Priority:    req.Priority,
		Category:    req.Category,
		Status:      models.StatusActive,
		Notes:       req.Notes,
		Reward:      req.Reward,
		Tags:        req.Tags,
		Order:       0, // Will be set by bulk operation
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Set deadline if provided
	if req.Deadline != nil {
		task.Deadline = *req.Deadline
	}

	// Create objectives
	objectives := make([]models.Objective, len(req.Objectives))
	for i, objReq := range req.Objectives {
		objectives[i] = models.Objective{
			ID:        s.idGen.Generate(),
			TaskID:    task.ID,
			Text:      strings.TrimSpace(objReq.Text),
			Completed: false,
			Order:     objReq.Order,
			CreatedAt: now,
		}
	}
	task.Objectives = objectives

	// Calculate progress
	task.Progress = s.calculateProgress(objectives)

	return task, nil
}

func (s *TaskServiceImpl) wrapValidationError(err error) error {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return ErrInvalidInput
	}

	errors := make([]ValidationError, len(validationErrs))
	for i, fieldErr := range validationErrs {
		errors[i] = ValidationError{
			Field:   fieldErr.Field(),
			Message: s.getErrorMessage(fieldErr),
		}
	}

	return &MultiValidationError{Errors: errors}
}

func (s *TaskServiceImpl) getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}
