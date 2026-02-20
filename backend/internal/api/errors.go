package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LaV72/quest-todo/internal/models"
	"github.com/LaV72/quest-todo/internal/service"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse sends an error response to the client
func ErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.Error(code, message))
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(w http.ResponseWriter, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make(map[string]string)
		for _, fe := range ve {
			fields[fe.Field()] = getValidationMessage(fe)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ValidationError(fields))
		return
	}

	// Fallback for generic validation errors
	ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
}

// HandleServiceError maps service errors to HTTP responses
func HandleServiceError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	// Check for specific service errors
	switch {
	case errors.Is(err, service.ErrTaskNotFound):
		ErrorResponse(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
	case errors.Is(err, service.ErrObjectiveNotFound):
		ErrorResponse(w, http.StatusNotFound, "OBJECTIVE_NOT_FOUND", "Objective not found")
	case errors.Is(err, service.ErrCategoryNotFound):
		ErrorResponse(w, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found")
	case errors.Is(err, service.ErrDeadlineInPast):
		ErrorResponse(w, http.StatusBadRequest, "INVALID_DEADLINE", "Deadline must be in the future")
	case errors.Is(err, service.ErrAlreadyCompleted):
		ErrorResponse(w, http.StatusBadRequest, "ALREADY_COMPLETED", "Task is already completed")
	case errors.Is(err, service.ErrCannotCompleteFailedTask):
		ErrorResponse(w, http.StatusBadRequest, "CANNOT_COMPLETE_FAILED", "Cannot complete a failed task")
	case errors.Is(err, service.ErrObjectivesIncomplete):
		ErrorResponse(w, http.StatusBadRequest, "OBJECTIVES_INCOMPLETE", "Cannot complete task with incomplete objectives")
	case errors.Is(err, service.ErrCannotDeleteCategory):
		ErrorResponse(w, http.StatusConflict, "CATEGORY_HAS_TASKS", "Cannot delete category with active tasks")
	case errors.Is(err, service.ErrBulkSizeTooLarge):
		ErrorResponse(w, http.StatusBadRequest, "BULK_SIZE_TOO_LARGE", "Too many items in bulk request")
	case errors.Is(err, service.ErrVersionConflict):
		ErrorResponse(w, http.StatusConflict, "VERSION_CONFLICT", "Resource was modified by another request")
	default:
		// Generic internal server error
		ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}

// getValidationMessage returns a human-readable validation error message
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value is too short or too small"
	case "max":
		return "Value is too long or too large"
	case "email":
		return "Invalid email format"
	case "hexcolor":
		return "Invalid hex color format"
	case "oneof":
		return "Invalid value, must be one of: " + fe.Param()
	default:
		return "Invalid value"
	}
}

// SuccessResponse sends a success response to the client
func SuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.Success(data))
}

// SuccessResponseWithMeta sends a success response with metadata
func SuccessResponseWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta *models.MetaData) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.SuccessWithMeta(data, meta))
}

// NoContentResponse sends a 204 No Content response
func NoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
