package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	// Load .env file if it exists (ignore error in production)
	if err := godotenv.Load(); err != nil && os.Getenv("APP_ENV") != "production" {
		// In development, we want to know if .env file is missing
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	config := &Config{
		App:      loadAppConfig(),
		Database: loadDatabaseConfig(),
		Cache:    loadCacheConfig(),
		External: loadExternalConfig(),
		Security: loadSecurityConfig(),
		Logging:  loadLoggingConfig(),
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
