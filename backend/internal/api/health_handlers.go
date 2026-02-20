package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/LaV72/quest-todo/internal/models"
)

// HealthCheck handles GET /health
func (api *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	uptime := int64(time.Since(api.StartTime).Seconds())

	response := models.HealthResponse{
		Status:  "ok",
		Version: api.AppVersion,
		Uptime:  uptime,
		Storage: "available",
	}

	SuccessResponse(w, http.StatusOK, response)
}

// Version handles GET /version
func (api *API) Version(w http.ResponseWriter, r *http.Request) {
	response := models.VersionResponse{
		Version:   api.AppVersion,
		GoVersion: runtime.Version(),
	}

	SuccessResponse(w, http.StatusOK, response)
}
