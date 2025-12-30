package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"employee-service/http/response"
	usermodel "employee-service/models/user"
	"employee-service/utils/jwt"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const userClaimsKey contextKey = "user_claims"

// JWTMiddleware validates JWT tokens
func JWTMiddleware(jwtMgr *jwt.JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			// Extract Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Verify token
			claims, err := jwtMgr.ExtractClaims(tokenString)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), userClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole checks if the user has the required role
func RequireRole(requiredRole string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get claims from context
		claims, ok := r.Context().Value(userClaimsKey).(*usermodel.JWTClaims)
		if !ok || claims == nil {
			response.Error(w, http.StatusUnauthorized, "user claims not found")
			return
		}			// Check role
			if claims.Role != requiredRole && requiredRole != "" {
				response.Error(w, http.StatusForbidden, "insufficient permissions for this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(r *http.Request) (*usermodel.JWTClaims, error) {
	claims, ok := r.Context().Value(userClaimsKey).(*usermodel.JWTClaims)
	if !ok || claims == nil {
		return nil, errors.New("user not authenticated")
	}
	return claims, nil
}

// OptionalJWTMiddleware validates JWT tokens but allows unauthenticated requests
func OptionalJWTMiddleware(jwtMgr *jwt.JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			
			if authHeader != "" {
				// Extract Bearer token
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString := parts[1]

					// Verify token
					claims, err := jwtMgr.ExtractClaims(tokenString)
					if err == nil {
						// Add claims to request context
						ctx := context.WithValue(r.Context(), userClaimsKey, claims)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// No token or invalid token, continue without authentication
			next.ServeHTTP(w, r)
		})
	}
}
