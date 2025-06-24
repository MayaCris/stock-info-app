package logger

import (
	"context"
	"time"
)

// IntegrityLogger define métodos específicos para servicios de integridad
type IntegrityLogger interface {
	Logger

	// Eventos específicos de validación de integridad
	LogValidationStart(ctx context.Context, validationType string, recordCount int)
	LogValidationEnd(ctx context.Context, validationType string, issuesFound int, duration time.Duration)
	LogOrphanDetected(ctx context.Context, entityType string, entityID string, reason string, severity string)
	LogInconsistencyDetected(ctx context.Context, entityType string, field string, expected interface{}, actual interface{})
	LogDuplicateDetected(ctx context.Context, entityType string, criteria string, count int)
	LogBusinessRuleViolation(ctx context.Context, rule string, entityType string, entityID string, details string)
	LogRepairAttempt(ctx context.Context, issueType string, entityID string, action string)
	LogRepairResult(ctx context.Context, issueType string, entityID string, success bool, error error)
}

// CacheLogger define métodos específicos para servicios de cache
type CacheLogger interface {
	Logger

	// Eventos específicos de cache
	LogCacheHit(ctx context.Context, key string, cacheType string)
	LogCacheMiss(ctx context.Context, key string, cacheType string)
	LogCacheSet(ctx context.Context, key string, cacheType string, ttl time.Duration)
	LogCacheDelete(ctx context.Context, key string, cacheType string)
	LogCacheClear(ctx context.Context, cacheType string)
	LogCacheError(ctx context.Context, operation string, key string, err error)
	LogCacheStats(ctx context.Context, cacheType string, hits int, misses int, hitRate float64)
}

// DatabaseLogger define métodos específicos para operaciones de base de datos
type DatabaseLogger interface {
	Logger

	// Eventos específicos de base de datos
	LogConnectionAttempt(ctx context.Context, host string, database string)
	LogConnectionSuccess(ctx context.Context, host string, database string, duration time.Duration)
	LogConnectionFailure(ctx context.Context, host string, database string, err error, duration time.Duration)
	LogQueryStart(ctx context.Context, operation string, table string)
	LogQueryEnd(ctx context.Context, operation string, table string, rowsAffected int64, duration time.Duration)
	LogQueryError(ctx context.Context, operation string, table string, err error, duration time.Duration)
	LogTransactionStart(ctx context.Context, transactionID string)
	LogTransactionCommit(ctx context.Context, transactionID string, duration time.Duration)
	LogTransactionRollback(ctx context.Context, transactionID string, reason string, duration time.Duration)
	LogMigrationStart(ctx context.Context, version string)
	LogMigrationEnd(ctx context.Context, version string, success bool, duration time.Duration)
}

// APILogger define métodos específicos para clientes de API externa
type APILogger interface {
	Logger

	// Eventos específicos de API
	LogAPIRequest(ctx context.Context, method string, url string, headers map[string]string)
	LogAPIResponse(ctx context.Context, method string, url string, statusCode int, duration time.Duration)
	LogAPIError(ctx context.Context, method string, url string, err error, duration time.Duration)
	LogAPIRetry(ctx context.Context, method string, url string, attempt int, delay time.Duration)
	LogAPIRateLimit(ctx context.Context, method string, url string, resetTime time.Time)
	LogAPICircuitBreaker(ctx context.Context, method string, url string, state string)
}

// RepositoryLogger define métodos específicos para repositorios
type RepositoryLogger interface {
	Logger

	// Eventos específicos de repositorio
	LogEntityCreate(ctx context.Context, entityType string, entityID string)
	LogEntityUpdate(ctx context.Context, entityType string, entityID string, changedFields []string)
	LogEntityDelete(ctx context.Context, entityType string, entityID string)
	LogEntityGet(ctx context.Context, entityType string, entityID string, found bool)
	LogEntityList(ctx context.Context, entityType string, filters map[string]interface{}, count int)
	LogRepositoryError(ctx context.Context, operation string, entityType string, err error)
}

// ServiceLogger define métodos específicos para servicios de aplicación
type ServiceLogger interface {
	Logger

	// Eventos específicos de servicios
	LogServiceStart(ctx context.Context, serviceName string, operation string)
	LogServiceEnd(ctx context.Context, serviceName string, operation string, success bool, duration time.Duration)
	LogServiceError(ctx context.Context, serviceName string, operation string, err error)
	LogServiceMetrics(ctx context.Context, serviceName string, metrics map[string]interface{})
	LogBusinessLogicStart(ctx context.Context, operation string, input interface{})
	LogBusinessLogicEnd(ctx context.Context, operation string, output interface{}, duration time.Duration)
	LogValidationError(ctx context.Context, operation string, field string, value interface{}, rule string)
}

// ServerLogger define métodos específicos para logging de operaciones del servidor web
type ServerLogger interface {
	Logger

	// Eventos del ciclo de vida del servidor
	LogServerStart(ctx context.Context, address string, mode string, config ServerStartConfig)
	LogServerShutdown(ctx context.Context, reason string, duration time.Duration, graceful bool)
	LogServerReady(ctx context.Context, endpoints []string, features ServerFeatures)

	// HTTP Request/Response logging
	LogHTTPRequest(ctx context.Context, request HTTPRequestInfo)
	LogHTTPResponse(ctx context.Context, response HTTPResponseInfo)
	LogHTTPError(ctx context.Context, request HTTPRequestInfo, err error, statusCode int)

	// Performance y métricas
	LogPerformanceMetrics(ctx context.Context, metrics ServerPerformanceMetrics)
	LogConnectionMetrics(ctx context.Context, metrics ConnectionMetrics)
	LogThroughputMetrics(ctx context.Context, metrics ThroughputMetrics)

	// Eventos de configuración
	LogConfigurationLoaded(ctx context.Context, configSummary ConfigSummary)
	LogMiddlewareRegistered(ctx context.Context, middlewareName string, priority int)
	LogRouteRegistered(ctx context.Context, method string, path string, handlerName string)

	// Eventos de salud del servidor
	LogHealthCheck(ctx context.Context, component string, status string, duration time.Duration, details map[string]interface{})
	LogResourceUsage(ctx context.Context, cpu float64, memory uint64, goroutines int)

	// Eventos de seguridad
	LogSecurityEvent(ctx context.Context, eventType string, clientIP string, details map[string]interface{})
	LogRateLimitExceeded(ctx context.Context, clientIP string, limit int, window time.Duration)

	// Shutdown hooks específicos
	LogShutdownHookStart(ctx context.Context, hookName string, priority int)
	LogShutdownHookComplete(ctx context.Context, hookName string, duration time.Duration, success bool)
}

// ServerStartConfig contiene información de configuración del servidor al iniciar
type ServerStartConfig struct {
	Host            string        `json:"host"`
	Port            string        `json:"port"`
	Mode            string        `json:"mode"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	MaxHeaderBytes  int           `json:"max_header_bytes"`
	TrustedProxies  []string      `json:"trusted_proxies,omitempty"`
}

// ServerFeatures contiene información sobre las características habilitadas del servidor
type ServerFeatures struct {
	SwaggerEnabled      bool `json:"swagger_enabled"`
	HealthChecksEnabled bool `json:"health_checks_enabled"`
	MetricsEnabled      bool `json:"metrics_enabled"`
	ProfilingEnabled    bool `json:"profiling_enabled"`
	RateLimitEnabled    bool `json:"rate_limit_enabled"`
	CORSEnabled         bool `json:"cors_enabled"`
}

// HTTPRequestInfo contiene información detallada de una request HTTP
type HTTPRequestInfo struct {
	RequestID     string            `json:"request_id"`
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	Query         string            `json:"query,omitempty"`
	ClientIP      string            `json:"client_ip"`
	UserAgent     string            `json:"user_agent"`
	ContentType   string            `json:"content_type,omitempty"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}

// HTTPResponseInfo contiene información detallada de una response HTTP
type HTTPResponseInfo struct {
	RequestID       string            `json:"request_id"`
	StatusCode      int               `json:"status_code"`
	ResponseSize    int               `json:"response_size"`
	Duration        time.Duration     `json:"duration"`
	ContentType     string            `json:"content_type,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	CacheHit        bool              `json:"cache_hit,omitempty"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	HandlerName     string            `json:"handler_name,omitempty"`
	MiddlewareChain []string          `json:"middleware_chain,omitempty"`
}

// ServerPerformanceMetrics contiene métricas de performance del servidor
type ServerPerformanceMetrics struct {
	RequestsPerSecond   float64       `json:"requests_per_second"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	ErrorRate           float64       `json:"error_rate"`
	ActiveConnections   int           `json:"active_connections"`
	TotalRequests       uint64        `json:"total_requests"`
	TotalErrors         uint64        `json:"total_errors"`
	UptimeSeconds       int64         `json:"uptime_seconds"`
}

// ConnectionMetrics contiene métricas de conexiones del servidor
type ConnectionMetrics struct {
	ActiveConnections   int           `json:"active_connections"`
	TotalConnections    uint64        `json:"total_connections"`
	ConnectionsAccepted uint64        `json:"connections_accepted"`
	ConnectionsRejected uint64        `json:"connections_rejected"`
	AverageConnDuration time.Duration `json:"average_conn_duration"`
	MaxConnections      int           `json:"max_connections"`
	KeepAliveEnabled    bool          `json:"keep_alive_enabled"`
}

// ThroughputMetrics contiene métricas de throughput del servidor
type ThroughputMetrics struct {
	BytesReceived      uint64  `json:"bytes_received"`
	BytesSent          uint64  `json:"bytes_sent"`
	RequestsHandled    uint64  `json:"requests_handled"`
	ResponsesGenerated uint64  `json:"responses_generated"`
	ThroughputMBps     float64 `json:"throughput_mbps"`
	PacketsReceived    uint64  `json:"packets_received,omitempty"`
	PacketsSent        uint64  `json:"packets_sent,omitempty"`
}

// ConfigSummary contiene un resumen de la configuración cargada
type ConfigSummary struct {
	Environment     string            `json:"environment"`
	ServerMode      string            `json:"server_mode"`
	LogLevel        string            `json:"log_level"`
	DatabaseEnabled bool              `json:"database_enabled"`
	CacheEnabled    bool              `json:"cache_enabled"`
	ExternalAPIs    int               `json:"external_apis_count"`
	FeatureFlags    map[string]bool   `json:"feature_flags"`
	CustomSettings  map[string]string `json:"custom_settings,omitempty"`
}
