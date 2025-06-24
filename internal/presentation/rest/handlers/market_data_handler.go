package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// MarketDataHandler handles market data related requests
type MarketDataHandler struct {
	marketDataService interfaces.MarketDataService
	logger            logger.Logger
}

// NewMarketDataHandler creates a new market data handler
func NewMarketDataHandler(marketDataService interfaces.MarketDataService, logger logger.Logger) *MarketDataHandler {
	return &MarketDataHandler{
		marketDataService: marketDataService,
		logger:            logger,
	}
}

// GetRealTimeQuote godoc
// @Summary Get real-time quote for a stock
// @Description Get real-time market data for a specific stock symbol
// @Tags market-data
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Success 200 {object} response.APIResponse[response.MarketDataResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/market-data/quote/{symbol} [get]
func (h *MarketDataHandler) GetRealTimeQuote(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get symbol from path
	symbol := c.Param("symbol")
	if symbol == "" {
		h.logger.Warn(ctx, "Missing symbol parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Symbol parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Getting real-time quote",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
	)

	// Get market data
	marketData, err := h.marketDataService.GetRealTimeQuote(ctx, symbol)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Market data retrieval failed",
				logger.String("request_id", requestID),
				logger.String("symbol", symbol),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during market data retrieval", err,
			logger.String("request_id", requestID),
			logger.String("symbol", symbol),
		)

		errorResp := response.InternalServerError("Failed to retrieve market data")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Market data retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
		logger.Float64("price", marketData.CurrentPrice),
	)

	apiResponse := response.Success(marketData)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetCompanyProfile godoc
// @Summary Get company profile
// @Description Get detailed company profile information
// @Tags market-data
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Success 200 {object} response.APIResponse[response.CompanyProfileResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/market-data/profile/{symbol} [get]
func (h *MarketDataHandler) GetCompanyProfile(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get symbol from path
	symbol := c.Param("symbol")
	if symbol == "" {
		h.logger.Warn(ctx, "Missing symbol parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Symbol parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Getting company profile",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
	)

	// Get company profile
	profile, err := h.marketDataService.GetCompanyProfile(ctx, symbol)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Company profile retrieval failed",
				logger.String("request_id", requestID),
				logger.String("symbol", symbol),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during company profile retrieval", err,
			logger.String("request_id", requestID),
			logger.String("symbol", symbol),
		)

		errorResp := response.InternalServerError("Failed to retrieve company profile")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company profile retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
		logger.String("company_name", profile.Name),
	)

	apiResponse := response.Success(profile)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetCompanyNews godoc
// @Summary Get company news
// @Description Get recent news for a specific company
// @Tags market-data
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Param days query int false "Number of days to look back" default(7) minimum(1) maximum(30)
// @Success 200 {object} response.APIResponse[[]response.NewsResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/market-data/news/{symbol} [get]
func (h *MarketDataHandler) GetCompanyNews(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get symbol from path
	symbol := c.Param("symbol")
	if symbol == "" {
		h.logger.Warn(ctx, "Missing symbol parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Symbol parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Parse days parameter
	days := 7 // Default
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err != nil {
			h.logger.Warn(ctx, "Invalid days parameter",
				logger.String("request_id", requestID),
				logger.String("days", daysStr),
			)

			errorResp := response.BadRequest("Invalid days parameter")
			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		} else if d < 1 {
			days = 1
		} else if d > 30 {
			days = 30
		} else {
			days = d
		}
	}

	h.logger.Info(ctx, "Getting company news",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
		logger.Int("days", days),
	)

	// Get company news
	news, err := h.marketDataService.GetCompanyNews(ctx, symbol, days)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Company news retrieval failed",
				logger.String("request_id", requestID),
				logger.String("symbol", symbol),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during company news retrieval", err,
			logger.String("request_id", requestID),
			logger.String("symbol", symbol),
		)

		errorResp := response.InternalServerError("Failed to retrieve company news")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company news retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
		logger.Int("news_count", len(news)),
	)

	apiResponse := response.Success(news)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetBasicFinancials godoc
// @Summary Get basic financial metrics
// @Description Get basic financial metrics for a specific company
// @Tags market-data
// @Accept json
// @Produce json
// @Param symbol path string true "Stock symbol (e.g., AAPL)"
// @Success 200 {object} response.APIResponse[response.BasicFinancialsResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/market-data/financials/{symbol} [get]
func (h *MarketDataHandler) GetBasicFinancials(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get symbol from path
	symbol := c.Param("symbol")
	if symbol == "" {
		h.logger.Warn(ctx, "Missing symbol parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Symbol parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Getting basic financials",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
	)

	// Get basic financials
	financials, err := h.marketDataService.GetBasicFinancials(ctx, symbol)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Basic financials retrieval failed",
				logger.String("request_id", requestID),
				logger.String("symbol", symbol),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during basic financials retrieval", err,
			logger.String("request_id", requestID),
			logger.String("symbol", symbol),
		)

		errorResp := response.InternalServerError("Failed to retrieve basic financials")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Basic financials retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("symbol", symbol),
	)

	apiResponse := response.Success(financials)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetMarketOverview godoc
// @Summary Get market overview
// @Description Get general market overview statistics
// @Tags market-data
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse[response.MarketOverviewResponse]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/market-data/overview [get]
func (h *MarketDataHandler) GetMarketOverview(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	h.logger.Info(ctx, "Getting market overview",
		logger.String("request_id", requestID),
	)

	// Get market overview
	overview, err := h.marketDataService.GetMarketOverview(ctx)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Market overview retrieval failed",
				logger.String("request_id", requestID),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during market overview retrieval", err,
			logger.String("request_id", requestID),
		)

		errorResp := response.InternalServerError("Failed to retrieve market overview")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Market overview retrieved successfully",
		logger.String("request_id", requestID),
		logger.Int("total_stocks", overview.TotalStocks),
	)

	apiResponse := response.Success(overview)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}
