package api

import (
	"net/http"

	"github.com/LaV72/quest-todo/internal/models"
)

// CreateCategory handles POST /api/categories
func (api *API) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CategoryCreateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	category, err := api.CategoryService.CreateCategory(r.Context(), req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusCreated, category)
}

// GetCategory handles GET /api/categories/{id}
func (api *API) GetCategory(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/categories/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Category ID is required")
		return
	}

	category, err := api.CategoryService.GetCategory(r.Context(), id)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, category)
}

// UpdateCategory handles PUT /api/categories/{id}
func (api *API) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/categories/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Category ID is required")
		return
	}

	var req models.CategoryUpdateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if err := api.Validator.Struct(req); err != nil {
		ValidationErrorResponse(w, err)
		return
	}

	category, err := api.CategoryService.UpdateCategory(r.Context(), id, req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, category)
}

// DeleteCategory handles DELETE /api/categories/{id}
func (api *API) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := ExtractID(r, "/api/categories/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_ID", "Category ID is required")
		return
	}

	if err := api.CategoryService.DeleteCategory(r.Context(), id); err != nil {
		HandleServiceError(w, err)
		return
	}

	NoContentResponse(w)
}

// ListCategories handles GET /api/categories
func (api *API) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := api.CategoryService.ListCategories(r.Context())
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, categories)
}
