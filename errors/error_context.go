package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorContext provides detailed error information with stack traces
type ErrorContext struct {
	ErrorType   string                 `json:"error_type"`
	Message     string                 `json:"message"`
	Code        string                 `json:"code"`
	StatusCode  int                    `json:"status_code"`
	StackTrace  []StackFrame           `json:"stack_trace"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Cause       string                 `json:"cause,omitempty"`
	Timestamp   string                 `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// StackFrame represents a single stack frame
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

// NewErrorContext creates a new error context with stack trace
func NewErrorContext(errorType, message, code string, statusCode int) *ErrorContext {
	ec := &ErrorContext{
		ErrorType:  errorType,
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
		Context:    make(map[string]interface{}),
		StackTrace: captureStackTrace(3), // Skip NewErrorContext, caller, and one more level
		Timestamp:  getCurrentTimestamp(),
	}
	return ec
}

// WithContext adds contextual information to the error
func (ec *ErrorContext) WithContext(key string, value interface{}) *ErrorContext {
	ec.Context[key] = value
	return ec
}

// WithContextMap adds multiple contextual values
func (ec *ErrorContext) WithContextMap(contextMap map[string]interface{}) *ErrorContext {
	for key, value := range contextMap {
		ec.Context[key] = value
	}
	return ec
}

// WithCause adds the underlying cause of the error
func (ec *ErrorContext) WithCause(cause error) *ErrorContext {
	if cause != nil {
		ec.Cause = cause.Error()
	}
	return ec
}

// WithRequestID adds the request ID for tracing
func (ec *ErrorContext) WithRequestID(requestID string) *ErrorContext {
	ec.RequestID = requestID
	return ec
}

// Error implements the error interface
func (ec *ErrorContext) Error() string {
	return fmt.Sprintf("[%s] %s (Code: %s)", ec.ErrorType, ec.Message, ec.Code)
}

// String returns a detailed string representation
func (ec *ErrorContext) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n=== ERROR CONTEXT ===\n"))
	sb.WriteString(fmt.Sprintf("Type: %s\n", ec.ErrorType))
	sb.WriteString(fmt.Sprintf("Message: %s\n", ec.Message))
	sb.WriteString(fmt.Sprintf("Code: %s\n", ec.Code))
	sb.WriteString(fmt.Sprintf("Status: %d\n", ec.StatusCode))

	if ec.Cause != "" {
		sb.WriteString(fmt.Sprintf("Cause: %s\n", ec.Cause))
	}

	if len(ec.Context) > 0 {
		sb.WriteString(fmt.Sprintf("Context:\n"))
		for key, value := range ec.Context {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	if ec.RequestID != "" {
		sb.WriteString(fmt.Sprintf("Request ID: %s\n", ec.RequestID))
	}

	sb.WriteString(fmt.Sprintf("\nStack Trace:\n"))
	for i, frame := range ec.StackTrace {
		sb.WriteString(fmt.Sprintf("  [%d] %s:%d in %s (%s)\n", i, frame.File, frame.Line, frame.Function, frame.Package))
	}

	return sb.String()
}

// ToJSON returns the error as a JSON-ready map
func (ec *ErrorContext) ToJSON() map[string]interface{} {
	stackTrace := make([]map[string]interface{}, len(ec.StackTrace))
	for i, frame := range ec.StackTrace {
		stackTrace[i] = map[string]interface{}{
			"function": frame.Function,
			"file":     frame.File,
			"line":     frame.Line,
			"package":  frame.Package,
		}
	}

	jsonMap := map[string]interface{}{
		"error_type":  ec.ErrorType,
		"message":     ec.Message,
		"code":        ec.Code,
		"status_code": ec.StatusCode,
		"timestamp":   ec.Timestamp,
		"stack_trace": stackTrace,
	}

	if len(ec.Context) > 0 {
		jsonMap["context"] = ec.Context
	}

	if ec.Cause != "" {
		jsonMap["cause"] = ec.Cause
	}

	if ec.RequestID != "" {
		jsonMap["request_id"] = ec.RequestID
	}

	return jsonMap
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) []StackFrame {
	var frames []StackFrame
	pc := make([]uintptr, 20)
	n := runtime.Callers(skip, pc)

	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pc[i])
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc[i])

		// Skip internal Go runtime frames
		if strings.Contains(file, "runtime/") {
			continue
		}

		frame := StackFrame{
			Function: fn.Name(),
			File:     cleanFilePath(file),
			Line:     line,
			Package:  extractPackage(fn.Name()),
		}

		frames = append(frames, frame)
	}

	return frames
}

// cleanFilePath removes the absolute path, keeping only relative path
func cleanFilePath(filePath string) string {
	// Remove common workspace paths
	separators := []string{
		"\\Go\\src\\Task\\",
		"/go/src/task/",
	}

	for _, sep := range separators {
		if idx := strings.LastIndex(filePath, sep); idx != -1 {
			return filePath[idx+len(sep):]
		}
	}

	// Fallback: return last part of path
	parts := strings.FieldsFunc(filePath, func(r rune) bool { return r == '/' || r == '\\' })
	if len(parts) > 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}

	return filePath
}

// extractPackage extracts the package name from a fully qualified function name
func extractPackage(fullName string) string {
	// Format: package.function or package.(*Type).Method
	parts := strings.Split(fullName, ".")
	if len(parts) > 1 {
		// Remove the main package part
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return "unknown"
}

// getCurrentTimestamp returns current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// ============================================================================
// Common Error Context Creators
// ============================================================================

// ValidationErrorContext creates a validation error context
func ValidationErrorContext(message string, fieldErrors map[string]string) *ErrorContext {
	ec := NewErrorContext("ValidationError", message, "VALIDATION_ERROR", 400)
	ec.WithContext("field_errors", fieldErrors)
	return ec
}

// NotFoundErrorContext creates a not found error context
func NotFoundErrorContext(resource string, identifier interface{}) *ErrorContext {
	message := fmt.Sprintf("%s not found: %v", resource, identifier)
	ec := NewErrorContext("NotFoundError", message, "NOT_FOUND", 404)
	ec.WithContext("resource", resource)
	ec.WithContext("identifier", identifier)
	return ec
}

// UnauthorizedErrorContext creates an unauthorized error context
func UnauthorizedErrorContext(reason string) *ErrorContext {
	ec := NewErrorContext("UnauthorizedError", reason, "UNAUTHORIZED", 401)
	ec.WithContext("reason", reason)
	return ec
}

// ForbiddenErrorContext creates a forbidden error context
func ForbiddenErrorContext(reason string) *ErrorContext {
	ec := NewErrorContext("ForbiddenError", reason, "FORBIDDEN", 403)
	ec.WithContext("reason", reason)
	return ec
}

// ConflictErrorContext creates a conflict error context
func ConflictErrorContext(message string, conflictField string) *ErrorContext {
	ec := NewErrorContext("ConflictError", message, "CONFLICT", 409)
	ec.WithContext("conflict_field", conflictField)
	return ec
}

// DatabaseErrorContext creates a database error context
func DatabaseErrorContext(operation string, cause error) *ErrorContext {
	ec := NewErrorContext("DatabaseError", fmt.Sprintf("Database %s failed", operation), "DATABASE_ERROR", 500)
	ec.WithContext("operation", operation)
	ec.WithCause(cause)
	return ec
}

// ValidationFieldErrors is a helper to create field error map
func ValidationFieldErrorsMap(errors map[string]string) map[string]string {
	return errors
}

// ExternalServiceErrorContext creates an error for external service failures
func ExternalServiceErrorContext(service string, cause error) *ErrorContext {
	ec := NewErrorContext("ExternalServiceError", fmt.Sprintf("%s service unavailable", service), "EXTERNAL_SERVICE_ERROR", 503)
	ec.WithContext("service", service)
	ec.WithCause(cause)
	return ec
}

// RateLimitErrorContext creates a rate limit error context
func RateLimitErrorContext(retryAfter int) *ErrorContext {
	ec := NewErrorContext("RateLimitError", "Too many requests", "RATE_LIMIT_EXCEEDED", 429)
	ec.WithContext("retry_after_seconds", retryAfter)
	return ec
}

// InternalServerErrorContext creates a generic internal server error context
func InternalServerErrorContext(message string, cause error) *ErrorContext {
	ec := NewErrorContext("InternalServerError", message, "INTERNAL_SERVER_ERROR", 500)
	ec.WithCause(cause)
	return ec
}
