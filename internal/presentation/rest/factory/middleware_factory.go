package factory

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/middleware"
)

// MiddlewareFactory crea instancias de middlewares para la API REST
type MiddlewareFactory struct {
	config *config.Config
	logger logger.Logger
	// Cached middleware instances for reuse
	corsMiddleware      gin.HandlerFunc
	loggingMiddleware   gin.HandlerFunc
	errorMiddleware     gin.HandlerFunc
	requestIDMiddleware gin.HandlerFunc
	rateLimitMiddleware gin.HandlerFunc
}

// NewMiddlewareFactory crea una nueva factory para middlewares
func NewMiddlewareFactory(cfg *config.Config, appLogger logger.Logger) *MiddlewareFactory {
	return &MiddlewareFactory{
		config: cfg,
		logger: appLogger,
	}
}

// MiddlewareSet representa un conjunto completo de middlewares configurados
type MiddlewareSet struct {
	CORS      gin.HandlerFunc
	Logging   gin.HandlerFunc
	Error     gin.HandlerFunc
	RequestID gin.HandlerFunc
	RateLimit gin.HandlerFunc
}

// CreateAllMiddlewares crea todos los middlewares necesarios para la API
func (f *MiddlewareFactory) CreateAllMiddlewares() *MiddlewareSet {
	return &MiddlewareSet{
		CORS:      f.CreateCORSMiddleware(),
		Logging:   f.CreateLoggingMiddleware(),
		Error:     f.CreateErrorMiddleware(),
		RequestID: f.CreateRequestIDMiddleware(),
		RateLimit: f.CreateRateLimitMiddleware(),
	}
}

// CreateCORSMiddleware crea el middleware de CORS usando la configuración
func (f *MiddlewareFactory) CreateCORSMiddleware() gin.HandlerFunc {
	if f.corsMiddleware != nil {
		return f.corsMiddleware
	}

	f.corsMiddleware = middleware.CORSMiddleware(f.config.CORS)
	return f.corsMiddleware
}

// CreateSimpleCORSMiddleware crea un middleware CORS simple para desarrollo
func (f *MiddlewareFactory) CreateSimpleCORSMiddleware() gin.HandlerFunc {
	return middleware.SimpleCORSMiddleware()
}

// CreateLoggingMiddleware crea el middleware de logging usando el logger existente
func (f *MiddlewareFactory) CreateLoggingMiddleware() gin.HandlerFunc {
	if f.loggingMiddleware != nil {
		return f.loggingMiddleware
	}

	f.loggingMiddleware = middleware.LoggingMiddleware(f.logger)
	return f.loggingMiddleware
}

// CreateErrorMiddleware crea el middleware de manejo de errores
func (f *MiddlewareFactory) CreateErrorMiddleware() gin.HandlerFunc {
	if f.errorMiddleware != nil {
		return f.errorMiddleware
	}

	f.errorMiddleware = middleware.ErrorHandlingMiddleware(f.logger)
	return f.errorMiddleware
}

// CreateRequestIDMiddleware crea el middleware de Request ID
func (f *MiddlewareFactory) CreateRequestIDMiddleware() gin.HandlerFunc {
	if f.requestIDMiddleware != nil {
		return f.requestIDMiddleware
	}

	f.requestIDMiddleware = middleware.RequestIDMiddleware()
	return f.requestIDMiddleware
}

// CreateRateLimitMiddleware crea el middleware de rate limiting usando la configuración
func (f *MiddlewareFactory) CreateRateLimitMiddleware() gin.HandlerFunc {
	if f.rateLimitMiddleware != nil {
		return f.rateLimitMiddleware
	}

	f.rateLimitMiddleware = middleware.RateLimitMiddleware(f.config.RateLimit)
	return f.rateLimitMiddleware
}

// CreateCustomRateLimitMiddleware crea un middleware de rate limiting con parámetros personalizados
func (f *MiddlewareFactory) CreateCustomRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	customConfig := config.RateLimitConfig{
		Enabled:     true,
		RequestsPer: window,
		Limit:       limit,
		KeyFunc:     "ip",
	}
	return middleware.RateLimitMiddleware(customConfig)
}

// CreateDevelopmentMiddlewares crea un conjunto de middlewares optimizado para desarrollo
func (f *MiddlewareFactory) CreateDevelopmentMiddlewares() *MiddlewareSet {
	return &MiddlewareSet{
		CORS:      f.CreateSimpleCORSMiddleware(), // CORS más permisivo para desarrollo
		Logging:   f.CreateLoggingMiddleware(),
		Error:     f.CreateErrorMiddleware(),
		RequestID: f.CreateRequestIDMiddleware(),
		RateLimit: f.CreateDevelopmentRateLimitMiddleware(),
	}
}

// CreateProductionMiddlewares crea un conjunto de middlewares optimizado para producción
func (f *MiddlewareFactory) CreateProductionMiddlewares() *MiddlewareSet {
	return &MiddlewareSet{
		CORS:      f.CreateCORSMiddleware(), // CORS estricto basado en configuración
		Logging:   f.CreateLoggingMiddleware(),
		Error:     f.CreateErrorMiddleware(),
		RequestID: f.CreateRequestIDMiddleware(),
		RateLimit: f.CreateRateLimitMiddleware(),
	}
}

// CreateDevelopmentRateLimitMiddleware crea un rate limiter más permisivo para desarrollo
func (f *MiddlewareFactory) CreateDevelopmentRateLimitMiddleware() gin.HandlerFunc {
	devConfig := config.RateLimitConfig{
		Enabled:     true,
		RequestsPer: time.Minute,
		Limit:       1000, // Más permisivo
		KeyFunc:     "ip",
	}
	return middleware.RateLimitMiddleware(devConfig)
}

// GetMiddlewaresByEnvironment retorna los middlewares apropiados según el entorno
func (f *MiddlewareFactory) GetMiddlewaresByEnvironment() *MiddlewareSet {
	if f.config.App.IsDevelopment() {
		return f.CreateDevelopmentMiddlewares()
	}
	return f.CreateProductionMiddlewares()
}

// CreateSecurityMiddlewares crea middlewares específicos de seguridad
func (f *MiddlewareFactory) CreateSecurityMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.CreateRateLimitMiddleware(),
		f.CreateCORSMiddleware(),
		f.CreateRequestIDMiddleware(),
	}
}

// CreateMonitoringMiddlewares crea middlewares para monitoreo y observabilidad
func (f *MiddlewareFactory) CreateMonitoringMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.CreateRequestIDMiddleware(),
		f.CreateLoggingMiddleware(),
		f.CreateErrorMiddleware(),
	}
}

// CreateBasicMiddlewares crea un conjunto mínimo de middlewares esenciales
func (f *MiddlewareFactory) CreateBasicMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.CreateRequestIDMiddleware(),
		f.CreateLoggingMiddleware(),
		f.CreateErrorMiddleware(),
	}
}

// ResetCache limpia las instancias de middleware en caché (útil para testing)
func (f *MiddlewareFactory) ResetCache() {
	f.corsMiddleware = nil
	f.loggingMiddleware = nil
	f.errorMiddleware = nil
	f.requestIDMiddleware = nil
	f.rateLimitMiddleware = nil
}
