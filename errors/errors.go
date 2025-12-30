package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Custom error types
var (
	ErrNotFound       = errors.New("resource not found")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInternalServer = errors.New("internal server error")
	ErrConflict       = errors.New("resource already exists")
	ErrInvalidInput   = errors.New("invalid input")
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error implements error interface
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new AppError
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// GetHTTPStatusCode returns appropriate HTTP status code for error
func GetHTTPStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}

	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrBadRequest, ErrInvalidInput:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// GetErrorMessage returns appropriate error message
func GetErrorMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}

	switch err {
	case ErrNotFound:
		return "Resource not found"
	case ErrBadRequest:
		return "Bad request"
	case ErrUnauthorized:
		return "Unauthorized"
	case ErrConflict:
		return "Resource already exists"
	case ErrInvalidInput:
		return "Invalid input provided"
	default:
		return "Internal server error"
	}
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *AppError {
	return NewAppError(http.StatusNotFound, fmt.Sprintf("%s not found", resource), ErrNotFound)
}

// BadRequestError creates a bad request error
func BadRequestError(message string) *AppError {
	return NewAppError(http.StatusBadRequest, message, ErrBadRequest)
}

// InternalServerError creates an internal server error
func InternalServerError(message string) *AppError {
	return NewAppError(http.StatusInternalServerError, message, ErrInternalServer)
}

// ConflictError creates a conflict error
func ConflictError(resource string) *AppError {
	return NewAppError(http.StatusConflict, fmt.Sprintf("%s already exists", resource), ErrConflict)
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *AppError {
	return NewAppError(http.StatusUnauthorized, message, ErrUnauthorized)
}