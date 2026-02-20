package api

import (
	"net/http"

	"github.com/LaV72/quest-todo/internal/models"
)

// CreateTask handles POST /api/tasks
func (api *API) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.TaskCreateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	task, err := api.TaskService.CreateTask(r.Context(), req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusCreated, task)
}

// GetTask handles GET /api/tasks/{id}
func (api *API) GetTask(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/tasks/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Task ID is required")
		return
	}

	task, err := api.TaskService.GetTask(r.Context(), id)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, task)
}

// UpdateTask handles PUT /api/tasks/{id}
func (api *API) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/tasks/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Task ID is required")
		return
	}

	var req models.TaskUpdateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	task, err := api.TaskService.UpdateTask(r.Context(), id, req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, task)
}

// DeleteTask handles DELETE /api/tasks/{id}
func (api *API) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/tasks/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Task ID is required")
		return
	}

	if err := api.TaskService.DeleteTask(r.Context(), id); err != nil {
		HandleServiceError(w, err)
		return
	}

	NoContentResponse(w)
}

// ListTasks handles GET /api/tasks
func (api *API) ListTasks(w http.ResponseWriter, r *http.Request) {
	filter := ParseTaskFilter(r)

	tasks, err := api.TaskService.ListTasks(r.Context(), filter)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	// Get total count if pagination is used
	if filter.Limit > 0 {
		total, err := api.TaskService.CountTasks(r.Context(), filter)
		if err != nil {
			HandleServiceError(w, err)
			return
		}

		meta := &models.MetaData{
			Total:   total,
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			HasMore: filter.Offset+len(tasks) < total,
		}
		SuccessResponseWithMeta(w, http.StatusOK, tasks, meta)
		return
	}

	SuccessResponse(w, http.StatusOK, tasks)
}

// SearchTasks handles GET /api/tasks/search?q=query
func (api *API) SearchTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		ErrorResponse(w, http.StatusBadRequest, "MISSING_QUERY", "Search query is required")
		return
	}

	tasks, err := api.TaskService.SearchTasks(r.Context(), query)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, tasks)
}

// CreateTasksBulk handles POST /api/tasks/bulk
func (api *API) CreateTasksBulk(w http.ResponseWriter, r *http.Request) {
	var req models.BulkTaskCreateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	tasks, err := api.TaskService.CreateTasksBulk(r.Context(), req.Tasks)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusCreated, tasks)
}

// DeleteTasksBulk handles DELETE /api/tasks/bulk
func (api *API) DeleteTasksBulk(w http.ResponseWriter, r *http.Request) {
	var req models.BulkTaskDeleteRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	if err := api.TaskService.DeleteTasksBulk(r.Context(), req.IDs); err != nil {
		HandleServiceError(w, err)
		return
	}

	NoContentResponse(w)
}

// CompleteTask handles POST /api/tasks/{id}/complete
func (api *API) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/tasks/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Task ID is required")
		return
	}

	task, err := api.TaskService.CompleteTask(r.Context(), id)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, task)
}

// FailTask handles POST /api/tasks/{id}/fail
func (api *API) FailTask(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/tasks/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Task ID is required")
		return
	}

	task, err := api.TaskService.FailTask(r.Context(), id)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, task)
}

// ReactivateTask handles POST /api/tasks/{id}/reactivate
func (api *API) ReactivateTask(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/tasks/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Task ID is required")
		return
	}

	task, err := api.TaskService.ReactivateTask(r.Context(), id)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, task)
}

// ReorderTasks handles POST /api/tasks/reorder
func (api *API) ReorderTasks(w http.ResponseWriter, r *http.Request) {
	var req models.TaskReorderRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	if err := api.TaskService.ReorderTasks(r.Context(), req.IDs); err != nil {
		HandleServiceError(w, err)
		return
	}

	NoContentResponse(w)
}
