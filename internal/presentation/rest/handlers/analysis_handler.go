package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	serviceInterfaces "github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// AnalysisHandler maneja los endpoints relacionados con análisis y recomendaciones
type AnalysisHandler struct {
	analysisService serviceInterfaces.AnalysisService
	logger          logger.Logger
}

// NewAnalysisHandler crea una nueva instancia del handler de análisis
func NewAnalysisHandler(analysisService serviceInterfaces.AnalysisService, appLogger logger.Logger) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
		logger:          appLogger,
	}
}

// GetCompanyAnalysis godoc
// @Summary Get company analysis by ID
// @Description Get detailed analysis and recommendations for a specific company
// @Tags analysis
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} response.APIResponse[response.AnalysisResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/companies/{id} [get]
func (h *AnalysisHandler) GetCompanyAnalysis(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Parse and validate company ID
	companyIDStr := c.Param("id")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("company_id", companyIDStr),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Get company analysis
	analysisResp, err := h.analysisService.GetCompanyAnalysis(ctx, companyID)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Company analysis retrieval failed",
				logger.String("request_id", requestID),
				logger.String("company_id", companyID.String()),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during company analysis retrieval", err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to retrieve company analysis")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company analysis retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
		logger.String("ticker", analysisResp.Ticker),
		logger.Int("total_ratings", analysisResp.TotalRatings),
	)

	apiResponse := response.Success(analysisResp)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetCompanyAnalysisByTicker godoc
// @Summary Get company analysis by ticker
// @Description Get detailed analysis and recommendations for a company by ticker symbol
// @Tags analysis
// @Accept json
// @Produce json
// @Param ticker path string true "Company ticker symbol"
// @Success 200 {object} response.APIResponse[response.AnalysisResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/companies/ticker/{ticker} [get]
func (h *AnalysisHandler) GetCompanyAnalysisByTicker(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get ticker from path
	ticker := c.Param("ticker")
	if ticker == "" {
		h.logger.Warn(ctx, "Missing ticker parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Ticker parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Get company analysis by ticker
	analysisResp, err := h.analysisService.GetCompanyAnalysisByTicker(ctx, ticker)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Company analysis by ticker retrieval failed",
				logger.String("request_id", requestID),
				logger.String("ticker", ticker),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during company analysis by ticker retrieval", err,
			logger.String("request_id", requestID),
			logger.String("ticker", ticker),
		)

		errorResp := response.InternalServerError("Failed to retrieve company analysis")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company analysis by ticker retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("ticker", ticker),
		logger.String("company_id", analysisResp.CompanyID.String()),
		logger.Int("total_ratings", analysisResp.TotalRatings),
	)

	apiResponse := response.Success(analysisResp)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetMarketOverview godoc
// @Summary Get market overview
// @Description Get overall market statistics and overview
// @Tags analysis
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse[map[string]interface{}]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/market/overview [get]
func (h *AnalysisHandler) GetMarketOverview(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get market overview
	overview, err := h.analysisService.GetMarketOverview(ctx)
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
	)

	apiResponse := response.Success(overview)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetSectorAnalysis godoc
// @Summary Get sector analysis
// @Description Get analysis for a specific sector
// @Tags analysis
// @Accept json
// @Produce json
// @Param sector path string true "Sector name"
// @Success 200 {object} response.APIResponse[map[string]interface{}]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/sectors/{sector} [get]
func (h *AnalysisHandler) GetSectorAnalysis(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get sector from path
	sector := c.Param("sector")
	if sector == "" {
		h.logger.Warn(ctx, "Missing sector parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Sector parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Get sector analysis
	analysis, err := h.analysisService.GetSectorAnalysis(ctx, sector)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Sector analysis retrieval failed",
				logger.String("request_id", requestID),
				logger.String("sector", sector),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during sector analysis retrieval", err,
			logger.String("request_id", requestID),
			logger.String("sector", sector),
		)

		errorResp := response.InternalServerError("Failed to retrieve sector analysis")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Sector analysis retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("sector", sector),
	)

	apiResponse := response.Success(analysis)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetTopRatedCompanies godoc
// @Summary Get top rated companies
// @Description Get list of top rated companies based on rating count
// @Tags analysis
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of companies to return" default(10) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse[[]response.CompanyListResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/companies/top-rated [get]
func (h *AnalysisHandler) GetTopRatedCompanies(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Parse limit parameter
	limit := 10 // Default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err != nil {
			h.logger.Warn(ctx, "Invalid limit parameter",
				logger.String("request_id", requestID),
				logger.String("limit", limitStr),
			)

			errorResp := response.BadRequest("Invalid limit parameter")
			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		} else if l < 1 {
			limit = 1
		} else if l > 100 {
			limit = 100
		} else {
			limit = l
		}
	}

	// Get top rated companies
	companies, err := h.analysisService.GetTopRatedCompanies(ctx, limit)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Top rated companies retrieval failed",
				logger.String("request_id", requestID),
				logger.Int("limit", limit),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during top rated companies retrieval", err,
			logger.String("request_id", requestID),
			logger.Int("limit", limit),
		)

		errorResp := response.InternalServerError("Failed to retrieve top rated companies")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Top rated companies retrieved successfully",
		logger.String("request_id", requestID),
		logger.Int("limit", limit),
		logger.Int("count", len(companies)),
	)

	apiResponse := response.Success(companies)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRatingTrends godoc
// @Summary Get rating trends
// @Description Get rating trends over a specified time period
// @Tags analysis
// @Accept json
// @Produce json
// @Param period query string false "Time period (week, month, quarter, year)" default("month")
// @Success 200 {object} response.APIResponse[map[string]interface{}]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/trends/ratings [get]
func (h *AnalysisHandler) GetRatingTrends(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Parse period parameter
	period := c.DefaultQuery("period", "month")
	validPeriods := map[string]bool{
		"week":    true,
		"month":   true,
		"quarter": true,
		"year":    true,
	}

	if !validPeriods[period] {
		h.logger.Warn(ctx, "Invalid period parameter",
			logger.String("request_id", requestID),
			logger.String("period", period),
		)

		errorResp := response.BadRequest("Invalid period parameter. Valid values: week, month, quarter, year")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Get rating trends
	trends, err := h.analysisService.GetRatingTrends(ctx, period)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Rating trends retrieval failed",
				logger.String("request_id", requestID),
				logger.String("period", period),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during rating trends retrieval", err,
			logger.String("request_id", requestID),
			logger.String("period", period),
		)

		errorResp := response.InternalServerError("Failed to retrieve rating trends")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Rating trends retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("period", period),
	)

	apiResponse := response.Success(trends)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetBrokerageActivity godoc
// @Summary Get brokerage activity analysis
// @Description Get brokerage activity analysis over a specified time period
// @Tags analysis
// @Accept json
// @Produce json
// @Param period query string false "Time period (week, month, quarter, year)" default("month")
// @Success 200 {object} response.APIResponse[map[string]interface{}]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/trends/brokerages [get]
func (h *AnalysisHandler) GetBrokerageActivity(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Parse period parameter
	period := c.DefaultQuery("period", "month")
	validPeriods := map[string]bool{
		"week":    true,
		"month":   true,
		"quarter": true,
		"year":    true,
	}

	if !validPeriods[period] {
		h.logger.Warn(ctx, "Invalid period parameter",
			logger.String("request_id", requestID),
			logger.String("period", period),
		)

		errorResp := response.BadRequest("Invalid period parameter. Valid values: week, month, quarter, year")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Get brokerage activity
	activity, err := h.analysisService.GetBrokerageActivity(ctx, period)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage activity retrieval failed",
				logger.String("request_id", requestID),
				logger.String("period", period),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage activity retrieval", err,
			logger.String("request_id", requestID),
			logger.String("period", period),
		)

		errorResp := response.InternalServerError("Failed to retrieve brokerage activity")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage activity retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("period", period),
	)

	apiResponse := response.Success(activity)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GenerateRecommendation godoc
// @Summary Generate recommendation for a company
// @Description Generate investment recommendation for a specific company
// @Tags analysis
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} response.APIResponse[map[string]string]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/recommendations/companies/{id} [get]
func (h *AnalysisHandler) GenerateRecommendation(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Parse and validate company ID
	companyIDStr := c.Param("id")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("company_id", companyIDStr),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Generate recommendation
	recommendation, err := h.analysisService.GenerateRecommendation(ctx, companyID)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Recommendation generation failed",
				logger.String("request_id", requestID),
				logger.String("company_id", companyID.String()),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during recommendation generation", err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to generate recommendation")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Recommendation generated successfully",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
		logger.String("recommendation", recommendation),
	)

	result := map[string]string{
		"company_id":     companyID.String(),
		"recommendation": recommendation,
	}

	apiResponse := response.Success(result)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRecommendationsByRating godoc
// @Summary Get companies by recommendation rating
// @Description Get companies that have a specific recommendation rating
// @Tags analysis
// @Accept json
// @Produce json
// @Param rating path string true "Rating type (BUY, SELL, HOLD, etc.)"
// @Param limit query int false "Maximum number of companies to return" default(10) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse[[]response.CompanyListResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/analysis/recommendations/rating/{rating} [get]
func (h *AnalysisHandler) GetRecommendationsByRating(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Get rating from path
	rating := c.Param("rating")
	if rating == "" {
		h.logger.Warn(ctx, "Missing rating parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Rating parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Parse limit parameter
	limit := 10 // Default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err != nil {
			h.logger.Warn(ctx, "Invalid limit parameter",
				logger.String("request_id", requestID),
				logger.String("limit", limitStr),
			)

			errorResp := response.BadRequest("Invalid limit parameter")
			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		} else if l < 1 {
			limit = 1
		} else if l > 100 {
			limit = 100
		} else {
			limit = l
		}
	}

	// Get recommendations by rating
	companies, err := h.analysisService.GetRecommendationsByRating(ctx, rating, limit)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Recommendations by rating retrieval failed",
				logger.String("request_id", requestID),
				logger.String("rating", rating),
				logger.Int("limit", limit),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during recommendations by rating retrieval", err,
			logger.String("request_id", requestID),
			logger.String("rating", rating),
			logger.Int("limit", limit),
		)

		errorResp := response.InternalServerError("Failed to retrieve recommendations by rating")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Recommendations by rating retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("rating", rating),
		logger.Int("limit", limit),
		logger.Int("count", len(companies)),
	)

	apiResponse := response.Success(companies)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}
