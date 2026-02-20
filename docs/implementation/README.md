# Implementation Documentation

This folder contains detailed implementation documentation for the Quest Todo backend.

## Overview

These documents explain **how** the system is built, with deep dives into:
- Code structure and organization
- Design decisions and rationale
- Go language features and patterns
- Performance considerations
- Best practices

## Documents

### [Go Project Structure](./go-project-structure.md)

Complete guide to the backend project setup:
- Why Go was chosen over Node.js/Python
- Directory structure (`cmd/`, `internal/`, packages)
- Go module system (`go.mod`, imports, dependencies)
- Configuration setup (JSON, environment variables)
- Development tools (building, testing, linting)
- Build process and binary optimization

**Read this first** to understand the foundation.

### [Data Models Implementation](./data-models-implementation.md)

Deep dive into Go data structures:
- Struct design principles
- Struct tags (JSON, database, validation)
- Core models line-by-line (Task, Category, Objective)
- Custom types for type safety (TaskStatus)
- Request/response separation
- Pointers vs values (when to use each)
- Best practices and common pitfalls

**Read this** to understand how data flows through the system.

### [SQLite Implementation](./sqlite-implementation.md)

Everything about the database layer:
- Why SQLite (vs PostgreSQL, JSON)
- Connection setup (WAL mode, busy timeout)
- Schema design (tables, foreign keys, constraints)
- Migration system
- Indexing strategy (B-trees, composite indexes)
- Performance optimization
- Backup and maintenance

**Read this** to understand data persistence and performance.

### [Service Layer Implementation](./service-layer-implementation.md)

Business logic and domain services:
- Service layer architecture and responsibilities
- Dependency injection (storage, clock, ID generator)
- Business rules and validation
- Progress calculation and status transitions
- Error handling and wrapping
- Configurable behavior patterns
- Service testing strategies

**Read this** to understand business logic and orchestration.

### [API Layer Implementation](./api-layer-implementation.md)

REST API handlers and HTTP layer:
- HTTP handler architecture
- Request parsing and validation
- Error mapping (service errors to HTTP status codes)
- Response formatting (success/error envelopes)
- Query parameter parsing and filtering
- Pagination and metadata
- Complete API endpoint reference
- REST best practices

**Read this** to understand the HTTP interface and API design.

### [Router and Middleware Implementation](./router-middleware-implementation.md)

HTTP routing and cross-cutting concerns:
- Router architecture using net/http ServeMux
- Middleware pattern and composition
- Request ID generation and tracing
- Request/response logging with timing
- Panic recovery and error handling
- CORS configuration and preflight handling
- Content-Type validation
- Complete route table (26 endpoints)
- Middleware testing strategies

**Read this** to understand request routing and middleware functionality.

### [Main Server Implementation](./main-server-implementation.md)

Server entry point and application bootstrap:
- Configuration management via environment variables
- Dependency injection patterns
- Server initialization and lifecycle
- Graceful shutdown handling
- HTTP server configuration (timeouts, connection pooling)
- Signal handling (SIGINT, SIGTERM)
- Logging strategy
- Production deployment considerations
- Docker and Kubernetes examples

**Read this** to understand how the application starts, configures, and shuts down.

### [Custom Binary Storage Implementation](./custom-binary-storage-implementation.md)

Custom binary storage format design (planned):
- Binary file format specification
- Multi-section layout (header, indexes, data, WAL)
- Index structures for fast queries
- Memory-mapped I/O strategies
- Implementation phases
- Performance characteristics

**Read this** to understand the future custom storage backend.

## Learning Path

### For New Developers

1. **Start with:** [Go Project Structure](./go-project-structure.md)
   - Get oriented with the codebase layout
   - Understand Go modules and tooling

2. **Then read:** [Data Models Implementation](./data-models-implementation.md)
   - See how data is modeled
   - Learn struct tags and validation

3. **Storage Layer:** [SQLite Implementation](./sqlite-implementation.md)
   - Understand data persistence
   - Learn about performance and indexing

4. **Business Logic:** [Service Layer Implementation](./service-layer-implementation.md)
   - Learn business rules and validation
   - Understand dependency injection
   - See how services orchestrate operations

5. **HTTP Interface:** [API Layer Implementation](./api-layer-implementation.md)
   - Learn REST API design
   - Understand request/response handling
   - See error mapping and status codes

6. **Routing & Middleware:** [Router and Middleware Implementation](./router-middleware-implementation.md)
   - Learn HTTP routing patterns
   - Understand middleware composition
   - See cross-cutting concerns (logging, CORS, recovery)

7. **Server Bootstrap:** [Main Server Implementation](./main-server-implementation.md)
   - Learn application initialization
   - Understand dependency injection
   - See configuration management and graceful shutdown

### For Backend Architects

Focus on these documents to understand the full architecture:
- [Service Layer Implementation](./service-layer-implementation.md) - Business logic patterns
- [API Layer Implementation](./api-layer-implementation.md) - HTTP interface design
- [Router and Middleware Implementation](./router-middleware-implementation.md) - Routing and middleware
- [Custom Binary Storage Implementation](./custom-binary-storage-implementation.md) - Future storage optimization

### For Quick Reference

Each document has a **Table of Contents** at the top for quick navigation to specific topics.

## Related Documentation

- **Design docs** (`../`) - High-level architecture and interfaces
- **Development guide** (`../development-guide.md`) - Setup and workflow

## Purpose

These implementation documents serve several purposes:

1. **Onboarding** - Help new developers understand the codebase
2. **Reference** - Quick lookup for "why did we do it this way?"
3. **Knowledge transfer** - Preserve design decisions and rationale
4. **Learning** - Detailed explanations of Go patterns and best practices
5. **Maintenance** - Context for future changes and refactoring

## Writing Style

These documents:
- ✅ Explain **what** the code does
- ✅ Explain **why** decisions were made
- ✅ Show alternatives and trade-offs
- ✅ Include code examples with annotations
- ✅ Highlight common pitfalls
- ✅ Reference Go best practices

## Keeping Documentation Updated

When making significant changes:
1. Update the relevant implementation doc
2. Add notes about new patterns or decisions
3. Update examples if interfaces change
4. Keep rationale sections current

---

Last updated: February 20, 2026 (added Router & Middleware, Main Server documentation)
