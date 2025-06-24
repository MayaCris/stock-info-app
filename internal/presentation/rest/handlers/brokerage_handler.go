package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/application/dto/request"
	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	serviceInterfaces "github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// BrokerageHandler maneja los endpoints relacionados con brokerages
type BrokerageHandler struct {
	brokerageService serviceInterfaces.BrokerageService
	logger           logger.Logger
}

// NewBrokerageHandler crea una nueva instancia del handler de brokerages
func NewBrokerageHandler(brokerageService serviceInterfaces.BrokerageService, appLogger logger.Logger) *BrokerageHandler {
	return &BrokerageHandler{
		brokerageService: brokerageService,
		logger:           appLogger,
	}
}

// CreateBrokerage godoc
// @Summary Create a new brokerage
// @Description Create a new brokerage with the provided details
// @Tags brokerages
// @Accept json
// @Produce json
// @Param brokerage body request.CreateBrokerageRequest true "Brokerage creation details"
// @Success 201 {object} response.APIResponse[response.BrokerageResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 409 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages [post]
func (h *BrokerageHandler) CreateBrokerage(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	var req request.CreateBrokerageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(ctx, "Invalid request body for brokerage creation",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Invalid request body")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Warn(ctx, "Brokerage creation validation failed",
			logger.String("request_id", requestID),
			logger.String("name", req.Name),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Validation failed")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Create brokerage
	brokerageResp, err := h.brokerageService.CreateBrokerage(ctx, &req)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage creation failed",
				logger.String("request_id", requestID),
				logger.String("name", req.Name),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage creation", err,
			logger.String("request_id", requestID),
			logger.String("name", req.Name),
		)

		errorResp := response.InternalServerError("Failed to create brokerage")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage created successfully",
		logger.String("request_id", requestID),
		logger.String("brokerage_id", brokerageResp.ID.String()),
		logger.String("name", brokerageResp.Name),
	)

	apiResponse := response.Success(brokerageResp)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusCreated, apiResponse)
}

// GetBrokerageByID godoc
// @Summary Get brokerage by ID
// @Description Get detailed information about a specific brokerage
// @Tags brokerages
// @Accept json
// @Produce json
// @Param id path string true "Brokerage ID"
// @Success 200 {object} response.APIResponse[response.BrokerageResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages/{id} [get]
func (h *BrokerageHandler) GetBrokerageByID(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	// Parse and validate brokerage ID
	brokerageIDStr := c.Param("id")
	brokerageID, err := uuid.Parse(brokerageIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid brokerage ID format",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageIDStr),
		)

		errorResp := response.BadRequest("Invalid brokerage ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Get brokerage
	brokerageResp, err := h.brokerageService.GetBrokerageByID(ctx, brokerageID)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage retrieval failed",
				logger.String("request_id", requestID),
				logger.String("brokerage_id", brokerageID.String()),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage retrieval", err,
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
		)

		errorResp := response.InternalServerError("Failed to retrieve brokerage")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("brokerage_id", brokerageID.String()),
		logger.String("name", brokerageResp.Name),
	)

	apiResponse := response.Success(brokerageResp)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// UpdateBrokerage godoc
// @Summary Update brokerage
// @Description Update an existing brokerage with the provided details
// @Tags brokerages
// @Accept json
// @Produce json
// @Param id path string true "Brokerage ID"
// @Param brokerage body request.UpdateBrokerageRequest true "Brokerage update details"
// @Success 200 {object} response.APIResponse[response.BrokerageResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 422 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages/{id} [put]
func (h *BrokerageHandler) UpdateBrokerage(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	// Parse and validate brokerage ID
	brokerageIDStr := c.Param("id")
	brokerageID, err := uuid.Parse(brokerageIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid brokerage ID format",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageIDStr),
		)

		errorResp := response.BadRequest("Invalid brokerage ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	var req request.UpdateBrokerageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(ctx, "Invalid request body for brokerage update",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Invalid request body")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Warn(ctx, "Brokerage update validation failed",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Validation failed")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Update brokerage
	brokerageResp, err := h.brokerageService.UpdateBrokerage(ctx, brokerageID, &req)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage update failed",
				logger.String("request_id", requestID),
				logger.String("brokerage_id", brokerageID.String()),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage update", err,
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
		)

		errorResp := response.InternalServerError("Failed to update brokerage")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage updated successfully",
		logger.String("request_id", requestID),
		logger.String("brokerage_id", brokerageID.String()),
		logger.String("name", brokerageResp.Name),
	)

	apiResponse := response.Success(brokerageResp)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// DeleteBrokerage godoc
// @Summary Delete brokerage
// @Description Delete an existing brokerage (soft delete)
// @Tags brokerages
// @Accept json
// @Produce json
// @Param id path string true "Brokerage ID"
// @Success 200 {object} response.APIResponse[any]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages/{id} [delete]
func (h *BrokerageHandler) DeleteBrokerage(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	// Parse and validate brokerage ID
	brokerageIDStr := c.Param("id")
	brokerageID, err := uuid.Parse(brokerageIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid brokerage ID format",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageIDStr),
		)

		errorResp := response.BadRequest("Invalid brokerage ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Delete brokerage
	err = h.brokerageService.DeleteBrokerage(ctx, brokerageID)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage deletion failed",
				logger.String("request_id", requestID),
				logger.String("brokerage_id", brokerageID.String()),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage deletion", err,
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
		)

		errorResp := response.InternalServerError("Failed to delete brokerage")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage deleted successfully",
		logger.String("request_id", requestID),
		logger.String("brokerage_id", brokerageID.String()),
	)

	apiResponse := response.Success(map[string]string{"message": "Brokerage deleted successfully"})
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// ListBrokerages godoc
// @Summary List brokerages with filters
// @Description Get a paginated list of brokerages with optional filters
// @Tags brokerages
// @Accept json
// @Produce json
// @Param name query string false "Filter by name (partial match)"
// @Param is_active query boolean false "Filter by active status"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.BrokerageResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages [get]
func (h *BrokerageHandler) ListBrokerages(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id") // Parse pagination parameters
	pagination := h.parsePagination(c)

	// Parse filter parameters
	var filter request.BrokerageFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Warn(ctx, "Invalid filter parameters",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()),
		)

		errorResp := response.BadRequest("Invalid filter parameters")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Get brokerages
	paginatedBrokerages, err := h.brokerageService.ListBrokerages(ctx, &filter, pagination)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage listing failed",
				logger.String("request_id", requestID),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage listing", err,
			logger.String("request_id", requestID),
		)

		errorResp := response.InternalServerError("Failed to list brokerages")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerages listed successfully",
		logger.String("request_id", requestID),
		logger.Int("total", paginatedBrokerages.Meta.Total),
		logger.Int("page", paginatedBrokerages.Meta.Page),
		logger.Int("per_page", paginatedBrokerages.Meta.PerPage),
	)

	apiResponse := response.Success(paginatedBrokerages)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// ListActiveBrokerages godoc
// @Summary List active brokerages
// @Description Get a paginated list of active brokerages only
// @Tags brokerages
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.BrokerageResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages/active [get]
func (h *BrokerageHandler) ListActiveBrokerages(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	// Parse pagination parameters
	pagination := h.parsePagination(c)

	// Get active brokerages
	paginatedBrokerages, err := h.brokerageService.ListActiveBrokerages(ctx, pagination)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Active brokerage listing failed",
				logger.String("request_id", requestID),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during active brokerage listing", err,
			logger.String("request_id", requestID),
		)

		errorResp := response.InternalServerError("Failed to list active brokerages")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Active brokerages listed successfully",
		logger.String("request_id", requestID),
		logger.Int("total", paginatedBrokerages.Meta.Total),
		logger.Int("page", paginatedBrokerages.Meta.Page),
		logger.Int("per_page", paginatedBrokerages.Meta.PerPage),
	)

	apiResponse := response.Success(paginatedBrokerages)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// ActivateBrokerage godoc
// @Summary Activate brokerage
// @Description Mark a brokerage as active
// @Tags brokerages
// @Accept json
// @Produce json
// @Param id path string true "Brokerage ID"
// @Success 200 {object} response.APIResponse[any]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages/{id}/activate [patch]
func (h *BrokerageHandler) ActivateBrokerage(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	// Parse and validate brokerage ID
	brokerageIDStr := c.Param("id")
	brokerageID, err := uuid.Parse(brokerageIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid brokerage ID format",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageIDStr),
		)

		errorResp := response.BadRequest("Invalid brokerage ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Activate brokerage
	err = h.brokerageService.ActivateBrokerage(ctx, brokerageID)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage activation failed",
				logger.String("request_id", requestID),
				logger.String("brokerage_id", brokerageID.String()),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage activation", err,
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
		)

		errorResp := response.InternalServerError("Failed to activate brokerage")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage activated successfully",
		logger.String("request_id", requestID),
		logger.String("brokerage_id", brokerageID.String()),
	)

	apiResponse := response.Success(map[string]string{"message": "Brokerage activated successfully"})
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// DeactivateBrokerage godoc
// @Summary Deactivate brokerage
// @Description Mark a brokerage as inactive
// @Tags brokerages
// @Accept json
// @Produce json
// @Param id path string true "Brokerage ID"
// @Success 200 {object} response.APIResponse[any]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages/{id}/deactivate [patch]
func (h *BrokerageHandler) DeactivateBrokerage(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	// Parse and validate brokerage ID
	brokerageIDStr := c.Param("id")
	brokerageID, err := uuid.Parse(brokerageIDStr)
	if err != nil {
		h.logger.Warn(ctx, "Invalid brokerage ID format",
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageIDStr),
		)

		errorResp := response.BadRequest("Invalid brokerage ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Deactivate brokerage
	err = h.brokerageService.DeactivateBrokerage(ctx, brokerageID)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage deactivation failed",
				logger.String("request_id", requestID),
				logger.String("brokerage_id", brokerageID.String()),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage deactivation", err,
			logger.String("request_id", requestID),
			logger.String("brokerage_id", brokerageID.String()),
		)

		errorResp := response.InternalServerError("Failed to deactivate brokerage")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage deactivated successfully",
		logger.String("request_id", requestID),
		logger.String("brokerage_id", brokerageID.String()),
	)

	apiResponse := response.Success(map[string]string{"message": "Brokerage deactivated successfully"})
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// SearchBrokeragesByName godoc
// @Summary Search brokerages by name
// @Description Search brokerages by name with partial matching
// @Tags brokerages
// @Accept json
// @Produce json
// @Param name query string true "Name to search for"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.BrokerageResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/brokerages/search [get]
func (h *BrokerageHandler) SearchBrokeragesByName(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	// Parse and validate name parameter
	name := c.Query("name")
	if name == "" {
		h.logger.Warn(ctx, "Missing name parameter for brokerage search",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Name parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	// Parse pagination parameters
	pagination := h.parsePagination(c)

	// Search brokerages
	paginatedBrokerages, err := h.brokerageService.SearchBrokeragesByName(ctx, name, pagination)
	if err != nil {
		if errorResp, ok := err.(*response.ErrorResponse); ok {
			h.logger.Warn(ctx, "Brokerage search failed",
				logger.String("request_id", requestID),
				logger.String("name", name),
				logger.String("error", errorResp.Message),
			)

			apiResponse := errorResp.ToAPIResponse()
			apiResponse.RequestID = requestID

			c.JSON(errorResp.StatusCode, apiResponse)
			return
		}

		h.logger.Error(ctx, "Unexpected error during brokerage search", err,
			logger.String("request_id", requestID),
			logger.String("name", name),
		)

		errorResp := response.InternalServerError("Failed to search brokerages")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Brokerage search completed successfully",
		logger.String("request_id", requestID),
		logger.String("name", name),
		logger.Int("total", paginatedBrokerages.Meta.Total),
		logger.Int("page", paginatedBrokerages.Meta.Page),
		logger.Int("per_page", paginatedBrokerages.Meta.PerPage),
	)

	apiResponse := response.Success(paginatedBrokerages)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// parsePagination extrae y valida los parámetros de paginación
func (h *BrokerageHandler) parsePagination(c *gin.Context) *response.PaginationRequest {
	pageParam := c.Query("page")
	perPageParam := c.Query("per_page")
	
	return response.ParsePaginationFromQuery(pageParam, perPageParam)
}
