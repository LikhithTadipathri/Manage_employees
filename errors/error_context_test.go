package errors_test

import (
	"testing"

	"employee-service/errors"
)

// TestErrorContext tests error context creation and details
func TestErrorContext(t *testing.T) {
	t.Run("Create error context with stack trace", func(t *testing.T) {
		ec := errors.NewErrorContext("TestError", "This is a test error", "TEST_ERROR", 400)

		if ec.ErrorType != "TestError" {
			t.Errorf("Expected ErrorType 'TestError', got %s", ec.ErrorType)
		}

		if ec.Code != "TEST_ERROR" {
			t.Errorf("Expected Code 'TEST_ERROR', got %s", ec.Code)
		}

		if ec.StatusCode != 400 {
			t.Errorf("Expected StatusCode 400, got %d", ec.StatusCode)
		}

		if len(ec.StackTrace) == 0 {
			t.Error("Expected StackTrace to be populated")
		}
	})

	t.Run("Add context information", func(t *testing.T) {
		ec := errors.NewErrorContext("Test", "Test", "TEST", 400)
		ec.WithContext("user_id", 123)
		ec.WithContext("action", "create")

		if ec.Context["user_id"] != 123 {
			t.Error("Expected context to be set")
		}

		if ec.Context["action"] != "create" {
			t.Error("Expected context to be set")
		}
	})

	t.Run("Add context map", func(t *testing.T) {
		ec := errors.NewErrorContext("Test", "Test", "TEST", 400)
		contextMap := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}
		ec.WithContextMap(contextMap)

		if ec.Context["field1"] != "value1" {
			t.Error("Expected context map to be set")
		}
	})

	t.Run("Add cause error", func(t *testing.T) {
		causeErr := errors.NewAppError(500, "Underlying error", errors.ErrInternalServer)
		ec := errors.NewErrorContext("Test", "Test", "TEST", 400)
		ec.WithCause(causeErr)

		if ec.Cause == "" {
			t.Error("Expected cause to be set")
		}
	})

	t.Run("Add request ID", func(t *testing.T) {
		ec := errors.NewErrorContext("Test", "Test", "TEST", 400)
		ec.WithRequestID("req-123-456")

		if ec.RequestID != "req-123-456" {
			t.Errorf("Expected RequestID 'req-123-456', got %s", ec.RequestID)
		}
	})

	t.Run("Error interface implementation", func(t *testing.T) {
		ec := errors.NewErrorContext("ValidationError", "Invalid email", "VALIDATION_ERROR", 400)
		errMsg := ec.Error()

		if errMsg == "" {
			t.Error("Expected Error() to return non-empty string")
		}

		// Should contain error type and code
		if !contains(errMsg, "ValidationError") || !contains(errMsg, "VALIDATION_ERROR") {
			t.Errorf("Expected error message to contain type and code, got: %s", errMsg)
		}
	})

	t.Run("String representation", func(t *testing.T) {
		ec := errors.NewErrorContext("Test", "Test message", "TEST", 400)
		str := ec.String()

		if len(str) == 0 {
			t.Error("Expected String() to return non-empty value")
		}

		// Should contain detailed information
		if !contains(str, "ERROR CONTEXT") || !contains(str, "Test message") {
			t.Error("Expected String() to contain detailed information")
		}
	})

	t.Run("ToJSON conversion", func(t *testing.T) {
		ec := errors.NewErrorContext("Test", "Test", "TEST", 400)
		ec.WithContext("user_id", 123)

		jsonMap := ec.ToJSON()

		if jsonMap["error_type"] != "Test" {
			t.Error("Expected error_type in JSON")
		}

		if jsonMap["code"] != "TEST" {
			t.Error("Expected code in JSON")
		}

		if jsonMap["status_code"] != 400 {
			t.Error("Expected status_code in JSON")
		}

		if jsonMap["context"] == nil {
			t.Error("Expected context in JSON")
		}
	})
}

// TestErrorContextHelpers tests helper functions
func TestErrorContextHelpers(t *testing.T) {
	t.Run("ValidationErrorContext", func(t *testing.T) {
		fieldErrors := map[string]string{
			"email": "Invalid email format",
			"phone": "Invalid phone number",
		}
		ec := errors.ValidationErrorContext("Validation failed", fieldErrors)

		if ec.Code != "VALIDATION_ERROR" {
			t.Errorf("Expected Code 'VALIDATION_ERROR', got %s", ec.Code)
		}

		if ec.StatusCode != 400 {
			t.Errorf("Expected StatusCode 400, got %d", ec.StatusCode)
		}
	})

	t.Run("NotFoundErrorContext", func(t *testing.T) {
		ec := errors.NotFoundErrorContext("Employee", 123)

		if ec.Code != "NOT_FOUND" {
			t.Error("Expected Code 'NOT_FOUND'")
		}

		if ec.StatusCode != 404 {
			t.Error("Expected StatusCode 404")
		}
	})

	t.Run("UnauthorizedErrorContext", func(t *testing.T) {
		ec := errors.UnauthorizedErrorContext("Invalid token")

		if ec.Code != "UNAUTHORIZED" {
			t.Error("Expected Code 'UNAUTHORIZED'")
		}

		if ec.StatusCode != 401 {
			t.Error("Expected StatusCode 401")
		}
	})

	t.Run("ForbiddenErrorContext", func(t *testing.T) {
		ec := errors.ForbiddenErrorContext("Access denied")

		if ec.Code != "FORBIDDEN" {
			t.Error("Expected Code 'FORBIDDEN'")
		}

		if ec.StatusCode != 403 {
			t.Error("Expected StatusCode 403")
		}
	})

	t.Run("ConflictErrorContext", func(t *testing.T) {
		ec := errors.ConflictErrorContext("Email already exists", "email")

		if ec.Code != "CONFLICT" {
			t.Error("Expected Code 'CONFLICT'")
		}

		if ec.StatusCode != 409 {
			t.Error("Expected StatusCode 409")
		}
	})

	t.Run("DatabaseErrorContext", func(t *testing.T) {
		underlyingErr := errors.ErrInternalServer
		ec := errors.DatabaseErrorContext("INSERT", underlyingErr)

		if ec.Code != "DATABASE_ERROR" {
			t.Error("Expected Code 'DATABASE_ERROR'")
		}

		if ec.StatusCode != 500 {
			t.Error("Expected StatusCode 500")
		}
	})

	t.Run("RateLimitErrorContext", func(t *testing.T) {
		ec := errors.RateLimitErrorContext(60)

		if ec.Code != "RATE_LIMIT_EXCEEDED" {
			t.Error("Expected Code 'RATE_LIMIT_EXCEEDED'")
		}

		if ec.StatusCode != 429 {
			t.Error("Expected StatusCode 429")
		}

		if ec.Context["retry_after_seconds"] != 60 {
			t.Error("Expected retry_after_seconds in context")
		}
	})
}

// Helper function
func contains(str, substr string) bool {
	for i := 0; i < len(str)-len(substr)+1; i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
