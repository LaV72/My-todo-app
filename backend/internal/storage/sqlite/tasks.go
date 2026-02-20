package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

	// Add sorting
	if filter.SortBy != "" {
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

	// Load objectives and tags for each task
	for _, task := range tasks {
		task.Objectives, err = s.loadObjectives(ctx, task.ID)
		if err != nil {
			return nil, fmt.Errorf("load objectives: %w", err)
		}

		task.Tags, err = s.loadTags(ctx, task.ID)
		if err != nil {
			return nil, fmt.Errorf("load tags: %w", err)
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

	// Load related data
	for _, task := range tasks {
		task.Objectives, _ = s.loadObjectives(ctx, task.ID)
		task.Tags, _ = s.loadTags(ctx, task.ID)
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
	for i, id := range ids {
		_, err = tx.ExecContext(ctx, `
			UPDATE tasks SET order_index = ? WHERE id = ?
		`, i, id)
		if err != nil {
			return fmt.Errorf("update task order: %w", err)
		}
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

	// Build IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM tasks WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete tasks: %w", err)
	}

	return nil
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
