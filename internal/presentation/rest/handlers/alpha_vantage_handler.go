package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// AlphaVantageHandler handles Alpha Vantage API endpoints
type AlphaVantageHandler struct {
	alphaVantageService interfaces.AlphaVantageService
	logger              logger.Logger
}

// NewAlphaVantageHandler creates a new Alpha Vantage handler
func NewAlphaVantageHandler(service interfaces.AlphaVantageService, log logger.Logger) *AlphaVantageHandler {
	return &AlphaVantageHandler{
		alphaVantageService: service,
		logger:              log,
	}
}

// GetHistoricalData retrieves historical price data
// @Summary Get historical data for a symbol
// @Description Retrieves historical price data from Alpha Vantage API
// @Tags alpha
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Param period query string false "Time period: daily, weekly, monthly" default(daily)
// @Param outputsize query string false "Output size: compact or full" default(compact)
// @Param interval query string false "Interval for intraday data: 1min, 5min, 15min, 30min, 60min"
// @Param adjusted query string false "Whether to return adjusted data: true or false"
// @Success 200 {object} response.HistoricalDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha/historical/{symbol} [get]
func (h *AlphaVantageHandler) GetHistoricalData(ctx *gin.Context) {
	symbol := ctx.Param("symbol")
	period := ctx.DefaultQuery("period", "daily")
	outputSize := ctx.DefaultQuery("outputsize", "compact")
	interval := ctx.DefaultQuery("interval", "")
	adjusted := ctx.DefaultQuery("adjusted", "")

	if symbol == "" {
		h.logger.Warn(ctx.Request.Context(), "Missing symbol parameter")
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}

	start := time.Now()

	data, err := h.alphaVantageService.GetHistoricalDataFromAPI(ctx.Request.Context(), symbol, period, outputSize, interval, adjusted)
	if err != nil {
		h.logger.Error(ctx.Request.Context(), "Failed to get historical data", err,
			logger.String("symbol", symbol),
			logger.String("period", period),
			logger.String("outputsize", outputSize),
			logger.String("interval", interval),
			logger.String("adjusted", adjusted))

		ctx.JSON(500, response.InternalServerError("Failed to retrieve historical data"))
		return
	}
	h.logger.Info(ctx.Request.Context(), "Historical data retrieved successfully",
		logger.String("symbol", symbol),
		logger.String("period", period),
		logger.String("outputsize", outputSize),
		logger.String("interval", interval),
		logger.String("adjusted", adjusted),
		logger.Duration("duration", time.Since(start)))

	ctx.JSON(200, response.Success(data))
}

// GetTechnicalIndicators retrieves technical indicators
// @Summary Get technical indicators for a symbol
// @Description Retrieves technical indicators from Alpha Vantage API
// @Tags alpha
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Param indicator query string true "Technical indicator (e.g., RSI, MACD, SMA)"
// @Param interval query string false "Time interval" default(daily)
// @Param time_period query int false "Time period for calculation" default(14)
// @Success 200 {object} response.TechnicalIndicatorResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha/technical/{symbol} [get]
func (h *AlphaVantageHandler) GetTechnicalIndicators(ctx *gin.Context) {
	symbol := ctx.Param("symbol")
	indicator := ctx.Query("indicator")
	interval := ctx.DefaultQuery("interval", "daily")
	timePeriod := ctx.DefaultQuery("time_period", "14")
	seriesType := ctx.DefaultQuery("series_type", "close")

	if symbol == "" {
		h.logger.Warn(ctx.Request.Context(), "Missing symbol parameter")
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}
	if indicator == "" {
		h.logger.Warn(ctx.Request.Context(), "Missing indicator parameter")
		ctx.JSON(400, response.BadRequest("Indicator is required"))
		return
	}
	start := time.Now()
	// Use the specific indicator method with parameters
	data, err := h.alphaVantageService.GetTechnicalIndicatorFromAPI(ctx.Request.Context(), symbol, indicator, interval, timePeriod, seriesType)
	if err != nil {
		h.logger.Error(ctx.Request.Context(), "Failed to get technical indicators", err,
			logger.String("symbol", symbol),
			logger.String("indicator", indicator),
			logger.String("interval", interval),
			logger.String("time_period", timePeriod),
			logger.String("series_type", seriesType))

		ctx.JSON(500, response.InternalServerError("Failed to retrieve technical indicators"))
		return
	}

	h.logger.Info(ctx.Request.Context(), "Technical indicators retrieved successfully",
		logger.String("symbol", symbol),
		logger.String("indicator", indicator),
		logger.String("interval", interval),
		logger.String("time_period", timePeriod),
		logger.String("series_type", seriesType),
		logger.Duration("duration", time.Since(start)))

	ctx.JSON(200, response.Success(data))
}

// GetFinancialMetrics retrieves financial metrics
// @Summary Get financial metrics for a symbol
// @Description Retrieves financial metrics from Alpha Vantage API
// @Tags alpha
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Param function query string false "Function type" default(OVERVIEW)
// @Success 200 {object} response.FinancialMetricsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha/financials/{symbol} [get]
func (h *AlphaVantageHandler) GetFinancialMetrics(ctx *gin.Context) {
	symbol := ctx.Param("symbol")
	function := ctx.DefaultQuery("function", "OVERVIEW")

	if symbol == "" {
		h.logger.Warn(ctx.Request.Context(), "Missing symbol parameter")
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}
	start := time.Now()

	data, err := h.alphaVantageService.GetFinancialMetricsFromAPI(ctx.Request.Context(), symbol)
	if err != nil {
		h.logger.Error(ctx.Request.Context(), "Failed to get financial metrics", err,
			logger.String("symbol", symbol),
			logger.String("function", function))

		ctx.JSON(500, response.InternalServerError("Failed to retrieve financial metrics"))
		return
	}

	h.logger.Info(ctx.Request.Context(), "Financial metrics retrieved successfully",
		logger.String("symbol", symbol),
		logger.String("function", function),
		logger.Duration("duration", time.Since(start)))

	ctx.JSON(200, response.Success(data))
}

// GetEarnings retrieves earnings data
// @Summary Get earnings data for a symbol
// @Description Retrieves earnings data from Alpha Vantage API
// @Tags alpha
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Success 200 {object} response.EarningsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha/earnings/{symbol} [get]
func (h *AlphaVantageHandler) GetEarnings(ctx *gin.Context) {
	symbol := ctx.Param("symbol")

	if symbol == "" {
		h.logger.Warn(ctx.Request.Context(), "Missing symbol parameter")
		ctx.JSON(400, response.BadRequest("Symbol is required"))
		return
	}

	start := time.Now()

	// For now, return a placeholder response since earnings method might not be implemented yet
	data := map[string]interface{}{
		"symbol":  symbol,
		"message": "Earnings data endpoint - implementation pending",
	}

	h.logger.Info(ctx.Request.Context(), "Earnings data request processed",
		logger.String("symbol", symbol),
		logger.Duration("duration", time.Since(start)))

	ctx.JSON(200, response.Success(data))
}

// HealthCheck performs a health check on the Alpha Vantage service
// @Summary Health check for Alpha Vantage service
// @Description Checks if Alpha Vantage service is healthy and responsive
// @Tags alpha
// @Accept json
// @Produce json
// @Success 200 {object} response.HealthResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/alpha/health [get]
func (h *AlphaVantageHandler) HealthCheck(ctx *gin.Context) {
	start := time.Now()
	// Use the service to perform a basic health check
	// This could involve checking API connectivity, rate limits, etc.
	isHealthy := true

	// You could add actual health check logic here
	// For example: err := h.alphaVantageService.HealthCheck(ctx.Request.Context())

	h.logger.Info(ctx.Request.Context(), "Alpha Vantage health check completed",
		logger.Bool("healthy", isHealthy),
		logger.Duration("duration", time.Since(start)))
	if isHealthy {
		ctx.JSON(200, response.Success(map[string]interface{}{
			"service":   "alpha_vantage",
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		}))
	} else {
		ctx.JSON(500, response.InternalServerError("Alpha Vantage service is unhealthy"))
	}
}
