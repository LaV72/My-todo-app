package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
)

// CreateCategory creates a new category
func (s *SQLiteStorage) CreateCategory(ctx context.Context, cat *models.Category) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO categories (id, name, color, icon, type, order_index, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, cat.ID, cat.Name, cat.Color, cat.Icon, cat.Type, cat.Order, cat.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}

	return nil
}

// GetCategory retrieves a single category by ID
func (s *SQLiteStorage) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	cat := &models.Category{}

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, color, icon, type, order_index, created_at
		FROM categories
		WHERE id = ?
	`, id).Scan(&cat.ID, &cat.Name, &cat.Color, &cat.Icon, &cat.Type, &cat.Order, &cat.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: category %s", storage.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("query category: %w", err)
	}

	return cat, nil
}

// ListCategories retrieves all categories
func (s *SQLiteStorage) ListCategories(ctx context.Context) ([]*models.Category, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, color, icon, type, order_index, created_at
		FROM categories
		ORDER BY order_index ASC, name ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	categories := []*models.Category{}
	for rows.Next() {
		cat := &models.Category{}
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Color, &cat.Icon, &cat.Type, &cat.Order, &cat.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

// UpdateCategory updates an existing category
func (s *SQLiteStorage) UpdateCategory(ctx context.Context, cat *models.Category) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE categories
		SET name = ?, color = ?, icon = ?, type = ?, order_index = ?
		WHERE id = ?
	`, cat.Name, cat.Color, cat.Icon, cat.Type, cat.Order, cat.ID)

	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: category %s", storage.ErrNotFound, cat.ID)
	}

	return nil
}

// DeleteCategory deletes a category
func (s *SQLiteStorage) DeleteCategory(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: category %s", storage.ErrNotFound, id)
	}

	return nil
}
