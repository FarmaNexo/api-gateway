// internal/presentation/http/middlewares/auth_middleware.go
package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/farmanexo/api-gateway/internal/shared/common"
	"github.com/farmanexo/api-gateway/internal/presentation/dto/responses"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware valida JWT tokens en el gateway
type AuthMiddleware struct {
	jwtSecret []byte
	logger    *zap.Logger
}

// NewAuthMiddleware crea un nuevo middleware de autenticación
func NewAuthMiddleware(jwtSecret string, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
		logger:    logger,
	}
}

// RequireAuth middleware que valida el JWT token
// NO re-valida el token en el servicio downstream — solo en el gateway
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondUnauthorized(w, "Token de autenticación requerido")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			m.respondUnauthorized(w, "Formato de token inválido. Usar: Bearer {token}")
			return
		}

		tokenString := parts[1]

		// Parsear y validar el JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validar algoritmo
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil {
			m.logger.Debug("JWT validation failed",
				zap.Error(err),
				zap.String("path", r.URL.Path),
			)
			m.respondUnauthorized(w, "Token inválido o expirado")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			m.respondUnauthorized(w, "Token inválido")
			return
		}

		// Extraer user_id del claim "sub"
		sub, _ := claims.GetSubject()
		if sub == "" {
			m.respondUnauthorized(w, "Token sin identificador de usuario")
			return
		}

		// Inyectar user_id en el contexto
		ctx := context.WithValue(r.Context(), UserIDKey, sub)

		// Inyectar access token en el contexto (para rate limiting por user)
		ctx = context.WithValue(ctx, AccessTokenKey, tokenString)

		// Pasar el Authorization header al downstream tal cual
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// respondUnauthorized envía una respuesta 401 con formato estándar
func (m *AuthMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	resp := common.UnauthorizedResponse[responses.EmptyResponse](message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(resp)
}
