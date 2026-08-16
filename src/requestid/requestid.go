// Package requestid propagates a per-request correlation ID through the
// request context so it can be logged consistently across a request's
// lifecycle and returned to the caller via the X-Request-Id header.
package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const key contextKey = "requestID"

// Header is the HTTP header used to propagate/return the request ID.
const Header = "X-Request-Id"

// NewContext returns a context carrying the given request ID.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

// FromContext extracts the request ID from context, if any.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(key).(string)
	return v
}

// Middleware assigns a request ID (reusing an inbound X-Request-Id header if
// present) to every request, exposes it via the response header, and stores
// it in the request context for handlers/logging to use.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(Header)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(Header, id)
		ctx := NewContext(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
