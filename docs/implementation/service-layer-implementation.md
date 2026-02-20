# Service Layer Implementation

## Overview

The service layer contains all **business logic** for the Quest Todo application. It sits between the API layer (HTTP handlers) and the Storage layer (database), orchestrating operations and enforcing business rules.

**Purpose**: Separate business logic from infrastructure concerns

**Responsibilities**:
- Input validation and sanitization
- Business rule enforcement
- Computed field calculation (progress, stats)
- Transaction coordination
- Error handling and wrapping
- Authorization logic (future)

**Pattern**: Service interface with pluggable storage backend

## Architecture

### Layer Separation

```
┌─────────────────────────────────────────┐
│ API Layer (HTTP Handlers)               │
│ - Parse requests                        │
│ - Serialize responses                   │
│ - HTTP status codes                     │
└─────────────────┬───────────────────────┘
                  │
                  ↓ Calls service methods
┌─────────────────────────────────────────┐
│ Service Layer (Business Logic)          │
│ - Validate inputs                       │
│ - Enforce business rules                │
│ - Calculate computed fields             │
│ - Coordinate storage operations         │
└─────────────────┬───────────────────────┘
                  │
                  ↓ Calls storage methods
┌─────────────────────────────────────────┐
│ Storage Layer (Data Persistence)        │
│ - CRUD operations                       │
│ - Queries and filters                   │
│ - Transactions                          │
└─────────────────────────────────────────┘
```

### Why Separate Layers?

**Without Service Layer** (anti-pattern):
```go
// API handler doing business logic
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
    var req TaskCreateRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Validation in handler ❌
    if req.Priority < 1 || req.Priority > 5 {
        http.Error(w, "invalid priority", 400)
        return
    }

    // Business logic in handler ❌
    task := &Task{
        ID:       generateID(),
        Title:    req.Title,
        Progress: calculateProgress(req.Objectives),
    }

    storage.CreateTask(ctx, task)
    json.NewEncoder(w).Encode(task)
}
```

**Problems**:
- ❌ Business logic mixed with HTTP concerns
- ❌ Can't reuse logic (CLI, gRPC, WebSocket would duplicate)
- ❌ Hard to test (need HTTP mocks)
- ❌ Can't change API independently

**With Service Layer** (correct):
```go
// API handler - thin translation layer
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
    var req TaskCreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", 400)
        return
    }

    // Delegate to service
    task, err := taskService.CreateTask(r.Context(), req)
    if err != nil {
        handleServiceError(w, err)  // Map service errors to HTTP codes
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}

// Service - business logic
func (s *TaskService) CreateTask(ctx context.Context, req TaskCreateRequest) (*Task, error) {
    // Validation
    if err := s.validate.Struct(req); err != nil {
        return nil, ErrInvalidInput
    }

    // Business logic
    task := &Task{
        ID:       s.generateID(),
        Title:    req.Title,
        Priority: req.Priority,
        Status:   StatusActive,
        CreatedAt: time.Now(),
    }

    // Calculate computed fields
    task.Progress = s.calculateProgress(task.Objectives)

    // Persist
    if err := s.storage.CreateTask(ctx, task); err != nil {
        return nil, fmt.Errorf("create task: %w", err)
    }

    return task, nil
}
```

**Benefits**:
- ✅ Business logic reusable (REST, gRPC, CLI all use same service)
- ✅ Easy to test (mock storage, no HTTP)
- ✅ Clear separation of concerns
- ✅ Can change API independently

## Service Interface

### TaskService Interface

```go
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
```

### Service Dependencies

```go
type TaskServiceImpl struct {
    storage   storage.Storage
    validator *validator.Validate
    idGen     IDGenerator
    clock     Clock
}

// IDGenerator generates unique IDs
type IDGenerator interface {
    Generate() string
}

// Clock provides current time (for testing)
type Clock interface {
    Now() time.Time
}
```

**Why interfaces for dependencies?**
- `storage`: Can mock for unit tests
- `validator`: Can swap validation library
- `idGen`: Can mock for deterministic tests
- `clock`: Can freeze time in tests

## Business Logic Details

### 1. Task Creation

**Business Rules**:
- ✅ Title is required (1-200 characters)
- ✅ Priority must be 1-5
- ✅ Status defaults to "active"
- ✅ CreatedAt/UpdatedAt set to current time
- ✅ Order defaults to max+1
- ✅ ID auto-generated (UUID v4)
- ✅ Progress calculated from objectives
- ✅ Deadline validation (date must be future)

**Implementation**:
```go
func (s *TaskServiceImpl) CreateTask(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error) {
    // 1. Validate input
    if err := s.validator.Struct(req); err != nil {
        return nil, NewValidationError(err)
    }

    // 2. Additional business rules
    if req.Deadline != nil && req.Deadline.Date != nil {
        if req.Deadline.Date.Before(s.clock.Now()) {
            return nil, ErrDeadlineInPast
        }
    }

    // 3. Build task
    now := s.clock.Now()
    task := &models.Task{
        ID:          s.idGen.Generate(),
        Title:       strings.TrimSpace(req.Title),
        Description: strings.TrimSpace(req.Description),
        Priority:    req.Priority,
        Deadline:    req.Deadline,
        Category:    req.Category,
        Status:      models.StatusActive,
        Notes:       req.Notes,
        Reward:      req.Reward,
        Tags:        req.Tags,
        Order:       s.getNextOrder(ctx),
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    // 4. Create objectives
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

    // 5. Calculate progress
    task.Progress = s.calculateProgress(objectives)

    // 6. Persist to storage
    if err := s.storage.CreateTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to create task: %w", err)
    }

    return task, nil
}
```

### 2. Progress Calculation

**Business Rule**: Progress = (completed objectives / total objectives) × 100

```go
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
```

**Edge Cases**:
- No objectives: 0% progress
- All completed: 100% progress
- Partial: Proportional percentage

### 3. Task Completion

**Business Rules**:
- ✅ Can only complete "active" tasks
- ✅ Sets CompletedAt to current time
- ✅ Status changes to "completed"
- ✅ Cannot complete if required objectives incomplete (optional rule)

**Implementation**:
```go
func (s *TaskServiceImpl) CompleteTask(ctx context.Context, id string) (*models.Task, error) {
    // 1. Get current task
    task, err := s.storage.GetTask(ctx, id)
    if err != nil {
        return nil, err
    }

    // 2. Validate state transition
    if task.Status == models.StatusCompleted {
        return nil, ErrAlreadyCompleted
    }

    if task.Status == models.StatusFailed {
        return nil, ErrCannotCompleteFailedTask
    }

    // 3. Optional: Check objectives
    if s.config.RequireAllObjectives && task.Progress < 100 {
        return nil, ErrObjectivesIncomplete
    }

    // 4. Update task
    now := s.clock.Now()
    task.Status = models.StatusCompleted
    task.CompletedAt = &now
    task.UpdatedAt = now

    // 5. Persist
    if err := s.storage.UpdateTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to complete task: %w", err)
    }

    return task, nil
}
```

### 4. Task Update

**Business Rules**:
- ✅ Only update fields that are provided (partial update)
- ✅ Cannot change ID or CreatedAt
- ✅ UpdatedAt always set to current time
- ✅ Validate changed fields
- ✅ Recalculate progress if objectives changed

**Implementation**:
```go
func (s *TaskServiceImpl) UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error) {
    // 1. Get existing task
    task, err := s.storage.GetTask(ctx, id)
    if err != nil {
        return nil, err
    }

    // 2. Validate input
    if err := s.validator.Struct(req); err != nil {
        return nil, NewValidationError(err)
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
        return nil, fmt.Errorf("failed to update task: %w", err)
    }

    return task, nil
}
```

### 5. Objective Toggle

**Business Rules**:
- ✅ Toggle objective completion status
- ✅ Recalculate parent task progress
- ✅ Update parent task's UpdatedAt
- ✅ Auto-complete task if all objectives done (optional)

**Implementation**:
```go
func (s *ObjectiveServiceImpl) ToggleObjective(ctx context.Context, id string) (*models.Objective, error) {
    // 1. Get objective
    objective, err := s.storage.GetObjective(ctx, id)
    if err != nil {
        return nil, err
    }

    // 2. Toggle completion
    objective.Completed = !objective.Completed

    // 3. Update objective
    if err := s.storage.UpdateObjective(ctx, objective); err != nil {
        return nil, fmt.Errorf("failed to toggle objective: %w", err)
    }

    // 4. Get parent task
    task, err := s.storage.GetTask(ctx, objective.TaskID)
    if err != nil {
        return nil, err
    }

    // 5. Recalculate task progress
    task.Progress = s.calculateProgress(task.Objectives)
    task.UpdatedAt = s.clock.Now()

    // 6. Optional: Auto-complete task
    if s.config.AutoCompleteOnFullProgress && task.Progress == 100 {
        task.Status = models.StatusCompleted
        now := s.clock.Now()
        task.CompletedAt = &now
    }

    // 7. Update task
    if err := s.storage.UpdateTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to update task progress: %w", err)
    }

    return objective, nil
}
```

### 6. Bulk Operations

**Business Rules**:
- ✅ All-or-nothing (transaction semantics)
- ✅ Validate all items before creating any
- ✅ Return partial errors with details
- ✅ Limit batch size (e.g., max 50 items)

**Implementation**:
```go
func (s *TaskServiceImpl) CreateTasksBulk(ctx context.Context, reqs []models.TaskCreateRequest) ([]*models.Task, error) {
    // 1. Validate batch size
    if len(reqs) > s.config.MaxBulkSize {
        return nil, ErrBulkSizeTooLarge
    }

    // 2. Validate all requests first
    for i, req := range reqs {
        if err := s.validator.Struct(req); err != nil {
            return nil, fmt.Errorf("request %d invalid: %w", i, NewValidationError(err))
        }
    }

    // 3. Build all tasks
    tasks := make([]*models.Task, len(reqs))
    for i, req := range reqs {
        task, err := s.buildTaskFromRequest(req)
        if err != nil {
            return nil, fmt.Errorf("failed to build task %d: %w", i, err)
        }
        tasks[i] = task
    }

    // 4. Persist all (atomic if storage supports transactions)
    if err := s.storage.CreateTasksBulk(ctx, tasks); err != nil {
        return nil, fmt.Errorf("failed to create tasks: %w", err)
    }

    return tasks, nil
}
```

## Validation

### Input Validation

Use `go-playground/validator` for declarative validation:

```go
import "github.com/go-playground/validator/v10"

type TaskCreateRequest struct {
    Title       string   `json:"title" validate:"required,min=1,max=200"`
    Description string   `json:"description" validate:"max=2000"`
    Priority    int      `json:"priority" validate:"required,min=1,max=5"`
    Category    string   `json:"category,omitempty"`
    Tags        []string `json:"tags,omitempty" validate:"dive,min=1,max=50"`
}

// In service initialization
validate := validator.New()

// Custom validators
validate.RegisterValidation("future_date", func(fl validator.FieldLevel) bool {
    date, ok := fl.Field().Interface().(time.Time)
    if !ok {
        return false
    }
    return date.After(time.Now())
})
```

### Validation Error Handling

```go
type ValidationError struct {
    Field   string
    Message string
}

func NewValidationError(err error) error {
    validationErrs, ok := err.(validator.ValidationErrors)
    if !ok {
        return err
    }

    errors := make([]ValidationError, len(validationErrs))
    for i, fieldErr := range validationErrs {
        errors[i] = ValidationError{
            Field:   fieldErr.Field(),
            Message: getErrorMessage(fieldErr),
        }
    }

    return &MultiValidationError{Errors: errors}
}

func getErrorMessage(err validator.FieldError) string {
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
```

## Error Handling

### Service Errors

Define domain-specific errors:

```go
package service

import "errors"

var (
    // Input errors
    ErrInvalidInput         = errors.New("invalid input")
    ErrTaskNotFound         = errors.New("task not found")
    ErrObjectiveNotFound    = errors.New("objective not found")
    ErrCategoryNotFound     = errors.New("category not found")

    // Business rule errors
    ErrDeadlineInPast       = errors.New("deadline must be in the future")
    ErrAlreadyCompleted     = errors.New("task already completed")
    ErrCannotCompleteFailedTask = errors.New("cannot complete a failed task")
    ErrObjectivesIncomplete = errors.New("all objectives must be completed first")
    ErrBulkSizeTooLarge     = errors.New("bulk operation exceeds maximum size")
    ErrCannotDeleteCategory = errors.New("cannot delete category with active tasks")

    // Concurrency errors
    ErrVersionConflict      = errors.New("version conflict, task was modified")
)
```

### Error Wrapping

Wrap storage errors with context:

```go
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
```

**Why wrap?**
- Service errors are domain-specific
- Storage errors are implementation details
- API layer only needs to know about service errors

## Testing Strategy

### Unit Tests (With Mocked Storage)

Test business logic in isolation:

```go
type MockStorage struct {
    tasks map[string]*models.Task
}

func (m *MockStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
    task, ok := m.tasks[id]
    if !ok {
        return nil, storage.ErrNotFound
    }
    return task, nil
}

func TestTaskService_CompleteTask(t *testing.T) {
    // Arrange
    mockStorage := &MockStorage{
        tasks: map[string]*models.Task{
            "task-1": {
                ID:     "task-1",
                Title:  "Test",
                Status: models.StatusActive,
            },
        },
    }

    service := NewTaskService(mockStorage, fixedClock, uuidGen, validator)

    // Act
    task, err := service.CompleteTask(context.Background(), "task-1")

    // Assert
    require.NoError(t, err)
    assert.Equal(t, models.StatusCompleted, task.Status)
    assert.NotNil(t, task.CompletedAt)
}

func TestTaskService_CompleteTask_AlreadyCompleted(t *testing.T) {
    // Arrange
    mockStorage := &MockStorage{
        tasks: map[string]*models.Task{
            "task-1": {
                ID:     "task-1",
                Status: models.StatusCompleted,
            },
        },
    }

    service := NewTaskService(mockStorage, fixedClock, uuidGen, validator)

    // Act
    _, err := service.CompleteTask(context.Background(), "task-1")

    // Assert
    assert.ErrorIs(t, err, ErrAlreadyCompleted)
}
```

### Integration Tests (With Real Storage)

Test service + storage together:

```go
func TestTaskService_Integration(t *testing.T) {
    // Use real SQLite storage
    storage, _ := sqlite.New(":memory:")
    defer storage.Close()

    service := NewTaskService(storage, realClock, uuidGen, validator)

    // Create task
    task, err := service.CreateTask(ctx, models.TaskCreateRequest{
        Title:    "Integration Test",
        Priority: 3,
    })
    require.NoError(t, err)

    // Retrieve task
    retrieved, err := service.GetTask(ctx, task.ID)
    require.NoError(t, err)
    assert.Equal(t, task.Title, retrieved.Title)

    // Complete task
    completed, err := service.CompleteTask(ctx, task.ID)
    require.NoError(t, err)
    assert.Equal(t, models.StatusCompleted, completed.Status)
}
```

### Test Helpers

```go
// Fixed clock for deterministic tests
type FixedClock struct {
    time time.Time
}

func (c *FixedClock) Now() time.Time {
    return c.time
}

// Fixed ID generator
type FixedIDGen struct {
    ids []string
    idx int
}

func (g *FixedIDGen) Generate() string {
    id := g.ids[g.idx]
    g.idx++
    return id
}
```

## Configuration

### Service Config

```go
type Config struct {
    // Validation
    MaxTitleLength      int
    MaxDescriptionLength int
    MaxBulkSize         int

    // Business rules
    RequireAllObjectives      bool
    AutoCompleteOnFullProgress bool
    AllowPastDeadlines        bool

    // Features
    EnableCategoryRestrictions bool
    EnableRewardSystem        bool
}

func DefaultConfig() *Config {
    return &Config{
        MaxTitleLength:             200,
        MaxDescriptionLength:       2000,
        MaxBulkSize:                50,
        RequireAllObjectives:       false,
        AutoCompleteOnFullProgress: true,
        AllowPastDeadlines:         false,
        EnableCategoryRestrictions: true,
        EnableRewardSystem:         true,
    }
}
```

## Service Factory

```go
package service

import (
    "github.com/go-playground/validator/v10"
    "github.com/google/uuid"
)

// Services holds all service instances
type Services struct {
    Task       TaskService
    Objective  ObjectiveService
    Category   CategoryService
    Stats      StatsService
}

// NewServices creates all services with shared dependencies
func NewServices(storage storage.Storage, config *Config) *Services {
    // Shared dependencies
    validate := validator.New()
    idGen := &UUIDGenerator{}
    clock := &SystemClock{}

    // Create services
    return &Services{
        Task:      NewTaskService(storage, clock, idGen, validate, config),
        Objective: NewObjectiveService(storage, clock, idGen, validate, config),
        Category:  NewCategoryService(storage, validate, config),
        Stats:     NewStatsService(storage),
    }
}

// UUIDGenerator generates UUIDs
type UUIDGenerator struct{}

func (g *UUIDGenerator) Generate() string {
    return uuid.New().String()
}

// SystemClock returns current system time
type SystemClock struct{}

func (c *SystemClock) Now() time.Time {
    return time.Now()
}
```

## Implementation Checklist

### Phase 1: Core Services (Day 1-2)

- [ ] Define service interfaces
- [ ] Implement TaskService
  - [ ] CreateTask with validation
  - [ ] GetTask
  - [ ] UpdateTask (partial updates)
  - [ ] DeleteTask
  - [ ] ListTasks (delegate to storage)
- [ ] Implement progress calculation
- [ ] Add input validation
- [ ] Write unit tests with mocked storage

### Phase 2: Advanced Operations (Day 3)

- [ ] Implement status transitions
  - [ ] CompleteTask
  - [ ] FailTask
  - [ ] ReactivateTask
- [ ] Implement ObjectiveService
  - [ ] CreateObjective
  - [ ] UpdateObjective
  - [ ] ToggleObjective (with progress recalc)
  - [ ] DeleteObjective
- [ ] Add business rule enforcement
- [ ] Write integration tests

### Phase 3: Bulk Operations (Day 4)

- [ ] Implement bulk create
- [ ] Implement bulk update
- [ ] Implement bulk delete
- [ ] Add transaction handling
- [ ] Test edge cases (partial failures)

### Phase 4: Supporting Services (Day 5)

- [ ] Implement CategoryService
  - [ ] CRUD operations
  - [ ] Validate category references
  - [ ] Prevent deleting categories with tasks
- [ ] Implement StatsService
  - [ ] GetStats
  - [ ] GetCategoryStats
- [ ] Add computed field calculations

## Directory Structure

```
backend/internal/service/
├── service.go              # Service interfaces
├── errors.go               # Service errors
├── config.go               # Service configuration
├── factory.go              # Service factory
├── testing.go              # Test helpers
├── task_service.go         # TaskService implementation
├── task_service_test.go    # TaskService unit tests
├── objective_service.go    # ObjectiveService implementation
├── objective_service_test.go
├── category_service.go     # CategoryService implementation
├── category_service_test.go
├── stats_service.go        # StatsService implementation
├── stats_service_test.go
└── integration_test.go     # Integration tests
```

## References

- Storage Interface: `backend/internal/storage/storage.go`
- Data Models: `backend/internal/models/models.go`
- Request Models: `backend/internal/models/requests.go`
- Validator: https://github.com/go-playground/validator

## Related Documentation

- [Testing Philosophy](../testing/README.md)
- [Storage Layer Tests](../testing/storage-layer-tests.md)
- [Data Models Implementation](./data-models-implementation.md)
- [SQLite Implementation](./sqlite-implementation.md)
