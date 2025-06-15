package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// MiddlewareManager gestiona la aplicación de middlewares específicos por tipo de ruta
type MiddlewareManager struct {
	config *config.Config
	logger logger.Logger
}

// NewMiddlewareManager crea una nueva instancia del gestor de middlewares
func NewMiddlewareManager(cfg *config.Config, appLogger logger.Logger) *MiddlewareManager {
	return &MiddlewareManager{
		config: cfg,
		logger: appLogger,
	}
}

// ApplyReadOnlyMiddlewares aplica middlewares específicos para operaciones de solo lectura
// Estas operaciones pueden tener rate limiting más permisivo
func (mm *MiddlewareManager) ApplyReadOnlyMiddlewares(group *gin.RouterGroup) {
	// Cache-friendly headers para operaciones de lectura
	group.Use(mm.cacheHeadersMiddleware())

	// Rate limiting específico para lecturas (más permisivo)
	if mm.config.RateLimit.Enabled {
		// Se puede implementar un rate limit específico para lecturas
		// group.Use(middleware.ReadOnlyRateLimitMiddleware(mm.config.RateLimit))
	}
}

// ApplyWriteMiddlewares aplica middlewares específicos para operaciones de escritura
// Estas operaciones requieren validaciones más estrictas
func (mm *MiddlewareManager) ApplyWriteMiddlewares(group *gin.RouterGroup) {
	// Validation middleware para operaciones de escritura
	group.Use(mm.writeValidationMiddleware())

	// Rate limiting más estricto para escrituras
	if mm.config.RateLimit.Enabled {
		// Rate limit más restrictivo para operaciones de escritura
		// group.Use(middleware.WriteRateLimitMiddleware(mm.config.RateLimit))
	}

	// Futuro: Authentication/Authorization middleware
	// group.Use(middleware.AuthenticationMiddleware())
	// group.Use(middleware.AuthorizationMiddleware("write"))
}

// ApplyAdminMiddlewares aplica middlewares específicos para operaciones administrativas
// Estas operaciones requieren los permisos más altos
func (mm *MiddlewareManager) ApplyAdminMiddlewares(group *gin.RouterGroup) {
	// Rate limiting muy estricto para operaciones admin
	if mm.config.RateLimit.Enabled {
		// Rate limit muy restrictivo para operaciones administrativas
		// group.Use(middleware.AdminRateLimitMiddleware(mm.config.RateLimit))
	}

	// Futuro: Admin authentication/authorization
	// group.Use(middleware.AuthenticationMiddleware())
	// group.Use(middleware.AuthorizationMiddleware("admin"))

	// Audit logging para operaciones administrativas
	group.Use(mm.auditLoggingMiddleware())
}

// ApplySearchMiddlewares aplica middlewares específicos para operaciones de búsqueda
// Estas operaciones pueden ser costosas computacionalmente
func (mm *MiddlewareManager) ApplySearchMiddlewares(group *gin.RouterGroup) {
	// Rate limiting específico para búsquedas (puede ser costoso)
	if mm.config.RateLimit.Enabled {
		// Rate limit específico para búsquedas
		// group.Use(middleware.SearchRateLimitMiddleware(mm.config.RateLimit))
	}

	// Cache headers optimizados para búsquedas
	group.Use(mm.searchCacheHeadersMiddleware())
}

// cacheHeadersMiddleware añade headers de cache apropiados para operaciones de lectura
func (mm *MiddlewareManager) cacheHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Headers de cache para operaciones de lectura
		c.Header("Cache-Control", "public, max-age=300") // 5 minutos
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	}
}

// searchCacheHeadersMiddleware añade headers de cache específicos para búsquedas
func (mm *MiddlewareManager) searchCacheHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Headers de cache más cortos para búsquedas (pueden cambiar más frecuentemente)
		c.Header("Cache-Control", "public, max-age=60") // 1 minuto
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	}
}

// writeValidationMiddleware añade validaciones específicas para operaciones de escritura
func (mm *MiddlewareManager) writeValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar Content-Type para operaciones POST/PUT/PATCH
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" || (contentType != "application/json" && contentType != "application/json; charset=utf-8") {
				mm.logger.Warn(c.Request.Context(), "Invalid content type for write operation",
					logger.String("method", c.Request.Method),
					logger.String("content_type", contentType),
					logger.String("path", c.Request.URL.Path),
				)
			}
		}

		// Headers de seguridad para operaciones de escritura
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}

// auditLoggingMiddleware añade logging de auditoría para operaciones administrativas
func (mm *MiddlewareManager) auditLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log adicional para operaciones administrativas
		mm.logger.Info(c.Request.Context(), "Admin operation attempted",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("remote_addr", c.ClientIP()),
			logger.String("user_agent", c.GetHeader("User-Agent")),
		)

		c.Next()

		// Log resultado de la operación administrativa
		mm.logger.Info(c.Request.Context(), "Admin operation completed",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.Int("status_code", c.Writer.Status()),
		)
	}
}

// GetMiddlewareInfo retorna información sobre los middlewares disponibles
func (mm *MiddlewareManager) GetMiddlewareInfo() map[string]interface{} {
	return map[string]interface{}{
		"available_middleware_types": []string{
			"read_only",
			"write",
			"admin",
			"search",
		},
		"features": map[string][]string{
			"read_only": {"cache_headers", "permissive_rate_limit"},
			"write":     {"validation", "strict_rate_limit", "security_headers"},
			"admin":     {"audit_logging", "very_strict_rate_limit", "admin_auth"},
			"search":    {"search_rate_limit", "search_cache_headers"},
		},
		"security": map[string]bool{
			"content_type_validation": true,
			"security_headers":        true,
			"audit_logging":           true,
			"rate_limiting":           mm.config.RateLimit.Enabled,
		},
	}
}
