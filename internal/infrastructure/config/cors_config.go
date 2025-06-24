package config

import (
	"time"
)

// CORSConfig holds CORS (Cross-Origin Resource Sharing) configuration
type CORSConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
	AllowWildcard    bool          `mapstructure:"allow_wildcard"`
}

// GetDefaultCORSConfig returns a sensible default CORS configuration
func GetDefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		Enabled: true,
		AllowOrigins: []string{
			"http://localhost:3000", // React development server
			"http://localhost:8080", // Vue development server
			"http://localhost:4200", // Angular development server
			"http://localhost:5173", // Vite development server (your frontend)
			"http://localhost:5174",
			"http://localhost:5175",
			"http://localhost:5176",
			"http://localhost:5177",
			"http://127.0.0.1:5173", // Vite development server (your frontend)
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Response-Time",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowWildcard:    false,
	}
}

// GetProductionCORSConfig returns a production-safe CORS configuration
func GetProductionCORSConfig() *CORSConfig {
	return &CORSConfig{
		Enabled:          true,
		AllowOrigins:     []string{}, // Should be set via environment variables
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-Response-Time"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
		AllowWildcard:    false,
	}
}

// IsOriginAllowed checks if an origin is allowed
func (c *CORSConfig) IsOriginAllowed(origin string) bool {
	if !c.Enabled {
		return false
	}

	if c.AllowWildcard {
		return true
	}

	for _, allowedOrigin := range c.AllowOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}

	return false
}
