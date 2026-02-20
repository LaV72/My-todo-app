package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/LaV72/quest-todo/internal/storage"
	_ "modernc.org/sqlite" // SQLite driver
)

// SQLiteStorage implements the storage.Storage interface using SQLite
type SQLiteStorage struct {
	db         *sql.DB
	dbPath     string
	openedAt   time.Time
}

// New creates a new SQLite storage instance
func New(path string) (*SQLiteStorage, error) {
	// Open database connection
	// Note: ?_journal_mode=WAL enables Write-Ahead Logging for better concurrency
	// ?_busy_timeout=5000 waits up to 5 seconds if database is locked
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on", path)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrConnectionFailed, err)
	}

	// Configure connection pool
	// SQLite can only handle one writer at a time
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Connections never expire

	s := &SQLiteStorage{
		db:       db,
		dbPath:   path,
		openedAt: time.Now(),
	}

	// Test connection
	if err := s.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Run migrations to create/update schema
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (s *SQLiteStorage) Ping() error {
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("%w: %v", storage.ErrConnectionFailed, err)
	}
	return nil
}

// Backup creates a backup of the database
func (s *SQLiteStorage) Backup(dest string) error {
	// SQLite backup using VACUUM INTO (available in SQLite 3.27.0+)
	_, err := s.db.Exec("VACUUM INTO ?", dest)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	return nil
}

// DB returns the underlying *sql.DB for advanced usage
// This is useful for transactions or custom queries
func (s *SQLiteStorage) DB() *sql.DB {
	return s.db
}

// Uptime returns how long the storage has been running
func (s *SQLiteStorage) Uptime() time.Duration {
	return time.Since(s.openedAt)
}
