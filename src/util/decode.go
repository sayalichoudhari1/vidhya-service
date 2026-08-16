package util

import (
	"encoding/json"
	"net/http"

	"vidhya-service/src/apperrors"
)

// DecodeJSON decodes the request body into dst, returning a well-formed
// apperrors.AppError (400 INVALID_PAYLOAD) on malformed JSON.
func DecodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apperrors.WithMessage(apperrors.ErrInvalidPayload, "malformed JSON body: "+err.Error())
	}
	return nil
}
