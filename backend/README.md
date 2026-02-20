# Quest Todo - Backend

Go backend service for Quest Todo application.

## Directory Structure

```
backend/
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── models/           # Data models
│   ├── storage/          # Storage layer
│   │   ├── sqlite/       # SQLite implementation
│   │   ├── json/         # JSON file implementation
│   │   └── memory/       # In-memory (testing)
│   ├── service/          # Business logic layer
│   ├── api/              # HTTP API layer
│   │   ├── handlers/     # Request handlers
│   │   └── middleware/   # HTTP middleware
│   ├── config/           # Configuration management
│   └── utils/            # Utility functions
├── data/                 # Data storage (gitignored)
├── config.json           # Configuration file
└── go.mod                # Go module definition
```

## Getting Started

### Prerequisites

- Go 1.21 or higher

### Install Dependencies

```bash
go mod download
```

### Run Server

```bash
go run cmd/server/main.go
```

### Build

```bash
go build -o bin/quest-todo cmd/server/main.go
```

### Configuration

Edit `config.json` to change server port or storage type:

```json
{
  "server": {
    "port": 3000
  },
  "storage": {
    "type": "sqlite"  // or "json"
  }
}
```

## API Documentation

See [API Specification](../docs/api-spec.md) for complete API documentation.

## Development

### Run Tests

```bash
go test ./...
```

### Run with Hot Reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```
