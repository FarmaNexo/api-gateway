// internal/presentation/http/middlewares/request_logger.go
package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter captura el status code de la respuesta
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// RequestLogger middleware que logea cada request con correlation ID y duración
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			correlationID := GetCorrelationID(r.Context())

			logger.Info("Request entrante",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("correlation_id", correlationID),
				zap.String("user_agent", r.UserAgent()),
			)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.Info("Request completado",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.statusCode),
				zap.Duration("duration", duration),
				zap.String("correlation_id", correlationID),
			)
		})
	}
}
