# API Specification

Complete REST API documentation for Quest Todo backend service.

## Base URL

```
http://localhost:3000/api/v1
```

## Response Format

All responses follow this structure:

**Success Response:**
```json
{
  "success": true,
  "data": { /* response data */ }
}
```

**Success Response with Metadata:**
```json
{
  "success": true,
  "data": [ /* array of items */ ],
  "meta": {
    "total": 100,
    "limit": 20,
    "offset": 0,
    "hasMore": true
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "fields": {
      "fieldName": "Field-specific error"
    }
  }
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Invalid input data |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Duplicate or constraint violation |
| `INTERNAL_ERROR` | Server error |
| `INVALID_REQUEST` | Malformed request |

---

## Tasks

### List Tasks

**GET** `/tasks`

Retrieve tasks with optional filtering, sorting, and pagination.

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `status` | string | Comma-separated status values | `active,in_progress` |
| `priority` | string | Comma-separated priority levels (1-5) | `4,5` |
| `category` | string | Comma-separated category IDs | `work,personal` |
| `tags` | string | Comma-separated tags | `urgent,important` |
| `deadline` | string | Deadline type filter | `short` |
| `include_completed` | boolean | Include completed tasks | `true` |
| `sort` | string | Sort field | `priority_desc`, `deadline_asc` |
| `limit` | integer | Max results (default: 20) | `50` |
| `offset` | integer | Skip N results (default: 0) | `20` |

**Sort Options:**
- `priority_asc` / `priority_desc`
- `deadline_asc` / `deadline_desc`
- `created_asc` / `created_desc`
- `updated_asc` / `updated_desc`
- `title_asc` / `title_desc`

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "task-123",
      "title": "Complete Project Proposal",
      "description": "Draft and finalize Q1 project proposal",
      "priority": 4,
      "deadline": {
        "type": "short",
        "date": "2026-02-12T23:59:59Z"
      },
      "category": "work",
      "status": "active",
      "objectives": [
        {
          "id": "obj-1",
          "text": "Research requirements",
          "completed": true,
          "order": 0,
          "createdAt": "2026-02-09T10:00:00Z"
        }
      ],
      "notes": "Check with team lead",
      "reward": 50,
      "tags": ["important", "q1"],
      "order": 0,
      "progress": 0.5,
      "isOverdue": false,
      "daysLeft": 3,
      "createdAt": "2026-02-09T10:00:00Z",
      "updatedAt": "2026-02-09T15:00:00Z",
      "completedAt": null
    }
  ],
  "meta": {
    "total": 45,
    "limit": 20,
    "offset": 0,
    "hasMore": true
  }
}
```

---

### Get Task

**GET** `/tasks/:id`

Retrieve a single task by ID.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "task-123",
    "title": "Complete Project Proposal",
    /* ... full task object ... */
  }
}
```

**Error (404):**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Task not found"
  }
}
```

---

### Create Task

**POST** `/tasks`

Create a new task.

**Request Body:**
```json
{
  "title": "Complete Project Proposal",
  "description": "Draft and finalize Q1 project proposal for review",
  "priority": 4,
  "deadline": {
    "type": "short",
    "date": "2026-02-12T23:59:59Z"
  },
  "category": "work",
  "objectives": [
    { "text": "Research requirements" },
    { "text": "Draft outline" },
    { "text": "Write full proposal" }
  ],
  "notes": "Check with team lead before submitting",
  "reward": 50,
  "tags": ["important", "q1"]
}
```

**Required Fields:**
- `title` (1-200 characters)
- `priority` (1-5)

**Optional Fields:**
- `description` (max 2000 characters)
- `deadline`
- `category`
- `objectives` (array)
- `notes`
- `reward` (integer)
- `tags` (array of strings)

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "task-uuid-generated",
    "title": "Complete Project Proposal",
    "status": "active",
    "createdAt": "2026-02-09T10:00:00Z",
    "updatedAt": "2026-02-09T10:00:00Z",
    /* ... full task object ... */
  }
}
```

**Error (400):**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "priority": "Priority must be between 1 and 5",
      "title": "Title is required"
    }
  }
}
```

---

### Update Task

**PUT** `/tasks/:id`

Fully update a task (replaces all fields).

**Request Body:** Same as Create Task

**Response:**
```json
{
  "success": true,
  "data": {
    /* ... updated task object ... */
  }
}
```

---

### Partial Update Task

**PATCH** `/tasks/:id`

Partially update a task (only specified fields).

**Request Body:**
```json
{
  "title": "Updated Title",
  "priority": 5
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    /* ... updated task object ... */
  }
}
```

---

### Delete Task

**DELETE** `/tasks/:id`

Delete a task (soft delete by default).

**Query Parameters:**
- `permanent=true` - Permanently delete (cannot be recovered)

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Task deleted successfully"
  }
}
```

---

### Update Task Status

**PATCH** `/tasks/:id/status`

Update only the task status.

**Request Body:**
```json
{
  "status": "complete"
}
```

**Valid Status Values:**
- `active`
- `in_progress`
- `complete`
- `failed`
- `archived`

**Response:**
```json
{
  "success": true,
  "data": {
    /* ... updated task object ... */
  }
}
```

---

### Search Tasks

**GET** `/tasks/search`

Full-text search across task titles and descriptions.

**Query Parameters:**
- `q` (required) - Search query

**Example:**
```
GET /tasks/search?q=project+proposal
```

**Response:**
```json
{
  "success": true,
  "data": [
    /* ... matching tasks ... */
  ]
}
```

---

### Get Upcoming Tasks

**GET** `/tasks/upcoming`

Get tasks with approaching deadlines (next 7 days).

**Response:**
```json
{
  "success": true,
  "data": [
    /* ... tasks with near deadlines ... */
  ]
}
```

---

### Get Overdue Tasks

**GET** `/tasks/overdue`

Get tasks past their deadline.

**Response:**
```json
{
  "success": true,
  "data": [
    /* ... overdue tasks ... */
  ]
}
```

---

## Objectives

### Add Objective

**POST** `/tasks/:id/objectives`

Add a new objective to a task.

**Request Body:**
```json
{
  "text": "Review with team lead"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "obj-new",
    "text": "Review with team lead",
    "completed": false,
    "order": 3,
    "createdAt": "2026-02-09T16:00:00Z"
  }
}
```

---

### Update Objective

**PATCH** `/tasks/:id/objectives/:objId`

Update an objective (toggle completion or change text).

**Request Body:**
```json
{
  "completed": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    /* ... updated objective ... */
  }
}
```

---

### Delete Objective

**DELETE** `/tasks/:id/objectives/:objId`

Remove an objective from a task.

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Objective deleted successfully"
  }
}
```

---

## Bulk Operations

### Create Multiple Tasks

**POST** `/tasks/bulk`

Create multiple tasks in one request.

**Request Body:**
```json
{
  "tasks": [
    {
      "title": "Task 1",
      "priority": 3
    },
    {
      "title": "Task 2",
      "priority": 4
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "created": 2,
    "tasks": [
      /* ... created task objects ... */
    ]
  }
}
```

---

### Update Multiple Tasks

**PATCH** `/tasks/bulk`

Update multiple tasks at once.

**Request Body:**
```json
{
  "ids": ["task-1", "task-2", "task-3"],
  "updates": {
    "status": "archived"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "updated": 3
  }
}
```

---

### Delete Multiple Tasks

**DELETE** `/tasks/bulk`

Delete multiple tasks at once.

**Request Body:**
```json
{
  "ids": ["task-1", "task-2", "task-3"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "deleted": 3
  }
}
```

---

### Reorder Tasks

**POST** `/tasks/reorder`

Change the display order of tasks.

**Request Body:**
```json
{
  "ids": ["task-3", "task-1", "task-2"]
}
```

The order in the array determines the new order (first = order 0).

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Tasks reordered successfully"
  }
}
```

---

## Categories

### List Categories

**GET** `/categories`

Retrieve all categories.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "cat-1",
      "name": "Work",
      "color": "#3A7F8F",
      "icon": "briefcase",
      "type": "main",
      "order": 0,
      "createdAt": "2026-02-09T10:00:00Z"
    }
  ]
}
```

---

### Get Category

**GET** `/categories/:id`

Retrieve a single category.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "cat-1",
    "name": "Work",
    /* ... */
  }
}
```

---

### Create Category

**POST** `/categories`

Create a new category.

**Request Body:**
```json
{
  "name": "Personal",
  "color": "#6BA573",
  "icon": "person",
  "type": "side"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "cat-new",
    "name": "Personal",
    /* ... */
  }
}
```

---

### Update Category

**PUT** `/categories/:id`

Update a category.

**Request Body:**
```json
{
  "name": "Personal Projects",
  "color": "#5A9F67"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    /* ... updated category ... */
  }
}
```

---

### Delete Category

**DELETE** `/categories/:id`

Delete a category.

**Note:** Tasks in this category will have their category field set to null.

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Category deleted successfully"
  }
}
```

---

## Statistics

### Get Overall Stats

**GET** `/stats`

Get overall statistics.

**Response:**
```json
{
  "success": true,
  "data": {
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
}
```

---

### Get Daily Stats

**GET** `/stats/daily`

Get daily completion statistics.

**Query Parameters:**
- `from` - Start date (YYYY-MM-DD)
- `to` - End date (YYYY-MM-DD)

**Example:**
```
GET /stats/daily?from=2026-02-01&to=2026-02-09
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "date": "2026-02-09",
      "completed": 5,
      "created": 3,
      "failed": 0
    },
    {
      "date": "2026-02-08",
      "completed": 7,
      "created": 5,
      "failed": 1
    }
  ]
}
```

---

### Get Weekly Stats

**GET** `/stats/weekly`

Get weekly trend statistics.

**Response:**
```json
{
  "success": true,
  "data": {
    "thisWeek": {
      "completed": 25,
      "created": 20
    },
    "lastWeek": {
      "completed": 30,
      "created": 28
    },
    "trend": "down"
  }
}
```

---

### Get Category Stats

**GET** `/stats/categories`

Get statistics broken down by category.

**Response:**
```json
{
  "success": true,
  "data": {
    "work": {
      "total": 75,
      "completed": 50,
      "active": 20,
      "completionRate": 0.71
    },
    "personal": {
      "total": 50,
      "completed": 35,
      "active": 10,
      "completionRate": 0.78
    }
  }
}
```

---

### Get Productivity Stats

**GET** `/stats/productivity`

Get productivity metrics.

**Response:**
```json
{
  "success": true,
  "data": {
    "tasksPerDay": 3.5,
    "averageCompletionTime": 48.5,
    "mostProductiveDay": "Tuesday",
    "mostProductiveHour": 10,
    "streakDays": 14,
    "longestStreak": 30
  }
}
```

---

## History

### Get History

**GET** `/history`

Get completed tasks (history view).

**Query Parameters:**
- `from` - Start date
- `to` - End date
- `limit` - Results per page
- `offset` - Pagination offset

**Response:**
```json
{
  "success": true,
  "data": [
    {
      /* ... completed task with completedAt timestamp ... */
    }
  ],
  "meta": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

## Settings

### Get Settings

**GET** `/settings`

Retrieve application settings.

**Response:**
```json
{
  "success": true,
  "data": {
    "theme": "trails-journal",
    "deadlineThresholds": {
      "short": 3,
      "medium": 7,
      "long": 14
    },
    "pointsEnabled": true,
    "notifications": true,
    "defaultReward": 10
  }
}
```

---

### Update Settings

**PUT** `/settings`

Update application settings.

**Request Body:**
```json
{
  "deadlineThresholds": {
    "short": 2,
    "medium": 5,
    "long": 10
  },
  "pointsEnabled": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    /* ... updated settings ... */
  }
}
```

---

## Export/Import

### Export Data

**GET** `/export`

Export all data as JSON.

**Response:**
```json
{
  "success": true,
  "data": {
    "version": "1.0",
    "exportedAt": "2026-02-09T20:00:00Z",
    "tasks": [ /* ... */ ],
    "categories": [ /* ... */ ],
    "settings": { /* ... */ }
  }
}
```

---

### Import Data

**POST** `/import`

Import data from JSON export.

**Request Body:** Same format as export response

**Response:**
```json
{
  "success": true,
  "data": {
    "imported": {
      "tasks": 100,
      "categories": 5
    }
  }
}
```

---

### Create Backup

**GET** `/backup`

Create a backup of the database.

**Response:**
Downloads a backup file (SQLite database or JSON file).

---

### Restore Backup

**POST** `/restore`

Restore from a backup file.

**Request:** Multipart form data with backup file

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Backup restored successfully"
  }
}
```

---

## Health & Status

### Health Check

**GET** `/health`

Check if the service is healthy.

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600,
  "storage": "ok",
  "timestamp": "2026-02-09T20:00:00Z"
}
```

---

### Version Info

**GET** `/version`

Get API version information.

**Response:**
```json
{
  "version": "1.0.0",
  "apiVersion": "v1",
  "buildDate": "2026-02-01",
  "goVersion": "1.22"
}
```

---

### Ping

**GET** `/ping`

Simple ping endpoint.

**Response:**
```json
{
  "message": "pong"
}
```

---

## Rate Limiting

Currently not implemented (local-only service).

Future consideration for networked deployments:
- 100 requests per minute per client
- Burst allowance: 20 requests
- Header: `X-RateLimit-Remaining`

## Authentication

Currently not required (local-only service).

Future consideration for networked deployments:
- JWT-based authentication
- Header: `Authorization: Bearer <token>`

## CORS

CORS is enabled for local development:
- Allowed origins: `http://localhost:*`
- Allowed methods: GET, POST, PUT, PATCH, DELETE
- Allowed headers: Content-Type, Authorization

## Content Type

All requests and responses use:
```
Content-Type: application/json
```

## Date/Time Format

All timestamps use ISO 8601 format with UTC timezone:
```
2026-02-09T20:00:00Z
```

## Pagination

Default pagination limits:
- Default `limit`: 20
- Maximum `limit`: 100
- Default `offset`: 0

## Summary

This API provides full CRUD operations for tasks, categories, and settings, along with advanced features like:
- Filtering and searching
- Bulk operations
- Statistics and analytics
- Export/Import capabilities
- Health monitoring

All endpoints follow RESTful conventions and return consistent JSON responses.
