package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

// APIRoutes encapsula la configuración de rutas de la API con versioning
type APIRoutes struct {
	config            *config.Config
	middlewareManager *MiddlewareManager
}

// NewAPIRoutes crea una nueva instancia del configurador de rutas de API
func NewAPIRoutes(cfg *config.Config, middlewareManager *MiddlewareManager) *APIRoutes {
	return &APIRoutes{
		config:            cfg,
		middlewareManager: middlewareManager,
	}
}

// SetupAPIRoutes configura todas las rutas de la API con versioning
// Esta función es el punto de entrada principal para configurar todas las rutas de la API
func (ar *APIRoutes) SetupAPIRoutes(engine *gin.Engine, handlers *Handlers) {
	// API v1 group - configuración del versionado principal
	v1 := ar.setupAPIv1Group(engine)

	// Configurar rutas por entidades en el grupo v1
	ar.setupEntityRoutes(v1, handlers)

	// Futuras versiones se pueden agregar aquí
	// v2 := ar.setupAPIv2Group(engine)
}

// setupAPIv1Group configura el grupo base para la API v1
func (ar *APIRoutes) setupAPIv1Group(engine *gin.Engine) *gin.RouterGroup {
	// Crear grupo con base path y versión
	basePath := ar.config.RESTAPI.BasePath
	v1GroupPath := fmt.Sprintf("%s/v1", basePath)

	v1 := engine.Group(v1GroupPath)

	// Middleware específicos para API v1 se pueden agregar aquí
	// v1.Use(middleware.APIVersionMiddleware("v1"))

	return v1
}

// setupEntityRoutes configura las rutas específicas de cada entidad en el grupo v1
func (ar *APIRoutes) setupEntityRoutes(v1 *gin.RouterGroup, handlers *Handlers) {
	// Configurar rutas de stocks usando StockRoutes
	if handlers.Stock != nil {
		stockRoutes := NewStockRoutes(ar.middlewareManager)
		stockRoutes.SetupStockRoutes(v1, handlers.Stock)
	}

	// Configurar rutas de companies usando CompanyRoutes
	if handlers.Company != nil {
		companyRoutes := NewCompanyRoutes(ar.middlewareManager)
		companyRoutes.SetupCompanyRoutes(v1, handlers.Company)
	}

	// Configurar rutas de brokerages usando BrokerageRoutes
	if handlers.Brokerage != nil {
		brokerageRoutes := NewBrokerageRoutes(ar.middlewareManager)
		brokerageRoutes.SetupBrokerageRoutes(v1, handlers.Brokerage)
	}

	// Configurar rutas de analysis usando AnalysisRoutes
	if handlers.Analysis != nil {
		analysisRoutes := NewAnalysisRoutes(ar.middlewareManager)
		analysisRoutes.SetupAnalysisRoutes(v1, handlers.Analysis)
	}

	// Configurar rutas de market data usando MarketDataRoutes
	if handlers.MarketData != nil {
		marketDataRoutes := NewMarketDataRoutes(ar.middlewareManager)
		marketDataRoutes.SetupMarketDataRoutes(v1, handlers.MarketData)
	}
}

// GetAPIInfo retorna información sobre las versiones de API disponibles
func (ar *APIRoutes) GetAPIInfo() map[string]interface{} {
	return map[string]interface{}{
		"current_version":    "v1",
		"supported_versions": []string{"v1"},
		"base_path":          ar.config.RESTAPI.BasePath,
		"endpoints": map[string]string{
			"v1": fmt.Sprintf("%s/v1", ar.config.RESTAPI.BasePath),
		},
	}
}
