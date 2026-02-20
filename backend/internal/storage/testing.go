package storage

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
)

// StorageTestSuite is a reusable test suite for any Storage implementation
// Pass a factory function that creates a fresh Storage instance for each test
type StorageTestSuite struct {
	// Factory creates a new storage instance for each test
	Factory func(t *testing.T) Storage
	// Cleanup is called after each test (optional)
	Cleanup func(t *testing.T, s Storage)
}

// RunAllTests runs the complete test suite
func (suite *StorageTestSuite) RunAllTests(t *testing.T) {
	t.Run("Tasks", func(t *testing.T) {
		suite.TestTaskCRUD(t)
		suite.TestTaskList(t)
		suite.TestTaskFilter(t)
		suite.TestTaskSearch(t)
		suite.TestTaskBulkOperations(t)
		suite.TestTaskNotFound(t)
	})

	t.Run("Objectives", func(t *testing.T) {
		suite.TestObjectiveCRUD(t)
	})

	t.Run("Categories", func(t *testing.T) {
		suite.TestCategoryCRUD(t)
	})

	t.Run("Stats", func(t *testing.T) {
		suite.TestStats(t)
	})

	t.Run("Lifecycle", func(t *testing.T) {
		suite.TestPingAndClose(t)
	})
}

// TestTaskCRUD tests basic task CRUD operations
func (suite *StorageTestSuite) TestTaskCRUD(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()

	// Create a task
	now := time.Now()
	task := &models.Task{
		ID:          "task-1",
		Title:       "Test Task",
		Description: "This is a test task",
		Priority:    4,
		Deadline: models.Deadline{
			Type: "short",
			Date: &now,
		},
		Category: "work",
		Status:   models.StatusActive,
		Notes:    "Some notes",
		Reward:   50,
		Order:    0,
		CreatedAt: now,
		UpdatedAt: now,
		Objectives: []models.Objective{
			{
				ID:        "obj-1",
				TaskID:    "task-1",
				Text:      "Step 1",
				Completed: false,
				Order:     0,
				CreatedAt: now,
			},
			{
				ID:        "obj-2",
				TaskID:    "task-1",
				Text:      "Step 2",
				Completed: true,
				Order:     1,
				CreatedAt: now,
			},
		},
		Tags: []string{"urgent", "important"},
	}

	// Test Create
	err := s.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Test Get
	retrieved, err := s.GetTask(ctx, "task-1")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	// Verify fields
	if retrieved.ID != task.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, task.ID)
	}
	if retrieved.Title != task.Title {
		t.Errorf("Title mismatch: got %s, want %s", retrieved.Title, task.Title)
	}
	if retrieved.Priority != task.Priority {
		t.Errorf("Priority mismatch: got %d, want %d", retrieved.Priority, task.Priority)
	}
	if len(retrieved.Objectives) != 2 {
		t.Errorf("Objectives count mismatch: got %d, want 2", len(retrieved.Objectives))
	}
	if len(retrieved.Tags) != 2 {
		t.Errorf("Tags count mismatch: got %d, want 2", len(retrieved.Tags))
	}

	// Test Update
	retrieved.Title = "Updated Task"
	retrieved.Priority = 5
	retrieved.UpdatedAt = time.Now()
	err = s.UpdateTask(ctx, retrieved)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// Verify update
	updated, err := s.GetTask(ctx, "task-1")
	if err != nil {
		t.Fatalf("GetTask after update failed: %v", err)
	}
	if updated.Title != "Updated Task" {
		t.Errorf("Title not updated: got %s, want 'Updated Task'", updated.Title)
	}
	if updated.Priority != 5 {
		t.Errorf("Priority not updated: got %d, want 5", updated.Priority)
	}

	// Test Delete
	err = s.DeleteTask(ctx, "task-1")
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Verify deletion
	_, err = s.GetTask(ctx, "task-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound after delete, got: %v", err)
	}
}

// TestTaskList tests listing tasks
func (suite *StorageTestSuite) TestTaskList(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()

	// Create multiple tasks
	now := time.Now()
	for i := 1; i <= 5; i++ {
		task := &models.Task{
			ID:          fmt.Sprintf("task-%d", i),
			Title:       fmt.Sprintf("Task %d", i),
			Description: "Test task",
			Priority:    i,
			Deadline:    models.Deadline{Type: "medium"},
			Status:      models.StatusActive,
			CreatedAt:   now.Add(time.Duration(i) * time.Minute),
			UpdatedAt:   now.Add(time.Duration(i) * time.Minute),
		}
		err := s.CreateTask(ctx, task)
		if err != nil {
			t.Fatalf("CreateTask %d failed: %v", i, err)
		}
	}

	// Test list all
	tasks, err := s.ListTasks(ctx, models.TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(tasks))
	}

	// Test pagination
	tasks, err = s.ListTasks(ctx, models.TaskFilter{
		Limit:  2,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListTasks with pagination failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks with limit=2, got %d", len(tasks))
	}

	// Test sorting
	tasks, err = s.ListTasks(ctx, models.TaskFilter{
		SortBy:    "priority",
		SortOrder: "desc",
	})
	if err != nil {
		t.Fatalf("ListTasks with sorting failed: %v", err)
	}
	if len(tasks) > 0 && tasks[0].Priority != 5 {
		t.Errorf("Expected highest priority first, got priority %d", tasks[0].Priority)
	}
}

// TestTaskFilter tests filtering tasks
func (suite *StorageTestSuite) TestTaskFilter(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()

	// Create tasks with different statuses and priorities
	now := time.Now()
	tasks := []*models.Task{
		{
			ID:        "task-1",
			Title:     "Active High Priority",
			Priority:  5,
			Status:    models.StatusActive,
			Category:  "work",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "task-2",
			Title:     "Active Low Priority",
			Priority:  2,
			Status:    models.StatusActive,
			Category:  "personal",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "task-3",
			Title:     "Completed Task",
			Priority:  3,
			Status:    models.StatusComplete,
			Category:  "work",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, task := range tasks {
		if err := s.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	// Filter by status
	filtered, err := s.ListTasks(ctx, models.TaskFilter{
		Status: []models.TaskStatus{models.StatusActive},
	})
	if err != nil {
		t.Fatalf("ListTasks with status filter failed: %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("Expected 2 active tasks, got %d", len(filtered))
	}

	// Filter by priority
	filtered, err = s.ListTasks(ctx, models.TaskFilter{
		Priority: []int{5},
	})
	if err != nil {
		t.Fatalf("ListTasks with priority filter failed: %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("Expected 1 task with priority 5, got %d", len(filtered))
	}

	// Filter by category (include completed)
	filtered, err = s.ListTasks(ctx, models.TaskFilter{
		Categories:       []string{"work"},
		IncludeCompleted: true,
	})
	if err != nil {
		t.Fatalf("ListTasks with category filter failed: %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("Expected 2 work tasks, got %d", len(filtered))
	}

	// Test count
	count, err := s.CountTasks(ctx, models.TaskFilter{
		Status: []models.TaskStatus{models.StatusActive},
	})
	if err != nil {
		t.Fatalf("CountTasks failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// TestTaskSearch tests searching tasks
func (suite *StorageTestSuite) TestTaskSearch(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()

	// Create tasks
	now := time.Now()
	tasks := []*models.Task{
		{
			ID:          "task-1",
			Title:       "Project Alpha",
			Description: "Working on the Alpha project",
			Priority:    3,
			Status:      models.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "task-2",
			Title:       "Project Beta",
			Description: "Working on the Beta project",
			Priority:    3,
			Status:      models.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "task-3",
			Title:       "Review Code",
			Description: "Review pull requests",
			Priority:    3,
			Status:      models.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, task := range tasks {
		if err := s.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	// Search by title
	results, err := s.SearchTasks(ctx, "Project")
	if err != nil {
		t.Fatalf("SearchTasks failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Project', got %d", len(results))
	}

	// Search by description
	results, err = s.SearchTasks(ctx, "pull requests")
	if err != nil {
		t.Fatalf("SearchTasks failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'pull requests', got %d", len(results))
	}
}

// TestTaskBulkOperations tests bulk create/update/delete
func (suite *StorageTestSuite) TestTaskBulkOperations(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()
	now := time.Now()

	// Bulk create
	tasks := []*models.Task{
		{
			ID:        "bulk-1",
			Title:     "Bulk Task 1",
			Priority:  3,
			Status:    models.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "bulk-2",
			Title:     "Bulk Task 2",
			Priority:  4,
			Status:    models.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "bulk-3",
			Title:     "Bulk Task 3",
			Priority:  5,
			Status:    models.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err := s.CreateTasksBulk(ctx, tasks)
	if err != nil {
		t.Fatalf("CreateTasksBulk failed: %v", err)
	}

	// Verify creation
	retrieved, err := s.GetTask(ctx, "bulk-1")
	if err != nil {
		t.Fatalf("GetTask after bulk create failed: %v", err)
	}
	if retrieved.Title != "Bulk Task 1" {
		t.Errorf("Task not created correctly")
	}

	// Bulk update
	for _, task := range tasks {
		task.Status = models.StatusComplete
		task.UpdatedAt = time.Now()
	}
	err = s.UpdateTasksBulk(ctx, tasks)
	if err != nil {
		t.Fatalf("UpdateTasksBulk failed: %v", err)
	}

	// Verify update
	retrieved, err = s.GetTask(ctx, "bulk-1")
	if err != nil {
		t.Fatalf("GetTask after bulk update failed: %v", err)
	}
	if retrieved.Status != models.StatusComplete {
		t.Errorf("Task status not updated, got %s", retrieved.Status)
	}

	// Bulk delete
	ids := []string{"bulk-1", "bulk-2", "bulk-3"}
	err = s.DeleteTasksBulk(ctx, ids)
	if err != nil {
		t.Fatalf("DeleteTasksBulk failed: %v", err)
	}

	// Verify deletion
	_, err = s.GetTask(ctx, "bulk-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound after bulk delete, got: %v", err)
	}
}

// TestTaskNotFound tests error handling for non-existent tasks
func (suite *StorageTestSuite) TestTaskNotFound(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()

	// Get non-existent task
	_, err := s.GetTask(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}

	// Update non-existent task
	task := &models.Task{
		ID:        "nonexistent",
		Title:     "Test",
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	err = s.UpdateTask(ctx, task)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound on update, got: %v", err)
	}

	// Delete non-existent task
	err = s.DeleteTask(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound on delete, got: %v", err)
	}
}

// TestObjectiveCRUD tests objective operations
func (suite *StorageTestSuite) TestObjectiveCRUD(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()
	now := time.Now()

	// Create a task first
	task := &models.Task{
		ID:        "task-obj",
		Title:     "Task with Objectives",
		Priority:  3,
		Status:    models.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := s.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Add objective
	obj := &models.Objective{
		ID:        "obj-1",
		TaskID:    "task-obj",
		Text:      "New Objective",
		Completed: false,
		Order:     0,
		CreatedAt: now,
	}
	err = s.AddObjective(ctx, "task-obj", obj)
	if err != nil {
		t.Fatalf("AddObjective failed: %v", err)
	}

	// Verify addition
	retrieved, err := s.GetTask(ctx, "task-obj")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if len(retrieved.Objectives) != 1 {
		t.Errorf("Expected 1 objective, got %d", len(retrieved.Objectives))
	}

	// Update objective
	obj.Text = "Updated Objective"
	obj.Completed = true
	err = s.UpdateObjective(ctx, "task-obj", "obj-1", obj)
	if err != nil {
		t.Fatalf("UpdateObjective failed: %v", err)
	}

	// Verify update
	retrieved, err = s.GetTask(ctx, "task-obj")
	if err != nil {
		t.Fatalf("GetTask after update failed: %v", err)
	}
	if retrieved.Objectives[0].Text != "Updated Objective" {
		t.Errorf("Objective text not updated")
	}
	if !retrieved.Objectives[0].Completed {
		t.Errorf("Objective not marked as completed")
	}

	// Delete objective
	err = s.DeleteObjective(ctx, "task-obj", "obj-1")
	if err != nil {
		t.Fatalf("DeleteObjective failed: %v", err)
	}

	// Verify deletion
	retrieved, err = s.GetTask(ctx, "task-obj")
	if err != nil {
		t.Fatalf("GetTask after delete failed: %v", err)
	}
	if len(retrieved.Objectives) != 0 {
		t.Errorf("Expected 0 objectives after delete, got %d", len(retrieved.Objectives))
	}
}

// TestCategoryCRUD tests category operations
func (suite *StorageTestSuite) TestCategoryCRUD(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()
	now := time.Now()

	// Create category
	cat := &models.Category{
		ID:        "cat-1",
		Name:      "Work",
		Color:     "#3A7F8F",
		Icon:      "briefcase",
		Type:      "main",
		Order:     0,
		CreatedAt: now,
	}
	err := s.CreateCategory(ctx, cat)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	// Get category
	retrieved, err := s.GetCategory(ctx, "cat-1")
	if err != nil {
		t.Fatalf("GetCategory failed: %v", err)
	}
	if retrieved.Name != "Work" {
		t.Errorf("Category name mismatch: got %s, want Work", retrieved.Name)
	}

	// List categories
	cats, err := s.ListCategories(ctx)
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	if len(cats) != 1 {
		t.Errorf("Expected 1 category, got %d", len(cats))
	}

	// Update category
	cat.Name = "Professional"
	cat.Color = "#FF0000"
	err = s.UpdateCategory(ctx, cat)
	if err != nil {
		t.Fatalf("UpdateCategory failed: %v", err)
	}

	// Verify update
	updated, err := s.GetCategory(ctx, "cat-1")
	if err != nil {
		t.Fatalf("GetCategory after update failed: %v", err)
	}
	if updated.Name != "Professional" {
		t.Errorf("Category name not updated")
	}

	// Delete category
	err = s.DeleteCategory(ctx, "cat-1")
	if err != nil {
		t.Fatalf("DeleteCategory failed: %v", err)
	}

	// Verify deletion
	_, err = s.GetCategory(ctx, "cat-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound after delete, got: %v", err)
	}
}

// TestStats tests statistics operations
func (suite *StorageTestSuite) TestStats(t *testing.T) {
	s := suite.Factory(t)
	if suite.Cleanup != nil {
		defer suite.Cleanup(t, s)
	}

	ctx := context.Background()
	now := time.Now()

	// Create tasks with different statuses
	tasks := []*models.Task{
		{
			ID:        "stat-1",
			Title:     "Active 1",
			Status:    models.StatusActive,
			Priority:  5,
			Category:  "work",
			Reward:    10,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "stat-2",
			Title:     "Active 2",
			Status:    models.StatusActive,
			Priority:  4,
			Category:  "work",
			Reward:    20,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "stat-3",
			Title:       "Completed",
			Status:      models.StatusComplete,
			Priority:    3,
			Category:    "personal",
			Reward:      30,
			CreatedAt:   now,
			UpdatedAt:   now,
			CompletedAt: &now,
		},
	}

	for _, task := range tasks {
		if err := s.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	// Get stats
	stats, err := s.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalTasks != 3 {
		t.Errorf("Expected 3 total tasks, got %d", stats.TotalTasks)
	}
	if stats.ActiveTasks != 2 {
		t.Errorf("Expected 2 active tasks, got %d", stats.ActiveTasks)
	}
	if stats.CompletedTasks != 1 {
		t.Errorf("Expected 1 completed task, got %d", stats.CompletedTasks)
	}
	if stats.TotalRewards != 30 { // Only completed task reward
		t.Errorf("Expected 30 total rewards, got %d", stats.TotalRewards)
	}

	// Test category stats
	catStats, err := s.GetCategoryStats(ctx)
	if err != nil {
		t.Fatalf("GetCategoryStats failed: %v", err)
	}
	if len(catStats) != 2 { // work and personal
		t.Errorf("Expected 2 categories in stats, got %d", len(catStats))
	}
}

// TestPingAndClose tests lifecycle methods
func (suite *StorageTestSuite) TestPingAndClose(t *testing.T) {
	s := suite.Factory(t)

	// Test ping
	err := s.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// Test close
	err = s.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
