package util

import (
	"net/http"
	"strconv"
)

// Pagination holds parsed limit/offset query parameters.
type Pagination struct {
	Limit  int
	Offset int
}

// ParsePagination reads "limit" and "offset" query params from the request,
// applying sane defaults/bounds so a caller can't accidentally request an
// unbounded result set.
func ParsePagination(r *http.Request) Pagination {
	const (
		defaultLimit = 25
		maxLimit     = 200
	)

	limit := defaultLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return Pagination{Limit: limit, Offset: offset}
}
