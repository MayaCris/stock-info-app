package unit

import (
	"testing"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

func TestLoggingConfig_DefaultConfiguration(t *testing.T) {
	cfg := config.DefaultServerLoggingConfig()

	// Verificar configuración básica
	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if cfg.Level != "info" {
		t.Errorf("Expected Level to be 'info', got '%s'", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("Expected Format to be 'json', got '%s'", cfg.Format)
	}

	// Verificar configuración de middleware
	if !cfg.Middleware.LogRequests {
		t.Error("Expected LogRequests to be true")
	}
	if !cfg.Middleware.LogSlowRequests {
		t.Error("Expected LogSlowRequests to be true")
	}
	if cfg.Middleware.SlowThreshold != 1*time.Second {
		t.Errorf("Expected SlowThreshold to be 1s, got %v", cfg.Middleware.SlowThreshold)
	}

	// Verificar que hay al menos un output
	if len(cfg.Outputs) == 0 {
		t.Error("Expected at least one output")
	}
	if len(cfg.Outputs) > 0 && cfg.Outputs[0].Type != "file" {
		t.Errorf("Expected first output type to be 'file', got '%s'", cfg.Outputs[0].Type)
	}
}

func TestLoggingConfig_DevelopmentConfiguration(t *testing.T) {
	cfg := config.DevelopmentServerLoggingConfig()

	if cfg.Level != "debug" {
		t.Errorf("Expected debug level for development, got '%s'", cfg.Level)
	}
	if cfg.Format != "text" {
		t.Errorf("Expected text format for development, got '%s'", cfg.Format)
	}

	// Debe tener al menos 2 outputs (file + console)
	if len(cfg.Outputs) < 2 {
		t.Errorf("Expected at least 2 outputs for development, got %d", len(cfg.Outputs))
	}
}

func TestLoggingConfig_ProductionConfiguration(t *testing.T) {
	cfg := config.ProductionServerLoggingConfig()

	if cfg.Level != "warn" {
		t.Errorf("Expected warn level for production, got '%s'", cfg.Level)
	}
	if cfg.Middleware.LogHeaders {
		t.Error("Expected LogHeaders to be false in production")
	}
}

func TestLoggingConfig_BasicValidation(t *testing.T) {
	// Test disabled config (should be valid)
	disabledConfig := config.ServerLoggingConfig{Enabled: false}
	if err := disabledConfig.Validate(); err != nil {
		t.Errorf("Disabled config should be valid, got error: %v", err)
	}

	// Test valid config
	validConfig := config.ServerLoggingConfig{
		Enabled: true,
		Level:   "info",
		Format:  "json",
		Outputs: []config.OutputConfig{{Type: "file"}},
	}
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should pass validation, got error: %v", err)
	}

	// Test invalid level
	invalidLevelConfig := config.ServerLoggingConfig{
		Enabled: true,
		Level:   "invalid",
		Format:  "json",
		Outputs: []config.OutputConfig{{Type: "file"}},
	}
	if err := invalidLevelConfig.Validate(); err == nil {
		t.Error("Invalid level should fail validation")
	}
}

func TestLoggingConfig_ToLoggerConfigConversion(t *testing.T) {
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
			},
		},
	}

	loggerConfig := serverConfig.ToLoggerConfig()

	if loggerConfig.Format != "text" {
		t.Errorf("Expected format 'text', got '%s'", loggerConfig.Format)
	}
	if !loggerConfig.EnableFile {
		t.Error("Expected file logging to be enabled")
	}
	if !loggerConfig.EnableConsole {
		t.Error("Expected console logging to be enabled")
	}
	if !loggerConfig.ColorOutput {
		t.Error("Expected color output to be enabled")
	}
}

// Benchmark simple para verificar performance
func BenchmarkLoggingConfig_Validate(b *testing.B) {
	cfg := config.DefaultServerLoggingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}

func BenchmarkLoggingConfig_ToLoggerConfig(b *testing.B) {
	cfg := config.DefaultServerLoggingConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.ToLoggerConfig()
	}
}
