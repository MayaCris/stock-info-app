package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

func main() {
	// Parse command line flags
	var (
		help        = flag.Bool("help", false, "Show help message")
		version     = flag.Bool("version", false, "Show version information")
		configCheck = flag.Bool("config-check", false, "Validate configuration and exit")
		dryRun      = flag.Bool("dry-run", false, "Validate setup without starting server")
	)
	flag.Parse()

	// For help and version, we need to load config first to get app name and version
	if *help || *version {
		// Load configuration early for help/version commands
		cfg, err := config.Load()
		if err != nil {
			// If config fails, use defaults for help/version
			appName := "Stock Info API"
			appVersion := "1.0.0"
			if *help {
				showHelp(appName, appVersion)
			} else {
				showVersion(appName, appVersion)
			}
			fmt.Fprintf(os.Stderr, "⚠️ Warning: Could not load configuration: %v\n", err)
			return
		}

		if *help {
			showHelp(cfg.App.Name, cfg.App.Version)
		} else {
			showVersion(cfg.App.Name, cfg.App.Version)
		}
		return
	}

	// Initialize logger first
	appLogger, err := logger.InitializeGlobalLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := appLogger.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "⚠️ Warning: Failed to close logger: %v\n", closeErr)
		}
	}()

	ctx := context.Background() // Load configuration
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatal(ctx, "Failed to load configuration", err,
			logger.String("component", "main"),
		)
		return
	}

	// Log application startup
	appLogger.Info(ctx, "Starting Stock Info API Server",
		logger.String("component", "main"),
		logger.String("app_name", cfg.App.Name),
		logger.String("version", cfg.App.Version),
		logger.String("environment", cfg.App.Env),
		logger.String("server_mode", cfg.Server.Mode),
	)

	// Validate configuration if requested
	if *configCheck {
		if err := validateConfiguration(cfg, appLogger); err != nil {
			appLogger.Fatal(ctx, "Configuration validation failed", err,
				logger.String("component", "config_validation"),
			)
			return
		}
		appLogger.Info(ctx, "✅ Configuration validation passed")
		return
	}

	// Create and configure server
	server, err := NewServer(cfg, appLogger)
	if err != nil {
		appLogger.Fatal(ctx, "Failed to create server", err,
			logger.String("component", "server_creation"),
		)
		return
	}

	// Perform health check before starting
	if err := server.HealthCheck(); err != nil {
		appLogger.Fatal(ctx, "Server health check failed", err,
			logger.String("component", "health_check"),
		)
		return
	}

	// Dry run - validate setup without starting server
	if *dryRun {
		appLogger.Info(ctx, "✅ Dry run completed successfully - server is ready to start",
			logger.String("address", server.GetServerAddress()),
			logger.String("mode", cfg.Server.Mode),
		)
		return
	}

	// Log startup information
	appLogger.Info(ctx, "Server configuration loaded successfully",
		logger.String("address", server.GetServerAddress()),
		logger.String("mode", cfg.Server.Mode),
		logger.String("api_version", cfg.RESTAPI.Version),
		logger.String("base_path", cfg.RESTAPI.BasePath),
		logger.Bool("swagger_enabled", cfg.RESTAPI.EnableSwagger),
		logger.Bool("health_checks_enabled", cfg.RESTAPI.EnableHealthChecks),
	)
	// Start server (blocking call with graceful shutdown)
	appLogger.Info(ctx, "🚀 Starting HTTP server...",
		logger.String("address", server.GetServerAddress()),
	)

	// Configurar shutdown hooks personalizados
	customHooks := setupCustomShutdownHooks(cfg, appLogger)
	appLogger.Info(ctx, "Configured custom shutdown hooks",
		logger.Int("custom_hooks", len(customHooks)),
	)

	if err := server.StartWithCustomShutdownHooks(customHooks); err != nil {
		appLogger.Fatal(ctx, "Server failed to start or encountered an error", err,
			logger.String("component", "server_start"),
		)
		return
	}

	// This line will only be reached after graceful shutdown
	appLogger.Info(ctx, "✅ Server shutdown completed successfully",
		logger.String("component", "main"),
	)
}

// showHelp displays help information
func showHelp(appName, appVersion string) {
	fmt.Printf("%s - %s\n\n", appName, appVersion)
	fmt.Println("USAGE:")
	fmt.Printf("  %s [options]\n\n", os.Args[0])
	fmt.Println("OPTIONS:")
	fmt.Println("  -help          Show this help message")
	fmt.Println("  -version       Show version information")
	fmt.Println("  -config-check  Validate configuration and exit")
	fmt.Println("  -dry-run       Validate setup without starting server")
	fmt.Println("")
	fmt.Println("ENVIRONMENT:")
	fmt.Println("  Configuration is loaded from environment variables and .env file")
	fmt.Println("  See docs/api/ for detailed configuration options")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Printf("  %s                    # Start the server\n", os.Args[0])
	fmt.Printf("  %s -config-check      # Validate configuration\n", os.Args[0])
	fmt.Printf("  %s -dry-run           # Test setup without starting\n", os.Args[0])
	fmt.Printf("  %s -version           # Show version\n", os.Args[0])
	fmt.Println("")
	fmt.Println("API ENDPOINTS:")
	fmt.Println("  GET  /                Health check and API info")
	fmt.Println("  GET  /health          Detailed health status")
	fmt.Println("  GET  /api/v1/*        REST API endpoints")
	fmt.Println("  GET  /swagger/*       API documentation (debug mode)")
	fmt.Println("")
}

// showVersion displays version information
func showVersion(appName, appVersion string) {
	fmt.Printf("%s\n", appName)
	fmt.Printf("Version: %s\n", appVersion)
	fmt.Printf("Built with: Go\n")
	fmt.Printf("Framework: Gin Web Framework\n")
	fmt.Printf("Architecture: Clean Architecture\n")
}

// validateConfiguration performs comprehensive configuration validation
func validateConfiguration(cfg *config.Config, appLogger logger.Logger) error {
	ctx := context.Background()

	// Validate app configuration
	if cfg.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if cfg.App.Env == "" {
		return fmt.Errorf("app environment is required")
	}
	if cfg.App.Port == "" {
		return fmt.Errorf("app port is required")
	}

	appLogger.Info(ctx, "✅ App configuration valid",
		logger.String("name", cfg.App.Name),
		logger.String("env", cfg.App.Env),
		logger.String("port", cfg.App.Port),
	)

	// Validate server configuration
	if cfg.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if cfg.Server.Mode == "" {
		return fmt.Errorf("server mode is required")
	}

	appLogger.Info(ctx, "✅ Server configuration valid",
		logger.String("host", cfg.Server.Host),
		logger.String("port", cfg.Server.Port),
		logger.String("mode", cfg.Server.Mode),
	)

	// Validate database configuration
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Port == "" {
		return fmt.Errorf("database port is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	appLogger.Info(ctx, "✅ Database configuration valid",
		logger.String("host", cfg.Database.Host),
		logger.String("port", cfg.Database.Port),
		logger.String("name", cfg.Database.Name),
		logger.String("user", cfg.Database.User),
	)

	// Validate API configuration
	if cfg.RESTAPI.Version == "" {
		return fmt.Errorf("API version is required")
	}
	if cfg.RESTAPI.BasePath == "" {
		return fmt.Errorf("API base path is required")
	}

	appLogger.Info(ctx, "✅ API configuration valid",
		logger.String("version", cfg.RESTAPI.Version),
		logger.String("base_path", cfg.RESTAPI.BasePath),
		logger.Bool("swagger_enabled", cfg.RESTAPI.EnableSwagger),
	)

	// Validate external APIs (if configured)
	if cfg.External.Primary.BaseURL != "" {
		appLogger.Info(ctx, "✅ Primary external API configured",
			logger.String("name", cfg.External.Primary.Name),
			logger.String("base_url", cfg.External.Primary.BaseURL),
		)
	}

	if cfg.External.Secondary.BaseURL != "" {
		appLogger.Info(ctx, "✅ Secondary external API configured",
			logger.String("name", cfg.External.Secondary.Name),
			logger.String("base_url", cfg.External.Secondary.BaseURL),
		)
	}

	// Validate cache configuration (optional)
	if cfg.Cache.Host != "" {
		appLogger.Info(ctx, "✅ Cache configuration valid",
			logger.String("host", cfg.Cache.Host),
			logger.String("port", cfg.Cache.Port),
			logger.Int("db", cfg.Cache.DB),
		)
	} else {
		appLogger.Warn(ctx, "⚠️ Cache not configured - running without cache")
	}

	appLogger.Info(ctx, "🎉 All configuration validation checks passed")
	return nil
}

// setupCustomShutdownHooks configura hooks de shutdown específicos para la aplicación
func setupCustomShutdownHooks(cfg *config.Config, appLogger logger.Logger) []ShutdownHook {
	var hooks []ShutdownHook

	// Hook para guardar métricas finales
	hooks = append(hooks, ShutdownHook{
		Name:     "save_metrics",
		Priority: 10,
		Cleanup: func(ctx context.Context) error {
			appLogger.Info(ctx, "💾 Saving final application metrics")
			// Aquí iría la lógica para guardar métricas
			return nil
		},
	})

	// Hook para notificar sistemas externos
	hooks = append(hooks, ShutdownHook{
		Name:     "notify_external_systems",
		Priority: 20,
		Cleanup: func(ctx context.Context) error {
			appLogger.Info(ctx, "📢 Notifying external systems of shutdown")
			// Aquí iría la lógica para notificar a sistemas externos
			return nil
		},
	})

	// Hook para limpiar archivos temporales
	hooks = append(hooks, ShutdownHook{
		Name:     "cleanup_temp_files",
		Priority: 80,
		Cleanup: func(ctx context.Context) error {
			appLogger.Info(ctx, "🧹 Cleaning up temporary files")
			// Aquí iría la lógica para limpiar archivos temporales
			return nil
		},
	})

	// Hook condicional para entorno de desarrollo
	if cfg.App.IsDevelopment() {
		hooks = append(hooks, ShutdownHook{
			Name:     "dev_cleanup",
			Priority: 85,
			Cleanup: func(ctx context.Context) error {
				appLogger.Info(ctx, "🔧 Performing development environment cleanup")
				// Lógica específica para desarrollo
				return nil
			},
		})
	}

	return hooks
}
