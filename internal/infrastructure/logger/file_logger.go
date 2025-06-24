package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// FileLogger implementa Logger con escritura a archivos
type FileLogger struct {
	config   *LogConfig
	level    LogLevel
	context  string
	fields   map[string]interface{}
	writers  []io.Writer
	mu       sync.RWMutex
	rotators map[string]*lumberjack.Logger
	closed   bool
}

// NewFileLogger crea un nuevo logger basado en archivos
func NewFileLogger(config *LogConfig) (*FileLogger, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if err := config.EnsureLogDir(); err != nil {
		return nil, err
	}

	logger := &FileLogger{
		config:   config,
		level:    config.Level,
		fields:   make(map[string]interface{}),
		writers:  make([]io.Writer, 0),
		rotators: make(map[string]*lumberjack.Logger),
	}

	// Configurar writers
	if err := logger.setupWriters(); err != nil {
		return nil, err
	}

	return logger, nil
}

// setupWriters configura los writers para archivo y consola
func (l *FileLogger) setupWriters() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Limpiar writers existentes
	l.writers = l.writers[:0]

	// Configurar writer de archivo principal
	if l.config.EnableFile {
		rotator := &lumberjack.Logger{
			Filename:   l.config.GetLogFilePath(),
			MaxSize:    l.config.MaxSize,
			MaxBackups: l.config.MaxBackups,
			MaxAge:     l.config.MaxAge,
			Compress:   l.config.Compress,
		}
		l.rotators["main"] = rotator
		l.writers = append(l.writers, rotator)
	}

	// Configurar writer de consola
	if l.config.EnableConsole {
		l.writers = append(l.writers, os.Stdout)
	}

	return nil
}

// getPopulationWriter retorna el writer para logs de población
func (l *FileLogger) getPopulationWriter() io.Writer {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.config.EnableFile || !l.config.Population.SeparateFile {
		// Usar writers principales
		if len(l.writers) > 0 {
			return io.MultiWriter(l.writers...)
		}
		return os.Stdout
	}

	// Crear/obtener rotator específico para población
	if rotator, exists := l.rotators["population"]; exists {
		writers := []io.Writer{rotator}
		if l.config.EnableConsole {
			writers = append(writers, os.Stdout)
		}
		return io.MultiWriter(writers...)
	}

	// Crear nuevo rotator para población
	rotator := &lumberjack.Logger{
		Filename:   l.config.GetPopulationLogFilePath(),
		MaxSize:    l.config.MaxSize,
		MaxBackups: l.config.MaxBackups,
		MaxAge:     l.config.MaxAge,
		Compress:   l.config.Compress,
	}
	l.rotators["population"] = rotator

	writers := []io.Writer{rotator}
	if l.config.EnableConsole {
		writers = append(writers, os.Stdout)
	}
	return io.MultiWriter(writers...)
}

// log escribe una entrada de log
func (l *FileLogger) log(ctx context.Context, level LogLevel, message string, err error, fields []Field) {
	if level < l.level || l.closed {
		return
	}

	entry := l.createLogEntry(ctx, level, message, err, fields)
	formatted := l.formatLogEntry(entry)

	// Determinar writer apropiado
	var writer io.Writer
	if l.isPopulationContext() {
		writer = l.getPopulationWriter()
	} else {
		l.mu.RLock()
		if len(l.writers) > 0 {
			writer = io.MultiWriter(l.writers...)
		} else {
			writer = os.Stdout
		}
		l.mu.RUnlock()
	}

	// Escribir log
	if _, writeErr := writer.Write([]byte(formatted + "\n")); writeErr != nil {
		// Fallback a stderr en caso de error
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\nOriginal log: %s\n", writeErr, formatted)
	}
}

// createLogEntry crea una entrada de log estructurada
func (l *FileLogger) createLogEntry(ctx context.Context, level LogLevel, message string, err error, fields []Field) *LogEntry {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Context:   l.context,
		Fields:    make(map[string]interface{}),
	}

	// Agregar trace ID del contexto si existe
	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		entry.TraceID = traceID
	}

	// Agregar campos del logger
	l.mu.RLock()
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	l.mu.RUnlock()

	// Agregar campos adicionales
	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}

	// Agregar error si existe
	if err != nil {
		entry.Error = err.Error()
	}

	// Agregar información de caller
	if _, file, line, ok := runtime.Caller(3); ok {
		entry.Fields["caller"] = fmt.Sprintf("%s:%d", file, line)
	}

	return entry
}

// formatLogEntry formatea la entrada según la configuración
func (l *FileLogger) formatLogEntry(entry *LogEntry) string {
	switch l.config.Format {
	case "json":
		return l.formatJSON(entry)
	case "text":
		return l.formatText(entry)
	default:
		return l.formatJSON(entry)
	}
}

// formatJSON formatea como JSON
func (l *FileLogger) formatJSON(entry *LogEntry) string {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf(`{"timestamp":"%s","level":"ERROR","message":"Failed to marshal log entry: %v"}`,
			time.Now().Format(time.RFC3339), err)
	}
	return string(data)
}

// formatText formatea como texto legible
func (l *FileLogger) formatText(entry *LogEntry) string {
	var builder strings.Builder

	// Timestamp y level
	builder.WriteString(entry.Timestamp.Format(l.config.TimeFormat))
	builder.WriteString(" [")
	builder.WriteString(entry.Level.String())
	builder.WriteString("]")

	// Context
	if entry.Context != "" {
		builder.WriteString(" [")
		builder.WriteString(entry.Context)
		builder.WriteString("]")
	}

	// Message
	builder.WriteString(" ")
	builder.WriteString(entry.Message)

	// Error
	if entry.Error != "" {
		builder.WriteString(" error=")
		builder.WriteString(entry.Error)
	}

	// Fields
	if len(entry.Fields) > 0 {
		for k, v := range entry.Fields {
			builder.WriteString(" ")
			builder.WriteString(k)
			builder.WriteString("=")
			builder.WriteString(fmt.Sprintf("%v", v))
		}
	}

	// Trace ID
	if entry.TraceID != "" {
		builder.WriteString(" trace_id=")
		builder.WriteString(entry.TraceID)
	}

	return builder.String()
}

// isPopulationContext determina si el contexto actual es de población
func (l *FileLogger) isPopulationContext() bool {
	return strings.Contains(strings.ToLower(l.context), "population") ||
		strings.Contains(strings.ToLower(l.context), "populate")
}

// Implementación de la interfaz Logger
func (l *FileLogger) Debug(ctx context.Context, message string, fields ...Field) {
	l.log(ctx, DebugLevel, message, nil, fields)
}

func (l *FileLogger) Info(ctx context.Context, message string, fields ...Field) {
	l.log(ctx, InfoLevel, message, nil, fields)
}

func (l *FileLogger) Warn(ctx context.Context, message string, fields ...Field) {
	l.log(ctx, WarnLevel, message, nil, fields)
}

func (l *FileLogger) Error(ctx context.Context, message string, err error, fields ...Field) {
	l.log(ctx, ErrorLevel, message, err, fields)
}

func (l *FileLogger) Fatal(ctx context.Context, message string, err error, fields ...Field) {
	l.log(ctx, FatalLevel, message, err, fields)
	os.Exit(1)
}

func (l *FileLogger) WithContext(context string) Logger {
	l.mu.RLock()
	fields := make(map[string]interface{})
	for k, v := range l.fields {
		fields[k] = v
	}
	l.mu.RUnlock()

	return &FileLogger{
		config:   l.config,
		level:    l.level,
		context:  context,
		fields:   fields,
		writers:  l.writers,
		rotators: l.rotators,
	}
}

func (l *FileLogger) WithFields(fields ...Field) Logger {
	l.mu.RLock()
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	l.mu.RUnlock()

	for _, field := range fields {
		newFields[field.Key] = field.Value
	}

	return &FileLogger{
		config:   l.config,
		level:    l.level,
		context:  l.context,
		fields:   newFields,
		writers:  l.writers,
		rotators: l.rotators,
	}
}

func (l *FileLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true

	// Cerrar todos los rotators
	var lastErr error
	for _, rotator := range l.rotators {
		if err := rotator.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Helper function para extraer trace ID del contexto
func getTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}
