package middlewares

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitMiddleware creates a rate limiter middleware
// Limits to 100 requests per minute per IP address
func RateLimitMiddleware() func(next http.Handler) http.Handler {
	return httprate.LimitByIP(100, time.Minute)
}

// StrictRateLimitMiddleware creates a stricter rate limiter
// Limits to 10 requests per minute per IP address
func StrictRateLimitMiddleware() func(next http.Handler) http.Handler {
	return httprate.LimitByIP(10, time.Minute)
}
