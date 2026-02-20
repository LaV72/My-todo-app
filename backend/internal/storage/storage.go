package storage

import (
	"context"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
)

// Storage is the main interface for all storage implementations
// It composes TaskStorage, CategoryStorage, and StatsStorage
type Storage interface {
	TaskStorage
	CategoryStorage
	StatsStorage

	// Lifecycle methods
	Close() error
	Ping() error
	Backup(dest string) error
}

// TaskStorage handles all task-related operations
type TaskStorage interface {
	// Basic CRUD operations
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, id string) error

	// Query operations
	ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error)
	CountTasks(ctx context.Context, filter models.TaskFilter) (int, error)
	SearchTasks(ctx context.Context, query string) ([]*models.Task, error)

	// Bulk operations
	CreateTasksBulk(ctx context.Context, tasks []*models.Task) error
	UpdateTasksBulk(ctx context.Context, tasks []*models.Task) error
	DeleteTasksBulk(ctx context.Context, ids []string) error

	// Task-specific operations
	UpdateTaskStatus(ctx context.Context, id string, status models.TaskStatus) error
	ReorderTasks(ctx context.Context, ids []string) error

	// Objective operations
	AddObjective(ctx context.Context, taskID string, obj *models.Objective) error
	UpdateObjective(ctx context.Context, taskID, objID string, obj *models.Objective) error
	DeleteObjective(ctx context.Context, taskID, objID string) error
}

// CategoryStorage handles category operations
type CategoryStorage interface {
	CreateCategory(ctx context.Context, cat *models.Category) error
	GetCategory(ctx context.Context, id string) (*models.Category, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
	UpdateCategory(ctx context.Context, cat *models.Category) error
	DeleteCategory(ctx context.Context, id string) error
}

// StatsStorage handles statistics queries
type StatsStorage interface {
	GetStats(ctx context.Context) (*models.Stats, error)
	GetDailyStats(ctx context.Context, from, to time.Time) ([]*models.DailyStat, error)
	GetCategoryStats(ctx context.Context) (map[string]*models.CategoryStat, error)
}

// Config holds the configuration for creating a storage instance
type Config struct {
	Type string // "sqlite", "json", "memory"
	Path string // Path to database file or data directory
}

// New creates a new storage instance based on the provided configuration
// Note: This factory is optional. You can also instantiate implementations directly:
//   store, err := sqlite.New(path)
//   store, err := json.New(path)
//
// The direct approach avoids circular import issues and is more explicit.
func New(cfg Config) (Storage, error) {
	// Note: To avoid circular imports, implementations are not imported here.
	// Call implementation constructors directly in your main application:
	//
	// Example:
	//   import "github.com/LaV72/quest-todo/internal/storage/sqlite"
	//   store, err := sqlite.New(cfg.Path)
	//
	switch cfg.Type {
	case "sqlite", "json", "memory":
		return nil, ErrNotImplemented
	default:
		return nil, ErrUnknownStorageType
	}
}
