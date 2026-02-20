package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// MockTaskService implements a minimal task service for testing
type MockTaskService struct {
	CreateTaskFunc      func(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error)
	GetTaskFunc         func(ctx context.Context, id string) (*models.Task, error)
	UpdateTaskFunc      func(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error)
	DeleteTaskFunc      func(ctx context.Context, id string) error
	ListTasksFunc       func(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error)
	CountTasksFunc      func(ctx context.Context, filter models.TaskFilter) (int, error)
	SearchTasksFunc     func(ctx context.Context, query string) ([]*models.Task, error)
	CreateTasksBulkFunc func(ctx context.Context, tasks []models.TaskCreateRequest) ([]*models.Task, error)
	UpdateTasksBulkFunc func(ctx context.Context, updates map[string]models.TaskUpdateRequest) error
	DeleteTasksBulkFunc func(ctx context.Context, ids []string) error
	CompleteTaskFunc    func(ctx context.Context, id string) (*models.Task, error)
	FailTaskFunc        func(ctx context.Context, id string) (*models.Task, error)
	ReactivateTaskFunc  func(ctx context.Context, id string) (*models.Task, error)
	ReorderTasksFunc    func(ctx context.Context, ids []string) error
}

func (m *MockTaskService) CreateTask(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error) {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, req)
	}
	return &models.Task{ID: "task-1", Title: req.Title}, nil
}

func (m *MockTaskService) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if m.GetTaskFunc != nil {
		return m.GetTaskFunc(ctx, id)
	}
	return &models.Task{ID: id, Title: "Test Task"}, nil
}

func (m *MockTaskService) UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error) {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, id, req)
	}
	return &models.Task{ID: id, Title: "Updated Task"}, nil
}

func (m *MockTaskService) DeleteTask(ctx context.Context, id string) error {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(ctx, id)
	}
	return nil
}

func (m *MockTaskService) ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error) {
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx, filter)
	}
	return []*models.Task{{ID: "task-1"}}, nil
}

func (m *MockTaskService) CountTasks(ctx context.Context, filter models.TaskFilter) (int, error) {
	if m.CountTasksFunc != nil {
		return m.CountTasksFunc(ctx, filter)
	}
	return 1, nil
}

func (m *MockTaskService) SearchTasks(ctx context.Context, query string) ([]*models.Task, error) {
	if m.SearchTasksFunc != nil {
		return m.SearchTasksFunc(ctx, query)
	}
	return []*models.Task{{ID: "task-1"}}, nil
}

func (m *MockTaskService) CreateTasksBulk(ctx context.Context, tasks []models.TaskCreateRequest) ([]*models.Task, error) {
	if m.CreateTasksBulkFunc != nil {
		return m.CreateTasksBulkFunc(ctx, tasks)
	}
	return []*models.Task{{ID: "task-1"}}, nil
}

func (m *MockTaskService) UpdateTasksBulk(ctx context.Context, updates map[string]models.TaskUpdateRequest) error {
	if m.UpdateTasksBulkFunc != nil {
		return m.UpdateTasksBulkFunc(ctx, updates)
	}
	return nil
}

func (m *MockTaskService) DeleteTasksBulk(ctx context.Context, ids []string) error {
	if m.DeleteTasksBulkFunc != nil {
		return m.DeleteTasksBulkFunc(ctx, ids)
	}
	return nil
}

func (m *MockTaskService) CompleteTask(ctx context.Context, id string) (*models.Task, error) {
	if m.CompleteTaskFunc != nil {
		return m.CompleteTaskFunc(ctx, id)
	}
	return &models.Task{ID: id, Status: models.StatusComplete}, nil
}

func (m *MockTaskService) FailTask(ctx context.Context, id string) (*models.Task, error) {
	if m.FailTaskFunc != nil {
		return m.FailTaskFunc(ctx, id)
	}
	return &models.Task{ID: id, Status: models.StatusFailed}, nil
}

func (m *MockTaskService) ReactivateTask(ctx context.Context, id string) (*models.Task, error) {
	if m.ReactivateTaskFunc != nil {
		return m.ReactivateTaskFunc(ctx, id)
	}
	return &models.Task{ID: id, Status: models.StatusActive}, nil
}

func (m *MockTaskService) ReorderTasks(ctx context.Context, ids []string) error {
	if m.ReorderTasksFunc != nil {
		return m.ReorderTasksFunc(ctx, ids)
	}
	return nil
}

// MockObjectiveService implements a minimal objective service for testing
type MockObjectiveService struct {
	CreateObjectiveFunc func(ctx context.Context, taskID string, req models.ObjectiveRequest) (*models.Objective, error)
	UpdateObjectiveFunc func(ctx context.Context, id string, req models.ObjectiveUpdateRequest) (*models.Objective, error)
	DeleteObjectiveFunc func(ctx context.Context, id string) error
	ToggleObjectiveFunc func(ctx context.Context, id string) (*models.Objective, error)
}

func (m *MockObjectiveService) CreateObjective(ctx context.Context, taskID string, req models.ObjectiveRequest) (*models.Objective, error) {
	if m.CreateObjectiveFunc != nil {
		return m.CreateObjectiveFunc(ctx, taskID, req)
	}
	return &models.Objective{ID: "obj-1", TaskID: taskID, Text: req.Text}, nil
}

func (m *MockObjectiveService) UpdateObjective(ctx context.Context, id string, req models.ObjectiveUpdateRequest) (*models.Objective, error) {
	if m.UpdateObjectiveFunc != nil {
		return m.UpdateObjectiveFunc(ctx, id, req)
	}
	return &models.Objective{ID: id, Text: "Updated"}, nil
}

func (m *MockObjectiveService) DeleteObjective(ctx context.Context, id string) error {
	if m.DeleteObjectiveFunc != nil {
		return m.DeleteObjectiveFunc(ctx, id)
	}
	return nil
}

func (m *MockObjectiveService) ToggleObjective(ctx context.Context, id string) (*models.Objective, error) {
	if m.ToggleObjectiveFunc != nil {
		return m.ToggleObjectiveFunc(ctx, id)
	}
	return &models.Objective{ID: id, Completed: true}, nil
}

// MockCategoryService implements a minimal category service for testing
type MockCategoryService struct {
	CreateCategoryFunc func(ctx context.Context, req models.CategoryCreateRequest) (*models.Category, error)
	GetCategoryFunc    func(ctx context.Context, id string) (*models.Category, error)
	UpdateCategoryFunc func(ctx context.Context, id string, req models.CategoryUpdateRequest) (*models.Category, error)
	DeleteCategoryFunc func(ctx context.Context, id string) error
	ListCategoriesFunc func(ctx context.Context) ([]*models.Category, error)
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, req models.CategoryCreateRequest) (*models.Category, error) {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, req)
	}
	return &models.Category{ID: "cat-1", Name: req.Name}, nil
}

func (m *MockCategoryService) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	if m.GetCategoryFunc != nil {
		return m.GetCategoryFunc(ctx, id)
	}
	return &models.Category{ID: id, Name: "Test Category"}, nil
}

func (m *MockCategoryService) UpdateCategory(ctx context.Context, id string, req models.CategoryUpdateRequest) (*models.Category, error) {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, id, req)
	}
	return &models.Category{ID: id, Name: "Updated"}, nil
}

func (m *MockCategoryService) DeleteCategory(ctx context.Context, id string) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, id)
	}
	return nil
}

func (m *MockCategoryService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	if m.ListCategoriesFunc != nil {
		return m.ListCategoriesFunc(ctx)
	}
	return []*models.Category{{ID: "cat-1"}}, nil
}

// MockStatsService implements a minimal stats service for testing
type MockStatsService struct {
	GetStatsFunc         func(ctx context.Context) (*models.Stats, error)
	GetCategoryStatsFunc func(ctx context.Context) ([]models.CategoryStat, error)
}

func (m *MockStatsService) GetStats(ctx context.Context) (*models.Stats, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx)
	}
	return &models.Stats{TotalTasks: 10}, nil
}

func (m *MockStatsService) GetCategoryStats(ctx context.Context) ([]models.CategoryStat, error) {
	if m.GetCategoryStatsFunc != nil {
		return m.GetCategoryStatsFunc(ctx)
	}
	return []models.CategoryStat{{CategoryID: "cat-1", TotalTasks: 5}}, nil
}

// createTestAPI creates an API instance with mock services for testing
func createTestAPI() *API {
	return NewAPI(
		&MockTaskService{},
		&MockObjectiveService{},
		&MockCategoryService{},
		&MockStatsService{},
		validator.New(),
		"test-version",
	)
}

// TestNewRouter verifies router initialization
func TestNewRouter(t *testing.T) {
	t.Run("creates router with default config", func(t *testing.T) {
		api := createTestAPI()
		config := RouterConfig{
			AllowedOrigins: []string{"*"},
			EnableCORS:     false,
			EnableLogging:  false,
		}

		router := NewRouter(api, config)

		assert.NotNil(t, router)
	})

	t.Run("creates router with CORS enabled", func(t *testing.T) {
		api := createTestAPI()
		config := RouterConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			EnableCORS:     true,
			EnableLogging:  false,
		}

		router := NewRouter(api, config)

		assert.NotNil(t, router)
	})

	t.Run("creates router with logging enabled", func(t *testing.T) {
		api := createTestAPI()
		config := RouterConfig{
			EnableCORS:    false,
			EnableLogging: true,
		}

		router := NewRouter(api, config)

		assert.NotNil(t, router)
	})
}

// TestHealthRoutes tests health check and version endpoints
func TestHealthRoutes(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	t.Run("GET /health returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "success")
	})

	t.Run("GET /version returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/version", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "test-version")
	})
}

// TestTaskRoutes tests task endpoint routing
func TestTaskRoutes(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	t.Run("POST /api/tasks routes to CreateTask", func(t *testing.T) {
		body := `{"title":"Test","priority":3}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("GET /api/tasks routes to ListTasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/tasks?q=query routes to SearchTasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks?q=test", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/tasks/{id} routes to GetTask", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/task-123", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("PUT /api/tasks/{id} routes to UpdateTask", func(t *testing.T) {
		body := `{"title":"Updated"}`
		req := httptest.NewRequest(http.MethodPut, "/api/tasks/task-123", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("DELETE /api/tasks/{id} routes to DeleteTask", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/task-123", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("POST /api/tasks/{id}/complete routes to CompleteTask", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/task-123/complete", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST /api/tasks/{id}/fail routes to FailTask", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/task-123/fail", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST /api/tasks/{id}/reactivate routes to ReactivateTask", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/task-123/reactivate", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST /api/tasks/bulk routes to CreateTasksBulk", func(t *testing.T) {
		body := `{"tasks":[{"title":"Task1","priority":3}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/bulk", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("DELETE /api/tasks/bulk routes to DeleteTasksBulk", func(t *testing.T) {
		body := `{"ids":["task-1","task-2"]}`
		req := httptest.NewRequest(http.MethodDelete, "/api/tasks/bulk", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("POST /api/tasks/reorder routes to ReorderTasks", func(t *testing.T) {
		body := `{"ids":["task-1","task-2"]}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/reorder", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}

// TestObjectiveRoutes tests objective endpoint routing
func TestObjectiveRoutes(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	t.Run("POST /api/tasks/{taskId}/objectives routes to CreateObjective", func(t *testing.T) {
		body := `{"text":"Objective 1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/task-123/objectives", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("PUT /api/objectives/{id} routes to UpdateObjective", func(t *testing.T) {
		body := `{"text":"Updated"}`
		req := httptest.NewRequest(http.MethodPut, "/api/objectives/obj-123", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("DELETE /api/objectives/{id} routes to DeleteObjective", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/objectives/obj-123", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("POST /api/objectives/{id}/toggle routes to ToggleObjective", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/objectives/obj-123/toggle", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// TestCategoryRoutes tests category endpoint routing
func TestCategoryRoutes(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	t.Run("POST /api/categories routes to CreateCategory", func(t *testing.T) {
		body := `{"name":"Category","color":"#FF0000","type":"main"}`
		req := httptest.NewRequest(http.MethodPost, "/api/categories", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("GET /api/categories routes to ListCategories", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/categories/{id} routes to GetCategory", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/categories/cat-123", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("PUT /api/categories/{id} routes to UpdateCategory", func(t *testing.T) {
		body := `{"name":"Updated"}`
		req := httptest.NewRequest(http.MethodPut, "/api/categories/cat-123", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("DELETE /api/categories/{id} routes to DeleteCategory", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/categories/cat-123", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}

// TestStatsRoutes tests stats endpoint routing
func TestStatsRoutes(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	t.Run("GET /api/stats routes to GetStats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/stats/categories routes to GetCategoryStats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stats/categories", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// TestMethodNotAllowed tests that invalid methods return 405
func TestMethodNotAllowed(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"PATCH on tasks collection", http.MethodPatch, "/api/tasks"},
		{"DELETE on tasks collection", http.MethodDelete, "/api/tasks"},
		{"POST on task detail", http.MethodPost, "/api/tasks/task-123"},
		{"GET on categories collection with POST", http.MethodPost, "/api/stats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
			assert.Contains(t, rec.Body.String(), "METHOD_NOT_ALLOWED")
		})
	}
}

// TestCORSIntegration tests CORS middleware integration
func TestCORSIntegration(t *testing.T) {
	api := createTestAPI()
	config := RouterConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		EnableCORS:     true,
		EnableLogging:  false,
	}
	router := NewRouter(api, config)

	t.Run("CORS headers added for allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS preflight request handled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/tasks", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	})
}

// TestRequestIDIntegration tests request ID middleware integration
func TestRequestIDIntegration(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	t.Run("request ID added to response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
	})

	t.Run("existing request ID preserved", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-Request-ID", "test-id-123")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, "test-id-123", rec.Header().Get("X-Request-ID"))
	})
}

// TestContentTypeIntegration tests content type middleware integration
func TestContentTypeIntegration(t *testing.T) {
	api := createTestAPI()
	router := NewRouter(api, RouterConfig{})

	t.Run("rejects non-JSON content type for POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader("data"))
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
		assert.Contains(t, rec.Body.String(), "UNSUPPORTED_MEDIA_TYPE")
	})

	t.Run("allows JSON content type for POST", func(t *testing.T) {
		body := `{"title":"Test","priority":3}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusUnsupportedMediaType, rec.Code)
	})
}

// TestRecoveryIntegration tests panic recovery middleware integration
func TestRecoveryIntegration(t *testing.T) {
	// Create API with handler that panics
	api := createTestAPI()
	api.TaskService = &MockTaskService{
		ListTasksFunc: func(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error) {
			panic("test panic")
		},
	}

	router := NewRouter(api, RouterConfig{})

	t.Run("recovers from panic and returns 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		rec := httptest.NewRecorder()

		// Should not panic
		assert.NotPanics(t, func() {
			router.ServeHTTP(rec, req)
		})

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "INTERNAL_ERROR")
	})
}

// TestMiddlewareChainOrder tests that middleware executes in correct order
func TestMiddlewareChainOrder(t *testing.T) {
	api := createTestAPI()
	config := RouterConfig{
		AllowedOrigins: []string{"*"},
		EnableCORS:     true,
		EnableLogging:  false,
	}
	router := NewRouter(api, config)

	t.Run("middleware chain executes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		// Verify multiple middleware effects
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"), "RequestID middleware should set header")
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"), "CORS middleware should set header")
		assert.Equal(t, http.StatusOK, rec.Code, "Handler should execute successfully")
	})
}
