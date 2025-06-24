package handlers

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/application/dto/request"
	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	serviceInterfaces "github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// StockHandler maneja los endpoints relacionados con stock ratings
type StockHandler struct {
	stockService serviceInterfaces.StockRatingService
	logger       logger.Logger
}

// NewStockHandler crea una nueva instancia del handler de stocks
func NewStockHandler(stockService serviceInterfaces.StockRatingService, appLogger logger.Logger) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		logger:       appLogger,
	}
}

// CreateStockRating godoc
// @Summary Create a new stock rating
// @Description Create a new stock rating with the provided details
// @Tags stocks
// @Accept json
// @Produce json
// @Param stock body request.CreateStockRatingRequest true "Stock rating creation details"
// @Success 201 {object} response.APIResponse[response.StockRatingResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 422 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks [post]
func (h *StockHandler) CreateStockRating(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	var req request.CreateStockRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(ctx, "Invalid request body for stock rating creation",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Invalid request body")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Creating stock rating",
		logger.String("request_id", requestID),
		logger.String("company_id", req.CompanyID.String()),
		logger.String("brokerage_id", req.BrokerageID.String()),
		logger.String("action", req.Action),
	)

	stockRating, err := h.stockService.CreateStockRating(ctx, &req)
	if err != nil {
		h.logger.Error(ctx, "Failed to create stock rating",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", req.CompanyID.String()),
			logger.String("brokerage_id", req.BrokerageID.String()),
		)

		errorResp := response.InternalServerError("Failed to create stock rating")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Stock rating created successfully",
		logger.String("request_id", requestID),
		logger.String("stock_rating_id", stockRating.ID.String()),
	)

	apiResponse := response.Success(stockRating)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusCreated, apiResponse)
}

// GetStockRatingByID godoc
// @Summary Get stock rating by ID
// @Description Get a specific stock rating by its ID
// @Tags stocks
// @Accept json
// @Produce json
// @Param id path string true "Stock Rating ID"
// @Success 200 {object} response.APIResponse[response.StockRatingResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/{id} [get]
func (h *StockHandler) GetStockRatingByID(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid stock rating ID format",
			logger.String("request_id", requestID),
			logger.String("id", idParam),
		)

		errorResp := response.BadRequest("Invalid stock rating ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Getting stock rating by ID",
		logger.String("request_id", requestID),
		logger.String("id", id.String()),
	)

	stockRating, err := h.stockService.GetStockRatingByID(ctx, id)
	if err != nil {
		h.logger.Error(ctx, "Failed to get stock rating",
			err,
			logger.String("request_id", requestID),
			logger.String("id", id.String()),
		)

		errorResp := response.NotFound("Stock rating")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(stockRating)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// DeleteStockRating godoc
// @Summary Delete stock rating
// @Description Delete a stock rating by ID
// @Tags stocks
// @Accept json
// @Produce json
// @Param id path string true "Stock Rating ID"
// @Success 204 "No content"
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/{id} [delete]
func (h *StockHandler) DeleteStockRating(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid stock rating ID format",
			logger.String("request_id", requestID),
			logger.String("id", idParam),
		)

		errorResp := response.BadRequest("Invalid stock rating ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Deleting stock rating",
		logger.String("request_id", requestID),
		logger.String("id", id.String()),
	)

	err = h.stockService.DeleteStockRating(ctx, id)
	if err != nil {
		h.logger.Error(ctx, "Failed to delete stock rating",
			err,
			logger.String("request_id", requestID),
			logger.String("id", id.String()),
		)

		errorResp := response.InternalServerError("Failed to delete stock rating")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Stock rating deleted successfully",
		logger.String("request_id", requestID),
		logger.String("id", id.String()),
	)

	c.Status(http.StatusNoContent)
}

// ListStockRatings godoc
// @Summary List stock ratings
// @Description Get a paginated list of stock ratings with optional filters
// @Tags stocks
// @Accept json
// @Produce json
// @Param company_id query string false "Company ID filter"
// @Param brokerage_id query string false "Brokerage ID filter"
// @Param ticker query string false "Company ticker filter"
// @Param action query string false "Rating action filter"
// @Param rating_to query string false "Rating to filter"
// @Param date_from query string false "Date from filter (YYYY-MM-DD)"
// @Param date_to query string false "Date to filter (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.StockRatingListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks [get]
func (h *StockHandler) ListStockRatings(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Parse filters
	var filter request.StockRatingFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Warn(ctx, "Invalid query parameters",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Invalid query parameters")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Parse pagination
	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Listing stock ratings",
		logger.String("request_id", requestID),
		logger.Int("page", pagination.Page),
		logger.Int("per_page", pagination.PerPage),
	)

	stockRatings, err := h.stockService.ListStockRatings(ctx, &filter, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to list stock ratings",
			err,
			logger.String("request_id", requestID),
		)

		errorResp := response.InternalServerError("Failed to list stock ratings")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(stockRatings)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRatingsByCompany godoc
// @Summary Get ratings by company
// @Description Get stock ratings for a specific company
// @Tags stocks
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.StockRatingListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/company/{company_id} [get]
func (h *StockHandler) GetRatingsByCompany(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	companyIDParam := c.Param("company_id")
	companyID, err := uuid.Parse(companyIDParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("company_id", companyIDParam),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Getting ratings by company",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	stockRatings, err := h.stockService.GetRatingsByCompany(ctx, companyID, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to get ratings by company",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to get ratings by company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(stockRatings)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRatingsByTicker godoc
// @Summary Get ratings by ticker
// @Description Get stock ratings for a specific company ticker
// @Tags stocks
// @Accept json
// @Produce json
// @Param ticker path string true "Company ticker"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.StockRatingListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/ticker/{ticker} [get]
func (h *StockHandler) GetRatingsByTicker(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	ticker := c.Param("ticker")
	if ticker == "" {
		h.logger.Warn(ctx, "Empty ticker parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Ticker parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Getting ratings by ticker",
		logger.String("request_id", requestID),
		logger.String("ticker", ticker),
	)

	stockRatings, err := h.stockService.GetRatingsByTicker(ctx, ticker, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to get ratings by ticker",
			err,
			logger.String("request_id", requestID),
			logger.String("ticker", ticker),
		)

		errorResp := response.InternalServerError("Failed to get ratings by ticker")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(stockRatings)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRatingsByBrokerage godoc
// @Summary Get ratings by brokerage
// @Description Get stock ratings for a specific brokerage
// @Tags stocks
// @Accept json
// @Produce json
// @Param brokerage_id path string true "Brokerage ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.StockRatingListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/brokerage/{brokerage_id} [get]
func (h *StockHandler) GetRatingsByBrokerage(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	brokerageIDParam := c.Param("brokerage_id")
	brokerageID, err := uuid.Parse(brokerageIDParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid brokerage ID format",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageIDParam),
		)

		errorResp := response.BadRequest("Invalid brokerage ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Getting ratings by brokerage",
		logger.String("request_id", requestID),
		logger.String("brokerage_id", brokerageID.String()),
	)

	stockRatings, err := h.stockService.GetRatingsByBrokerage(ctx, brokerageID, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to get ratings by brokerage",
			err,
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
		)

		errorResp := response.InternalServerError("Failed to get ratings by brokerage")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(stockRatings)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRecentRatings godoc
// @Summary Get recent ratings
// @Description Get the most recent stock ratings
// @Tags stocks
// @Accept json
// @Produce json
// @Param limit query int false "Number of recent ratings to return" default(10)
// @Success 200 {object} response.APIResponse[[]response.StockRatingListResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/recent [get]
func (h *StockHandler) GetRecentRatings(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	limitParam := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 || limit > 100 {
		h.logger.Warn(ctx, "Invalid limit parameter",
			logger.String("request_id", requestID),
			logger.String("limit", limitParam),
		)

		errorResp := response.BadRequest("Invalid limit parameter (must be between 1 and 100)")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Getting recent ratings",
		logger.String("request_id", requestID),
		logger.Int("limit", limit),
	)

	recentRatings, err := h.stockService.GetRecentRatings(ctx, limit)
	if err != nil {
		h.logger.Error(ctx, "Failed to get recent ratings",
			err,
			logger.String("request_id", requestID),
		)

		errorResp := response.InternalServerError("Failed to get recent ratings")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(recentRatings)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRatingsByDateRange godoc
// @Summary Get ratings by date range
// @Description Get stock ratings within a specific date range
// @Tags stocks
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.StockRatingListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/date-range [get]
func (h *StockHandler) GetRatingsByDateRange(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		h.logger.Warn(ctx, "Missing date range parameters",
			logger.String("request_id", requestID),
			logger.String("start_date", startDate),
			logger.String("end_date", endDate),
		)

		errorResp := response.BadRequest("Both start_date and end_date parameters are required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Getting ratings by date range",
		logger.String("request_id", requestID),
		logger.String("start_date", startDate),
		logger.String("end_date", endDate),
	)

	stockRatings, err := h.stockService.GetRatingsByDateRange(ctx, startDate, endDate, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to get ratings by date range",
			err,
			logger.String("request_id", requestID),
			logger.String("start_date", startDate),
			logger.String("end_date", endDate),
		)

		errorResp := response.InternalServerError("Failed to get ratings by date range")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(stockRatings)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetRatingStatsByCompany godoc
// @Summary Get rating statistics by company
// @Description Get statistical information about ratings for a specific company
// @Tags stocks
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Success 200 {object} response.APIResponse[map[string]interface{}]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/stocks/company/{company_id}/stats [get]
func (h *StockHandler) GetRatingStatsByCompany(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	companyIDParam := c.Param("company_id")
	companyID, err := uuid.Parse(companyIDParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("company_id", companyIDParam),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Getting rating stats by company",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	stats, err := h.stockService.GetRatingStatsByCompany(ctx, companyID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get rating stats by company",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to get rating statistics")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(stats)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// parsePagination extrae y valida los parámetros de paginación
func (h *StockHandler) parsePagination(c *gin.Context) *response.PaginationRequest {
	pageParam := c.Query("page")
	perPageParam := c.Query("per_page")
	
	return response.ParsePaginationFromQuery(pageParam, perPageParam)
}
