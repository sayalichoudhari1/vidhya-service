// Package apperrors defines a small set of API-facing error codes shared by
// all Vidhya Service handlers, so every endpoint returns a consistent JSON
// error shape:
//
//	{ "error": { "code": "NOT_FOUND", "message": "student not found" } }
package apperrors

import "net/http"

// AppError is a structured error carrying an HTTP status and a stable code.
type AppError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// New creates an AppError with a custom message.
func New(httpStatus int, code, message string) *AppError {
	return &AppError{HTTPStatus: httpStatus, Code: code, Message: message}
}

// Predefined, reusable errors covering the common REST scenarios.
var (
	ErrBadRequest     = New(http.StatusBadRequest, "BAD_REQUEST", "the request could not be understood")
	ErrUnauthorized   = New(http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
	ErrForbidden      = New(http.StatusForbidden, "FORBIDDEN", "you do not have permission to perform this action")
	ErrNotFound       = New(http.StatusNotFound, "NOT_FOUND", "the requested resource was not found")
	ErrConflict       = New(http.StatusConflict, "CONFLICT", "the resource already exists")
	ErrInternal       = New(http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	ErrInvalidPayload = New(http.StatusBadRequest, "INVALID_PAYLOAD", "the request payload is invalid")
)

// WithMessage returns a copy of the base error with a more specific message.
func WithMessage(base *AppError, message string) *AppError {
	return &AppError{HTTPStatus: base.HTTPStatus, Code: base.Code, Message: message}
}
