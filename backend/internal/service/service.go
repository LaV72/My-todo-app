package service

import (
	"context"

	"github.com/LaV72/quest-todo/internal/models"
)

// TaskService handles all task-related business logic
type TaskService interface {
	// CRUD operations
	CreateTask(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error)
	DeleteTask(ctx context.Context, id string) error

	// Query operations
	ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error)
	CountTasks(ctx context.Context, filter models.TaskFilter) (int, error)
	SearchTasks(ctx context.Context, query string) ([]*models.Task, error)

	// Bulk operations
	CreateTasksBulk(ctx context.Context, tasks []models.TaskCreateRequest) ([]*models.Task, error)
	UpdateTasksBulk(ctx context.Context, updates map[string]models.TaskUpdateRequest) error
	DeleteTasksBulk(ctx context.Context, ids []string) error

	// Status transitions
	CompleteTask(ctx context.Context, id string) (*models.Task, error)
	FailTask(ctx context.Context, id string) (*models.Task, error)
	ReactivateTask(ctx context.Context, id string) (*models.Task, error)

	// Ordering
	ReorderTasks(ctx context.Context, ids []string) error
}

// ObjectiveService handles objective-related business logic
type ObjectiveService interface {
	CreateObjective(ctx context.Context, taskID string, req models.ObjectiveRequest) (*models.Objective, error)
	UpdateObjective(ctx context.Context, id string, req models.ObjectiveUpdateRequest) (*models.Objective, error)
	DeleteObjective(ctx context.Context, id string) error
	ToggleObjective(ctx context.Context, id string) (*models.Objective, error)
}

// CategoryService handles category-related business logic
type CategoryService interface {
	CreateCategory(ctx context.Context, req models.CategoryCreateRequest) (*models.Category, error)
	GetCategory(ctx context.Context, id string) (*models.Category, error)
	UpdateCategory(ctx context.Context, id string, req models.CategoryUpdateRequest) (*models.Category, error)
	DeleteCategory(ctx context.Context, id string) error
	ListCategories(ctx context.Context) ([]*models.Category, error)
}

// StatsService handles statistics and analytics
type StatsService interface {
	GetStats(ctx context.Context) (*models.Stats, error)
	GetCategoryStats(ctx context.Context) ([]models.CategoryStat, error)
}
