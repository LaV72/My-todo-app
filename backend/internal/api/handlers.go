package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/go-playground/validator/v10"
)

// API holds the service dependencies for HTTP handlers
type API struct {
	TaskService      interface {
		CreateTask(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error)
		GetTask(ctx context.Context, id string) (*models.Task, error)
		UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error)
		DeleteTask(ctx context.Context, id string) error
		ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error)
		CountTasks(ctx context.Context, filter models.TaskFilter) (int, error)
		SearchTasks(ctx context.Context, query string) ([]*models.Task, error)
		CreateTasksBulk(ctx context.Context, tasks []models.TaskCreateRequest) ([]*models.Task, error)
		UpdateTasksBulk(ctx context.Context, updates map[string]models.TaskUpdateRequest) error
		DeleteTasksBulk(ctx context.Context, ids []string) error
		CompleteTask(ctx context.Context, id string) (*models.Task, error)
		FailTask(ctx context.Context, id string) (*models.Task, error)
		ReactivateTask(ctx context.Context, id string) (*models.Task, error)
		ReorderTasks(ctx context.Context, ids []string) error
	}
	ObjectiveService interface {
		CreateObjective(ctx context.Context, taskID string, req models.ObjectiveRequest) (*models.Objective, error)
		UpdateObjective(ctx context.Context, id string, req models.ObjectiveUpdateRequest) (*models.Objective, error)
		DeleteObjective(ctx context.Context, id string) error
		ToggleObjective(ctx context.Context, id string) (*models.Objective, error)
	}
	CategoryService interface {
		CreateCategory(ctx context.Context, req models.CategoryCreateRequest) (*models.Category, error)
		GetCategory(ctx context.Context, id string) (*models.Category, error)
		UpdateCategory(ctx context.Context, id string, req models.CategoryUpdateRequest) (*models.Category, error)
		DeleteCategory(ctx context.Context, id string) error
		ListCategories(ctx context.Context) ([]*models.Category, error)
	}
	StatsService interface {
		GetStats(ctx context.Context) (*models.Stats, error)
		GetCategoryStats(ctx context.Context) ([]models.CategoryStat, error)
	}
	Validator  *validator.Validate
	StartTime  time.Time
	AppVersion string
}

// NewAPI creates a new API instance
func NewAPI(
	taskService interface {
		CreateTask(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error)
		GetTask(ctx context.Context, id string) (*models.Task, error)
		UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error)
		DeleteTask(ctx context.Context, id string) error
		ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error)
		CountTasks(ctx context.Context, filter models.TaskFilter) (int, error)
		SearchTasks(ctx context.Context, query string) ([]*models.Task, error)
		CreateTasksBulk(ctx context.Context, tasks []models.TaskCreateRequest) ([]*models.Task, error)
		UpdateTasksBulk(ctx context.Context, updates map[string]models.TaskUpdateRequest) error
		DeleteTasksBulk(ctx context.Context, ids []string) error
		CompleteTask(ctx context.Context, id string) (*models.Task, error)
		FailTask(ctx context.Context, id string) (*models.Task, error)
		ReactivateTask(ctx context.Context, id string) (*models.Task, error)
		ReorderTasks(ctx context.Context, ids []string) error
	},
	objectiveService interface {
		CreateObjective(ctx context.Context, taskID string, req models.ObjectiveRequest) (*models.Objective, error)
		UpdateObjective(ctx context.Context, id string, req models.ObjectiveUpdateRequest) (*models.Objective, error)
		DeleteObjective(ctx context.Context, id string) error
		ToggleObjective(ctx context.Context, id string) (*models.Objective, error)
	},
	categoryService interface {
		CreateCategory(ctx context.Context, req models.CategoryCreateRequest) (*models.Category, error)
		GetCategory(ctx context.Context, id string) (*models.Category, error)
		UpdateCategory(ctx context.Context, id string, req models.CategoryUpdateRequest) (*models.Category, error)
		DeleteCategory(ctx context.Context, id string) error
		ListCategories(ctx context.Context) ([]*models.Category, error)
	},
	statsService interface {
		GetStats(ctx context.Context) (*models.Stats, error)
		GetCategoryStats(ctx context.Context) ([]models.CategoryStat, error)
	},
	v *validator.Validate,
	version string,
) *API {
	return &API{
		TaskService:      taskService,
		ObjectiveService: objectiveService,
		CategoryService:  categoryService,
		StatsService:     statsService,
		Validator:        v,
		StartTime:        time.Now(),
		AppVersion:       version,
	}
}

// DecodeJSONBody decodes a JSON request body into the target struct
func DecodeJSONBody(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return ErrEmptyBody
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

// ExtractID extracts an ID from the URL path
// Example: /tasks/{id} -> ExtractID(r, "/tasks/")
func ExtractID(r *http.Request, prefix string) string {
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	id = strings.TrimSuffix(id, "/")
	// Handle nested routes like /tasks/{id}/complete
	parts := strings.Split(id, "/")
	return parts[0]
}

// ParseTaskFilter parses query parameters into a TaskFilter
func ParseTaskFilter(r *http.Request) models.TaskFilter {
	q := r.URL.Query()
	filter := models.TaskFilter{}

	// Status filter (can be multiple)
	if statuses := q.Get("status"); statuses != "" {
		for _, s := range strings.Split(statuses, ",") {
			filter.Status = append(filter.Status, models.TaskStatus(strings.TrimSpace(s)))
		}
	}

	// Priority filter (can be multiple)
	if priorities := q.Get("priority"); priorities != "" {
		for _, p := range strings.Split(priorities, ",") {
			if pInt, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				filter.Priority = append(filter.Priority, pInt)
			}
		}
	}

	// Category filter (can be multiple)
	if categories := q.Get("category"); categories != "" {
		filter.Categories = strings.Split(categories, ",")
	}

	// Tags filter (can be multiple)
	if tags := q.Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	// Deadline type
	if deadlineType := q.Get("deadlineType"); deadlineType != "" {
		filter.DeadlineType = deadlineType
	}

	// Date range
	if dateFrom := q.Get("dateFrom"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := q.Get("dateTo"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filter.DateTo = &t
		}
	}

	// Include completed
	if includeCompleted := q.Get("includeCompleted"); includeCompleted == "true" {
		filter.IncludeCompleted = true
	}

	// Sorting
	if sortBy := q.Get("sortBy"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := q.Get("sortOrder"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	// Pagination
	if limit := q.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if offset := q.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	return filter
}

// Common errors
var (
	ErrEmptyBody = &jsonError{message: "Request body is empty"}
)

type jsonError struct {
	message string
}

func (e *jsonError) Error() string {
	return e.message
}
