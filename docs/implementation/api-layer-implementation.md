# REST API Layer Implementation

## Overview

The REST API layer is the HTTP interface for the Quest Todo application. It handles:
- HTTP request parsing and validation
- Request routing to service layer operations
- Response formatting and serialization
- Error handling and HTTP status code mapping
- Query parameter parsing for filtering and pagination

This layer sits between HTTP clients (frontend, mobile apps, external APIs) and the service layer, translating HTTP semantics into business operations.

## Architecture

### Layer Responsibilities

```
┌─────────────────────────────────────────────────┐
│              HTTP Clients                       │
│   (Web Frontend, Mobile Apps, CLI Tools)        │
└─────────────────┬───────────────────────────────┘
                  │ HTTP Requests
                  ▼
┌─────────────────────────────────────────────────┐
│            API Layer (Handlers)                 │
│  - Parse HTTP requests                          │
│  - Validate JSON bodies                         │
│  - Extract URL parameters                       │
│  - Parse query strings                          │
│  - Format responses                             │
│  - Map errors to HTTP status codes              │
└─────────────────┬───────────────────────────────┘
                  │ Service method calls
                  ▼
┌─────────────────────────────────────────────────┐
│            Service Layer                        │
│  - Business logic                               │
│  - Validation                                   │
│  - Orchestration                                │
└─────────────────────────────────────────────────┘
```

### Design Principles

1. **Thin Controllers**: Handlers are thin - they parse requests, call services, format responses
2. **No Business Logic**: All business logic stays in the service layer
3. **Consistent API Format**: All responses use a standard envelope format
4. **HTTP Semantics**: Proper use of HTTP methods, status codes, and headers
5. **Validation at Boundary**: Input validation happens at the API boundary
6. **Error Mapping**: Service errors map to appropriate HTTP status codes

## File Structure

```
internal/api/
├── handlers.go              # Core API structure and utilities
├── errors.go                # Error handling and response utilities
├── task_handlers.go         # Task-related HTTP handlers
├── objective_handlers.go    # Objective-related HTTP handlers
├── category_handlers.go     # Category-related HTTP handlers
├── stats_handlers.go        # Statistics HTTP handlers
└── health_handlers.go       # Health check and version endpoints
```

## Core Components

### 1. API Structure (`handlers.go`)

The `API` struct holds all dependencies needed by handlers:

```go
type API struct {
    TaskService      TaskServiceInterface
    ObjectiveService ObjectiveServiceInterface
    CategoryService  CategoryServiceInterface
    StatsService     StatsServiceInterface
    Validator        *validator.Validate
    StartTime        time.Time
    AppVersion       string
}
```

**Design Decision**: We use interface types rather than concrete service implementations. This provides:
- Testability: Easy to mock services in handler tests
- Flexibility: Can swap implementations without changing handlers
- Decoupling: API layer doesn't depend on service implementation details

**Constructor**:
```go
func NewAPI(
    taskService TaskServiceInterface,
    objectiveService ObjectiveServiceInterface,
    categoryService CategoryServiceInterface,
    statsService StatsServiceInterface,
    validator *validator.Validate,
    version string,
) *API
```

The constructor captures the start time for uptime reporting and stores all service dependencies.

### 2. Request Parsing Utilities

#### `DecodeJSONBody()`

Decodes JSON request bodies with strict validation:

```go
func DecodeJSONBody(r *http.Request, dst interface{}) error {
    if r.Body == nil {
        return ErrEmptyBody
    }
    defer r.Body.Close()

    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()  // Reject unknown fields

    if err := decoder.Decode(dst); err != nil {
        return err
    }

    return nil
}
```

**Key Feature**: `DisallowUnknownFields()` prevents clients from sending unexpected fields. This:
- Catches typos in field names
- Prevents future compatibility issues
- Makes API contract explicit

#### `ExtractID()`

Extracts resource IDs from URL paths:

```go
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
```

**Usage Examples**:
- `/api/tasks/task-123` → `ExtractID(r, "/api/tasks/")` → `"task-123"`
- `/api/tasks/task-123/complete` → `ExtractID(r, "/api/tasks/")` → `"task-123"`

**Design Decision**: We use simple string parsing instead of a routing library for ID extraction. This keeps the code simple and doesn't lock us into a specific router. The actual routing will be handled by middleware/router layer.

#### `ParseTaskFilter()`

Parses query parameters into a structured filter:

```go
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

    // Pagination
    if limit := q.Get("limit"); limit != "" {
        if l, err := strconv.Atoi(limit); err == nil && l > 0 {
            filter.Limit = l
        }
    }

    // ... more filters
    return filter
}
```

**Supported Query Parameters**:
- `status=active,in_progress` - Multiple statuses (comma-separated)
- `priority=4,5` - Multiple priorities
- `category=cat-1,cat-2` - Multiple categories
- `tags=urgent,bug` - Multiple tags
- `deadlineType=short` - Deadline type filter
- `dateFrom=2024-01-01T00:00:00Z` - RFC3339 date
- `dateTo=2024-12-31T23:59:59Z` - RFC3339 date
- `includeCompleted=true` - Include completed tasks
- `sortBy=priority` - Sort field
- `sortOrder=desc` - Sort direction
- `limit=20` - Pagination limit
- `offset=0` - Pagination offset

**Example Request**:
```
GET /api/tasks?status=active&priority=4,5&sortBy=deadline&limit=20&offset=0
```

**Design Decision**: We use a single filter struct instead of individual parameters. This:
- Makes function signatures cleaner
- Allows easy filter composition
- Provides consistent API across all list endpoints

## Error Handling (`errors.go`)

### Error Response Strategy

The API layer translates internal errors into HTTP-appropriate responses with consistent format.

#### Error Response Functions

**`ErrorResponse()`** - Generic error response:
```go
func ErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(models.Error(code, message))
}
```

**Response Format**:
```json
{
  "success": false,
  "error": {
    "code": "TASK_NOT_FOUND",
    "message": "Task not found"
  }
}
```

**`ValidationErrorResponse()`** - Validation error with field details:
```go
func ValidationErrorResponse(w http.ResponseWriter, err error) {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        fields := make(map[string]string)
        for _, fe := range ve {
            fields[fe.Field()] = getValidationMessage(fe)
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.ValidationError(fields))
        return
    }
    // Fallback
    ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
}
```

**Response Format**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "title": "This field is required",
      "priority": "Value is too small or too large"
    }
  }
}
```

### Service Error Mapping

`HandleServiceError()` maps service layer errors to HTTP responses:

```go
func HandleServiceError(w http.ResponseWriter, err error) {
    if err == nil {
        return
    }

    switch {
    case errors.Is(err, service.ErrTaskNotFound):
        ErrorResponse(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
    case errors.Is(err, service.ErrObjectiveNotFound):
        ErrorResponse(w, http.StatusNotFound, "OBJECTIVE_NOT_FOUND", "Objective not found")
    case errors.Is(err, service.ErrCategoryNotFound):
        ErrorResponse(w, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found")
    case errors.Is(err, service.ErrDeadlineInPast):
        ErrorResponse(w, http.StatusBadRequest, "INVALID_DEADLINE", "Deadline must be in the future")
    case errors.Is(err, service.ErrObjectivesIncomplete):
        ErrorResponse(w, http.StatusBadRequest, "OBJECTIVES_INCOMPLETE", "Cannot complete task with incomplete objectives")
    case errors.Is(err, service.ErrCannotDeleteCategory):
        ErrorResponse(w, http.StatusConflict, "CATEGORY_HAS_TASKS", "Cannot delete category with active tasks")
    // ... more mappings
    default:
        ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
    }
}
```

**Error Mapping Table**:

| Service Error | HTTP Status | Error Code | Use Case |
|---------------|-------------|------------|----------|
| `ErrTaskNotFound` | 404 | `TASK_NOT_FOUND` | Task doesn't exist |
| `ErrObjectiveNotFound` | 404 | `OBJECTIVE_NOT_FOUND` | Objective doesn't exist |
| `ErrCategoryNotFound` | 404 | `CATEGORY_NOT_FOUND` | Category doesn't exist |
| `ErrDeadlineInPast` | 400 | `INVALID_DEADLINE` | Invalid deadline date |
| `ErrAlreadyCompleted` | 400 | `ALREADY_COMPLETED` | Task is already complete |
| `ErrCannotCompleteFailedTask` | 400 | `CANNOT_COMPLETE_FAILED` | Cannot complete failed task |
| `ErrObjectivesIncomplete` | 400 | `OBJECTIVES_INCOMPLETE` | Objectives not done |
| `ErrCannotDeleteCategory` | 409 | `CATEGORY_HAS_TASKS` | Category has active tasks |
| `ErrBulkSizeTooLarge` | 400 | `BULK_SIZE_TOO_LARGE` | Too many bulk items |
| `ErrVersionConflict` | 409 | `VERSION_CONFLICT` | Concurrent modification |

**Design Decision**: We use `errors.Is()` for error checking instead of type assertions. This supports error wrapping and makes error handling more robust.

### Success Response Functions

**`SuccessResponse()`** - Simple success response:
```go
func SuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(models.Success(data))
}
```

**`SuccessResponseWithMeta()`** - Success with pagination metadata:
```go
func SuccessResponseWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta *models.MetaData) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(models.SuccessWithMeta(data, meta))
}
```

**Response Format**:
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

**`NoContentResponse()`** - 204 No Content for DELETE operations:
```go
func NoContentResponse(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNoContent)
}
```

## Task Handlers (`task_handlers.go`)

### Handler Pattern

All handlers follow a consistent pattern:

1. **Extract Parameters**: Get ID from URL, parse query params
2. **Parse Request Body**: Decode JSON if present
3. **Validate Input**: Run validation on request struct
4. **Call Service**: Delegate to service layer
5. **Handle Errors**: Map service errors to HTTP responses
6. **Format Response**: Send success response with appropriate status code

### Example: CreateTask Handler

```go
func (api *API) CreateTask(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    var req models.TaskCreateRequest
    if err := DecodeJSONBody(r, &req); err != nil {
        ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
        return
    }

    // 2. Validate input
    if err := api.Validator.Struct(req); err != nil {
        ValidationErrorResponse(w, err)
        return
    }

    // 3. Call service
    task, err := api.TaskService.CreateTask(r.Context(), req)
    if err != nil {
        HandleServiceError(w, err)
        return
    }

    // 4. Format response
    SuccessResponse(w, http.StatusCreated, task)
}
```

**Request**:
```http
POST /api/tasks
Content-Type: application/json

{
  "title": "Implement API handlers",
  "description": "Create REST API handlers for all resources",
  "priority": 4,
  "deadline": {
    "type": "short",
    "date": "2024-01-15T23:59:59Z"
  },
  "category": "cat-work",
  "objectives": [
    {"text": "Create task handlers", "order": 1},
    {"text": "Create objective handlers", "order": 2}
  ],
  "tags": ["backend", "api"],
  "reward": 100
}
```

**Success Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "id": "task-123",
    "title": "Implement API handlers",
    "description": "Create REST API handlers for all resources",
    "priority": 4,
    "deadline": {
      "type": "short",
      "date": "2024-01-15T23:59:59Z"
    },
    "category": "cat-work",
    "status": "active",
    "objectives": [
      {
        "id": "obj-1",
        "taskId": "task-123",
        "text": "Create task handlers",
        "completed": false,
        "order": 1,
        "createdAt": "2024-01-10T10:00:00Z"
      },
      {
        "id": "obj-2",
        "taskId": "task-123",
        "text": "Create objective handlers",
        "completed": false,
        "order": 2,
        "createdAt": "2024-01-10T10:00:00Z"
      }
    ],
    "notes": "",
    "reward": 100,
    "tags": ["backend", "api"],
    "order": 0,
    "createdAt": "2024-01-10T10:00:00Z",
    "updatedAt": "2024-01-10T10:00:00Z",
    "progress": 0.0,
    "isOverdue": false
  }
}
```

**Validation Error Response** (400 Bad Request):
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "title": "This field is required",
      "priority": "Value is too small or too large"
    }
  }
}
```

### CRUD Operations

#### Create Task
- **Endpoint**: `POST /api/tasks`
- **Request Body**: `TaskCreateRequest`
- **Success**: 201 Created with task data
- **Errors**: 400 (validation), 500 (internal)

#### Get Task
- **Endpoint**: `GET /api/tasks/{id}`
- **Success**: 200 OK with task data
- **Errors**: 400 (invalid ID), 404 (not found), 500 (internal)

#### Update Task
- **Endpoint**: `PUT /api/tasks/{id}`
- **Request Body**: `TaskUpdateRequest` (partial updates supported)
- **Success**: 200 OK with updated task
- **Errors**: 400 (validation), 404 (not found), 500 (internal)

**Partial Update Example**:
```http
PUT /api/tasks/task-123
Content-Type: application/json

{
  "priority": 5,
  "notes": "Updated priority to urgent"
}
```

Only the specified fields are updated; all other fields remain unchanged.

#### Delete Task
- **Endpoint**: `DELETE /api/tasks/{id}`
- **Success**: 204 No Content
- **Errors**: 404 (not found), 500 (internal)

### List and Search Operations

#### List Tasks with Filtering
- **Endpoint**: `GET /api/tasks?status=active&priority=4,5&limit=20&offset=0`
- **Success**: 200 OK with tasks array and pagination metadata
- **Query Parameters**: See `ParseTaskFilter()` section above

**Example Request**:
```http
GET /api/tasks?status=active&priority=4,5&category=cat-work&sortBy=deadline&sortOrder=asc&limit=20&offset=0
```

**Response**:
```json
{
  "success": true,
  "data": [
    { "id": "task-1", "title": "Task 1", "priority": 5, ... },
    { "id": "task-2", "title": "Task 2", "priority": 4, ... }
  ],
  "meta": {
    "total": 45,
    "limit": 20,
    "offset": 0,
    "hasMore": true
  }
}
```

**Design Decision**: We return metadata with pagination info so clients can:
- Show "Page 1 of 3" indicators
- Implement "Load More" buttons
- Build pagination controls
- Know when to stop fetching

#### Search Tasks
- **Endpoint**: `GET /api/tasks/search?q=implement+api`
- **Success**: 200 OK with matching tasks
- **Errors**: 400 (missing query), 500 (internal)

**Example Request**:
```http
GET /api/tasks/search?q=implement api handlers
```

Searches in task title, description, and notes.

### Bulk Operations

#### Bulk Create
- **Endpoint**: `POST /api/tasks/bulk`
- **Request Body**: `BulkTaskCreateRequest` (max 50 tasks)
- **Success**: 201 Created with array of created tasks
- **Errors**: 400 (validation, size limit), 500 (internal)

**Example**:
```http
POST /api/tasks/bulk
Content-Type: application/json

{
  "tasks": [
    {
      "title": "Task 1",
      "priority": 3,
      "category": "cat-work"
    },
    {
      "title": "Task 2",
      "priority": 4,
      "category": "cat-personal"
    }
  ]
}
```

**Use Case**: Importing tasks, batch creation from templates, migration scripts.

#### Bulk Delete
- **Endpoint**: `DELETE /api/tasks/bulk`
- **Request Body**: `BulkTaskDeleteRequest`
- **Success**: 204 No Content
- **Errors**: 400 (validation), 500 (internal)

**Example**:
```http
DELETE /api/tasks/bulk
Content-Type: application/json

{
  "ids": ["task-1", "task-2", "task-3"]
}
```

**Use Case**: Deleting multiple completed tasks, cleanup operations, batch archival.

### Status Transition Operations

#### Complete Task
- **Endpoint**: `POST /api/tasks/{id}/complete`
- **Success**: 200 OK with updated task
- **Errors**: 400 (objectives incomplete, already completed, failed task), 404 (not found)

**Example**:
```http
POST /api/tasks/task-123/complete
```

**Business Rules** (enforced by service layer):
- All objectives must be completed
- Task cannot be already completed
- Task cannot be failed

#### Fail Task
- **Endpoint**: `POST /api/tasks/{id}/fail`
- **Success**: 200 OK with updated task
- **Errors**: 404 (not found), 500 (internal)

**Example**:
```http
POST /api/tasks/task-123/fail
```

**Use Case**: Mark tasks as failed when they can't be completed (cancelled, expired, no longer relevant).

#### Reactivate Task
- **Endpoint**: `POST /api/tasks/{id}/reactivate`
- **Success**: 200 OK with updated task
- **Errors**: 404 (not found), 500 (internal)

**Example**:
```http
POST /api/tasks/task-123/reactivate
```

**Use Case**: Reopen completed or failed tasks for rework.

### Reorder Operation

#### Reorder Tasks
- **Endpoint**: `POST /api/tasks/reorder`
- **Request Body**: `TaskReorderRequest` with ordered IDs
- **Success**: 204 No Content
- **Errors**: 400 (validation), 500 (internal)

**Example**:
```http
POST /api/tasks/reorder
Content-Type: application/json

{
  "ids": ["task-3", "task-1", "task-2"]
}
```

**Use Case**: Manual task ordering in UI (drag-and-drop quest list).

## Objective Handlers (`objective_handlers.go`)

Objectives are sub-tasks within a main task. They follow similar patterns to task handlers.

### Create Objective
- **Endpoint**: `POST /api/tasks/{taskId}/objectives`
- **Request Body**: `ObjectiveRequest`
- **Success**: 201 Created with objective data
- **Errors**: 400 (validation), 404 (task not found), 500 (internal)

**Example**:
```http
POST /api/tasks/task-123/objectives
Content-Type: application/json

{
  "text": "Write unit tests",
  "order": 3
}
```

**Side Effect**: Task progress is automatically recalculated after objective creation.

### Update Objective
- **Endpoint**: `PUT /api/objectives/{id}`
- **Request Body**: `ObjectiveUpdateRequest` (partial updates)
- **Success**: 200 OK with updated objective
- **Errors**: 400 (validation), 404 (not found), 500 (internal)

**Example**:
```http
PUT /api/objectives/obj-123
Content-Type: application/json

{
  "text": "Write unit and integration tests",
  "completed": true
}
```

**Side Effect**: If `completed` is changed, task progress is recalculated.

### Delete Objective
- **Endpoint**: `DELETE /api/objectives/{id}`
- **Success**: 204 No Content
- **Errors**: 404 (not found), 500 (internal)

**Side Effect**: Task progress is recalculated after deletion.

### Toggle Objective
- **Endpoint**: `POST /api/objectives/{id}/toggle`
- **Success**: 200 OK with toggled objective
- **Errors**: 404 (not found), 500 (internal)

**Example**:
```http
POST /api/objectives/obj-123/toggle
```

**Behavior**: Flips the `completed` state: `false → true` or `true → false`.

**Use Case**: Quick toggling in UI (checkbox click).

**Side Effects**:
- Task progress is recalculated
- If all objectives are completed and auto-complete is enabled, task is automatically marked as complete

## Category Handlers (`category_handlers.go`)

Categories are used to organize tasks (Main Quests, Side Quests, Projects, etc.).

### Create Category
- **Endpoint**: `POST /api/categories`
- **Request Body**: `CategoryCreateRequest`
- **Success**: 201 Created with category data
- **Errors**: 400 (validation), 500 (internal)

**Example**:
```http
POST /api/categories
Content-Type: application/json

{
  "name": "Main Quest",
  "color": "#FF5733",
  "icon": "sword",
  "type": "main"
}
```

**Validation**:
- `name`: Required, 1-100 characters
- `color`: Required, valid hex color (e.g., `#FF5733`)
- `type`: Required, must be "main" or "side"

### Get Category
- **Endpoint**: `GET /api/categories/{id}`
- **Success**: 200 OK with category data
- **Errors**: 404 (not found), 500 (internal)

### Update Category
- **Endpoint**: `PUT /api/categories/{id}`
- **Request Body**: `CategoryUpdateRequest` (partial updates)
- **Success**: 200 OK with updated category
- **Errors**: 400 (validation), 404 (not found), 500 (internal)

**Example**:
```http
PUT /api/categories/cat-123
Content-Type: application/json

{
  "name": "Main Story Quest",
  "icon": "crown"
}
```

### Delete Category
- **Endpoint**: `DELETE /api/categories/{id}`
- **Success**: 204 No Content
- **Errors**: 404 (not found), 409 (has active tasks), 500 (internal)

**Business Rule**: Cannot delete categories with active tasks (when protection is enabled in config).

**Error Response** (409 Conflict):
```json
{
  "success": false,
  "error": {
    "code": "CATEGORY_HAS_TASKS",
    "message": "Cannot delete category with active tasks"
  }
}
```

### List Categories
- **Endpoint**: `GET /api/categories`
- **Success**: 200 OK with categories array
- **Errors**: 500 (internal)

**Example Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": "cat-1",
      "name": "Main Quest",
      "color": "#FF5733",
      "icon": "sword",
      "type": "main",
      "order": 1,
      "createdAt": "2024-01-01T00:00:00Z"
    },
    {
      "id": "cat-2",
      "name": "Side Quest",
      "color": "#33C3FF",
      "icon": "scroll",
      "type": "side",
      "order": 2,
      "createdAt": "2024-01-02T00:00:00Z"
    }
  ]
}
```

## Stats Handlers (`stats_handlers.go`)

Statistics endpoints provide analytics and dashboard data.

### Get Overall Stats
- **Endpoint**: `GET /api/stats`
- **Success**: 200 OK with statistics
- **Errors**: 500 (internal)

**Example Response**:
```json
{
  "success": true,
  "data": {
    "totalTasks": 150,
    "activeTasks": 45,
    "completedTasks": 100,
    "failedTasks": 5,
    "totalRewards": 12500,
    "completionRate": 0.67,
    "averageTimeToComplete": 72.5,
    "streakDays": 14,
    "categoryStats": {
      "cat-main": 80,
      "cat-side": 70
    },
    "priorityStats": {
      "1": 20,
      "2": 30,
      "3": 40,
      "4": 35,
      "5": 25
    }
  }
}
```

**Fields Explained**:
- `totalTasks`: All tasks ever created
- `activeTasks`: Currently active tasks
- `completedTasks`: Successfully completed tasks
- `failedTasks`: Failed/cancelled tasks
- `totalRewards`: Sum of all earned rewards
- `completionRate`: Percentage of completed tasks (0.0-1.0)
- `averageTimeToComplete`: Average hours to complete a task
- `streakDays`: Consecutive days with task completions
- `categoryStats`: Task count per category
- `priorityStats`: Task count per priority level

### Get Category Stats
- **Endpoint**: `GET /api/stats/categories`
- **Success**: 200 OK with per-category statistics
- **Errors**: 500 (internal)

**Example Response**:
```json
{
  "success": true,
  "data": [
    {
      "categoryId": "cat-main",
      "totalTasks": 80,
      "completedTasks": 65,
      "completionRate": 0.8125
    },
    {
      "categoryId": "cat-side",
      "totalTasks": 70,
      "completedTasks": 35,
      "completionRate": 0.5
    }
  ]
}
```

**Use Case**: Dashboard showing progress per category (e.g., "Main Quest: 81% complete").

## Health Handlers (`health_handlers.go`)

Health check and version endpoints for monitoring and operations.

### Health Check
- **Endpoint**: `GET /health`
- **Success**: 200 OK with health status
- **Errors**: None (always succeeds if server is running)

**Example Response**:
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "version": "1.0.0",
    "uptime": 3600,
    "storage": "available"
  }
}
```

**Fields**:
- `status`: Always "ok" if responding
- `version`: Application version
- `uptime`: Seconds since server start
- `storage`: Storage availability status

**Use Case**: Load balancer health checks, monitoring systems, DevOps dashboards.

### Version Info
- **Endpoint**: `GET /version`
- **Success**: 200 OK with version information
- **Errors**: None

**Example Response**:
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "goVersion": "go1.21.0"
  }
}
```

**Use Case**: Debugging, support tickets, compatibility checks.

## HTTP Status Code Strategy

### Status Code Decision Tree

```
Request Type:
├─ POST (Create)
│  ├─ Success → 201 Created
│  ├─ Validation Error → 400 Bad Request
│  ├─ Resource Not Found (parent) → 404 Not Found
│  ├─ Business Rule Violation → 400 Bad Request
│  └─ Server Error → 500 Internal Server Error
│
├─ GET (Read)
│  ├─ Success → 200 OK
│  ├─ Not Found → 404 Not Found
│  └─ Server Error → 500 Internal Server Error
│
├─ PUT (Update)
│  ├─ Success → 200 OK
│  ├─ Validation Error → 400 Bad Request
│  ├─ Not Found → 404 Not Found
│  ├─ Conflict → 409 Conflict
│  └─ Server Error → 500 Internal Server Error
│
└─ DELETE
   ├─ Success → 204 No Content
   ├─ Not Found → 404 Not Found
   ├─ Conflict (dependencies exist) → 409 Conflict
   └─ Server Error → 500 Internal Server Error
```

### Status Codes Used

| Code | Name | Usage |
|------|------|-------|
| 200 | OK | Successful GET, PUT, or action endpoint |
| 201 | Created | Successful POST (resource created) |
| 204 | No Content | Successful DELETE (no body to return) |
| 400 | Bad Request | Validation errors, invalid input, business rule violations |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Version conflict, deletion conflict (has dependencies) |
| 500 | Internal Server Error | Unexpected errors, database failures |

**Not Used** (Future Considerations):
- 401 Unauthorized - Will be added when authentication is implemented
- 403 Forbidden - Will be added for permission-based access control
- 422 Unprocessable Entity - Could be used instead of 400 for semantic validation errors

## Design Patterns and Best Practices

### 1. Request/Response Pattern

Every handler follows this flow:
```
Parse → Validate → Execute → Respond
```

```go
func (api *API) HandlerName(w http.ResponseWriter, r *http.Request) {
    // 1. Parse: Extract parameters and decode body
    id := ExtractID(r, "/api/resource/")
    var req RequestType
    if err := DecodeJSONBody(r, &req); err != nil {
        ErrorResponse(...)
        return
    }

    // 2. Validate: Check input constraints
    if err := api.Validator.Struct(req); err != nil {
        ValidationErrorResponse(w, err)
        return
    }

    // 3. Execute: Call service layer
    result, err := api.Service.Method(r.Context(), id, req)
    if err != nil {
        HandleServiceError(w, err)
        return
    }

    // 4. Respond: Format success response
    SuccessResponse(w, http.StatusOK, result)
}
```

### 2. Early Return Pattern

Handlers use early returns for error cases:
```go
// ✅ Good: Early return on error
if err != nil {
    ErrorResponse(...)
    return
}
// Continue with success path

// ❌ Bad: Nested if-else
if err == nil {
    // Happy path deeply nested
} else {
    // Error handling
}
```

**Benefits**:
- Main success path is at the left margin (easy to read)
- Error handling is immediate and obvious
- Reduces cognitive load and nesting depth

### 3. Consistent Error Handling

All service errors go through `HandleServiceError()`:
```go
result, err := api.Service.Method(...)
if err != nil {
    HandleServiceError(w, err)  // Centralized mapping
    return
}
```

**Benefits**:
- Consistent error responses across all endpoints
- Easy to add new error mappings
- Single source of truth for HTTP status codes

### 4. Type-Safe Request/Response

Using structs for all requests and responses:
```go
// ✅ Good: Type-safe request
var req models.TaskCreateRequest
DecodeJSONBody(r, &req)
api.Validator.Struct(req)

// ❌ Bad: Map-based request
var req map[string]interface{}
json.Decode(&req)
title := req["title"].(string)  // Unsafe cast
```

**Benefits**:
- Compile-time type checking
- Auto-complete in IDEs
- Clear API contract
- Validation at struct level

### 5. Context Propagation

Always pass `r.Context()` to services:
```go
result, err := api.TaskService.CreateTask(r.Context(), req)
```

**Benefits**:
- Request cancellation support
- Timeout propagation
- Trace ID propagation (for future observability)
- Database transaction context

### 6. Interface-Based Dependencies

API depends on service interfaces, not concrete types:
```go
type API struct {
    TaskService TaskServiceInterface  // Interface, not *taskService
}
```

**Benefits**:
- Easy to mock for testing
- Flexible implementation swapping
- Decoupled layers

## Testing Considerations

### Handler Testing Strategy

Handlers will be tested using `httptest`:

```go
func TestCreateTask(t *testing.T) {
    // Arrange
    mockService := &MockTaskService{
        CreateTaskFunc: func(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error) {
            return &models.Task{ID: "task-1", Title: req.Title}, nil
        },
    }
    api := NewAPI(mockService, ..., validator.New(), "1.0.0")

    reqBody := `{"title":"Test Task","priority":3}`
    req := httptest.NewRequest("POST", "/api/tasks", strings.NewReader(reqBody))
    rec := httptest.NewRecorder()

    // Act
    api.CreateTask(rec, req)

    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
    var resp models.APIResponse
    json.Unmarshal(rec.Body.Bytes(), &resp)
    assert.True(t, resp.Success)
}
```

**Test Coverage Goals**:
- ✅ Successful requests (200, 201, 204)
- ✅ Validation errors (400)
- ✅ Not found errors (404)
- ✅ Conflict errors (409)
- ✅ Service error mapping
- ✅ JSON parsing errors
- ✅ Query parameter parsing
- ✅ Response format consistency

### Testing Tools

- `net/http/httptest` - HTTP request/response testing
- Mock services - Implement service interfaces
- `testify/assert` - Assertions
- `testify/require` - Required checks

## Error Response Examples

### 400 Bad Request - Invalid JSON
```json
{
  "success": false,
  "error": {
    "code": "INVALID_JSON",
    "message": "Invalid JSON in request body"
  }
}
```

### 400 Bad Request - Validation Error
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "title": "This field is required",
      "priority": "Value must be between 1 and 5"
    }
  }
}
```

### 400 Bad Request - Business Rule Violation
```json
{
  "success": false,
  "error": {
    "code": "OBJECTIVES_INCOMPLETE",
    "message": "Cannot complete task with incomplete objectives"
  }
}
```

### 404 Not Found
```json
{
  "success": false,
  "error": {
    "code": "TASK_NOT_FOUND",
    "message": "Task not found"
  }
}
```

### 409 Conflict - Version Conflict
```json
{
  "success": false,
  "error": {
    "code": "VERSION_CONFLICT",
    "message": "Resource was modified by another request"
  }
}
```

### 409 Conflict - Deletion Conflict
```json
{
  "success": false,
  "error": {
    "code": "CATEGORY_HAS_TASKS",
    "message": "Cannot delete category with active tasks"
  }
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An internal error occurred"
  }
}
```

## Complete API Endpoint Reference

### Tasks

| Method | Endpoint | Description | Success Status |
|--------|----------|-------------|----------------|
| POST | `/api/tasks` | Create task | 201 |
| GET | `/api/tasks/{id}` | Get task | 200 |
| PUT | `/api/tasks/{id}` | Update task | 200 |
| DELETE | `/api/tasks/{id}` | Delete task | 204 |
| GET | `/api/tasks` | List tasks with filters | 200 |
| GET | `/api/tasks/search?q={query}` | Search tasks | 200 |
| POST | `/api/tasks/bulk` | Bulk create tasks | 201 |
| DELETE | `/api/tasks/bulk` | Bulk delete tasks | 204 |
| POST | `/api/tasks/{id}/complete` | Complete task | 200 |
| POST | `/api/tasks/{id}/fail` | Fail task | 200 |
| POST | `/api/tasks/{id}/reactivate` | Reactivate task | 200 |
| POST | `/api/tasks/reorder` | Reorder tasks | 204 |

### Objectives

| Method | Endpoint | Description | Success Status |
|--------|----------|-------------|----------------|
| POST | `/api/tasks/{taskId}/objectives` | Create objective | 201 |
| PUT | `/api/objectives/{id}` | Update objective | 200 |
| DELETE | `/api/objectives/{id}` | Delete objective | 204 |
| POST | `/api/objectives/{id}/toggle` | Toggle objective | 200 |

### Categories

| Method | Endpoint | Description | Success Status |
|--------|----------|-------------|----------------|
| POST | `/api/categories` | Create category | 201 |
| GET | `/api/categories/{id}` | Get category | 200 |
| PUT | `/api/categories/{id}` | Update category | 200 |
| DELETE | `/api/categories/{id}` | Delete category | 204 |
| GET | `/api/categories` | List categories | 200 |

### Statistics

| Method | Endpoint | Description | Success Status |
|--------|----------|-------------|----------------|
| GET | `/api/stats` | Get overall stats | 200 |
| GET | `/api/stats/categories` | Get category stats | 200 |

### Health & Version

| Method | Endpoint | Description | Success Status |
|--------|----------|-------------|----------------|
| GET | `/health` | Health check | 200 |
| GET | `/version` | Version info | 200 |

## Implementation Notes

### What's Implemented ✅

1. **Core Handler Functions**: All CRUD operations for all resources
2. **Error Handling**: Comprehensive error mapping from service to HTTP
3. **Request Parsing**: JSON body parsing, URL parameter extraction, query string parsing
4. **Response Formatting**: Success and error responses with consistent format
5. **Validation**: Input validation using struct tags
6. **Filtering**: Complex query parameter parsing for task listing
7. **Pagination**: Limit/offset support with metadata
8. **Bulk Operations**: Batch create and delete endpoints
9. **Status Transitions**: Complete, fail, and reactivate operations
10. **Health Checks**: Health and version endpoints

### What's Not Implemented (Future)

1. **Router**: Need to wire handlers to HTTP paths (Task #8)
2. **Middleware**: CORS, logging, authentication, rate limiting
3. **Authentication**: JWT, API keys, session management
4. **Authorization**: Role-based access control (RBAC)
5. **Rate Limiting**: Request throttling per client
6. **Caching**: Response caching for read-heavy endpoints
7. **WebSocket**: Real-time updates for task changes
8. **File Uploads**: Attachment support for tasks
9. **Export**: CSV/JSON export endpoints
10. **Batch Updates**: Bulk update endpoint (currently only bulk create/delete)

### Next Steps

1. **Create Router** (Task #8): Wire handlers to HTTP routes
   - Use `http.ServeMux`, `gorilla/mux`, or `chi`
   - Define URL patterns
   - Apply middleware

2. **Add Middleware**: Implement cross-cutting concerns
   - CORS for frontend integration
   - Request logging
   - Recovery from panics
   - Request ID generation

3. **Write Handler Tests**: Test HTTP layer
   - Use `httptest` package
   - Mock service layer
   - Test all status codes
   - Verify response formats

4. **Main Server Entry Point** (Task #9): Bootstrap application
   - Load configuration
   - Initialize dependencies
   - Start HTTP server
   - Graceful shutdown

## Summary

The REST API layer provides a clean HTTP interface to the Quest Todo application's business logic. Key characteristics:

- **Thin Layer**: No business logic, just HTTP translation
- **Consistent**: All endpoints follow same patterns
- **Type-Safe**: Strong typing with Go structs
- **Well-Documented**: Clear endpoint definitions with examples
- **Error-Friendly**: Comprehensive error handling with proper HTTP status codes
- **Testable**: Interface-based design enables easy testing
- **RESTful**: Follows REST conventions and HTTP semantics

The implementation is complete and ready to be integrated with a router and middleware layer to create a fully functional HTTP server.
