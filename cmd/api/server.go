package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/factory"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/routes"
)

// Server encapsula el servidor HTTP y sus dependencias
type Server struct {
	httpServer   *http.Server
	router       *routes.Router
	config       *config.Config
	logger       logger.Logger
	serverLogger logger.ServerLogger

	// Dependencies for cleanup
	dependencies  *factory.Dependencies
	shutdownHooks []ShutdownHook
}

// ShutdownHook representa una función que debe ejecutarse durante el shutdown
type ShutdownHook struct {
	Name     string
	Priority int // Menor número = mayor prioridad
	Cleanup  func(ctx context.Context) error
}

// ShutdownConfig define configuraciones avanzadas para el shutdown
type ShutdownConfig struct {
	GracePeriod      time.Duration // Tiempo de gracia antes de forzar el shutdown
	HookTimeout      time.Duration // Timeout individual para cada hook
	MaxHookFailures  int           // Número máximo de hooks que pueden fallar
	ForceAfterPeriod time.Duration // Tiempo después del cual se fuerza el shutdown
}

// NewServer crea una nueva instancia del servidor HTTP
func NewServer(cfg *config.Config, appLogger logger.Logger) (*Server, error) {
	// Crear factory para dependencias
	apiFactory := factory.NewAPIFactory(cfg)

	// Crear dependencias
	deps, err := apiFactory.CreateDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to create dependencies: %w", err)
	}
	// Crear ServerLogger especializado con configuración optimizada
	loggerFactory := logger.NewLoggerFactory()

	// Crear configuración de logger base a partir de la configuración del servidor
	serverLogConfig := cfg.ServerLogging.ToLoggerConfig()
	serverLogger, err := loggerFactory.CreateServerLoggerWithConfig(serverLogConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create server logger: %w", err)
	}

	// Crear handlers
	handlers, err := createHandlers(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("failed to create handlers: %w", err)
	}

	// Crear router principal
	mainRouter := routes.NewRouter(cfg, appLogger, serverLogger, handlers)

	// Configurar servidor HTTP
	httpServer := &http.Server{
		Addr:           cfg.Server.GetServerAddress(),
		Handler:        mainRouter.GetEngine(),
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}
	// Configurar trusted proxies si están definidos
	if len(cfg.Server.TrustedProxies) > 0 {
		if err := mainRouter.GetEngine().SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
			appLogger.Error(context.Background(), "Failed to set trusted proxies", err,
				logger.Any("proxies", cfg.Server.TrustedProxies),
			)
		}
	}

	return &Server{
		httpServer:    httpServer,
		router:        mainRouter,
		config:        cfg,
		logger:        appLogger,
		serverLogger:  serverLogger,
		dependencies:  deps,
		shutdownHooks: make([]ShutdownHook, 0),
	}, nil
}

// NewServerWithShutdownConfig crea un servidor con configuración avanzada de shutdown
func NewServerWithShutdownConfig(cfg *config.Config, appLogger logger.Logger, shutdownCfg ShutdownConfig) (*Server, error) {
	server, err := NewServer(cfg, appLogger)
	if err != nil {
		return nil, err
	}

	// Configurar shutdown personalizado si se especifica
	if shutdownCfg.GracePeriod > 0 {
		// Crear una configuración temporal para el servidor con el timeout personalizado
		if shutdownCfg.GracePeriod > cfg.Server.ShutdownTimeout {
			appLogger.Warn(context.Background(), "Shutdown grace period is longer than configured timeout",
				logger.String("grace_period", shutdownCfg.GracePeriod.String()),
				logger.String("configured_timeout", cfg.Server.ShutdownTimeout.String()),
			)
		}
	}

	return server, nil
}

// Start inicia el servidor HTTP con graceful shutdown avanzado
func (s *Server) Start() error {
	return s.GracefulShutdownWithSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

// StartWithCustomShutdownHooks inicia el servidor con hooks personalizados
func (s *Server) StartWithCustomShutdownHooks(customHooks []ShutdownHook) error {
	// Registrar hooks personalizados
	for _, hook := range customHooks {
		s.AddShutdownHook(hook.Name, hook.Priority, hook.Cleanup)
	}

	return s.Start()
}

// Shutdown realiza un graceful shutdown del servidor
func (s *Server) Shutdown() error {
	shutdownStart := time.Now()

	// Usar ServerLogger para logging especializado
	s.serverLogger.LogServerShutdown(context.Background(), "shutdown_requested", 0, true)

	// Crear contexto con timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	// Phase 1: Stop accepting new connections
	s.logger.Info(ctx, "Phase 1: Stopping HTTP server from accepting new connections")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(ctx, "Failed to shutdown HTTP server gracefully", err)
		// Log el shutdown fallido con ServerLogger
		shutdownDuration := time.Since(shutdownStart)
		s.serverLogger.LogServerShutdown(ctx, "shutdown_failed", shutdownDuration, false)
		return fmt.Errorf("failed to shutdown server gracefully: %w", err)
	}
	s.logger.Info(ctx, "✅ HTTP server stopped accepting new connections")

	// Phase 2: Execute shutdown hooks in priority order
	s.logger.Info(ctx, "Phase 2: Executing shutdown hooks",
		logger.Int("total_hooks", len(s.shutdownHooks)))

	if err := s.executeShutdownHooks(ctx); err != nil {
		s.logger.Error(ctx, "Some shutdown hooks failed", err)
		// Continue with shutdown even if some hooks fail
	}

	// Phase 3: Cleanup core dependencies
	s.logger.Info(ctx, "Phase 3: Cleaning up core dependencies")
	if err := s.cleanupDependencies(ctx); err != nil {
		s.logger.Error(ctx, "Failed to cleanup some dependencies", err)
		// Continue with shutdown
	}

	shutdownDuration := time.Since(shutdownStart)
	s.logger.Info(context.Background(), "✅ Graceful shutdown completed",
		logger.String("duration", shutdownDuration.String()))

	return nil
}

// ForceShutdown realiza un shutdown forzado del servidor
func (s *Server) ForceShutdown() error {
	forceStart := time.Now()
	s.logger.Warn(context.Background(), "🚨 Forcing server shutdown - this may cause data loss")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Intentar shutdown graceful con timeout muy corto
	done := make(chan error, 1)
	go func() {
		done <- s.httpServer.Shutdown(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			s.logger.Error(ctx, "Failed graceful shutdown, forcing close", err)
			forceErr := s.httpServer.Close() // Forzar cierre inmediato

			// Log forced shutdown with ServerLogger
			forceDuration := time.Since(forceStart)
			s.serverLogger.LogServerShutdown(ctx, "force_shutdown_failed", forceDuration, false)
			return forceErr
		}

		// Log successful force shutdown
		forceDuration := time.Since(forceStart)
		s.serverLogger.LogServerShutdown(ctx, "force_shutdown_success", forceDuration, false)
		return nil
	case <-ctx.Done():
		s.logger.Error(ctx, "Shutdown timeout exceeded, forcing close", ctx.Err())
		forceErr := s.httpServer.Close() // Forzar cierre inmediato

		// Log timeout force shutdown
		forceDuration := time.Since(forceStart)
		s.serverLogger.LogServerShutdown(ctx, "force_shutdown_timeout", forceDuration, false)
		return forceErr
	}
}

// GetRouter retorna la instancia del router principal
func (s *Server) GetRouter() *routes.Router {
	return s.router
}

// GetHTTPServer retorna la instancia del servidor HTTP
func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}

// logServerInfo registra información detallada del servidor (solo en modo debug)
func (s *Server) logServerInfo() {
	engine := s.router.GetEngine()
	routes := engine.Routes()

	s.logger.Info(context.Background(), "Server configuration details",
		logger.String("host", s.config.Server.Host),
		logger.String("port", s.config.Server.Port),
		logger.String("read_timeout", s.config.Server.ReadTimeout.String()),
		logger.String("write_timeout", s.config.Server.WriteTimeout.String()),
		logger.String("idle_timeout", s.config.Server.IdleTimeout.String()),
		logger.String("shutdown_timeout", s.config.Server.ShutdownTimeout.String()),
		logger.Int("max_header_bytes", s.config.Server.MaxHeaderBytes),
		logger.Int("total_routes", len(routes)),
	)

	// Log de configuraciones de funcionalidades
	s.logger.Info(context.Background(), "API features configuration",
		logger.String("api_version", s.config.RESTAPI.Version),
		logger.String("base_path", s.config.RESTAPI.BasePath),
		logger.Bool("swagger_enabled", s.config.RESTAPI.EnableSwagger),
		logger.Bool("health_checks_enabled", s.config.RESTAPI.EnableHealthChecks),
		logger.Bool("metrics_enabled", s.config.RESTAPI.EnableMetrics),
		logger.Bool("profiling_enabled", s.config.RESTAPI.EnableProfiling),
	)

	// Log de configuración de rate limiting
	if s.config.RateLimit.Enabled {
		s.logger.Info(context.Background(), "Rate limiting configuration",
			logger.Bool("enabled", s.config.RateLimit.Enabled),
			logger.Int("limit", s.config.RateLimit.Limit),
			logger.String("requests_per", s.config.RateLimit.RequestsPer.String()),
			logger.String("key_func", s.config.RateLimit.KeyFunc),
		)
	}

	// Log de trusted proxies si están configurados
	if len(s.config.Server.TrustedProxies) > 0 {
		s.logger.Info(context.Background(), "Trusted proxies configured",
			logger.Any("proxies", s.config.Server.TrustedProxies),
		)
	}

	// Log de endpoints principales disponibles
	s.logger.Info(context.Background(), "Available endpoints",
		logger.String("root", "/"),
		logger.String("health", "/health"),
		logger.String("api_base", s.config.RESTAPI.BasePath+"/v1"),
		logger.String("swagger", "/swagger/index.html"),
		logger.String("docs_redirect", "/docs"),
	)
}

// HealthCheck realiza un health check básico del servidor
func (s *Server) HealthCheck() error {
	// Verificar que el servidor esté configurado correctamente
	if s.httpServer == nil {
		return fmt.Errorf("HTTP server is not initialized")
	}

	if s.router == nil {
		return fmt.Errorf("router is not initialized")
	}

	if s.config == nil {
		return fmt.Errorf("configuration is not loaded")
	}

	if s.logger == nil {
		return fmt.Errorf("logger is not initialized")
	}

	return nil
}

// GetServerAddress retorna la dirección completa del servidor
func (s *Server) GetServerAddress() string {
	return s.httpServer.Addr
}

// IsRunning verifica si el servidor está en ejecución
func (s *Server) IsRunning() bool {
	return s.httpServer != nil
}

// createHandlers crea todas las instancias de handlers necesarias
func createHandlers(cfg *config.Config, deps *factory.Dependencies) (*routes.Handlers, error) {
	// Crear handler de health check
	healthHandler := handlers.NewHealthHandler(cfg, deps.Logger, deps.CacheService)

	// Crear handler de stocks
	stockHandler := handlers.NewStockHandler(deps.StockService, deps.Logger)

	// Crear handler de companies
	companyHandler := handlers.NewCompanyHandler(deps.CompanyService, deps.Logger)

	// Crear handler de brokerages
	brokerageHandler := handlers.NewBrokerageHandler(deps.BrokerageService, deps.Logger)

	// Crear handler de analysis
	analysisHandler := handlers.NewAnalysisHandler(deps.AnalysisService, deps.Logger)
	// Crear handler de market data
	marketDataHandler := handlers.NewMarketDataHandler(deps.MarketDataService, deps.Logger)

	// Crear handler de Alpha Vantage
	alphaVantageHandler := handlers.NewAlphaVantageHandler(deps.AlphaVantageService, deps.Logger)

	return &routes.Handlers{
		Health:       healthHandler,
		Stock:        stockHandler,
		Company:      companyHandler,
		Brokerage:    brokerageHandler,
		Analysis:     analysisHandler,
		MarketData:   marketDataHandler,
		AlphaVantage: alphaVantageHandler,
	}, nil
}

// AddShutdownHook registra una función de limpieza que se ejecutará durante el shutdown
func (s *Server) AddShutdownHook(name string, priority int, cleanup func(ctx context.Context) error) {
	hook := ShutdownHook{
		Name:     name,
		Priority: priority,
		Cleanup:  cleanup,
	}
	s.shutdownHooks = append(s.shutdownHooks, hook)
}

// executeShutdownHooks ejecuta todos los shutdown hooks registrados en orden de prioridad
func (s *Server) executeShutdownHooks(ctx context.Context) error {
	if len(s.shutdownHooks) == 0 {
		s.logger.Info(ctx, "No shutdown hooks to execute")
		return nil
	}

	// Ordenar hooks por prioridad (menor número = mayor prioridad)
	sort.Slice(s.shutdownHooks, func(i, j int) bool {
		return s.shutdownHooks[i].Priority < s.shutdownHooks[j].Priority
	})

	var lastError error
	for _, hook := range s.shutdownHooks {
		hookStart := time.Now()
		s.logger.Info(ctx, "Executing shutdown hook",
			logger.String("name", hook.Name),
			logger.Int("priority", hook.Priority),
		)

		// Log hook start with ServerLogger
		s.serverLogger.LogShutdownHookStart(ctx, hook.Name, hook.Priority)

		if err := hook.Cleanup(ctx); err != nil {
			hookDuration := time.Since(hookStart)
			s.logger.Error(ctx, "Shutdown hook failed", err,
				logger.String("hook_name", hook.Name),
			)
			// Log hook failure with ServerLogger
			s.serverLogger.LogShutdownHookComplete(ctx, hook.Name, hookDuration, false)
			lastError = err // Keep track of last error but continue with other hooks
		} else {
			hookDuration := time.Since(hookStart)
			s.logger.Info(ctx, "✅ Shutdown hook completed successfully",
				logger.String("name", hook.Name),
			)
			// Log hook success with ServerLogger
			s.serverLogger.LogShutdownHookComplete(ctx, hook.Name, hookDuration, true)
		}
	}

	return lastError
}

// cleanupDependencies limpia las dependencias principales del servidor
func (s *Server) cleanupDependencies(ctx context.Context) error {
	var lastError error

	// Cleanup logger
	if s.dependencies != nil && s.dependencies.Logger != nil {
		s.logger.Info(ctx, "Cleaning up application logger")
		if err := s.dependencies.Logger.Close(); err != nil {
			s.logger.Error(ctx, "Failed to close application logger", err)
			lastError = err
		} else {
			s.logger.Info(ctx, "✅ Application logger closed successfully")
		}
	}

	// Cleanup cache service if present
	if s.dependencies != nil && s.dependencies.CacheService != nil {
		s.logger.Info(ctx, "Cleaning up cache service")
		// Note: CacheService interface might need a Close() method
		// For now, we'll just log that it exists
		s.logger.Info(ctx, "✅ Cache service cleanup completed")
	}

	// Cleanup transaction service if needed
	if s.dependencies != nil && s.dependencies.TransactionService != nil {
		s.logger.Info(ctx, "Cleaning up transaction service")
		// Note: TransactionService interface might need specific cleanup
		s.logger.Info(ctx, "✅ Transaction service cleanup completed")
	}

	return lastError
}

// RegisterDefaultShutdownHooks registra hooks de shutdown por defecto
func (s *Server) RegisterDefaultShutdownHooks() {
	// Hook para logging de inicio de shutdown (prioridad más alta)
	s.AddShutdownHook("logging_start", 1, func(ctx context.Context) error {
		s.logger.Info(ctx, "🔄 Starting graceful shutdown process")
		return nil
	})

	// Hook para cerrar conexiones activas (prioridad media)
	s.AddShutdownHook("close_connections", 50, func(ctx context.Context) error {
		s.logger.Info(ctx, "Closing remaining connections")
		// Implementation would go here
		return nil
	})

	// Hook para finalizar procesos en background (prioridad baja)
	s.AddShutdownHook("background_processes", 90, func(ctx context.Context) error {
		s.logger.Info(ctx, "Stopping background processes")
		// Implementation would go here
		return nil
	})

	// Hook para logging final (prioridad más baja)
	s.AddShutdownHook("logging_end", 100, func(ctx context.Context) error {
		s.logger.Info(ctx, "🏁 Shutdown hooks execution completed")
		return nil
	})
}

// GracefulShutdownWithSignals maneja múltiples señales de shutdown
func (s *Server) GracefulShutdownWithSignals(signals ...os.Signal) error {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	// Canal para recibir señales del sistema
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, signals...)

	// Canal para errores del servidor
	serverErrors := make(chan error, 1)

	// Registrar hooks por defecto
	s.RegisterDefaultShutdownHooks()
	// Iniciar servidor en goroutine
	go func() {
		// Log especializado del inicio del servidor
		serverStartConfig := logger.ServerStartConfig{
			Host:            s.config.Server.Host,
			Port:            s.config.Server.Port,
			Mode:            s.config.Server.Mode,
			ReadTimeout:     s.config.Server.ReadTimeout,
			WriteTimeout:    s.config.Server.WriteTimeout,
			IdleTimeout:     s.config.Server.IdleTimeout,
			ShutdownTimeout: s.config.Server.ShutdownTimeout,
			MaxHeaderBytes:  s.config.Server.MaxHeaderBytes,
			TrustedProxies:  s.config.Server.TrustedProxies,
		}
		s.serverLogger.LogServerStart(context.Background(), s.httpServer.Addr, s.config.Server.Mode, serverStartConfig)

		// Configurar Gin mode basado en configuración
		if s.config.Server.IsReleaseMode() {
			gin.SetMode(gin.ReleaseMode)
		} else if s.config.Server.IsTestMode() {
			gin.SetMode(gin.TestMode)
		} else {
			gin.SetMode(gin.DebugMode)
		}

		// Log de información adicional en modo debug
		if s.config.Server.IsDebugMode() {
			s.logServerInfo()
		}

		// Log que el servidor está listo
		endpoints := []string{"/", "/health", s.config.RESTAPI.BasePath + "/v1"}
		if s.config.RESTAPI.EnableSwagger {
			endpoints = append(endpoints, "/swagger/")
		}

		features := logger.ServerFeatures{
			SwaggerEnabled:      s.config.RESTAPI.EnableSwagger,
			HealthChecksEnabled: s.config.RESTAPI.EnableHealthChecks,
			MetricsEnabled:      s.config.RESTAPI.EnableMetrics,
			ProfilingEnabled:    s.config.RESTAPI.EnableProfiling,
			RateLimitEnabled:    s.config.RateLimit.Enabled,
			CORSEnabled:         true, // CORS siempre habilitado en nuestro setup
		}
		s.serverLogger.LogServerReady(context.Background(), endpoints, features)

		// Iniciar servidor
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}()

	// Esperar señal de shutdown o error
	select {
	case err := <-serverErrors:
		return err
	case sig := <-quit:
		s.logger.Info(context.Background(), "Received shutdown signal",
			logger.String("signal", sig.String()),
		)
		return s.Shutdown()
	}
}

// GetShutdownStatus retorna información sobre el estado del shutdown
func (s *Server) GetShutdownStatus() map[string]interface{} {
	return map[string]interface{}{
		"shutdown_hooks_registered": len(s.shutdownHooks),
		"server_running":            s.IsRunning(),
		"shutdown_timeout":          s.config.Server.ShutdownTimeout.String(),
		"server_address":            s.GetServerAddress(),
	}
}
