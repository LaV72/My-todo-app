# Architecture Overview

## System Architecture

Quest Todo follows a client-server architecture with clear separation between backend (core) and frontend (plugin).

```
┌──────────────────────────────────────────────────────┐
│                  Frontend Layer                      │
│  ┌────────────────────────────────────────────────┐  │
│  │         Swift/SwiftUI (macOS App)              │  │
│  │  ├─ Views (Journal UI)                         │  │
│  │  ├─ ViewModels (MVVM)                          │  │
│  │  ├─ API Service (HTTP Client)                  │  │
│  │  └─ Models (Swift structs)                     │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────┬───────────────────────────────┘
                       │ REST API (HTTP/JSON)
                       │ http://localhost:3000
┌──────────────────────▼───────────────────────────────┐
│                  Backend Service                      │
│  ┌────────────────────────────────────────────────┐  │
│  │              API Layer                         │  │
│  │  ├─ HTTP Handlers                              │  │
│  │  ├─ Middleware (CORS, Logging, Recovery)      │  │
│  │  └─ Router                                     │  │
│  └────────────────────┬───────────────────────────┘  │
│  ┌────────────────────▼───────────────────────────┐  │
│  │            Service Layer                       │  │
│  │  ├─ TaskService (Business Logic)              │  │
│  │  ├─ CategoryService                           │  │
│  │  ├─ StatsService                              │  │
│  │  └─ Validation & Processing                   │  │
│  └────────────────────┬───────────────────────────┘  │
│  ┌────────────────────▼───────────────────────────┐  │
│  │         Storage Interface                      │  │
│  │  (Abstract contract - pluggable)               │  │
│  └────────────────────┬───────────────────────────┘  │
│           ┌───────────┴───────────┬──────────────┐   │
│  ┌────────▼──────┐  ┌────────────▼──┐  ┌────────▼┐  │
│  │ SQLite        │  │ JSON          │  │ Memory  │  │
│  │ (Production)  │  │ (Simple/Debug)│  │ (Tests) │  │
│  └───────────────┘  └───────────────┘  └─────────┘  │
└──────────────────────────────────────────────────────┘
```

## Design Principles

### 1. Separation of Concerns

**Backend (Core)**
- Data storage and retrieval
- Business logic validation
- API endpoint exposure
- Cross-platform compatibility
- No UI knowledge

**Frontend (Plugin)**
- User interface rendering
- User interaction handling
- API consumption
- Platform-specific optimizations
- No data storage logic

### 2. Pluggable Storage

The storage layer is abstracted behind interfaces, allowing easy swapping of storage implementations without affecting business logic.

**Benefits:**
- Start simple (JSON), scale up (SQLite)
- Easy to add new storage backends
- Test with in-memory storage
- Migrate data without code changes

### 3. Clean Architecture

```
External → API → Service → Storage
         ←     ←         ←
```

**Layers:**
1. **API Layer**: HTTP request/response handling
2. **Service Layer**: Business logic, validation, orchestration
3. **Storage Layer**: Data persistence abstraction
4. **Implementation Layer**: Concrete storage implementations

**Rules:**
- Outer layers depend on inner layers
- Inner layers know nothing about outer layers
- Business logic independent of frameworks
- Storage is just a detail

### 4. Interface-Driven Design

All major components communicate through interfaces:

```go
type Storage interface {
    TaskStorage
    CategoryStorage
    StatsStorage
    Close() error
}
```

**Benefits:**
- Easy mocking for tests
- Loose coupling
- Dependency injection
- Swappable implementations

## Component Details

### Backend Components

#### 1. API Layer (`internal/api`)
**Responsibilities:**
- Parse HTTP requests
- Validate request format
- Call service layer
- Format responses
- Handle HTTP errors
- CORS, logging, recovery middleware

**Does NOT:**
- Contain business logic
- Access storage directly
- Validate business rules

#### 2. Service Layer (`internal/service`)
**Responsibilities:**
- Business logic implementation
- Data validation (business rules)
- Orchestrate storage operations
- Calculate computed fields
- Transaction management

**Example:**
```go
type TaskService struct {
    store storage.Storage
}

func (s *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error) {
    // Validate business rules
    if err := s.validateTask(req); err != nil {
        return nil, err
    }

    // Apply business logic
    task := s.buildTask(req)

    // Delegate to storage
    return s.store.CreateTask(ctx, task)
}
```

#### 3. Storage Interface (`internal/storage`)
**Responsibilities:**
- Define storage contract
- Common models and types
- Storage errors
- Factory for implementations

**Key Interfaces:**
- `Storage` - Main interface
- `TaskStorage` - Task operations
- `CategoryStorage` - Category operations
- `StatsStorage` - Statistics queries

#### 4. Storage Implementations

**SQLite (`internal/storage/sqlite`)**
- Production storage
- B-tree indexes for performance
- ACID compliance
- Full-text search capability
- Single file database

**JSON (`internal/storage/json`)**
- Simple file-based storage
- Human-readable format
- Good for debugging
- Easy version control
- Suitable for <1000 tasks

**Memory (`internal/storage/memory`)**
- In-memory storage
- Fast for testing
- No persistence
- Simple implementation

### Frontend Components (Swift/macOS)

#### 1. Views
**Responsibilities:**
- UI rendering
- User interaction
- Display data from ViewModels
- Journal-style aesthetics

**Structure:**
- `MainView` - Container with tabs
- `TaskListView` - Left panel (task list)
- `TaskDetailView` - Right panel (task details)
- `HistoryView` - Completed tasks
- Components (reusable UI elements)

#### 2. ViewModels (MVVM)
**Responsibilities:**
- Fetch data from API
- Transform data for views
- Handle user actions
- State management
- Error handling

**Example:**
```swift
class TaskListViewModel: ObservableObject {
    @Published var tasks: [Task] = []
    @Published var isLoading = false

    private let apiService: APIService

    func loadTasks() async {
        isLoading = true
        tasks = try await apiService.getTasks()
        isLoading = false
    }
}
```

#### 3. API Service
**Responsibilities:**
- HTTP communication
- Request/response serialization
- Error handling
- Retry logic (optional)

#### 4. Models
**Responsibilities:**
- Swift structs matching API models
- Codable conformance
- Computed properties for UI

## Communication Protocol

### REST API

**Format:** JSON over HTTP
**Base URL:** `http://localhost:3000/api/v1`

**Request Example:**
```http
POST /api/v1/tasks
Content-Type: application/json

{
  "title": "Complete project",
  "priority": 4,
  "deadline": {
    "type": "short",
    "date": "2026-02-12T23:59:59Z"
  }
}
```

**Response Example:**
```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "success": true,
  "data": {
    "id": "task-123",
    "title": "Complete project",
    "createdAt": "2026-02-09T10:00:00Z"
  }
}
```

### Error Handling

**Error Response Format:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Priority must be between 1 and 5",
    "fields": {
      "priority": "Invalid value"
    }
  }
}
```

**Error Codes:**
- `VALIDATION_ERROR` - Invalid input
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Duplicate or constraint violation
- `INTERNAL_ERROR` - Server error

## Data Flow

### Creating a Task

```
User Input
    │
    ▼
[Frontend View]
    │ User fills form
    ▼
[ViewModel]
    │ Calls API service
    ▼
[API Service]
    │ POST /tasks
    ▼
────────────── HTTP ──────────────
    │
    ▼
[API Handler]
    │ Parse request
    ▼
[Service Layer]
    │ Validate business rules
    │ Generate ID, timestamps
    ▼
[Storage Interface]
    │
    ▼
[SQLite/JSON Implementation]
    │ Persist data
    ▼
────────────── HTTP ──────────────
    │
    ▼
[API Service]
    │ Parse response
    ▼
[ViewModel]
    │ Update state
    ▼
[View]
    │ Refresh UI
    ▼
User sees new task
```

## Deployment Architecture

### Development

```
┌─────────────────┐
│   Developer     │
│   Machine       │
│                 │
│  ┌───────────┐  │
│  │ Frontend  │  │
│  │   :3001   │  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼─────┐  │
│  │ Backend   │  │
│  │   :3000   │  │
│  └───────────┘  │
│                 │
│  data/todo.db   │
└─────────────────┘
```

### Production (Single User)

```
┌──────────────────────────┐
│      User's Mac          │
│                          │
│  ┌────────────────────┐  │
│  │  QuestTodo.app     │  │
│  │  (Frontend)        │  │
│  └──────┬─────────────┘  │
│         │ localhost      │
│  ┌──────▼─────────────┐  │
│  │  Backend Service   │  │
│  │  (LaunchAgent)     │  │
│  │  :3000             │  │
│  └────────────────────┘  │
│                          │
│  ~/Library/Application   │
│    Support/QuestTodo/    │
│      └─ data/todo.db     │
└──────────────────────────┘
```

**Backend runs as:**
- LaunchAgent (starts on login)
- Background service
- Localhost only

## Technology Choices

### Backend: Go

**Why Go:**
- Single binary deployment
- Fast compilation and runtime
- Excellent standard library
- Built-in HTTP server
- Cross-platform compilation
- No runtime dependencies
- Small memory footprint

**Not Node.js:**
- Avoid npm dependencies bloat
- No JavaScript runtime needed
- Better performance
- Simpler deployment

### Frontend: Swift/SwiftUI

**Why Swift:**
- Native macOS integration
- Optimal performance
- Access to system APIs
- Modern declarative UI (SwiftUI)
- Type safety
- First-class Apple platform support

### Storage: SQLite + JSON

**SQLite:**
- B-tree indexes for fast queries
- ACID compliance
- Single file
- Embedded (no server)
- Battle-tested
- Cross-platform

**JSON:**
- Simple fallback
- Human-readable
- Easy debugging
- Version control friendly
- Good for small datasets

## Scalability Considerations

### Current Architecture (Single User)
- Backend runs locally
- SQLite handles millions of tasks
- No network latency
- Fast and responsive

### Future Extensions

**Multi-device Sync:**
```
[Mac Frontend] ─┐
                 ├─► [Sync Service] ─► [Cloud Storage]
[iOS Frontend] ─┘
```

**Web Frontend:**
```
[Swift/macOS]  ─┐
                 ├─► [Go Backend] ─► [SQLite]
[React/Web]    ─┘
```

**Cloud Backend:**
```
[Frontend] ─► [API Gateway] ─► [Backend Cluster] ─► [PostgreSQL]
```

The pluggable architecture makes these extensions possible without rewriting core logic.

## Security Considerations

### Current (Local-only)
- Backend binds to localhost only
- No external network access
- Data stored locally
- OS-level security (file permissions)

### Future (If networked)
- API authentication (JWT/OAuth)
- HTTPS/TLS encryption
- Rate limiting
- Input validation
- CORS configuration
- SQL injection prevention (parameterized queries)

## Testing Strategy

### Backend Tests
- **Unit Tests**: Service layer with mocked storage
- **Integration Tests**: API handlers with memory storage
- **Storage Tests**: Each implementation (SQLite, JSON)

### Frontend Tests
- **Unit Tests**: ViewModels with mocked API
- **UI Tests**: SwiftUI view testing
- **Integration Tests**: Full app with test backend

### Example Test Structure
```go
func TestTaskService_CreateTask(t *testing.T) {
    // Use memory storage for testing
    store := memory.New()
    service := NewTaskService(store)

    task, err := service.CreateTask(ctx, &CreateTaskRequest{
        Title: "Test Task",
        Priority: 3,
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, task.ID)
}
```

## Performance Characteristics

### Backend
- **Task Creation**: ~1-5ms (SQLite), ~10-50ms (JSON)
- **Task Retrieval**: ~1-5ms (indexed query)
- **List Tasks (100)**: ~5-10ms (SQLite), ~50-100ms (JSON)
- **Search**: ~10-20ms (SQLite FTS), ~100-500ms (JSON scan)

### Frontend
- **API Call Latency**: <5ms (localhost)
- **UI Rendering**: 60fps with SwiftUI
- **List Scrolling**: Smooth with lazy loading

### Storage Limits
- **SQLite**: Millions of tasks
- **JSON**: Comfortable up to ~10,000 tasks
- **Memory**: Limited by RAM (testing only)

## Configuration

### Backend Config
```json
{
  "server": {
    "host": "localhost",
    "port": 3000
  },
  "storage": {
    "type": "sqlite",
    "path": "./data/todo.db"
  }
}
```

### Frontend Config
```swift
struct AppConfig {
    static let apiBaseURL = "http://localhost:3000/api/v1"
    static let requestTimeout: TimeInterval = 30
}
```

## Monitoring & Logging

### Backend Logging
- Request/response logging (middleware)
- Error logging with stack traces
- Storage operation timing
- Health check endpoint

### Frontend Logging
- API call logging
- Error reporting
- User action tracking (optional)

## Summary

Quest Todo's architecture prioritizes:
1. **Simplicity**: Clear separation of concerns
2. **Flexibility**: Pluggable components
3. **Maintainability**: Clean interfaces and layers
4. **Performance**: Efficient storage with indexes
5. **Testability**: Mockable interfaces
6. **Extensibility**: Easy to add features

The backend is a lean Go service with pluggable storage, while the frontend is a native Swift app that communicates via REST API. This architecture allows independent evolution of both components while maintaining a clean contract between them.
