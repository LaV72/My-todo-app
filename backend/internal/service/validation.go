package service

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Valid deadline types
var validDeadlineTypes = map[string]bool{
	"short":  true,
	"medium": true,
	"long":   true,
	"none":   true,
}

// validateDeadlineType checks if a deadline type is valid
func validateDeadlineType(deadlineType string) bool {
	return validDeadlineTypes[deadlineType]
}

// wrapValidationError converts validator errors to our custom error format
func wrapValidationError(err error) error {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return ErrInvalidInput
	}

	errors := make([]ValidationError, len(validationErrs))
	for i, fieldErr := range validationErrs {
		errors[i] = ValidationError{
			Field:   fieldErr.Field(),
			Message: getValidationErrorMessage(fieldErr),
		}
	}

	return &MultiValidationError{Errors: errors}
}

func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", err.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}
