package sqlite

import (
	"context"
	"fmt"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
)

// AddObjective adds a new objective to a task
func (s *SQLiteStorage) AddObjective(ctx context.Context, taskID string, obj *models.Objective) error {
	// Verify task exists
	_, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err // Will be ErrNotFound if task doesn't exist
	}

	// Insert objective
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO objectives (id, task_id, text, completed, order_index, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, obj.ID, taskID, obj.Text, obj.Completed, obj.Order, obj.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert objective: %w", err)
	}

	return nil
}

// UpdateObjective updates an existing objective
func (s *SQLiteStorage) UpdateObjective(ctx context.Context, taskID, objID string, obj *models.Objective) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE objectives
		SET text = ?, completed = ?, order_index = ?
		WHERE id = ? AND task_id = ?
	`, obj.Text, obj.Completed, obj.Order, objID, taskID)

	if err != nil {
		return fmt.Errorf("update objective: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: objective %s in task %s", storage.ErrNotFound, objID, taskID)
	}

	return nil
}

// DeleteObjective deletes an objective from a task
func (s *SQLiteStorage) DeleteObjective(ctx context.Context, taskID, objID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM objectives WHERE id = ? AND task_id = ?
	`, objID, taskID)

	if err != nil {
		return fmt.Errorf("delete objective: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: objective %s in task %s", storage.ErrNotFound, objID, taskID)
	}

	return nil
}
