// internal/presentation/http/routes/routes.go
package routes

import (
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/farmanexo/api-gateway/internal/infrastructure/proxy"
	"github.com/farmanexo/api-gateway/internal/presentation/http/controllers"
	"github.com/farmanexo/api-gateway/internal/presentation/http/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// ServiceProxies contiene los proxies a los microservicios
type ServiceProxies struct {
	Auth     *proxy.ServiceProxy
	User     *proxy.ServiceProxy
	Catalog  *proxy.ServiceProxy
	Pharmacy *proxy.ServiceProxy
	Price    *proxy.ServiceProxy
	Order    *proxy.ServiceProxy
}

// SetupRoutes configura todas las rutas del API Gateway
func SetupRoutes(
	gatewayController *controllers.GatewayController,
	authMiddleware *middlewares.AuthMiddleware,
	rateLimitMiddleware *middlewares.RateLimitMiddleware,
	cbManager *middlewares.CircuitBreakerManager,
	proxies *ServiceProxies,
	logger *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// ========================================
	// MIDDLEWARES GLOBALES (orden importa)
	// ========================================

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// X-Ray tracing — crea un segment por request. Tras Recoverer para que
	// un panic del handler no rompa el trace. No-op si no hay daemon.
	r.Use(func(next http.Handler) http.Handler {
		return xray.Handler(xray.NewFixedSegmentNamer("api-gateway"), next)
	})

	r.Use(middlewares.CorrelationID)
	r.Use(middlewares.RequestLogger(logger))
	r.Use(middlewares.SecurityHeaders)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"https://farmanexo.pe",
			"https://farmanexo.com.pe",
			"https://*.farmanexo.com.pe",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Correlation-ID"},
		ExposedHeaders:   []string{"Link", "X-Correlation-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting (antes de proxy para proteger downstreams)
	r.Use(rateLimitMiddleware.RateLimit)

	// ========================================
	// HEALTH CHECK (sin auth, sin proxy)
	// ========================================

	r.Get("/health", gatewayController.HealthCheck)
	r.Get("/", gatewayController.HealthCheck)

	// ========================================
	// PASSTHROUGH ROUTES (health checks duales, sin rewrite)
	//
	// Cada microservicio expone /{recurso}/health además de /health, para
	// que el ALB pueda apuntar el target group a una ruta path-based bajo
	// el prefijo permitido. Solo los 9 prefijos de dominio son válidos
	// (services/CLAUDE.md §5.5). NO se permiten top-level como /legal,
	// /admin, /public — son violaciones del contrato del ALB interno.
	// ========================================

	r.Route("/auth", func(r chi.Router) {
		r.Handle("/*", cbManager.Wrap("auth-service", proxies.Auth))
	})

	r.Route("/users", func(r chi.Router) {
		r.Handle("/*", cbManager.Wrap("user-service", proxies.User))
	})

	r.Route("/products", func(r chi.Router) {
		r.Handle("/*", cbManager.Wrap("catalog-service", proxies.Catalog))
	})

	r.Route("/pharmacies", func(r chi.Router) {
		r.Handle("/*", cbManager.Wrap("pharmacy-service", proxies.Pharmacy))
	})

	r.Route("/prices", func(r chi.Router) {
		r.Handle("/*", cbManager.Wrap("price-service", proxies.Price))
	})

	r.Route("/orders", func(r chi.Router) {
		r.Handle("/*", cbManager.Wrap("order-service", proxies.Order))
	})

	// ========================================
	// API ROUTES — Reverse Proxy
	// ========================================

	// ========================================
	// API v1 — Reverse Proxy
	//
	// Modelo de auth: cada microservicio downstream tiene su propio
	// authMiddleware (validación JWT + admin/owner role). El gateway NO
	// es la única autoridad de auth; defensa en profundidad delegada.
	//
	// - /auth/*: rutas públicas (register/login/refresh/legal) y protegidas
	//   (logout/me/*) coexisten; auth-service decide internamente. El gateway
	//   no impone JWT aquí porque rompería login y la página /terminos.
	//
	// - /products, /categories, /brands, /pharmacies, /prices: cada servicio
	//   marca cada endpoint individualmente. Lecturas públicas (GET catálogo,
	//   GET farmacias, comparador de precios) no requieren JWT. Mutaciones
	//   y endpoints user-context (alertas de precio, registro de farmacia,
	//   admin) son protegidos por el authMiddleware del propio servicio.
	//   No hay /api/v1/public/* — el modo de acceso vive en middleware,
	//   no en path (services/CLAUDE.md §5.5).
	//
	// - /users, /cart, /orders: TODAS sus rutas requieren user-context.
	//   El gateway impone JWT acá como gate adicional (valida el token y
	//   inyecta user_id en el contexto antes del proxy).
	// ========================================
	r.Route("/api/v1", func(r chi.Router) {
		// --- Auth (incluye /auth/legal/*) — auth-service decide internamente ---
		r.Route("/auth", func(r chi.Router) {
			r.Handle("/*", cbManager.Wrap("auth-service",
				http.StripPrefix("/api/v1/auth", withPrefix("/api/v1/auth", proxies.Auth)),
			))
		})

		// --- Catálogo y disponibilidad: auth descentralizada en cada servicio ---
		r.Route("/products", func(r chi.Router) {
			r.Handle("/*", cbManager.Wrap("catalog-service",
				http.StripPrefix("/api/v1/products", withPrefix("/api/v1/products", proxies.Catalog)),
			))
		})
		r.Route("/categories", func(r chi.Router) {
			r.Handle("/*", cbManager.Wrap("catalog-service",
				http.StripPrefix("/api/v1/categories", withPrefix("/api/v1/categories", proxies.Catalog)),
			))
		})
		r.Route("/brands", func(r chi.Router) {
			r.Handle("/*", cbManager.Wrap("catalog-service",
				http.StripPrefix("/api/v1/brands", withPrefix("/api/v1/brands", proxies.Catalog)),
			))
		})
		r.Route("/pharmacies", func(r chi.Router) {
			r.Handle("/*", cbManager.Wrap("pharmacy-service",
				http.StripPrefix("/api/v1/pharmacies", withPrefix("/api/v1/pharmacies", proxies.Pharmacy)),
			))
		})
		r.Route("/prices", func(r chi.Router) {
			r.Handle("/*", cbManager.Wrap("price-service",
				http.StripPrefix("/api/v1/prices", withPrefix("/api/v1/prices", proxies.Price)),
			))
		})

		// --- User-context obligatorio (gateway-enforced JWT) ---
		// Todas las rutas bajo estos prefijos requieren user_id del JWT.
		// Catalog/pharmacy/price NO van aquí — su lectura es pública.
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

			r.Route("/users", func(r chi.Router) {
				r.Handle("/*", cbManager.Wrap("user-service",
					http.StripPrefix("/api/v1/users", withPrefix("/api/v1/users", proxies.User)),
				))
			})

			r.Route("/cart", func(r chi.Router) {
				r.Handle("/*", cbManager.Wrap("order-service",
					http.StripPrefix("/api/v1/cart", withPrefix("/api/v1/cart", proxies.Order)),
				))
			})

			r.Route("/orders", func(r chi.Router) {
				r.Handle("/*", cbManager.Wrap("order-service",
					http.StripPrefix("/api/v1/orders", withPrefix("/api/v1/orders", proxies.Order)),
				))
			})
		})
	})

	// ========================================
	// 404 handler
	// ========================================
	r.NotFound(gatewayController.NotFound)
	r.MethodNotAllowed(gatewayController.MethodNotAllowed)

	return r
}

// withPrefix reescribe el path del request para que el downstream reciba el path correcto
func withPrefix(prefix string, sp *proxy.ServiceProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Restaurar el path original que el downstream espera
		r.URL.Path = prefix + r.URL.Path
		if r.URL.Path == prefix+"/" {
			r.URL.Path = prefix
		}
		sp.ServeHTTP(w, r)
	})
}
