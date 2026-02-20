package service

import (
	"context"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
)

// MockStorage is a mock implementation of storage.Storage for testing
type MockStorage struct {
	// Data stores
	Tasks      map[string]*models.Task
	Objectives map[string]*models.Objective
	Categories map[string]*models.Category

	// Function hooks for testing
	CreateTaskFunc       func(ctx context.Context, task *models.Task) error
	GetTaskFunc          func(ctx context.Context, id string) (*models.Task, error)
	UpdateTaskFunc       func(ctx context.Context, task *models.Task) error
	DeleteTaskFunc       func(ctx context.Context, id string) error
	ListTasksFunc        func(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error)
	CountTasksFunc       func(ctx context.Context, filter models.TaskFilter) (int, error)
	SearchTasksFunc      func(ctx context.Context, query string) ([]*models.Task, error)
	CreateTasksBulkFunc  func(ctx context.Context, tasks []*models.Task) error
	AddObjectiveFunc     func(ctx context.Context, taskID string, obj *models.Objective) error
	UpdateObjectiveFunc  func(ctx context.Context, taskID, objID string, obj *models.Objective) error
	DeleteObjectiveFunc  func(ctx context.Context, taskID, objID string) error
	CreateCategoryFunc   func(ctx context.Context, cat *models.Category) error
	GetCategoryFunc      func(ctx context.Context, id string) (*models.Category, error)
	ListCategoriesFunc   func(ctx context.Context) ([]*models.Category, error)
	GetStatsFunc         func(ctx context.Context) (*models.Stats, error)
	GetCategoryStatsFunc func(ctx context.Context) (map[string]*models.CategoryStat, error)
}

// NewMockStorage creates a new mock storage with empty data stores
func NewMockStorage() *MockStorage {
	return &MockStorage{
		Tasks:      make(map[string]*models.Task),
		Objectives: make(map[string]*models.Objective),
		Categories: make(map[string]*models.Category),
	}
}

// Default implementations that use in-memory storage

func (m *MockStorage) CreateTask(ctx context.Context, task *models.Task) error {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, task)
	}
	m.Tasks[task.ID] = task
	return nil
}

func (m *MockStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if m.GetTaskFunc != nil {
		return m.GetTaskFunc(ctx, id)
	}
	task, ok := m.Tasks[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return task, nil
}

func (m *MockStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, task)
	}
	if _, ok := m.Tasks[task.ID]; !ok {
		return storage.ErrNotFound
	}
	m.Tasks[task.ID] = task
	return nil
}

func (m *MockStorage) DeleteTask(ctx context.Context, id string) error {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(ctx, id)
	}
	if _, ok := m.Tasks[id]; !ok {
		return storage.ErrNotFound
	}
	delete(m.Tasks, id)
	return nil
}

func (m *MockStorage) ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error) {
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx, filter)
	}
	tasks := make([]*models.Task, 0, len(m.Tasks))
	for _, task := range m.Tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *MockStorage) CountTasks(ctx context.Context, filter models.TaskFilter) (int, error) {
	if m.CountTasksFunc != nil {
		return m.CountTasksFunc(ctx, filter)
	}
	return len(m.Tasks), nil
}

func (m *MockStorage) SearchTasks(ctx context.Context, query string) ([]*models.Task, error) {
	if m.SearchTasksFunc != nil {
		return m.SearchTasksFunc(ctx, query)
	}
	return []*models.Task{}, nil
}

func (m *MockStorage) CreateTasksBulk(ctx context.Context, tasks []*models.Task) error {
	if m.CreateTasksBulkFunc != nil {
		return m.CreateTasksBulkFunc(ctx, tasks)
	}
	for _, task := range tasks {
		m.Tasks[task.ID] = task
	}
	return nil
}

func (m *MockStorage) UpdateTasksBulk(ctx context.Context, tasks []*models.Task) error {
	for _, task := range tasks {
		m.Tasks[task.ID] = task
	}
	return nil
}

func (m *MockStorage) DeleteTasksBulk(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.Tasks, id)
	}
	return nil
}

func (m *MockStorage) UpdateTaskStatus(ctx context.Context, id string, status models.TaskStatus) error {
	if task, ok := m.Tasks[id]; ok {
		task.Status = status
		return nil
	}
	return storage.ErrNotFound
}

func (m *MockStorage) ReorderTasks(ctx context.Context, ids []string) error {
	return nil
}

func (m *MockStorage) AddObjective(ctx context.Context, taskID string, obj *models.Objective) error {
	if m.AddObjectiveFunc != nil {
		return m.AddObjectiveFunc(ctx, taskID, obj)
	}
	task, ok := m.Tasks[taskID]
	if !ok {
		return storage.ErrNotFound
	}
	task.Objectives = append(task.Objectives, *obj)
	m.Objectives[obj.ID] = obj
	return nil
}

func (m *MockStorage) UpdateObjective(ctx context.Context, taskID, objID string, obj *models.Objective) error {
	if m.UpdateObjectiveFunc != nil {
		return m.UpdateObjectiveFunc(ctx, taskID, objID, obj)
	}
	task, ok := m.Tasks[taskID]
	if !ok {
		return storage.ErrNotFound
	}
	for i := range task.Objectives {
		if task.Objectives[i].ID == objID {
			task.Objectives[i] = *obj
			m.Objectives[objID] = obj
			return nil
		}
	}
	return storage.ErrNotFound
}

func (m *MockStorage) DeleteObjective(ctx context.Context, taskID, objID string) error {
	if m.DeleteObjectiveFunc != nil {
		return m.DeleteObjectiveFunc(ctx, taskID, objID)
	}
	task, ok := m.Tasks[taskID]
	if !ok {
		return storage.ErrNotFound
	}
	for i := range task.Objectives {
		if task.Objectives[i].ID == objID {
			task.Objectives = append(task.Objectives[:i], task.Objectives[i+1:]...)
			delete(m.Objectives, objID)
			return nil
		}
	}
	return storage.ErrNotFound
}

func (m *MockStorage) CreateCategory(ctx context.Context, cat *models.Category) error {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, cat)
	}
	m.Categories[cat.ID] = cat
	return nil
}

func (m *MockStorage) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	if m.GetCategoryFunc != nil {
		return m.GetCategoryFunc(ctx, id)
	}
	cat, ok := m.Categories[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return cat, nil
}

func (m *MockStorage) ListCategories(ctx context.Context) ([]*models.Category, error) {
	if m.ListCategoriesFunc != nil {
		return m.ListCategoriesFunc(ctx)
	}
	categories := make([]*models.Category, 0, len(m.Categories))
	for _, cat := range m.Categories {
		categories = append(categories, cat)
	}
	return categories, nil
}

func (m *MockStorage) UpdateCategory(ctx context.Context, cat *models.Category) error {
	if _, ok := m.Categories[cat.ID]; !ok {
		return storage.ErrNotFound
	}
	m.Categories[cat.ID] = cat
	return nil
}

func (m *MockStorage) DeleteCategory(ctx context.Context, id string) error {
	if _, ok := m.Categories[id]; !ok {
		return storage.ErrNotFound
	}
	delete(m.Categories, id)
	return nil
}

func (m *MockStorage) GetStats(ctx context.Context) (*models.Stats, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx)
	}
	return &models.Stats{
		TotalTasks:  len(m.Tasks),
		ActiveTasks: 0,
	}, nil
}

func (m *MockStorage) GetDailyStats(ctx context.Context, from, to time.Time) ([]*models.DailyStat, error) {
	return []*models.DailyStat{}, nil
}

func (m *MockStorage) GetCategoryStats(ctx context.Context) (map[string]*models.CategoryStat, error) {
	if m.GetCategoryStatsFunc != nil {
		return m.GetCategoryStatsFunc(ctx)
	}
	return make(map[string]*models.CategoryStat), nil
}

func (m *MockStorage) Ping() error {
	return nil
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) Backup(dest string) error {
	return nil
}

// FixedClock is a clock that always returns the same time (for testing)
type FixedClock struct {
	Time time.Time
}

func (c *FixedClock) Now() time.Time {
	return c.Time
}

// NewFixedClock creates a clock with a specific time
func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{Time: t}
}

// FixedIDGenerator generates IDs from a predefined list (for testing)
type FixedIDGenerator struct {
	IDs []string
	idx int
}

func (g *FixedIDGenerator) Generate() string {
	if g.idx >= len(g.IDs) {
		return "default-id"
	}
	id := g.IDs[g.idx]
	g.idx++
	return id
}

// NewFixedIDGenerator creates an ID generator with predefined IDs
func NewFixedIDGenerator(ids ...string) *FixedIDGenerator {
	return &FixedIDGenerator{IDs: ids}
}
