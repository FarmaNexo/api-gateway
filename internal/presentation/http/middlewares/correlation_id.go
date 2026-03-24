// internal/presentation/http/middlewares/correlation_id.go
package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	UserIDKey        contextKey = "user_id"
	AccessTokenKey   contextKey = "access_token"
)

// CorrelationID es un middleware que agrega correlation ID a cada request
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Agregar al header de respuesta
		w.Header().Set("X-Correlation-ID", correlationID)

		// Agregar al request header para que lo propague el proxy
		r.Header.Set("X-Correlation-ID", correlationID)

		// Agregar al contexto
		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCorrelationID obtiene el correlation ID del contexto
func GetCorrelationID(ctx context.Context) string {
	if val := ctx.Value(CorrelationIDKey); val != nil {
		if corrID, ok := val.(string); ok {
			return corrID
		}
	}
	return ""
}

// GetUserIDFromContext obtiene el user ID del contexto
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	if val := ctx.Value(UserIDKey); val != nil {
		if userID, ok := val.(string); ok {
			return userID, true
		}
	}
	return "", false
}
