// internal/infrastructure/cache/redis_client.go
package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/farmanexo/api-gateway/pkg/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient envuelve el cliente de Redis con logging
type RedisClient struct {
	Client *redis.Client
	logger *zap.Logger
}

// NewRedisClient crea una nueva conexión a Redis y verifica conectividad
func NewRedisClient(cfg config.RedisConfig, environment string, logger *zap.Logger) (*RedisClient, error) {
	opts := &redis.Options{
		Addr:       cfg.GetAddr(),
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: cfg.MaxRetries,
		PoolSize:   cfg.PoolSize,
	}

	// TLS requerido en ambientes AWS (development, production)
	if environment != "local" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		logger.Info("Redis TLS habilitado", zap.String("environment", environment))
	}

	client := redis.NewClient(opts)

	// Verificar conexión con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error conectando a Redis (%s): %w", cfg.GetAddr(), err)
	}

	logger.Info("Conexión a Redis establecida",
		zap.String("addr", cfg.GetAddr()),
		zap.Int("db", cfg.DB),
		zap.Int("pool_size", cfg.PoolSize),
	)

	return &RedisClient{
		Client: client,
		logger: logger,
	}, nil
}

// Close cierra la conexión a Redis
func (r *RedisClient) Close() error {
	r.logger.Info("Cerrando conexión a Redis")
	return r.Client.Close()
}
