package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// CorrelationIDMiddleware adds a correlation ID to each request for tracing
func CorrelationIDMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Add to response header
			w.Header().Set("X-Correlation-ID", correlationID)

			// Add to context
			ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return ""
}
