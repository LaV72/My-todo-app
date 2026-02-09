# Quest Todo

A to-do application inspired by the quest menu system from Nihon Falcom's Trails series, featuring a journal-style interface with a clean separation between backend and frontend services.

## Documentation

- [Architecture Overview](./docs/architecture.md) - System architecture and design principles
- [API Specification](./docs/api-spec.md) - Complete REST API documentation
- [Data Models](./docs/data-models.md) - Database schema and data structures
- [Storage Layer](./docs/storage-layer.md) - Storage abstraction and implementations
- [Frontend Design](./docs/frontend-design.md) - UI/UX design inspired by Trails series
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
- **Pluggable Storage**: Swap between SQLite, JSON, or custom storage
- **Cross-platform Backend**: Go backend runs anywhere
- **Native Frontend**: Swift/SwiftUI for optimal macOS experience

### Technology Stack

**Backend:**
- Language: Go
- HTTP Router: gorilla/mux (or standard library)
- Storage: SQLite (modernc.org/sqlite), JSON fallback
- Architecture: Clean architecture with service layer

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
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
open QuestTodo.xcodeproj
# Press Cmd+R in Xcode to run
```

## Design Philosophy

1. **Separation of Concerns**: Backend handles data, frontend handles presentation
2. **Pluggable Storage**: Storage layer is abstracted behind interfaces
3. **Clean Architecture**: Business logic separated from infrastructure
4. **JRPG Aesthetics**: UI inspired by Trails series quest journals
5. **Cross-platform Core**: Backend works everywhere, frontend is platform-specific

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

Currently in design phase. Documentation represents the planned architecture and features.

## Contributing

This is a personal project, but feedback and suggestions are welcome through issues.

## License

TBD
