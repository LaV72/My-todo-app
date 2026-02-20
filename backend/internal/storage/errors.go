package storage

import "errors"

// Standard storage errors
// These are sentinel errors that can be used with errors.Is()
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrDuplicate is returned when trying to create a resource that already exists
	ErrDuplicate = errors.New("duplicate resource")

	// ErrConflict is returned when there's a conflict with existing data
	ErrConflict = errors.New("resource conflict")

	// ErrInvalid is returned when the data is invalid
	ErrInvalid = errors.New("invalid data")

	// ErrInternal is returned for internal storage errors
	ErrInternal = errors.New("internal storage error")

	// ErrNotImplemented is returned when a storage type is not yet implemented
	ErrNotImplemented = errors.New("storage type not implemented")

	// ErrUnknownStorageType is returned when an unknown storage type is requested
	ErrUnknownStorageType = errors.New("unknown storage type")

	// ErrConnectionFailed is returned when connection to storage fails
	ErrConnectionFailed = errors.New("failed to connect to storage")

	// ErrTransactionFailed is returned when a transaction fails
	ErrTransactionFailed = errors.New("transaction failed")
)

// Error codes for API responses
const (
	CodeNotFound    = "NOT_FOUND"
	CodeDuplicate   = "DUPLICATE"
	CodeConflict    = "CONFLICT"
	CodeInvalid     = "INVALID_DATA"
	CodeInternal    = "INTERNAL_ERROR"
	CodeUnknown     = "UNKNOWN_ERROR"
)

// GetErrorCode returns the appropriate error code string for an error
// This is useful for converting storage errors to API error codes
func GetErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrNotFound):
		return CodeNotFound
	case errors.Is(err, ErrDuplicate):
		return CodeDuplicate
	case errors.Is(err, ErrConflict):
		return CodeConflict
	case errors.Is(err, ErrInvalid):
		return CodeInvalid
	case errors.Is(err, ErrInternal):
		return CodeInternal
	default:
		return CodeUnknown
	}
}

// GetErrorMessage returns a user-friendly message for an error
func GetErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrNotFound):
		return "The requested resource was not found"
	case errors.Is(err, ErrDuplicate):
		return "A resource with this identifier already exists"
	case errors.Is(err, ErrConflict):
		return "The operation conflicts with existing data"
	case errors.Is(err, ErrInvalid):
		return "The provided data is invalid"
	case errors.Is(err, ErrInternal):
		return "An internal error occurred"
	case errors.Is(err, ErrNotImplemented):
		return "This feature is not yet implemented"
	case errors.Is(err, ErrUnknownStorageType):
		return "Unknown storage type specified"
	case errors.Is(err, ErrConnectionFailed):
		return "Failed to connect to storage"
	case errors.Is(err, ErrTransactionFailed):
		return "The transaction failed"
	default:
		return "An unknown error occurred"
	}
}
