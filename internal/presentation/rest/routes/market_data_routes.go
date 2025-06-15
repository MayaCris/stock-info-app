package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
)

// MarketDataRoutes encapsula la configuración de rutas de market data
type MarketDataRoutes struct {
	middlewareManager *MiddlewareManager
}

// NewMarketDataRoutes crea una nueva instancia del configurador de rutas de market data
func NewMarketDataRoutes(middlewareManager *MiddlewareManager) *MarketDataRoutes {
	return &MarketDataRoutes{
		middlewareManager: middlewareManager,
	}
}

// SetupMarketDataRoutes configura todas las rutas de market data
func (mr *MarketDataRoutes) SetupMarketDataRoutes(group *gin.RouterGroup, handler *handlers.MarketDataHandler) {
	// Market data base group
	marketData := group.Group("/market-data")
	{
		// Real-time market data endpoints
		marketData.GET("/quote/:symbol", handler.GetRealTimeQuote)

		// Company profile endpoints
		marketData.GET("/profile/:symbol", handler.GetCompanyProfile)

		// News endpoints
		marketData.GET("/news/:symbol", handler.GetCompanyNews)

		// Financial metrics endpoints
		marketData.GET("/financials/:symbol", handler.GetBasicFinancials)

		// Market overview endpoints
		marketData.GET("/overview", handler.GetMarketOverview)
	}
}
