package helpers

import (
	"strconv"
	"strings"
)

// StringToInt converts string to int with default value
func StringToInt(str string, defaultValue int) int {
	if str == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}

	return val
}

// StringToFloat converts string to float64 with default value
func StringToFloat(str string, defaultValue float64) float64 {
	if str == "" {
		return defaultValue
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return defaultValue
	}

	return val
}

// TrimSpace trims whitespace from string
func TrimSpace(str string) string {
	return strings.TrimSpace(str)
}

// IsEmpty checks if string is empty after trimming
func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// Contains checks if slice contains a value
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
