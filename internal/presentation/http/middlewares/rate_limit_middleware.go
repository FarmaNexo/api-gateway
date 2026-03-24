// internal/presentation/http/middlewares/rate_limit_middleware.go
package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/farmanexo/api-gateway/internal/shared/common"
	"github.com/farmanexo/api-gateway/internal/presentation/dto/responses"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitMiddleware controla el rate limiting por IP y por JWT sub
type RateLimitMiddleware struct {
	client            *redis.Client
	requestsPerMinute int
	burst             int
	logger            *zap.Logger
}

// NewRateLimitMiddleware crea un nuevo middleware de rate limiting
func NewRateLimitMiddleware(
	client *redis.Client,
	requestsPerMinute int,
	burst int,
	logger *zap.Logger,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		client:            client,
		requestsPerMinute: requestsPerMinute,
		burst:             burst,
		logger:            logger,
	}
}

// RateLimit middleware que limita requests por IP
func (rl *RateLimitMiddleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener identificador: user_id si autenticado, IP si no
		identifier := rl.getIdentifier(r)
		key := fmt.Sprintf("ratelimit:gw:%s", identifier)

		allowed, err := rl.isAllowed(r.Context(), key)
		if err != nil {
			// Si Redis falla, permitir el request (fail-open)
			rl.logger.Warn("Rate limiter error, allowing request",
				zap.Error(err),
				zap.String("identifier", identifier),
			)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			rl.logger.Warn("Rate limit exceeded",
				zap.String("identifier", identifier),
				zap.String("path", r.URL.Path),
			)
			resp := common.TooManyRequestsResponse[responses.EmptyResponse](
				"Demasiadas solicitudes. Intente nuevamente en un momento",
			)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(resp)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAllowed verifica si el request está dentro del límite usando sliding window
func (rl *RateLimitMiddleware) isAllowed(ctx context.Context, key string) (bool, error) {
	pipe := rl.client.Pipeline()

	// Incrementar contador
	incr := pipe.Incr(ctx, key)

	// Establecer TTL solo si es la primera vez
	pipe.Expire(ctx, key, time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := incr.Val()
	return count <= int64(rl.requestsPerMinute), nil
}

// getIdentifier obtiene el identificador para rate limiting
func (rl *RateLimitMiddleware) getIdentifier(r *http.Request) string {
	// Si hay un user_id autenticado, usar eso
	if userID, ok := GetUserIDFromContext(r.Context()); ok {
		return "user:" + userID
	}

	// Sino, usar IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return "ip:" + ip
}
