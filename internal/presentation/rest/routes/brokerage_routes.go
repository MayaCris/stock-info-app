package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
)

// BrokerageRoutes encapsula la configuración de rutas de brokerages
type BrokerageRoutes struct {
	middlewareManager *MiddlewareManager
}

// NewBrokerageRoutes crea una nueva instancia del configurador de rutas de brokerages
func NewBrokerageRoutes(middlewareManager *MiddlewareManager) *BrokerageRoutes {
	return &BrokerageRoutes{
		middlewareManager: middlewareManager,
	}
}

// SetupBrokerageRoutes configura todas las rutas relacionadas con brokerages
// Esta función configura operaciones CRUD, gestión de estado y búsquedas
func (br *BrokerageRoutes) SetupBrokerageRoutes(routerGroup *gin.RouterGroup, brokerageHandler *handlers.BrokerageHandler) {
	// Verificar que el handler existe
	if brokerageHandler == nil {
		return
	}

	// Configurar el grupo de rutas de brokerages
	brokerages := routerGroup.Group("/brokerages")
	{
		// CRUD operations
		br.setupCRUDRoutes(brokerages, brokerageHandler)

		// State management operations
		br.setupStateRoutes(brokerages, brokerageHandler)

		// Search operations
		br.setupSearchRoutes(brokerages, brokerageHandler)
	}
}

// setupCRUDRoutes configura las operaciones básicas CRUD
func (br *BrokerageRoutes) setupCRUDRoutes(brokerages *gin.RouterGroup, brokerageHandler *handlers.BrokerageHandler) {
	// Create - Crear un nuevo brokerage
	brokerages.POST("/", brokerageHandler.CreateBrokerage)

	// Read - Obtener brokerage por ID
	brokerages.GET("/:id", brokerageHandler.GetBrokerageByID)

	// Update - Actualizar brokerage completo
	brokerages.PUT("/:id", brokerageHandler.UpdateBrokerage)

	// Delete - Eliminar brokerage
	brokerages.DELETE("/:id", brokerageHandler.DeleteBrokerage)

	// List operations
	brokerages.GET("/", brokerageHandler.ListBrokerages)
	brokerages.GET("/active", brokerageHandler.ListActiveBrokerages)
}

// setupStateRoutes configura las rutas de gestión de estado
func (br *BrokerageRoutes) setupStateRoutes(brokerages *gin.RouterGroup, brokerageHandler *handlers.BrokerageHandler) {
	// Activación y desactivación
	brokerages.PATCH("/:id/activate", brokerageHandler.ActivateBrokerage)
	brokerages.PATCH("/:id/deactivate", brokerageHandler.DeactivateBrokerage)

	// Futuras operaciones de estado se pueden agregar aquí
	// brokerages.PATCH("/:id/suspend", brokerageHandler.SuspendBrokerage)
	// brokerages.PATCH("/:id/verify", brokerageHandler.VerifyBrokerage)
}

// setupSearchRoutes configura las rutas de búsqueda
func (br *BrokerageRoutes) setupSearchRoutes(brokerages *gin.RouterGroup, brokerageHandler *handlers.BrokerageHandler) {
	// Búsqueda por nombre
	brokerages.GET("/search", brokerageHandler.SearchBrokeragesByName)

	// Futuras búsquedas se pueden agregar aquí
	// brokerages.GET("/country/:country", brokerageHandler.GetBrokeragesByCountry)
	// brokerages.GET("/type/:type", brokerageHandler.GetBrokeragesByType)
}

// GetBrokerageRoutesInfo retorna información sobre las rutas de brokerages disponibles
func (br *BrokerageRoutes) GetBrokerageRoutesInfo() map[string]interface{} {
	return map[string]interface{}{
		"entity":    "brokerages",
		"base_path": "/brokerages",
		"operations": map[string][]string{
			"crud": {
				"POST /brokerages",
				"GET /brokerages/:id",
				"PUT /brokerages/:id",
				"DELETE /brokerages/:id",
				"GET /brokerages",
				"GET /brokerages/active",
			},
			"state_management": {
				"PATCH /brokerages/:id/activate",
				"PATCH /brokerages/:id/deactivate",
			},
			"search": {
				"GET /brokerages/search",
			},
		},
	}
}
