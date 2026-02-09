# Data Models

Complete data model specifications for Quest Todo.

## Core Models

### Task

The primary entity representing a to-do item or quest.

```go
type Task struct {
    // Identity
    ID          string      `json:"id"`

    // Core Fields
    Title       string      `json:"title" validate:"required,min=1,max=200"`
    Description string      `json:"description" validate:"max=2000"`
    Priority    int         `json:"priority" validate:"min=1,max=5"`
    Deadline    Deadline    `json:"deadline"`
    Category    string      `json:"category"`
    Status      TaskStatus  `json:"status"`

    // Progress Tracking
    Objectives  []Objective `json:"objectives"`
    Notes       string      `json:"notes"`
    Reward      int         `json:"reward"`
    Tags        []string    `json:"tags"`

    // Metadata
    Order       int         `json:"order"`         // Manual ordering
    CreatedAt   time.Time   `json:"createdAt"`
    UpdatedAt   time.Time   `json:"updatedAt"`
    CompletedAt *time.Time  `json:"completedAt"`   // nil if not completed

    // Computed Fields (calculated on read, not stored)
    Progress    float64     `json:"progress"`      // 0.0 - 1.0
    IsOverdue   bool        `json:"isOverdue"`
    DaysLeft    int         `json:"daysLeft"`
}
```

**Field Descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (UUID) |
| `title` | string | Yes | Task name (1-200 chars) |
| `description` | string | No | Detailed description (max 2000 chars) |
| `priority` | int | Yes | Priority level 1-5 (★ to ★★★★★) |
| `deadline` | Deadline | No | Deadline information |
| `category` | string | No | Category ID |
| `status` | TaskStatus | Yes | Current status |
| `objectives` | []Objective | No | Sub-tasks/steps |
| `notes` | string | No | Additional notes |
| `reward` | int | No | Points/reward value |
| `tags` | []string | No | Tags for organization |
| `order` | int | Yes | Display order (0-based) |
| `createdAt` | timestamp | Yes | Creation timestamp |
| `updatedAt` | timestamp | Yes | Last update timestamp |
| `completedAt` | timestamp | No | Completion timestamp (null if incomplete) |
| `progress` | float | Computed | Completion percentage (0.0-1.0) |
| `isOverdue` | bool | Computed | Whether task is past deadline |
| `daysLeft` | int | Computed | Days until deadline (negative if overdue) |

**Validation Rules:**
- Title: required, 1-200 characters
- Description: max 2000 characters
- Priority: 1-5 (inclusive)
- Status: must be valid TaskStatus
- Tags: max 10 tags, each max 50 characters

**Example JSON:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Complete Project Proposal",
  "description": "Draft and finalize Q1 project proposal for review",
  "priority": 4,
  "deadline": {
    "type": "short",
    "date": "2026-02-12T23:59:59Z"
  },
  "category": "work",
  "status": "in_progress",
  "objectives": [
    {
      "id": "obj-1",
      "text": "Research requirements",
      "completed": true,
      "order": 0,
      "createdAt": "2026-02-09T10:00:00Z"
    },
    {
      "id": "obj-2",
      "text": "Draft outline",
      "completed": true,
      "order": 1,
      "createdAt": "2026-02-09T10:00:00Z"
    },
    {
      "id": "obj-3",
      "text": "Write full proposal",
      "completed": false,
      "order": 2,
      "createdAt": "2026-02-09T10:00:00Z"
    }
  ],
  "notes": "Check with team lead before submitting",
  "reward": 50,
  "tags": ["important", "q1", "deadline"],
  "order": 0,
  "createdAt": "2026-02-09T10:00:00Z",
  "updatedAt": "2026-02-09T15:30:00Z",
  "completedAt": null,
  "progress": 0.67,
  "isOverdue": false,
  "daysLeft": 3
}
```

---

### TaskStatus

Enum representing the current state of a task.

```go
type TaskStatus string

const (
    StatusActive     TaskStatus = "active"      // Task is active, not started
    StatusInProgress TaskStatus = "in_progress" // Task is being worked on
    StatusComplete   TaskStatus = "complete"    // Task is finished
    StatusFailed     TaskStatus = "failed"      // Task could not be completed
    StatusArchived   TaskStatus = "archived"    // Task is archived
)
```

**Status Flow:**
```
active → in_progress → complete
  ↓           ↓
archived    failed
```

**Description:**
- `active`: Task exists but work hasn't started
- `in_progress`: Actively being worked on
- `complete`: Successfully finished
- `failed`: Could not complete (missed deadline, blocked, etc.)
- `archived`: Removed from active view but not deleted

---

### Deadline

Deadline information for a task.

```go
type Deadline struct {
    Type string     `json:"type"`  // "short", "medium", "long", "none"
    Date *time.Time `json:"date"`  // Actual deadline date (optional)
}
```

**Deadline Types:**

| Type | Description | Default Days | Color Code |
|------|-------------|--------------|------------|
| `short` | Urgent deadline | 1-3 days | Red (#D94A4A) |
| `medium` | Moderate deadline | 4-7 days | Orange (#E8A958) |
| `long` | Relaxed deadline | 8+ days | Green (#6BA573) |
| `none` | No deadline | - | Gray (#A8A8A8) |

**Examples:**
```json
// Type-only deadline
{
  "type": "short",
  "date": null
}

// Specific date deadline
{
  "type": "short",
  "date": "2026-02-12T23:59:59Z"
}

// No deadline
{
  "type": "none",
  "date": null
}
```

**Calculation Logic:**
- If `date` is provided, calculate `daysLeft` from current date
- If only `type` is provided, use default threshold
- Type determines visual styling regardless of actual date

---

### Objective

Sub-task or step within a task.

```go
type Objective struct {
    ID        string    `json:"id"`
    Text      string    `json:"text" validate:"required"`
    Completed bool      `json:"completed"`
    Order     int       `json:"order"`
    CreatedAt time.Time `json:"createdAt"`
}
```

**Field Descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `text` | string | Yes | Objective description |
| `completed` | bool | Yes | Completion status |
| `order` | int | Yes | Display order (0-based) |
| `createdAt` | timestamp | Yes | Creation timestamp |

**Example:**
```json
{
  "id": "obj-550e8400-e29b-41d4-a716-446655440001",
  "text": "Research requirements and gather data",
  "completed": true,
  "order": 0,
  "createdAt": "2026-02-09T10:00:00Z"
}
```

**Notes:**
- Objectives are always part of a parent task
- Completion toggles with star indicators (★/☆) in UI
- Order determines display sequence
- Progress calculation: `completed objectives / total objectives`

---

### Category

Organizational category for tasks.

```go
type Category struct {
    ID        string    `json:"id"`
    Name      string    `json:"name" validate:"required"`
    Color     string    `json:"color" validate:"hexcolor"`
    Icon      string    `json:"icon"`
    Type      string    `json:"type"`  // "main", "side"
    Order     int       `json:"order"`
    CreatedAt time.Time `json:"createdAt"`
}
```

**Field Descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `name` | string | Yes | Category name |
| `color` | string | Yes | Hex color code (#RRGGBB) |
| `icon` | string | No | Icon identifier |
| `type` | string | Yes | "main" or "side" |
| `order` | int | Yes | Display order |
| `createdAt` | timestamp | Yes | Creation timestamp |

**Category Types:**
- `main`: Main quest category (high priority)
- `side`: Side quest category (optional tasks)

**Example:**
```json
{
  "id": "cat-work",
  "name": "Work",
  "color": "#3A7F8F",
  "icon": "briefcase",
  "type": "main",
  "order": 0,
  "createdAt": "2026-02-09T10:00:00Z"
}
```

**Predefined Categories:**
```json
[
  {
    "id": "work",
    "name": "Work",
    "color": "#3A7F8F",
    "icon": "briefcase",
    "type": "main"
  },
  {
    "id": "personal",
    "name": "Personal",
    "color": "#6BA573",
    "icon": "person",
    "type": "side"
  },
  {
    "id": "learning",
    "name": "Learning",
    "color": "#8B6FB0",
    "icon": "book",
    "type": "side"
  },
  {
    "id": "health",
    "name": "Health",
    "color": "#D9896A",
    "icon": "heart",
    "type": "side"
  }
]
```

---

## Statistics Models

### Stats

Overall application statistics.

```go
type Stats struct {
    TotalTasks      int               `json:"totalTasks"`
    ActiveTasks     int               `json:"activeTasks"`
    CompletedTasks  int               `json:"completedTasks"`
    FailedTasks     int               `json:"failedTasks"`
    TotalRewards    int               `json:"totalRewards"`
    CompletionRate  float64           `json:"completionRate"`
    AverageTime     float64           `json:"averageTimeToComplete"` // hours
    StreakDays      int               `json:"streakDays"`
    CategoryStats   map[string]int    `json:"categoryStats"`
    PriorityStats   map[int]int       `json:"priorityStats"`
}
```

**Example:**
```json
{
  "totalTasks": 150,
  "activeTasks": 45,
  "completedTasks": 100,
  "failedTasks": 5,
  "totalRewards": 5000,
  "completionRate": 0.67,
  "averageTimeToComplete": 48.5,
  "streakDays": 14,
  "categoryStats": {
    "work": 75,
    "personal": 50,
    "learning": 25
  },
  "priorityStats": {
    "1": 10,
    "2": 20,
    "3": 40,
    "4": 50,
    "5": 30
  }
}
```

**Calculations:**
- `completionRate` = completed / (completed + failed)
- `averageTimeToComplete` = average hours from creation to completion
- `streakDays` = consecutive days with at least one completed task

---

### DailyStat

Statistics for a specific day.

```go
type DailyStat struct {
    Date      string `json:"date"`      // YYYY-MM-DD
    Completed int    `json:"completed"`
    Created   int    `json:"created"`
    Failed    int    `json:"failed"`
}
```

**Example:**
```json
{
  "date": "2026-02-09",
  "completed": 5,
  "created": 3,
  "failed": 0
}
```

---

### CategoryStat

Statistics for a specific category.

```go
type CategoryStat struct {
    Total          int     `json:"total"`
    Completed      int     `json:"completed"`
    Active         int     `json:"active"`
    CompletionRate float64 `json:"completionRate"`
}
```

**Example:**
```json
{
  "work": {
    "total": 75,
    "completed": 50,
    "active": 20,
    "completionRate": 0.71
  }
}
```

---

## Configuration Models

### AppSettings

Application-wide settings.

```go
type AppSettings struct {
    Theme              string             `json:"theme"`
    DeadlineThresholds DeadlineThresholds `json:"deadlineThresholds"`
    DefaultReward      int                `json:"defaultReward"`
    PointsEnabled      bool               `json:"pointsEnabled"`
    Notifications      bool               `json:"notifications"`
}

type DeadlineThresholds struct {
    Short  int `json:"short"`  // days
    Medium int `json:"medium"` // days
    Long   int `json:"long"`   // days
}
```

**Example:**
```json
{
  "theme": "trails-journal",
  "deadlineThresholds": {
    "short": 3,
    "medium": 7,
    "long": 14
  },
  "defaultReward": 10,
  "pointsEnabled": true,
  "notifications": true
}
```

---

## Filter & Query Models

### TaskFilter

Query filter for listing tasks.

```go
type TaskFilter struct {
    // Filters
    Status      []TaskStatus
    Priority    []int
    Categories  []string
    Tags        []string
    DeadlineType string
    DateFrom    *time.Time
    DateTo      *time.Time
    IncludeCompleted bool

    // Sorting
    SortBy      string  // field name
    SortOrder   string  // "asc" or "desc"

    // Pagination
    Limit       int
    Offset      int
}
```

**Example:**
```json
{
  "status": ["active", "in_progress"],
  "priority": [4, 5],
  "categories": ["work"],
  "tags": ["urgent"],
  "deadlineType": "short",
  "includeCompleted": false,
  "sortBy": "priority",
  "sortOrder": "desc",
  "limit": 20,
  "offset": 0
}
```

---

## Database Schema (SQLite)

### Tasks Table

```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    priority INTEGER NOT NULL CHECK(priority >= 1 AND priority <= 5),
    deadline_type TEXT,
    deadline_date DATETIME,
    category TEXT,
    status TEXT NOT NULL,
    notes TEXT,
    reward INTEGER DEFAULT 0,
    order_index INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    completed_at DATETIME,

    FOREIGN KEY (category) REFERENCES categories(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority DESC);
CREATE INDEX idx_tasks_deadline ON tasks(deadline_date);
CREATE INDEX idx_tasks_category ON tasks(category);
CREATE INDEX idx_tasks_created ON tasks(created_at DESC);
CREATE INDEX idx_tasks_order ON tasks(order_index);

-- Full-text search
CREATE VIRTUAL TABLE tasks_fts USING fts5(
    id UNINDEXED,
    title,
    description,
    content=tasks,
    content_rowid=rowid
);
```

### Objectives Table

```sql
CREATE TABLE objectives (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    text TEXT NOT NULL,
    completed BOOLEAN DEFAULT 0,
    order_index INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,

    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_objectives_task ON objectives(task_id);
CREATE INDEX idx_objectives_order ON objectives(task_id, order_index);
```

### Tags Table

```sql
CREATE TABLE tags (
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,

    PRIMARY KEY (task_id, tag),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX idx_tags_tag ON tags(tag);
```

### Categories Table

```sql
CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL,
    icon TEXT,
    type TEXT NOT NULL CHECK(type IN ('main', 'side')),
    order_index INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL
);

CREATE INDEX idx_categories_order ON categories(order_index);
```

### Settings Table

```sql
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

---

## JSON Storage Format

When using JSON storage, data is stored as:

```json
{
  "version": "1.0",
  "lastModified": "2026-02-09T20:00:00Z",
  "tasks": {
    "task-id-1": { /* task object */ },
    "task-id-2": { /* task object */ }
  },
  "categories": {
    "cat-id-1": { /* category object */ },
    "cat-id-2": { /* category object */ }
  },
  "settings": {
    /* settings object */
  }
}
```

---

## Data Validation

### Task Validation

```go
func ValidateTask(task *Task) error {
    if task.Title == "" {
        return errors.New("title is required")
    }
    if len(task.Title) > 200 {
        return errors.New("title too long (max 200 chars)")
    }
    if len(task.Description) > 2000 {
        return errors.New("description too long (max 2000 chars)")
    }
    if task.Priority < 1 || task.Priority > 5 {
        return errors.New("priority must be between 1 and 5")
    }
    if !isValidStatus(task.Status) {
        return errors.New("invalid status")
    }
    return nil
}
```

### Category Validation

```go
func ValidateCategory(cat *Category) error {
    if cat.Name == "" {
        return errors.New("name is required")
    }
    if !isValidHexColor(cat.Color) {
        return errors.New("invalid color format")
    }
    if cat.Type != "main" && cat.Type != "side" {
        return errors.New("type must be 'main' or 'side'")
    }
    return nil
}
```

---

## Migration Strategy

### Version 1.0 (Initial)
- Tasks, objectives, tags, categories
- Basic statistics

### Future Versions
- v1.1: Add task templates
- v1.2: Add recurring tasks
- v1.3: Add task dependencies
- v1.4: Add attachments/files

**Migration Process:**
1. Check current schema version
2. Apply migrations sequentially
3. Update version number
4. Create backup before migration

---

## Summary

The data models provide:
- **Tasks** with priorities, deadlines, and objectives
- **Categories** for organization (main/side quests)
- **Statistics** for tracking progress
- **Flexible storage** (SQLite with indexes or JSON)
- **Validation** for data integrity
- **Future extensibility** through clean schema design

All models are designed to support the quest journal aesthetic while maintaining clean separation between storage implementations.
