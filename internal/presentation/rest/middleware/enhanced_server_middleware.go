package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// EnhancedRateLimitMiddleware middleware de rate limiting con logging avanzado usando ServerLogger
func EnhancedRateLimitMiddleware(serverLogger logger.ServerLogger, rateLimitConfig config.RateLimitConfig) gin.HandlerFunc {
	// Usar el rate limiter existente
	rateLimiter := NewInMemoryRateLimiter(rateLimitConfig.Limit, rateLimitConfig.RequestsPer)

	return gin.HandlerFunc(func(c *gin.Context) {
		if !rateLimitConfig.Enabled {
			c.Next()
			return
		}

		// Obtener la key para rate limiting
		key := getRateLimitKey(c, rateLimitConfig.KeyFunc)

		// Verificar si la request es permitida
		if !rateLimiter.Allow(key) {
			// Log del rate limit exceeded con ServerLogger
			serverLogger.LogRateLimitExceeded(
				c.Request.Context(),
				c.ClientIP(),
				rateLimitConfig.Limit,
				rateLimitConfig.RequestsPer,
			)

			// Agregar headers de rate limit
			c.Header("X-RateLimit-Limit", string(rune(rateLimitConfig.Limit)))
			c.Header("X-RateLimit-Window", rateLimitConfig.RequestsPer.String())
			c.Header("X-RateLimit-Remaining", "0")
			resetTime := rateLimiter.Reset(key)
			c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

			// Responder con 429
			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests",
				"retry_after": resetTime.Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// Agregar headers informativos de rate limit
		remaining := rateLimiter.Remaining(key)
		c.Header("X-RateLimit-Limit", string(rune(rateLimitConfig.Limit)))
		c.Header("X-RateLimit-Remaining", string(rune(remaining)))
		c.Header("X-RateLimit-Window", rateLimitConfig.RequestsPer.String())

		c.Next()
	})
}

// getRateLimitKey obtiene la key para rate limiting basada en la configuración
func getRateLimitKey(c *gin.Context, keyFunc string) string {
	switch keyFunc {
	case "ip":
		return c.ClientIP()
	case "user_id":
		// Futuro: obtener user ID del contexto de autenticación
		return c.ClientIP() // Fallback a IP
	case "api_key":
		// Futuro: obtener API key del header
		return c.ClientIP() // Fallback a IP
	default:
		return c.ClientIP()
	}
}

// ConnectionMetricsMiddleware middleware para tracking de métricas de conexión
func ConnectionMetricsMiddleware(serverLogger logger.ServerLogger, interval time.Duration) gin.HandlerFunc {
	var (
		activeConnections   int
		totalConnections    uint64
		connectionsAccepted uint64
		lastLogTime         = time.Now()
	)

	return gin.HandlerFunc(func(c *gin.Context) {
		// Incrementar contadores
		activeConnections++
		totalConnections++
		connectionsAccepted++

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Decrementar conexiones activas
		activeConnections--

		// Log métricas periódicamente
		if time.Since(lastLogTime) >= interval {
			metrics := logger.ConnectionMetrics{
				ActiveConnections:   activeConnections,
				TotalConnections:    totalConnections,
				ConnectionsAccepted: connectionsAccepted,
				AverageConnDuration: duration,
				MaxConnections:      100, // Configurable
				KeepAliveEnabled:    true,
			}

			serverLogger.LogConnectionMetrics(c.Request.Context(), metrics)
			lastLogTime = time.Now()
		}
	})
}

// ThroughputMetricsMiddleware middleware para tracking de métricas de throughput
func ThroughputMetricsMiddleware(serverLogger logger.ServerLogger, interval time.Duration) gin.HandlerFunc {
	var (
		bytesReceived      uint64
		bytesSent          uint64
		requestsHandled    uint64
		responsesGenerated uint64
		lastLogTime        = time.Now()
	)

	return gin.HandlerFunc(func(c *gin.Context) {
		// Incrementar requests handled
		requestsHandled++

		// Obtener tamaño de request
		if c.Request.ContentLength > 0 {
			bytesReceived += uint64(c.Request.ContentLength)
		}

		c.Next()

		// Incrementar responses generated y bytes sent
		responsesGenerated++
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			bytesSent += uint64(responseSize)
		}

		// Log métricas periódicamente
		if time.Since(lastLogTime) >= interval {
			// Calcular throughput en MB/s
			elapsed := time.Since(lastLogTime).Seconds()
			throughputMBps := float64(bytesSent+bytesReceived) / (1024 * 1024) / elapsed

			metrics := logger.ThroughputMetrics{
				BytesReceived:      bytesReceived,
				BytesSent:          bytesSent,
				RequestsHandled:    requestsHandled,
				ResponsesGenerated: responsesGenerated,
				ThroughputMBps:     throughputMBps,
			}

			serverLogger.LogThroughputMetrics(c.Request.Context(), metrics)
			lastLogTime = time.Now()

			// Reset counters
			bytesReceived = 0
			bytesSent = 0
		}
	})
}

// ResourceMonitoringMiddleware middleware para monitoreo de recursos del sistema
func ResourceMonitoringMiddleware(serverLogger logger.ServerLogger, interval time.Duration) gin.HandlerFunc {
	var lastLogTime = time.Now()

	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Log recursos del sistema periódicamente
		if time.Since(lastLogTime) >= interval {
			// Obtener métricas del sistema (esto requeriría una librería como gopsutil)
			// Por ahora, usamos valores mock para demostrar la funcionalidad
			cpuUsage := 25.5                         // % CPU usage
			memoryUsage := uint64(512 * 1024 * 1024) // 512 MB en bytes
			goroutineCount := 45                     // Número de goroutines

			serverLogger.LogResourceUsage(
				c.Request.Context(),
				cpuUsage,
				memoryUsage,
				goroutineCount,
			)

			lastLogTime = time.Now()
		}
	})
}

// HealthCheckMiddleware middleware para logging de health checks
func HealthCheckMiddleware(serverLogger logger.ServerLogger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Solo aplicar en rutas de health check
		if c.Request.URL.Path != "/health" && c.Request.URL.Path != "/health/ready" && c.Request.URL.Path != "/health/live" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Determinar el estado basado en el código de respuesta
		status := "healthy"
		if c.Writer.Status() >= 400 {
			status = "unhealthy"
		}

		// Crear detalles del health check
		details := map[string]interface{}{
			"endpoint":         c.Request.URL.Path,
			"status_code":      c.Writer.Status(),
			"response_time_ms": duration.Milliseconds(),
		}

		// Log del health check
		serverLogger.LogHealthCheck(
			c.Request.Context(),
			"http_endpoint",
			status,
			duration,
			details,
		)
	})
}
