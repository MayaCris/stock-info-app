package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/middleware"
)

// Router encapsula el router principal de Gin y sus dependencias
type Router struct {
	engine       *gin.Engine
	config       *config.Config
	logger       logger.Logger
	serverLogger logger.ServerLogger
}

// Handlers contiene todas las instancias de handlers
type Handlers struct {
	Health       *handlers.HealthHandler
	Stock        *handlers.StockHandler
	Company      *handlers.CompanyHandler
	Brokerage    *handlers.BrokerageHandler
	Analysis     *handlers.AnalysisHandler
	MarketData   *handlers.MarketDataHandler
	AlphaVantage *handlers.AlphaVantageHandler
}

// NewRouter crea una nueva instancia del router principal
func NewRouter(cfg *config.Config, appLogger logger.Logger, serverLogger logger.ServerLogger, handlers *Handlers) *Router {
	// Configurar modo de Gin
	gin.SetMode(cfg.Server.Mode)

	// Crear engine de Gin
	engine := gin.New()

	router := &Router{
		engine:       engine,
		config:       cfg,
		logger:       appLogger,
		serverLogger: serverLogger,
	}

	// Configurar middlewares globales
	router.setupGlobalMiddlewares()

	// Configurar rutas
	router.setupRoutes(handlers)

	return router
}

// GetEngine retorna el engine de Gin configurado
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// setupGlobalMiddlewares configura los middlewares globales
func (r *Router) setupGlobalMiddlewares() {
	// Advanced Recovery middleware usando ServerLogger - debe ir primero
	r.engine.Use(middleware.AdvancedRecoveryMiddleware(r.serverLogger))

	// Request ID middleware - para trazabilidad
	r.engine.Use(middleware.RequestIDMiddleware())

	// Enhanced Server Logging middleware - usar configuración del config
	serverLoggingConfig := r.config.ServerLogging
	middlewareConfig := convertToMiddlewareConfig(serverLoggingConfig.Middleware)
	r.engine.Use(middleware.ServerLoggingMiddleware(r.serverLogger, middlewareConfig))

	// Security Logging middleware - para eventos de seguridad
	if serverLoggingConfig.Handlers.LogErrors {
		r.engine.Use(middleware.SecurityLoggingMiddleware(r.serverLogger))
	}

	// Performance Logging middleware - para métricas en tiempo real (cada 30 segundos)
	if serverLoggingConfig.Handlers.LogMetrics {
		r.engine.Use(middleware.PerformanceLoggingMiddleware(r.serverLogger, 30*time.Second))
	}

	// CORS middleware - para permitir requests cross-origin
	r.engine.Use(middleware.CORSMiddleware(r.config.CORS))

	// Rate limiting middleware - para controlar el tráfico
	r.engine.Use(middleware.RateLimitMiddleware(r.config.RateLimit))

	// Error Response middleware - para estandarizar respuestas de error
	r.engine.Use(middleware.ErrorResponseMiddleware())
}

// setupRoutes configura todas las rutas de la aplicación
func (r *Router) setupRoutes(handlers *Handlers) {
	// Crear el gestor de middlewares
	middlewareManager := NewMiddlewareManager(r.config, r.logger)

	// Ruta raíz
	r.engine.GET("/", r.rootHandler)

	// Health check routes - delegado a HealthRoutes
	healthRoutes := NewHealthRoutes(r.config)
	healthRoutes.SetupHealthRoutes(r.engine, handlers.Health)

	// API routes con versioning - delegado a APIRoutes
	apiRoutes := NewAPIRoutes(r.config, middlewareManager)
	apiRoutes.SetupAPIRoutes(r.engine, handlers)

	// Swagger documentation - solo en modo debug
	if r.config.Server.IsDebugMode() && r.config.RESTAPI.EnableSwagger {
		r.setupSwaggerRoutes()
	}

	// Profiling routes - solo en modo debug
	if r.config.Server.IsDebugMode() && r.config.RESTAPI.EnableProfiling {
		r.setupProfilingRoutes()
	}

	// 404 handler
	r.engine.NoRoute(r.notFoundHandler)
}

// setupSwaggerRoutes configura las rutas de documentación Swagger
func (r *Router) setupSwaggerRoutes() {
	// Swagger documentation endpoint
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Redirect from /docs to /swagger/index.html for convenience
	r.engine.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

// setupProfilingRoutes configura las rutas de profiling (pprof)
func (r *Router) setupProfilingRoutes() {
	// Note: In a real implementation, you would import _ "net/http/pprof"
	// and set up the pprof routes here for performance profiling
	debug := r.engine.Group("/debug")
	{
		debug.GET("/vars", r.debugVarsHandler)
		debug.GET("/routes", r.debugRoutesHandler)
	}
}

// rootHandler maneja la ruta raíz
func (r *Router) rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"service":     "Stock Info API",
		"version":     r.config.RESTAPI.Version,
		"status":      "running",
		"environment": r.config.Server.Mode,
		"endpoints": map[string]string{
			"health":  "/health",
			"api":     r.config.RESTAPI.BasePath + "/v1",
			"swagger": "/swagger/index.html",
		},
	}))
}

// notFoundHandler maneja las rutas no encontradas
func (r *Router) notFoundHandler(c *gin.Context) {
	r.logger.Warn(c.Request.Context(), "Route not found",
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
		logger.String("remote_addr", c.ClientIP()),
	)

	errorResp := response.NotFound("Route")
	apiResponse := errorResp.ToAPIResponse()
	if requestID := c.GetString("request_id"); requestID != "" {
		apiResponse.RequestID = requestID
	}

	c.JSON(errorResp.StatusCode, apiResponse)
}

// debugVarsHandler proporciona información de variables del sistema
func (r *Router) debugVarsHandler(c *gin.Context) {
	vars := map[string]interface{}{
		"config": map[string]interface{}{
			"mode":    r.config.Server.Mode,
			"version": r.config.RESTAPI.Version,
			"host":    r.config.Server.Host,
			"port":    r.config.Server.Port,
		},
		"runtime": map[string]interface{}{
			"go_version": "go1.21+",
			"gin_mode":   gin.Mode(),
		},
	}

	c.JSON(http.StatusOK, response.Success(vars))
}

// debugRoutesHandler lista todas las rutas registradas
func (r *Router) debugRoutesHandler(c *gin.Context) {
	routes := r.engine.Routes()
	routeInfo := make([]map[string]string, len(routes))

	for i, route := range routes {
		routeInfo[i] = map[string]string{
			"method": route.Method,
			"path":   route.Path,
		}
	}
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"total":  len(routes),
		"routes": routeInfo,
	}))
}

// convertToMiddlewareConfig convierte la configuración del servidor a configuración de middleware
func convertToMiddlewareConfig(middlewareLogConfig config.MiddlewareLogConfig) middleware.ServerLoggingConfig {
	return middleware.ServerLoggingConfig{
		LogHeaders:           middlewareLogConfig.LogHeaders,
		LogRequestBody:       false, // Por seguridad, no logear bodies por defecto
		LogResponseBody:      false,
		SkipPaths:            middlewareLogConfig.SkipPaths,
		SkipSuccessfulPaths:  middlewareLogConfig.SkipSuccessfulPaths,
		LogSlowRequests:      middlewareLogConfig.LogSlowRequests,
		SlowRequestThreshold: middlewareLogConfig.SlowThreshold,
		SensitiveHeaders: []string{
			"authorization", "cookie", "set-cookie", "x-api-key",
			"x-auth-token", "proxy-authorization", "www-authenticate",
		},
	}
}
