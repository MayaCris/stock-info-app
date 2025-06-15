package logger

import (
	"context"
	"time"
)

// LogLevel representa los niveles de logging
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry representa una entrada de log estructurada
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Context   string                 `json:"context,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// Logger define la interfaz principal para logging
type Logger interface {
	// Métodos de logging por nivel
	Debug(ctx context.Context, message string, fields ...Field)
	Info(ctx context.Context, message string, fields ...Field)
	Warn(ctx context.Context, message string, fields ...Field)
	Error(ctx context.Context, message string, err error, fields ...Field)
	Fatal(ctx context.Context, message string, err error, fields ...Field)

	// Métodos con contexto específico
	WithContext(context string) Logger
	WithFields(fields ...Field) Logger

	// Control del logger
	SetLevel(level LogLevel)
	Close() error
}

// PopulationLogger define métodos específicos para el proceso de población
type PopulationLogger interface {
	Logger

	// Eventos específicos de población
	LogPopulationStart(ctx context.Context, config PopulationConfig)
	LogPopulationEnd(ctx context.Context, result PopulationResult, duration time.Duration)
	LogPageProcessing(ctx context.Context, pageNum, totalPages int, itemCount int)
	LogBatchProcessing(ctx context.Context, batchSize int, operation string)
	LogEntityCreated(ctx context.Context, entityType string, identifier string, fields ...Field)
	LogEntitySkipped(ctx context.Context, entityType string, identifier string, reason string)
	LogEntityError(ctx context.Context, entityType string, identifier string, err error, fields ...Field)
	LogIntegrityValidation(ctx context.Context, status string, issues int, duration time.Duration)
	LogTransactionOperation(ctx context.Context, operation string, retry int, success bool, duration time.Duration)
}

// Field representa un campo de log estructurado
type Field struct {
	Key   string
	Value interface{}
}

// Funciones helper para crear campos
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Uint64(key string, value uint64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value.Format(time.RFC3339)}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Error crear un campo de error
func ErrorField(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// PopulationConfig representa la configuración de población para logging
type PopulationConfig struct {
	BatchSize     int           `json:"batch_size"`
	MaxPages      int           `json:"max_pages"`
	DelayBetween  time.Duration `json:"delay_between"`
	ClearFirst    bool          `json:"clear_first"`
	UseCache      bool          `json:"use_cache"`
	DryRun        bool          `json:"dry_run"`
	ValidateAfter bool          `json:"validate_after"`
}

// PopulationResult representa el resultado de población para logging
type PopulationResult struct {
	TotalPages     int           `json:"total_pages"`     // Páginas con datos procesadas
	PagesRequested int           `json:"pages_requested"` // Total de páginas consultadas (incluyendo vacías)
	TotalItems     int           `json:"total_items"`
	ProcessedItems int           `json:"processed_items"`
	SkippedItems   int           `json:"skipped_items"`
	ErrorCount     int           `json:"error_count"`
	Companies      int           `json:"companies"`
	Brokerages     int           `json:"brokerages"`
	StockRatings   int           `json:"stock_ratings"`
	Duration       time.Duration `json:"duration"`
	Errors         []string      `json:"errors,omitempty"`
}
