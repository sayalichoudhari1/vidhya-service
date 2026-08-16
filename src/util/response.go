// Package util contains small HTTP, security and validation helpers shared
// across the web handlers.
package util

import (
	"context"
	"encoding/json"
	"net/http"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/logger"
	"vidhya-service/src/requestid"
)

// RequestIDFromContext returns the correlation ID for the current request.
func RequestIDFromContext(ctx context.Context) string {
	return requestid.FromContext(ctx)
}

// errorEnvelope is the JSON shape returned for every failed API call.
type errorEnvelope struct {
	Error apperrors.AppError `json:"error"`
}

// WriteJSON writes any value as a JSON response body with the given status.
func WriteJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Log.Error("", "RESPONSE_ENCODE_ERROR", "failed to write JSON response: %s", err)
	}
}

// WriteError writes a structured error response. If err is an *apperrors.AppError
// its HTTP status/code/message are used as-is; any other error is treated as
// an internal server error (its details are logged, not exposed to the client).
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		WriteJSON(w, appErr.HTTPStatus, errorEnvelope{Error: *appErr})
		return
	}

	// Safety net: a raw Postgres constraint violation (e.g. a duplicate
	// slipping past an application-level pre-check due to a race) is
	// translated into a clean 4xx instead of leaking a 500.
	if translated, ok := apperrors.TranslateDBError(err).(*apperrors.AppError); ok {
		WriteJSON(w, translated.HTTPStatus, errorEnvelope{Error: *translated})
		return
	}

	reqID := RequestIDFromContext(r.Context())
	logger.Log.Error(reqID, "UNHANDLED_ERROR", "%s %s: %v", r.Method, r.URL.Path, err)
	WriteJSON(w, http.StatusInternalServerError, errorEnvelope{Error: *apperrors.ErrInternal})
}

// WriteCreated writes a 201 Created response with the given body.
func WriteCreated(w http.ResponseWriter, body interface{}) {
	WriteJSON(w, http.StatusCreated, body)
}

// WriteOK writes a 200 OK response with the given body.
func WriteOK(w http.ResponseWriter, body interface{}) {
	WriteJSON(w, http.StatusOK, body)
}

// WriteNoContent writes a 204 No Content response.
func WriteNoContent(w http.ResponseWriter) {
	WriteJSON(w, http.StatusNoContent, nil)
}
