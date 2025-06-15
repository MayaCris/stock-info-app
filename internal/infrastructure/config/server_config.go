package config

import (
	"time"
)

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host" validate:"required"`
	Port            string        `mapstructure:"port" validate:"required"`
	Mode            string        `mapstructure:"mode" validate:"required,oneof=debug release test"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" validate:"required"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes" validate:"min=1"`
	TrustedProxies  []string      `mapstructure:"trusted_proxies"`
}

// RESTAPIConfig holds REST API-specific configuration
type RESTAPIConfig struct {
	Version            string `mapstructure:"version" validate:"required"`
	BasePath           string `mapstructure:"base_path" validate:"required"`
	EnableSwagger      bool   `mapstructure:"enable_swagger"`
	EnableHealthChecks bool   `mapstructure:"enable_health_checks"`
	EnableMetrics      bool   `mapstructure:"enable_metrics"`
	EnableProfiling    bool   `mapstructure:"enable_profiling"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	RequestsPer time.Duration `mapstructure:"requests_per" validate:"required"`
	Limit       int           `mapstructure:"limit" validate:"min=1"`
	KeyFunc     string        `mapstructure:"key_func" validate:"oneof=ip user_id"`
}

// GetServerAddress returns the full server address
func (s *ServerConfig) GetServerAddress() string {
	return s.Host + ":" + s.Port
}

// IsDebugMode returns true if server is in debug mode
func (s *ServerConfig) IsDebugMode() bool {
	return s.Mode == "debug"
}

// IsReleaseMode returns true if server is in release mode
func (s *ServerConfig) IsReleaseMode() bool {
	return s.Mode == "release"
}

// IsTestMode returns true if server is in test mode
func (s *ServerConfig) IsTestMode() bool {
	return s.Mode == "test"
}
