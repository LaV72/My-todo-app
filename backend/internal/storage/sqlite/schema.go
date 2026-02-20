package sqlite

import (
	"database/sql"
	"fmt"
)

// migrate runs database migrations to create or update the schema
func (s *SQLiteStorage) migrate() error {
	// Get current schema version
	version, err := s.getSchemaVersion()
	if err != nil {
		return err
	}

	// Apply migrations in order
	migrations := []migration{
		{version: 1, name: "initial_schema", up: migrateV1},
		// Future migrations go here:
		// {version: 2, name: "add_templates", up: migrateV2},
	}

	for _, m := range migrations {
		if version < m.version {
			if err := m.up(s.db); err != nil {
				return fmt.Errorf("migration %d (%s) failed: %w", m.version, m.name, err)
			}
			if err := s.setSchemaVersion(m.version); err != nil {
				return fmt.Errorf("failed to update schema version: %w", err)
			}
			version = m.version
		}
	}

	return nil
}

// migration represents a database migration
type migration struct {
	version int
	name    string
	up      func(*sql.DB) error
}

// getSchemaVersion returns the current schema version
func (s *SQLiteStorage) getSchemaVersion() (int, error) {
	// Create schema_version table if it doesn't exist
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return 0, err
	}

	var version int
	err = s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

// setSchemaVersion records that a migration has been applied
func (s *SQLiteStorage) setSchemaVersion(version int) error {
	_, err := s.db.Exec("INSERT INTO schema_version (version) VALUES (?)", version)
	return err
}

// migrateV1 creates the initial database schema
func migrateV1(db *sql.DB) error {
	schema := `
	-- Tasks table
	CREATE TABLE IF NOT EXISTS tasks (
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

	-- Objectives table (sub-tasks within a task)
	CREATE TABLE IF NOT EXISTS objectives (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		text TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0,
		order_index INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	-- Tags table (many-to-many relationship with tasks)
	CREATE TABLE IF NOT EXISTS task_tags (
		task_id TEXT NOT NULL,
		tag TEXT NOT NULL,
		PRIMARY KEY (task_id, tag),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	-- Categories table
	CREATE TABLE IF NOT EXISTS categories (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		color TEXT NOT NULL,
		icon TEXT,
		type TEXT NOT NULL DEFAULT 'main',
		order_index INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL
	);

	-- Performance indexes
	-- Tasks indexes (for fast queries)
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority DESC);
	CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline_date);
	CREATE INDEX IF NOT EXISTS idx_tasks_category ON tasks(category);
	CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_tasks_order ON tasks(order_index);

	-- Composite indexes for common query patterns
	CREATE INDEX IF NOT EXISTS idx_tasks_status_priority ON tasks(status, priority DESC);
	CREATE INDEX IF NOT EXISTS idx_tasks_status_deadline ON tasks(status, deadline_date);
	CREATE INDEX IF NOT EXISTS idx_tasks_category_status ON tasks(category, status);

	-- Objectives indexes
	CREATE INDEX IF NOT EXISTS idx_objectives_task ON objectives(task_id);
	CREATE INDEX IF NOT EXISTS idx_objectives_order ON objectives(task_id, order_index);

	-- Tags indexes
	CREATE INDEX IF NOT EXISTS idx_tags_task ON task_tags(task_id);
	CREATE INDEX IF NOT EXISTS idx_tags_tag ON task_tags(tag);

	-- Categories indexes
	CREATE INDEX IF NOT EXISTS idx_categories_type ON categories(type);
	CREATE INDEX IF NOT EXISTS idx_categories_order ON categories(order_index);
	`

	_, err := db.Exec(schema)
	return err
}
