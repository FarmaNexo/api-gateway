// pkg/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config representa toda la configuración del API Gateway
type Config struct {
	Environment    string              `mapstructure:"environment"`
	Server         ServerConfig        `mapstructure:"server"`
	Services       ServicesConfig      `mapstructure:"services"`
	JWT            JWTConfig           `mapstructure:"jwt"`
	Redis          RedisConfig         `mapstructure:"redis"`
	RateLimit      RateLimitConfig     `mapstructure:"rate_limit"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
	Log            LogConfig           `mapstructure:"log"`
}

// ServerConfig configuración del servidor HTTP
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// ServicesConfig URLs de los microservicios downstream
type ServicesConfig struct {
	AuthServiceURL     string        `mapstructure:"auth_service_url"`
	UserServiceURL     string        `mapstructure:"user_service_url"`
	CatalogServiceURL  string        `mapstructure:"catalog_service_url"`
	PharmacyServiceURL string        `mapstructure:"pharmacy_service_url"`
	PriceServiceURL    string        `mapstructure:"price_service_url"`
	OrderServiceURL    string        `mapstructure:"order_service_url"`
	DefaultTimeout     time.Duration `mapstructure:"default_timeout"`
}

// JWTConfig configuración de JWT (solo validación, no generación)
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// RedisConfig configuración de Redis
type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

// GetAddr retorna la dirección host:port de Redis
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RateLimitConfig configuración de rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}

// CircuitBreakerConfig configuración del circuit breaker
type CircuitBreakerConfig struct {
	MaxRequests      uint32        `mapstructure:"max_requests"`
	Interval         time.Duration `mapstructure:"interval"`
	Timeout          time.Duration `mapstructure:"timeout"`
	FailureThreshold uint32        `mapstructure:"failure_threshold"`
}

// LogConfig configuración de logging
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

// ========================================
// LOAD CONFIG
// ========================================

// LoadConfig carga la configuración basada en el environment
func LoadConfig(environment string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	configFile := findConfigFile(environment)
	if configFile == "" {
		return nil, fmt.Errorf("config file config.%s.yaml not found", environment)
	}

	// Leer el archivo YAML como string
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configFile, err)
	}

	// Expandir variables de entorno ${VAR} en el YAML
	expanded := os.ExpandEnv(string(content))

	// Pasar el YAML expandido a Viper
	if err := v.ReadConfig(strings.NewReader(expanded)); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	// Unmarshal a struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	config.Environment = environment

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// findConfigFile busca el archivo de configuración en múltiples rutas
func findConfigFile(environment string) string {
	filename := fmt.Sprintf("config.%s.yaml", environment)
	paths := []string{"./configs", "../configs", "../../configs"}

	for _, dir := range paths {
		path := fmt.Sprintf("%s/%s", dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// ========================================
// VALIDATION
// ========================================

func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	if config.Services.AuthServiceURL == "" {
		return fmt.Errorf("auth_service_url is required")
	}

	return nil
}

// ========================================
// HELPERS
// ========================================

// IsProduction verifica si está en producción
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsLocal verifica si está en máquina local
func (c *Config) IsLocal() bool {
	return c.Environment == "local"
}

// IsDevelopment verifica si está en desarrollo (desplegado)
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
