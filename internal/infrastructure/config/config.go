package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Server        ServerConfig        `mapstructure:"server"`
	RESTAPI       RESTAPIConfig       `mapstructure:"rest_api"`
	CORS          CORSConfig          `mapstructure:"cors"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Cache         CacheConfig         `mapstructure:"cache"`
	External      ExternalConfig      `mapstructure:"external"`
	Security      SecurityConfig      `mapstructure:"security"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	ServerLogging ServerLoggingConfig `mapstructure:"server_logging"`
	ThirdStockAPI ThirdStockAPIConfig `mapstructure:"third_stock_api"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name      string `mapstructure:"name" validate:"required"`
	Version   string `mapstructure:"version" validate:"required"`
	Env       string `mapstructure:"env" validate:"required,oneof=development staging production"`
	Port      string `mapstructure:"port" validate:"required"`
	RateLimit int    `mapstructure:"rate_limit" validate:"min=1"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host" validate:"required"`
	Port            string        `mapstructure:"port" validate:"required"`
	User            string        `mapstructure:"user" validate:"required"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name" validate:"required"`
	SSLMode         string        `mapstructure:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"required"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0"`
	Username string `mapstructure:"username"`
}

// ExternalConfig holds external APIs configuration
type ExternalConfig struct {
	Primary   APIConfig `mapstructure:"primary"`   // Finnhub - Real-time data
	Secondary APIConfig `mapstructure:"secondary"` // Alpha Vantage - Historical analysis
}

// APIConfig holds API configuration
type APIConfig struct {
	Name      string `mapstructure:"name"`
	Key       string `mapstructure:"key" validate:"required"`
	SecretKey string `mapstructure:"secret_key"` // Opcional
	BaseURL   string `mapstructure:"base_url" validate:"required,url"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret string `mapstructure:"jwt_secret" validate:"required,min=16"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format string `mapstructure:"format" validate:"required,oneof=json text"`
}

// ThirdStockAPIConfig holds configuration for a third-party stock API
type ThirdStockAPIConfig struct {
	Name    string `mapstructure:"name" validate:"required"`
	Auth    string `mapstructure:"auth" validate:"required"`
	BaseURL string `mapstructure:"base_url" validate:"required,url"`
}

// GetDSN returns the database connection string for CockroachDB
func (d DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// IsDevelopment returns true if the app is running in development mode
func (a AppConfig) IsDevelopment() bool {
	return a.Env == "development"
}

// IsProduction returns true if the app is running in production mode
func (a AppConfig) IsProduction() bool {
	return a.Env == "production"
}

// GetRedisAddr returns the Redis connection address
func (c CacheConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// GetPrimaryAPI returns the primary API configuration (Finnhub)
func (e ExternalConfig) GetPrimaryAPI() APIConfig {
	return e.Primary
}

// GetSecondaryAPI returns the secondary API configuration (Alpha Vantage)
func (e ExternalConfig) GetSecondaryAPI() APIConfig {
	return e.Secondary
}
