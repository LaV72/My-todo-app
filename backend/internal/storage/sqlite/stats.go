package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
)

// GetStats retrieves overall statistics
func (s *SQLiteStorage) GetStats(ctx context.Context) (*models.Stats, error) {
	stats := &models.Stats{
		CategoryStats: make(map[string]int),
		PriorityStats: make(map[int]int),
	}

	// Get total task count
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks`).Scan(&stats.TotalTasks)
	if err != nil {
		return nil, fmt.Errorf("count total tasks: %w", err)
	}

	// Get active tasks count
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks WHERE status IN (?, ?)
	`, models.StatusActive, models.StatusInProgress).Scan(&stats.ActiveTasks)
	if err != nil {
		return nil, fmt.Errorf("count active tasks: %w", err)
	}

	// Get completed tasks count
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks WHERE status = ?
	`, models.StatusComplete).Scan(&stats.CompletedTasks)
	if err != nil {
		return nil, fmt.Errorf("count completed tasks: %w", err)
	}

	// Get failed tasks count
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks WHERE status = ?
	`, models.StatusFailed).Scan(&stats.FailedTasks)
	if err != nil {
		return nil, fmt.Errorf("count failed tasks: %w", err)
	}

	// Get total rewards
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(reward), 0) FROM tasks WHERE status = ?
	`, models.StatusComplete).Scan(&stats.TotalRewards)
	if err != nil {
		return nil, fmt.Errorf("sum rewards: %w", err)
	}

	// Calculate completion rate
	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks)
	}

	// Get category stats
	rows, err := s.db.QueryContext(ctx, `
		SELECT category, COUNT(*) as count
		FROM tasks
		WHERE category != ''
		GROUP BY category
	`)
	if err != nil {
		return nil, fmt.Errorf("query category stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}
		stats.CategoryStats[category] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	// Get priority stats
	rows, err = s.db.QueryContext(ctx, `
		SELECT priority, COUNT(*) as count
		FROM tasks
		GROUP BY priority
		ORDER BY priority DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query priority stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var priority, count int
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, fmt.Errorf("scan priority stat: %w", err)
		}
		stats.PriorityStats[priority] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	// TODO: Calculate average time to complete (requires tracking completion time)
	// TODO: Calculate streak days (requires daily activity tracking)

	return stats, nil
}

// GetDailyStats retrieves statistics for each day in a date range
func (s *SQLiteStorage) GetDailyStats(ctx context.Context, from, to time.Time) ([]*models.DailyStat, error) {
	// This is a simplified version - in production, you'd want a daily_stats table
	// that gets populated as tasks are completed/created

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as tasks_created
		FROM tasks
		WHERE DATE(created_at) BETWEEN DATE(?) AND DATE(?)
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, from, to)

	if err != nil {
		return nil, fmt.Errorf("query daily stats: %w", err)
	}
	defer rows.Close()

	stats := []*models.DailyStat{}
	for rows.Next() {
		stat := &models.DailyStat{}
		var dateStr string
		if err := rows.Scan(&dateStr, &stat.TasksCreated); err != nil {
			return nil, fmt.Errorf("scan daily stat: %w", err)
		}

		// Parse date string
		stat.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date: %w", err)
		}

		stats = append(stats, stat)
	}

	// Get completion stats in a single query
	completionRows, err := s.db.QueryContext(ctx, `
		SELECT
			DATE(completed_at) as date,
			COUNT(*) as tasks_completed,
			COALESCE(SUM(reward), 0) as rewards_earned
		FROM tasks
		WHERE completed_at IS NOT NULL
			AND DATE(completed_at) BETWEEN DATE(?) AND DATE(?)
		GROUP BY DATE(completed_at)
	`, from, to)

	if err != nil {
		return nil, fmt.Errorf("query completion stats: %w", err)
	}
	defer completionRows.Close()

	// Create map for fast lookup
	completionMap := make(map[string]*models.DailyStat)
	for completionRows.Next() {
		var dateStr string
		var completed, rewards int
		if err := completionRows.Scan(&dateStr, &completed, &rewards); err != nil {
			return nil, fmt.Errorf("scan completion stat: %w", err)
		}

		completionMap[dateStr] = &models.DailyStat{
			TasksCompleted: completed,
			RewardsEarned:  rewards,
		}
	}

	// Merge completion data with creation data
	for _, stat := range stats {
		dateStr := stat.Date.Format("2006-01-02")
		if completion, exists := completionMap[dateStr]; exists {
			stat.TasksCompleted = completion.TasksCompleted
			stat.RewardsEarned = completion.RewardsEarned
		}
	}

	return stats, nil
}

// GetCategoryStats retrieves statistics per category
func (s *SQLiteStorage) GetCategoryStats(ctx context.Context) (map[string]*models.CategoryStat, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			category,
			COUNT(*) as total,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as completed
		FROM tasks
		WHERE category != ''
		GROUP BY category
	`, models.StatusComplete)

	if err != nil {
		return nil, fmt.Errorf("query category stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]*models.CategoryStat)
	for rows.Next() {
		var categoryID string
		var total, completed int

		if err := rows.Scan(&categoryID, &total, &completed); err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}

		completionRate := 0.0
		if total > 0 {
			completionRate = float64(completed) / float64(total)
		}

		stats[categoryID] = &models.CategoryStat{
			CategoryID:     categoryID,
			TotalTasks:     total,
			CompletedTasks: completed,
			CompletionRate: completionRate,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return stats, nil
}
