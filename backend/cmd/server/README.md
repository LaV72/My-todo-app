# Quest Todo API Server

The main entry point for the Quest Todo backend HTTP server.

## Quick Start

### Build

```bash
# From backend directory
go build -o bin/quest-todo-server ./cmd/server
```

### Run

```bash
# Run with defaults
./bin/quest-todo-server

# Run with custom configuration
PORT=3000 DB_PATH=/data/quest.db ./bin/quest-todo-server
```

The server will:
1. Initialize the SQLite database
2. Run migrations automatically
3. Start the HTTP server on configured port
4. Log all initialization steps

## Configuration

All configuration is done via environment variables with sensible defaults.

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `localhost` | Server host address |
| `PORT` | `8080` | Server port |
| `READ_TIMEOUT` | `15s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `15s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_PATH` | `./quest-todo.db` | Path to SQLite database file |

### Service Layer Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MAX_TITLE_LENGTH` | `200` | Maximum task/category title length |
| `MAX_DESCRIPTION_LENGTH` | `2000` | Maximum description length |
| `MAX_BULK_SIZE` | `100` | Maximum items in bulk operations |
| `REQUIRE_ALL_OBJECTIVES` | `false` | Require all objectives complete for task completion |
| `AUTO_COMPLETE_ON_FULL_PROGRESS` | `true` | Auto-complete tasks at 100% progress |
| `ALLOW_PAST_DEADLINES` | `true` | Allow setting deadlines in the past |
| `ENABLE_CATEGORY_RESTRICTIONS` | `false` | Enable category-based task restrictions |
| `ENABLE_REWARD_SYSTEM` | `false` | Enable reward tracking (future feature) |

### Router/Middleware Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins (comma-separated) |
| `ENABLE_CORS` | `true` | Enable CORS middleware |
| `ENABLE_LOGGING` | `true` | Enable request/response logging |

## Examples

### Development

```bash
# Default development setup
./bin/quest-todo-server
```

### Production

```bash
# Production configuration
HOST=0.0.0.0 \
PORT=8080 \
DB_PATH=/var/lib/quest-todo/data.db \
ALLOWED_ORIGINS=https://quest-todo.com,https://app.quest-todo.com \
ENABLE_LOGGING=true \
MAX_BULK_SIZE=50 \
./bin/quest-todo-server
```

### Testing

```bash
# Test configuration with strict validation
MAX_TITLE_LENGTH=100 \
MAX_DESCRIPTION_LENGTH=500 \
REQUIRE_ALL_OBJECTIVES=true \
AUTO_COMPLETE_ON_FULL_PROGRESS=false \
./bin/quest-todo-server
```

## Graceful Shutdown

The server handles `SIGINT` (Ctrl+C) and `SIGTERM` signals gracefully:

1. Stops accepting new requests
2. Waits for active requests to complete (30s timeout)
3. Closes database connections
4. Exits cleanly

## Logging

The server logs all important events:

```
2026/02/20 16:49:10 main.go:173: Starting Quest Todo API Server
2026/02/20 16:49:10 main.go:174: Version: 0.1.0
2026/02/20 16:49:10 main.go:177: Initializing SQLite database at ./quest-todo.db
2026/02/20 16:49:10 main.go:187: Database initialized and migrated successfully
2026/02/20 16:49:10 main.go:195: Initializing services...
2026/02/20 16:49:10 main.go:200: Services initialized successfully
2026/02/20 16:49:10 main.go:203: Initializing API handlers...
2026/02/20 16:49:10 main.go:212: API handlers initialized successfully
2026/02/20 16:49:10 main.go:215: Setting up HTTP router and middleware...
2026/02/20 16:49:10 main.go:217: Router configured successfully
2026/02/20 16:49:10 main.go:231: Server listening on http://localhost:8080
2026/02/20 16:49:10 main.go:232: Press Ctrl+C to shut down
```

When `ENABLE_LOGGING=true` (default), the server also logs all HTTP requests:

```
[request-id] GET /api/tasks 200 23ms
[request-id] POST /api/tasks 201 45ms
```

## Dependencies

The server automatically wires up:

1. **Storage Layer**: SQLite with WAL mode, foreign keys, proper timeouts
2. **Service Layer**: Task, Objective, Category, and Stats services
3. **API Layer**: REST API handlers with validation
4. **Router**: HTTP routing with middleware chain:
   - Recovery (panic handling)
   - ContentType validation
   - RequestID generation
   - Logger (optional)
   - CORS (optional)

## Endpoints

Once started, the server exposes 26 REST API endpoints:

### Health
- `GET /health` - Health check
- `GET /version` - API version

### Tasks (13 endpoints)
- `GET /api/tasks` - List tasks
- `POST /api/tasks` - Create task
- `GET /api/tasks/:id` - Get task
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task
- `GET /api/tasks?q=query` - Search tasks
- `POST /api/tasks/bulk` - Bulk create
- `PUT /api/tasks/bulk` - Bulk update
- `DELETE /api/tasks/bulk` - Bulk delete
- `POST /api/tasks/:id/complete` - Complete task
- `POST /api/tasks/:id/fail` - Fail task
- `POST /api/tasks/:id/reactivate` - Reactivate task
- `POST /api/tasks/reorder` - Reorder tasks

### Objectives (4 endpoints)
- `POST /api/tasks/:id/objectives` - Create objective
- `PUT /api/objectives/:id` - Update objective
- `DELETE /api/objectives/:id` - Delete objective
- `POST /api/objectives/:id/toggle` - Toggle completion

### Categories (5 endpoints)
- `GET /api/categories` - List categories
- `POST /api/categories` - Create category
- `GET /api/categories/:id` - Get category
- `PUT /api/categories/:id` - Update category
- `DELETE /api/categories/:id` - Delete category

### Stats (2 endpoints)
- `GET /api/stats` - Get overall stats
- `GET /api/stats/categories` - Get category stats

## Architecture

The server follows a layered architecture with dependency injection:

```
main.go
├── Config (environment variables)
├── Storage (SQLite)
│   └── Migrations (automatic)
├── Services
│   ├── TaskService
│   ├── ObjectiveService
│   ├── CategoryService
│   └── StatsService
├── API Handlers
│   └── Validator
└── Router + Middleware
    ├── Recovery
    ├── ContentType
    ├── RequestID
    ├── Logger
    └── CORS
```

## Error Handling

The server handles errors at multiple levels:

1. **Database errors**: Caught during initialization, cause fatal exit
2. **Migration errors**: Caught during startup, cause fatal exit
3. **HTTP handler errors**: Logged and returned as JSON error responses
4. **Panics**: Caught by recovery middleware, logged with stack trace
5. **Shutdown errors**: Logged but don't prevent graceful exit

## Building for Production

### Optimized Build

```bash
# Build with optimizations
go build -ldflags="-s -w" -o bin/quest-todo-server ./cmd/server
```

Flags:
- `-s`: Strip symbol table
- `-w`: Strip DWARF debugging info

This reduces binary size by ~30%.

### Static Binary (for Docker/Alpine)

```bash
# Build fully static binary
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/quest-todo-server ./cmd/server
```

Note: SQLite requires CGO. For a static binary with SQLite, use `modernc.org/sqlite` which is a pure Go implementation (already in use).

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 8080
lsof -i :8080

# Use a different port
PORT=9000 ./bin/quest-todo-server
```

### Database Locked

SQLite uses WAL mode which should prevent most locking issues. If you still see "database locked" errors:

1. Check that only one server instance is running
2. Ensure no other processes have the database open
3. Check file permissions on the database file

### Migration Failures

Migrations run automatically on startup. If migrations fail:

1. Check database file permissions
2. Check disk space
3. Review migration error in logs
4. Delete database file to start fresh (dev only!)

### Performance Issues

If the server is slow:

1. Check `ENABLE_LOGGING=false` to disable request logging
2. Verify SQLite database is on fast storage (SSD)
3. Review service layer timeouts and limits
4. Check for slow queries in application logs
