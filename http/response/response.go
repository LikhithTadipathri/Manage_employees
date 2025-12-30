package response

import (
	"encoding/json"
	"net/http"
)

// Response is a generic response wrapper
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse is an error response wrapper
type ErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// Success sends a successful response
func Success(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// SuccessNoData sends a successful response without data
func SuccessNoData(w http.ResponseWriter, statusCode int, message string) {
	Success(w, statusCode, nil, message)
}

// Error sends an error response
func Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// ErrorWithFields sends an error response with field-level errors
func ErrorWithFields(w http.ResponseWriter, statusCode int, message string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Message: message,
		Errors:  fields,
	}

	json.NewEncoder(w).Encode(response)
}

// JSONResponse encodes data as JSON and sends it as response
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(data)
}
