package middlewares

import (
	"context"
	"net/http"
	"strings"

	"employee-service/http/response"
)

// roleKey is used for storing role in request context
type roleKey struct{}


func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get authorization header
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		role := "user"
		if token == "admin-token" || token == "admin" {
			role = "admin"
		}

		// attach role to context
		ctx := context.WithValue(r.Context(), roleKey{}, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware is auth middleware that doesn't require auth
// If token provided, role is set; otherwise role is "anonymous".
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		role := "anonymous"

		if authHeader != "" {
			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == "admin-token" || token == "admin" {
				role = "admin"
			} else {
				role = "user"
			}
		}

		ctx := context.WithValue(r.Context(), roleKey{}, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRole returns the role stored in request context ("admin", "user", or "anonymous")
func GetRole(r *http.Request) string {
	if v := r.Context().Value(roleKey{}); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "anonymous"
}
