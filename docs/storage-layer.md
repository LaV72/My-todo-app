# Storage Layer

Complete documentation for the pluggable storage abstraction layer.

## Overview

The storage layer provides a clean abstraction between business logic and data persistence, allowing different storage implementations to be swapped without affecting the rest of the application.

```
Service Layer
      ↓
Storage Interface (Abstract Contract)
      ↓
┌─────┴─────┬──────────┐
│           │          │
SQLite     JSON     Memory
```

## Design Principles

### 1. Interface Segregation

Storage is divided into logical interfaces:

```go
type Storage interface {
    TaskStorage      // Task operations
    CategoryStorage  // Category operations
    StatsStorage     // Statistics queries

    // Lifecycle
    Close() error
    Ping() error
    Backup(dest string) error
}
```

Each interface focuses on a specific domain, making implementations easier to maintain.

### 2. Implementation Independence

Business logic depends only on interfaces, never concrete implementations:

```go
// Service depends on interface
type TaskService struct {
    store storage.Storage  // Interface, not concrete type
}

// Can use any implementation
service := NewTaskService(sqlite.New(path))
service := NewTaskService(json.New(path))
service := NewTaskService(memory.New())
```

### 3. Factory Pattern

Storage instances are created through a factory function:

```go
store, err := storage.New(storage.Config{
    Type: "sqlite",  // or "json", "memory"
    Path: "./data/todo.db",
})
```

---

## Storage Interface

### Complete Interface Definition

```go
package storage

import (
    "context"
    "time"
)

// Storage is the main interface
type Storage interface {
    TaskStorage
    CategoryStorage
    StatsStorage

    Close() error
    Ping() error
    Backup(dest string) error
}

// TaskStorage handles all task operations
type TaskStorage interface {
    // Basic CRUD
    CreateTask(ctx context.Context, task *Task) error
    GetTask(ctx context.Context, id string) (*Task, error)
    UpdateTask(ctx context.Context, task *Task) error
    DeleteTask(ctx context.Context, id string) error

    // Queries
    ListTasks(ctx context.Context, filter TaskFilter) ([]*Task, error)
    CountTasks(ctx context.Context, filter TaskFilter) (int, error)
    SearchTasks(ctx context.Context, query string) ([]*Task, error)

    // Bulk operations
    CreateTasksBulk(ctx context.Context, tasks []*Task) error
    UpdateTasksBulk(ctx context.Context, tasks []*Task) error
    DeleteTasksBulk(ctx context.Context, ids []string) error

    // Task-specific
    UpdateTaskStatus(ctx context.Context, id string, status TaskStatus) error
    ReorderTasks(ctx context.Context, ids []string) error

    // Objectives
    AddObjective(ctx context.Context, taskID string, obj *Objective) error
    UpdateObjective(ctx context.Context, taskID, objID string, obj *Objective) error
    DeleteObjective(ctx context.Context, taskID, objID string) error
}

// CategoryStorage handles category operations
type CategoryStorage interface {
    CreateCategory(ctx context.Context, cat *Category) error
    GetCategory(ctx context.Context, id string) (*Category, error)
    ListCategories(ctx context.Context) ([]*Category, error)
    UpdateCategory(ctx context.Context, cat *Category) error
    DeleteCategory(ctx context.Context, id string) error
}

// StatsStorage handles statistics
type StatsStorage interface {
    GetStats(ctx context.Context) (*Stats, error)
    GetDailyStats(ctx context.Context, from, to time.Time) ([]*DailyStat, error)
    GetCategoryStats(ctx context.Context) (map[string]*CategoryStat, error)
}
```

### TaskFilter

```go
type TaskFilter struct {
    // Filters
    Status      []TaskStatus
    Priority    []int
    Categories  []string
    Tags        []string
    DeadlineType string
    DateFrom    *time.Time
    DateTo      *time.Time
    IncludeCompleted bool

    // Sorting
    SortBy      string  // "priority", "deadline", "created_at", etc.
    SortOrder   string  // "asc", "desc"

    // Pagination
    Limit       int
    Offset      int
}
```

---

## Storage Errors

### Standard Errors

```go
package storage

import "errors"

var (
    ErrNotFound    = errors.New("resource not found")
    ErrDuplicate   = errors.New("duplicate resource")
    ErrConflict    = errors.New("resource conflict")
    ErrInvalid     = errors.New("invalid data")
    ErrInternal    = errors.New("internal storage error")
)
```

### Error Wrapping

```go
import "fmt"

func (s *SQLiteStorage) GetTask(ctx context.Context, id string) (*Task, error) {
    // ...
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("%w: task %s", ErrNotFound, id)
    }
    return nil, fmt.Errorf("get task: %w", err)
}
```

---

## SQLite Implementation

### Directory Structure

```
internal/storage/sqlite/
├── sqlite.go          # Main implementation
├── tasks.go           # Task operations
├── categories.go      # Category operations
├── stats.go           # Statistics
├── migrations.go      # Schema migrations
└── queries.go         # SQL query constants
```

### Main Implementation

```go
// internal/storage/sqlite/sqlite.go
package sqlite

import (
    "database/sql"
    "github.com/yourusername/quest-todo/internal/storage"
    _ "modernc.org/sqlite"
)

type SQLiteStorage struct {
    db *sql.DB
}

func New(path string) (*SQLiteStorage, error) {
    // Open database
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, err
    }

    // Configure connection pool
    db.SetMaxOpenConns(1)  // SQLite limitation
    db.SetMaxIdleConns(1)

    s := &SQLiteStorage{db: db}

    // Run migrations
    if err := s.migrate(); err != nil {
        db.Close()
        return nil, err
    }

    return s, nil
}

func (s *SQLiteStorage) Close() error {
    return s.db.Close()
}

func (s *SQLiteStorage) Ping() error {
    return s.db.Ping()
}

func (s *SQLiteStorage) Backup(dest string) error {
    // Use SQLite backup API
    // Implementation details...
    return nil
}
```

### Task Operations

```go
// internal/storage/sqlite/tasks.go
package sqlite

import (
    "context"
    "database/sql"
    "github.com/yourusername/quest-todo/internal/storage"
)

func (s *SQLiteStorage) CreateTask(ctx context.Context, task *storage.Task) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Insert task
    _, err = tx.ExecContext(ctx, `
        INSERT INTO tasks (
            id, title, description, priority, deadline_type, deadline_date,
            category, status, notes, reward, order_index, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, task.ID, task.Title, task.Description, task.Priority,
       task.Deadline.Type, task.Deadline.Date, task.Category,
       task.Status, task.Notes, task.Reward, task.Order,
       task.CreatedAt, task.UpdatedAt)

    if err != nil {
        return err
    }

    // Insert objectives
    for _, obj := range task.Objectives {
        _, err = tx.ExecContext(ctx, `
            INSERT INTO objectives (id, task_id, text, completed, order_index, created_at)
            VALUES (?, ?, ?, ?, ?, ?)
        `, obj.ID, task.ID, obj.Text, obj.Completed, obj.Order, obj.CreatedAt)

        if err != nil {
            return err
        }
    }

    // Insert tags
    for _, tag := range task.Tags {
        _, err = tx.ExecContext(ctx, `
            INSERT INTO tags (task_id, tag) VALUES (?, ?)
        `, task.ID, tag)

        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func (s *SQLiteStorage) ListTasks(ctx context.Context, filter storage.TaskFilter) ([]*storage.Task, error) {
    // Build dynamic query
    query, args := s.buildTaskQuery(filter)

    rows, err := s.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []*storage.Task
    for rows.Next() {
        task := &storage.Task{}
        err := rows.Scan(
            &task.ID, &task.Title, &task.Description, &task.Priority,
            &task.Deadline.Type, &task.Deadline.Date,
            &task.Category, &task.Status, &task.Notes, &task.Reward,
            &task.Order, &task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
        )
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, task)
    }

    // Load objectives and tags for each task
    for _, task := range tasks {
        if err := s.loadTaskRelations(ctx, task); err != nil {
            return nil, err
        }
    }

    return tasks, nil
}

func (s *SQLiteStorage) buildTaskQuery(filter storage.TaskFilter) (string, []interface{}) {
    query := "SELECT * FROM tasks WHERE 1=1"
    args := []interface{}{}

    // Add filters
    if len(filter.Status) > 0 {
        query += " AND status IN (?" + strings.Repeat(",?", len(filter.Status)-1) + ")"
        for _, s := range filter.Status {
            args = append(args, s)
        }
    }

    // Add sorting
    if filter.SortBy != "" {
        order := "ASC"
        if filter.SortOrder == "desc" {
            order = "DESC"
        }
        query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
    }

    // Add pagination
    if filter.Limit > 0 {
        query += " LIMIT ? OFFSET ?"
        args = append(args, filter.Limit, filter.Offset)
    }

    return query, args
}
```

### Migrations

```go
// internal/storage/sqlite/migrations.go
package sqlite

func (s *SQLiteStorage) migrate() error {
    // Check current version
    version, err := s.getSchemaVersion()
    if err != nil {
        return err
    }

    // Apply migrations
    migrations := []migration{
        {version: 1, up: migrateV1},
        {version: 2, up: migrateV2},
    }

    for _, m := range migrations {
        if version < m.version {
            if err := m.up(s.db); err != nil {
                return err
            }
            if err := s.setSchemaVersion(m.version); err != nil {
                return err
            }
        }
    }

    return nil
}

func migrateV1(db *sql.DB) error {
    schema := `
    CREATE TABLE IF NOT EXISTS tasks (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT,
        priority INTEGER NOT NULL,
        deadline_type TEXT,
        deadline_date DATETIME,
        category TEXT,
        status TEXT NOT NULL,
        notes TEXT,
        reward INTEGER DEFAULT 0,
        order_index INTEGER DEFAULT 0,
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL,
        completed_at DATETIME
    );

    CREATE INDEX idx_tasks_status ON tasks(status);
    CREATE INDEX idx_tasks_priority ON tasks(priority DESC);
    CREATE INDEX idx_tasks_deadline ON tasks(deadline_date);

    -- Additional tables...
    `

    _, err := db.Exec(schema)
    return err
}
```

### Indexes

```sql
-- Performance indexes
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority DESC);
CREATE INDEX idx_tasks_deadline ON tasks(deadline_date);
CREATE INDEX idx_tasks_category ON tasks(category);
CREATE INDEX idx_tasks_created ON tasks(created_at DESC);
CREATE INDEX idx_tasks_order ON tasks(order_index);

-- Composite indexes for common queries
CREATE INDEX idx_tasks_status_priority ON tasks(status, priority DESC);
CREATE INDEX idx_tasks_status_deadline ON tasks(status, deadline_date);

-- Full-text search index
CREATE VIRTUAL TABLE tasks_fts USING fts5(
    id UNINDEXED,
    title,
    description,
    content=tasks
);
```

---

## JSON Implementation

### Directory Structure

```
internal/storage/json/
├── json.go           # Main implementation
├── tasks.go          # Task operations
├── categories.go     # Category operations
└── stats.go          # Statistics
```

### Main Implementation

```go
// internal/storage/json/json.go
package json

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"
    "github.com/yourusername/quest-todo/internal/storage"
)

type JSONStorage struct {
    dataPath string
    data     *Data
    mu       sync.RWMutex
}

type Data struct {
    Version    string                       `json:"version"`
    Tasks      map[string]*storage.Task     `json:"tasks"`
    Categories map[string]*storage.Category `json:"categories"`
    Settings   map[string]string            `json:"settings"`
}

func New(path string) (*JSONStorage, error) {
    s := &JSONStorage{
        dataPath: filepath.Join(path, "data.json"),
    }

    if err := s.load(); err != nil {
        // Initialize empty data
        s.data = &Data{
            Version:    "1.0",
            Tasks:      make(map[string]*storage.Task),
            Categories: make(map[string]*storage.Category),
            Settings:   make(map[string]string),
        }
        if err := s.save(); err != nil {
            return nil, err
        }
    }

    return s, nil
}

func (s *JSONStorage) load() error {
    data, err := os.ReadFile(s.dataPath)
    if err != nil {
        return err
    }

    return json.Unmarshal(data, &s.data)
}

func (s *JSONStorage) save() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    data, err := json.MarshalIndent(s.data, "", "  ")
    if err != nil {
        return err
    }

    // Atomic write: write to temp file then rename
    tempPath := s.dataPath + ".tmp"
    if err := os.WriteFile(tempPath, data, 0644); err != nil {
        return err
    }

    return os.Rename(tempPath, s.dataPath)
}

func (s *JSONStorage) Close() error {
    return s.save()
}

func (s *JSONStorage) Ping() error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return nil
}

func (s *JSONStorage) Backup(dest string) error {
    s.mu.RLock()
    defer s.mu.RUnlock()

    data, err := os.ReadFile(s.dataPath)
    if err != nil {
        return err
    }

    return os.WriteFile(dest, data, 0644)
}
```

### Task Operations

```go
// internal/storage/json/tasks.go
package json

import (
    "context"
    "github.com/yourusername/quest-todo/internal/storage"
    "sort"
)

func (s *JSONStorage) CreateTask(ctx context.Context, task *storage.Task) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, exists := s.data.Tasks[task.ID]; exists {
        return storage.ErrDuplicate
    }

    s.data.Tasks[task.ID] = task
    return s.save()
}

func (s *JSONStorage) GetTask(ctx context.Context, id string) (*storage.Task, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    task, ok := s.data.Tasks[id]
    if !ok {
        return nil, storage.ErrNotFound
    }

    // Return a copy to prevent external modification
    return copyTask(task), nil
}

func (s *JSONStorage) ListTasks(ctx context.Context, filter storage.TaskFilter) ([]*storage.Task, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    var tasks []*storage.Task

    // Filter tasks
    for _, task := range s.data.Tasks {
        if s.matchesFilter(task, filter) {
            tasks = append(tasks, copyTask(task))
        }
    }

    // Sort tasks
    s.sortTasks(tasks, filter.SortBy, filter.SortOrder)

    // Paginate
    tasks = s.paginateTasks(tasks, filter.Limit, filter.Offset)

    return tasks, nil
}

func (s *JSONStorage) matchesFilter(task *storage.Task, filter storage.TaskFilter) bool {
    // Check status
    if len(filter.Status) > 0 {
        match := false
        for _, status := range filter.Status {
            if task.Status == status {
                match = true
                break
            }
        }
        if !match {
            return false
        }
    }

    // Check priority
    if len(filter.Priority) > 0 {
        match := false
        for _, p := range filter.Priority {
            if task.Priority == p {
                match = true
                break
            }
        }
        if !match {
            return false
        }
    }

    // Additional filters...

    return true
}

func (s *JSONStorage) sortTasks(tasks []*storage.Task, sortBy, order string) {
    less := func(i, j int) bool {
        switch sortBy {
        case "priority":
            if order == "desc" {
                return tasks[i].Priority > tasks[j].Priority
            }
            return tasks[i].Priority < tasks[j].Priority
        case "created_at":
            if order == "desc" {
                return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
            }
            return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
        default:
            return i < j
        }
    }

    sort.Slice(tasks, less)
}
```

---

## Memory Implementation

For testing purposes only.

```go
// internal/storage/memory/memory.go
package memory

import (
    "context"
    "github.com/yourusername/quest-todo/internal/storage"
    "sync"
)

type MemoryStorage struct {
    tasks      map[string]*storage.Task
    categories map[string]*storage.Category
    mu         sync.RWMutex
}

func New() *MemoryStorage {
    return &MemoryStorage{
        tasks:      make(map[string]*storage.Task),
        categories: make(map[string]*storage.Category),
    }
}

// Implement all Storage interface methods...
```

---

## Factory Function

```go
// internal/storage/storage.go
package storage

type Config struct {
    Type string
    Path string
}

func New(cfg Config) (Storage, error) {
    switch cfg.Type {
    case "sqlite":
        return sqlite.New(cfg.Path)
    case "json":
        return json.New(cfg.Path)
    case "memory":
        return memory.New()
    default:
        return nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
    }
}
```

---

## Usage Examples

### In Service Layer

```go
type TaskService struct {
    store storage.Storage
}

func NewTaskService(store storage.Storage) *TaskService {
    return &TaskService{store: store}
}

func (s *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*storage.Task, error) {
    task := &storage.Task{
        ID:       uuid.New().String(),
        Title:    req.Title,
        Priority: req.Priority,
        // ...
    }

    if err := s.store.CreateTask(ctx, task); err != nil {
        return nil, err
    }

    return task, nil
}
```

### In Main

```go
func main() {
    cfg := loadConfig()

    store, err := storage.New(storage.Config{
        Type: cfg.Storage.Type,
        Path: cfg.Storage.Path,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    service := service.NewTaskService(store)
    // ...
}
```

---

## Testing

### Unit Tests with Memory Storage

```go
func TestTaskService_CreateTask(t *testing.T) {
    store := memory.New()
    service := NewTaskService(store)

    task, err := service.CreateTask(context.Background(), &CreateTaskRequest{
        Title:    "Test Task",
        Priority: 3,
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, task.ID)
}
```

### Integration Tests with SQLite

```go
func TestSQLiteStorage_CreateTask(t *testing.T) {
    // Use in-memory SQLite
    store, err := sqlite.New(":memory:")
    require.NoError(t, err)
    defer store.Close()

    task := &storage.Task{
        ID:       "test-1",
        Title:    "Test Task",
        Priority: 3,
    }

    err = store.CreateTask(context.Background(), task)
    assert.NoError(t, err)

    retrieved, err := store.GetTask(context.Background(), "test-1")
    assert.NoError(t, err)
    assert.Equal(t, task.Title, retrieved.Title)
}
```

---

## Performance Considerations

### SQLite
- **Indexes**: Critical for query performance
- **Transactions**: Use for multi-operation atomicity
- **Connection Pool**: Single connection (SQLite limitation)
- **WAL Mode**: Consider enabling for better concurrency

### JSON
- **File Size**: Keep under 10MB for good performance
- **Atomic Writes**: Use temp file + rename pattern
- **In-Memory**: Entire dataset loaded in memory
- **Locking**: Read-write mutex for concurrency

### Memory
- **Speed**: Fastest (no I/O)
- **Persistence**: None (testing only)
- **Limits**: RAM-bound

---

## Summary

The storage layer provides:

✅ **Clean abstraction** - Business logic independent of storage
✅ **Pluggable implementations** - Easy to swap SQLite/JSON/Memory
✅ **Performance** - SQLite with B-tree indexes
✅ **Testability** - Mock with memory storage
✅ **Maintainability** - Clear interface contracts
✅ **Extensibility** - Add new storage backends easily

This architecture allows starting simple (JSON) and scaling up (SQLite) without rewriting application code.
