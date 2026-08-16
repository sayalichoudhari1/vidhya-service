package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/model"
)

// resolveSchoolID reads the {school_id} path variable and enforces tenant
// isolation: ADMIN may operate on any school; every other role may only
// operate within their own school (the one on their JWT).
func resolveSchoolID(r *http.Request) (string, error) {
	schoolID := mux.Vars(r)["school_id"]
	if schoolID == "" {
		return "", apperrors.WithMessage(apperrors.ErrInvalidPayload, "school_id path parameter is required")
	}

	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		return "", apperrors.ErrUnauthorized
	}
	if claims.Role == model.RoleAdmin {
		return schoolID, nil
	}
	if claims.SchoolID != schoolID {
		return "", apperrors.WithMessage(apperrors.ErrForbidden, "you do not have access to this school")
	}
	return schoolID, nil
}

// Chain applies middleware in the given order (left-most runs first) and
// returns the fully wrapped handler, ready to register on the router.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
