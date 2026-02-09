# Development Guide

Complete guide for setting up and developing Quest Todo.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Project Setup](#project-setup)
3. [Backend Development](#backend-development)
4. [Frontend Development](#frontend-development)
5. [Development Workflow](#development-workflow)
6. [Testing](#testing)
7. [Building & Deployment](#building--deployment)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Backend (Go)

**Required:**
- Go 1.21 or higher
- Git

**Installation (macOS):**
```bash
# Using Homebrew
brew install go

# Verify installation
go version
```

### Frontend (Swift/macOS)

**Required:**
- Xcode 15 or higher
- macOS 13 (Ventura) or higher
- Swift 5.9 or higher

**Installation:**
- Download Xcode from Mac App Store
- Install Command Line Tools: `xcode-select --install`

### Optional Tools

**Development Tools:**
- [Postman](https://www.postman.com/) or [Insomnia](https://insomnia.rest/) - API testing
- [DB Browser for SQLite](https://sqlitebrowser.org/) - Database inspection
- [TablePlus](https://tableplus.com/) - Database GUI (alternative)

**Code Quality:**
- `golangci-lint` - Go linter
- `swiftlint` - Swift linter

```bash
# Install tools
brew install golangci-lint
brew install swiftlint
```

---

## Project Setup

### Clone Repository

```bash
git clone https://github.com/yourusername/quest-todo.git
cd quest-todo
```

### Project Structure

```
quest-todo/
├── backend/              # Go backend service
│   ├── cmd/
│   ├── internal/
│   ├── data/            # Data storage (gitignored)
│   ├── go.mod
│   └── config.json
├── frontend/            # Swift/macOS frontend
│   ├── QuestTodo.xcodeproj
│   ├── QuestTodo/
│   │   ├── Views/
│   │   ├── ViewModels/
│   │   ├── Models/
│   │   └── Services/
│   └── Assets.xcassets
├── docs/                # Documentation
└── README.md
```

---

## Backend Development

### Initial Setup

```bash
cd backend

# Initialize Go module
go mod init github.com/yourusername/quest-todo

# Install dependencies
go get github.com/gorilla/mux
go get modernc.org/sqlite
go get github.com/google/uuid
```

### Directory Structure

```bash
mkdir -p cmd/server
mkdir -p internal/{api,service,storage,config}
mkdir -p internal/api/{handlers,middleware}
mkdir -p internal/storage/{sqlite,json,memory}
mkdir -p data
```

### Configuration

Create `config.json`:

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

### Run Backend

```bash
# Development mode
go run cmd/server/main.go

# With hot reload (using air)
go install github.com/cosmtrek/air@latest
air

# Build and run
go build -o bin/quest-todo cmd/server/main.go
./bin/quest-todo
```

**Expected Output:**
```
Starting Quest Todo Backend
Storage: SQLite (./data/todo.db)
Server listening on http://localhost:3000
```

### Development Commands

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter
golangci-lint run

# Generate test coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build for production
go build -ldflags="-s -w" -o bin/quest-todo cmd/server/main.go
```

### API Testing

```bash
# Health check
curl http://localhost:3000/health

# Create a task
curl -X POST http://localhost:3000/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Task",
    "description": "This is a test",
    "priority": 4,
    "deadline": {
      "type": "short",
      "date": "2026-02-15T23:59:59Z"
    }
  }'

# List tasks
curl http://localhost:3000/api/v1/tasks

# Get specific task
curl http://localhost:3000/api/v1/tasks/{task-id}
```

### Database Inspection

```bash
# Using SQLite CLI
sqlite3 data/todo.db

# Common queries
sqlite> .tables
sqlite> SELECT * FROM tasks;
sqlite> .schema tasks
sqlite> .quit

# Using DB Browser (GUI)
open data/todo.db
```

---

## Frontend Development

### Initial Setup

```bash
cd frontend

# Open in Xcode
open QuestTodo.xcodeproj
```

### Project Configuration

**In Xcode:**
1. Set deployment target: macOS 13.0+
2. Configure signing team
3. Set bundle identifier: `com.yourname.quest-todo`

### Project Structure

```
QuestTodo/
├── QuestTodoApp.swift       # App entry point
├── Views/
│   ├── MainView.swift       # Main container
│   ├── TaskListView.swift   # Left panel
│   ├── TaskDetailView.swift # Right panel
│   ├── HistoryView.swift    # History tab
│   └── Components/
│       ├── TaskRowView.swift
│       ├── PriorityStars.swift
│       ├── DeadlineLabel.swift
│       └── ObjectiveRow.swift
├── ViewModels/
│   ├── TaskListViewModel.swift
│   └── TaskDetailViewModel.swift
├── Models/
│   ├── Task.swift
│   ├── Category.swift
│   └── Objective.swift
├── Services/
│   ├── APIService.swift
│   └── NetworkClient.swift
└── Resources/
    ├── Assets.xcassets
    └── Colors.xcassets
```

### API Service Setup

Create `APIService.swift`:

```swift
import Foundation

class APIService {
    static let shared = APIService()

    private let baseURL = "http://localhost:3000/api/v1"
    private let session = URLSession.shared

    func getTasks() async throws -> [Task] {
        let url = URL(string: "\(baseURL)/tasks")!
        let (data, _) = try await session.data(from: url)

        let response = try JSONDecoder().decode(TasksResponse.self, from: data)
        return response.data
    }

    func createTask(_ task: TaskCreateRequest) async throws -> Task {
        let url = URL(string: "\(baseURL)/tasks")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(task)

        let (data, _) = try await session.data(for: request)
        let response = try JSONDecoder().decode(TaskResponse.self, from: data)
        return response.data
    }

    // Additional methods...
}
```

### Run Frontend

**In Xcode:**
1. Select target: QuestTodo
2. Select destination: My Mac
3. Press Cmd+R to build and run

**Or from command line:**
```bash
xcodebuild -project QuestTodo.xcodeproj \
  -scheme QuestTodo \
  -configuration Debug \
  build
```

### Preview Mode

Use SwiftUI previews for rapid development:

```swift
struct TaskListView_Previews: PreviewProvider {
    static var previews: some View {
        TaskListView(viewModel: TaskListViewModel.mock)
            .frame(width: 800, height: 600)
    }
}
```

### Development Commands

```bash
# Run SwiftLint
swiftlint

# Auto-fix lint issues
swiftlint --fix

# Build from command line
xcodebuild -project QuestTodo.xcodeproj \
  -scheme QuestTodo \
  build

# Run tests
xcodebuild test \
  -project QuestTodo.xcodeproj \
  -scheme QuestTodo \
  -destination 'platform=macOS'
```

---

## Development Workflow

### 1. Start Backend

```bash
# Terminal 1: Backend
cd backend
go run cmd/server/main.go
```

**Verify backend is running:**
```bash
curl http://localhost:3000/health
```

### 2. Start Frontend

```bash
# Terminal 2: Frontend (or use Xcode)
cd frontend
open QuestTodo.xcodeproj
# Press Cmd+R in Xcode
```

### 3. Typical Development Cycle

**Backend Changes:**
1. Edit Go files
2. Save (auto-restart if using `air`)
3. Test with curl or Postman
4. Verify in frontend

**Frontend Changes:**
1. Edit Swift files
2. Save
3. SwiftUI hot reload shows changes (in previews)
4. Build and run to test full integration

### 4. Feature Development Flow

```bash
# 1. Create feature branch
git checkout -b feature/task-templates

# 2. Backend: Add storage methods
# Edit: internal/storage/storage.go

# 3. Backend: Add service logic
# Edit: internal/service/task_service.go

# 4. Backend: Add API endpoints
# Edit: internal/api/handlers/tasks.go

# 5. Backend: Test
go test ./internal/service/...

# 6. Frontend: Update models
# Edit: Models/Task.swift

# 7. Frontend: Update API service
# Edit: Services/APIService.swift

# 8. Frontend: Update UI
# Edit: Views/TaskListView.swift

# 9. Test integration
# Use app to create/edit tasks

# 10. Commit
git add .
git commit -m "Add task templates feature"
git push origin feature/task-templates
```

---

## Testing

### Backend Tests

**Unit Tests (Service Layer):**

```go
// internal/service/task_service_test.go
package service

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/quest-todo/internal/storage/memory"
)

func TestTaskService_CreateTask(t *testing.T) {
    store := memory.New()
    service := NewTaskService(store)

    req := &CreateTaskRequest{
        Title:    "Test Task",
        Priority: 3,
    }

    task, err := service.CreateTask(context.Background(), req)

    assert.NoError(t, err)
    assert.NotEmpty(t, task.ID)
    assert.Equal(t, "Test Task", task.Title)
}
```

**Integration Tests (API Layer):**

```go
// internal/api/handlers/tasks_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestCreateTaskHandler(t *testing.T) {
    // Setup
    router := setupTestRouter()

    reqBody := map[string]interface{}{
        "title":    "Test Task",
        "priority": 3,
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
    rec := httptest.NewRecorder()

    // Execute
    router.ServeHTTP(rec, req)

    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

**Storage Tests:**

```go
// internal/storage/sqlite/tasks_test.go
package sqlite

import (
    "context"
    "testing"
    "github.com/yourusername/quest-todo/internal/storage"
)

func TestSQLiteStorage_CreateTask(t *testing.T) {
    store, err := New(":memory:")
    require.NoError(t, err)
    defer store.Close()

    task := &storage.Task{
        ID:       "test-1",
        Title:    "Test",
        Priority: 3,
    }

    err = store.CreateTask(context.Background(), task)
    assert.NoError(t, err)

    retrieved, err := store.GetTask(context.Background(), "test-1")
    assert.NoError(t, err)
    assert.Equal(t, task.Title, retrieved.Title)
}
```

**Run Tests:**
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/service/

# Verbose output
go test -v ./...

# With race detection
go test -race ./...
```

### Frontend Tests

**Unit Tests:**

```swift
// Tests/TaskViewModelTests.swift
import XCTest
@testable import QuestTodo

class TaskViewModelTests: XCTestCase {
    func testLoadTasks() async throws {
        let viewModel = TaskListViewModel(apiService: MockAPIService())

        await viewModel.loadTasks()

        XCTAssertEqual(viewModel.tasks.count, 3)
        XCTAssertEqual(viewModel.tasks.first?.title, "Test Task")
    }

    func testFilterTasksByStatus() {
        let viewModel = TaskListViewModel()
        viewModel.tasks = [
            Task(id: "1", title: "Active", status: .active),
            Task(id: "2", title: "Complete", status: .complete),
        ]

        let active = viewModel.activeTasks
        XCTAssertEqual(active.count, 1)
        XCTAssertEqual(active.first?.title, "Active")
    }
}
```

**UI Tests:**

```swift
// UITests/TaskListUITests.swift
import XCTest

class TaskListUITests: XCTestCase {
    func testCreateNewTask() {
        let app = XCUIApplication()
        app.launch()

        app.buttons["New Task"].tap()
        app.textFields["Title"].tap()
        app.textFields["Title"].typeText("UI Test Task")
        app.buttons["Save"].tap()

        XCTAssertTrue(app.staticTexts["UI Test Task"].exists)
    }
}
```

**Run Tests:**
```bash
# In Xcode: Cmd+U

# From command line
xcodebuild test \
  -project QuestTodo.xcodeproj \
  -scheme QuestTodo \
  -destination 'platform=macOS'
```

---

## Building & Deployment

### Backend Build

**Development Build:**
```bash
go build -o bin/quest-todo cmd/server/main.go
```

**Production Build:**
```bash
# Optimized build
go build -ldflags="-s -w" -o bin/quest-todo cmd/server/main.go

# Cross-compile for different platforms
GOOS=darwin GOARCH=amd64 go build -o bin/quest-todo-mac-amd64 cmd/server/main.go
GOOS=darwin GOARCH=arm64 go build -o bin/quest-todo-mac-arm64 cmd/server/main.go
```

**Universal Binary (macOS):**
```bash
# Build for both architectures
GOOS=darwin GOARCH=amd64 go build -o bin/quest-todo-amd64 cmd/server/main.go
GOOS=darwin GOARCH=arm64 go build -o bin/quest-todo-arm64 cmd/server/main.go

# Create universal binary
lipo -create -output bin/quest-todo \
  bin/quest-todo-amd64 \
  bin/quest-todo-arm64
```

### Frontend Build

**Debug Build:**
```bash
xcodebuild -project QuestTodo.xcodeproj \
  -scheme QuestTodo \
  -configuration Debug \
  build
```

**Release Build:**
```bash
xcodebuild -project QuestTodo.xcodeproj \
  -scheme QuestTodo \
  -configuration Release \
  build
```

**Archive for Distribution:**
```bash
xcodebuild -project QuestTodo.xcodeproj \
  -scheme QuestTodo \
  -configuration Release \
  -archivePath build/QuestTodo.xcarchive \
  archive
```

**Export .app:**
```bash
xcodebuild -exportArchive \
  -archivePath build/QuestTodo.xcarchive \
  -exportPath build/ \
  -exportOptionsPlist exportOptions.plist
```

### Installation

**Backend (LaunchAgent):**

1. Install binary:
```bash
sudo cp bin/quest-todo /usr/local/bin/
sudo chmod +x /usr/local/bin/quest-todo
```

2. Create LaunchAgent plist:
```xml
<!-- ~/Library/LaunchAgents/com.quest-todo.backend.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.quest-todo.backend</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/quest-todo</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/quest-todo.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/quest-todo-error.log</string>
</dict>
</plist>
```

3. Load LaunchAgent:
```bash
launchctl load ~/Library/LaunchAgents/com.quest-todo.backend.plist
```

**Frontend:**

1. Copy to Applications:
```bash
cp -r build/QuestTodo.app /Applications/
```

2. Launch:
```bash
open /Applications/QuestTodo.app
```

---

## Troubleshooting

### Backend Issues

**Port already in use:**
```bash
# Find process using port 3000
lsof -i :3000

# Kill process
kill -9 <PID>
```

**Database locked:**
```bash
# Close all connections
# Delete database and restart
rm data/todo.db
go run cmd/server/main.go
```

**Module errors:**
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy dependencies
go mod tidy
```

### Frontend Issues

**Cannot connect to backend:**
- Verify backend is running: `curl http://localhost:3000/health`
- Check firewall settings
- Ensure no VPN/proxy interfering

**Build errors:**
```bash
# Clean build folder
xcodebuild clean -project QuestTodo.xcodeproj

# Or in Xcode: Product > Clean Build Folder (Cmd+Shift+K)

# Reset package cache
rm -rf ~/Library/Developer/Xcode/DerivedData
```

**SwiftUI preview issues:**
```bash
# Restart Xcode
# Or: Editor > Canvas > Restart Canvas (Cmd+Opt+P)
```

### General Issues

**Git issues:**
```bash
# Reset to clean state
git reset --hard HEAD
git clean -fd

# Update from main
git fetch origin
git rebase origin/main
```

**Environment issues:**
```bash
# Verify Go version
go version

# Verify Xcode version
xcodebuild -version

# Check system info
sw_vers
```

---

## Development Tips

### Backend

1. **Use air for hot reload** - Auto-restart on file changes
2. **Log extensively** - Use structured logging
3. **Write tests first** - TDD for reliability
4. **Use contexts** - Pass context.Context everywhere
5. **Handle errors properly** - Wrap errors with context

### Frontend

1. **Use previews** - Develop UI components in isolation
2. **Mock data** - Create mock services for development
3. **Async/await** - Use modern concurrency
4. **MVVM pattern** - Keep views dumb, logic in ViewModels
5. **Extract components** - Reuse UI components

### Git Workflow

```bash
# Start new feature
git checkout -b feature/my-feature

# Commit frequently
git add .
git commit -m "Add feature X"

# Keep branch updated
git fetch origin
git rebase origin/main

# Push changes
git push origin feature/my-feature

# Create PR on GitHub
```

---

## Resources

### Documentation
- [Go Documentation](https://go.dev/doc/)
- [Swift Documentation](https://docs.swift.org)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui)

### Tools
- [Postman](https://www.postman.com/)
- [DB Browser for SQLite](https://sqlitebrowser.org/)
- [Charles Proxy](https://www.charlesproxy.com/) - Network debugging

### Communities
- [Golang Reddit](https://reddit.com/r/golang)
- [Swift Forums](https://forums.swift.org)
- [Stack Overflow](https://stackoverflow.com)

---

## Summary

This guide covers:
- ✅ Complete development environment setup
- ✅ Backend and frontend development workflows
- ✅ Testing strategies for both layers
- ✅ Building and deployment processes
- ✅ Troubleshooting common issues
- ✅ Development best practices

Follow this guide to get started with Quest Todo development. For architecture details, see [architecture.md](./architecture.md). For API details, see [api-spec.md](./api-spec.md).
