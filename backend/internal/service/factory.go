package service

import (
	"github.com/LaV72/quest-todo/internal/storage"
	"github.com/go-playground/validator/v10"
)

// Services holds all service instances
type Services struct {
	Task      TaskService
	Objective ObjectiveService
	Category  CategoryService
	Stats     StatsService
}

// NewServices creates all services with shared dependencies
func NewServices(storage storage.Storage, config *Config) *Services {
	if config == nil {
		config = DefaultConfig()
	}

	// Shared dependencies
	validate := validator.New()
	idGen := &UUIDGenerator{}
	clock := &SystemClock{}

	// Create services
	return &Services{
		Task:      NewTaskService(storage, clock, idGen, validate, config),
		Objective: NewObjectiveService(storage, clock, idGen, validate, config),
		Category:  NewCategoryService(storage, validate, config),
		Stats:     NewStatsService(storage),
	}
}
