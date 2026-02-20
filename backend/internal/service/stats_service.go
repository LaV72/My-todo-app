package service

import (
	"context"
	"fmt"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/storage"
)

// StatsServiceImpl implements StatsService
type StatsServiceImpl struct {
	storage storage.Storage
}

// NewStatsService creates a new StatsService
func NewStatsService(storage storage.Storage) StatsService {
	return &StatsServiceImpl{
		storage: storage,
	}
}

// GetStats retrieves overall statistics
func (s *StatsServiceImpl) GetStats(ctx context.Context) (*models.Stats, error) {
	stats, err := s.storage.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return stats, nil
}

// GetCategoryStats retrieves per-category statistics
func (s *StatsServiceImpl) GetCategoryStats(ctx context.Context) ([]models.CategoryStat, error) {
	statsMap, err := s.storage.GetCategoryStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get category stats: %w", err)
	}

	// Convert map to slice
	stats := make([]models.CategoryStat, 0, len(statsMap))
	for _, stat := range statsMap {
		stats = append(stats, *stat)
	}

	return stats, nil
}
