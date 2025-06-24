package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
)

// AlphaVantageRoutes encapsula la configuración de rutas de Alpha Vantage
type AlphaVantageRoutes struct {
	middlewareManager *MiddlewareManager
}

// NewAlphaVantageRoutes crea una nueva instancia del configurador de rutas de Alpha Vantage
func NewAlphaVantageRoutes(middlewareManager *MiddlewareManager) *AlphaVantageRoutes {
	return &AlphaVantageRoutes{
		middlewareManager: middlewareManager,
	}
}

// SetupAlphaVantageRoutes configura todas las rutas de Alpha Vantage usando el handler pattern
// @Summary Configure Alpha Vantage API routes
// @Description Sets up all routes for Alpha Vantage data retrieval including historical data,
//
//	technical indicators, financial metrics, and health checks
//
// @Tags alpha
// @Router /api/v1/alpha [group]
func (ar *AlphaVantageRoutes) SetupAlphaVantageRoutes(v1 *gin.RouterGroup, handler *handlers.AlphaVantageHandler) {
	// Crear grupo de rutas para Alpha Vantage
	alphaVantage := v1.Group("/alpha")

	// Aplicar middlewares específicos si es necesario
	// alphaVantage.Use(ar.middlewareManager.RateLimitMiddleware())

	// Configurar rutas con documentación Swagger actualizada
	{
		// Health check endpoint
		// @Summary Health check for Alpha Vantage API
		// @Description Checks if Alpha Vantage API is accessible and operational
		// @Tags alpha
		// @Accept json
		// @Produce json
		// @Success 200 {object} response.HealthCheckResponse
		// @Failure 500 {object} response.ErrorResponse
		// @Router /api/v1/alpha/health [get]
		alphaVantage.GET("/health", handler.HealthCheck)

		// Historical data endpoint
		// @Summary Get historical data for a symbol
		// @Description Retrieves historical price data from Alpha Vantage API
		// @Tags alpha
		// @Accept json
		// @Produce json
		// @Param symbol path string true "Stock symbol (e.g., AAPL)"
		// @Param period query string false "Time period: daily, weekly, monthly" default(daily)
		// @Param outputsize query string false "Output size: compact or full" default(compact)
		// @Success 200 {object} response.HistoricalDataResponse
		// @Failure 400 {object} response.ErrorResponse
		// @Failure 404 {object} response.ErrorResponse
		// @Failure 500 {object} response.ErrorResponse
		// @Router /api/v1/alpha/historical/{symbol} [get]
		alphaVantage.GET("/historical/:symbol", handler.GetHistoricalData)

		// Technical indicators endpoint
		// @Summary Get technical indicators for a symbol
		// @Description Retrieves technical indicators from Alpha Vantage API
		// @Tags alpha
		// @Accept json		// @Produce json
		// @Param symbol path string true "Stock symbol (e.g., AAPL)"
		// @Param indicator query string true "Indicator type: RSI, MACD, SMA, EMA, BBANDS"
		// @Param interval query string false "Time interval: daily, weekly, monthly" default(daily)
		// @Param time_period query string false "Time period for calculation" default(14)
		// @Param series_type query string false "Price series type: close, open, high, low" default(close)
		// @Success 200 {object} response.TechnicalIndicatorsResponse
		// @Failure 400 {object} response.ErrorResponse
		// @Failure 404 {object} response.ErrorResponse
		// @Failure 500 {object} response.ErrorResponse
		// @Router /api/v1/alpha/technical/{symbol} [get]
		alphaVantage.GET("/technical/:symbol", handler.GetTechnicalIndicators)

		// Financial metrics endpoint
		// @Summary Get financial metrics for a symbol
		// @Description Retrieves fundamental financial data from Alpha Vantage API
		// @Tags alpha
		// @Accept json
		// @Produce json
		// @Param symbol path string true "Stock symbol (e.g., AAPL)"
		// @Success 200 {object} response.FundamentalDataResponse
		// @Failure 400 {object} response.ErrorResponse
		// @Failure 404 {object} response.ErrorResponse
		// @Failure 500 {object} response.ErrorResponse
		// @Router /api/v1/alpha/financials/{symbol} [get]
		alphaVantage.GET("/financials/:symbol", handler.GetFinancialMetrics)

		// Earnings data endpoint
		// @Summary Get earnings data for a symbol
		// @Description Retrieves earnings data from Alpha Vantage API
		// @Tags alpha
		// @Accept json
		// @Produce json
		// @Param symbol path string true "Stock symbol (e.g., AAPL)"
		// @Success 200 {object} response.EarningsDataResponse
		// @Failure 400 {object} response.ErrorResponse
		// @Failure 404 {object} response.ErrorResponse
		// @Failure 500 {object} response.ErrorResponse
		// @Router /api/v1/alpha/earnings/{symbol} [get]
		alphaVantage.GET("/earnings/:symbol", handler.GetEarnings)
	}
}
