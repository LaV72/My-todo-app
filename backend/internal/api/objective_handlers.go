package api

import (
	"net/http"
	"strings"

	"github.com/LaV72/quest-todo/internal/models"
)

// CreateObjective handles POST /api/tasks/{taskId}/objectives
func (api *API) CreateObjective(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from path like /api/tasks/{taskId}/objectives
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/objectives")
	parts := strings.Split(path, "/")
	taskID := parts[0]

	if taskID == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Task ID is required")
		return
	}

	var req models.ObjectiveRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	objective, err := api.ObjectiveService.CreateObjective(r.Context(), taskID, req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusCreated, objective)
}

// UpdateObjective handles PUT /api/objectives/{id}
func (api *API) UpdateObjective(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/objectives/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Objective ID is required")
		return
	}

	var req models.ObjectiveUpdateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	objective, err := api.ObjectiveService.UpdateObjective(r.Context(), id, req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, objective)
}

// DeleteObjective handles DELETE /api/objectives/{id}
func (api *API) DeleteObjective(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/objectives/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Objective ID is required")
		return
	}

	if err := api.ObjectiveService.DeleteObjective(r.Context(), id); err != nil {
		HandleServiceError(w, err)
		return
	}

	NoContentResponse(w)
}

// ToggleObjective handles POST /api/objectives/{id}/toggle
func (api *API) ToggleObjective(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/objectives/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Objective ID is required")
		return
	}

	objective, err := api.ObjectiveService.ToggleObjective(r.Context(), id)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, objective)
}
