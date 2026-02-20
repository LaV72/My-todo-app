package models

// APIResponse is a generic wrapper for all API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *MetaData   `json:"meta,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"` // For validation errors
}

// MetaData represents pagination and other metadata
type MetaData struct {
	Total   int  `json:"total,omitempty"`
	Limit   int  `json:"limit,omitempty"`
	Offset  int  `json:"offset,omitempty"`
	HasMore bool `json:"hasMore,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"` // seconds
	Storage string `json:"storage"`
}

// VersionResponse represents the version information
type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime,omitempty"`
	GoVersion string `json:"goVersion,omitempty"`
}

// Success creates a successful API response
func Success(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// SuccessWithMeta creates a successful API response with metadata
func SuccessWithMeta(data interface{}, meta *MetaData) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// Error creates an error API response
func Error(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}

// ValidationError creates a validation error response
func ValidationError(fields map[string]string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Fields:  fields,
		},
	}
}
