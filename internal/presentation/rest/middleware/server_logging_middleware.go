package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// ContextKey define el tipo para keys de contexto para evitar colisiones
type ContextKey string

const (
	// RequestIDKey es la key para el request ID en el contexto
	RequestIDKey ContextKey = "request_id"
)

// ServerLoggingConfig configura el comportamiento del middleware de logging del servidor
type ServerLoggingConfig struct {
	LogHeaders           bool          `json:"log_headers"`
	LogRequestBody       bool          `json:"log_request_body"`
	LogResponseBody      bool          `json:"log_response_body"`
	SkipPaths            []string      `json:"skip_paths"`
	SkipSuccessfulPaths  []string      `json:"skip_successful_paths"`
	LogSlowRequests      bool          `json:"log_slow_requests"`
	SlowRequestThreshold time.Duration `json:"slow_request_threshold"`
	SensitiveHeaders     []string      `json:"sensitive_headers"`
}

// DefaultServerLoggingConfig retorna la configuración por defecto para el middleware
func DefaultServerLoggingConfig() ServerLoggingConfig {
	return ServerLoggingConfig{
		LogHeaders:           true,
		LogRequestBody:       false, // Por seguridad, no logear bodies por defecto
		LogResponseBody:      false,
		SkipPaths:            []string{"/health", "/metrics", "/favicon.ico"},
		SkipSuccessfulPaths:  []string{"/health"},
		LogSlowRequests:      true,
		SlowRequestThreshold: 1 * time.Second,
		SensitiveHeaders: []string{
			"authorization", "cookie", "set-cookie", "x-api-key",
			"x-auth-token", "proxy-authorization", "www-authenticate",
		},
	}
}

// ServerLoggingMiddleware crea un middleware avanzado de logging usando ServerLogger
func ServerLoggingMiddleware(serverLogger logger.ServerLogger, config ServerLoggingConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Verificar si debemos skip este path
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Generar o obtener request ID
		requestID := getOrGenerateRequestID(c)

		// Crear contexto con request ID
		ctx := createContextWithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Preparar información de la request
		requestInfo := buildHTTPRequestInfo(c, requestID, path, rawQuery, config)

		// Log de request start
		serverLogger.LogHTTPRequest(ctx, requestInfo)

		// Process request
		c.Next()

		// Calcular duración y preparar información de respuesta
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		// Verificar si debemos skip logging de respuestas exitosas para ciertos paths
		if shouldSkipSuccessfulResponse(statusCode, path, config.SkipSuccessfulPaths) {
			return
		}

		// Preparar información de la response
		responseInfo := buildHTTPResponseInfo(c, requestID, statusCode, responseSize, duration, config)

		// Determinar si hay errores
		hasError := len(c.Errors) > 0 || statusCode >= 400

		// Log según el tipo de respuesta
		if hasError {
			var err error
			if len(c.Errors) > 0 {
				err = c.Errors.Last()
			}
			serverLogger.LogHTTPError(ctx, requestInfo, err, statusCode)
		} else {
			serverLogger.LogHTTPResponse(ctx, responseInfo)
		}

		// Log de requests lentas si está habilitado
		if config.LogSlowRequests && duration > config.SlowRequestThreshold {
			logSlowRequest(ctx, serverLogger, requestInfo, responseInfo, duration)
		}
	})
}

// getOrGenerateRequestID obtiene o genera un request ID único
func getOrGenerateRequestID(c *gin.Context) string {
	// Intentar obtener de headers
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = c.GetHeader("X-Correlation-ID")
	}
	if requestID == "" {
		requestID = c.GetHeader("X-Trace-ID")
	}

	// Si no existe, generar uno nuevo
	if requestID == "" {
		requestID = uuid.New().String()
	}

	// Almacenar en el contexto de Gin
	c.Set("request_id", requestID)

	return requestID
}

// createContextWithRequestID crea un contexto con el request ID
func createContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// buildHTTPRequestInfo construye la estructura de información de request
func buildHTTPRequestInfo(c *gin.Context, requestID string, path string, rawQuery string, config ServerLoggingConfig) logger.HTTPRequestInfo {
	fullPath := path
	if rawQuery != "" {
		fullPath = path + "?" + rawQuery
	}

	requestInfo := logger.HTTPRequestInfo{
		RequestID:     requestID,
		Method:        c.Request.Method,
		Path:          fullPath,
		Query:         rawQuery,
		ClientIP:      c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		ContentType:   c.GetHeader("Content-Type"),
		ContentLength: c.Request.ContentLength,
		Timestamp:     time.Now(),
	}

	// Agregar headers si está habilitado
	if config.LogHeaders {
		requestInfo.Headers = filterSensitiveHeaders(c.Request.Header, config.SensitiveHeaders)
	}

	return requestInfo
}

// buildHTTPResponseInfo construye la estructura de información de response
func buildHTTPResponseInfo(c *gin.Context, requestID string, statusCode int, responseSize int, duration time.Duration, config ServerLoggingConfig) logger.HTTPResponseInfo {
	responseInfo := logger.HTTPResponseInfo{
		RequestID:    requestID,
		StatusCode:   statusCode,
		ResponseSize: responseSize,
		Duration:     duration,
		ContentType:  c.Writer.Header().Get("Content-Type"),
	}

	// Agregar headers si está habilitado
	if config.LogHeaders {
		responseInfo.Headers = convertHeaderToMap(c.Writer.Header())
	}

	// Agregar información de error si existe
	if len(c.Errors) > 0 {
		errorMessages := make([]string, len(c.Errors))
		for i, err := range c.Errors {
			errorMessages[i] = err.Error()
		}
		responseInfo.ErrorMessage = strings.Join(errorMessages, "; ")
	}

	// Obtener handler name si está disponible
	if handlerName := c.HandlerName(); handlerName != "" {
		responseInfo.HandlerName = handlerName
	}

	return responseInfo
}

// filterSensitiveHeaders filtra headers sensibles para logging seguro
func filterSensitiveHeaders(headers map[string][]string, sensitiveHeaders []string) map[string]string {
	filtered := make(map[string]string)

	for key, values := range headers {
		lowerKey := strings.ToLower(key)

		// Verificar si es un header sensible
		isSensitive := false
		for _, sensitive := range sensitiveHeaders {
			if strings.ToLower(sensitive) == lowerKey {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			filtered[key] = "[REDACTED]"
		} else {
			if len(values) == 1 {
				filtered[key] = values[0]
			} else if len(values) > 1 {
				filtered[key] = strings.Join(values, ", ")
			}
		}
	}

	return filtered
}

// convertHeaderToMap convierte headers de response a map
func convertHeaderToMap(headers map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) == 1 {
			result[key] = values[0]
		} else if len(values) > 1 {
			result[key] = strings.Join(values, ", ")
		}
	}
	return result
}

// shouldSkipPath verifica si debemos omitir el logging para un path específico
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// shouldSkipSuccessfulResponse verifica si debemos omitir el logging de respuestas exitosas
func shouldSkipSuccessfulResponse(statusCode int, path string, skipSuccessfulPaths []string) bool {
	if statusCode >= 400 {
		return false // Nunca omitir errores
	}

	for _, skipPath := range skipSuccessfulPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// logSlowRequest registra información adicional para requests lentas
func logSlowRequest(ctx context.Context, serverLogger logger.ServerLogger, requestInfo logger.HTTPRequestInfo, responseInfo logger.HTTPResponseInfo, duration time.Duration) {
	// Crear métricas de performance para la request lenta
	metrics := logger.ServerPerformanceMetrics{
		AverageResponseTime: duration,
		TotalRequests:       1,
		TotalErrors:         0,
	}

	if responseInfo.StatusCode >= 400 {
		metrics.TotalErrors = 1
		metrics.ErrorRate = 100.0
	}

	serverLogger.LogPerformanceMetrics(ctx, metrics)
}

// PerformanceLoggingMiddleware middleware para logging de métricas de performance en tiempo real
func PerformanceLoggingMiddleware(serverLogger logger.ServerLogger, interval time.Duration) gin.HandlerFunc {
	var (
		totalRequests uint64
		totalErrors   uint64
		lastLogTime   = time.Now()
	)

	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		totalRequests++
		if statusCode >= 400 {
			totalErrors++
		}

		// Log métricas periódicamente
		if time.Since(lastLogTime) >= interval {
			errorRate := float64(totalErrors) / float64(totalRequests) * 100
			rps := float64(totalRequests) / time.Since(lastLogTime).Seconds()

			metrics := logger.ServerPerformanceMetrics{
				RequestsPerSecond:   rps,
				AverageResponseTime: duration,
				ErrorRate:           errorRate,
				TotalRequests:       totalRequests,
				TotalErrors:         totalErrors,
			}

			serverLogger.LogPerformanceMetrics(c.Request.Context(), metrics)
			lastLogTime = time.Now()
		}
	})
}

// SecurityLoggingMiddleware middleware para logging de eventos de seguridad
func SecurityLoggingMiddleware(serverLogger logger.ServerLogger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Detectar patrones sospechosos
		suspiciousPatterns := []string{
			"<script", "javascript:", "eval(", "union select",
			"../", "..\\", "/etc/passwd", "cmd.exe",
		}

		requestPath := c.Request.URL.Path

		// Verificar patrones sospechosos en path y query
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(strings.ToLower(requestPath), pattern) ||
				strings.Contains(strings.ToLower(c.Request.URL.RawQuery), pattern) {

				details := map[string]interface{}{
					"pattern":    pattern,
					"path":       requestPath,
					"query":      c.Request.URL.RawQuery,
					"user_agent": userAgent,
					"method":     c.Request.Method,
				}

				serverLogger.LogSecurityEvent(
					c.Request.Context(),
					"suspicious_pattern_detected",
					clientIP,
					details,
				)
				break
			}
		}

		c.Next()

		// Log de intentos de acceso no autorizado
		statusCode := c.Writer.Status()
		if statusCode == 401 || statusCode == 403 {
			details := map[string]interface{}{
				"status_code": statusCode,
				"path":        requestPath,
				"method":      c.Request.Method,
				"user_agent":  userAgent,
			}

			serverLogger.LogSecurityEvent(
				c.Request.Context(),
				"unauthorized_access_attempt",
				clientIP,
				details,
			)
		}
	})
}
