// internal/presentation/http/middlewares/security_headers.go
package middlewares

import (
	"net/http"
)

// SecurityHeaders agrega headers de seguridad estándar a todas las respuestas
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevenir MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevenir clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Protección XSS del navegador
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// HSTS (solo en HTTPS)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// No cachear respuestas de la API por defecto
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Eliminar header Server si lo puso Go
		w.Header().Del("Server")

		next.ServeHTTP(w, r)
	})
}
