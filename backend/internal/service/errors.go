package service

import "errors"

var (
	// Input errors
	ErrInvalidInput      = errors.New("invalid input")
	ErrTaskNotFound      = errors.New("task not found")
	ErrObjectiveNotFound = errors.New("objective not found")
	ErrCategoryNotFound  = errors.New("category not found")

	// Business rule errors
	ErrDeadlineInPast           = errors.New("deadline must be in the future")
	ErrAlreadyCompleted         = errors.New("task already completed")
	ErrCannotCompleteFailedTask = errors.New("cannot complete a failed task")
	ErrObjectivesIncomplete     = errors.New("all objectives must be completed first")
	ErrBulkSizeTooLarge         = errors.New("bulk operation exceeds maximum size")
	ErrCannotDeleteCategory     = errors.New("cannot delete category with active tasks")

	// Concurrency errors
	ErrVersionConflict = errors.New("version conflict, task was modified")
)

// ValidationError represents validation errors for multiple fields
type ValidationError struct {
	Field   string
	Message string
}

// MultiValidationError holds multiple validation errors
type MultiValidationError struct {
	Errors []ValidationError
}

func (e *MultiValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return e.Errors[0].Message
}
