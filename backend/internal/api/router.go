package api

import (
	"net/http"
	"strings"

	"github.com/LaV72/quest-todo/internal/api/middleware"
)

// RouterConfig holds configuration for the router
type RouterConfig struct {
	AllowedOrigins []string
	EnableCORS     bool
	EnableLogging  bool
}

// NewRouter creates and configures the HTTP router with all routes and middleware
func NewRouter(api *API, config RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Health and version endpoints (no /api prefix)
	mux.HandleFunc("/health", api.HealthCheck)
	mux.HandleFunc("/version", api.Version)

	// Task routes
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Check for search query parameter
			if r.URL.Query().Get("q") != "" {
				api.SearchTasks(w, r)
			} else {
				api.ListTasks(w, r)
			}
		case http.MethodPost:
			api.CreateTask(w, r)
		default:
			methodNotAllowed(w, r)
		}
	})

	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		path = strings.TrimSuffix(path, "/") // Remove trailing slash

		// Reject empty segments (e.g., /api/tasks//complete)
		if strings.Contains(path, "//") {
			http.NotFound(w, r)
			return
		}

		parts := strings.Split(path, "/")

		// Handle special routes first
		switch parts[0] {
		case "bulk":
			// Validate: exactly 1 segment
			if len(parts) != 1 {
				http.NotFound(w, r)
				return
			}
			switch r.Method {
			case http.MethodPost:
				api.CreateTasksBulk(w, r)
			case http.MethodDelete:
				api.DeleteTasksBulk(w, r)
			default:
				methodNotAllowed(w, r)
			}
			return
		case "reorder":
			// Validate: exactly 1 segment
			if len(parts) != 1 {
				http.NotFound(w, r)
				return
			}
			if r.Method == http.MethodPost {
				api.ReorderTasks(w, r)
			} else {
				methodNotAllowed(w, r)
			}
			return
		}

		// Handle task ID routes
		if len(parts) >= 1 && parts[0] != "" {
			// Check for action routes
			if len(parts) == 2 {
				action := parts[1]
				switch action {
				case "complete":
					if r.Method == http.MethodPost {
						api.CompleteTask(w, r)
					} else {
						methodNotAllowed(w, r)
					}
					return
				case "fail":
					if r.Method == http.MethodPost {
						api.FailTask(w, r)
					} else {
						methodNotAllowed(w, r)
					}
					return
				case "reactivate":
					if r.Method == http.MethodPost {
						api.ReactivateTask(w, r)
					} else {
						methodNotAllowed(w, r)
					}
					return
				case "objectives":
					if r.Method == http.MethodPost {
						api.CreateObjective(w, r)
					} else {
						methodNotAllowed(w, r)
					}
					return
				default:
					// Unknown action with 2 segments
					http.NotFound(w, r)
					return
				}
			} else if len(parts) > 2 {
				// Too many segments (e.g., /api/tasks/123/extra/stuff)
				http.NotFound(w, r)
				return
			}

			// Standard CRUD operations on task (exactly 1 segment: the ID)
			switch r.Method {
			case http.MethodGet:
				api.GetTask(w, r)
			case http.MethodPut:
				api.UpdateTask(w, r)
			case http.MethodDelete:
				api.DeleteTask(w, r)
			default:
				methodNotAllowed(w, r)
			}
			return
		}

		http.NotFound(w, r)
	})

	// Objective routes
	mux.HandleFunc("/api/objectives/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/objectives/")
		path = strings.TrimSuffix(path, "/") // Remove trailing slash

		// Reject empty segments
		if strings.Contains(path, "//") {
			http.NotFound(w, r)
			return
		}

		parts := strings.Split(path, "/")

		if len(parts) >= 1 && parts[0] != "" {
			// Check for toggle action
			if len(parts) == 2 && parts[1] == "toggle" {
				if r.Method == http.MethodPost {
					api.ToggleObjective(w, r)
				} else {
					methodNotAllowed(w, r)
				}
				return
			} else if len(parts) > 2 {
				// Too many segments
				http.NotFound(w, r)
				return
			} else if len(parts) == 2 {
				// Unknown action with 2 segments
				http.NotFound(w, r)
				return
			}

			// Standard CRUD operations (exactly 1 segment: the ID)
			switch r.Method {
			case http.MethodPut:
				api.UpdateObjective(w, r)
			case http.MethodDelete:
				api.DeleteObjective(w, r)
			default:
				methodNotAllowed(w, r)
			}
			return
		}

		http.NotFound(w, r)
	})

	// Category routes
	mux.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.ListCategories(w, r)
		case http.MethodPost:
			api.CreateCategory(w, r)
		default:
			methodNotAllowed(w, r)
		}
	})

	mux.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/categories/")
		path = strings.TrimSuffix(path, "/") // Remove trailing slash

		// Reject empty segments
		if strings.Contains(path, "//") || path == "" {
			http.NotFound(w, r)
			return
		}

		parts := strings.Split(path, "/")

		// Validate: exactly 1 segment (the category ID)
		if len(parts) != 1 {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.GetCategory(w, r)
		case http.MethodPut:
			api.UpdateCategory(w, r)
		case http.MethodDelete:
			api.DeleteCategory(w, r)
		default:
			methodNotAllowed(w, r)
		}
	})

	// Stats routes
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetStats(w, r)
		} else {
			methodNotAllowed(w, r)
		}
	})

	mux.HandleFunc("/api/stats/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetCategoryStats(w, r)
		} else {
			methodNotAllowed(w, r)
		}
	})

	// Apply middleware chain
	var handler http.Handler = mux

	// Always apply these middleware (in reverse order of execution)
	// Recovery should be outermost to catch all panics
	handler = middleware.ContentType(handler)
	handler = middleware.RequestID(handler)
	handler = middleware.Recovery(handler)

	// Optional middleware
	if config.EnableLogging {
		handler = middleware.Logger(handler)
	}

	if config.EnableCORS {
		handler = middleware.CORS(config.AllowedOrigins)(handler)
	}

	return handler
}

// methodNotAllowed returns a 405 Method Not Allowed response
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
}
