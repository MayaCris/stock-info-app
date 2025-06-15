package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// AlphaVantageController handles Alpha Vantage API endpoints
type AlphaVantageController struct {
	marketDataService interfaces.MarketDataService
	logger            logger.Logger
}

// NewAlphaVantageController creates a new Alpha Vantage controller
func NewAlphaVantageController(service interfaces.MarketDataService, log logger.Logger) *AlphaVantageController {
	return &AlphaVantageController{
		marketDataService: service,
		logger:            log,
	}
}

// GetHistoricalData retrieves historical price data
// @Summary Get historical data for a symbol
// @Description Retrieves historical price data from Alpha Vantage API
// @Tags alpha-vantage
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Param period query string false "Time period: daily, weekly, monthly" default(daily)
// @Param outputsize query string false "Output size: compact or full" default(compact)
// @Success 200 {object} response.HistoricalDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha-vantage/historical/{symbol} [get]
func (c *AlphaVantageController) GetHistoricalData(ctx *gin.Context) {
	symbol := ctx.Param("symbol")
	period := ctx.DefaultQuery("period", "daily")
	outputSize := ctx.DefaultQuery("outputsize", "compact")

	if symbol == "" {
		c.logger.Warn(ctx.Request.Context(), "Missing symbol parameter")
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}

	c.logger.Info(ctx.Request.Context(), "Fetching historical data",
		logger.String("symbol", symbol),
		logger.String("period", period),
		logger.String("output_size", outputSize))

	// This would be implemented in the enhanced MarketDataService
	historicalData, err := c.marketDataService.GetHistoricalData(ctx.Request.Context(), symbol, period, outputSize)
	if err != nil {
		c.logger.Error(ctx.Request.Context(), "Failed to get historical data", err,
			logger.String("symbol", symbol))
		ctx.JSON(500, response.InternalServerError("Failed to retrieve historical data"))
		return
	}

	ctx.JSON(200, historicalData)
}

// GetTechnicalIndicators retrieves technical indicators
// @Summary Get technical indicators for a symbol
// @Description Retrieves technical indicators from Alpha Vantage API
// @Tags alpha-vantage
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Param indicator query string true "Indicator type: RSI, MACD, SMA, EMA, BBANDS"
// @Param interval query string false "Time interval: daily, weekly, monthly" default(daily)
// @Param time_period query string false "Time period for calculation" default(14)
// @Success 200 {object} response.TechnicalIndicatorsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha-vantage/technical/{symbol} [get]
func (c *AlphaVantageController) GetTechnicalIndicators(ctx *gin.Context) {
	symbol := ctx.Param("symbol")
	indicator := ctx.Query("indicator")
	interval := ctx.DefaultQuery("interval", "daily")
	timePeriod := ctx.DefaultQuery("time_period", "14")

	if symbol == "" {
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}

	if indicator == "" {
		ctx.JSON(400, response.BadRequest("Indicator is required"))
		return
	}

	c.logger.Info(ctx.Request.Context(), "Fetching technical indicators",
		logger.String("symbol", symbol),
		logger.String("indicator", indicator),
		logger.String("interval", interval))

	// This would be implemented in the enhanced MarketDataService
	indicators, err := c.marketDataService.GetTechnicalIndicators(ctx.Request.Context(), symbol, indicator, interval, timePeriod)
	if err != nil {
		c.logger.Error(ctx.Request.Context(), "Failed to get technical indicators", err,
			logger.String("symbol", symbol),
			logger.String("indicator", indicator))
		ctx.JSON(500, response.InternalServerError("Failed to retrieve technical indicators"))
		return
	}

	ctx.JSON(200, indicators)
}

// GetFundamentalData retrieves fundamental company data
// @Summary Get fundamental data for a symbol
// @Description Retrieves fundamental financial data from Alpha Vantage API
// @Tags alpha-vantage
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Success 200 {object} response.FundamentalDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha-vantage/fundamental/{symbol} [get]
func (c *AlphaVantageController) GetFundamentalData(ctx *gin.Context) {
	symbol := ctx.Param("symbol")

	if symbol == "" {
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}

	c.logger.Info(ctx.Request.Context(), "Fetching fundamental data",
		logger.String("symbol", symbol))

	// This would be implemented in the enhanced MarketDataService
	fundamental, err := c.marketDataService.GetFundamentalData(ctx.Request.Context(), symbol)
	if err != nil {
		c.logger.Error(ctx.Request.Context(), "Failed to get fundamental data", err,
			logger.String("symbol", symbol))
		ctx.JSON(500, response.InternalServerError("Failed to retrieve fundamental data"))
		return
	}

	ctx.JSON(200, fundamental)
}

// GetEarningsData retrieves earnings data
// @Summary Get earnings data for a symbol
// @Description Retrieves earnings data from Alpha Vantage API
// @Tags alpha-vantage
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Success 200 {object} response.EarningsDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha-vantage/earnings/{symbol} [get]
func (c *AlphaVantageController) GetEarningsData(ctx *gin.Context) {
	symbol := ctx.Param("symbol")

	if symbol == "" {
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}

	c.logger.Info(ctx.Request.Context(), "Fetching earnings data",
		logger.String("symbol", symbol))

	// This would be implemented in the enhanced MarketDataService
	earnings, err := c.marketDataService.GetEarningsData(ctx.Request.Context(), symbol)
	if err != nil {
		c.logger.Error(ctx.Request.Context(), "Failed to get earnings data", err,
			logger.String("symbol", symbol))
		ctx.JSON(500, response.InternalServerError("Failed to retrieve earnings data"))
		return
	}

	ctx.JSON(200, earnings)
}

// HealthCheck checks Alpha Vantage API connectivity
// @Summary Health check for Alpha Vantage API
// @Description Checks if Alpha Vantage API is accessible
// @Tags alpha-vantage
// @Accept json
// @Produce json
// @Success 200 {object} response.HealthCheckResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha-vantage/health [get]
func (c *AlphaVantageController) HealthCheck(ctx *gin.Context) {
	c.logger.Info(ctx.Request.Context(), "Performing Alpha Vantage health check")

	// This would be implemented in the enhanced MarketDataService
	healthy, err := c.marketDataService.AlphaVantageHealthCheck(ctx.Request.Context())
	if err != nil {
		c.logger.Error(ctx.Request.Context(), "Alpha Vantage health check failed", err)
		ctx.JSON(500, response.InternalServerError("Alpha Vantage API is not healthy"))
		return
	}

	if !healthy {
		ctx.JSON(500, response.InternalServerError("Alpha Vantage API is not responding"))
		return
	}

	ctx.JSON(200, gin.H{
		"status":    "healthy",
		"service":   "alpha-vantage",
		"timestamp": time.Now(),
	})
}

// RegisterAlphaVantageRoutes registers all Alpha Vantage routes
func RegisterAlphaVantageRoutes(router *gin.RouterGroup, controller *AlphaVantageController) {
	alphaVantage := router.Group("/alpha-vantage")
	{
		alphaVantage.GET("/health", controller.HealthCheck)
		alphaVantage.GET("/historical/:symbol", controller.GetHistoricalData)
		alphaVantage.GET("/technical/:symbol", controller.GetTechnicalIndicators)
		alphaVantage.GET("/fundamental/:symbol", controller.GetFundamentalData)
		alphaVantage.GET("/earnings/:symbol", controller.GetEarningsData)
	}
}
