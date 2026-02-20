# Main Server Implementation

This document explains the implementation of the main server entry point (`cmd/server/main.go`), which bootstraps the entire application and wires all layers together.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Configuration System](#configuration-system)
4. [Dependency Injection](#dependency-injection)
5. [Server Initialization](#server-initialization)
6. [Graceful Shutdown](#graceful-shutdown)
7. [Logging](#logging)
8. [Error Handling](#error-handling)
9. [Production Considerations](#production-considerations)
10. [Best Practices](#best-practices)

## Overview

The main server entry point is responsible for:

1. **Configuration Management** - Loading configuration from environment variables
2. **Dependency Injection** - Wiring up all layers with proper dependencies
3. **Server Lifecycle** - Starting the HTTP server and handling shutdown
4. **Error Handling** - Handling initialization and runtime errors
5. **Logging** - Providing visibility into server state and operations

### Key Design Principles

- **Fail Fast** - Fatal errors during initialization prevent the server from starting
- **Graceful Degradation** - Runtime errors are logged but don't crash the server (handled by recovery middleware)
- **Explicit Dependencies** - All dependencies are created and injected explicitly (no globals)
- **Configuration via Environment** - All configuration through environment variables (12-factor app principle)
- **Signal Handling** - Proper shutdown on SIGINT/SIGTERM for clean exits

## Architecture

### Initialization Flow

```
main()
├── 1. Load Configuration (env vars)
├── 2. Initialize Storage Layer
│   ├── Open SQLite connection
│   ├── Configure connection pool
│   ├── Test connection (Ping)
│   └── Run migrations (automatic)
├── 3. Initialize Shared Dependencies
│   ├── Clock (RealClock)
│   ├── ID Generator (UUIDGenerator)
│   └── Validator (go-playground/validator)
├── 4. Initialize Service Layer
│   ├── TaskService
│   ├── ObjectiveService
│   ├── CategoryService
│   └── StatsService
├── 5. Initialize API Layer
│   └── API handlers (with services + validator)
├── 6. Initialize Router + Middleware
│   ├── Route registration
│   └── Middleware chain
├── 7. Start HTTP Server
│   ├── Listen on configured address
│   └── Handle requests
└── 8. Wait for Shutdown Signal
    ├── Graceful shutdown (30s timeout)
    └── Close database connection
```

### Layer Dependencies

```
┌─────────────────────────────────────────┐
│         HTTP Server (net/http)          │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│        Router + Middleware              │
│  (RequestID → Logger → ContentType      │
│   → Recovery → CORS)                    │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│           API Handlers                  │
│  (Request parsing, validation,          │
│   response formatting)                  │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│          Service Layer                  │
│  (Business logic, validation,           │
│   progress calculation)                 │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│          Storage Layer                  │
│  (SQLite, migrations, queries)          │
└─────────────────────────────────────────┘
```

## Configuration System

### Configuration Structure

All configuration is managed through a hierarchical struct:

```go
type Config struct {
    Server   ServerConfig      // HTTP server settings
    Service  service.Config    // Business logic configuration
    Router   api.RouterConfig  // Router/middleware settings
    Database DatabaseConfig    // Database connection settings
}
```

### Environment Variables

#### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `localhost` | Server host address |
| `PORT` | `8080` | Server port |
| `READ_TIMEOUT` | `15s` | Maximum duration for reading request |
| `WRITE_TIMEOUT` | `15s` | Maximum duration for writing response |
| `IDLE_TIMEOUT` | `60s` | Maximum idle time before closing connection |

**Why these defaults?**
- `localhost` - Safe default for development (not exposed externally)
- `8080` - Common alternative to 80 for unprivileged users
- 15s timeouts - Balance between slow clients and preventing resource exhaustion
- 60s idle - Standard HTTP keep-alive timeout

#### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_PATH` | `./quest-todo.db` | Path to SQLite database file |

**Why SQLite?**
- Zero configuration (no separate database server)
- Single file deployment
- Excellent performance for single-server apps
- ACID compliance with WAL mode

#### Service Layer Configuration

| Variable | Default | Description | Rationale |
|----------|---------|-------------|-----------|
| `MAX_TITLE_LENGTH` | `200` | Max task title length | Fits on one line in most UIs |
| `MAX_DESCRIPTION_LENGTH` | `2000` | Max description length | ~300 words, reasonable detail |
| `MAX_BULK_SIZE` | `100` | Max bulk operation size | Prevents memory issues |
| `REQUIRE_ALL_OBJECTIVES` | `false` | Must complete all objectives? | Flexible default |
| `AUTO_COMPLETE_ON_FULL_PROGRESS` | `true` | Auto-complete at 100%? | Convenient default |
| `ALLOW_PAST_DEADLINES` | `true` | Allow past dates? | Flexible for retroactive entry |
| `ENABLE_CATEGORY_RESTRICTIONS` | `false` | Enable category rules? | Optional strictness |
| `ENABLE_REWARD_SYSTEM` | `false` | Enable rewards? | Future feature flag |

#### Router/Middleware Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins (comma-separated) |
| `ENABLE_CORS` | `true` | Enable CORS middleware |
| `ENABLE_LOGGING` | `true` | Enable request/response logging |

**CORS Default:**
- `*` (wildcard) is convenient for development
- **Must** be changed to specific origins in production for security

### Configuration Loading

```go
func loadConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Host:         getEnv("HOST", defaultHost),
            Port:         getEnv("PORT", defaultPort),
            ReadTimeout:  getDurationEnv("READ_TIMEOUT", defaultReadTimeout),
            WriteTimeout: getDurationEnv("WRITE_TIMEOUT", defaultWriteTimeout),
            IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", defaultIdleTimeout),
        },
        // ... other configurations
    }
}
```

**Helper Functions:**

- `getEnv(key, default)` - String values
- `getIntEnv(key, default)` - Integer values (with parsing)
- `getBoolEnv(key, default)` - Boolean values (true/1/yes)
- `getDurationEnv(key, default)` - Duration values (Go duration format)
- `parseOrigins(s)` - Comma-separated list parsing

**Why Environment Variables?**

1. **12-Factor App Principle** - Configuration in environment, not code
2. **No Config Files to Manage** - Simpler deployment
3. **Container-Friendly** - Easy to configure in Docker/Kubernetes
4. **Security** - Secrets via environment, not checked into version control
5. **Platform Agnostic** - Works on any OS

### Configuration Validation

No explicit validation in `loadConfig()` because:

1. **Type Safety** - Helper functions handle parsing with fallback to defaults
2. **Fail Fast** - Invalid values use defaults (logged if needed)
3. **Service Layer Validation** - Business rules validated by services
4. **Runtime Validation** - Connection failures caught during initialization

For strict validation, you could add:

```go
func (c *Config) Validate() error {
    if c.Server.Port == "" {
        return errors.New("PORT cannot be empty")
    }
    if c.Database.Path == "" {
        return errors.New("DB_PATH cannot be empty")
    }
    // ... more validation
    return nil
}
```

## Dependency Injection

### Explicit Dependency Injection

All dependencies are created explicitly and injected into components that need them:

```go
// 1. Shared dependencies
clock := RealClock{}
idGen := UUIDGenerator{}
validate := validator.New()

// 2. Service layer (depends on storage + shared deps)
taskService := service.NewTaskService(storage, clock, idGen, validate, &config.Service)
objectiveService := service.NewObjectiveService(storage, clock, idGen, validate, &config.Service)
categoryService := service.NewCategoryService(storage, validate, &config.Service)
statsService := service.NewStatsService(storage)

// 3. API layer (depends on services)
apiHandler := api.NewAPI(
    taskService,
    objectiveService,
    categoryService,
    statsService,
    validate,
    Version,
)

// 4. Router layer (depends on API)
router := api.NewRouter(apiHandler, config.Router)
```

**Benefits:**

- ✅ **Testable** - Easy to inject mocks/fakes for testing
- ✅ **Explicit** - All dependencies visible in code
- ✅ **Type-Safe** - Compile-time checking of dependencies
- ✅ **No Magic** - No reflection, no dependency injection framework
- ✅ **Simple** - Easy to understand and debug

**Comparison with Other Approaches:**

| Approach | Pros | Cons |
|----------|------|------|
| **Manual DI** (current) | Simple, explicit, no magic | More boilerplate code |
| **DI Framework** (wire, fx) | Less boilerplate | Adds dependency, uses reflection |
| **Service Locator** | Flexible | Hidden dependencies, harder to test |
| **Global Variables** | Convenient | Not testable, implicit dependencies |

### Interface-Based Dependencies

Services depend on interfaces, not concrete types:

```go
// TaskService depends on storage.Storage interface
func NewTaskService(storage storage.Storage, ...) TaskService {
    // Can be SQLite, Memory, JSON, or any storage.Storage implementation
}

// API depends on service interfaces (inline)
type API struct {
    TaskService interface {
        CreateTask(ctx context.Context, req models.TaskCreateRequest) (*models.Task, error)
        // ...
    }
}
```

**Benefits:**

- ✅ **Swappable** - Easy to swap implementations
- ✅ **Testable** - Easy to mock with minimal code
- ✅ **Decoupled** - Layers don't depend on concrete types

### Clock and ID Generator Interfaces

Custom implementations for testability:

```go
// RealClock implements service.Clock interface
type RealClock struct{}

func (RealClock) Now() time.Time {
    return time.Now()
}

// UUIDGenerator implements service.IDGenerator interface
type UUIDGenerator struct{}

func (UUIDGenerator) Generate() string {
    return uuid.New().String()
}
```

**Why?**

In production, we use real implementations. In tests, we can inject:

- **Fake Clock** - Control time for deadline testing
- **Sequential ID Generator** - Predictable IDs for test assertions

This is a common pattern in Go for testability without heavyweight mocking frameworks.

## Server Initialization

### HTTP Server Configuration

```go
server := &http.Server{
    Addr:         config.Server.Address(),  // "host:port"
    Handler:      router,                    // Middleware-wrapped router
    ReadTimeout:  config.Server.ReadTimeout, // Prevent slow clients
    WriteTimeout: config.Server.WriteTimeout,// Prevent slow responses
    IdleTimeout:  config.Server.IdleTimeout, // Connection keep-alive
}
```

**Timeout Configuration:**

- **ReadTimeout** - Maximum time to read entire request (headers + body)
  - Prevents slow-loris attacks (slow clients holding connections)
  - Default: 15s (sufficient for normal requests)

- **WriteTimeout** - Maximum time to write response
  - Prevents slow clients from blocking server goroutines
  - Default: 15s (sufficient for database queries + JSON encoding)

- **IdleTimeout** - Maximum idle time before closing connection
  - Implements HTTP keep-alive timeout
  - Default: 60s (standard HTTP keep-alive)

**Why these timeouts matter:**

Without timeouts, a malicious or buggy client can:
- Hold server resources indefinitely
- Exhaust server goroutines
- Cause memory leaks from incomplete requests

### Goroutine-Based Server Start

```go
serverErrors := make(chan error, 1)
go func() {
    log.Printf("Server listening on http://%s", server.Addr)
    serverErrors <- server.ListenAndServe()
}()
```

**Why a goroutine?**

- `ListenAndServe()` blocks until the server stops
- We need to listen for shutdown signals concurrently
- Use buffered channel (size 1) to prevent goroutine leak

### Database Connection Management

```go
storage, err := sqlite.New(config.Database.Path)
if err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}
defer func() {
    if err := storage.Close(); err != nil {
        log.Printf("Error closing database: %v", err)
    }
}()
```

**Initialization:**

- `sqlite.New()` automatically:
  1. Opens connection with WAL mode
  2. Configures connection pool (max 1 writer for SQLite)
  3. Tests connection with Ping()
  4. Runs migrations to create/update schema

**Cleanup:**

- `defer storage.Close()` ensures cleanup even if main() panics
- Log error but don't fail (server is already shutting down)

**Connection Pooling:**

SQLite-specific configuration in `sqlite.New()`:

```go
db.SetMaxOpenConns(1)    // SQLite: only 1 writer at a time
db.SetMaxIdleConns(1)    // Keep 1 connection alive
db.SetConnMaxLifetime(0) // Connections never expire
```

This prevents "database locked" errors with SQLite.

## Graceful Shutdown

### Signal Handling

```go
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

select {
case err := <-serverErrors:
    log.Fatalf("Server error: %v", err)

case sig := <-shutdown:
    log.Printf("Received signal: %v", sig)
    // ... graceful shutdown
}
```

**Why handle signals?**

- **SIGINT** (Ctrl+C) - User-initiated shutdown
- **SIGTERM** - System/container shutdown (Docker, Kubernetes)
- Without signal handling, shutdown is immediate (connections dropped)

### Graceful Shutdown Process

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Printf("Error during shutdown: %v", err)
    if err := server.Close(); err != nil {
        log.Printf("Error forcing shutdown: %v", err)
    }
} else {
    log.Println("Server stopped gracefully")
}
```

**Shutdown Steps:**

1. **Stop Accepting New Requests** - Server stops listening
2. **Wait for Active Requests** - Up to 30 seconds for in-flight requests to complete
3. **Timeout Handling** - If timeout expires, force close remaining connections
4. **Database Cleanup** - `defer storage.Close()` runs after shutdown

**Why 30 seconds?**

- Most HTTP requests complete in < 5 seconds
- 30s gives ample time for slow queries or bulk operations
- Long enough for graceful, short enough to not block deployments

**Force Shutdown:**

If graceful shutdown times out:

```go
if err := server.Close(); err != nil {
    log.Printf("Error forcing shutdown: %v", err)
}
```

- `Close()` immediately closes all connections
- Active requests will see "connection reset" errors
- Last resort to ensure server actually stops

### Shutdown in Different Environments

**Development:**
- Ctrl+C triggers SIGINT
- Graceful shutdown completes in < 1 second (no active requests)

**Production (systemd):**
- `systemctl stop quest-todo` sends SIGTERM
- systemd waits for process to exit (configurable timeout)
- If process doesn't exit, systemd sends SIGKILL (force kill)

**Docker:**
- `docker stop` sends SIGTERM, waits 10 seconds (default), then SIGKILL
- Our 30s timeout may be cut short by Docker's 10s timeout
- Can configure: `docker stop -t 60 quest-todo-container`

**Kubernetes:**
- Sends SIGTERM, waits for `terminationGracePeriodSeconds` (default 30s)
- Then sends SIGKILL
- Matches our 30s graceful shutdown timeout

## Logging

### Log Configuration

```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

**Log Flags:**

- `log.LstdFlags` - Date and time (2009/01/23 01:23:23)
- `log.Lshortfile` - Filename and line number (main.go:173)

**Output:**
```
2026/02/20 16:49:10 main.go:173: Starting Quest Todo API Server
```

### Initialization Logging

Server logs all major initialization steps:

```go
log.Println("Starting Quest Todo API Server")
log.Printf("Version: %s", Version)
log.Printf("Initializing SQLite database at %s", config.Database.Path)
log.Println("Database initialized and migrated successfully")
log.Println("Initializing services...")
log.Println("Services initialized successfully")
log.Println("Initializing API handlers...")
log.Println("API handlers initialized successfully")
log.Println("Setting up HTTP router and middleware...")
log.Println("Router configured successfully")
log.Printf("Server listening on http://%s", server.Addr)
```

**Why log initialization?**

1. **Debugging** - Identify where initialization fails
2. **Monitoring** - Verify server started correctly
3. **Performance** - Measure startup time
4. **Auditing** - Track server restarts

### Request Logging

When `ENABLE_LOGGING=true` (default), the Logger middleware logs every request:

```
[request-id] GET /api/tasks 200 23ms
[request-id] POST /api/tasks 201 45ms
[request-id] GET /health 200 1ms
```

Format: `[request-id] METHOD PATH STATUS_CODE DURATION`

**Why structured logging?**

- Request ID enables tracing across logs
- Method + Path identifies the endpoint
- Status code shows success/error
- Duration shows performance

### Error Logging

Different log levels based on severity:

```go
// Fatal errors (prevent startup)
log.Fatalf("Failed to initialize database: %v", err)

// Error logs (runtime errors, continue serving)
log.Printf("Error during shutdown: %v", err)

// Info logs (normal operations)
log.Println("Server stopped gracefully")
```

**log.Fatalf() vs log.Printf():**

- `log.Fatalf()` - Calls `os.Exit(1)` after logging (startup only)
- `log.Printf()` - Logs and continues (runtime errors, shutdown)

### Production Logging Considerations

For production, consider:

1. **Structured Logging** (JSON format):
   ```go
   {"time":"2026-02-20T16:49:10Z","level":"info","msg":"Server starting","version":"0.1.0"}
   ```
   - Use libraries like `zerolog` or `zap`
   - Easier to parse and aggregate

2. **Log Levels** (debug, info, warn, error):
   - Configure verbosity based on environment
   - `INFO` for production, `DEBUG` for development

3. **Log Aggregation** (ELK, Splunk, CloudWatch):
   - Centralized logging for multiple instances
   - Search and analyze across servers

4. **Log Rotation**:
   - Prevent log files from filling disk
   - Use `logrotate` on Linux

## Error Handling

### Initialization Errors (Fatal)

Errors during initialization prevent the server from starting:

```go
storage, err := sqlite.New(config.Database.Path)
if err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}
```

**Why fatal?**

- Server cannot function without database
- Better to fail fast than serve errors
- Kubernetes/systemd will restart the pod/service

**What causes initialization errors?**

- Database file permissions
- Disk full
- Migration failures
- Invalid configuration

### Runtime Errors (Non-Fatal)

Runtime errors are handled by middleware and services:

```go
// Recovery middleware catches panics
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v\n%s", err, debug.Stack())
                // Return 500 error to client
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

**Error Propagation:**

```
Handler Error
    ↓
Service Error
    ↓
Storage Error (SQL error, constraint violation)
    ↓
HTTP Error Response (400/404/500)
```

Errors are mapped to HTTP status codes by the API layer.

### Shutdown Errors (Logged)

Errors during shutdown are logged but don't prevent exit:

```go
if err := storage.Close(); err != nil {
    log.Printf("Error closing database: %v", err)
}
```

**Why not fatal?**

- Server is already shutting down
- Forcing a crash doesn't help
- Log for post-mortem analysis

## Production Considerations

### Security

**CORS:**
```bash
# Development (allow all)
ALLOWED_ORIGINS=*

# Production (specific origins only)
ALLOWED_ORIGINS=https://quest-todo.com,https://app.quest-todo.com
```

**HTTPS:**

Current implementation uses HTTP. For production, add TLS:

```go
// Option 1: Let server handle TLS
server.ListenAndServeTLS("cert.pem", "key.pem")

// Option 2: Use reverse proxy (recommended)
// Nginx/Caddy handles TLS, forwards to HTTP server
```

Reverse proxy is recommended because:
- Centralized certificate management
- Better performance (TLS handshake offloading)
- Additional features (caching, compression, rate limiting)

**Environment Variables:**

Never commit secrets to version control:

```bash
# Wrong: Hard-coded in code
const dbPassword = "secret123"

# Right: Environment variable
DB_PASSWORD="secret123" ./quest-todo-server
```

### Performance

**Connection Pooling:**

SQLite configuration (already done in `sqlite.New()`):
```go
db.SetMaxOpenConns(1)  // SQLite: 1 writer only
db.SetMaxIdleConns(1)  // Keep connection alive
```

For PostgreSQL/MySQL, increase limits:
```go
db.SetMaxOpenConns(25)  // 25 concurrent connections
db.SetMaxIdleConns(5)   // Keep 5 idle connections
```

**Timeouts:**

Already configured (15s read/write, 60s idle).

For high-traffic production, consider:
- Lower timeouts to free up resources faster
- Add request timeout middleware for slow endpoints

**Middleware Order:**

Current order (reverse of application):
```go
CORS → Logger → RequestID → ContentType → Recovery
```

Why this order?
1. **CORS** - First, handle preflight requests early
2. **Logger** - Log all requests (even if they fail later)
3. **RequestID** - Add request ID for logging/tracing
4. **ContentType** - Validate content type
5. **Recovery** - Last, catch panics from handlers

### Monitoring

**Health Checks:**

Already implemented:
```bash
curl http://localhost:8080/health
# {"status":"ok","database":"healthy"}
```

Use for:
- Kubernetes liveness/readiness probes
- Load balancer health checks
- Monitoring systems (Datadog, New Relic)

**Metrics:**

Current implementation logs basic metrics (request duration).

For production, add:
- Prometheus metrics (request count, duration, errors)
- Custom business metrics (tasks created, completion rate)

**Distributed Tracing:**

Request IDs enable basic tracing:
```
[abc-123] Processing request...
[abc-123] Querying database...
[abc-123] Request completed in 45ms
```

For advanced tracing, integrate:
- OpenTelemetry
- Jaeger
- Zipkin

### Deployment

**Docker:**

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -ldflags="-s -w" -o quest-todo-server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/quest-todo-server .
EXPOSE 8080
CMD ["./quest-todo-server"]
```

**Kubernetes:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quest-todo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: quest-todo
  template:
    metadata:
      labels:
        app: quest-todo
    spec:
      containers:
      - name: quest-todo
        image: quest-todo:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: DB_PATH
          value: "/data/quest-todo.db"
        - name: ALLOWED_ORIGINS
          value: "https://quest-todo.com"
        volumeMounts:
        - name: data
          mountPath: /data
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 5
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: quest-todo-data
```

**systemd:**

```ini
[Unit]
Description=Quest Todo API Server
After=network.target

[Service]
Type=simple
User=quest-todo
WorkingDirectory=/opt/quest-todo
ExecStart=/opt/quest-todo/quest-todo-server
Restart=on-failure
RestartSec=5s
Environment="PORT=8080"
Environment="DB_PATH=/var/lib/quest-todo/data.db"

[Install]
WantedBy=multi-user.target
```

## Best Practices

### Do's ✅

1. **Use Environment Variables for Configuration**
   - Easier deployment
   - Container-friendly
   - No config files to manage

2. **Fail Fast on Initialization Errors**
   - Better than serving errors
   - Let orchestrator (K8s, systemd) restart

3. **Handle Shutdown Signals Gracefully**
   - SIGINT (Ctrl+C) and SIGTERM
   - Wait for active requests to complete

4. **Log All Initialization Steps**
   - Easier debugging
   - Monitoring and auditing

5. **Use Explicit Dependency Injection**
   - Testable
   - Clear dependencies
   - No magic

6. **Close Resources with defer**
   - Guaranteed cleanup
   - Even if panic occurs

7. **Use Goroutines for Non-Blocking Operations**
   - Server start doesn't block signal handling
   - Buffered channel prevents goroutine leak

8. **Configure HTTP Timeouts**
   - Prevents resource exhaustion
   - Protects against slow clients

### Don'ts ❌

1. **Don't Use Global Variables for Dependencies**
   ```go
   // ❌ Bad: Global database connection
   var db *sql.DB

   // ✅ Good: Pass as dependency
   storage := sqlite.New(dbPath)
   service := service.NewTaskService(storage, ...)
   ```

2. **Don't Ignore Errors During Initialization**
   ```go
   // ❌ Bad: Ignoring error
   storage, _ := sqlite.New(dbPath)

   // ✅ Good: Check and fail fast
   storage, err := sqlite.New(dbPath)
   if err != nil {
       log.Fatalf("Failed to initialize: %v", err)
   }
   ```

3. **Don't Use time.Sleep() for Synchronization**
   ```go
   // ❌ Bad: Arbitrary wait
   go server.ListenAndServe()
   time.Sleep(100 * time.Millisecond)

   // ✅ Good: Wait for signal or error
   serverErrors := make(chan error, 1)
   go func() { serverErrors <- server.ListenAndServe() }()
   select {
   case err := <-serverErrors:
       // Handle error
   case <-shutdown:
       // Handle shutdown
   }
   ```

4. **Don't Forget to Close Resources**
   ```go
   // ❌ Bad: Resource leak
   storage, _ := sqlite.New(dbPath)
   // storage never closed

   // ✅ Good: Defer cleanup
   storage, err := sqlite.New(dbPath)
   if err != nil { ... }
   defer storage.Close()
   ```

5. **Don't Use Unbuffered Channels for Goroutine Communication**
   ```go
   // ❌ Bad: Unbuffered (goroutine leak if no receiver)
   serverErrors := make(chan error)
   go func() { serverErrors <- server.ListenAndServe() }()

   // ✅ Good: Buffered (no leak)
   serverErrors := make(chan error, 1)
   go func() { serverErrors <- server.ListenAndServe() }()
   ```

6. **Don't Start Server Without Graceful Shutdown**
   ```go
   // ❌ Bad: No graceful shutdown
   log.Fatal(http.ListenAndServe(":8080", handler))

   // ✅ Good: Signal handling + graceful shutdown
   shutdown := make(chan os.Signal, 1)
   signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
   // ... select statement
   server.Shutdown(ctx)
   ```

### Common Pitfalls

**1. Forgetting to Handle SIGTERM**

Only handling SIGINT (Ctrl+C) works in development but fails in production:

```go
// ❌ Bad: Only SIGINT
signal.Notify(shutdown, os.Interrupt)

// ✅ Good: Both SIGINT and SIGTERM
signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
```

**2. Not Setting HTTP Timeouts**

Without timeouts, slow clients can exhaust server resources:

```go
// ❌ Bad: No timeouts (default: none)
server := &http.Server{Addr: ":8080", Handler: router}

// ✅ Good: Reasonable timeouts
server := &http.Server{
    Addr:         ":8080",
    Handler:      router,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

**3. Using log.Fatal() for Runtime Errors**

`log.Fatal()` calls `os.Exit()`, which:
- Doesn't run defer statements
- Doesn't allow graceful shutdown
- Immediately kills the server

```go
// ❌ Bad: Fatal for runtime error
if err := processRequest(); err != nil {
    log.Fatal(err)  // Kills entire server!
}

// ✅ Good: Log and return error response
if err := processRequest(); err != nil {
    log.Printf("Error: %v", err)
    http.Error(w, "Internal error", 500)
    return
}
```

Use `log.Fatal()` only during initialization, not for runtime errors.

## Summary

The main server implementation:

1. **Loads Configuration** - From environment variables with sensible defaults
2. **Initializes Dependencies** - Explicit dependency injection (storage → services → API → router)
3. **Starts HTTP Server** - With proper timeouts and graceful shutdown
4. **Handles Signals** - SIGINT and SIGTERM for clean exits
5. **Logs Operations** - Initialization, requests, errors, shutdown
6. **Fails Fast** - Fatal errors during startup prevent running broken server
7. **Cleans Up** - defer statements ensure resource cleanup

This design follows Go best practices and 12-factor app principles for production-ready applications.

**Related Documentation:**
- [Router and Middleware Implementation](./router-middleware-implementation.md)
- [API Layer Implementation](./api-layer-implementation.md)
- [Service Layer Implementation](./service-layer-implementation.md)
- [Server Usage Documentation](../../backend/cmd/server/README.md)
