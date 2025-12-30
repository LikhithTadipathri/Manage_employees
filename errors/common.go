package errors

import (
	"fmt"
	"log"
)

// LogError logs error message with context
func LogError(message string, err error) {
	if err != nil {
		log.Printf("ERROR - %s: %v\n", message, err)
	}
}

// LogInfo logs info message
func LogInfo(message string) {
	log.Printf("INFO - %s\n", message)
}

// LogWarning logs warning message
func LogWarning(message string) {
	log.Printf("WARNING - %s\n", message)
}

// WrapError wraps error with additional context
func WrapError(context string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// NotFoundErrorType represents a not found error (used for type assertions)
type NotFoundErrorType struct {
	Message string
}

func (e *NotFoundErrorType) Error() string {
	return e.Message
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *NotFoundErrorType {
	return &NotFoundErrorType{Message: message}
}

// ForbiddenError represents a forbidden error
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{Message: message}
}
