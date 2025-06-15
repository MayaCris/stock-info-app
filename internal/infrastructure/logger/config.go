package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogConfig contiene la configuración del sistema de logging
type LogConfig struct {
	// Configuración general
	Level      LogLevel `json:"level"`
	Format     string   `json:"format"`      // "json" or "text"
	TimeFormat string   `json:"time_format"` // RFC3339, RFC3339Nano, etc.

	// Configuración de archivos
	EnableFile     bool   `json:"enable_file"`
	LogDir         string `json:"log_dir"`
	LogFileName    string `json:"log_file_name"`
	MaxSize        int    `json:"max_size"`        // MB
	MaxBackups     int    `json:"max_backups"`     // número de archivos de respaldo
	MaxAge         int    `json:"max_age"`         // días
	Compress       bool   `json:"compress"`        // comprimir archivos antiguos
	EnableRotation bool   `json:"enable_rotation"` // rotación automática

	// Configuración de consola
	EnableConsole bool `json:"enable_console"`
	ColorOutput   bool `json:"color_output"`

	// Configuración específica de población
	Population PopulationLogConfig `json:"population"`
}

// PopulationLogConfig configuración específica para logs de población
type PopulationLogConfig struct {
	SeparateFile      bool   `json:"separate_file"`       // archivo separado para población
	FileName          string `json:"file_name"`           // nombre del archivo específico
	LogProgress       bool   `json:"log_progress"`        // log de progreso detallado
	LogTransactions   bool   `json:"log_transactions"`    // log de operaciones transaccionales
	LogCacheOps       bool   `json:"log_cache_ops"`       // log de operaciones de cache
	LogIntegrityCheck bool   `json:"log_integrity_check"` // log de verificaciones de integridad
	BatchLogInterval  int    `json:"batch_log_interval"`  // intervalo para log de lotes
}

// DefaultLogConfig retorna la configuración por defecto
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      InfoLevel,
		Format:     "json",
		TimeFormat: time.RFC3339,

		// Archivo
		EnableFile:     true,
		LogDir:         "logs",
		LogFileName:    "application.log",
		MaxSize:        100, // 100MB
		MaxBackups:     5,
		MaxAge:         30, // 30 días
		Compress:       true,
		EnableRotation: true,

		// Consola
		EnableConsole: false, // Deshabilitado por defecto para producción
		ColorOutput:   false,

		// Población
		Population: PopulationLogConfig{
			SeparateFile:      true,
			FileName:          "database_population.log",
			LogProgress:       true,
			LogTransactions:   true,
			LogCacheOps:       false, // Solo en debug
			LogIntegrityCheck: true,
			BatchLogInterval:  100, // cada 100 items
		},
	}
}

// DevelopmentLogConfig retorna configuración para desarrollo
func DevelopmentLogConfig() *LogConfig {
	config := DefaultLogConfig()
	config.Level = DebugLevel
	config.EnableConsole = true
	config.ColorOutput = true
	config.Format = "text"
	config.Population.LogCacheOps = true
	return config
}

// ProductionLogConfig retorna configuración para producción
func ProductionLogConfig() *LogConfig {
	config := DefaultLogConfig()
	config.Level = InfoLevel
	config.EnableConsole = false
	config.ColorOutput = false
	config.Format = "json"
	config.MaxSize = 500 // 500MB para producción
	config.MaxBackups = 10
	config.Population.LogCacheOps = false
	return config
}

// Validate valida la configuración
func (c *LogConfig) Validate() error {
	if c.EnableFile {
		if c.LogDir == "" {
			return fmt.Errorf("log_dir cannot be empty when file logging is enabled")
		}
		if c.LogFileName == "" {
			return fmt.Errorf("log_file_name cannot be empty when file logging is enabled")
		}
		if c.MaxSize <= 0 {
			return fmt.Errorf("max_size must be greater than 0")
		}
	}

	if c.Format != "json" && c.Format != "text" {
		return fmt.Errorf("format must be either 'json' or 'text'")
	}

	if !c.EnableFile && !c.EnableConsole {
		return fmt.Errorf("at least one output (file or console) must be enabled")
	}

	return nil
}

// EnsureLogDir crea el directorio de logs si no existe
func (c *LogConfig) EnsureLogDir() error {
	if !c.EnableFile {
		return nil
	}

	if err := os.MkdirAll(c.LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", c.LogDir, err)
	}

	return nil
}

// GetLogFilePath retorna la ruta completa del archivo de log principal
func (c *LogConfig) GetLogFilePath() string {
	if !c.EnableFile {
		return ""
	}
	return filepath.Join(c.LogDir, c.LogFileName)
}

// GetPopulationLogFilePath retorna la ruta del archivo de log de población
func (c *LogConfig) GetPopulationLogFilePath() string {
	if !c.EnableFile || !c.Population.SeparateFile {
		return c.GetLogFilePath()
	}
	return filepath.Join(c.LogDir, c.Population.FileName)
}

// ShouldLogProgress determina si se debe registrar el progreso
func (c *LogConfig) ShouldLogProgress() bool {
	return c.Population.LogProgress
}

// ShouldLogTransactions determina si se deben registrar las transacciones
func (c *LogConfig) ShouldLogTransactions() bool {
	return c.Population.LogTransactions
}

// ShouldLogCacheOps determina si se deben registrar las operaciones de cache
func (c *LogConfig) ShouldLogCacheOps() bool {
	return c.Population.LogCacheOps
}

// ShouldLogIntegrityCheck determina si se deben registrar las verificaciones de integridad
func (c *LogConfig) ShouldLogIntegrityCheck() bool {
	return c.Population.LogIntegrityCheck
}
