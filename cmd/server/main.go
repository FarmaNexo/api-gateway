// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/farmanexo/api-gateway/internal/infrastructure/cache"
	"github.com/farmanexo/api-gateway/internal/infrastructure/proxy"
	"github.com/farmanexo/api-gateway/internal/presentation/http/controllers"
	"github.com/farmanexo/api-gateway/internal/presentation/http/middlewares"
	"github.com/farmanexo/api-gateway/internal/presentation/http/routes"
	"github.com/farmanexo/api-gateway/pkg/config"

	// Swagger docs
	_ "github.com/farmanexo/api-gateway/docs"

	"go.uber.org/zap"
)

// @title           FarmaNexo API Gateway
// @version         1.0
// @description     API Gateway para FarmaNexo — Reverse proxy hacia microservicios con auth, rate limiting y circuit breaker
// @termsOfService  https://farmanexo.pe/terms

// @contact.name    FarmaNexo API Support
// @contact.url     https://farmanexo.pe/support
// @contact.email   support@farmanexo.pe

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

func main() {
	env := getEnvironment()
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic(fmt.Sprintf("Error cargando configuración: %v", err))
	}

	logger := initLogger(cfg)
	defer logger.Sync()

	logger.Info("Iniciando API Gateway",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	// ========================================
	// REDIS
	// ========================================
	redisClient, err := cache.NewRedisClient(cfg.Redis, cfg.Environment, logger)
	if err != nil {
		logger.Fatal("Error conectando a Redis", zap.Error(err))
	}

	// ========================================
	// SERVICE PROXIES
	// ========================================
	serviceTimeout := cfg.Services.DefaultTimeout
	if serviceTimeout == 0 {
		serviceTimeout = 30 * time.Second
	}

	authProxy, err := proxy.NewServiceProxy("auth-service", cfg.Services.AuthServiceURL, serviceTimeout, logger)
	if err != nil {
		logger.Fatal("Error creando proxy auth-service", zap.Error(err))
	}

	userProxy, err := proxy.NewServiceProxy("user-service", cfg.Services.UserServiceURL, serviceTimeout, logger)
	if err != nil {
		logger.Fatal("Error creando proxy user-service", zap.Error(err))
	}

	catalogProxy, err := proxy.NewServiceProxy("catalog-service", cfg.Services.CatalogServiceURL, serviceTimeout, logger)
	if err != nil {
		logger.Fatal("Error creando proxy catalog-service", zap.Error(err))
	}

	pharmacyProxy, err := proxy.NewServiceProxy("pharmacy-service", cfg.Services.PharmacyServiceURL, serviceTimeout, logger)
	if err != nil {
		logger.Fatal("Error creando proxy pharmacy-service", zap.Error(err))
	}

	priceProxy, err := proxy.NewServiceProxy("price-service", cfg.Services.PriceServiceURL, serviceTimeout, logger)
	if err != nil {
		logger.Fatal("Error creando proxy price-service", zap.Error(err))
	}

	orderProxy, err := proxy.NewServiceProxy("order-service", cfg.Services.OrderServiceURL, serviceTimeout, logger)
	if err != nil {
		logger.Fatal("Error creando proxy order-service", zap.Error(err))
	}

	logger.Info("Service proxies configurados",
		zap.String("auth", cfg.Services.AuthServiceURL),
		zap.String("user", cfg.Services.UserServiceURL),
		zap.String("catalog", cfg.Services.CatalogServiceURL),
		zap.String("pharmacy", cfg.Services.PharmacyServiceURL),
		zap.String("price", cfg.Services.PriceServiceURL),
		zap.String("order", cfg.Services.OrderServiceURL),
	)

	// ========================================
	// CIRCUIT BREAKERS
	// ========================================
	cbManager := middlewares.NewCircuitBreakerManager(logger)

	cbSettings := middlewares.CircuitBreakerSettings{
		MaxRequests:      cfg.CircuitBreaker.MaxRequests,
		Interval:         cfg.CircuitBreaker.Interval,
		Timeout:          cfg.CircuitBreaker.Timeout,
		FailureThreshold: cfg.CircuitBreaker.FailureThreshold,
	}

	serviceNames := []string{
		"auth-service", "user-service", "catalog-service",
		"pharmacy-service", "price-service", "order-service",
	}
	for _, name := range serviceNames {
		cbManager.Register(name, cbSettings)
	}

	logger.Info("Circuit breakers configurados",
		zap.Int("services", len(serviceNames)),
		zap.Uint32("failure_threshold", cbSettings.FailureThreshold),
	)

	// ========================================
	// MIDDLEWARES
	// ========================================
	authMiddleware := middlewares.NewAuthMiddleware(cfg.JWT.Secret, logger)
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(
		redisClient.Client,
		cfg.RateLimit.RequestsPerMinute,
		cfg.RateLimit.Burst,
		logger,
	)

	// ========================================
	// CONTROLLERS Y RUTAS
	// ========================================
	gatewayController := controllers.NewGatewayController(logger)

	proxies := &routes.ServiceProxies{
		Auth:     authProxy,
		User:     userProxy,
		Catalog:  catalogProxy,
		Pharmacy: pharmacyProxy,
		Price:    priceProxy,
		Order:    orderProxy,
	}

	router := routes.SetupRoutes(
		gatewayController,
		authMiddleware,
		rateLimitMiddleware,
		cbManager,
		proxies,
		logger,
	)

	// ========================================
	// SERVIDOR HTTP
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("API Gateway iniciado",
			zap.String("address", server.Addr),
			zap.String("health", fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port)),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error iniciando servidor", zap.Error(err))
		}
	}()

	// ========================================
	// GRACEFUL SHUTDOWN
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Iniciando graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error en shutdown", zap.Error(err))
	}

	if err := redisClient.Close(); err != nil {
		logger.Error("Error cerrando conexión Redis", zap.Error(err))
	}

	logger.Info("API Gateway detenido exitosamente")
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return env
}

func initLogger(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.IsProduction() {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("Error inicializando logger: %v", err))
	}

	return logger
}
