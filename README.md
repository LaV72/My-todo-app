# Quest Todo

A to-do application inspired by the quest menu system from Nihon Falcom's Trails series, featuring a journal-style interface with a clean separation between backend and frontend services.

## Documentation

### Design Documentation
- [Architecture Overview](./docs/architecture.md) - System architecture and design principles
- [API Specification](./docs/api-spec.md) - Complete REST API documentation
- [Data Models](./docs/data-models.md) - Database schema and data structures
- [Storage Layer](./docs/storage-layer.md) - Pluggable storage abstraction (SQLite, JSON, Memory)
- [Communication Layer](./docs/communication-layer.md) - Pluggable protocols (REST, gRPC, WebSocket, Mock)
- [Frontend Design](./docs/frontend-design.md) - UI/UX design inspired by Trails series

### Implementation Documentation
- [Go Project Structure](./docs/implementation/go-project-structure.md) - Project layout, Go modules, and development setup
- [Data Models Implementation](./docs/implementation/data-models-implementation.md) - Struct design, tags, types, and best practices
- [SQLite Implementation](./docs/implementation/sqlite-implementation.md) - Database schema, indexing, and performance
- [Service Layer Implementation](./docs/implementation/service-layer-implementation.md) - Business logic, dependency injection, and validation
- [API Layer Implementation](./docs/implementation/api-layer-implementation.md) - REST handlers, error mapping, and response formatting
- [Router and Middleware Implementation](./docs/implementation/router-middleware-implementation.md) - HTTP routing and cross-cutting concerns

### Testing Documentation
- [Testing Philosophy](./docs/testing/README.md) - Overall testing approach and best practices
- [Storage Layer Tests](./docs/testing/storage-layer-tests.md) - Integration tests with real database
- [Service Layer Tests](./docs/testing/service-layer-tests.md) - Unit tests with mocked dependencies
- [Router and Middleware Tests](./docs/testing/router-middleware-tests.md) - HTTP routing and middleware tests

### Development
- [Development Guide](./docs/development-guide.md) - Setup and development workflow

## Quick Overview

### Architecture

```
┌─────────────────────────────────────────┐
│         Frontend (Swift/macOS)          │
│   Journal-style UI inspired by          │
│   Trails series quest menus             │
└─────────────────┬───────────────────────┘
                  │ REST API
┌─────────────────▼───────────────────────┐
│         Backend (Go)                    │
│   ├─ API Layer                          │
│   ├─ Service Layer (Business Logic)    │
│   └─ Storage Layer (Pluggable)         │
│       ├─ SQLite (with indexes)         │
│       ├─ JSON (simple/backup)          │
│       └─ Memory (testing)              │
└─────────────────────────────────────────┘
```

### Key Features

- **Quest-style Tasks**: Main tasks, side tasks, with priority ratings (★★★★★)
- **Deadline System**: Short/Medium/Long deadlines with visual indicators
- **Objectives**: Sub-tasks with progress tracking
- **Journal UI**: Parchment-style interface with ornate borders
- **Pluggable Storage**: Swap between SQLite, JSON, or Memory storage via configuration
- **Pluggable Communication**: Swap between REST, gRPC, WebSocket, or Mock via configuration
- **Cross-platform Backend**: Go backend runs anywhere
- **Native Frontend**: Swift/SwiftUI for optimal macOS experience

### Technology Stack

**Backend:**
- Language: Go 1.25
- HTTP Router: Standard library (net/http ServeMux)
- Storage: SQLite (modernc.org/sqlite) - pure Go implementation
- Validation: go-playground/validator
- Testing: testify/assert + httptest
- Architecture: Clean layered architecture (Storage → Service → API → Router)

**Frontend:**
- Platform: macOS
- Language: Swift
- UI Framework: SwiftUI
- Architecture: MVVM pattern

## Getting Started

See [Development Guide](./docs/development-guide.md) for complete setup instructions.

### Quick Start

**Backend:**
```bash
cd backend

# Using Makefile (recommended)
make run          # Build and run
make test         # Run tests
make coverage     # Run tests with coverage

# Or using Go directly
go run cmd/server/main.go

# With custom configuration
PORT=3000 DB_PATH=/data/quest.db ./bin/quest-todo-server
```

The server will start on `http://localhost:8080` by default.

See [Server Documentation](./backend/cmd/server/README.md) for configuration options.

**Frontend:**
```bash
cd frontend
open QuestTodo.xcodeproj
# Press Cmd+R in Xcode to run
```

*Note: Frontend is not yet implemented. Backend is fully functional.*

## Design Philosophy

1. **Separation of Concerns**: Backend handles data, frontend handles presentation
2. **Pluggable Architecture**: Both storage and communication layers are swappable
   - Storage: SQLite, JSON, Memory
   - Communication: REST, gRPC, WebSocket, Mock
3. **Clean Architecture**: Business logic separated from infrastructure
4. **Interface-Driven Design**: Depend on abstractions, not implementations
5. **JRPG Aesthetics**: UI inspired by Trails series quest journals
6. **Experimentation Friendly**: Easy to swap implementations and compare approaches

## Project Structure

```
quest-todo/
├── backend/              # Go backend service
│   ├── cmd/             # Application entry points
│   ├── internal/        # Internal packages
│   └── data/            # Data storage
├── frontend/            # Swift/macOS frontend
│   └── QuestTodo/       # SwiftUI app
├── docs/                # Documentation
└── README.md            # This file
```

## Project Status

### Backend: ✅ Complete

The Go backend is fully implemented and tested:

- ✅ **Storage Layer** - SQLite with automatic migrations, indexing, WAL mode
- ✅ **Service Layer** - Complete business logic with validation and configurable rules
- ✅ **API Layer** - 26 REST endpoints with request validation and error handling
- ✅ **Router & Middleware** - HTTP routing with logging, CORS, recovery, request ID
- ✅ **Server** - Production-ready with graceful shutdown and configuration via env vars
- ✅ **Tests** - 127+ tests across all layers with excellent coverage (< 3s execution)
- ✅ **Documentation** - Comprehensive docs for implementation and testing

**Test Coverage:**
- Storage: 100% interface coverage (real database tests)
- Service: 100% business logic coverage (50+ unit tests)
- Router/Middleware: ~95% coverage (77 tests)
- Total: 8 test files, 127+ tests, < 3 seconds execution time

**To run the backend:**
```bash
cd backend
make run    # or: go run cmd/server/main.go
```

Server will start on `http://localhost:8080` with SQLite database.

### Frontend: 🔲 Planned

Swift/macOS frontend not yet implemented.

## Contributing

This is a personal project, but feedback and suggestions are welcome through issues.

## License

TBD
