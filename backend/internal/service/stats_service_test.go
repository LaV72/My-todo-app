package service

import (
	"context"
	"testing"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsService_GetStats(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()

		mockStorage.GetStatsFunc = func(ctx context.Context) (*models.Stats, error) {
			return &models.Stats{
				TotalTasks:             100,
				ActiveTasks:            30,
				CompletedTasks:         60,
				FailedTasks:            10,
				CompletionRate:         60.0,
				TotalRewards:           5000,
				AverageTimeToComplete:  2.5,
				StreakDays:             7,
			}, nil
		}

		service := NewStatsService(mockStorage)

		// Act
		stats, err := service.GetStats(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 100, stats.TotalTasks)
		assert.Equal(t, 30, stats.ActiveTasks)
		assert.Equal(t, 60, stats.CompletedTasks)
		assert.Equal(t, 10, stats.FailedTasks)
		assert.Equal(t, 60.0, stats.CompletionRate)
		assert.Equal(t, 5000, stats.TotalRewards)
		assert.Equal(t, 2.5, stats.AverageTimeToComplete)
		assert.Equal(t, 7, stats.StreakDays)
	})

	t.Run("empty stats", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()

		mockStorage.GetStatsFunc = func(ctx context.Context) (*models.Stats, error) {
			return &models.Stats{
				TotalTasks:             0,
				ActiveTasks:            0,
				CompletedTasks:         0,
				FailedTasks:            0,
				CompletionRate:         0,
				TotalRewards:           0,
				AverageTimeToComplete:  0,
				StreakDays:             0,
			}, nil
		}

		service := NewStatsService(mockStorage)

		// Act
		stats, err := service.GetStats(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 0, stats.TotalTasks)
		assert.Equal(t, 0.0, stats.CompletionRate)
	})
}

func TestStatsService_GetCategoryStats(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()

		mockStorage.GetCategoryStatsFunc = func(ctx context.Context) (map[string]*models.CategoryStat, error) {
			return map[string]*models.CategoryStat{
				"work": {
					CategoryID:     "work",
					TotalTasks:     50,
					CompletedTasks: 30,
					CompletionRate: 60.0,
				},
				"personal": {
					CategoryID:     "personal",
					TotalTasks:     25,
					CompletedTasks: 20,
					CompletionRate: 80.0,
				},
			}, nil
		}

		service := NewStatsService(mockStorage)

		// Act
		stats, err := service.GetCategoryStats(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Len(t, stats, 2)

		// Find each stat
		var workStat, personalStat *models.CategoryStat
		for i := range stats {
			if stats[i].CategoryID == "work" {
				workStat = &stats[i]
			} else if stats[i].CategoryID == "personal" {
				personalStat = &stats[i]
			}
		}

		require.NotNil(t, workStat)
		assert.Equal(t, 50, workStat.TotalTasks)
		assert.Equal(t, 30, workStat.CompletedTasks)
		assert.Equal(t, 60.0, workStat.CompletionRate)

		require.NotNil(t, personalStat)
		assert.Equal(t, 25, personalStat.TotalTasks)
		assert.Equal(t, 20, personalStat.CompletedTasks)
		assert.Equal(t, 80.0, personalStat.CompletionRate)
	})

	t.Run("empty stats", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()

		mockStorage.GetCategoryStatsFunc = func(ctx context.Context) (map[string]*models.CategoryStat, error) {
			return make(map[string]*models.CategoryStat), nil
		}

		service := NewStatsService(mockStorage)

		// Act
		stats, err := service.GetCategoryStats(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Empty(t, stats)
	})

	t.Run("single category", func(t *testing.T) {
		// Arrange
		mockStorage := NewMockStorage()

		mockStorage.GetCategoryStatsFunc = func(ctx context.Context) (map[string]*models.CategoryStat, error) {
			return map[string]*models.CategoryStat{
				"work": {
					CategoryID:     "work",
					TotalTasks:     100,
					CompletedTasks: 75,
					CompletionRate: 75.0,
				},
			}, nil
		}

		service := NewStatsService(mockStorage)

		// Act
		stats, err := service.GetCategoryStats(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Len(t, stats, 1)
		assert.Equal(t, "work", stats[0].CategoryID)
		assert.Equal(t, 100, stats[0].TotalTasks)
	})
}
