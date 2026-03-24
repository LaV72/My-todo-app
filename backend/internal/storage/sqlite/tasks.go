package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
)

// CreateTask creates a new task in the database
func (s *SQLiteStorage) CreateTask(ctx context.Context, task *models.Task) error {
	// Start a transaction (for atomic insert of task + objectives + tags)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if we don't commit

	// Insert task
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tasks (
			id, title, description, priority, deadline_type, deadline_date,
			category, status, notes, reward, order_index, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		task.ID, task.Title, task.Description, task.Priority,
		task.Deadline.Type, task.Deadline.Date,
		task.Category, task.Status, task.Notes, task.Reward, task.Order,
		task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	// Insert objectives
	for _, obj := range task.Objectives {
		err = s.insertObjective(ctx, tx, &obj)
		if err != nil {
			return fmt.Errorf("insert objective: %w", err)
		}
	}

	// Insert tags
	for _, tag := range task.Tags {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO task_tags (task_id, tag) VALUES (?, ?)
		`, task.ID, tag)
		if err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetTask retrieves a single task by ID
func (s *SQLiteStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	task := &models.Task{}

	// Query task
	err := s.db.QueryRowContext(ctx, `
		SELECT
			id, title, description, priority, deadline_type, deadline_date,
			category, status, notes, reward, order_index,
			created_at, updated_at, completed_at
		FROM tasks
		WHERE id = ?
	`, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Priority,
		&task.Deadline.Type, &task.Deadline.Date,
		&task.Category, &task.Status, &task.Notes, &task.Reward, &task.Order,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: task %s", storage.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("query task: %w", err)
	}

	// Load objectives
	task.Objectives, err = s.loadObjectives(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("load objectives: %w", err)
	}

	// Load tags
	task.Tags, err = s.loadTags(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("load tags: %w", err)
	}

	// Calculate computed fields
	s.calculateComputedFields(task)

	return task, nil
}

// UpdateTask updates an existing task
func (s *SQLiteStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update task
	result, err := tx.ExecContext(ctx, `
		UPDATE tasks SET
			title = ?, description = ?, priority = ?,
			deadline_type = ?, deadline_date = ?,
			category = ?, status = ?, notes = ?, reward = ?,
			order_index = ?, updated_at = ?, completed_at = ?
		WHERE id = ?
	`,
		task.Title, task.Description, task.Priority,
		task.Deadline.Type, task.Deadline.Date,
		task.Category, task.Status, task.Notes, task.Reward,
		task.Order, task.UpdatedAt, task.CompletedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// Check if task exists
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: task %s", storage.ErrNotFound, task.ID)
	}

	// Delete and re-insert objectives (simpler than diffing)
	_, err = tx.ExecContext(ctx, `DELETE FROM objectives WHERE task_id = ?`, task.ID)
	if err != nil {
		return fmt.Errorf("delete old objectives: %w", err)
	}

	for _, obj := range task.Objectives {
		err = s.insertObjective(ctx, tx, &obj)
		if err != nil {
			return fmt.Errorf("insert objective: %w", err)
		}
	}

	// Delete and re-insert tags
	_, err = tx.ExecContext(ctx, `DELETE FROM task_tags WHERE task_id = ?`, task.ID)
	if err != nil {
		return fmt.Errorf("delete old tags: %w", err)
	}

	for _, tag := range task.Tags {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO task_tags (task_id, tag) VALUES (?, ?)
		`, task.ID, tag)
		if err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteTask deletes a task by ID
func (s *SQLiteStorage) DeleteTask(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: task %s", storage.ErrNotFound, id)
	}

	// Objectives and tags are deleted automatically via CASCADE

	return nil
}

// ListTasks retrieves tasks with filtering, sorting, and pagination
func (s *SQLiteStorage) ListTasks(ctx context.Context, filter models.TaskFilter) ([]*models.Task, error) {
	// Build query dynamically based on filters
	query := `
		SELECT
			id, title, description, priority, deadline_type, deadline_date,
			category, status, notes, reward, order_index,
			created_at, updated_at, completed_at
		FROM tasks
		WHERE 1=1
	`
	args := []interface{}{}

	// Add filters
	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			placeholders[i] = "?"
			args = append(args, status)
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filter.Priority) > 0 {
		placeholders := make([]string, len(filter.Priority))
		for i, priority := range filter.Priority {
			placeholders[i] = "?"
			args = append(args, priority)
		}
		query += fmt.Sprintf(" AND priority IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filter.Categories) > 0 {
		placeholders := make([]string, len(filter.Categories))
		for i, category := range filter.Categories {
			placeholders[i] = "?"
			args = append(args, category)
		}
		query += fmt.Sprintf(" AND category IN (%s)", strings.Join(placeholders, ","))
	}

	if filter.DeadlineType != "" {
		query += " AND deadline_type = ?"
		args = append(args, filter.DeadlineType)
	}

	if filter.DateFrom != nil {
		query += " AND created_at >= ?"
		args = append(args, filter.DateFrom)
	}

	if filter.DateTo != nil {
		query += " AND created_at <= ?"
		args = append(args, filter.DateTo)
	}

	if !filter.IncludeCompleted {
		query += " AND status != ?"
		args = append(args, models.StatusComplete)
	}

	// Add sorting with SQL injection protection
	if filter.SortBy != "" {
		// Whitelist of allowed sort columns
		allowedColumns := map[string]bool{
			"priority":      true,
			"created_at":    true,
			"updated_at":    true,
			"title":         true,
			"status":        true,
			"deadline_date": true,
			"order_index":   true,
		}

		// Validate sort column
		if !allowedColumns[filter.SortBy] {
			filter.SortBy = "created_at" // safe default
		}

		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	} else {
		// Default sort
		query += " ORDER BY order_index ASC, created_at DESC"
	}

	// Add pagination
	if filter.Limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, filter.Limit, filter.Offset)
	}

	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	// Scan results
	tasks := []*models.Task{}
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Priority,
			&task.Deadline.Type, &task.Deadline.Date,
			&task.Category, &task.Status, &task.Notes, &task.Reward, &task.Order,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Batch load objectives and tags to avoid N+1 queries
	if len(tasks) > 0 {
		// Collect all task IDs
		taskIDs := make([]string, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}

		// Load all objectives in one query
		objectivesMap, err := s.loadObjectivesBatch(ctx, taskIDs)
		if err != nil {
			return nil, fmt.Errorf("batch load objectives: %w", err)
		}

		// Load all tags in one query
		tagsMap, err := s.loadTagsBatch(ctx, taskIDs)
		if err != nil {
			return nil, fmt.Errorf("batch load tags: %w", err)
		}

		// Assign objectives and tags to tasks
		for _, task := range tasks {
			task.Objectives = objectivesMap[task.ID]
			task.Tags = tagsMap[task.ID]

			// Ensure empty slices instead of nil
			if task.Objectives == nil {
				task.Objectives = []models.Objective{}
			}
			if task.Tags == nil {
				task.Tags = []string{}
			}

			// Calculate computed fields (Progress, IsOverdue, DaysLeft)
			s.calculateComputedFields(task)
		}
	}

	return tasks, nil
}

// CountTasks counts tasks matching the filter
func (s *SQLiteStorage) CountTasks(ctx context.Context, filter models.TaskFilter) (int, error) {
	query := "SELECT COUNT(*) FROM tasks WHERE 1=1"
	args := []interface{}{}

	// Add same filters as ListTasks (without sorting/pagination)
	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			placeholders[i] = "?"
			args = append(args, status)
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filter.Priority) > 0 {
		placeholders := make([]string, len(filter.Priority))
		for i, priority := range filter.Priority {
			placeholders[i] = "?"
			args = append(args, priority)
		}
		query += fmt.Sprintf(" AND priority IN (%s)", strings.Join(placeholders, ","))
	}

	if !filter.IncludeCompleted {
		query += " AND status != ?"
		args = append(args, models.StatusComplete)
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}

	return count, nil
}

// SearchTasks performs full-text search on tasks
func (s *SQLiteStorage) SearchTasks(ctx context.Context, query string) ([]*models.Task, error) {
	// Simple LIKE search (can be upgraded to FTS5 later)
	searchPattern := "%" + query + "%"

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id, title, description, priority, deadline_type, deadline_date,
			category, status, notes, reward, order_index,
			created_at, updated_at, completed_at
		FROM tasks
		WHERE title LIKE ? OR description LIKE ?
		ORDER BY created_at DESC
		LIMIT 50
	`, searchPattern, searchPattern)

	if err != nil {
		return nil, fmt.Errorf("search tasks: %w", err)
	}
	defer rows.Close()

	tasks := []*models.Task{}
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Priority,
			&task.Deadline.Type, &task.Deadline.Date,
			&task.Category, &task.Status, &task.Notes, &task.Reward, &task.Order,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	// Batch load objectives and tags to avoid N+1 queries
	if len(tasks) > 0 {
		// Collect all task IDs
		taskIDs := make([]string, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}

		// Load all objectives in one query
		objectivesMap, err := s.loadObjectivesBatch(ctx, taskIDs)
		if err != nil {
			return nil, fmt.Errorf("batch load objectives: %w", err)
		}

		// Load all tags in one query
		tagsMap, err := s.loadTagsBatch(ctx, taskIDs)
		if err != nil {
			return nil, fmt.Errorf("batch load tags: %w", err)
		}

		// Assign objectives and tags to tasks
		for _, task := range tasks {
			task.Objectives = objectivesMap[task.ID]
			task.Tags = tagsMap[task.ID]

			// Ensure empty slices instead of nil
			if task.Objectives == nil {
				task.Objectives = []models.Objective{}
			}
			if task.Tags == nil {
				task.Tags = []string{}
			}

			// Calculate computed fields
			s.calculateComputedFields(task)
		}
	}

	return tasks, nil
}

// UpdateTaskStatus updates only the status of a task
func (s *SQLiteStorage) UpdateTaskStatus(ctx context.Context, id string, status models.TaskStatus) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, status, id)

	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: task %s", storage.ErrNotFound, id)
	}

	return nil
}

// ReorderTasks updates the order_index for multiple tasks
func (s *SQLiteStorage) ReorderTasks(ctx context.Context, ids []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update order_index based on position in array
	var totalUpdated int64
	for i, id := range ids {
		result, err := tx.ExecContext(ctx, `
			UPDATE tasks SET order_index = ? WHERE id = ?
		`, i, id)
		if err != nil {
			return fmt.Errorf("update task order: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("get rows affected: %w", err)
		}
		totalUpdated += rowsAffected
	}

	// Verify at least some tasks were reordered
	if totalUpdated == 0 {
		return fmt.Errorf("%w: no tasks found with provided IDs", storage.ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// CreateTasksBulk creates multiple tasks in a single transaction
func (s *SQLiteStorage) CreateTasksBulk(ctx context.Context, tasks []*models.Task) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, task := range tasks {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO tasks (
				id, title, description, priority, deadline_type, deadline_date,
				category, status, notes, reward, order_index, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			task.ID, task.Title, task.Description, task.Priority,
			task.Deadline.Type, task.Deadline.Date,
			task.Category, task.Status, task.Notes, task.Reward, task.Order,
			task.CreatedAt, task.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert task %s: %w", task.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// UpdateTasksBulk updates multiple tasks in a single transaction
func (s *SQLiteStorage) UpdateTasksBulk(ctx context.Context, tasks []*models.Task) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, task := range tasks {
		_, err = tx.ExecContext(ctx, `
			UPDATE tasks SET
				title = ?, description = ?, priority = ?,
				status = ?, updated_at = ?
			WHERE id = ?
		`,
			task.Title, task.Description, task.Priority,
			task.Status, task.UpdatedAt, task.ID,
		)
		if err != nil {
			return fmt.Errorf("update task %s: %w", task.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteTasksBulk deletes multiple tasks in a single transaction
func (s *SQLiteStorage) DeleteTasksBulk(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Use transaction for consistency with other bulk operations
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM tasks WHERE id IN (%s)", strings.Join(placeholders, ","))
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete tasks: %w", err)
	}

	// Verify at least some rows were deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: no tasks found with provided IDs", storage.ErrNotFound)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetMaxOrderIndex returns the maximum order_index value across all tasks
func (s *SQLiteStorage) GetMaxOrderIndex(ctx context.Context) (int, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT MAX(order_index) FROM tasks
	`).Scan(&maxOrder)

	if err != nil {
		return 0, fmt.Errorf("get max order index: %w", err)
	}

	// If no tasks exist, MAX returns NULL, so return -1 (next will be 0)
	if !maxOrder.Valid {
		return -1, nil
	}

	return int(maxOrder.Int64), nil
}

// Helper functions

// insertObjective inserts an objective within a transaction
func (s *SQLiteStorage) insertObjective(ctx context.Context, tx *sql.Tx, obj *models.Objective) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO objectives (id, task_id, text, completed, order_index, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, obj.ID, obj.TaskID, obj.Text, obj.Completed, obj.Order, obj.CreatedAt)
	return err
}

// loadObjectives loads all objectives for a task
func (s *SQLiteStorage) loadObjectives(ctx context.Context, taskID string) ([]models.Objective, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, task_id, text, completed, order_index, created_at
		FROM objectives
		WHERE task_id = ?
		ORDER BY order_index ASC
	`, taskID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	objectives := []models.Objective{}
	for rows.Next() {
		var obj models.Objective
		err := rows.Scan(&obj.ID, &obj.TaskID, &obj.Text, &obj.Completed, &obj.Order, &obj.CreatedAt)
		if err != nil {
			return nil, err
		}
		objectives = append(objectives, obj)
	}

	return objectives, rows.Err()
}

// loadTags loads all tags for a task
func (s *SQLiteStorage) loadTags(ctx context.Context, taskID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag FROM task_tags WHERE task_id = ? ORDER BY tag
	`, taskID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// calculateComputedFields calculates Progress, IsOverdue, and DaysLeft for a task
func (s *SQLiteStorage) calculateComputedFields(task *models.Task) {
	// Calculate progress based on completed objectives
	if len(task.Objectives) > 0 {
		completed := 0
		for _, obj := range task.Objectives {
			if obj.Completed {
				completed++
			}
		}
		task.Progress = float64(completed) / float64(len(task.Objectives)) * 100
	} else {
		task.Progress = 0
	}

	// Calculate IsOverdue and DaysLeft based on deadline
	if task.Deadline.Date != nil {
		now := time.Now()
		task.IsOverdue = task.Deadline.Date.Before(now) && task.Status == "active"

		days := int(time.Until(*task.Deadline.Date).Hours() / 24)
		task.DaysLeft = &days
	} else {
		task.IsOverdue = false
		task.DaysLeft = nil
	}
}

// loadObjectivesBatch loads objectives for multiple tasks in one query
// Returns a map of taskID -> objectives
func (s *SQLiteStorage) loadObjectivesBatch(ctx context.Context, taskIDs []string) (map[string][]models.Objective, error) {
	if len(taskIDs) == 0 {
		return map[string][]models.Objective{}, nil
	}

	// Build IN clause with placeholders
	placeholders := make([]string, len(taskIDs))
	args := make([]interface{}, len(taskIDs))
	for i, id := range taskIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, task_id, text, completed, order_index, created_at
		FROM objectives
		WHERE task_id IN (%s)
		ORDER BY task_id, order_index ASC
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map of task_id -> objectives
	result := make(map[string][]models.Objective)
	for rows.Next() {
		var obj models.Objective
		if err := rows.Scan(&obj.ID, &obj.TaskID, &obj.Text, &obj.Completed, &obj.Order, &obj.CreatedAt); err != nil {
			return nil, err
		}
		result[obj.TaskID] = append(result[obj.TaskID], obj)
	}

	return result, rows.Err()
}

// loadTagsBatch loads tags for multiple tasks in one query
// Returns a map of taskID -> tags
func (s *SQLiteStorage) loadTagsBatch(ctx context.Context, taskIDs []string) (map[string][]string, error) {
	if len(taskIDs) == 0 {
		return map[string][]string{}, nil
	}

	// Build IN clause with placeholders
	placeholders := make([]string, len(taskIDs))
	args := make([]interface{}, len(taskIDs))
	for i, id := range taskIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT task_id, tag
		FROM task_tags
		WHERE task_id IN (%s)
		ORDER BY task_id, tag
	`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map of task_id -> tags
	result := make(map[string][]string)
	for rows.Next() {
		var taskID, tag string
		if err := rows.Scan(&taskID, &tag); err != nil {
			return nil, err
		}
		result[taskID] = append(result[taskID], tag)
	}

	return result, rows.Err()
}
