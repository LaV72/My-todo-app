package service

import "time"

// Config holds configuration for service layer business rules
type Config struct {
	// Validation limits
	MaxTitleLength       int
	MaxDescriptionLength int
	MaxBulkSize          int

	// Business rules
	RequireAllObjectives       bool
	AutoCompleteOnFullProgress bool
	AllowPastDeadlines         bool

	// Features
	EnableCategoryRestrictions bool
	EnableRewardSystem         bool
}

// DefaultConfig returns default service configuration
func DefaultConfig() *Config {
	return &Config{
		MaxTitleLength:             200,
		MaxDescriptionLength:       2000,
		MaxBulkSize:                50,
		RequireAllObjectives:       false,
		AutoCompleteOnFullProgress: true,
		AllowPastDeadlines:         false,
		EnableCategoryRestrictions: true,
		EnableRewardSystem:         true,
	}
}

// IDGenerator generates unique IDs
type IDGenerator interface {
	Generate() string
}

// Clock provides current time (abstracted for testing)
type Clock interface {
	Now() time.Time
}

// SystemClock returns current system time
type SystemClock struct{}

func (c *SystemClock) Now() time.Time {
	return time.Now()
}
