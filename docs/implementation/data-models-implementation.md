# Data Models Implementation

Comprehensive documentation for the Go data models implementation, covering struct design, tags, types, and best practices.

## Table of Contents

1. [Overview](#overview)
2. [Struct Design Principles](#struct-design-principles)
3. [Struct Tags Explained](#struct-tags-explained)
4. [Core Models](#core-models)
5. [Request/Response Separation](#requestresponse-separation)
6. [Type Safety](#type-safety)
7. [Pointers vs Values](#pointers-vs-values)
8. [Best Practices](#best-practices)

---

## Overview

The `internal/models/` package contains all data structures used throughout the application. These models define how data is:
- Stored in the database
- Transmitted over HTTP (JSON)
- Validated before processing
- Presented to the user

**Package structure:**
```
internal/models/
├── models.go       # Core domain models
├── requests.go     # API request DTOs
└── responses.go    # API response wrappers
```

---

## Struct Design Principles

### 1. Separation of Concerns

**Three types of models:**

```go
// Domain model (what data is)
type Task struct {
	ID        string
	Title     string
	CreatedAt time.Time
	// ...
}

// Request model (what clients send)
type TaskCreateRequest struct {
	Title string  // Only fields client can set
	// No ID, CreatedAt (server-controlled)
}

// Response model (how data is returned)
type APIResponse struct {
	Success bool
	Data    interface{}
	Error   *APIError
}
```

**Why separate?**

**Problem: Using Task directly for create:**
```go
// Client sends:
{
  "id": "malicious-id",      // ❌ Client shouldn't control ID
  "title": "My Task",
  "createdAt": "1970-01-01", // ❌ Client shouldn't set timestamp
  "progress": 1.0            // ❌ Computed field
}

// Server must ignore some fields, error-prone
```

**Solution: Separate request type:**
```go
type TaskCreateRequest struct {
	Title    string  // ✅ Only what client provides
	Priority int
	// Server generates: ID, CreatedAt, Progress
}
```

### 2. Explicit Fields

**Avoid embedding for core models:**

```go
// ❌ Bad: Hidden fields
type Task struct {
	BaseModel  // What fields does this have?
	Title string
}

// ✅ Good: Explicit
type Task struct {
	ID        string     // Clear
	CreatedAt time.Time  // Clear
	UpdatedAt time.Time  // Clear
	Title     string
}
```

**Trade-off:**
- ❌ Embedding: Less repetition, but harder to understand
- ✅ Explicit: More typing, but crystal clear

**For Quest Todo:** Clarity > Brevity

### 3. Zero Values

**Design for zero values:**

```go
type Task struct {
	Status TaskStatus  // Zero value = "" (empty string)
	Reward int         // Zero value = 0
	Tags   []string    // Zero value = nil (empty slice)
}

// Create task without setting everything:
task := Task{
	Title: "My Task",
	// Status: "",  // Implicitly zero
	// Reward: 0,   // Implicitly zero
	// Tags: nil,   // Implicitly zero
}
```

**Why this matters:**

```go
// Must check zero values:
if task.Status == "" {
	task.Status = StatusActive  // Default
}

if task.Reward == 0 {
	task.Reward = 10  // Default
}

if task.Tags == nil {
	task.Tags = []string{}  // Empty slice, not nil
}
```

---

## Struct Tags Explained

### What Are Struct Tags?

**Metadata attached to struct fields:**

```go
type Task struct {
	Title string `json:"title" db:"title" validate:"required,max=200"`
	//             ──────────── ─────────── ─────────────────────────
	//              JSON tag     DB tag      Validation tag
}
```

**Accessed via reflection at runtime:**
```go
field := reflect.TypeOf(Task{}).Field(0)
jsonTag := field.Tag.Get("json")  // "title"
```

### JSON Tags

**Purpose:** Control JSON serialization/deserialization

#### Basic JSON Tags

```go
type Task struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Marshaling to JSON:
task := Task{ID: "123", Title: "My Task"}
json.Marshal(task)
// Result: {"id":"123","title":"My Task"}

// Without tags:
// Result: {"ID":"123","Title":"My Task"}  // Capitalized
```

#### `omitempty` Modifier

```go
type Task struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}
```

**Without `omitempty`:**
```json
{
  "title": "My Task",
  "description": "",      // Empty string included
  "completedAt": null     // Null included
}
```

**With `omitempty`:**
```json
{
  "title": "My Task"
  // description and completedAt omitted (not included at all)
}
```

**When to use:**
- Optional fields
- Pointer fields (nil = not set)
- Reduce JSON payload size

#### `-` (Ignore Field)

```go
type Task struct {
	Title    string  `json:"title"`
	internal string  `json:"-"`  // Never in JSON
	Progress float64 `json:"progress" db:"-"`  // In JSON, not DB
}
```

**Use cases:**
- Computed fields (Progress calculated, not stored)
- Sensitive data (passwords)
- Internal bookkeeping

### Database Tags

**Purpose:** Map struct fields to database columns

```go
type Task struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`  // Snake case in DB
	OrderIndex int     `db:"order_index"`
}
```

**Naming convention:**
- Go: CamelCase (`CreatedAt`)
- Database: snake_case (`created_at`)

**Why different?**
- Go convention: Exported fields start with capital
- SQL convention: lowercase with underscores

**Tag maps between them:**
```go
// Go code:
task.CreatedAt = time.Now()

// SQL:
INSERT INTO tasks (created_at) VALUES (?)
```

#### `-` in Database Tags

```go
type Task struct {
	Progress  float64 `json:"progress" db:"-"`
	IsOverdue bool    `json:"isOverdue" db:"-"`
}
```

**Meaning:** Don't persist to database (computed on-the-fly)

**Why computed fields?**
```go
// Progress = completed objectives / total objectives
func (t *Task) calculateProgress() float64 {
	if len(t.Objectives) == 0 {
		return 0
	}

	completed := 0
	for _, obj := range t.Objectives {
		if obj.Completed {
			completed++
		}
	}

	return float64(completed) / float64(len(t.Objectives))
}

// Always accurate, no stale data
```

### Validation Tags

**Purpose:** Define validation rules (used by validation libraries)

```go
type TaskCreateRequest struct {
	Title    string `validate:"required,min=1,max=200"`
	Priority int    `validate:"required,min=1,max=5"`
	Email    string `validate:"email"`
}
```

**Common rules:**
- `required` - Cannot be empty
- `min=1` - Minimum value/length
- `max=200` - Maximum value/length
- `email` - Must be valid email format
- `url` - Must be valid URL
- `oneof=main side` - Must be one of these values

**Usage with validator library:**
```go
import "github.com/go-playground/validator/v10"

validate := validator.New()

req := TaskCreateRequest{Title: "", Priority: 10}
err := validate.Struct(req)
// Error: Title is required, Priority exceeds max
```

---

## Core Models

### Task Model

```go
type Task struct {
	// Identity
	ID string `json:"id" db:"id"`

	// Core fields
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Priority    int        `json:"priority" db:"priority"`

	// Deadline (embedded struct)
	Deadline    Deadline   `json:"deadline"`

	// References
	Category    string     `json:"category" db:"category"`
	Status      TaskStatus `json:"status" db:"status"`

	// Relations (not in DB, loaded separately)
	Objectives  []Objective `json:"objectives"`
	Tags        []string    `json:"tags"`

	// Metadata
	Notes       string     `json:"notes" db:"notes"`
	Reward      int        `json:"reward" db:"reward"`
	Order       int        `json:"order" db:"order_index"`

	// Timestamps
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	CompletedAt *time.Time `json:"completedAt,omitempty" db:"completed_at"`

	// Computed fields
	Progress  float64 `json:"progress" db:"-"`
	IsOverdue bool    `json:"isOverdue" db:"-"`
	DaysLeft  *int    `json:"daysLeft,omitempty" db:"-"`
}
```

#### Design Decisions

**1. ID as string (UUID)**
```go
ID string `json:"id"`
// "550e8400-e29b-41d4-a009-426614174000"
```

**Why string instead of int?**
- ✅ Globally unique (can merge databases)
- ✅ Client-side generation (don't need DB)
- ✅ Non-sequential (security: can't guess IDs)
- ✅ Distributed systems friendly
- ❌ Larger (36 bytes vs 8 bytes)

**For our scale (1000s of tasks): String UUID is better**

**2. Priority as int (1-5)**
```go
Priority int `json:"priority" db:"priority"`
```

**Why int instead of enum?**
- ✅ Simple to store and compare
- ✅ Easy arithmetic: `priority >= 4` (high priority)
- ✅ Can sort: `ORDER BY priority DESC`
- ✅ Validation in CHECK constraint: `CHECK(priority BETWEEN 1 AND 5)`

**Alternative considered:**
```go
type Priority string
const (
	PriorityLowest  Priority = "lowest"
	PriorityLow     Priority = "low"
	PriorityMedium  Priority = "medium"
	PriorityHigh    Priority = "high"
	PriorityHighest Priority = "highest"
)
```

**Rejected because:**
- ❌ Can't sort easily: `"medium" > "high"` is alphabetical, not priority order
- ❌ Requires mapping: "high" = 4 stars
- ❌ More complex

**3. Deadline as embedded struct**
```go
Deadline Deadline `json:"deadline"`

type Deadline struct {
	Type string     `json:"type" db:"deadline_type"`
	Date *time.Time `json:"date,omitempty" db:"deadline_date"`
}
```

**Why separate type and date?**
```json
// Can have type without date:
{"deadline": {"type": "short", "date": null}}

// Can query by type:
SELECT * FROM tasks WHERE deadline_type = 'short'

// Can calculate date later:
if deadline.Type == "short" && deadline.Date == nil {
	deadline.Date = time.Now().Add(3 * 24 * time.Hour)
}
```

**4. Objectives as slice (not in DB column)**
```go
Objectives []Objective `json:"objectives"`
```

**Stored in separate table:**
```sql
CREATE TABLE objectives (
	id TEXT PRIMARY KEY,
	task_id TEXT,  -- Foreign key to tasks
	text TEXT,
	completed BOOLEAN
)
```

**Loaded via JOIN:**
```go
// Load task
task := loadTaskFromDB(id)

// Load objectives separately
task.Objectives = loadObjectivesForTask(task.ID)
```

**5. CompletedAt as pointer**
```go
CompletedAt *time.Time `json:"completedAt,omitempty"`
```

**Why pointer?**

**Non-pointer problem:**
```go
// Zero value is Jan 1, 0001
CompletedAt time.Time  // Default: 0001-01-01 00:00:00

// Must check if set:
if task.CompletedAt.IsZero() {
	// Not completed yet
}

// JSON output:
{"completedAt": "0001-01-01T00:00:00Z"}  // Confusing!
```

**Pointer solution:**
```go
CompletedAt *time.Time  // Default: nil

// Clear check:
if task.CompletedAt == nil {
	// Not completed yet
}

// JSON output:
{}  // Field omitted with omitempty
// or
{"completedAt": null}  // Clearly null
```

**6. Computed fields with `db:"-"`**
```go
Progress  float64 `json:"progress" db:"-"`
IsOverdue bool    `json:"isOverdue" db:"-"`
```

**Calculated in code:**
```go
func enrichTask(task *Task) {
	// Calculate progress
	if len(task.Objectives) > 0 {
		completed := 0
		for _, obj := range task.Objectives {
			if obj.Completed {
				completed++
			}
		}
		task.Progress = float64(completed) / float64(len(task.Objectives))
	}

	// Calculate overdue
	if task.Deadline.Date != nil {
		task.IsOverdue = time.Now().After(*task.Deadline.Date)
	}

	// Calculate days left
	if task.Deadline.Date != nil {
		days := int(time.Until(*task.Deadline.Date).Hours() / 24)
		task.DaysLeft = &days
	}
}
```

**Why not store?**
- ✅ Always accurate (no stale data)
- ✅ Changes automatically when objectives change
- ✅ No need to update when time passes
- ✅ Saves database space

### Custom Types for Type Safety

#### TaskStatus

```go
type TaskStatus string

const (
	StatusActive     TaskStatus = "active"
	StatusInProgress TaskStatus = "in_progress"
	StatusComplete   TaskStatus = "complete"
	StatusFailed     TaskStatus = "failed"
	StatusArchived   TaskStatus = "archived"
)
```

**Why custom type instead of plain string?**

**Without custom type:**
```go
// Any string accepted
task.Status = "active"    // ✅ OK
task.Status = "comlpete" // ❌ Typo, but compiles!
task.Status = "done"      // ❌ Invalid, but compiles!
```

**With custom type:**
```go
// Only defined constants accepted
task.Status = StatusActive    // ✅ OK
task.Status = "comlpete"      // ❌ Compile error: cannot use string as TaskStatus
task.Status = StatusComplete  // ✅ OK (autocomplete helps)
```

**Benefits:**
1. **Compile-time safety** - Invalid values caught before running
2. **IDE autocomplete** - Type `Status` and see all options
3. **Self-documenting** - `const` block shows all valid values
4. **Refactoring** - Change value in one place

**Can add methods:**
```go
func (s TaskStatus) IsTerminal() bool {
	return s == StatusComplete || s == StatusFailed || s == StatusArchived
}

func (s TaskStatus) CanTransitionTo(next TaskStatus) bool {
	// Define valid state transitions
	switch s {
	case StatusActive:
		return next == StatusInProgress || next == StatusComplete || next == StatusFailed
	case StatusInProgress:
		return next == StatusComplete || next == StatusFailed
	case StatusComplete, StatusFailed:
		return next == StatusArchived
	}
	return false
}
```

---

## Request/Response Separation

### Why Separate Request Models?

**Problem: Using Task directly:**
```go
func CreateTask(task Task) error {
	// Client sent:
	// {"id": "123", "createdAt": "1970-01-01", "progress": 1.0}

	// Must ignore/override client-provided values:
	task.ID = generateID()           // Ignore client ID
	task.CreatedAt = time.Now()      // Ignore client timestamp
	task.Progress = 0                // Ignore client progress

	// Error-prone and confusing
}
```

**Solution: Separate request type:**
```go
type TaskCreateRequest struct {
	Title       string   `json:"title" validate:"required,max=200"`
	Priority    int      `json:"priority" validate:"required,min=1,max=5"`
	// Only fields client is allowed to set
}

func CreateTask(req TaskCreateRequest) (*Task, error) {
	// Server generates protected fields
	task := &Task{
		ID:        generateID(),      // Server controls
		Title:     req.Title,         // From client
		Priority:  req.Priority,      // From client
		CreatedAt: time.Now(),        // Server controls
		Progress:  0,                 // Computed
	}

	return task, nil
}
```

### Pointers in Update Requests

**Problem: Distinguishing "not provided" from "zero value"**

```go
// Without pointers:
type TaskUpdateRequest struct {
	Priority int  // 0 is zero value
}

// Client wants to update title only:
{"title": "New Title"}

// Go unmarshals to:
{Title: "New Title", Priority: 0}

// Problem: Did client want Priority=0, or did they not provide it?
```

**Solution: Pointers**

```go
type TaskUpdateRequest struct {
	Title    *string `json:"title,omitempty"`
	Priority *int    `json:"priority,omitempty"`
}

// Client sends: {"title": "New Title"}
// Go unmarshals to: {Title: &"New Title", Priority: nil}
// nil = not provided, don't update
```

**Usage:**
```go
func UpdateTask(id string, req TaskUpdateRequest) error {
	task := getTask(id)

	// Only update provided fields
	if req.Title != nil {
		task.Title = *req.Title  // Dereference pointer
	}

	if req.Priority != nil {
		task.Priority = *req.Priority
	}

	// Fields with nil pointers remain unchanged

	return saveTask(task)
}
```

**Setting to zero value explicitly:**
```go
// Client wants Priority = 0:
zero := 0
req := TaskUpdateRequest{
	Priority: &zero,  // Explicitly set to 0
}

// Server receives: Priority = &0 (not nil)
// Updates: task.Priority = 0
```

### Response Wrapper

```go
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *MetaData   `json:"meta,omitempty"`
}
```

**Consistent format for all endpoints:**

**Success:**
```json
{
  "success": true,
  "data": {"id": "123", "title": "My Task"}
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Task not found"
  }
}
```

**With pagination:**
```json
{
  "success": true,
  "data": [...],
  "meta": {
    "total": 100,
    "limit": 20,
    "offset": 0,
    "hasMore": true
  }
}
```

**Benefits:**
- ✅ Client always checks `success` field
- ✅ Error handling is consistent
- ✅ Easy to add metadata (pagination, etc.)
- ✅ Self-documenting (success/error is explicit)

**Helper functions:**
```go
// In handler:
return Success(task)
// Instead of manually: APIResponse{Success: true, Data: task}

return Error("NOT_FOUND", "Task not found")
// Instead of manually creating APIError

return SuccessWithMeta(tasks, &MetaData{Total: 100, Limit: 20})
```

---

## Pointers vs Values

### When to Use Pointers

**General rule:** Use pointer when:
1. **Nil has meaning** (optional, not set)
2. **Large struct** (avoid copying)
3. **Need to modify** (pass by reference)

#### 1. Nil Has Meaning

```go
// Pointer: nil = not completed
CompletedAt *time.Time

if task.CompletedAt == nil {
	// Task not completed
} else {
	// Task completed at: *task.CompletedAt
}
```

#### 2. Large Struct (Avoid Copying)

```go
// Task struct is ~200 bytes
// Passing by value copies all 200 bytes
func UpdateTask(task Task) {  // Copies 200 bytes
	task.Title = "New"
	// Modifies copy, original unchanged
}

// Passing by pointer copies 8 bytes (memory address)
func UpdateTask(task *Task) {  // Copies 8 bytes
	task.Title = "New"
	// Modifies original through pointer
}
```

**Rule of thumb:**
- Struct < 64 bytes: Value is fine
- Struct > 64 bytes: Pointer is better
- Task struct (~200 bytes): Use pointer

#### 3. Need to Modify

```go
// Can't modify through value
func SetTitle(task Task) {
	task.Title = "New"  // Modifies copy
}

task := Task{Title: "Old"}
SetTitle(task)
fmt.Println(task.Title)  // Still "Old"

// Can modify through pointer
func SetTitle(task *Task) {
	task.Title = "New"  // Modifies original
}

task := &Task{Title: "Old"}
SetTitle(task)
fmt.Println(task.Title)  // Now "New"
```

### When to Use Values

**Use value when:**
1. **Small struct** (<64 bytes)
2. **Immutable** (don't need to modify)
3. **No nil semantics** needed

```go
// Small struct, value is fine
type Deadline struct {
	Type string     // 16 bytes
	Date *time.Time // 8 bytes
}
// Total: 24 bytes, OK as value

// Embedded in Task
type Task struct {
	Deadline Deadline  // Value (not pointer)
	// Always has a deadline (even if type="none")
}
```

---

## Best Practices

### 1. Explicit Over Implicit

```go
// ❌ Bad: Magic behavior
type Task struct {
	Status string  // What are valid values?
}

// ✅ Good: Explicit constants
type TaskStatus string
const (
	StatusActive TaskStatus = "active"
	// ...
)
```

### 2. Zero Values

```go
// ❌ Bad: Must initialize
type Task struct {
	Tags []string  // nil
}
task := Task{}
task.Tags = []string{}  // Must remember to initialize

// ✅ Good: Zero value is usable
type Task struct {
	Tags []string  // nil is OK (len(nil) = 0)
}
task := Task{}
// Can use immediately: append(task.Tags, "work")
```

### 3. Computed Fields

```go
// ❌ Bad: Store computed field
type Task struct {
	Objectives []Objective
	Progress   float64  `db:"progress"`  // Stored in DB
}
// Problem: Can become stale when objectives change

// ✅ Good: Calculate on demand
type Task struct {
	Objectives []Objective
	Progress   float64  `json:"progress" db:"-"`  // Computed
}
func (t *Task) CalculateProgress() {
	// Always accurate
}
```

### 4. Validation

```go
// ❌ Bad: Validation scattered
func CreateTask(task Task) error {
	if task.Title == "" {
		return errors.New("title required")
	}
	if task.Priority < 1 || task.Priority > 5 {
		return errors.New("invalid priority")
	}
	// ...
}

// ✅ Good: Centralized validation
type TaskCreateRequest struct {
	Title    string `validate:"required,max=200"`
	Priority int    `validate:"required,min=1,max=5"`
}

validate := validator.New()
if err := validate.Struct(req); err != nil {
	return err
}
```

### 5. Immutable IDs

```go
// ❌ Bad: ID in update request
type TaskUpdateRequest struct {
	ID    string  // Can client change ID? Dangerous!
	Title string
}

// ✅ Good: ID only in URL/path
func UpdateTask(id string, req TaskUpdateRequest) {
	// ID from URL: PUT /tasks/{id}
	// Request body: {"title": "New Title"}
}
```

---

## Summary

### Key Principles

| Principle | Implementation |
|-----------|----------------|
| Separation | Domain models, Request DTOs, Response wrappers |
| Type safety | Custom types (TaskStatus), validation tags |
| Nil semantics | Pointers for optional fields |
| Efficiency | Pointers for large structs |
| Clarity | Explicit fields, no magic |
| Computed fields | Calculate on-the-fly, don't store |

### Struct Tags Usage

| Tag | Purpose | Example |
|-----|---------|---------|
| `json` | JSON serialization | `json:"title"` |
| `db` | Database mapping | `db:"created_at"` |
| `validate` | Input validation | `validate:"required,max=200"` |
| `omitempty` | Omit empty values | `json:"description,omitempty"` |
| `-` | Ignore field | `db:"-"` (computed field) |

### Pointers vs Values

| Use Case | Choice | Rationale |
|----------|--------|-----------|
| Optional field | Pointer | nil = not set |
| Large struct (>64 bytes) | Pointer | Avoid copying |
| Need to modify | Pointer | Modify original |
| Small struct (<64 bytes) | Value | Simple, efficient |
| Always present | Value | No nil confusion |

---

This design creates models that are:
- ✅ Type-safe
- ✅ Self-documenting
- ✅ Easy to validate
- ✅ Efficient
- ✅ Maintainable
