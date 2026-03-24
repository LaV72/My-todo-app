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

	// Create services (TaskService first, then ObjectiveService which depends on it)
	taskService := NewTaskService(storage, clock, idGen, validate, config)

	return &Services{
		Task:      taskService,
		Objective: NewObjectiveService(storage, taskService, clock, idGen, validate, config),
		Category:  NewCategoryService(storage, idGen, validate, config),
		Stats:     NewStatsService(storage),
	}
}
