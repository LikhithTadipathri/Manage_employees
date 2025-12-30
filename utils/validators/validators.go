package validators

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError holds validation errors
type ValidationError struct {
	Fields map[string][]string
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Fields: make(map[string][]string),
	}
}

func (ve *ValidationError) Add(field string, message string) *ValidationError {
	ve.Fields[field] = append(ve.Fields[field], message)
	return ve
}

func (ve *ValidationError) HasErrors() bool {
	return len(ve.Fields) > 0
}

func (ve *ValidationError) Error() string {
	var messages []string
	for field, errs := range ve.Fields {
		for _, err := range errs {
			messages = append(messages, fmt.Sprintf("%s: %s", field, err))
		}
	}
	return strings.Join(messages, "; ")
}

// Email validation
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// RFC 5322 simplified pattern
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	if len(email) > 254 {
		return fmt.Errorf("email too long (max 254 characters)")
	}

	return nil
}

// Phone number validation (basic international format)
func ValidatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number is required")
	}

	// Remove common separators
	cleaned := regexp.MustCompile(`[\s\-\(\)\.]+`).ReplaceAllString(phone, "")

	// Must be digits only and 7-15 characters (E.164 standard)
	if !regexp.MustCompile(`^\+?[0-9]{7,15}$`).MatchString(cleaned) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// Name validation
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}

	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}

	if len(name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}

	// Allow letters, spaces, hyphens, apostrophes
	if !regexp.MustCompile(`^[a-zA-Z\s\-']+$`).MatchString(name) {
		return fmt.Errorf("name contains invalid characters")
	}

	return nil
}

// PAN validation (India-specific)
func ValidatePAN(pan string) error {
	if pan == "" {
		return nil // PAN is optional
	}

	// PAN format: AAAAA0000A
	// 5 letters, 4 digits, 1 letter
	if !regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`).MatchString(strings.ToUpper(pan)) {
		return fmt.Errorf("invalid PAN format (expected: AAAAA0000A)")
	}

	return nil
}

// Aadhaar validation (India-specific)
func ValidateAadhaar(aadhaar string) error {
	if aadhaar == "" {
		return nil // Aadhaar is optional
	}

	// Aadhaar is 12 digits
	if !regexp.MustCompile(`^[0-9]{12}$`).MatchString(aadhaar) {
		return fmt.Errorf("invalid Aadhaar format (expected: 12 digits)")
	}

	return nil
}

// Gender validation
func ValidateGender(gender string) error {
	if gender == "" {
		return fmt.Errorf("gender is required")
	}

	valid := map[string]bool{
		"Male":   true,
		"Female": true,
		"Other":  true,
	}

	if !valid[gender] {
		return fmt.Errorf("invalid gender (must be Male, Female, or Other)")
	}

	return nil
}

// Date format validation (YYYY-MM-DD)
func ValidateDateFormat(date string) error {
	if date == "" {
		return fmt.Errorf("date is required")
	}

	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(date) {
		return fmt.Errorf("invalid date format (expected: YYYY-MM-DD)")
	}

	return nil
}

// UUID validation
func ValidateUUID(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("UUID is required")
	}

	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	if !regexp.MustCompile(pattern).MatchString(strings.ToLower(uuid)) {
		return fmt.Errorf("invalid UUID format")
	}

	return nil
}

// Required field validation
func ValidateRequired(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// Min length validation
func ValidateMinLength(value string, minLength int, fieldName string) error {
	if len(value) < minLength {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLength)
	}
	return nil
}

// Max length validation
func ValidateMaxLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, maxLength)
	}
	return nil
}

// Range validation
func ValidateRange(value int, min int, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", fieldName, min, max)
	}
	return nil
}
