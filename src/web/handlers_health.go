package web

import (
	"net/http"

	"vidhya-service/src/app"
	"vidhya-service/src/util"
	"vidhya-service/src/version"
)

// handleHealth handles GET /health - a liveness/readiness probe that also
// verifies the database connection, since that's the one hard dependency
// this service has today.
func handleHealth(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{"status": "UP", "database": "UP"}
		httpStatus := http.StatusOK

		sqlDB, err := a.DB.DB()
		if err != nil || sqlDB.PingContext(r.Context()) != nil {
			status["status"] = "DOWN"
			status["database"] = "DOWN"
			httpStatus = http.StatusServiceUnavailable
		}

		util.WriteJSON(w, httpStatus, status)
	}
}

// handleVersion handles GET /version.
func handleVersion(w http.ResponseWriter, r *http.Request) {
	util.WriteOK(w, map[string]string{
		"service": "vidhya-service",
		"version": version.BuildNumber,
		"commit":  version.BuildCommitSHA,
	})
}
