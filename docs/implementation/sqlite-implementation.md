# SQLite Implementation

Comprehensive documentation for the SQLite storage backend implementation.

## Table of Contents

1. [Overview](#overview)
2. [Why SQLite?](#why-sqlite)
3. [Database Connection](#database-connection)
4. [Schema Design](#schema-design)
5. [Migration System](#migration-system)
6. [Indexing Strategy](#indexing-strategy)
7. [Performance Considerations](#performance-considerations)
8. [Backup and Maintenance](#backup-and-maintenance)

---

## Overview

SQLite is an embedded, serverless SQL database engine. Unlike client-server databases (MySQL, PostgreSQL), SQLite runs **inside** the application process and stores data in a single file.

**Key characteristics:**
- ✅ Zero configuration - no setup required
- ✅ Single file - `quest-todo.db` contains everything
- ✅ ACID compliant - reliable transactions
- ✅ Cross-platform - works on macOS, Linux, Windows
- ✅ Fast for local applications
- ✅ No network latency - direct file access

---

## Why SQLite?

### Comparison with Alternatives

#### SQLite vs PostgreSQL/MySQL

| Feature | SQLite | PostgreSQL/MySQL |
|---------|--------|------------------|
| Setup | None - just open file | Install and configure server |
| Deployment | Ship single file | Manage separate database server |
| Concurrency | Multiple readers, one writer | Many concurrent connections |
| Data Size | Works well up to ~1TB | Handles massive datasets |
| Network | No network overhead | Network latency on every query |
| Backup | Copy file | Complex backup procedures |
| Use Case | Local apps, embedded | Web services, high concurrency |

#### SQLite vs JSON Files

| Feature | SQLite | JSON Files |
|---------|--------|------------|
| Query Speed | Fast (indexed, O(log n)) | Slow (must scan all data) |
| Data Integrity | ACID transactions | Manual handling |
| Relationships | Foreign keys, joins | Manual management |
| Concurrent Access | Built-in locking | Custom locking needed |
| Scalability | Handles large datasets | Slow with 1000+ items |

**Conclusion:** SQLite is perfect for Quest Todo because:
1. Single-user local application
2. Needs fast queries with filters/sorting
3. Benefits from indexes and transactions
4. Easy to backup (just copy the file)

---

## Database Connection

### Connection String (DSN)

```go
dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on", path)
db, err := sql.Open("sqlite", dsn)
```

### DSN Parameters Explained

#### 1. `_journal_mode=WAL` (Write-Ahead Logging)

**What is WAL?**

SQLite has two journaling modes:

**DELETE mode (default):**
```
Write operation:
1. Copy original data to journal file
2. Modify database file
3. Delete journal file when complete

Problem: Readers must wait for writers to finish
```

**WAL mode (Write-Ahead Log):**
```
Write operation:
1. Append changes to WAL file
2. Periodically checkpoint: merge WAL into main database

Benefit: Readers read from main file while writer writes to WAL
Result: Readers NEVER block writers, writers NEVER block readers
```

**Performance impact:**
```
Without WAL:
- Read while writing: BLOCKED (waits)
- Write while reading: BLOCKED (waits)

With WAL:
- Read while writing: ✅ Continues reading old data
- Write while reading: ✅ Continues writing to WAL
```

**Trade-offs:**
- ✅ Much better concurrency
- ✅ Faster writes (no blocking)
- ❌ Slightly more disk space (WAL file exists)
- ❌ Need periodic checkpoints

#### 2. `_busy_timeout=5000` (5 seconds)

**What it does:**
If database is locked, retry for up to 5 seconds before giving up.

**Without busy_timeout:**
```go
// Writer is holding lock
// Reader tries to access
// Immediate error: "database is locked"
```

**With busy_timeout=5000:**
```go
// Writer is holding lock (for 1 second)
// Reader tries to access
// Reader waits... checks every 100ms
// After 1 second: Writer releases lock
// Reader acquires lock and succeeds
```

**When does locking occur?**
- SQLite allows: **Multiple simultaneous readers** OR **one writer**
- Cannot have: Writer + Reader at same time (brief lock during WAL checkpoint)

#### 3. `_foreign_keys=on`

**Enables referential integrity:**

```sql
-- With foreign keys enabled:
DELETE FROM tasks WHERE id = 'task-123';
-- Also deletes all objectives with task_id = 'task-123' (CASCADE)

-- Prevents orphaned data:
INSERT INTO objectives (task_id, ...) VALUES ('nonexistent-id', ...);
-- ERROR: foreign key constraint failed
```

**Without foreign keys:**
- Orphaned objectives remain after task deletion
- Must manually clean up related data
- Data integrity not guaranteed

### Connection Pool Configuration

```go
db.SetMaxOpenConns(1)
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(0)
```

#### Why Only 1 Connection?

**SQLite concurrency model:**
- ✅ Unlimited concurrent readers (SELECT)
- ⚠️ Only ONE writer at a time (INSERT, UPDATE, DELETE)

**Multiple connections cause problems:**

```go
// With 2 connections:
Connection 1: BEGIN TRANSACTION; UPDATE tasks ...
Connection 2: BEGIN TRANSACTION; UPDATE tasks ...
// Result: "database is locked" error on Connection 2
```

**Single connection = serialized writes:**
```go
// With 1 connection:
Operation 1: UPDATE tasks ... (completes)
Operation 2: UPDATE tasks ... (waits, then executes)
// Result: No lock errors, operations queue naturally
```

**Trade-off:**
- ✅ No lock errors
- ✅ Simpler mental model
- ✅ Still fast (local disk is very fast)
- ❌ Writes are serialized (but this is fine for single-user app)

---

## Schema Design

### Table Structure

#### 1. Tasks Table

```sql
CREATE TABLE tasks (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT,
	priority INTEGER NOT NULL CHECK(priority BETWEEN 1 AND 5),
	deadline_type TEXT NOT NULL DEFAULT 'none',
	deadline_date DATETIME,
	category TEXT,
	status TEXT NOT NULL DEFAULT 'active',
	notes TEXT,
	reward INTEGER DEFAULT 0,
	order_index INTEGER DEFAULT 0,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	completed_at DATETIME
);
```

**Design decisions:**

##### ID as TEXT (UUID)
```go
// Could use INTEGER:
id INTEGER PRIMARY KEY AUTOINCREMENT

// Why TEXT (UUID)?
id TEXT PRIMARY KEY  // "550e8400-e29b-41d4-a009-426614174000"
```

**Benefits of UUID:**
- ✅ Globally unique (can merge databases)
- ✅ Non-sequential (no information leakage: can't guess next ID)
- ✅ Client-side generation (don't need DB to create IDs)
- ✅ Distributed systems friendly

**Trade-off:**
- ❌ Slightly larger (36 bytes vs 8 bytes)
- ❌ Not human-readable
- ✅ For our scale (1000s of tasks), size difference is negligible

##### CHECK Constraints
```sql
priority INTEGER NOT NULL CHECK(priority BETWEEN 1 AND 5)
```

**Database-level validation:**
```go
// Attempt to insert invalid data:
INSERT INTO tasks (priority) VALUES (10);
// ERROR: CHECK constraint failed: priority BETWEEN 1 AND 5
```

**Defense in depth:**
1. Frontend validation (user experience)
2. API validation (business logic)
3. Database constraints (data integrity)

##### Deadline as Two Columns
```sql
deadline_type TEXT NOT NULL DEFAULT 'none',
deadline_date DATETIME,
```

**Why not one column?**

**Option A: Single DATETIME column:**
```sql
deadline DATETIME
-- Problem: How to represent "short" vs "medium" vs "long"?
-- Must calculate type from date, or store type elsewhere
```

**Option B: Two columns (our choice):**
```sql
deadline_type TEXT,  -- "short", "medium", "long", "none"
deadline_date DATETIME  -- Actual date (nullable)
```

**Benefits:**
- Can query by type: `WHERE deadline_type = 'short'`
- Can have type without date: "short deadline, date TBD"
- Maps directly to Go struct:
```go
type Deadline struct {
	Type string     `db:"deadline_type"`
	Date *time.Time `db:"deadline_date"`
}
```

##### completed_at as Nullable
```sql
completed_at DATETIME  -- NULL = not completed
```

**Why nullable instead of zero value?**

```sql
-- Zero value approach:
completed_at DATETIME DEFAULT '1970-01-01'
-- Problem: Must remember that '1970-01-01' means "not completed"
-- Queries: WHERE completed_at != '1970-01-01'
-- Confusing and error-prone

-- Nullable approach:
completed_at DATETIME  -- NULL = not completed
-- Clear: WHERE completed_at IS NULL
-- Standard SQL idiom
```

#### 2. Objectives Table (Sub-tasks)

```sql
CREATE TABLE objectives (
	id TEXT PRIMARY KEY,
	task_id TEXT NOT NULL,
	text TEXT NOT NULL,
	completed BOOLEAN NOT NULL DEFAULT 0,
	order_index INTEGER DEFAULT 0,
	created_at DATETIME NOT NULL,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

**Foreign Key with CASCADE:**

```sql
FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
```

**What this does:**

```sql
-- Task has 3 objectives:
tasks:       id='task-1', title='Project'
objectives:  id='obj-1', task_id='task-1', text='Step 1'
objectives:  id='obj-2', task_id='task-1', text='Step 2'
objectives:  id='obj-3', task_id='task-1', text='Step 3'

-- Delete the task:
DELETE FROM tasks WHERE id = 'task-1';

-- CASCADE automatically deletes objectives:
-- obj-1, obj-2, obj-3 all deleted
-- No orphaned data!
```

**Without CASCADE:**
```sql
-- Must manually delete objectives first:
DELETE FROM objectives WHERE task_id = 'task-1';
DELETE FROM tasks WHERE id = 'task-1';

-- If you forget: orphaned objectives remain
-- Database becomes polluted over time
```

**BOOLEAN Storage in SQLite:**

SQLite doesn't have a native BOOLEAN type. It stores as INTEGER:
- `0` = false
- `1` = true

```go
// Go side:
completed := true

// SQLite side:
INSERT INTO objectives (completed) VALUES (1);

// Query back:
SELECT completed FROM objectives;  // Returns 1
// Go's sql package converts 1 → true automatically
```

#### 3. Task Tags (Many-to-Many)

```sql
CREATE TABLE task_tags (
	task_id TEXT NOT NULL,
	tag TEXT NOT NULL,
	PRIMARY KEY (task_id, tag),
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

**Composite Primary Key:**

```sql
PRIMARY KEY (task_id, tag)
```

**Prevents duplicates:**
```sql
-- First insert: OK
INSERT INTO task_tags VALUES ('task-1', 'work');

-- Duplicate insert: ERROR
INSERT INTO task_tags VALUES ('task-1', 'work');
-- ERROR: PRIMARY KEY constraint failed
```

**Example data:**
```
| task_id | tag          |
|---------|--------------|
| task-1  | work         |  ← Task 1 has "work" tag
| task-1  | urgent       |  ← Task 1 has "urgent" tag
| task-2  | work         |  ← Task 2 has "work" tag
| task-2  | code-review  |  ← Task 2 has "code-review" tag
| task-3  | personal     |  ← Task 3 has "personal" tag
```

**Query examples:**
```sql
-- Find all tasks with "work" tag:
SELECT tasks.* FROM tasks
JOIN task_tags ON tasks.id = task_tags.task_id
WHERE task_tags.tag = 'work';
-- Returns: task-1, task-2

-- Find all tags for task-1:
SELECT tag FROM task_tags WHERE task_id = 'task-1';
-- Returns: work, urgent

-- Count tasks per tag:
SELECT tag, COUNT(*) FROM task_tags GROUP BY tag;
-- Returns: work=2, urgent=1, code-review=1, personal=1
```

#### 4. Categories Table

```sql
CREATE TABLE categories (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	color TEXT NOT NULL,
	icon TEXT,
	type TEXT NOT NULL DEFAULT 'main',
	order_index INTEGER DEFAULT 0,
	created_at DATETIME NOT NULL
);
```

**Color as TEXT (Hex code):**
```sql
color TEXT NOT NULL  -- "#3A7F8F"
```

**Why not RGB integers?**

**Option A: Three columns:**
```sql
color_r INTEGER,  -- 58
color_g INTEGER,  -- 127
color_b INTEGER   -- 143
```

**Option B: Single hex string (our choice):**
```sql
color TEXT  -- "#3A7F8F"
```

**Benefits:**
- ✅ Matches web standards (CSS uses hex)
- ✅ Frontend can use directly: `<div style="color: {category.color}">`
- ✅ Single column (simpler schema)
- ✅ Easy to copy/paste from design tools

---

## Migration System

### How Migrations Work

```go
func (s *SQLiteStorage) migrate() error {
	version := s.getSchemaVersion()  // Current version

	migrations := []migration{
		{version: 1, name: "initial_schema", up: migrateV1},
		{version: 2, name: "add_templates", up: migrateV2},  // Future
	}

	for _, m := range migrations {
		if version < m.version {
			m.up(s.db)  // Apply migration
			s.setSchemaVersion(m.version)  // Record it
		}
	}
}
```

### Schema Version Tracking

```sql
CREATE TABLE schema_version (
	version INTEGER PRIMARY KEY,
	applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**After migrations run:**
```
| version | applied_at          |
|---------|---------------------|
| 1       | 2026-02-09 12:00:00 |
| 2       | 2026-02-15 14:30:00 |
```

### Migration Example

**Scenario: Adding a "priority_color" column**

```go
// Add to migrations list:
{version: 2, name: "add_priority_color", up: migrateV2}

func migrateV2(db *sql.DB) error {
	_, err := db.Exec(`
		ALTER TABLE tasks ADD COLUMN priority_color TEXT DEFAULT '#000000';
	`)
	return err
}
```

**What happens:**

```
User A (has v1 database):
1. Opens app
2. Migration system checks: current version = 1
3. Sees migration v2 needs to run
4. Runs: ALTER TABLE tasks ADD COLUMN priority_color ...
5. Records: schema_version = 2
6. Database now at v2

User B (has v2 database):
1. Opens app
2. Migration system checks: current version = 2
3. No migrations needed
4. App starts normally
```

### Benefits of Migration System

**1. Incremental Updates:**
- Don't recreate entire database
- Preserve existing data
- Add features without breaking old data

**2. Version Tracking:**
- Know exactly what schema version you have
- Can detect mismatches (e.g., old app on new database)
- Audit trail of changes

**3. Team Coordination:**
```
Developer A: Adds migration v2 (new column)
Developer B: Adds migration v3 (new table)

When merged:
- Both developers' databases update correctly
- No conflicts (migrations run in order)
```

**4. Rollback Support (future):**
```go
type migration struct {
	up   func(*sql.DB) error  // Apply change
	down func(*sql.DB) error  // Undo change
}

// Can add rollback functionality:
func (s *SQLiteStorage) rollback() error {
	// Run down migrations in reverse
}
```

---

## Indexing Strategy

### What Are Indexes?

An index is a data structure (B-tree) that makes searches fast.

**Without index:**
```
Finding status='active':
[task1: active] ✓ Found!
[task2: complete]
[task3: active] ✓ Found!
[task4: complete]
[task5: active] ✓ Found!
... scan all 10,000 rows
Time: O(n) - Linear scan
```

**With index:**
```
Index on status:
  active → [task1, task3, task5, ...]
  complete → [task2, task4, ...]

Finding status='active':
Look up "active" in index → [task1, task3, task5]
Time: O(log n) - B-tree lookup
```

### B-Tree Structure

```
              [M]
            /     \
        [D]         [S]
       /   \       /   \
     [A-C][E-L] [N-R][T-Z]
```

**Finding "Priority = 4":**
```
Root: Is 4 < M? No → Go right
Node: Is 4 < S? Yes → Go left
Leaf: Found 4!

Steps: 3 (instead of scanning all rows)
```

### Our Indexes

#### Single-Column Indexes

```sql
-- Fast lookups by status
CREATE INDEX idx_tasks_status ON tasks(status);
-- Query: WHERE status = 'active'
-- Time: O(log n) instead of O(n)

-- Fast lookups by priority (descending)
CREATE INDEX idx_tasks_priority ON tasks(priority DESC);
-- Query: ORDER BY priority DESC
-- Can read index in order, no sorting needed

-- Fast lookups by deadline
CREATE INDEX idx_tasks_deadline ON tasks(deadline_date);
-- Query: WHERE deadline_date < NOW()
-- Finds overdue tasks quickly

-- Fast lookups by category
CREATE INDEX idx_tasks_category ON tasks(category);
-- Query: WHERE category = 'work'
```

#### Composite Indexes

```sql
CREATE INDEX idx_tasks_status_priority ON tasks(status, priority DESC);
```

**What it optimizes:**
```sql
SELECT * FROM tasks
WHERE status = 'active'
ORDER BY priority DESC;
```

**How it works:**

Index stores rows in this order:
```
status='active', priority=5
status='active', priority=5
status='active', priority=4
status='active', priority=3
status='complete', priority=5
status='complete', priority=4
```

**Query execution:**
1. Seek to first "active" row in index
2. Read rows in order (already sorted by priority)
3. Stop when reaching "complete"

**No separate sort step needed!**

#### Index on Foreign Keys

```sql
CREATE INDEX idx_objectives_task ON objectives(task_id);
```

**Optimizes:**
```sql
-- Loading objectives for a task:
SELECT * FROM objectives WHERE task_id = 'task-123';
-- Without index: Scan all objectives (O(n))
-- With index: Direct lookup (O(log n))

-- Deleting task with CASCADE:
DELETE FROM tasks WHERE id = 'task-123';
-- Must find all related objectives to delete
-- Index makes this fast
```

### Index Size vs Query Speed

**Trade-offs:**

**Benefits:**
- ✅ 10-100x faster queries
- ✅ Critical for good user experience

**Costs:**
- ❌ Extra disk space (~10-20% of table size per index)
- ❌ Slower writes (must update indexes)
- ❌ More RAM usage (indexes cached in memory)

**For Quest Todo:**
- 1,000 tasks × 10 indexes ≈ 1 MB total
- Write performance impact: negligible (few writes per minute)
- Query speed: crucial (many queries per second)

**Conclusion: Indexes are worth it**

---

## Performance Considerations

### Query Performance

**Best practices:**

#### 1. Use Indexes
```sql
-- Bad: Full table scan
SELECT * FROM tasks WHERE title LIKE '%project%';
-- Must scan every row

-- Good: Use indexed column
SELECT * FROM tasks WHERE status = 'active';
-- Uses idx_tasks_status
```

#### 2. Limit Result Sets
```go
// Bad: Load everything
tasks, _ := storage.ListTasks(TaskFilter{})
// Loads all 10,000 tasks into memory

// Good: Paginate
tasks, _ := storage.ListTasks(TaskFilter{
	Limit: 50,
	Offset: 0,
})
// Loads only 50 tasks
```

#### 3. Use Prepared Statements
```go
// Bad: SQL injection risk, slower
query := fmt.Sprintf("SELECT * FROM tasks WHERE id = '%s'", id)
db.Query(query)

// Good: Safe, faster (query plan cached)
db.Query("SELECT * FROM tasks WHERE id = ?", id)
```

### Transaction Performance

**Single transaction vs multiple:**

```go
// Slow: 1000 separate transactions
for _, task := range tasks {
	db.Exec("INSERT INTO tasks ...", task)
}
// Each INSERT = separate disk write
// Time: ~1000ms

// Fast: Single transaction
tx, _ := db.Begin()
for _, task := range tasks {
	tx.Exec("INSERT INTO tasks ...", task)
}
tx.Commit()
// One disk write at commit
// Time: ~50ms
```

**Why?**
- Each transaction = fsync() call = wait for disk
- Batching = one fsync() for all operations

### Memory Usage

**SQLite caches pages in memory:**

```go
// Default cache: ~2MB
PRAGMA cache_size = -2000;  // 2000 KB

// For better performance:
PRAGMA cache_size = -10000;  // 10 MB
```

**Our data:**
- 1,000 tasks ≈ 200 KB
- With indexes ≈ 500 KB
- 10 MB cache = everything fits in RAM

**Result: All queries hit RAM cache (very fast)**

---

## Backup and Maintenance

### Backup Methods

#### 1. File Copy (Simple)

```bash
cp quest-todo.db quest-todo-backup.db
```

**Pros:**
- ✅ Simple
- ✅ Fast
- ✅ Works with any tool

**Cons:**
- ⚠️ Must ensure no writes during copy
- ⚠️ May copy inconsistent state if app is running

#### 2. VACUUM INTO (Safe)

```go
storage.Backup("./backups/quest-todo-backup.db")

// Internally runs:
VACUUM INTO './backups/quest-todo-backup.db'
```

**Pros:**
- ✅ Safe even while app is running
- ✅ Optimizes/defragments during copy
- ✅ Output is clean, compact file

**How it works:**
1. Reads all data from main database
2. Writes to new file in optimal format
3. No gaps, no fragmentation
4. Safe snapshot (transactionally consistent)

#### 3. Online Backup API (Best)

```go
import "github.com/mattn/go-sqlite3"

func backup(srcDB, destPath string) error {
	dest, _ := sql.Open("sqlite3", destPath)
	defer dest.Close()

	srcConn := srcDB.Driver().Open(srcDB)
	destConn := dest.Driver().Open(dest)

	backup, _ := destConn.Backup("main", srcConn, "main")
	backup.Step(-1)  // Copy all pages
	backup.Finish()
}
```

**Pros:**
- ✅ Official SQLite backup API
- ✅ Can copy while app is running
- ✅ Can show progress (step by step)
- ✅ Handles locking correctly

### Maintenance

#### VACUUM (Defragmentation)

```sql
VACUUM;
```

**What it does:**
- Rebuilds database file
- Removes deleted/fragmented space
- Optimizes layout

**When to run:**
- After deleting many tasks
- Database file larger than expected
- Periodic maintenance (monthly)

**Example:**
```
Before VACUUM:
- 1000 tasks created
- 500 tasks deleted
- File size: 5 MB (includes deleted data)

After VACUUM:
- File size: 2.5 MB (deleted data reclaimed)
```

#### ANALYZE (Statistics)

```sql
ANALYZE;
```

**What it does:**
- Gathers statistics about data distribution
- Helps query planner choose optimal indexes
- Improves query performance

**When to run:**
- After bulk imports
- After schema changes
- Periodic maintenance (weekly)

#### Integrity Check

```sql
PRAGMA integrity_check;
```

**Returns:**
- "ok" if database is healthy
- List of errors if corruption detected

**When to run:**
- After crashes
- Before important backups
- Periodic health checks

---

## Summary

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| WAL mode | Better concurrency for read-heavy workload |
| Single connection | Avoid lock errors, simpler model |
| UUID as TEXT | Globally unique, client-side generation |
| Separate deadline columns | Query by type, map to Go struct |
| Foreign keys with CASCADE | Prevent orphaned data |
| Composite indexes | Optimize common query patterns |
| Migration system | Safe schema evolution |

### Performance Characteristics

**For typical workload (1,000-10,000 tasks):**
- Single task lookup: <1ms (indexed)
- List 50 tasks: <5ms (indexed + cached)
- Create task: <10ms (single write)
- Bulk create 100 tasks: <50ms (batched)
- Full-text search: <50ms (depends on data size)

### Best Practices

1. ✅ Use transactions for bulk operations
2. ✅ Always use prepared statements (?)
3. ✅ Leverage indexes for common queries
4. ✅ Paginate large result sets
5. ✅ Run VACUUM periodically
6. ✅ Use VACUUM INTO for backups
7. ✅ Check database integrity after crashes

---

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [SQLite Query Planner](https://www.sqlite.org/queryplanner.html)
- [Write-Ahead Logging](https://www.sqlite.org/wal.html)
- [modernc.org/sqlite Driver](https://gitlab.com/cznic/sqlite)
