package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
)

// HealthRoutes encapsula la configuración de rutas de health check
type HealthRoutes struct {
	config *config.Config
}

// NewHealthRoutes crea una nueva instancia del configurador de rutas de health
func NewHealthRoutes(cfg *config.Config) *HealthRoutes {
	return &HealthRoutes{
		config: cfg,
	}
}

// SetupHealthRoutes configura todas las rutas de health check
// Esta función es el punto de entrada principal para configurar las rutas de monitoreo
func (hr *HealthRoutes) SetupHealthRoutes(engine *gin.Engine, healthHandler *handlers.HealthHandler) {
	// Verificar si los health checks están habilitados
	if !hr.config.RESTAPI.EnableHealthChecks {
		return
	}

	// Verificar que el handler existe
	if healthHandler == nil {
		return
	}

	// Configurar el grupo base de health
	hr.setupHealthGroup(engine, healthHandler)
}

// setupHealthGroup configura el grupo principal de rutas de health
func (hr *HealthRoutes) setupHealthGroup(engine *gin.Engine, healthHandler *handlers.HealthHandler) {
	health := engine.Group("/health")
	{
		// Basic health check - overview general del sistema
		health.GET("/", healthHandler.Health)

		// Liveness probe - indica si la aplicación está viva
		// Usado por Kubernetes para reiniciar pods que no responden
		health.GET("/live", healthHandler.Liveness)

		// Readiness probe - indica si la aplicación está lista para recibir tráfico
		// Usado por Kubernetes para agregar/quitar pods del load balancer
		health.GET("/ready", healthHandler.Readiness)
	}

	// Configurar rutas adicionales de health si están habilitadas
	hr.setupExtendedHealthRoutes(health, healthHandler)
}

// setupExtendedHealthRoutes configura rutas adicionales de health check
// Estas rutas proporcionan información más detallada sobre componentes específicos
func (hr *HealthRoutes) setupExtendedHealthRoutes(healthGroup *gin.RouterGroup, healthHandler *handlers.HealthHandler) {
	// Rutas de componentes específicos
	// Nota: Estas rutas requerirían métodos adicionales en el HealthHandler
	// Por ahora están comentadas hasta que se implementen

	// healthGroup.GET("/db", healthHandler.DatabaseHealth)
	// healthGroup.GET("/cache", healthHandler.CacheHealth)
	// healthGroup.GET("/external", healthHandler.ExternalServicesHealth)

	// Rutas de métricas básicas (si se implementan en el futuro)
	// healthGroup.GET("/metrics", healthHandler.Metrics)
	// healthGroup.GET("/info", healthHandler.SystemInfo)
}

// GetHealthEndpoints retorna información sobre los endpoints de health disponibles
func (hr *HealthRoutes) GetHealthEndpoints() map[string]interface{} {
	endpoints := map[string]interface{}{
		"enabled":   hr.config.RESTAPI.EnableHealthChecks,
		"base_path": "/health",
	}

	if hr.config.RESTAPI.EnableHealthChecks {
		endpoints["available_endpoints"] = map[string]string{
			"health":    "/health",
			"liveness":  "/health/live",
			"readiness": "/health/ready",
		}
		endpoints["description"] = map[string]string{
			"/health":       "General health status with all components",
			"/health/live":  "Liveness probe - application is running",
			"/health/ready": "Readiness probe - application is ready to serve traffic",
		}
	}

	return endpoints
}

// IsHealthCheckEnabled indica si los health checks están habilitados
func (hr *HealthRoutes) IsHealthCheckEnabled() bool {
	return hr.config.RESTAPI.EnableHealthChecks
}
