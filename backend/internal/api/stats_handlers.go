package api

import (
	"net/http"
)

// GetStats handles GET /api/stats
func (api *API) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := api.StatsService.GetStats(r.Context())
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, stats)
}

// GetCategoryStats handles GET /api/stats/categories
func (api *API) GetCategoryStats(w http.ResponseWriter, r *http.Request) {
	stats, err := api.StatsService.GetCategoryStats(r.Context())
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, stats)
}
