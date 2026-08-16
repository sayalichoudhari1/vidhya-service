// Package web contains the HTTP router, middleware, and handlers for every
// Vidhya Service API.
package web

import (
	"context"
	"net/http"
	"strings"

	"vidhya-service/src/apperrors"
	"vidhya-service/src/model"
	"vidhya-service/src/util"
)

type contextKey string

const claimsKey contextKey = "claims"

// ClaimsFromContext extracts the authenticated caller's claims, if any.
func ClaimsFromContext(ctx context.Context) *util.Claims {
	claims, _ := ctx.Value(claimsKey).(*util.Claims)
	return claims
}

// AuthMiddleware validates the Authorization: Bearer <token> header on every
// request and stores the decoded claims in the request context. It is the
// only thing standing in for the SSO/IdP the original service relied on.
func AuthMiddleware(issuer *util.JWTIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				util.WriteError(w, r, apperrors.WithMessage(apperrors.ErrUnauthorized, "missing or malformed Authorization header"))
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := issuer.ParseAccessToken(tokenStr)
			if err != nil {
				util.WriteError(w, r, apperrors.WithMessage(apperrors.ErrUnauthorized, "invalid or expired token"))
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRoles returns a middleware that only allows requests whose JWT role
// is one of the given roles. This is the entire RBAC enforcement point: every
// route declares exactly which roles may call it (see router.go).
func RequireRoles(roles ...model.Role) func(http.Handler) http.Handler {
	allowed := make(map[model.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				util.WriteError(w, r, apperrors.ErrUnauthorized)
				return
			}
			if !allowed[claims.Role] {
				util.WriteError(w, r, apperrors.WithMessage(apperrors.ErrForbidden,
					"role "+string(claims.Role)+" is not permitted to perform this action"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware allows the (separately developed) frontend running on a
// different localhost port to call this API during development.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := "*"
			if len(allowedOrigins) > 0 {
				origin = allowedOrigins[0]
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
