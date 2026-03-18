# Dynamic Task Modification - Discovering Work As You Go

One of the most realistic aspects of task management is that **you often don't know all the subtasks until you start working**. The Quest Todo API fully supports this workflow!

## Real-World Scenario

You start with a vague task like "Fix Production Bug" but discover what needs to be done as you investigate:

1. **Start**: Create high-level task with no objectives
2. **Investigate**: Add first objective "Check logs"
3. **Discover**: Add more objectives as root causes are found
4. **Refine**: Update objectives when requirements change
5. **Simplify**: Delete objectives that become unnecessary
6. **Clarify**: Update task title/description as understanding evolves

## API Operations

### 1. Create Task Without Objectives

Start with a high-level task:

```bash
POST /api/tasks
{
  "title": "Fix Production Database Performance",
  "description": "Users reporting slow queries",
  "priority": 5
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "task-123",
    "title": "Fix Production Database Performance",
    "objectives": []  // Empty - we'll add them later!
  }
}
```

### 2. Add Objectives (As You Discover Them)

After investigation, add your first subtask:

```bash
POST /api/tasks/task-123/objectives
{
  "text": "Check database logs for slow queries",
  "order": 1
}
```

Add more as you discover them:

```bash
POST /api/tasks/task-123/objectives
{"text": "Add index on categories.name", "order": 2}

POST /api/tasks/task-123/objectives
{"text": "Optimize user search query", "order": 3}
```

### 3. Update Objectives (When Plans Change)

Realized the user search query is fine? Update the objective:

```bash
PUT /api/objectives/obj-456
{
  "text": "Add index on categories.name for faster filtering"
}
```

### 4. Complete Objectives (As You Work)

Mark objectives complete with a simple toggle:

```bash
POST /api/objectives/obj-123/toggle
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "obj-123",
    "text": "Check database logs for slow queries",
    "completed": true
  }
}
```

### 5. Delete Objectives (Remove Unnecessary Work)

Found out an index already exists? Delete the objective:

```bash
DELETE /api/objectives/obj-789
```

### 6. Update Task Itself (Refine As You Learn)

As your understanding evolves, update the task:

```bash
PUT /api/tasks/task-123
{
  "title": "Optimize Database Indexes for Performance",
  "description": "Added missing indexes after profiling. Categories table was the bottleneck.",
  "priority": 5
}
```

### 7. Track Progress

Get the current state anytime:

```bash
GET /api/tasks/task-123
```

Response shows all objectives and their status:
```json
{
  "success": true,
  "data": {
    "title": "Optimize Database Indexes for Performance",
    "objectives": [
      {
        "text": "Check database logs",
        "completed": true
      },
      {
        "text": "Add index on categories.name",
        "completed": false
      },
      {
        "text": "Test query performance",
        "completed": false
      }
    ],
    "progress": 33
  }
}
```

## Complete Workflow Example

Here's a real workflow showing task evolution:

### Initial State
```
Task: "Fix Production Bug"
Objectives: (none)
```

### After Investigation
```
Task: "Fix Production Bug"
Objectives:
  1. ⏳ Check server logs
```

### After Finding Root Cause
```
Task: "Fix Production Bug"
Objectives:
  1. ✅ Check server logs
  2. ⏳ Optimize database query
  3. ⏳ Add missing index
  4. ⏳ Deploy fix
```

### After Refinement
```
Task: "Optimize Database Indexes for Performance"
Description: "Added indexes after profiling - categories table was bottleneck"
Objectives:
  1. ✅ Check server logs
  2. ✅ Add index on categories.name
  3. ⏳ Add index on tasks.created_at
  4. ⏳ Test query performance
```

Note: "Optimize database query" was removed (became unnecessary) and task title/description were updated to reflect actual work.

## API Endpoints Summary

| Operation | Endpoint | Method | Description |
|-----------|----------|--------|-------------|
| Add objective | `/api/tasks/{id}/objectives` | POST | Add subtask as discovered |
| Update objective | `/api/objectives/{id}` | PUT | Modify objective text |
| Toggle objective | `/api/objectives/{id}/toggle` | POST | Mark complete/incomplete |
| Delete objective | `/api/objectives/{id}` | DELETE | Remove unnecessary work |
| Update task | `/api/tasks/{id}` | PUT | Refine title/description |
| Get task | `/api/tasks/{id}` | GET | View current state |

## Key Features

✅ **Start Minimal**: Create tasks with no objectives
✅ **Add Dynamically**: Add objectives as you discover work
✅ **Update Freely**: Modify objectives when plans change
✅ **Remove Easily**: Delete objectives that become unnecessary
✅ **Refine Continuously**: Update task details as understanding evolves
✅ **Track Progress**: View completion status at any time
✅ **Auto-Complete**: Task auto-completes when all objectives done (configurable)

## Benefits

1. **Realistic Workflow**: Matches how people actually work
2. **Flexibility**: No need to plan everything upfront
3. **Clarity**: Task evolves to reflect actual work done
4. **Progress Tracking**: See completion percentage update automatically
5. **No Waste**: Remove objectives that become irrelevant

## Configuration

Auto-completion behavior is configurable via environment variables:

```bash
# Auto-complete task when all objectives are done (default: true)
AUTO_COMPLETE_ON_FULL_PROGRESS=true

# Require all objectives complete before manual completion (default: false)
REQUIRE_ALL_OBJECTIVES=false
```

## Example: Frontend Integration

```javascript
// Create task with no objectives
const task = await createTask({
  title: "Fix bug",
  priority: 5
});

// User starts investigating...
await addObjective(task.id, {
  text: "Check logs",
  order: 1
});

// User discovers more work...
await addObjective(task.id, {
  text: "Fix SQL syntax",
  order: 2
});

// User completes first step
await toggleObjective(objective1.id);

// User realizes plans changed
await updateObjective(objective2.id, {
  text: "Optimize query instead of fixing syntax"
});

// User refines task title
await updateTask(task.id, {
  title: "Optimize slow database query"
});
```

## Comparison with Static Planning

### Traditional Approach ❌
```
1. Plan all subtasks upfront
2. Create task with all objectives
3. Realize plans don't match reality
4. Objectives become outdated
5. Task completion is inaccurate
```

### Dynamic Approach ✅
```
1. Create high-level task
2. Start working
3. Add objectives as discovered
4. Update when plans change
5. Delete unnecessary work
6. Task accurately reflects work done
```

## Best Practices

1. **Start Simple**: Create task with just title and priority
2. **Add As You Go**: Add objectives during work, not before
3. **Update Freely**: Don't hesitate to modify objectives
4. **Delete Boldly**: Remove objectives that become unnecessary
5. **Refine Title**: Update task title to reflect final work
6. **Track Progress**: Use completion percentage to gauge work remaining

## Demo Output

See the complete demo in action:

```bash
cd backend
make run

# Follow the demo in docs/DYNAMIC_TASK_WORKFLOW_DEMO.sh
```

Example output:
```
STEP 1: Create task (0 objectives)
STEP 2: Add first objective after investigation
STEP 3: Add 3 more objectives as work progresses
STEP 4: Complete first objective
STEP 5: Update objective when plans change
STEP 6: Delete unnecessary objective
STEP 7: Update task title and description
STEP 8: Add final objective before deploying

Final result: Task evolved from vague "Fix Bug" to precise
"Optimize Database Indexes" with 4 specific subtasks (1 completed)
```

---

**This is how task management should work** - fluid, flexible, and matching real-world workflows! 🎯
