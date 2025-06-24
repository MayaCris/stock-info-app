package logger

import (
	"os"
	"strings"
)

// LoggerFactory crea instancias de logger según el entorno
type LoggerFactory struct{}

// NewLoggerFactory crea una nueva instancia del factory
func NewLoggerFactory() *LoggerFactory {
	return &LoggerFactory{}
}

// CreateLogger crea un logger según el entorno especificado
func (f *LoggerFactory) CreateLogger(environment string) (Logger, error) {
	var config *LogConfig

	switch strings.ToLower(environment) {
	case "development", "dev":
		config = DevelopmentLogConfig()
	case "production", "prod":
		config = ProductionLogConfig()
	case "test", "testing":
		config = TestLogConfig()
	default:
		config = DefaultLogConfig()
	}

	return NewFileLogger(config)
}

// CreateLoggerWithConfig crea un logger con configuración personalizada
func (f *LoggerFactory) CreateLoggerWithConfig(config *LogConfig) (Logger, error) {
	return NewFileLogger(config)
}

// CreatePopulationLogger crea un logger específico para población
func (f *LoggerFactory) CreatePopulationLogger(environment string) (PopulationLogger, error) {
	baseLogger, err := f.CreateLogger(environment)
	if err != nil {
		return nil, err
	}

	var config *LogConfig
	switch strings.ToLower(environment) {
	case "development", "dev":
		config = DevelopmentLogConfig()
	case "production", "prod":
		config = ProductionLogConfig()
	case "test", "testing":
		config = TestLogConfig()
	default:
		config = DefaultLogConfig()
	}

	return NewPopulationLogger(baseLogger, config), nil
}

// CreatePopulationLoggerWithConfig crea un logger de población con configuración personalizada
func (f *LoggerFactory) CreatePopulationLoggerWithConfig(config *LogConfig) (PopulationLogger, error) {
	baseLogger, err := f.CreateLoggerWithConfig(config)
	if err != nil {
		return nil, err
	}

	return NewPopulationLogger(baseLogger, config), nil
}

// CreateServerLogger crea un logger específico para operaciones del servidor web
func (f *LoggerFactory) CreateServerLogger(environment string) (ServerLogger, error) {
	baseLogger, err := f.CreateLogger(environment)
	if err != nil {
		return nil, err
	}

	var config *LogConfig
	switch strings.ToLower(environment) {
	case "development", "dev":
		config = DevelopmentLogConfig()
	case "production", "prod":
		config = ProductionLogConfig()
	case "test", "testing":
		config = TestLogConfig()
	default:
		config = DefaultLogConfig()
	}

	return NewServerLogger(baseLogger, config), nil
}

// CreateServerLoggerWithConfig crea un logger de servidor con configuración personalizada
func (f *LoggerFactory) CreateServerLoggerWithConfig(config *LogConfig) (ServerLogger, error) {
	baseLogger, err := f.CreateLoggerWithConfig(config)
	if err != nil {
		return nil, err
	}

	return NewServerLogger(baseLogger, config), nil
}

// TestLogConfig retorna configuración para testing
func TestLogConfig() *LogConfig {
	config := DefaultLogConfig()
	config.Level = DebugLevel
	config.EnableConsole = false
	config.EnableFile = true
	config.LogDir = "test_logs"
	config.LogFileName = "test.log"
	config.Population.SeparateFile = false
	config.Population.LogProgress = false
	config.Population.LogTransactions = false
	config.Population.LogCacheOps = false
	config.Population.LogIntegrityCheck = false
	return config
}

// GetEnvironmentFromEnv determina el entorno desde variables de entorno
func GetEnvironmentFromEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = os.Getenv("GO_ENV")
	}
	if env == "" {
		env = "development" // default
	}
	return env
}

// InitializeGlobalLogger inicializa un logger global
func InitializeGlobalLogger() (Logger, error) {
	factory := NewLoggerFactory()
	environment := GetEnvironmentFromEnv()
	return factory.CreateLogger(environment)
}

// InitializePopulationLogger inicializa un logger de población
func InitializePopulationLogger() (PopulationLogger, error) {
	factory := NewLoggerFactory()
	environment := GetEnvironmentFromEnv()
	return factory.CreatePopulationLogger(environment)
}

// LoggerBuilder permite construir loggers con configuración fluida
type LoggerBuilder struct {
	config *LogConfig
}

// NewLoggerBuilder crea un nuevo builder
func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		config: DefaultLogConfig(),
	}
}

// WithLevel establece el nivel de log
func (b *LoggerBuilder) WithLevel(level LogLevel) *LoggerBuilder {
	b.config.Level = level
	return b
}

// WithFormat establece el formato de log
func (b *LoggerBuilder) WithFormat(format string) *LoggerBuilder {
	b.config.Format = format
	return b
}

// WithFileOutput habilita salida a archivo
func (b *LoggerBuilder) WithFileOutput(enabled bool) *LoggerBuilder {
	b.config.EnableFile = enabled
	return b
}

// WithConsoleOutput habilita salida a consola
func (b *LoggerBuilder) WithConsoleOutput(enabled bool) *LoggerBuilder {
	b.config.EnableConsole = enabled
	return b
}

// WithLogDir establece el directorio de logs
func (b *LoggerBuilder) WithLogDir(dir string) *LoggerBuilder {
	b.config.LogDir = dir
	return b
}

// WithLogFile establece el nombre del archivo de log
func (b *LoggerBuilder) WithLogFile(filename string) *LoggerBuilder {
	b.config.LogFileName = filename
	return b
}

// WithRotation configura la rotación de archivos
func (b *LoggerBuilder) WithRotation(maxSize, maxBackups, maxAge int, compress bool) *LoggerBuilder {
	b.config.MaxSize = maxSize
	b.config.MaxBackups = maxBackups
	b.config.MaxAge = maxAge
	b.config.Compress = compress
	b.config.EnableRotation = true
	return b
}

// WithPopulationConfig configura opciones específicas de población
func (b *LoggerBuilder) WithPopulationConfig(separateFile bool, detailedLogging bool) *LoggerBuilder {
	b.config.Population.SeparateFile = separateFile
	b.config.Population.LogProgress = detailedLogging
	b.config.Population.LogTransactions = detailedLogging
	b.config.Population.LogIntegrityCheck = detailedLogging
	return b
}

// Build construye el logger
func (b *LoggerBuilder) Build() (Logger, error) {
	return NewFileLogger(b.config)
}

// BuildPopulation construye un logger de población
func (b *LoggerBuilder) BuildPopulation() (PopulationLogger, error) {
	baseLogger, err := b.Build()
	if err != nil {
		return nil, err
	}
	return NewPopulationLogger(baseLogger, b.config), nil
}

// Validate valida la configuración
func (b *LoggerBuilder) Validate() error {
	return b.config.Validate()
}

// GetConfig retorna la configuración actual
func (b *LoggerBuilder) GetConfig() *LogConfig {
	return b.config
}

// Quick builders para casos comunes
func NewDevelopmentLogger() (Logger, error) {
	return NewLoggerBuilder().
		WithLevel(DebugLevel).
		WithFormat("text").
		WithFileOutput(true).
		WithConsoleOutput(true).
		WithLogDir("logs").
		WithLogFile("development.log").
		Build()
}

func NewProductionLogger() (Logger, error) {
	return NewLoggerBuilder().
		WithLevel(InfoLevel).
		WithFormat("json").
		WithFileOutput(true).
		WithConsoleOutput(false).
		WithLogDir("logs").
		WithLogFile("application.log").
		WithRotation(500, 10, 30, true).
		Build()
}

// Helper para migrar desde log estándar
func ReplaceStandardLogger(logger Logger) {
	// Esta función puede ser usada para reemplazar el logger estándar
	// en casos donde se necesite migrar código existente
	if logger == nil {
		return
	}
	// TODO: Implementar wrapper para log estándar si es necesario
}
