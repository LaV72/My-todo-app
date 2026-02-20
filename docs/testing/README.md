# Testing Documentation

## Overview

This directory contains comprehensive documentation about our testing philosophy, strategies, and implementation across all layers of the Quest Todo application.

## Testing Philosophy

### Core Principles

Our testing approach is built on these fundamental principles that apply to **all layers**:

#### 1. Real Over Mocked

**Principle**: Prefer testing against real implementations over mocks when performance allows.

**Rationale**:
- Real tests catch bugs that mocks miss (SQL errors, constraint violations, serialization issues)
- No mock maintenance burden (mocks can drift from real behavior)
- Higher confidence that code actually works
- Simpler test code (no mock setup complexity)

**When to Use Real**:
- Storage layer: Real databases (SQLite in-memory is fast enough)
- HTTP handlers: Real `httptest` servers (no network, just in-process)
- Business logic: Real service instances with configurable dependencies

**When to Use Mocks**:
- External APIs (payment gateways, email services, third-party APIs)
- Slow operations that can't be sped up (actual network calls)
- Testing rare failure scenarios (network timeouts, service outages)
- Dependencies outside our control that charge per API call

#### 2. Layer-Appropriate Testing

**Principle**: Each layer gets the type of tests appropriate to its responsibilities.

**Test Types by Layer**:

| Layer | Primary Test Type | Secondary Test Type |
|-------|------------------|---------------------|
| Storage | Integration tests (real DB) | Unit tests (edge cases) |
| Service | Unit tests (business logic) | Integration tests (with storage) |
| API/HTTP | Handler tests (httptest) | Unit tests (validation) |
| Full Stack | End-to-end integration | - |

**Rationale**: Different layers have different concerns:
- Storage: Must verify SQL correctness and database behavior
- Service: Must verify business logic and calculations
- API: Must verify HTTP protocol and JSON serialization
- Integration: Must verify full workflows

#### 3. Test Independence

**Principle**: Each test should run in complete isolation.

**Requirements**:
- ✅ Create fresh instances for each test
- ✅ Clean up all resources after tests
- ✅ No shared mutable state between tests
- ✅ Tests can run in any order
- ✅ Tests can run in parallel (when safe)

**Implementation**:
```go
func TestSomething(t *testing.T) {
    // Create fresh instance
    storage := createFreshStorage(t)

    // Cleanup guaranteed even if test fails
    t.Cleanup(func() {
        storage.Close()
    })

    // Test code...
}
```

#### 4. Reusable Test Suites

**Principle**: Test interface implementations consistently using shared test suites.

**Pattern**:
```go
// Define reusable test suite for interface
type StorageTestSuite struct {
    Factory func(t *testing.T) Storage
    Cleanup func(t *testing.T, s Storage)
}

// Use it for SQLite
suite := &StorageTestSuite{
    Factory: func(t *testing.T) Storage {
        return sqlite.New(":memory:")
    },
}

// Use it for JSON
suite := &StorageTestSuite{
    Factory: func(t *testing.T) Storage {
        return json.New(t.TempDir())
    },
}
```

**Benefits**:
- All implementations tested consistently
- Single source of truth for interface behavior
- Easy to add new implementations
- High confidence in interface compliance

#### 5. Fast Feedback Loop

**Principle**: Tests should complete quickly enough for continuous integration.

**Targets**:
- Unit tests: < 1 second
- Integration tests: < 5 seconds
- Full test suite: < 10 seconds
- Use build tags for slow tests (performance, load)

**Strategies**:
- Use in-memory databases (SQLite `:memory:`)
- Use `httptest` for HTTP (no network)
- Minimize test data (5-10 records, not thousands)
- Run tests in parallel where safe
- Cache expensive setup operations

#### 6. Comprehensive Coverage

**Principle**: Test all critical paths, not just happy paths.

**Test Categories**:
1. **Happy Paths**: Normal, successful operations
2. **Error Paths**: Validation failures, constraint violations
3. **Edge Cases**: Boundary conditions, empty inputs, nil values
4. **Concurrent Operations**: Race conditions, transaction conflicts
5. **Resource Limits**: Full disks, connection limits, timeouts

**Coverage Goal**: Aim for 80%+ code coverage, but prioritize critical paths over 100% coverage.

## Testing Strategy by Layer

### Storage Layer

**Test Type**: Integration tests with real databases

**Why**: Storage layer must verify SQL correctness, constraints, and transactions. Mocking would miss these critical behaviors.

**Implementation**:
- Real SQLite databases (in-memory for speed)
- Reusable test suite for all storage backends
- Tests verify actual data persistence

**Details**: See [Storage Layer Tests](./storage-layer-tests.md)

### Service Layer

**Test Type**: Unit tests (with mocked storage)

**Why**: Business logic should be tested in isolation for fast feedback and focused testing.

**Unit Tests** (with mocked storage):
- Test business logic independently
- Test validation rules
- Test calculations (progress, statistics)
- Fast execution for quick feedback

**Details**: See [Service Layer Tests](./service-layer-tests.md)

### Router and Middleware Layer

**Test Type**: Unit tests (middleware) + Integration tests (routing)

**Why**: HTTP routing and cross-cutting concerns must be verified independently and together.

**Middleware Tests** (isolated unit tests):
- Test each middleware function independently
- Test RequestID, Logger, Recovery, CORS, ContentType
- Fast execution, no router overhead

**Router Tests** (integration with mocks):
- Test route registration and path matching
- Test HTTP method routing
- Test middleware integration
- Test all 26 API endpoints

**Details**: See [Router and Middleware Tests](./router-middleware-tests.md)

### API/HTTP Layer (Future)

**Test Type**: Handler tests with httptest

**Why**: HTTP handlers must verify request parsing, service integration, and response formatting.

**Implementation**:
- Use `httptest` for in-process HTTP testing
- Test request/response validation
- Test handler logic and error cases
- Test JSON parsing and serialization

**Details**: See [API Layer Tests](./api-layer-tests.md) _(coming soon)_

### Integration Tests (Future)

**Test Type**: End-to-end tests

**Why**: Verify complete workflows across all layers.

**Implementation**:
- Full application stack (API → Service → Storage)
- Test complete user scenarios
- Verify cross-layer error propagation
- Test API contract compliance

**Details**: See [Integration Tests](./integration-tests.md) _(coming soon)_

## Best Practices

### Writing Good Tests

#### 1. Test One Thing

Each test should verify one specific behavior:

```go
// ✅ Good: Tests one specific behavior
func TestTaskCreate_WithValidData_ReturnsNoError(t *testing.T) { ... }
func TestTaskCreate_WithDuplicateID_ReturnsError(t *testing.T) { ... }

// ❌ Bad: Tests multiple unrelated things
func TestTaskEverything(t *testing.T) { ... }
```

#### 2. Use Descriptive Names

Test names should describe what they test and expected outcome:

```go
// ✅ Good: Clear what's being tested
func TestTaskGet_WithInvalidID_ReturnsNotFoundError(t *testing.T)
func TestTaskUpdate_ChangesPersistedData(t *testing.T)
func TestTaskList_WithPagination_ReturnsCorrectPage(t *testing.T)

// ❌ Bad: Unclear what's being tested
func TestTask1(t *testing.T)
func TestError(t *testing.T)
func TestIt(t *testing.T)
```

#### 3. Arrange-Act-Assert Pattern

Structure tests with clear sections:

```go
func TestTaskUpdate(t *testing.T) {
    // Arrange: Set up test data and dependencies
    storage := createTestStorage(t)
    task := &models.Task{ID: "task-1", Title: "Original"}
    storage.CreateTask(ctx, task)

    // Act: Perform the operation being tested
    task.Title = "Updated"
    err := storage.UpdateTask(ctx, task)

    // Assert: Verify expected results
    require.NoError(t, err)
    retrieved, _ := storage.GetTask(ctx, "task-1")
    assert.Equal(t, "Updated", retrieved.Title)
}
```

#### 4. Use Table-Driven Tests

For testing multiple scenarios with similar logic:

```go
func TestTaskValidation(t *testing.T) {
    tests := []struct {
        name    string
        task    *models.Task
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid task",
            task:    &models.Task{Title: "Test", Priority: 3},
            wantErr: false,
        },
        {
            name:    "empty title",
            task:    &models.Task{Title: "", Priority: 3},
            wantErr: true,
            errMsg:  "title is required",
        },
        {
            name:    "invalid priority",
            task:    &models.Task{Title: "Test", Priority: 0},
            wantErr: true,
            errMsg:  "priority must be between 1 and 5",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateTask(tt.task)
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### 5. Clean Up Resources

Always clean up test resources, even if test fails:

```go
func TestDatabase(t *testing.T) {
    // Using t.Cleanup (preferred - runs even if test fails)
    db := openDatabase(t)
    t.Cleanup(func() {
        db.Close()
    })

    // Or using defer
    db := openDatabase(t)
    defer db.Close()

    // Test code...
}
```

#### 6. Test Error Cases

Don't just test happy paths:

```go
func TestTaskOperations(t *testing.T) {
    t.Run("successful creation", func(t *testing.T) {
        // Test normal case
    })

    t.Run("creation with duplicate ID", func(t *testing.T) {
        // Test error case
    })

    t.Run("creation with invalid data", func(t *testing.T) {
        // Test validation
    })

    t.Run("creation with closed database", func(t *testing.T) {
        // Test infrastructure failure
    })
}
```

#### 7. Use Assertion Libraries

Use `testify/assert` or `testify/require` for better error messages:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTask(t *testing.T) {
    // ✅ Good: Clear error messages, test continues
    assert.Equal(t, "expected", actual, "task title should match")
    assert.True(t, condition, "task should be active")

    // ✅ Good: Stops test immediately if fails (use for critical checks)
    require.NoError(t, err, "must create task successfully")
    require.NotNil(t, task, "task must not be nil")

    // ❌ Bad: Poor error messages
    if actual != "expected" {
        t.Error("not equal")
    }
}
```

### Common Pitfalls to Avoid

❌ **Don't use time.Sleep() for synchronization**
```go
// ❌ Bad: Flaky, slow
go doSomething()
time.Sleep(100 * time.Millisecond)
checkResult()

// ✅ Good: Explicit synchronization
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    doSomething()
}()
wg.Wait()
checkResult()
```

❌ **Don't depend on test execution order**
```go
// ❌ Bad: Test2 depends on Test1
func TestCreateUser(t *testing.T) {
    createUser("user-1")  // Created in global state
}

func TestGetUser(t *testing.T) {
    user := getUser("user-1")  // Expects user from Test1
}

// ✅ Good: Each test is independent
func TestGetUser(t *testing.T) {
    createUser("user-1")  // Create what we need
    user := getUser("user-1")
}
```

❌ **Don't share mutable state between tests**
```go
// ❌ Bad: Shared global state
var testDB *sql.DB

func TestA(t *testing.T) {
    testDB.Exec("INSERT...")  // Affects other tests
}

// ✅ Good: Fresh instance per test
func TestA(t *testing.T) {
    db := createTestDB(t)
    db.Exec("INSERT...")
}
```

❌ **Don't test implementation details**
```go
// ❌ Bad: Tests internal structure
func TestTaskStorage(t *testing.T) {
    storage := &SQLiteStorage{}
    assert.Equal(t, 1, storage.connectionPool.maxConns)
}

// ✅ Good: Tests behavior
func TestTaskStorage(t *testing.T) {
    storage := NewSQLiteStorage()
    task := &Task{ID: "1"}
    storage.CreateTask(ctx, task)
    retrieved, err := storage.GetTask(ctx, "1")
    assert.NoError(t, err)
    assert.Equal(t, task.ID, retrieved.ID)
}
```

❌ **Don't ignore test failures**
```go
// ❌ Bad: Commenting out failing tests
// func TestBrokenThing(t *testing.T) { ... }

// ❌ Bad: Skipping tests without fixing
func TestBrokenThing(t *testing.T) {
    t.Skip("TODO: Fix this later")
}

// ✅ Good: Fix the test or the code immediately
```

## Test Commands Reference

### Running Tests

```bash
cd backend

# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run tests in parallel
go test -parallel 4 ./...

# Run tests for specific package
go test -v ./internal/storage/...

# Run specific test function
go test -v -run TestTaskCRUD ./internal/storage/...

# Run tests matching pattern
go test -v -run "TestTask.*" ./internal/storage/...
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Show coverage percentage by package
go test -cover ./...

# Enforce minimum coverage threshold
go test -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | grep total | \
  awk '{if ($3+0 < 80) {print "Coverage below 80%"; exit 1}}'
```

### Test Performance

```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory statistics
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkTaskCreate ./...

# Run benchmarks multiple times for accuracy
go test -bench=. -benchtime=10s -count=5 ./...
```

### Test Debugging

```bash
# Run with race detector (catches concurrency bugs)
go test -race ./...

# Run with timeout
go test -timeout 30s ./...

# Run tests with verbose output
go test -v ./...

# Show test failures only
go test -short ./...

# Continue after first failure (don't stop)
go test -failfast=false ./...

# Run tests with CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

### CI/CD Commands

```bash
# Complete CI test suite
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report for CI
go test -coverprofile=coverage.out -covermode=atomic ./...

# Run tests with JSON output (for CI parsing)
go test -json ./... > test-results.json
```

## Test Organization

### Directory Structure

```
backend/
├── internal/
│   ├── storage/
│   │   ├── storage.go           # Interface
│   │   ├── errors.go            # Error definitions
│   │   ├── testing.go           # Reusable test suite (non-test file)
│   │   └── sqlite/
│   │       ├── sqlite.go        # Implementation
│   │       ├── sqlite_test.go   # Tests
│   │       └── ...
│   ├── service/                 # (Future)
│   │   ├── service.go
│   │   ├── testing.go           # Test helpers
│   │   ├── task_service.go
│   │   └── task_service_test.go
│   └── api/                     # (Future)
│       ├── handlers.go
│       ├── handlers_test.go
│       └── ...
└── cmd/
    └── server/
        └── main_integration_test.go
```

### File Naming

- `*_test.go` - Test files (Go convention)
- `testing.go` - Reusable test helpers (non-test file, accessible to other packages)
- `*_integration_test.go` - Integration tests (for clarity)
- `*_bench_test.go` - Benchmark tests (optional, for organization)

### Test Function Naming

```go
// Standard test
func TestFunctionName(t *testing.T)

// Test with subtests
func TestTaskOperations(t *testing.T) {
    t.Run("Create", func(t *testing.T) { ... })
    t.Run("Get", func(t *testing.T) { ... })
}

// Table-driven test with descriptive name
func TestTaskValidation_Various_Scenarios(t *testing.T)

// Benchmark
func BenchmarkTaskCreate(b *testing.B)

// Example (appears in godoc)
func ExampleTaskService_CreateTask()
```

## Test Metrics

### Current Status

| Layer | Status | Test Files | Execution Time | Coverage |
|-------|--------|-----------|----------------|----------|
| Storage | ✅ Complete | 1 | 0.867s | 100% interface |
| Service | ✅ Complete | 5 (50+ tests) | 0.796s | 100% business logic |
| Router/Middleware | ✅ Complete | 2 (77 tests) | 1.113s | ~95% routing/middleware |
| API Handlers | 🔲 Planned | - | - | - |
| Integration | 🔲 Planned | - | - | - |

**Total**: 8 test files, 127+ tests, < 3 seconds execution time

### Goals

- **Coverage**: 80%+ code coverage for all packages
- **Speed**: < 10 seconds for complete test suite
- **Reliability**: 0% flaky tests
- **Maintainability**: Tests should be easy to understand and update

## Documentation by Layer

### Implemented

- ✅ [Storage Layer Tests](./storage-layer-tests.md) - Complete analysis of storage testing implementation
- ✅ [Service Layer Tests](./service-layer-tests.md) - Complete unit testing of business logic with mocked dependencies
- ✅ [Router and Middleware Tests](./router-middleware-tests.md) - Complete unit and integration testing of HTTP routing and middleware

### Planned

- 🔲 [API Handler Tests](./api-handler-tests.md) - HTTP handler testing (request parsing, validation, responses)
- 🔲 [Integration Tests](./integration-tests.md) - End-to-end testing

## Related Documentation

- [Go Project Structure](../implementation/go-project-structure.md)
- [Data Models Implementation](../implementation/data-models-implementation.md)
- [SQLite Implementation](../implementation/sqlite-implementation.md)

## Summary

Our testing philosophy prioritizes:

1. **Real Over Mocked** - Test against real implementations when performance allows
2. **Layer-Appropriate** - Each layer gets the right type of tests
3. **Independence** - Tests run in isolation and can run in any order
4. **Reusability** - Shared test suites for interface implementations
5. **Fast Feedback** - Quick execution for continuous integration
6. **Comprehensive** - Cover happy paths, errors, and edge cases

This approach provides high confidence that our code works correctly while maintaining fast feedback loops for developers.
