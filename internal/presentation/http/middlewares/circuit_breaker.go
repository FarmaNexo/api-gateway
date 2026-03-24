// internal/presentation/http/middlewares/circuit_breaker.go
package middlewares

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/farmanexo/api-gateway/internal/shared/common"
	"github.com/farmanexo/api-gateway/internal/presentation/dto/responses"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
)

// CircuitBreakerManager gestiona circuit breakers por servicio
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker[struct{}]
	logger   *zap.Logger
}

// CircuitBreakerSettings configuración para el circuit breaker
type CircuitBreakerSettings struct {
	MaxRequests      uint32
	Interval         time.Duration
	Timeout          time.Duration
	FailureThreshold uint32
}

// NewCircuitBreakerManager crea un nuevo manager de circuit breakers
func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker[struct{}]),
		logger:   logger,
	}
}

// Register registra un circuit breaker para un servicio
func (cbm *CircuitBreakerManager) Register(serviceName string, settings CircuitBreakerSettings) {
	cbm.breakers[serviceName] = gobreaker.NewCircuitBreaker[struct{}](gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: settings.MaxRequests,
		Interval:    settings.Interval,
		Timeout:     settings.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= settings.FailureThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			cbm.logger.Warn("Circuit breaker state change",
				zap.String("service", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})
}

// Wrap envuelve un handler con circuit breaker protection
func (cbm *CircuitBreakerManager) Wrap(serviceName string, handler http.Handler) http.Handler {
	cb, exists := cbm.breakers[serviceName]
	if !exists {
		// Sin circuit breaker configurado, pasar directo
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := cb.Execute(func() (struct{}, error) {
			// Usar un responseWriter que captura el status
			rw := &circuitResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			handler.ServeHTTP(rw, r)

			// Si el downstream devolvió 5xx, considerarlo como fallo
			if rw.statusCode >= 500 {
				return struct{}{}, &upstreamError{statusCode: rw.statusCode}
			}
			return struct{}{}, nil
		})

		if err != nil {
			// Circuit breaker abierto
			if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
				cbm.logger.Warn("Circuit breaker open",
					zap.String("service", serviceName),
					zap.String("path", r.URL.Path),
				)
				resp := common.ServiceUnavailableResponse[responses.EmptyResponse](
					"Servicio temporalmente deshabilitado por fallos consecutivos. Reintente en unos segundos",
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(resp)
				return
			}
			// Otros errores ya fueron manejados por el handler (errores de proxy)
		}
	})
}

// circuitResponseWriter captura el status code para el circuit breaker
type circuitResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (crw *circuitResponseWriter) WriteHeader(code int) {
	if !crw.written {
		crw.statusCode = code
		crw.written = true
		crw.ResponseWriter.WriteHeader(code)
	}
}

// upstreamError error del servicio upstream
type upstreamError struct {
	statusCode int
}

func (e *upstreamError) Error() string {
	return "upstream error"
}
