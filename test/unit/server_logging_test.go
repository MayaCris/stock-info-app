package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

func TestServerLoggingConfig_DefaultConfiguration(t *testing.T) {
	cfg := config.DefaultServerLoggingConfig()

	// Verificar configuración básica
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)

	// Verificar configuración de middleware
	assert.True(t, cfg.Middleware.LogRequests)
	assert.True(t, cfg.Middleware.LogResponses)
	assert.True(t, cfg.Middleware.LogHeaders)
	assert.True(t, cfg.Middleware.LogSlowRequests)
	assert.Equal(t, 1*time.Second, cfg.Middleware.SlowThreshold)
	assert.Contains(t, cfg.Middleware.SkipPaths, "/health")
	assert.Contains(t, cfg.Middleware.SkipPaths, "/metrics")

	// Verificar configuración de handlers
	assert.True(t, cfg.Handlers.LogErrors)
	assert.True(t, cfg.Handlers.LogPanics)
	assert.True(t, cfg.Handlers.LogMetrics)

	// Verificar configuración de router
	assert.True(t, cfg.Router.LogRouteRegistration)
	assert.True(t, cfg.Router.LogStartup)
	assert.True(t, cfg.Router.LogShutdown)
	// Verificar outputs
	require.Len(t, cfg.Outputs, 1)
	assert.Equal(t, "file", cfg.Outputs[0].Type)
	assert.Equal(t, "info", cfg.Outputs[0].Level)
	assert.Equal(t, "json", cfg.Outputs[0].Format)
}

func TestServerLoggingConfig_DevelopmentConfiguration(t *testing.T) {
	cfg := config.DevelopmentServerLoggingConfig()

	// Verificar nivel de debug
	assert.Equal(t, "debug", cfg.Level)
	assert.Equal(t, "text", cfg.Format)

	// Debe tener más de un output (file + console)
	assert.GreaterOrEqual(t, len(cfg.Outputs), 2)

	// Verificar que hay output de consola
	var hasConsoleOutput bool
	for _, output := range cfg.Outputs {
		if output.Type == "console" {
			hasConsoleOutput = true
			assert.Equal(t, "debug", output.Level)
			assert.Equal(t, "text", output.Format)
			assert.Equal(t, "true", output.Settings["color"])
		}
	}
	assert.True(t, hasConsoleOutput, "Development config should have console output")
}

func TestServerLoggingConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  config.ServerLoggingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "disabled config is valid",
			config:  config.ServerLoggingConfig{Enabled: false},
			wantErr: false,
		},
		{
			name: "valid config",
			config: config.ServerLoggingConfig{
				Enabled: true,
				Level:   "info",
				Format:  "json",
				Outputs: []config.OutputConfig{{Type: "file"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerLoggingConfig_DevelopmentEnvironment(t *testing.T) {
	cfg := config.DevelopmentServerLoggingConfig()

	// Verificar nivel de debug
	assert.Equal(t, "debug", cfg.Level)
	assert.Equal(t, "text", cfg.Format)

	// Debe tener más de un output (file + console)
	assert.GreaterOrEqual(t, len(cfg.Outputs), 2)

	// Verificar que hay output de consola
	var hasConsoleOutput bool
	for _, output := range cfg.Outputs {
		if output.Type == "console" {
			hasConsoleOutput = true
			assert.Equal(t, "debug", output.Level)
			assert.Equal(t, "text", output.Format)
			assert.Equal(t, "true", output.Settings["color"])
		}
	}
	assert.True(t, hasConsoleOutput, "Development config should have console output")
}

func TestServerLoggingConfig_ProductionConfiguration(t *testing.T) {
	cfg := config.ProductionServerLoggingConfig()

	// Verificar nivel más restrictivo
	assert.Equal(t, "warn", cfg.Level)

	// Headers no deberían loggearse en producción
	assert.False(t, cfg.Middleware.LogHeaders)

	// Más paths deberían saltarse en producción
	assert.Contains(t, cfg.Middleware.SkipSuccessfulPaths, "/health")
	assert.Contains(t, cfg.Middleware.SkipSuccessfulPaths, "/metrics")
}

func TestServerLoggingConfig_AdvancedValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  config.ServerLoggingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "disabled config is valid",
			config:  config.ServerLoggingConfig{Enabled: false},
			wantErr: false,
		},
		{
			name: "valid config",
			config: config.ServerLoggingConfig{
				Enabled: true,
				Level:   "info",
				Format:  "json",
				Outputs: []config.OutputConfig{{Type: "file"}},
			},
			wantErr: false,
		},
		{
			name: "no outputs when enabled",
			config: config.ServerLoggingConfig{
				Enabled: true,
				Level:   "info",
				Format:  "json",
				Outputs: []config.OutputConfig{},
			},
			wantErr: true,
			errMsg:  "at least one output must be configured",
		},
		{
			name: "invalid log level",
			config: config.ServerLoggingConfig{
				Enabled: true,
				Level:   "invalid",
				Format:  "json",
				Outputs: []config.OutputConfig{{Type: "file"}},
			},
			wantErr: true,
			errMsg:  "invalid log level: invalid",
		},
		{
			name: "invalid format",
			config: config.ServerLoggingConfig{
				Enabled: true,
				Level:   "info",
				Format:  "invalid",
				Outputs: []config.OutputConfig{{Type: "file"}},
			},
			wantErr: true,
			errMsg:  "invalid log format: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerLoggingConfig_ToLoggerConfig(t *testing.T) {
	serverConfig := config.ServerLoggingConfig{
		Level:  "debug",
		Format: "text",
		Outputs: []config.OutputConfig{
			{
				Type:     "file",
				Settings: map[string]string{"filename": "custom.log"},
			},
			{
				Type:     "console",
				Settings: map[string]string{"color": "true"},
			}},
	}

	loggerConfig := serverConfig.ToLoggerConfig()

	// Verificar conversión de nivel
	assert.Equal(t, "DEBUG", loggerConfig.Level.String())

	// Verificar formato
	assert.Equal(t, "text", loggerConfig.Format)

	// Verificar que file está habilitado
	assert.True(t, loggerConfig.EnableFile)
	assert.Equal(t, "custom.log", loggerConfig.LogFileName)

	// Verificar que console está habilitado
	assert.True(t, loggerConfig.EnableConsole)
	assert.True(t, loggerConfig.ColorOutput)
}

func TestServerLoggingConfig_ToMiddlewareConfig(t *testing.T) {
	serverConfig := config.ServerLoggingConfig{
		Middleware: config.MiddlewareLogConfig{
			LogRequests:         true,
			LogResponses:        false,
			LogHeaders:          true,
			LogSlowRequests:     true,
			SlowThreshold:       2 * time.Second,
			SkipPaths:           []string{"/test"},
			SkipSuccessfulPaths: []string{"/health"},
		},
	}

	middlewareConfig := serverConfig.ToMiddlewareConfig()

	assert.True(t, middlewareConfig.LogRequests)
	assert.False(t, middlewareConfig.LogResponses)
	assert.True(t, middlewareConfig.LogHeaders)
	assert.True(t, middlewareConfig.LogSlowRequests)
	assert.Equal(t, 2*time.Second, middlewareConfig.SlowThreshold)
	assert.Equal(t, []string{"/test"}, middlewareConfig.SkipPaths)
	assert.Equal(t, []string{"/health"}, middlewareConfig.SkipSuccessfulPaths)
}

func TestOutputConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config config.OutputConfig
		valid  bool
	}{
		{
			name:   "valid file output",
			config: config.OutputConfig{Type: "file", Level: "info", Format: "json"},
			valid:  true,
		},
		{
			name:   "valid console output",
			config: config.OutputConfig{Type: "console", Level: "debug", Format: "text"},
			valid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Esta sería la lógica de validación si la implementáramos
			assert.NotEmpty(t, tt.config.Type)
			assert.NotEmpty(t, tt.config.Level)
			assert.NotEmpty(t, tt.config.Format)
		})
	}
}

// Benchmarks para verificar performance
func BenchmarkServerLoggingConfig_Validate(b *testing.B) {
	cfg := config.DefaultServerLoggingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}

func BenchmarkServerLoggingConfig_ToLoggerConfig(b *testing.B) {
	cfg := config.DefaultServerLoggingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.ToLoggerConfig()
	}
}

func BenchmarkServerLoggingConfig_ToMiddlewareConfig(b *testing.B) {
	cfg := config.DefaultServerLoggingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.ToMiddlewareConfig()
	}
}
