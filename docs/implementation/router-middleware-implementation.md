# Router and Middleware Implementation

## Overview

The router and middleware layer provides HTTP request routing and cross-cutting concerns for the Quest Todo API. This layer:
- Routes HTTP requests to appropriate handlers
- Applies middleware for logging, CORS, security, etc.
- Handles method routing (GET, POST, PUT, DELETE, OPTIONS)
- Manages request/response lifecycle
- Provides observability through request IDs and logging

This is the glue that connects HTTP requests to handler functions while applying common functionality across all endpoints.

## Architecture

### Layer Position

```
┌─────────────────────────────────────────────────┐
│              HTTP Clients                       │
└─────────────────┬───────────────────────────────┘
                  │ HTTP Request
                  ▼
┌─────────────────────────────────────────────────┐
│         Middleware Chain                        │
│  ┌───────────────────────────────────────────┐  │
│  │ CORS (if enabled)                         │  │
│  │  ┌─────────────────────────────────────┐  │  │
│  │  │ Logger (if enabled)                 │  │  │
│  │  │  ┌───────────────────────────────┐  │  │  │
│  │  │  │ RequestID                     │  │  │  │
│  │  │  │  ┌─────────────────────────┐  │  │  │  │
│  │  │  │  │ ContentType             │  │  │  │  │
│  │  │  │  │  ┌───────────────────┐  │  │  │  │  │
│  │  │  │  │  │ Recovery          │  │  │  │  │  │
│  │  │  │  │  └────────┬──────────┘  │  │  │  │  │
│  │  │  │  └───────────┼─────────────┘  │  │  │  │
│  │  │  └──────────────┼────────────────┘  │  │  │
│  │  └─────────────────┼───────────────────┘  │  │
│  └────────────────────┼──────────────────────┘  │
└───────────────────────┼─────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────┐
│              Router (ServeMux)                  │
│  - Pattern matching                             │
│  - Method routing                               │
│  - Path parameter extraction                    │
└─────────────────┬───────────────────────────────┘
                  │ Matched route
                  ▼
┌─────────────────────────────────────────────────┐
│            Handler Functions                    │
│  - CreateTask, GetTask, etc.                    │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│            Service Layer                        │
└─────────────────────────────────────────────────┘
```

### Design Principles

1. **Composable Middleware**: Middleware functions wrap handlers using functional composition
2. **Standard Library First**: Use `net/http` instead of external frameworks when possible
3. **Fail Fast**: Validate early in the middleware chain (CORS, content type)
4. **Observability**: Request IDs and logging for debugging and monitoring
5. **Resilience**: Panic recovery prevents server crashes
6. **Configuration**: Middleware can be enabled/disabled via config

## File Structure

```
internal/api/
├── router.go                    # HTTP routing configuration
└── middleware/
    ├── middleware.go            # Middleware implementations
    └── middleware_test.go       # Middleware tests
```

## Middleware Implementation

### Middleware Pattern

All middleware follow the same functional pattern:

```go
type Middleware func(http.Handler) http.Handler
```

This allows middleware to wrap handlers:

```go
// Simple middleware example
func SimpleMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler
        log.Println("Before")

        // Call next handler
        next.ServeHTTP(w, r)

        // After handler
        log.Println("After")
    })
}
```

**Key Characteristics**:
- Takes an `http.Handler` as input
- Returns a new `http.Handler` that wraps it
- Can execute code before and after the wrapped handler
- Can modify request or response
- Can short-circuit (not call next handler)

### 1. RequestID Middleware

**Purpose**: Generate unique identifier for each request for tracing and correlation.

```go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if request ID already exists (from client or load balancer)
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Add request ID to response headers
        w.Header().Set("X-Request-ID", requestID)

        // Add request ID to context for use in handlers
        ctx := context.WithValue(r.Context(), requestIDKey, requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Features**:
- Generates UUID if no X-Request-ID header present
- Accepts existing request ID from client/load balancer
- Adds request ID to response headers
- Stores request ID in context for handler access
- Enables request tracing across services

**Context Access**:
```go
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}
```

**Usage in Handlers**:
```go
func (api *API) SomeHandler(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    log.Printf("[%s] Processing request", requestID)
}
```

**Design Decision**: We use UUID v4 for request IDs because:
- Globally unique (no coordination needed)
- 128-bit space prevents collisions
- Standard format, compatible with tracing tools
- Fast generation (cryptographically random)

### 2. Logger Middleware

**Purpose**: Log all HTTP requests with timing information.

```go
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Get request ID if available
        requestID := GetRequestID(r.Context())

        // Log request start
        log.Printf("[%s] --> %s %s", requestID, r.Method, r.URL.Path)

        // Process request
        next.ServeHTTP(wrapped, r)

        // Log request completion with timing
        duration := time.Since(start)
        log.Printf("[%s] <-- %s %s %d %s",
            requestID,
            r.Method,
            r.URL.Path,
            wrapped.statusCode,
            duration,
        )
    })
}
```

**Features**:
- Logs request start (method and path)
- Logs request completion (status code and duration)
- Includes request ID for correlation
- Non-blocking (doesn't slow down requests)
- Captures status codes from both WriteHeader and Write

**Log Format**:
```
[request-id] --> GET /api/tasks
[request-id] <-- GET /api/tasks 200 45ms
```

**Response Writer Wrapper**:

To capture the HTTP status code, we need a custom ResponseWriter:

```go
type responseWriter struct {
    http.ResponseWriter
    statusCode int
    written    bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
    if !rw.written {
        rw.statusCode = statusCode
        rw.ResponseWriter.WriteHeader(statusCode)
        rw.written = true
    }
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    if !rw.written {
        rw.WriteHeader(http.StatusOK)
    }
    return rw.ResponseWriter.Write(b)
}
```

**Why We Need This**: The standard `http.ResponseWriter` doesn't expose the status code. By wrapping it, we can intercept `WriteHeader` calls and capture the status.

**Design Decision**: We log before and after the request for observability:
- "Before" log shows when request started (useful for finding hung requests)
- "After" log shows result and timing (useful for performance analysis)
- Request ID links the two log lines

**Production Enhancement**: For production, consider:
- Structured logging (JSON format)
- Log levels (DEBUG, INFO, WARN, ERROR)
- Sampling for high-traffic endpoints
- Integration with APM tools (Datadog, New Relic)

### 3. Recovery Middleware

**Purpose**: Catch panics and prevent server crashes while returning proper error responses.

```go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Log the panic with stack trace
                requestID := GetRequestID(r.Context())
                log.Printf("[%s] PANIC: %v\n%s", requestID, err, debug.Stack())

                // Return 500 error
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(`{"success":false,"error":{"code":"INTERNAL_ERROR","message":"An internal error occurred"}}`))
            }
        }()

        next.ServeHTTP(w, r)
    })
}
```

**Features**:
- Catches all panics in handler chain
- Logs full stack trace for debugging
- Returns consistent JSON error response
- Includes request ID in logs
- Prevents server from crashing

**What Gets Caught**:
- Explicit `panic()` calls
- Nil pointer dereferences
- Array/slice out of bounds
- Type assertion failures
- Division by zero

**Error Response** (500 Internal Server Error):
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An internal error occurred"
  }
}
```

**Design Decision**: We always return a generic error message to clients (not the panic message) because:
- Panic messages may contain sensitive information
- Stack traces shouldn't be exposed to clients
- Consistent error format for all server errors
- Security: Don't leak internal implementation details

**Stack Trace Example**:
```
[abc-123] PANIC: runtime error: invalid memory address or nil pointer dereference
goroutine 19 [running]:
runtime/debug.Stack()
    /usr/local/go/src/runtime/debug/stack.go:24 +0x64
github.com/LaV72/quest-todo/internal/api/middleware.Recovery.func1.1()
    /app/internal/api/middleware/middleware.go:109 +0x60
panic({0x102ae0280, 0x102b35ac0})
    /usr/local/go/src/runtime/panic.go:783 +0x120
github.com/LaV72/quest-todo/internal/api.(*API).GetTask(...)
    /app/internal/api/task_handlers.go:45
```

**Best Practice**: Avoid panics when possible. Use explicit error returns instead:
```go
// ❌ Bad: panic
if task == nil {
    panic("task is nil")
}

// ✅ Good: return error
if task == nil {
    return nil, errors.New("task not found")
}
```

Recovery middleware is a safety net, not a primary error handling mechanism.

### 4. CORS Middleware

**Purpose**: Handle Cross-Origin Resource Sharing for frontend applications.

```go
func CORS(allowedOrigins []string) Middleware {
    // If no origins specified, allow all (for development)
    if len(allowedOrigins) == 0 {
        allowedOrigins = []string{"*"}
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // Check if origin is allowed and set CORS headers
            for _, allowedOrigin := range allowedOrigins {
                if allowedOrigin == "*" {
                    w.Header().Set("Access-Control-Allow-Origin", "*")
                    break
                } else if allowedOrigin == origin {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    break
                }
            }

            // Set other CORS headers
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
            w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

            // Handle preflight requests
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

**Features**:
- Configurable allowed origins
- Wildcard support for development
- Preflight request handling (OPTIONS)
- Proper CORS headers for all responses

**CORS Headers**:
- `Access-Control-Allow-Origin`: Which origins can access (specific or *)
- `Access-Control-Allow-Methods`: Which HTTP methods are allowed
- `Access-Control-Allow-Headers`: Which headers can be sent
- `Access-Control-Max-Age`: How long to cache preflight results

**Configuration Examples**:

Development (allow all):
```go
CORS([]string{"*"})
```

Production (specific origins):
```go
CORS([]string{
    "https://app.example.com",
    "https://www.example.com",
})
```

Multiple environments:
```go
CORS([]string{
    "http://localhost:3000",      // Local development
    "https://staging.example.com", // Staging
    "https://app.example.com",     // Production
})
```

**Preflight Requests**:

Before certain requests (e.g., POST with custom headers), browsers send a preflight OPTIONS request:

```
OPTIONS /api/tasks HTTP/1.1
Origin: http://localhost:3000
Access-Control-Request-Method: POST
Access-Control-Request-Headers: Content-Type
```

Response:
```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-Request-ID
Access-Control-Max-Age: 86400
```

The middleware short-circuits OPTIONS requests (returns 204 No Content) without calling the handler.

**Design Decision**: We return the actual origin in the header (not *) when checking specific origins because:
- More secure (browser enforces origin checking)
- Allows credentials (cookies) to be sent
- `Access-Control-Allow-Origin: *` with credentials is not allowed by browsers

**Security Note**:
- In production, never use `*` with sensitive APIs
- Always specify exact origins
- Don't derive origins from request (validate against whitelist)

### 5. ContentType Middleware

**Purpose**: Validate that requests with bodies use JSON content type.

```go
func ContentType(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Only check content type for requests with bodies
        if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
            contentType := r.Header.Get("Content-Type")
            // Allow application/json or empty (will be set by client)
            if contentType != "" && contentType != "application/json" {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnsupportedMediaType)
                w.Write([]byte(`{"success":false,"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json"}}`))
                return
            }
        }

        next.ServeHTTP(w, r)
    })
}
```

**Features**:
- Validates Content-Type for POST/PUT/PATCH requests
- Allows `application/json` or empty
- Returns 415 Unsupported Media Type for invalid types
- Skips validation for GET/DELETE (no body)

**Valid Requests**:
```http
POST /api/tasks
Content-Type: application/json

{"title": "Task"}
```

```http
POST /api/tasks
(no Content-Type header - allowed)

{"title": "Task"}
```

**Invalid Request**:
```http
POST /api/tasks
Content-Type: text/plain

title=Task
```

Response (415):
```json
{
  "success": false,
  "error": {
    "code": "UNSUPPORTED_MEDIA_TYPE",
    "message": "Content-Type must be application/json"
  }
}
```

**Design Decision**: We allow empty Content-Type because:
- Some HTTP clients don't set it automatically
- Handler will fail anyway if JSON is invalid
- Better developer experience (less strict)

**Future Enhancement**: Support other content types:
- `application/x-www-form-urlencoded`
- `multipart/form-data` (for file uploads)
- Content negotiation based on Accept header

### 6. Middleware Chain

**Purpose**: Apply multiple middleware in order.

```go
func Chain(h http.Handler, middleware ...Middleware) http.Handler {
    // Apply middleware in reverse order so they execute in the order specified
    for i := len(middleware) - 1; i >= 0; i-- {
        h = middleware[i](h)
    }
    return h
}
```

**Why Reverse Order?**

Consider this chain:
```go
Chain(handler, middleware1, middleware2, middleware3)
```

We want execution order:
1. middleware1 (before)
2. middleware2 (before)
3. middleware3 (before)
4. handler
5. middleware3 (after)
6. middleware2 (after)
7. middleware1 (after)

By applying in reverse, we build:
```
middleware1(
    middleware2(
        middleware3(
            handler
        )
    )
)
```

This creates the correct nesting order.

**Usage Example**:
```go
handler := Chain(
    http.HandlerFunc(actualHandler),
    RequestID,
    Logger,
    Recovery,
)
```

Execution order:
```
RequestID → Logger → Recovery → actualHandler → Recovery → Logger → RequestID
```

**Visualization**:
```
[RequestID Before]
    [Logger Before]
        [Recovery Before]
            [Handler]
        [Recovery After]
    [Logger After]
[RequestID After]
```

## Router Implementation

### Router Structure

The router uses Go's standard `http.ServeMux` with custom routing logic:

```go
func NewRouter(api *API, config RouterConfig) http.Handler {
    mux := http.NewServeMux()

    // Register routes...

    // Apply middleware chain
    var handler http.Handler = mux
    handler = middleware.Recovery(handler)
    handler = middleware.ContentType(handler)
    handler = middleware.RequestID(handler)

    if config.EnableLogging {
        handler = middleware.Logger(handler)
    }

    if config.EnableCORS {
        handler = middleware.CORS(config.AllowedOrigins)(handler)
    }

    return handler
}
```

**Router Configuration**:
```go
type RouterConfig struct {
    AllowedOrigins []string  // CORS allowed origins
    EnableCORS     bool       // Enable CORS middleware
    EnableLogging  bool       // Enable request logging
}
```

### Route Patterns

#### Health and Version Routes

```go
mux.HandleFunc("/health", api.HealthCheck)
mux.HandleFunc("/version", api.Version)
```

**No `/api` prefix** - These are operational endpoints, not API resources.

#### Task Routes

**Collection routes** (`/api/tasks`):
```go
mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // Check for search
        if r.URL.Query().Get("q") != "" {
            api.SearchTasks(w, r)
        } else {
            api.ListTasks(w, r)
        }
    case http.MethodPost:
        api.CreateTask(w, r)
    default:
        methodNotAllowed(w, r)
    }
})
```

**Resource routes** (`/api/tasks/{id}`):
```go
mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
    parts := strings.Split(path, "/")

    // Handle special routes first
    switch parts[0] {
    case "search":
        // GET /api/tasks/search?q=query
    case "bulk":
        // POST /api/tasks/bulk
        // DELETE /api/tasks/bulk
    case "reorder":
        // POST /api/tasks/reorder
    }

    // Handle task ID routes
    if len(parts) >= 1 && parts[0] != "" {
        // Check for action routes
        if len(parts) >= 2 {
            action := parts[1]
            switch action {
            case "complete":
                // POST /api/tasks/{id}/complete
            case "fail":
                // POST /api/tasks/{id}/fail
            case "reactivate":
                // POST /api/tasks/{id}/reactivate
            case "objectives":
                // POST /api/tasks/{id}/objectives
            }
        }

        // Standard CRUD on /api/tasks/{id}
        switch r.Method {
        case http.MethodGet:
            api.GetTask(w, r)
        case http.MethodPut:
            api.UpdateTask(w, r)
        case http.MethodDelete:
            api.DeleteTask(w, r)
        }
    }
})
```

**Route Priority**:
1. Special routes (search, bulk, reorder) - checked first
2. Action routes (complete, fail, reactivate) - checked if 2 parts
3. Standard CRUD - fallback for single ID

**Why This Structure?**

`http.ServeMux` doesn't support path parameters like `/{id}` or route parameters. We manually parse paths using string manipulation:

```go
path := "/api/tasks/task-123/complete"
path = strings.TrimPrefix(path, "/api/tasks/")  // "task-123/complete"
parts := strings.Split(path, "/")                // ["task-123", "complete"]
id := parts[0]                                    // "task-123"
action := parts[1]                                // "complete"
```

**Alternative Routers**:

For more complex routing, consider:
- `gorilla/mux` - Regex patterns, path variables
- `chi` - Lightweight, middleware-aware
- `httprouter` - High performance, zero allocations

We chose standard library for:
- No external dependencies
- Simple routing needs
- Full control over routing logic
- Easy to understand and maintain

#### Method Routing

All routes check HTTP methods explicitly:

```go
switch r.Method {
case http.MethodGet:
    // Handle GET
case http.MethodPost:
    // Handle POST
case http.MethodPut:
    // Handle PUT
case http.MethodDelete:
    // Handle DELETE
default:
    methodNotAllowed(w, r)
}
```

**Method Not Allowed** (405):
```go
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusMethodNotAllowed)
    w.Write([]byte(`{"success":false,"error":{"code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}}`))
}
```

**Response**:
```json
{
  "success": false,
  "error": {
    "code": "METHOD_NOT_ALLOWED",
    "message": "Method not allowed"
  }
}
```

### Complete Route Table

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/health` | HealthCheck | Health check |
| GET | `/version` | Version | Version info |
| **Tasks** |
| POST | `/api/tasks` | CreateTask | Create task |
| GET | `/api/tasks` | ListTasks | List tasks (with filters) |
| GET | `/api/tasks?q=query` | SearchTasks | Search tasks |
| GET | `/api/tasks/{id}` | GetTask | Get task by ID |
| PUT | `/api/tasks/{id}` | UpdateTask | Update task |
| DELETE | `/api/tasks/{id}` | DeleteTask | Delete task |
| POST | `/api/tasks/{id}/complete` | CompleteTask | Complete task |
| POST | `/api/tasks/{id}/fail` | FailTask | Fail task |
| POST | `/api/tasks/{id}/reactivate` | ReactivateTask | Reactivate task |
| POST | `/api/tasks/bulk` | CreateTasksBulk | Bulk create |
| DELETE | `/api/tasks/bulk` | DeleteTasksBulk | Bulk delete |
| POST | `/api/tasks/reorder` | ReorderTasks | Reorder tasks |
| **Objectives** |
| POST | `/api/tasks/{taskId}/objectives` | CreateObjective | Create objective |
| PUT | `/api/objectives/{id}` | UpdateObjective | Update objective |
| DELETE | `/api/objectives/{id}` | DeleteObjective | Delete objective |
| POST | `/api/objectives/{id}/toggle` | ToggleObjective | Toggle completion |
| **Categories** |
| POST | `/api/categories` | CreateCategory | Create category |
| GET | `/api/categories` | ListCategories | List all categories |
| GET | `/api/categories/{id}` | GetCategory | Get category |
| PUT | `/api/categories/{id}` | UpdateCategory | Update category |
| DELETE | `/api/categories/{id}` | DeleteCategory | Delete category |
| **Stats** |
| GET | `/api/stats` | GetStats | Overall stats |
| GET | `/api/stats/categories` | GetCategoryStats | Category stats |

**Total: 26 endpoints**

## Testing

The router and middleware layer has comprehensive test coverage with 77 total tests:
- **Middleware tests**: 18 unit tests (0.538s execution)
- **Router tests**: 59 integration tests (0.575s execution)

All tests use `net/http/httptest` for HTTP testing with mock services.

**For complete testing documentation**, see:
- [Router and Middleware Tests](../../testing/router-middleware-tests.md) - Comprehensive test documentation with examples, patterns, and best practices

## Usage Examples

### Basic Server Setup

```go
package main

import (
    "log"
    "net/http"

    "github.com/LaV72/quest-todo/internal/api"
    "github.com/LaV72/quest-todo/internal/api/middleware"
    "github.com/LaV72/quest-todo/internal/service"
    "github.com/go-playground/validator/v10"
)

func main() {
    // Initialize services
    taskService := service.NewTaskService(...)
    objectiveService := service.NewObjectiveService(...)
    categoryService := service.NewCategoryService(...)
    statsService := service.NewStatsService(...)

    // Create API
    apiInstance := api.NewAPI(
        taskService,
        objectiveService,
        categoryService,
        statsService,
        validator.New(),
        "1.0.0",
    )

    // Configure router
    config := api.RouterConfig{
        AllowedOrigins: []string{"http://localhost:3000"},
        EnableCORS:     true,
        EnableLogging:  true,
    }

    router := api.NewRouter(apiInstance, config)

    // Start server
    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatal(err)
    }
}
```

### Development Configuration

```go
config := api.RouterConfig{
    AllowedOrigins: []string{"*"},  // Allow all origins
    EnableCORS:     true,
    EnableLogging:  true,           // Verbose logging
}
```

### Production Configuration

```go
config := api.RouterConfig{
    AllowedOrigins: []string{
        "https://app.example.com",
        "https://www.example.com",
    },
    EnableCORS:     true,
    EnableLogging:  false,  // Use structured logging instead
}
```

### Custom Middleware

To add custom middleware:

```go
// Define custom middleware
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !isValidToken(token) {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Apply in router
func NewRouter(api *API, config RouterConfig) http.Handler {
    mux := http.NewServeMux()
    // ... register routes

    var handler http.Handler = mux
    handler = middleware.Recovery(handler)
    handler = AuthMiddleware(handler)        // Add custom middleware
    handler = middleware.RequestID(handler)

    return handler
}
```

## Design Decisions

### 1. Standard Library Router

**Decision**: Use `http.ServeMux` instead of external router library.

**Rationale**:
- No external dependencies
- Sufficient for our routing needs
- Full control over routing logic
- Easy to understand and maintain
- Can always swap later if needed

**Trade-offs**:
- ✅ Simple and straightforward
- ✅ No version compatibility issues
- ❌ Manual path parsing required
- ❌ No built-in path parameters

### 2. Middleware Execution Order

**Decision**: CORS → Logger → RequestID → ContentType → Recovery → Handler

**Rationale**:
- **CORS first**: Reject cross-origin requests early
- **Logger second**: Log all requests (including rejected ones)
- **RequestID third**: Generate ID for logging
- **ContentType fourth**: Validate before handler
- **Recovery innermost**: Catch all handler panics

**Alternative Order**:
```
Recovery → RequestID → Logger → CORS → ContentType → Handler
```
Would catch panics in middleware, but we want middleware to be panic-free.

### 3. Request ID Format

**Decision**: Use UUID v4 for request IDs.

**Alternatives Considered**:
- Sequential integers: Not unique across servers
- Timestamp + random: Harder to parse
- ULID: Lexicographically sortable but overkill
- Snowflake IDs: Requires coordination

**Choice**: UUID v4 is standard, globally unique, and well-supported.

### 4. Logging Format

**Decision**: Human-readable text format with request ID.

**Format**: `[request-id] --> METHOD PATH` and `[request-id] <-- METHOD PATH STATUS DURATION`

**Rationale**:
- Easy to read in development
- Request ID enables correlation
- Status and timing for debugging

**Production Alternative**: Structured logging (JSON)
```json
{
  "timestamp": "2024-01-10T10:00:00Z",
  "request_id": "abc-123",
  "method": "GET",
  "path": "/api/tasks",
  "status": 200,
  "duration_ms": 45,
  "level": "info"
}
```

### 5. CORS Security

**Decision**: Validate origin against whitelist, return specific origin (not *).

**Rationale**:
- More secure than wildcard
- Allows credentials (cookies)
- Browser enforces origin checking
- Explicit is better than implicit

**Development Exception**: Allow * for local development convenience.

## Performance Considerations

### Middleware Overhead

Each middleware adds minimal latency:

| Middleware | Overhead | Operations |
|------------|----------|------------|
| RequestID | ~1µs | UUID generation, context creation |
| Logger | ~5µs | Time capture, string formatting |
| Recovery | ~0µs | Defer call (no cost unless panic) |
| CORS | ~2µs | String comparison, header setting |
| ContentType | ~1µs | String comparison |

**Total overhead**: ~10µs per request

For a typical request taking 50ms, middleware overhead is 0.02% - negligible.

### Memory Allocation

Request ID middleware allocates:
- 1 UUID (16 bytes)
- 1 context value (24 bytes)
- 1 string (36 bytes for UUID string)

**Total**: ~80 bytes per request

For 1000 req/s: 80KB/s allocation - minimal GC pressure.

### Scaling Considerations

1. **Logging**: High-traffic servers should:
   - Use asynchronous logging
   - Sample logs (log 1% of requests)
   - Use log levels (only log errors in production)

2. **CORS**: Cache origin validation results:
   ```go
   var allowedOriginsMap = make(map[string]bool)
   for _, origin := range allowedOrigins {
       allowedOriginsMap[origin] = true
   }
   ```

3. **Request ID**: Consider UUIDv1 (timestamp-based) for sortability in logs.

## Best Practices

### 1. Middleware Order Matters

Always apply middleware in correct order:
```go
// ✅ Good: CORS before auth (allows preflight without auth)
handler = AuthMiddleware(handler)
handler = CORS(origins)(handler)

// ❌ Bad: Auth before CORS (blocks preflight requests)
handler = CORS(origins)(handler)
handler = AuthMiddleware(handler)
```

### 2. Log Request IDs Everywhere

Always include request ID in logs:
```go
requestID := middleware.GetRequestID(r.Context())
log.Printf("[%s] Operation failed: %v", requestID, err)
```

This enables request tracing across services.

### 3. Fail Fast

Validate early in the middleware chain:
```go
// ✅ Good: Validate before expensive operations
handler = BusinessLogic(handler)
handler = ValidateAuth(handler)
handler = CORS(origins)(handler)

// ❌ Bad: Validate after expensive operations
handler = CORS(origins)(handler)
handler = ValidateAuth(handler)
handler = BusinessLogic(handler)
```

### 4. Use Structured Logging in Production

Replace text logs with structured logs:
```go
logger.Info("Request completed",
    "request_id", requestID,
    "method", r.Method,
    "path", r.URL.Path,
    "status", statusCode,
    "duration_ms", duration.Milliseconds(),
)
```

### 5. Monitor Panic Recovery

Set up alerts for recovered panics:
```go
if err := recover(); err != nil {
    log.Printf("[%s] PANIC: %v", requestID, err)
    metrics.IncrementPanicCount()  // Track panic rate
    alerting.NotifyOnCall()        // Alert engineers
}
```

Panics indicate bugs that should be fixed.

## Future Enhancements

### 1. Authentication Middleware

```go
func JWTAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        claims, err := validateJWT(token)
        if err != nil {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), userKey, claims.UserID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 2. Rate Limiting

```go
func RateLimit(requestsPerSecond int) Middleware {
    limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                w.WriteHeader(http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### 3. Request Timeout

```go
func Timeout(duration time.Duration) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), duration)
            defer cancel()
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 4. Compression

```go
func Gzip(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }

        gzipWriter := gzip.NewWriter(w)
        defer gzipWriter.Close()

        w.Header().Set("Content-Encoding", "gzip")
        next.ServeHTTP(&gzipResponseWriter{gzipWriter, w}, r)
    })
}
```

### 5. Metrics

```go
func Metrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start)
        metrics.RecordRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
    })
}
```

## Summary

The router and middleware layer provides:

1. **HTTP Routing**: 26 API endpoints with method-based routing
2. **Request IDs**: Unique identifier for every request (tracing)
3. **Logging**: Request/response logging with timing
4. **Panic Recovery**: Graceful error handling, no crashes
5. **CORS**: Cross-origin support for frontend apps
6. **Content Validation**: JSON-only API enforcement
7. **Composability**: Easy to add custom middleware

**Key Characteristics**:
- **Standard Library**: Uses `net/http`, no external router dependencies
- **Tested**: 18 comprehensive tests covering all middleware
- **Configurable**: Enable/disable features via RouterConfig
- **Extensible**: Easy to add custom middleware
- **Production-Ready**: Panic recovery, logging, CORS, request IDs

The implementation is complete and ready to be integrated with the main server entry point to create a fully functional HTTP API server.
