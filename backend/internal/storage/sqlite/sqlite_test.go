package sqlite_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
	"github.com/LaV72/quest-todo/internal/storage/sqlite"
)

// TestSQLiteStorage runs the complete storage test suite against SQLite
func TestSQLiteStorage(t *testing.T) {
	suite := &storage.StorageTestSuite{
		Factory: func(t *testing.T) storage.Storage {
			// Create temporary database file
			tempDir := t.TempDir()
			dbPath := filepath.Join(tempDir, "test.db")

			// Create SQLite storage
			s, err := sqlite.New(dbPath)
			if err != nil {
				t.Fatalf("Failed to create SQLite storage: %v", err)
			}

			return s
		},
		Cleanup: func(t *testing.T, s storage.Storage) {
			// Close the storage
			if err := s.Close(); err != nil {
				t.Errorf("Failed to close storage: %v", err)
			}
			// TempDir is automatically cleaned up by testing framework
		},
	}

	// Run all tests
	suite.RunAllTests(t)
}

// TestSQLiteInMemory tests SQLite with in-memory database
func TestSQLiteInMemory(t *testing.T) {
	suite := &storage.StorageTestSuite{
		Factory: func(t *testing.T) storage.Storage {
			// Use in-memory database for faster tests
			s, err := sqlite.New(":memory:")
			if err != nil {
				t.Fatalf("Failed to create in-memory SQLite storage: %v", err)
			}
			return s
		},
		Cleanup: func(t *testing.T, s storage.Storage) {
			if err := s.Close(); err != nil {
				t.Errorf("Failed to close storage: %v", err)
			}
		},
	}

	// Run all tests
	suite.RunAllTests(t)
}

// TestSQLiteBackup tests the backup functionality
func TestSQLiteBackup(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create storage and add some data
	s, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer s.Close()

	// Create a test task
	ctx := context.Background()
	task := &models.Task{
		ID:        "backup-test",
		Title:     "Backup Test Task",
		Priority:  3,
		Status:    models.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = s.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Create backup
	backupPath := filepath.Join(tempDir, "backup.db")
	err = s.Backup(backupPath)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Backup file does not exist")
	}

	// Open backup and verify data
	backupStorage, err := sqlite.New(backupPath)
	if err != nil {
		t.Fatalf("Failed to open backup: %v", err)
	}
	defer backupStorage.Close()

	// Verify task exists in backup
	retrieved, err := backupStorage.GetTask(ctx, "backup-test")
	if err != nil {
		t.Fatalf("Failed to retrieve task from backup: %v", err)
	}
	if retrieved.Title != "Backup Test Task" {
		t.Errorf("Task title mismatch in backup: got %s, want 'Backup Test Task'", retrieved.Title)
	}
}

// TestSQLiteConcurrency tests concurrent operations
func TestSQLiteConcurrency(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "concurrent.db")

	s, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	now := time.Now()

	// Run concurrent writes
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			task := &models.Task{
				ID:        fmt.Sprintf("concurrent-%d", id),
				Title:     fmt.Sprintf("Concurrent Task %d", id),
				Priority:  3,
				Status:    models.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}

			err := s.CreateTask(ctx, task)
			if err != nil {
				errors <- fmt.Errorf("CreateTask %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify all tasks were created
	tasks, err := s.ListTasks(ctx, models.TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(tasks) != 10 {
		t.Errorf("Expected 10 tasks, got %d", len(tasks))
	}
}

// TestSQLiteTransactionRollback tests transaction rollback on error
func TestSQLiteTransactionRollback(t *testing.T) {
	s, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	now := time.Now()

	// Create a task with invalid data that should fail
	// (assuming validation happens at database level)
	task := &models.Task{
		ID:        "rollback-test",
		Title:     "Test Task",
		Priority:  10, // Invalid priority (CHECK constraint: 1-5)
		Status:    models.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.CreateTask(ctx, task)
	// Should fail due to CHECK constraint
	if err == nil {
		t.Error("Expected error for invalid priority, got nil")
	}

	// Verify task was not created
	_, err = s.GetTask(ctx, "rollback-test")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Expected ErrNotFound (rollback worked), got: %v", err)
	}
}
