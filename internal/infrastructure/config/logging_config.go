package config

import (
	"fmt"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// ServerLoggingConfig configuración específica para logging del servidor web
type ServerLoggingConfig struct {
	// Configuración básica
	Enabled bool   `mapstructure:"enabled" json:"enabled"`
	Level   string `mapstructure:"level" json:"level"`
	Format  string `mapstructure:"format" json:"format"`

	// Configuración de componentes
	Middleware MiddlewareLogConfig `mapstructure:"middleware" json:"middleware"`
	Handlers   HandlersLogConfig   `mapstructure:"handlers" json:"handlers"`
	Router     RouterLogConfig     `mapstructure:"router" json:"router"`

	// Configuración de outputs
	Outputs []OutputConfig `mapstructure:"outputs" json:"outputs"`
}

// MiddlewareLogConfig configuración para logging de middleware
type MiddlewareLogConfig struct {
	LogRequests         bool          `mapstructure:"log_requests" json:"log_requests"`
	LogResponses        bool          `mapstructure:"log_responses" json:"log_responses"`
	LogHeaders          bool          `mapstructure:"log_headers" json:"log_headers"`
	LogSlowRequests     bool          `mapstructure:"log_slow_requests" json:"log_slow_requests"`
	SlowThreshold       time.Duration `mapstructure:"slow_threshold" json:"slow_threshold"`
	SkipPaths           []string      `mapstructure:"skip_paths" json:"skip_paths"`
	SkipSuccessfulPaths []string      `mapstructure:"skip_successful_paths" json:"skip_successful_paths"`
}

// HandlersLogConfig configuración para logging de handlers
type HandlersLogConfig struct {
	LogErrors  bool `mapstructure:"log_errors" json:"log_errors"`
	LogPanics  bool `mapstructure:"log_panics" json:"log_panics"`
	LogMetrics bool `mapstructure:"log_metrics" json:"log_metrics"`
}

// RouterLogConfig configuración para logging del router
type RouterLogConfig struct {
	LogRouteRegistration bool `mapstructure:"log_route_registration" json:"log_route_registration"`
	LogStartup           bool `mapstructure:"log_startup" json:"log_startup"`
	LogShutdown          bool `mapstructure:"log_shutdown" json:"log_shutdown"`
}

// OutputConfig configuración de destinos de logs
type OutputConfig struct {
	Type     string            `mapstructure:"type" json:"type"` // "file", "console"
	Level    string            `mapstructure:"level" json:"level"`
	Format   string            `mapstructure:"format" json:"format"`
	Settings map[string]string `mapstructure:"settings" json:"settings"`
}

// DefaultServerLoggingConfig retorna la configuración por defecto para logging del servidor
func DefaultServerLoggingConfig() ServerLoggingConfig {
	return ServerLoggingConfig{
		Enabled: true,
		Level:   "info",
		Format:  "json",

		Middleware: MiddlewareLogConfig{
			LogRequests:         true,
			LogResponses:        true,
			LogHeaders:          true,
			LogSlowRequests:     true,
			SlowThreshold:       1 * time.Second,
			SkipPaths:           []string{"/health", "/metrics", "/favicon.ico"},
			SkipSuccessfulPaths: []string{"/health"},
		},

		Handlers: HandlersLogConfig{
			LogErrors:  true,
			LogPanics:  true,
			LogMetrics: true,
		},

		Router: RouterLogConfig{
			LogRouteRegistration: true,
			LogStartup:           true,
			LogShutdown:          true,
		},

		Outputs: []OutputConfig{
			{
				Type:   "file",
				Level:  "info",
				Format: "json",
				Settings: map[string]string{
					"filename": "logs/server.log",
				},
			},
		},
	}
}

// DevelopmentServerLoggingConfig configuración para entorno de desarrollo
func DevelopmentServerLoggingConfig() ServerLoggingConfig {
	config := DefaultServerLoggingConfig()
	config.Level = "debug"
	config.Format = "text"

	// Agregar salida a consola para desarrollo
	config.Outputs = append(config.Outputs, OutputConfig{
		Type:   "console",
		Level:  "debug",
		Format: "text",
		Settings: map[string]string{
			"color": "true",
		},
	})

	return config
}

// ProductionServerLoggingConfig configuración para entorno de producción
func ProductionServerLoggingConfig() ServerLoggingConfig {
	config := DefaultServerLoggingConfig()
	config.Level = "warn"
	config.Middleware.LogHeaders = false
	config.Middleware.SkipSuccessfulPaths = []string{"/health", "/metrics"}

	return config
}

// ToMiddlewareConfig convierte a configuración de middleware
func (c ServerLoggingConfig) ToMiddlewareConfig() MiddlewareLogConfig {
	return c.Middleware
}

// ToLoggerConfig convierte a configuración del logger base
func (c ServerLoggingConfig) ToLoggerConfig() *logger.LogConfig {
	logConfig := logger.DefaultLogConfig()

	// Mapear nivel
	switch c.Level {
	case "debug":
		logConfig.Level = logger.DebugLevel
	case "info":
		logConfig.Level = logger.InfoLevel
	case "warn":
		logConfig.Level = logger.WarnLevel
	case "error":
		logConfig.Level = logger.ErrorLevel
	default:
		logConfig.Level = logger.InfoLevel
	}

	// Mapear formato
	logConfig.Format = c.Format

	// Configurar outputs
	for _, output := range c.Outputs {
		switch output.Type {
		case "file":
			logConfig.EnableFile = true
			if filename, ok := output.Settings["filename"]; ok {
				logConfig.LogFileName = filename
			}
		case "console":
			logConfig.EnableConsole = true
			if color, ok := output.Settings["color"]; ok && color == "true" {
				logConfig.ColorOutput = true
			}
		}
	}

	return logConfig
}

// Validate valida la configuración
func (c ServerLoggingConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	// Validar que al menos hay un output configurado
	if len(c.Outputs) == 0 {
		return fmt.Errorf("at least one output must be configured when server logging is enabled")
	}

	// Validar niveles válidos
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}

	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	// Validar formatos válidos
	validFormats := map[string]bool{
		"json": true, "text": true,
	}

	if !validFormats[c.Format] {
		return fmt.Errorf("invalid log format: %s", c.Format)
	}

	return nil
}
