package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// ErrorLoggingConfig configura el comportamiento del middleware de error logging
type ErrorLoggingConfig struct {
	LogStackTrace       bool `json:"log_stack_trace"`
	LogRequestDetails   bool `json:"log_request_details"`
	LogPanicRecovery    bool `json:"log_panic_recovery"`
	StackTraceSkipLines int  `json:"stack_trace_skip_lines"`
}

// DefaultErrorLoggingConfig retorna la configuración por defecto
func DefaultErrorLoggingConfig() ErrorLoggingConfig {
	return ErrorLoggingConfig{
		LogStackTrace:       true,
		LogRequestDetails:   true,
		LogPanicRecovery:    true,
		StackTraceSkipLines: 3, // Skip middleware stack frames
	}
}

// ServerErrorMiddleware middleware avanzado para manejo y logging de errores usando ServerLogger
func ServerErrorMiddleware(serverLogger logger.ServerLogger, config ErrorLoggingConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				handlePanicRecovery(c, serverLogger, err, config)
			}
		}()

		c.Next()

		// Procesar errores generados durante la request
		if len(c.Errors) > 0 {
			handleRequestErrors(c, serverLogger, config)
		}
	})
}

// handlePanicRecovery maneja la recuperación de panics
func handlePanicRecovery(c *gin.Context, serverLogger logger.ServerLogger, recovered interface{}, config ErrorLoggingConfig) {
	ctx := c.Request.Context()

	// Información básica del panic
	panicStr := fmt.Sprintf("%v", recovered)

	// Obtener stack trace si está habilitado
	var stackTrace string
	if config.LogStackTrace {
		buf := make([]byte, 1024*8) // 8KB buffer for stack trace
		n := runtime.Stack(buf, false)
		stackTrace = string(buf[:n])

		// Limpiar stack trace removiendo líneas irrelevantes
		if config.StackTraceSkipLines > 0 {
			stackTrace = cleanStackTrace(stackTrace, config.StackTraceSkipLines)
		}
	}

	// Crear información de request para el panic
	requestInfo := buildPanicRequestInfo(c, config)

	// Log del panic con ServerLogger
	if config.LogPanicRecovery {
		securityDetails := map[string]interface{}{
			"panic_value":  panicStr,
			"request_info": requestInfo,
			"timestamp":    time.Now(),
		}

		if stackTrace != "" {
			securityDetails["stack_trace"] = stackTrace
		}

		serverLogger.LogSecurityEvent(
			ctx,
			"panic_recovery",
			c.ClientIP(),
			securityDetails,
		)
	}

	// Responder con error 500
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":     "Internal Server Error",
		"message":   "An unexpected error occurred",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	// Abortar la cadena de middleware
	c.Abort()
}

// handleRequestErrors procesa errores generados durante la request
func handleRequestErrors(c *gin.Context, serverLogger logger.ServerLogger, config ErrorLoggingConfig) {
	ctx := c.Request.Context()

	for _, ginError := range c.Errors {
		// Crear información de request para el error
		var requestInfo map[string]interface{}
		if config.LogRequestDetails {
			requestInfo = buildErrorRequestInfo(c)
		}

		// Determinar el tipo de error y su severidad
		errorType := determineErrorType(ginError, c.Writer.Status())

		// Crear detalles del error
		errorDetails := map[string]interface{}{
			"error_message": ginError.Error(),
			"error_type":    errorType,
			"status_code":   c.Writer.Status(),
			"timestamp":     time.Now(),
		}

		if requestInfo != nil {
			errorDetails["request_info"] = requestInfo
		}

		// Log el error basado en su tipo
		switch errorType {
		case "validation_error", "bad_request":
			serverLogger.Warn(ctx, "Request validation error",
				logger.String("error", ginError.Error()),
				logger.Int("status_code", c.Writer.Status()),
				logger.Any("details", errorDetails),
			)
		case "authorization_error", "forbidden":
			serverLogger.LogSecurityEvent(
				ctx,
				"authorization_error",
				c.ClientIP(),
				errorDetails,
			)
		case "server_error", "internal_error":
			// Para errores del servidor, usar HTTPError con stack trace si está disponible
			if config.LogStackTrace {
				errorDetails["stack_trace"] = getStackTraceForError()
			}

			requestInfo := logger.HTTPRequestInfo{
				RequestID:     getRequestIDFromContext(ctx),
				Method:        c.Request.Method,
				Path:          c.Request.URL.Path,
				ClientIP:      c.ClientIP(),
				UserAgent:     c.Request.UserAgent(),
				ContentType:   c.GetHeader("Content-Type"),
				ContentLength: c.Request.ContentLength,
				Timestamp:     time.Now(),
			}

			serverLogger.LogHTTPError(ctx, requestInfo, ginError, c.Writer.Status())
		default:
			serverLogger.Error(ctx, "Unhandled request error", ginError,
				logger.Int("status_code", c.Writer.Status()),
				logger.Any("details", errorDetails),
			)
		}
	}
}

// buildPanicRequestInfo construye información de request para panics
func buildPanicRequestInfo(c *gin.Context, config ErrorLoggingConfig) map[string]interface{} {
	if !config.LogRequestDetails {
		return nil
	}

	return map[string]interface{}{
		"method":         c.Request.Method,
		"path":           c.Request.URL.Path,
		"query":          c.Request.URL.RawQuery,
		"client_ip":      c.ClientIP(),
		"user_agent":     c.Request.UserAgent(),
		"content_type":   c.GetHeader("Content-Type"),
		"content_length": c.Request.ContentLength,
		"headers":        extractSafeHeaders(c.Request.Header),
		"request_id":     getRequestIDFromGin(c),
	}
}

// buildErrorRequestInfo construye información de request para errores
func buildErrorRequestInfo(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"query":        c.Request.URL.RawQuery,
		"client_ip":    c.ClientIP(),
		"user_agent":   c.Request.UserAgent(),
		"content_type": c.GetHeader("Content-Type"),
		"request_id":   getRequestIDFromGin(c),
		"status_code":  c.Writer.Status(),
	}
}

// determineErrorType determina el tipo de error basado en el código de estado
func determineErrorType(ginError *gin.Error, statusCode int) string {
	switch {
	case statusCode >= 400 && statusCode < 500:
		switch statusCode {
		case 400:
			return "bad_request"
		case 401:
			return "authorization_error"
		case 403:
			return "forbidden"
		case 404:
			return "not_found"
		case 422:
			return "validation_error"
		case 429:
			return "rate_limit"
		default:
			return "client_error"
		}
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown_error"
	}
}

// cleanStackTrace limpia el stack trace removiendo líneas irrelevantes
func cleanStackTrace(stackTrace string, skipLines int) string {
	lines := strings.Split(stackTrace, "\n")

	// Encontrar donde empiezan las líneas relevantes
	start := 0
	skipped := 0
	for i, line := range lines {
		if strings.Contains(line, "panic(") ||
			strings.Contains(line, "middleware") {
			skipped++
			if skipped >= skipLines {
				start = i
				break
			}
		}
	}

	if start < len(lines) {
		return strings.Join(lines[start:], "\n")
	}

	return stackTrace
}

// getStackTraceForError obtiene stack trace para errores (no panics)
func getStackTraceForError() string {
	buf := make([]byte, 1024*4) // 4KB buffer
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// extractSafeHeaders extrae headers seguros para logging (sin información sensible)
func extractSafeHeaders(headers map[string][]string) map[string]interface{} {
	safeHeaders := map[string]interface{}{}

	sensitiveHeaders := map[string]bool{
		"authorization":       true,
		"cookie":              true,
		"set-cookie":          true,
		"x-api-key":           true,
		"x-auth-token":        true,
		"proxy-authorization": true,
	}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)

		if sensitiveHeaders[lowerKey] {
			safeHeaders[key] = "[REDACTED]"
		} else {
			if len(values) == 1 {
				safeHeaders[key] = values[0]
			} else {
				safeHeaders[key] = values
			}
		}
	}

	return safeHeaders
}

// getRequestIDFromGin obtiene el request ID del contexto de Gin
func getRequestIDFromGin(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	// Fallback a headers
	return c.GetHeader("X-Request-ID")
}

// getRequestIDFromContext obtiene el request ID del contexto
func getRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// AdvancedRecoveryMiddleware middleware avanzado de recovery con logging detallado
func AdvancedRecoveryMiddleware(serverLogger logger.ServerLogger) gin.HandlerFunc {
	return ServerErrorMiddleware(serverLogger, DefaultErrorLoggingConfig())
}

// CustomErrorResponse estructura para respuestas de error personalizadas
type CustomErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Timestamp string                 `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponseMiddleware middleware para estandarizar las respuestas de error
func ErrorResponseMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Solo procesar si hay errores y no se ha escrito respuesta aún
		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last()

			// Determinar código de estado basado en el tipo de error
			statusCode := c.Writer.Status()
			if statusCode == 200 {
				statusCode = http.StatusInternalServerError
			}

			response := CustomErrorResponse{
				Error:     http.StatusText(statusCode),
				Message:   err.Error(),
				Timestamp: time.Now().Format(time.RFC3339),
				RequestID: getRequestIDFromGin(c),
			}

			c.JSON(statusCode, response)
		}
	})
}
