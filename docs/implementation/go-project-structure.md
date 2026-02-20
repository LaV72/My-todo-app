# Go Project Structure

Comprehensive documentation for the Quest Todo backend project structure and setup.

## Table of Contents

1. [Why Go?](#why-go)
2. [Directory Structure](#directory-structure)
3. [Go Module System](#go-module-system)
4. [Configuration Setup](#configuration-setup)
5. [Development Tools](#development-tools)

---

## Why Go?

### Language Selection Rationale

**Go vs Node.js:**

| Feature | Go | Node.js |
|---------|-----|---------|
| Binary Size | Single binary (~10MB) | node + node_modules (~100MB+) |
| Startup Time | Instant | Slower (interpret JS) |
| Memory Usage | Low (~20MB) | Higher (~50MB+) |
| Dependencies | Compiled in | Must install node_modules |
| Performance | Native code | V8 JIT compilation |
| Type Safety | Compile-time types | Runtime (or TypeScript) |
| Concurrency | Goroutines (built-in) | Callbacks/Promises/async-await |
| Cross-compile | Easy (`GOOS=linux go build`) | More complex |

**Go vs Python:**

| Feature | Go | Python |
|---------|-----|--------|
| Speed | Compiled (fast) | Interpreted (slower) |
| Deployment | Single binary | Python + dependencies |
| Type Safety | Static typing | Dynamic (or type hints) |
| Concurrency | Goroutines | Threading (GIL limitations) |
| Learning Curve | Simple, small language | Large ecosystem, many ways |

### Why Go is Perfect for Quest Todo

✅ **Single binary deployment**
```bash
# Build once, run anywhere (no runtime needed)
go build -o quest-todo cmd/server/main.go
./quest-todo  # Just run it - no "npm install", no Python interpreter
```

✅ **Fast startup**
```bash
# Go
./quest-todo  # Starts in <10ms

# Node.js
node server.js  # Starts in ~100ms (load modules, parse JS)

# Python
python server.py  # Starts in ~200ms (import packages, interpret)
```

✅ **Low memory footprint**
- Go: ~20MB for idle server
- Perfect for always-running background service

✅ **Easy cross-compilation**
```bash
# Build for macOS (from any OS)
GOOS=darwin GOARCH=amd64 go build -o quest-todo-mac

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o quest-todo-linux

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o quest-todo.exe
```

✅ **Simple, readable code**
- Small language (25 keywords vs 50+ in other languages)
- One obvious way to do things
- Easy to maintain

---

## Directory Structure

### Standard Go Project Layout

```
backend/
├── cmd/                    # Command-line applications
│   └── server/
│       └── main.go        # Server entry point
├── internal/              # Private application code
│   ├── models/           # Data structures
│   ├── storage/          # Data persistence layer
│   ├── service/          # Business logic
│   ├── api/              # HTTP handlers
│   ├── config/           # Configuration
│   └── utils/            # Utilities
├── data/                 # Runtime data storage
├── go.mod                # Module definition
├── go.sum                # Dependency checksums
├── config.json           # Configuration file
└── README.md             # Documentation
```

### Why This Structure?

#### 1. `cmd/` Directory

**Purpose:** Contains main applications (entry points with `package main`)

```go
// cmd/server/main.go
package main

func main() {
    // Start server
}
```

**Why separate directory?**
- Can have multiple executables:
  ```
  cmd/
  ├── server/main.go      # Main server
  ├── cli/main.go         # Command-line tool
  └── migrator/main.go    # Database migration tool
  ```
- Each `cmd/` subdirectory produces one binary
- Clear separation: entry points vs library code

**Build process:**
```bash
# Build server
go build -o bin/quest-todo cmd/server/main.go

# Build CLI tool
go build -o bin/quest-cli cmd/cli/main.go
```

#### 2. `internal/` Directory

**Purpose:** Private code that can't be imported by other projects

**Go's special rule:**
```go
// Other projects CANNOT import from internal/
import "github.com/LaV72/quest-todo/internal/models"  // ❌ ERROR if external

// Only code within quest-todo can import internal/
import "github.com/LaV72/quest-todo/internal/models"  // ✅ OK within project
```

**Why `internal/`?**
- **Encapsulation:** Hide implementation details
- **API stability:** Free to change internal code without breaking others
- **Clear boundaries:** Public API (if any) vs private implementation

**Without `internal/`:**
```
backend/
├── models/        # Anyone can import this
├── storage/       # Anyone can import this
└── service/       # Anyone can import this

Problem: External projects might depend on your internal APIs
```

**With `internal/`:**
```
backend/
└── internal/
    ├── models/    # Only quest-todo can import
    ├── storage/   # Only quest-todo can import
    └── service/   # Only quest-todo can import

Benefit: Free to refactor without external dependencies
```

#### 3. Package Organization Within `internal/`

```
internal/
├── models/       # Data structures (what)
├── storage/      # Data persistence (how data is stored)
├── service/      # Business logic (how data is processed)
├── api/          # HTTP interface (how users interact)
├── config/       # Configuration (settings)
└── utils/        # Helpers (shared utilities)
```

**Layered architecture:**
```
HTTP Request
    ↓
api/ (handlers)          # Handles HTTP
    ↓
service/ (business logic) # Applies rules
    ↓
storage/ (persistence)    # Saves data
    ↓
Database
```

**Dependency flow (top to bottom only):**
- ✅ `api/` can import `service/` and `models/`
- ✅ `service/` can import `storage/` and `models/`
- ✅ `storage/` can import `models/`
- ❌ `models/` should NOT import anything (pure data)
- ❌ `storage/` should NOT import `service/` (layering violation)

**Benefits:**
- **Separation of concerns:** Each package has one job
- **Testability:** Can test each layer independently
- **Maintainability:** Changes isolated to one layer
- **Dependency management:** Clear, one-way dependencies

#### 4. `data/` Directory

**Purpose:** Runtime data storage (gitignored)

```
data/
├── quest-todo.db         # SQLite database
├── quest-todo.db-wal     # Write-ahead log
├── quest-todo.db-shm     # Shared memory file
└── backups/              # Database backups
    └── quest-todo-2026-02-09.db
```

**Why separate directory?**
- ✅ Clean separation: code vs data
- ✅ Easy to backup (just copy `data/` folder)
- ✅ Easy to gitignore (don't commit user data)
- ✅ Easy to reset (delete `data/`, restart fresh)

---

## Go Module System

### What is a Go Module?

**Definition:** A collection of Go packages versioned together

### `go.mod` File

```go
module github.com/LaV72/quest-todo

go 1.25.0
```

**Line by line:**

1. **`module github.com/LaV72/quest-todo`**
   - Module path (unique identifier)
   - Not necessarily a real URL (but conventionally matches repo)
   - Used in imports: `import "github.com/LaV72/quest-todo/internal/models"`

2. **`go 1.25.0`**
   - Minimum Go version required
   - Ensures language features are available
   - Forward compatible (Go 1.26 can use modules requiring 1.25)

### Module Path Explained

**Why `github.com/LaV72/quest-todo`?**

**Convention (not requirement):**
```
github.com/<username>/<repository>
gitlab.com/<username>/<repository>
bitbucket.org/<username>/<repository>
```

**Benefits:**
1. **Unique globally:** No name collisions
2. **Go get works:** `go get github.com/LaV72/quest-todo`
3. **Discoverable:** Others can find your code
4. **Version control:** Tied to repository

**Could be anything:**
```
module my-quest-todo        # Works, but not recommended
module quest-todo           # Works, but might conflict
module example.com/quest    # Works if you own example.com
```

**Best practice:** Use GitHub/GitLab path even for private projects

### Imports Using Module Path

```go
// Full import path = module path + package path
import "github.com/LaV72/quest-todo/internal/models"
//      ─────────────────────────────── ────────────
//              module path              package path

// Another example:
import "github.com/LaV72/quest-todo/internal/storage/sqlite"
```

### `go.sum` File

```
modernc.org/sqlite v1.44.3 h1:abc123...
modernc.org/sqlite v1.44.3/go.mod h1:def456...
```

**What is this?**
- Checksums of dependencies
- Verifies downloads haven't been tampered with
- Ensures reproducible builds

**Why two lines per dependency?**
1. Hash of module code
2. Hash of module's `go.mod` file

**Do I commit `go.sum`?**
✅ YES - Commit to version control
- Ensures all developers use exact same dependency versions
- Prevents supply chain attacks

### Dependency Management

**Adding a dependency:**
```bash
# Option 1: Import in code, then tidy
# In your .go file:
import "modernc.org/sqlite"

# Then run:
go mod tidy
# Downloads package, adds to go.mod and go.sum

# Option 2: Explicit get
go get modernc.org/sqlite
```

**Updating dependencies:**
```bash
# Update one package
go get -u modernc.org/sqlite

# Update all packages
go get -u ./...

# Update to specific version
go get modernc.org/sqlite@v1.44.3
```

**Removing unused dependencies:**
```bash
go mod tidy
# Removes packages not imported anywhere
```

### Vendor Directory (Optional)

**Copy dependencies into project:**
```bash
go mod vendor
# Creates vendor/ directory with all dependencies

backend/
├── vendor/
│   └── modernc.org/
│       └── sqlite/
│           └── (entire package)
```

**Why vendor?**
- ✅ Offline builds (no internet needed)
- ✅ Guaranteed availability (even if package deleted online)
- ❌ Larger repository size

**When to vendor:**
- Production deployments
- Air-gapped environments
- Critical dependencies

---

## Configuration Setup

### `config.json`

```json
{
  "server": {
    "host": "localhost",
    "port": 3000
  },
  "storage": {
    "type": "sqlite",
    "path": "./data/quest-todo.db"
  },
  "app": {
    "deadlineThresholds": {
      "short": 3,
      "medium": 7,
      "long": 14
    },
    "defaultReward": 10,
    "pointsEnabled": true
  }
}
```

### Why JSON Configuration?

**JSON vs YAML vs TOML:**

| Feature | JSON | YAML | TOML |
|---------|------|------|------|
| Simple | ✅ Yes | ❌ Complex | ✅ Yes |
| Comments | ❌ No | ✅ Yes | ✅ Yes |
| Go stdlib | ✅ Yes | ❌ No (needs package) | ❌ No |
| Human-readable | ✅ OK | ✅ Great | ✅ Good |
| Multi-line | ❌ No | ✅ Yes | ✅ Yes |

**Why JSON for Quest Todo:**
- ✅ Built into Go (`encoding/json`)
- ✅ No external dependencies
- ✅ Simple, well-known format
- ✅ Easy to generate/edit programmatically
- ✅ Frontend can also parse (Swift has JSON support)

**Trade-off:** No comments in JSON
```json
{
  "port": 3000  // Can't add comment here
}
```

**Solution:** Use separate documentation or environment variables for overrides

### Configuration Structure

```json
{
  "server": {           // HTTP server settings
    "host": "localhost",
    "port": 3000
  },
  "storage": {          // Database settings
    "type": "sqlite",   // "sqlite", "json", or "memory"
    "path": "./data/quest-todo.db"
  },
  "app": {              // Application settings
    "deadlineThresholds": {
      "short": 3,       // Days
      "medium": 7,
      "long": 14
    },
    "defaultReward": 10,
    "pointsEnabled": true
  }
}
```

**Design decisions:**

1. **Nested structure** - Groups related settings
2. **Type-switchable** - `storage.type` allows pluggable backends
3. **Sensible defaults** - Values that work out of the box
4. **Extensible** - Easy to add new settings

### Environment Variable Overrides (Future)

**Pattern:**
```bash
# Config file: port = 3000
# Override with environment variable:
export QUEST_TODO_PORT=8080

# Application reads:
port := config.Server.Port
if envPort := os.Getenv("QUEST_TODO_PORT"); envPort != "" {
    port = envPort
}
```

**Benefits:**
- ✅ Different settings per environment (dev/staging/prod)
- ✅ No need to modify config file
- ✅ Secure for secrets (don't commit passwords to git)

---

## .gitignore Setup

```gitignore
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Data directory
data/*.db
data/*.json
data/*.db-*

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Config overrides
config.local.json
```

### Why Ignore These Files?

#### Binaries (`bin/`, `*.exe`)
```bash
# Don't commit compiled binaries
bin/quest-todo  # ❌ Don't commit (200MB, platform-specific)

# Commit source code instead
cmd/server/main.go  # ✅ Commit (2KB, works everywhere)
```

**Rationale:**
- Binaries are platform-specific (Mac binary won't run on Windows)
- Large files (MBs) bloat repository
- Can rebuild from source: `go build`

#### Data Files (`data/*.db`)
```bash
# Don't commit user data
data/quest-todo.db  # ❌ Don't commit (contains user's tasks)
```

**Rationale:**
- User data is private
- Each developer has their own local data
- Could contain sensitive information

#### IDE Files (`.idea/`, `.vscode/`)
```bash
# Don't commit IDE settings
.vscode/settings.json  # ❌ Don't commit (your personal preferences)
```

**Rationale:**
- Personal preferences (theme, formatting, shortcuts)
- Different team members use different IDEs
- Creates merge conflicts

#### OS Files (`.DS_Store`, `Thumbs.db`)
```bash
# macOS creates .DS_Store in every folder
.DS_Store  # ❌ Don't commit (macOS metadata)

# Windows creates Thumbs.db for image folders
Thumbs.db  # ❌ Don't commit (Windows thumbnail cache)
```

**Rationale:**
- Not part of project
- OS-specific
- Clutters repository

#### Test Output (`*.test`, `*.out`)
```bash
# go test creates binaries
mypackage.test  # ❌ Don't commit (temporary test binary)

# go test -coverprofile creates this
coverage.out  # ❌ Don't commit (temporary coverage data)
```

**Rationale:**
- Generated during testing
- Can be regenerated: `go test`
- Changes frequently

---

## Development Tools

### Essential Tools

#### 1. Go Compiler
```bash
# Install Go
brew install go  # macOS
# or download from golang.org

# Verify
go version  # go version go1.25.0 darwin/amd64
```

#### 2. Build Commands
```bash
# Build
go build -o bin/quest-todo cmd/server/main.go

# Build with optimizations (smaller binary)
go build -ldflags="-s -w" -o bin/quest-todo cmd/server/main.go
# -s: Strip symbol table
# -w: Strip debug info
# Result: ~50% smaller binary

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o bin/quest-todo-linux cmd/server/main.go
```

#### 3. Run Without Building
```bash
# Run directly (builds in temp directory)
go run cmd/server/main.go
```

#### 4. Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Opens in browser
```

#### 5. Formatting
```bash
# Format all Go files (standard style)
go fmt ./...

# Or use gofmt directly
gofmt -w .
```

**Go's formatting:**
- One standard style (no debates)
- Built into language
- Automatic (IDE does it on save)

#### 6. Linting
```bash
# Install golangci-lint
brew install golangci-lint

# Run linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

**Common issues caught:**
- Unused variables
- Unreachable code
- Inefficient code
- Style violations
- Potential bugs

#### 7. Dependency Management
```bash
# Download dependencies
go mod download

# Clean up unused dependencies
go mod tidy

# Verify dependencies
go mod verify

# View dependency graph
go mod graph

# Why is this dependency included?
go mod why golang.org/x/sys
```

### Optional Tools

#### 1. Air (Hot Reload)
```bash
# Install
go install github.com/cosmtrek/air@latest

# Run with auto-reload
air

# Now code changes automatically restart server
```

**`.air.toml` configuration:**
```toml
[build]
  cmd = "go build -o ./bin/server cmd/server/main.go"
  bin = "./bin/server"
  include_ext = ["go"]
  exclude_dir = ["vendor", "data"]
  delay = 1000  # ms
```

#### 2. Delve (Debugger)
```bash
# Install
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug cmd/server/main.go

# Commands:
# break main.main  # Set breakpoint
# continue         # Run to breakpoint
# next            # Step over
# step            # Step into
# print varName   # Inspect variable
```

#### 3. Go Modules Graph Visualization
```bash
# Install
go install github.com/nikolaydubina/import-graph@latest

# Generate graph
import-graph -i github.com/LaV72/quest-todo/cmd/server | dot -Tpng -o graph.png
```

---

## Build Process Explained

### What Happens During `go build`?

```bash
go build -o bin/quest-todo cmd/server/main.go
```

**Steps:**

1. **Parse:** Read `main.go` and parse Go syntax
2. **Resolve imports:** Find all imported packages
3. **Compile packages:** Compile each package in dependency order
4. **Link:** Combine all packages into single binary
5. **Output:** Write binary to `bin/quest-todo`

**Result:**
- Single executable file
- No runtime needed (unlike Java, Python, Node.js)
- All dependencies included
- Platform-specific binary

### Static vs Dynamic Linking

**Go default: Static linking**
```bash
# One binary, no dependencies
./quest-todo  # Just works, no libraries needed
```

**Compare to C:**
```bash
# C program needs shared libraries
./myprogram
# Error: libssl.so.1.1 not found
```

**Why static linking is great:**
- ✅ Deploy one file
- ✅ Works on any system (same OS/arch)
- ✅ No "dependency hell"
- ❌ Larger binary size (acceptable trade-off)

### Binary Size

```bash
# Minimal Go binary: ~2MB (includes Go runtime, garbage collector)
# Our binary: ~10MB (includes SQLite, all packages)
```

**Size breakdown:**
- Go runtime: ~1MB
- Standard library: ~1MB
- SQLite: ~2MB
- Our code + dependencies: ~6MB

**Reducing size:**
```bash
# Strip debug info
go build -ldflags="-s -w" -o bin/quest-todo cmd/server/main.go
# Result: ~5MB (50% smaller)

# UPX compression (optional)
upx --best bin/quest-todo
# Result: ~2MB (80% smaller, but slower startup)
```

---

## Summary

### Key Decisions

| Decision | Rationale |
|----------|-----------|
| Go language | Single binary, fast, simple, low memory |
| `cmd/` directory | Separate entry points from library code |
| `internal/` directory | Prevent external imports, clear API boundary |
| Layered packages | Separation of concerns, testability |
| Go modules | Standard dependency management |
| JSON config | Built-in support, no dependencies |
| `.gitignore` rules | Don't commit binaries, data, or OS files |

### Project Characteristics

**Simplicity:**
- Small, focused codebase
- One obvious way to structure
- Minimal dependencies

**Maintainability:**
- Clear layer boundaries
- One-way dependencies
- Self-documenting structure

**Deployability:**
- Single binary
- No runtime needed
- Cross-platform compilation

---

## Next Steps

Once project structure is in place:
1. Implement models (data structures)
2. Implement storage interface
3. Implement storage backends (SQLite, JSON)
4. Implement business logic
5. Implement API handlers
6. Wire everything together in `main.go`

The solid foundation makes building on top straightforward and maintainable.
