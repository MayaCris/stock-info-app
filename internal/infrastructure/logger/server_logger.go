package logger

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ServerFileLogger implementa ServerLogger usando el sistema de logging existente
type ServerFileLogger struct {
	Logger
	config    *LogConfig
	startTime time.Time
}

// NewServerLogger crea un nuevo logger específico para operaciones del servidor web
func NewServerLogger(baseLogger Logger, config *LogConfig) ServerLogger {
	return &ServerFileLogger{
		Logger:    baseLogger.WithContext("SERVER"),
		config:    config,
		startTime: time.Now(),
	}
}

// LogServerStart registra el inicio del servidor
func (l *ServerFileLogger) LogServerStart(ctx context.Context, address string, mode string, config ServerStartConfig) {
	l.Info(ctx, "🚀 HTTP Server Starting",
		String("operation", "server_start"),
		String("address", address),
		String("mode", mode),
		String("host", config.Host),
		String("port", config.Port),
		Duration("read_timeout", config.ReadTimeout),
		Duration("write_timeout", config.WriteTimeout),
		Duration("idle_timeout", config.IdleTimeout),
		Duration("shutdown_timeout", config.ShutdownTimeout),
		Int("max_header_bytes", config.MaxHeaderBytes),
		Any("trusted_proxies", config.TrustedProxies),
	)
}

// LogServerShutdown registra el shutdown del servidor
func (l *ServerFileLogger) LogServerShutdown(ctx context.Context, reason string, duration time.Duration, graceful bool) {
	uptime := time.Since(l.startTime)

	status := "🔴"
	message := "HTTP Server Shutdown"
	if graceful {
		status = "✅"
		message = "HTTP Server Graceful Shutdown"
	}

	l.Info(ctx, fmt.Sprintf("%s %s", status, message),
		String("operation", "server_shutdown"),
		String("reason", reason),
		Duration("shutdown_duration", duration),
		Duration("uptime", uptime),
		Bool("graceful", graceful),
	)
}

// LogServerReady registra cuando el servidor está listo para recibir requests
func (l *ServerFileLogger) LogServerReady(ctx context.Context, endpoints []string, features ServerFeatures) {
	l.Info(ctx, "✅ HTTP Server Ready",
		String("operation", "server_ready"),
		Any("endpoints", endpoints),
		Bool("swagger_enabled", features.SwaggerEnabled),
		Bool("health_checks_enabled", features.HealthChecksEnabled),
		Bool("metrics_enabled", features.MetricsEnabled),
		Bool("profiling_enabled", features.ProfilingEnabled),
		Bool("rate_limit_enabled", features.RateLimitEnabled),
		Bool("cors_enabled", features.CORSEnabled),
	)
}

// LogHTTPRequest registra información detallada de una request HTTP
func (l *ServerFileLogger) LogHTTPRequest(ctx context.Context, request HTTPRequestInfo) {
	fields := []Field{
		String("operation", "http_request"),
		String("request_id", request.RequestID),
		String("method", request.Method),
		String("path", request.Path),
		String("client_ip", request.ClientIP),
		String("user_agent", request.UserAgent),
		Int64("content_length", request.ContentLength),
		Time("timestamp", request.Timestamp),
	}

	if request.Query != "" {
		fields = append(fields, String("query", request.Query))
	}
	if request.ContentType != "" {
		fields = append(fields, String("content_type", request.ContentType))
	}

	// Log headers solo en modo debug
	if l.config.Level == DebugLevel && len(request.Headers) > 0 {
		fields = append(fields, Any("headers", request.Headers))
	}

	l.Info(ctx, "📨 HTTP Request",
		fields...,
	)
}

// LogHTTPResponse registra información detallada de una response HTTP
func (l *ServerFileLogger) LogHTTPResponse(ctx context.Context, response HTTPResponseInfo) {
	// Determinar nivel de log basado en status code
	var message string

	switch {
	case response.StatusCode >= 500:
		message = "❌ HTTP Response Error"
	case response.StatusCode >= 400:
		message = "⚠️ HTTP Response Warning"
	case response.StatusCode >= 300:
		message = "🔄 HTTP Response Redirect"
	default:
		message = "✅ HTTP Response Success"
	}

	fields := []Field{
		String("operation", "http_response"),
		String("request_id", response.RequestID),
		Int("status_code", response.StatusCode),
		Int("response_size", response.ResponseSize),
		Duration("duration", response.Duration),
	}

	if response.ContentType != "" {
		fields = append(fields, String("content_type", response.ContentType))
	}
	if response.CacheHit {
		fields = append(fields, Bool("cache_hit", response.CacheHit))
	}
	if response.ErrorMessage != "" {
		fields = append(fields, String("error_message", response.ErrorMessage))
	}
	if response.HandlerName != "" {
		fields = append(fields, String("handler_name", response.HandlerName))
	}

	// Log headers y middleware chain solo en modo debug
	if l.config.Level == DebugLevel {
		if len(response.Headers) > 0 {
			fields = append(fields, Any("headers", response.Headers))
		}
		if len(response.MiddlewareChain) > 0 {
			fields = append(fields, Any("middleware_chain", response.MiddlewareChain))
		}
	}

	// Usar el método apropiado según el status code
	switch {
	case response.StatusCode >= 500:
		l.Error(ctx, message, fmt.Errorf("server error: status %d", response.StatusCode), fields...)
	case response.StatusCode >= 400:
		l.Warn(ctx, message, fields...)
	default:
		l.Info(ctx, message, fields...)
	}
}

// LogHTTPError registra errores específicos de HTTP
func (l *ServerFileLogger) LogHTTPError(ctx context.Context, request HTTPRequestInfo, err error, statusCode int) {
	l.Error(ctx, "💥 HTTP Request Error", err,
		String("operation", "http_error"),
		String("request_id", request.RequestID),
		String("method", request.Method),
		String("path", request.Path),
		String("client_ip", request.ClientIP),
		Int("status_code", statusCode),
		String("error_type", "request_processing"),
	)
}

// LogPerformanceMetrics registra métricas de performance del servidor
func (l *ServerFileLogger) LogPerformanceMetrics(ctx context.Context, metrics ServerPerformanceMetrics) {
	l.Info(ctx, "📊 Server Performance Metrics",
		String("operation", "performance_metrics"),
		Float64("requests_per_second", metrics.RequestsPerSecond),
		Duration("avg_response_time", metrics.AverageResponseTime),
		Duration("p95_response_time", metrics.P95ResponseTime),
		Duration("p99_response_time", metrics.P99ResponseTime),
		Float64("error_rate", metrics.ErrorRate),
		Int("active_connections", metrics.ActiveConnections),
		Uint64("total_requests", metrics.TotalRequests),
		Uint64("total_errors", metrics.TotalErrors),
		Int64("uptime_seconds", metrics.UptimeSeconds),
	)
}

// LogConnectionMetrics registra métricas de conexiones
func (l *ServerFileLogger) LogConnectionMetrics(ctx context.Context, metrics ConnectionMetrics) {
	l.Info(ctx, "🔌 Connection Metrics",
		String("operation", "connection_metrics"),
		Int("active_connections", metrics.ActiveConnections),
		Uint64("total_connections", metrics.TotalConnections),
		Uint64("connections_accepted", metrics.ConnectionsAccepted),
		Uint64("connections_rejected", metrics.ConnectionsRejected),
		Duration("avg_conn_duration", metrics.AverageConnDuration),
		Int("max_connections", metrics.MaxConnections),
		Bool("keep_alive_enabled", metrics.KeepAliveEnabled),
	)
}

// LogThroughputMetrics registra métricas de throughput
func (l *ServerFileLogger) LogThroughputMetrics(ctx context.Context, metrics ThroughputMetrics) {
	l.Info(ctx, "🚄 Throughput Metrics",
		String("operation", "throughput_metrics"),
		Uint64("bytes_received", metrics.BytesReceived),
		Uint64("bytes_sent", metrics.BytesSent),
		Uint64("requests_handled", metrics.RequestsHandled),
		Uint64("responses_generated", metrics.ResponsesGenerated),
		Float64("throughput_mbps", metrics.ThroughputMBps),
		Uint64("packets_received", metrics.PacketsReceived),
		Uint64("packets_sent", metrics.PacketsSent),
	)
}

// LogConfigurationLoaded registra cuando se carga la configuración
func (l *ServerFileLogger) LogConfigurationLoaded(ctx context.Context, configSummary ConfigSummary) {
	l.Info(ctx, "⚙️ Configuration Loaded",
		String("operation", "config_loaded"),
		String("environment", configSummary.Environment),
		String("server_mode", configSummary.ServerMode),
		String("log_level", configSummary.LogLevel),
		Bool("database_enabled", configSummary.DatabaseEnabled),
		Bool("cache_enabled", configSummary.CacheEnabled),
		Int("external_apis_count", configSummary.ExternalAPIs),
		Any("feature_flags", configSummary.FeatureFlags),
		Any("custom_settings", configSummary.CustomSettings),
	)
}

// LogMiddlewareRegistered registra cuando se registra un middleware
func (l *ServerFileLogger) LogMiddlewareRegistered(ctx context.Context, middlewareName string, priority int) {
	l.Debug(ctx, "🔧 Middleware Registered",
		String("operation", "middleware_registered"),
		String("middleware_name", middlewareName),
		Int("priority", priority),
	)
}

// LogRouteRegistered registra cuando se registra una ruta
func (l *ServerFileLogger) LogRouteRegistered(ctx context.Context, method string, path string, handlerName string) {
	l.Debug(ctx, "🛣️ Route Registered",
		String("operation", "route_registered"),
		String("method", method),
		String("path", path),
		String("handler_name", handlerName),
	)
}

// LogHealthCheck registra resultados de health checks
func (l *ServerFileLogger) LogHealthCheck(ctx context.Context, component string, status string, duration time.Duration, details map[string]interface{}) {
	var message string

	switch status {
	case "healthy":
		message = "✅ Health Check Passed"
	case "degraded":
		message = "⚠️ Health Check Degraded"
	default:
		message = "❌ Health Check Failed"
	}

	fields := []Field{
		String("operation", "health_check"),
		String("component", component),
		String("status", status),
		Duration("duration", duration),
	}

	if details != nil {
		fields = append(fields, Any("details", details))
	}

	// Usar el método apropiado según el status
	switch status {
	case "healthy":
		l.Info(ctx, message, fields...)
	case "degraded":
		l.Warn(ctx, message, fields...)
	default:
		l.Error(ctx, message, fmt.Errorf("health check failed for component: %s", component), fields...)
	}
}

// LogResourceUsage registra el uso de recursos del sistema
func (l *ServerFileLogger) LogResourceUsage(ctx context.Context, cpu float64, memory uint64, goroutines int) {
	l.Debug(ctx, "💻 Resource Usage",
		String("operation", "resource_usage"),
		Float64("cpu_percent", cpu),
		Uint64("memory_bytes", memory),
		Int("goroutines", goroutines),
		Int("runtime_goroutines", runtime.NumGoroutine()), // Agregar métricas del runtime
	)
}

// LogSecurityEvent registra eventos de seguridad
func (l *ServerFileLogger) LogSecurityEvent(ctx context.Context, eventType string, clientIP string, details map[string]interface{}) {
	l.Warn(ctx, "🛡️ Security Event",
		String("operation", "security_event"),
		String("event_type", eventType),
		String("client_ip", clientIP),
		Any("details", details),
	)
}

// LogRateLimitExceeded registra cuando se excede el rate limit
func (l *ServerFileLogger) LogRateLimitExceeded(ctx context.Context, clientIP string, limit int, window time.Duration) {
	l.Warn(ctx, "🚦 Rate Limit Exceeded",
		String("operation", "rate_limit_exceeded"),
		String("client_ip", clientIP),
		Int("limit", limit),
		Duration("window", window),
	)
}

// LogShutdownHookStart registra el inicio de un shutdown hook
func (l *ServerFileLogger) LogShutdownHookStart(ctx context.Context, hookName string, priority int) {
	l.Info(ctx, "🪝 Shutdown Hook Starting",
		String("operation", "shutdown_hook_start"),
		String("hook_name", hookName),
		Int("priority", priority),
	)
}

// LogShutdownHookComplete registra la finalización de un shutdown hook
func (l *ServerFileLogger) LogShutdownHookComplete(ctx context.Context, hookName string, duration time.Duration, success bool) {
	var message string

	if success {
		message = "✅ Shutdown Hook Completed"
	} else {
		message = "❌ Shutdown Hook Failed"
	}

	fields := []Field{
		String("operation", "shutdown_hook_complete"),
		String("hook_name", hookName),
		Duration("duration", duration),
		Bool("success", success),
	}

	// Usar el método apropiado según el resultado
	if success {
		l.Info(ctx, message, fields...)
	} else {
		l.Error(ctx, message, fmt.Errorf("shutdown hook failed: %s", hookName), fields...)
	}
}
