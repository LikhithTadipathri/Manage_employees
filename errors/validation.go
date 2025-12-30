package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// ValidationError holds field validation errors
type ValidationError struct {
	Code   int
	Fields map[string]string
}

// Error implements error interface
func (v *ValidationError) Error() string {
	var messages []string
	for field, message := range v.Fields {
		messages = append(messages, fmt.Sprintf("%s: %s", field, message))
	}
	return strings.Join(messages, "; ")
}

// NewValidationError creates a new validation error
func NewValidationError() *ValidationError {
	return &ValidationError{
		Code:   http.StatusBadRequest,
		Fields: make(map[string]string),
	}
}

// AddFieldError adds an error for a specific field
func (v *ValidationError) AddFieldError(field, message string) {
	v.Fields[field] = message
}

// AddField adds a field error and returns the validation error for chaining
func (v *ValidationError) AddField(field, message string) *ValidationError {
	v.Fields[field] = message
	return v
}

// HasErrors checks if there are any validation errors
func (v *ValidationError) HasErrors() bool {
	return len(v.Fields) > 0
}

// Validate checks if validation error has errors and returns it
func (v *ValidationError) Validate() error {
	if v.HasErrors() {
		return v
	}
	return nil
}
