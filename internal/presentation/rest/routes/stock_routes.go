package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
)

// StockRoutes encapsula la configuración de rutas de stock ratings
type StockRoutes struct {
	middlewareManager *MiddlewareManager
}

// NewStockRoutes crea una nueva instancia del configurador de rutas de stocks
func NewStockRoutes(middlewareManager *MiddlewareManager) *StockRoutes {
	return &StockRoutes{
		middlewareManager: middlewareManager,
	}
}

// SetupStockRoutes configura todas las rutas relacionadas con stock ratings
// Esta función configura tanto las operaciones CRUD como las consultas especializadas
func (sr *StockRoutes) SetupStockRoutes(routerGroup *gin.RouterGroup, stockHandler *handlers.StockHandler) {
	// Verificar que el handler existe
	if stockHandler == nil {
		return
	}

	// Configurar el grupo de rutas de stocks
	stocks := routerGroup.Group("/stocks")
	{
		// CRUD operations
		sr.setupCRUDRoutes(stocks, stockHandler)

		// Query operations
		sr.setupQueryRoutes(stocks, stockHandler)

		// Statistics operations
		sr.setupStatisticsRoutes(stocks, stockHandler)
	}
}

// setupCRUDRoutes configura las operaciones básicas CRUD
func (sr *StockRoutes) setupCRUDRoutes(stocks *gin.RouterGroup, stockHandler *handlers.StockHandler) {
	// Grupo para operaciones de escritura (CREATE, DELETE)
	writeOps := stocks.Group("")
	if sr.middlewareManager != nil {
		sr.middlewareManager.ApplyWriteMiddlewares(writeOps)
	}
	{
		// Create - Crear un nuevo rating de stock
		writeOps.POST("/", stockHandler.CreateStockRating)

		// Delete - Eliminar rating por ID
		writeOps.DELETE("/:id", stockHandler.DeleteStockRating)
	}

	// Grupo para operaciones de lectura (READ, LIST)
	readOps := stocks.Group("")
	if sr.middlewareManager != nil {
		sr.middlewareManager.ApplyReadOnlyMiddlewares(readOps)
	}
	{
		// Read - Obtener rating por ID
		readOps.GET("/:id", stockHandler.GetStockRatingByID)

		// List - Listar todos los ratings (con paginación)
		readOps.GET("/", stockHandler.ListStockRatings)
	}
}

// setupQueryRoutes configura las rutas de consultas especializadas
func (sr *StockRoutes) setupQueryRoutes(stocks *gin.RouterGroup, stockHandler *handlers.StockHandler) {
	// Grupo para operaciones de consulta (solo lectura pero pueden ser costosas)
	queryOps := stocks.Group("")
	if sr.middlewareManager != nil {
		sr.middlewareManager.ApplyReadOnlyMiddlewares(queryOps)
	}
	{
		// Consultas por entidad relacionada
		queryOps.GET("/company/:company_id", stockHandler.GetRatingsByCompany)
		queryOps.GET("/ticker/:ticker", stockHandler.GetRatingsByTicker)
		queryOps.GET("/brokerage/:brokerage_id", stockHandler.GetRatingsByBrokerage)

		// Consultas por tiempo
		queryOps.GET("/recent", stockHandler.GetRecentRatings)
		queryOps.GET("/date-range", stockHandler.GetRatingsByDateRange)
	}
}

// setupStatisticsRoutes configura las rutas de estadísticas y métricas
func (sr *StockRoutes) setupStatisticsRoutes(stocks *gin.RouterGroup, stockHandler *handlers.StockHandler) {
	// Grupo específico para estadísticas
	stats := stocks.Group("/stats")
	if sr.middlewareManager != nil {
		sr.middlewareManager.ApplyReadOnlyMiddlewares(stats)
	}
	{
		stats.GET("/company/:company_id", stockHandler.GetRatingStatsByCompany)
		// Futuras rutas de estadísticas se pueden agregar aquí
		// stats.GET("/brokerage/:brokerage_id", stockHandler.GetRatingStatsByBrokerage)
		// stats.GET("/sector/:sector", stockHandler.GetRatingStatsBySector)
	}
}

// GetStockRoutesInfo retorna información sobre las rutas de stocks disponibles
func (sr *StockRoutes) GetStockRoutesInfo() map[string]interface{} {
	return map[string]interface{}{
		"entity":    "stocks",
		"base_path": "/stocks",
		"operations": map[string][]string{
			"crud": {
				"POST /stocks",
				"GET /stocks/:id",
				"DELETE /stocks/:id",
				"GET /stocks",
			},
			"queries": {
				"GET /stocks/company/:company_id",
				"GET /stocks/ticker/:ticker",
				"GET /stocks/brokerage/:brokerage_id",
				"GET /stocks/recent",
				"GET /stocks/date-range",
			},
			"statistics": {
				"GET /stocks/stats/company/:company_id",
			},
		},
	}
}
