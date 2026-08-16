package web

import (
	"net/http"

	"vidhya-service/src/app"
	"vidhya-service/src/apperrors"
	"vidhya-service/src/services/auth"
	"vidhya-service/src/util"
)

// handleLogin handles POST /v1/auth/login.
func handleLogin(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req auth.LoginRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}

		resp, err := a.AuthService.Login(r.Context(), req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, resp)
	}
}

// handleMe handles GET /v1/auth/me - returns the authenticated caller's profile.
func handleMe(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil {
			util.WriteError(w, r, apperrors.ErrUnauthorized)
			return
		}

		me, err := a.AuthService.Me(r.Context(), claims.UserID)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, me)
	}
}
