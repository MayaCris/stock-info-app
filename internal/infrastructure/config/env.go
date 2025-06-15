package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file from multiple locations
	envPaths := []string{
		".env",       // Current directory
		"../.env",    // Parent directory
		"../../.env", // For tests running from test/integration
	}

	envLoaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			break
		}
	}
	// Only fail if we're not in production and no .env file was found
	if !envLoaded && os.Getenv("APP_ENV") != "production" {
		return nil, fmt.Errorf("failed to load .env file from any of the following locations: %v", envPaths)
	}

	config := &Config{
		App:           loadAppConfig(),
		Server:        loadServerConfig(),
		RESTAPI:       loadRESTAPIConfig(),
		CORS:          loadCORSConfig(),
		RateLimit:     loadRateLimitConfig(),
		Database:      loadDatabaseConfig(),
		Cache:         loadCacheConfig(),
		External:      loadExternalConfig(),
		Security:      loadSecurityConfig(),
		Logging:       loadLoggingConfig(),
		ServerLogging: loadServerLoggingConfig(),
		ThirdStockAPI: loadThirdStockAPIConfig(),
	}

	// Validate configuration
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

func loadAppConfig() AppConfig {
	return AppConfig{
		Name:      getEnvRequired("APP_NAME"),
		Version:   getEnvRequired("APP_VERSION"),
		Env:       getEnvRequired("APP_ENV"),
		Port:      getEnvRequired("APP_PORT"),
		RateLimit: getEnvAsIntRequired("API_RATE_LIMIT"),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            getEnvRequired("DB_HOST"),
		Port:            getEnvRequired("DB_PORT"),
		User:            getEnvRequired("DB_USER"),
		Password:        getEnvRequired("DB_PASSWORD"),
		Name:            getEnvRequired("DB_NAME"),
		SSLMode:         getEnvRequired("DB_SSL_MODE"),
		MaxOpenConns:    getEnvAsIntRequired("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    getEnvAsIntRequired("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: getEnvAsDurationRequired("DB_CONN_MAX_LIFETIME"),
	}
}

func loadCacheConfig() CacheConfig {
	return CacheConfig{
		Host:     getEnvRequired("REDIS_HOST"),
		Port:     getEnvRequired("REDIS_PORT"),
		Password: getEnvRequired("REDIS_PASSWORD"),
		Username: getEnvRequired("REDIS_USERNAME"),
		DB:       getEnvAsIntRequired("REDIS_DB"),
	}
}

func loadExternalConfig() ExternalConfig {
	return ExternalConfig{
		Primary: APIConfig{
			Name:      "Finnhub",
			Key:       getEnvRequired("PRIMARY_API_KEY"),
			SecretKey: getEnvRequired("PRIMARY_SECRET_KEY"),
			BaseURL:   getEnvRequired("PRIMARY_API_BASE_URL"),
		},
		Secondary: APIConfig{
			Name:    "Alpha Vantage",
			Key:     getEnvRequired("SECONDARY_API_KEY"),
			BaseURL: getEnvRequired("SECONDARY_API_BASE_URL"),
		},
	}
}

func loadSecurityConfig() SecurityConfig {
	return SecurityConfig{
		JWTSecret: getEnvRequired("JWT_SECRET"),
	}
}

func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:  getEnvRequired("LOG_LEVEL"),
		Format: getEnvRequired("LOG_FORMAT"),
	}
}

func loadServerLoggingConfig() ServerLoggingConfig {
	// Detectar el entorno para usar la configuración apropiada
	env := getEnvWithDefault("APP_ENV", "development")

	switch env {
	case "production":
		return ProductionServerLoggingConfig()
	case "development":
		return DevelopmentServerLoggingConfig()
	default:
		return DefaultServerLoggingConfig()
	}
}

func loadThirdStockAPIConfig() ThirdStockAPIConfig {
	return ThirdStockAPIConfig{
		Name:    "Third Stock API",
		Auth:    getEnvRequired("THIRD_STOCK_API_AUTH"),
		BaseURL: getEnvRequired("THIRD_STOCK_API_BASE_URL"),
	}
}

// loadServerConfig loads server configuration from environment variables
func loadServerConfig() ServerConfig {
	return ServerConfig{
		Host:            getEnvWithDefault("SERVER_HOST", "0.0.0.0"),
		Port:            getEnvWithDefault("SERVER_PORT", "8080"),
		Mode:            getEnvWithDefault("GIN_MODE", "debug"),
		ReadTimeout:     getEnvAsDurationWithDefault("SERVER_READ_TIMEOUT", "30s"),
		WriteTimeout:    getEnvAsDurationWithDefault("SERVER_WRITE_TIMEOUT", "30s"),
		IdleTimeout:     getEnvAsDurationWithDefault("SERVER_IDLE_TIMEOUT", "120s"),
		ShutdownTimeout: getEnvAsDurationWithDefault("SERVER_SHUTDOWN_TIMEOUT", "30s"),
		MaxHeaderBytes:  getEnvAsIntWithDefault("SERVER_MAX_HEADER_BYTES", 1048576), // 1MB
		TrustedProxies:  getEnvAsSlice("SERVER_TRUSTED_PROXIES"),
	}
}

// loadRESTAPIConfig loads REST API configuration from environment variables
func loadRESTAPIConfig() RESTAPIConfig {
	return RESTAPIConfig{
		Version:            getEnvWithDefault("API_VERSION", "v1"),
		BasePath:           getEnvWithDefault("API_BASE_PATH", "/api"),
		EnableSwagger:      getEnvAsBoolWithDefault("API_ENABLE_SWAGGER", true),
		EnableHealthChecks: getEnvAsBoolWithDefault("API_ENABLE_HEALTH_CHECKS", true),
		EnableMetrics:      getEnvAsBoolWithDefault("API_ENABLE_METRICS", false),
		EnableProfiling:    getEnvAsBoolWithDefault("API_ENABLE_PROFILING", false),
	}
}

// loadCORSConfig loads CORS configuration from environment variables
func loadCORSConfig() CORSConfig {
	// Check if specific environment is set, otherwise use defaults
	env := getEnvWithDefault("APP_ENV", "development")

	if env == "production" {
		config := GetProductionCORSConfig()
		// Override with environment variables if provided
		if origins := getEnvAsSlice("CORS_ALLOW_ORIGINS"); len(origins) > 0 {
			config.AllowOrigins = origins
		}
		return *config
	}

	// Development/staging defaults
	config := GetDefaultCORSConfig()
	config.Enabled = getEnvAsBoolWithDefault("CORS_ENABLED", true)
	config.AllowCredentials = getEnvAsBoolWithDefault("CORS_ALLOW_CREDENTIALS", true)
	config.AllowWildcard = getEnvAsBoolWithDefault("CORS_ALLOW_WILDCARD", false)

	if origins := getEnvAsSlice("CORS_ALLOW_ORIGINS"); len(origins) > 0 {
		config.AllowOrigins = origins
	}
	if methods := getEnvAsSlice("CORS_ALLOW_METHODS"); len(methods) > 0 {
		config.AllowMethods = methods
	}
	if headers := getEnvAsSlice("CORS_ALLOW_HEADERS"); len(headers) > 0 {
		config.AllowHeaders = headers
	}

	return *config
}

// loadRateLimitConfig loads rate limiting configuration from environment variables
func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:     getEnvAsBoolWithDefault("RATE_LIMIT_ENABLED", false),
		RequestsPer: getEnvAsDurationWithDefault("RATE_LIMIT_REQUESTS_PER", "1m"),
		Limit:       getEnvAsIntWithDefault("RATE_LIMIT_LIMIT", 100),
		KeyFunc:     getEnvWithDefault("RATE_LIMIT_KEY_FUNC", "ip"),
	}
}

// Helper functions for environment variable parsing

// getEnvRequired gets an environment variable or fails immediately if not found
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("❌ Required environment variable %s is not set", key)
	}
	return value
}

// getEnvAsIntRequired gets a required integer environment variable
func getEnvAsIntRequired(key string) int {
	value := getEnvRequired(key)
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("❌ Environment variable %s must be a valid integer, got: %s", key, value)
	}
	return intValue
}

// getEnvAsDurationRequired gets a required duration environment variable
func getEnvAsDurationRequired(key string) time.Duration {
	value := getEnvRequired(key)
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("❌ Environment variable %s must be a valid duration, got: %s", key, value)
	}
	return duration
}

// Helper functions for loading configuration with defaults

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsBoolWithDefault gets a boolean environment variable with a default value
func getEnvAsBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvAsIntWithDefault gets an integer environment variable with a default value
func getEnvAsIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvAsDurationWithDefault gets a duration environment variable with a default value
func getEnvAsDurationWithDefault(key, defaultValue string) time.Duration {
	value := getEnvWithDefault(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// If parsing fails, parse the default value
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

// getEnvAsSlice gets an environment variable as a comma-separated slice
func getEnvAsSlice(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{}
	}

	// Split by comma and trim spaces
	parts := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
