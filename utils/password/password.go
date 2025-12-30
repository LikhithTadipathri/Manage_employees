package password

import (
	"fmt"
	"regexp"
	"unicode"
)

const (
	MinLength = 12
	MaxLength = 128
)

// ValidationResult holds password validation details
type ValidationResult struct {
	IsValid bool
	Errors  []string
}

// Validate checks if a password meets security requirements
func Validate(password string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// Check length
	if len(password) < MinLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must be at least %d characters long", MinLength))
	}

	if len(password) > MaxLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must not exceed %d characters", MaxLength))
	}

	// Check for required character types
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		if unicode.IsUpper(ch) {
			hasUpper = true
		}
		if unicode.IsLower(ch) {
			hasLower = true
		}
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && !unicode.IsSpace(ch) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter (A-Z)")
	}

	if !hasLower {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter (a-z)")
	}

	if !hasDigit {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one digit (0-9)")
	}

	if !hasSpecial {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character (!@#$%^&*)")
	}

	// Check for common patterns (weak patterns)
	if isWeakPattern(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password appears to use a weak pattern (e.g., sequential numbers or repeated characters)")
	}

	return result
}

// isWeakPattern checks for common weak password patterns
func isWeakPattern(password string) bool {
	// Check for sequential numbers
	if regexp.MustCompile(`123|234|345|456|567|678|789|890`).MatchString(password) {
		return true
	}

	// Check for repeated characters (more than 3 in a row)
	if regexp.MustCompile(`(.)\1{3,}`).MatchString(password) {
		return true
	}

	// Check for common keyboard patterns
	if regexp.MustCompile(`qwerty|asdfgh|zxcvbn`).MatchString(password) {
		return true
	}

	return false
}

// GetErrorMessage returns a formatted error message
func GetErrorMessage(result *ValidationResult) string {
	if result.IsValid {
		return "Password is strong"
	}

	if len(result.Errors) == 1 {
		return result.Errors[0]
	}

	message := "Password validation failed:\n"
	for i, err := range result.Errors {
		message += fmt.Sprintf("%d. %s\n", i+1, err)
	}
	return message
}
