# Quest Todo Backend Makefile

# Binary name
BINARY_NAME=quest-todo-server
BINARY_PATH=bin/$(BINARY_NAME)

# Source directory
CMD_DIR=./cmd/server

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test coverage run dev help

# Default target
all: clean test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(CMD_DIR)
	@echo "Build complete: $(BINARY_PATH)"

# Build with optimizations (smaller binary)
build-prod:
	@echo "Building $(BINARY_NAME) for production..."
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) $(CMD_DIR)
	@echo "Production build complete: $(BINARY_PATH)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f *.db *.db-shm *.db-wal
	@rm -f coverage.out
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	@echo "\nCoverage report:"
	$(GOCMD) tool cover -func=coverage.out
	@echo "\nTo view HTML coverage report, run: make coverage-html"

# View coverage in browser
coverage-html: coverage
	$(GOCMD) tool cover -html=coverage.out

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

# Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_PATH)

# Run in development mode (rebuild on change not included, use air for that)
dev:
	@echo "Running in development mode..."
	@PORT=8080 ENABLE_LOGGING=true ./$(BINARY_PATH)

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Format complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies tidied"

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Dependencies verified"

# Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(CMD_DIR)
	@echo "Install complete"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...
	@echo "Lint complete"

# Run security check (requires gosec)
security:
	@echo "Running security check..."
	@which gosec > /dev/null || (echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
	gosec ./...
	@echo "Security check complete"

# Show help
help:
	@echo "Quest Todo Backend Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all           - Clean, test, and build (default)"
	@echo "  build         - Build the application"
	@echo "  build-prod    - Build with optimizations for production"
	@echo "  clean         - Remove build artifacts and test files"
	@echo "  test          - Run all tests"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  coverage-html - View coverage report in browser"
	@echo "  test-race     - Run tests with race detector"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Run in development mode"
	@echo "  fmt           - Format all Go code"
	@echo "  deps          - Download dependencies"
	@echo "  tidy          - Tidy dependencies (remove unused)"
	@echo "  verify        - Verify dependencies"
	@echo "  install       - Install binary to \$$GOPATH/bin"
	@echo "  lint          - Run golangci-lint (requires golangci-lint)"
	@echo "  security      - Run security check (requires gosec)"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make              # Clean, test, and build"
	@echo "  make test         # Run tests"
	@echo "  make run          # Build and run"
	@echo "  make build-prod   # Build optimized binary"
