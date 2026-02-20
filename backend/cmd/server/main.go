package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LaV72/quest-todo/internal/api"
	"github.com/LaV72/quest-todo/internal/service"
	"github.com/LaV72/quest-todo/internal/storage/sqlite"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	// Version of the application
	Version = "0.1.0"

	// Default configuration values
	defaultHost         = "localhost"
	defaultPort         = "8080"
	defaultDBPath       = "./quest-todo.db"
	defaultReadTimeout  = 15 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Server ServerConfig

	// Service layer configuration
	Service service.Config

	// Router configuration
	Router api.RouterConfig

	// Database configuration
	Database DatabaseConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string
}

// Address returns the full server address
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

// loadConfig loads configuration from environment variables with sensible defaults
func loadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("HOST", defaultHost),
			Port:         getEnv("PORT", defaultPort),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", defaultReadTimeout),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", defaultWriteTimeout),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", defaultIdleTimeout),
		},
		Service: service.Config{
			// Validation limits
			MaxTitleLength:       getIntEnv("MAX_TITLE_LENGTH", 200),
			MaxDescriptionLength: getIntEnv("MAX_DESCRIPTION_LENGTH", 2000),
			MaxBulkSize:          getIntEnv("MAX_BULK_SIZE", 100),

			// Business rules
			RequireAllObjectives:       getBoolEnv("REQUIRE_ALL_OBJECTIVES", false),
			AutoCompleteOnFullProgress: getBoolEnv("AUTO_COMPLETE_ON_FULL_PROGRESS", true),
			AllowPastDeadlines:         getBoolEnv("ALLOW_PAST_DEADLINES", true),

			// Features
			EnableCategoryRestrictions: getBoolEnv("ENABLE_CATEGORY_RESTRICTIONS", false),
			EnableRewardSystem:         getBoolEnv("ENABLE_REWARD_SYSTEM", false),
		},
		Router: api.RouterConfig{
			AllowedOrigins: parseOrigins(getEnv("ALLOWED_ORIGINS", "*")),
			EnableCORS:     getBoolEnv("ENABLE_CORS", true),
			EnableLogging:  getBoolEnv("ENABLE_LOGGING", true),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", defaultDBPath),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv gets an integer environment variable or returns a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var i int
		if _, err := fmt.Sscanf(value, "%d", &i); err == nil {
			return i
		}
	}
	return defaultValue
}

// getBoolEnv gets a boolean environment variable or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable or returns a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// parseOrigins parses a comma-separated list of origins
func parseOrigins(s string) []string {
	if s == "" {
		return []string{"*"}
	}
	origins := strings.Split(s, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
}

// RealClock implements service.Clock interface using actual time
type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

// UUIDGenerator implements service.IDGenerator interface using UUID v4
type UUIDGenerator struct{}

func (UUIDGenerator) Generate() string {
	return uuid.New().String()
}

func main() {
	// Load configuration
	config := loadConfig()

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Quest Todo API Server")
	log.Printf("Version: %s", Version)

	// Initialize database
	log.Printf("Initializing SQLite database at %s", config.Database.Path)
	storage, err := sqlite.New(config.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()
	log.Println("Database initialized and migrated successfully")

	// Initialize shared dependencies
	clock := RealClock{}
	idGen := UUIDGenerator{}
	validate := validator.New()

	// Initialize services
	log.Println("Initializing services...")
	taskService := service.NewTaskService(storage, clock, idGen, validate, &config.Service)
	objectiveService := service.NewObjectiveService(storage, clock, idGen, validate, &config.Service)
	categoryService := service.NewCategoryService(storage, validate, &config.Service)
	statsService := service.NewStatsService(storage)
	log.Println("Services initialized successfully")

	// Initialize API handlers
	log.Println("Initializing API handlers...")
	apiHandler := api.NewAPI(
		taskService,
		objectiveService,
		categoryService,
		statsService,
		validate,
		Version,
	)
	log.Println("API handlers initialized successfully")

	// Create router with middleware
	log.Println("Setting up HTTP router and middleware...")
	router := api.NewRouter(apiHandler, config.Router)
	log.Println("Router configured successfully")

	// Create HTTP server
	server := &http.Server{
		Addr:         config.Server.Address(),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server listening on http://%s", server.Addr)
		log.Println("Press Ctrl+C to shut down")
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)

	case sig := <-shutdown:
		log.Printf("Received signal: %v", sig)
		log.Println("Starting graceful shutdown...")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			log.Println("Forcing shutdown...")
			if err := server.Close(); err != nil {
				log.Printf("Error forcing shutdown: %v", err)
			}
		} else {
			log.Println("Server stopped gracefully")
		}
	}

	log.Println("Shutdown complete")
}
