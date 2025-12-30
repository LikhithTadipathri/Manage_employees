package middlewares

import (
	"log"
	"net/http"
	"time"
)

// LoggerMiddleware logs incoming HTTP requests
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		log.Printf(
			"[%s] %s %s - Status: %d - Duration: %v\n",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter is a wrapper around http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
