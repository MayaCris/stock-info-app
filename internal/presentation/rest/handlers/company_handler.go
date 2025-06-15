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

// CompanyHandler maneja los endpoints relacionados con companies
type CompanyHandler struct {
	companyService serviceInterfaces.CompanyService
	logger         logger.Logger
}

// NewCompanyHandler crea una nueva instancia del handler de companies
func NewCompanyHandler(companyService serviceInterfaces.CompanyService, appLogger logger.Logger) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
		logger:         appLogger,
	}
}

// CreateCompany godoc
// @Summary Create a new company
// @Description Create a new company with the provided details
// @Tags companies
// @Accept json
// @Produce json
// @Param company body request.CreateCompanyRequest true "Company creation details"
// @Success 201 {object} response.APIResponse[response.CompanyResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 409 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies [post]
func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	var req request.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(ctx, "Invalid request body for company creation",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Invalid request body")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Creating company",
		logger.String("request_id", requestID),
		logger.String("ticker", req.Ticker),
		logger.String("name", req.Name),
		logger.String("sector", req.Sector),
	)

	company, err := h.companyService.CreateCompany(ctx, &req)
	if err != nil {
		h.logger.Error(ctx, "Failed to create company",
			err,
			logger.String("request_id", requestID),
			logger.String("ticker", req.Ticker),
		)

		errorResp := response.InternalServerError("Failed to create company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company created successfully",
		logger.String("request_id", requestID),
		logger.String("company_id", company.ID.String()),
		logger.String("ticker", company.Ticker),
	)

	apiResponse := response.Success(company)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusCreated, apiResponse)
}

// GetCompanyByID godoc
// @Summary Get company by ID
// @Description Get a specific company by its ID
// @Tags companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} response.APIResponse[response.CompanyResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/{id} [get]
func (h *CompanyHandler) GetCompanyByID(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	idParam := c.Param("id")
	companyID, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("id", idParam),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Getting company by ID",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	company, err := h.companyService.GetCompanyByID(ctx, companyID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get company by ID",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.NotFound("Company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(company)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetCompanyByTicker godoc
// @Summary Get company by ticker
// @Description Get a specific company by its ticker symbol
// @Tags companies
// @Accept json
// @Produce json
// @Param ticker path string true "Company ticker symbol"
// @Success 200 {object} response.APIResponse[response.CompanyResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/ticker/{ticker} [get]
func (h *CompanyHandler) GetCompanyByTicker(c *gin.Context) {
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

	h.logger.Info(ctx, "Getting company by ticker",
		logger.String("request_id", requestID),
		logger.String("ticker", ticker),
	)

	company, err := h.companyService.GetCompanyByTicker(ctx, ticker)
	if err != nil {
		h.logger.Error(ctx, "Failed to get company by ticker",
			err,
			logger.String("request_id", requestID),
			logger.String("ticker", ticker),
		)

		errorResp := response.NotFound("Company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(company)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// UpdateCompany godoc
// @Summary Update a company
// @Description Update an existing company with the provided details
// @Tags companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param company body request.UpdateCompanyRequest true "Company update details"
// @Success 200 {object} response.APIResponse[response.CompanyResponse]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/{id} [put]
func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	idParam := c.Param("id")
	companyID, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("id", idParam),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	var req request.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(ctx, "Invalid request body for company update",
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Invalid request body")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Updating company",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	company, err := h.companyService.UpdateCompany(ctx, companyID, &req)
	if err != nil {
		h.logger.Error(ctx, "Failed to update company",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to update company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company updated successfully",
		logger.String("request_id", requestID),
		logger.String("company_id", company.ID.String()),
	)

	apiResponse := response.Success(company)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// DeleteCompany godoc
// @Summary Delete a company
// @Description Delete an existing company by ID
// @Tags companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/{id} [delete]
func (h *CompanyHandler) DeleteCompany(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	idParam := c.Param("id")
	companyID, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("id", idParam),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Deleting company",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	err = h.companyService.DeleteCompany(ctx, companyID)
	if err != nil {
		h.logger.Error(ctx, "Failed to delete company",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to delete company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company deleted successfully",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	c.Status(http.StatusNoContent)
}

// ListCompanies godoc
// @Summary List companies with filtering and pagination
// @Description Get a paginated list of companies with optional filters
// @Tags companies
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param ticker query string false "Filter by ticker"
// @Param name query string false "Filter by name (partial match)"
// @Param sector query string false "Filter by sector"
// @Param exchange query string false "Filter by exchange"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.CompanyListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies [get]
func (h *CompanyHandler) ListCompanies(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// Parse pagination
	pagination := h.parsePagination(c)

	// Parse filters
	var filter request.CompanyFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Warn(ctx, "Invalid query parameters for company listing",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()),
		)

		errorResp := response.BadRequest("Invalid query parameters")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Listing companies",
		logger.String("request_id", requestID),
		logger.Int("page", pagination.Page),
		logger.Int("per_page", pagination.PerPage),
		logger.String("ticker", filter.Ticker),
		logger.String("sector", filter.Sector),
	)

	companies, err := h.companyService.ListCompanies(ctx, &filter, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to list companies",
			err,
			logger.String("request_id", requestID),
		)

		errorResp := response.InternalServerError("Failed to list companies")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(companies)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// ListActiveCompanies godoc
// @Summary List active companies
// @Description Get a paginated list of active companies only
// @Tags companies
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.CompanyListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/active [get]
func (h *CompanyHandler) ListActiveCompanies(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Listing active companies",
		logger.String("request_id", requestID),
		logger.Int("page", pagination.Page),
		logger.Int("per_page", pagination.PerPage),
	)

	companies, err := h.companyService.ListActiveCompanies(ctx, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to list active companies",
			err,
			logger.String("request_id", requestID),
		)

		errorResp := response.InternalServerError("Failed to list active companies")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(companies)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// ActivateCompany godoc
// @Summary Activate a company
// @Description Activate an inactive company
// @Tags companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} response.APIResponse[any]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/{id}/activate [post]
func (h *CompanyHandler) ActivateCompany(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	idParam := c.Param("id")
	companyID, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("id", idParam),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Activating company",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	err = h.companyService.ActivateCompany(ctx, companyID)
	if err != nil {
		h.logger.Error(ctx, "Failed to activate company",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to activate company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company activated successfully",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	apiResponse := response.Success(map[string]string{"message": "Company activated successfully"})
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// DeactivateCompany godoc
// @Summary Deactivate a company
// @Description Deactivate an active company
// @Tags companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} response.APIResponse[any]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/{id}/deactivate [post]
func (h *CompanyHandler) DeactivateCompany(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	idParam := c.Param("id")
	companyID, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn(ctx, "Invalid company ID format",
			logger.String("request_id", requestID),
			logger.String("id", idParam),
		)

		errorResp := response.BadRequest("Invalid company ID format")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Deactivating company",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	err = h.companyService.DeactivateCompany(ctx, companyID)
	if err != nil {
		h.logger.Error(ctx, "Failed to deactivate company",
			err,
			logger.String("request_id", requestID),
			logger.String("company_id", companyID.String()),
		)

		errorResp := response.InternalServerError("Failed to deactivate company")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Company deactivated successfully",
		logger.String("request_id", requestID),
		logger.String("company_id", companyID.String()),
	)

	apiResponse := response.Success(map[string]string{"message": "Company deactivated successfully"})
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// SearchCompaniesByName godoc
// @Summary Search companies by name
// @Description Search companies by name with partial matching
// @Tags companies
// @Accept json
// @Produce json
// @Param name query string true "Company name to search"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.CompanyListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/search [get]
func (h *CompanyHandler) SearchCompaniesByName(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	name := c.Query("name")
	if name == "" {
		h.logger.Warn(ctx, "Empty name parameter for company search",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Name parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Searching companies by name",
		logger.String("request_id", requestID),
		logger.String("name", name),
		logger.Int("page", pagination.Page),
		logger.Int("per_page", pagination.PerPage),
	)

	companies, err := h.companyService.SearchCompaniesByName(ctx, name, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to search companies by name",
			err,
			logger.String("request_id", requestID),
			logger.String("name", name),
		)

		errorResp := response.InternalServerError("Failed to search companies")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(companies)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// GetCompaniesBySector godoc
// @Summary Get companies by sector
// @Description Get all companies in a specific sector
// @Tags companies
// @Accept json
// @Produce json
// @Param sector path string true "Sector name"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.APIResponse[response.PaginatedResponse[response.CompanyListResponse]]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/sector/{sector} [get]
func (h *CompanyHandler) GetCompaniesBySector(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	sector := c.Param("sector")
	if sector == "" {
		h.logger.Warn(ctx, "Empty sector parameter",
			logger.String("request_id", requestID),
		)

		errorResp := response.BadRequest("Sector parameter is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	pagination := h.parsePagination(c)

	h.logger.Info(ctx, "Getting companies by sector",
		logger.String("request_id", requestID),
		logger.String("sector", sector),
		logger.Int("page", pagination.Page),
		logger.Int("per_page", pagination.PerPage),
	)

	companies, err := h.companyService.GetCompaniesBySector(ctx, sector, pagination)
	if err != nil {
		h.logger.Error(ctx, "Failed to get companies by sector",
			err,
			logger.String("request_id", requestID),
			logger.String("sector", sector),
		)

		errorResp := response.InternalServerError("Failed to get companies by sector")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	apiResponse := response.Success(companies)
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// UpdateMarketCap godoc
// @Summary Update company market cap
// @Description Update the market capitalization of a company by ticker
// @Tags companies
// @Accept json
// @Produce json
// @Param ticker path string true "Company ticker symbol"
// @Param request body map[string]float64 true "Market cap update request"
// @Success 200 {object} response.APIResponse[any]
// @Failure 400 {object} response.APIResponse[any]
// @Failure 404 {object} response.APIResponse[any]
// @Failure 500 {object} response.APIResponse[any]
// @Router /api/v1/companies/ticker/{ticker}/market-cap [put]
func (h *CompanyHandler) UpdateMarketCap(c *gin.Context) {
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

	var req map[string]float64
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(ctx, "Invalid request body for market cap update",
			logger.String("request_id", requestID),
			logger.String("ticker", ticker),
			logger.String("error", err.Error()),
		)

		errorResp := response.ValidationFailed("Invalid request body")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	marketCap, exists := req["market_cap"]
	if !exists || marketCap < 0 {
		h.logger.Warn(ctx, "Invalid market cap value",
			logger.String("request_id", requestID),
			logger.String("ticker", ticker),
		)

		errorResp := response.BadRequest("Valid market_cap field is required")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Updating company market cap",
		logger.String("request_id", requestID),
		logger.String("ticker", ticker),
		logger.Float64("market_cap", marketCap),
	)

	err := h.companyService.UpdateMarketCap(ctx, ticker, marketCap)
	if err != nil {
		h.logger.Error(ctx, "Failed to update market cap",
			err,
			logger.String("request_id", requestID),
			logger.String("ticker", ticker),
		)

		errorResp := response.InternalServerError("Failed to update market cap")
		apiResponse := errorResp.ToAPIResponse()
		apiResponse.RequestID = requestID

		c.JSON(errorResp.StatusCode, apiResponse)
		return
	}

	h.logger.Info(ctx, "Market cap updated successfully",
		logger.String("request_id", requestID),
		logger.String("ticker", ticker),
	)

	apiResponse := response.Success(map[string]string{"message": "Market cap updated successfully"})
	apiResponse.RequestID = requestID

	c.JSON(http.StatusOK, apiResponse)
}

// parsePagination extrae y valida los parámetros de paginación
func (h *CompanyHandler) parsePagination(c *gin.Context) *response.PaginationRequest {
	pageParam := c.DefaultQuery("page", "1")
	perPageParam := c.DefaultQuery("per_page", "20")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageParam)
	if err != nil || perPage < 1 {
		perPage = 20
	}

	// Limit max per_page to prevent abuse
	if perPage > 100 {
		perPage = 100
	}

	return &response.PaginationRequest{
		Page:    page,
		PerPage: perPage,
	}
}
