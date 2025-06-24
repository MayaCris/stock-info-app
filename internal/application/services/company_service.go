package services

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/application/dto/request"
	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// companyService implements the CompanyService interface
type companyService struct {
	companyRepo repoInterfaces.CompanyRepository
	logger      logger.Logger
}

// NewCompanyService creates a new company service
func NewCompanyService(
	companyRepo repoInterfaces.CompanyRepository,
	logger logger.Logger,
) interfaces.CompanyService {
	return &companyService{
		companyRepo: companyRepo,
		logger:      logger,
	}
}

// CreateCompany creates a new company
func (s *companyService) CreateCompany(ctx context.Context, req *request.CreateCompanyRequest) (*response.CompanyResponse, error) {
	// Check if company already exists by ticker
	exists, err := s.companyRepo.ExistsByTicker(ctx, req.Ticker)
	if err != nil {
		s.logger.Error(ctx, "Failed to check company existence", err,
			logger.String("ticker", req.Ticker))
		return nil, response.InternalServerError("Failed to check company existence")
	}

	if exists {
		return nil, response.Conflict("Company with ticker already exists")
	}

	// Create company entity
	company := &entities.Company{
		ID:        uuid.New(),
		Ticker:    strings.ToUpper(req.Ticker),
		Name:      req.Name,
		Sector:    req.Sector,
		Exchange:  req.Exchange,
		MarketCap: req.MarketCap,
		Logo:      req.Logo,
		IsActive:  true,
	}

	// Save to repository
	if err := s.companyRepo.Create(ctx, company); err != nil {
		s.logger.Error(ctx, "Failed to create company", err,
			logger.String("ticker", req.Ticker),
			logger.String("name", req.Name))
		return nil, response.InternalServerError("Failed to create company")
	}

	s.logger.Info(ctx, "Company created successfully",
		logger.String("company_id", company.ID.String()),
		logger.String("ticker", company.Ticker),
		logger.String("name", company.Name))

	return s.convertToCompanyResponse(company), nil
}

// GetCompanyByID retrieves a company by ID
func (s *companyService) GetCompanyByID(ctx context.Context, id uuid.UUID) (*response.CompanyResponse, error) {
	company, err := s.companyRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get company by ID", err,
			logger.String("company_id", id.String()))
		return nil, response.NotFound("Company")
	}

	return s.convertToCompanyResponse(company), nil
}

// GetCompanyByTicker retrieves a company by ticker
func (s *companyService) GetCompanyByTicker(ctx context.Context, ticker string) (*response.CompanyResponse, error) {
	company, err := s.companyRepo.GetByTicker(ctx, strings.ToUpper(ticker))
	if err != nil {
		s.logger.Error(ctx, "Failed to get company by ticker", err,
			logger.String("ticker", ticker))
		return nil, response.NotFound("Company")
	}

	return s.convertToCompanyResponse(company), nil
}

// UpdateCompany updates an existing company
func (s *companyService) UpdateCompany(ctx context.Context, id uuid.UUID, req *request.UpdateCompanyRequest) (*response.CompanyResponse, error) {
	// Get existing company
	company, err := s.companyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, response.NotFound("Company")
	}

	// Update fields if provided
	if req.Name != nil {
		company.Name = *req.Name
	}
	if req.Sector != nil {
		company.Sector = *req.Sector
	}
	if req.Exchange != nil {
		company.Exchange = *req.Exchange
	}
	if req.MarketCap != nil {
		company.MarketCap = *req.MarketCap
	}
	if req.Logo != nil {
		company.Logo = *req.Logo
	}
	if req.IsActive != nil {
		company.IsActive = *req.IsActive
	}

	// Save changes
	if err := s.companyRepo.Update(ctx, company); err != nil {
		s.logger.Error(ctx, "Failed to update company", err,
			logger.String("company_id", id.String()))
		return nil, response.InternalServerError("Failed to update company")
	}

	s.logger.Info(ctx, "Company updated successfully",
		logger.String("company_id", company.ID.String()),
		logger.String("ticker", company.Ticker))

	return s.convertToCompanyResponse(company), nil
}

// DeleteCompany deletes a company
func (s *companyService) DeleteCompany(ctx context.Context, id uuid.UUID) error {
	// Check if exists
	_, err := s.companyRepo.GetByID(ctx, id)
	if err != nil {
		return response.NotFound("Company")
	}

	if err := s.companyRepo.Delete(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to delete company", err,
			logger.String("company_id", id.String()))
		return response.InternalServerError("Failed to delete company")
	}

	s.logger.Info(ctx, "Company deleted successfully",
		logger.String("company_id", id.String()))
	return nil
}

// ListCompanies lists companies with filters and pagination
func (s *companyService) ListCompanies(ctx context.Context, filter *request.CompanyFilterRequest, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	var companies []*entities.Company
	var total int64
	var err error
	// Apply filters
	if filter != nil {
		if filter.Sector != "" {
			companies, err = s.companyRepo.GetBySector(ctx, filter.Sector)
		} else if filter.Exchange != "" {
			companies, err = s.companyRepo.GetByExchange(ctx, filter.Exchange)
		} else if filter.IsActive != nil && *filter.IsActive {
			companies, err = s.companyRepo.GetAllActive(ctx)
		} else {
			companies, err = s.companyRepo.GetAll(ctx)
		}
	} else {
		companies, err = s.companyRepo.GetAll(ctx)
	}

	if err != nil {
		s.logger.Error(ctx, "Failed to get companies", err)
		return nil, response.InternalServerError("Failed to get companies")
	}

	total = int64(len(companies))

	// Apply pagination manually (in production, implement pagination in repository)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > len(companies) {
		start = len(companies)
	}
	if end > len(companies) {
		end = len(companies)
	}
	paginatedCompanies := companies[start:end]

	// Convert to list responses
	listResponses := make([]*response.CompanyListResponse, len(paginatedCompanies))
	for i, company := range paginatedCompanies {
		listResponses[i] = s.convertToCompanyListResponse(company)
	}

	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, int(total)), nil
}

// GetCompaniesBySector gets companies by sector
func (s *companyService) GetCompaniesBySector(ctx context.Context, sector string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	companies, err := s.companyRepo.GetBySector(ctx, sector)
	if err != nil {
		s.logger.Error(ctx, "Failed to get companies by sector", err,
			logger.String("sector", sector))
		return nil, response.InternalServerError("Failed to get companies")
	}

	// Apply pagination
	total := len(companies)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginatedCompanies := companies[start:end]

	// Convert to list responses
	listResponses := make([]*response.CompanyListResponse, len(paginatedCompanies))
	for i, company := range paginatedCompanies {
		listResponses[i] = s.convertToCompanyListResponse(company)
	}

	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, total), nil
}

// GetCompaniesByExchange gets companies by exchange
func (s *companyService) GetCompaniesByExchange(ctx context.Context, exchange string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	companies, err := s.companyRepo.GetByExchange(ctx, exchange)
	if err != nil {
		s.logger.Error(ctx, "Failed to get companies by exchange", err,
			logger.String("exchange", exchange))
		return nil, response.InternalServerError("Failed to get companies")
	}

	// Apply pagination
	total := len(companies)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginatedCompanies := companies[start:end]

	// Convert to list responses
	listResponses := make([]*response.CompanyListResponse, len(paginatedCompanies))
	for i, company := range paginatedCompanies {
		listResponses[i] = s.convertToCompanyListResponse(company)
	}

	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, total), nil
}

// GetLargestCompaniesByMarketCap gets largest companies by market cap
func (s *companyService) GetLargestCompaniesByMarketCap(ctx context.Context, limit int) ([]*response.CompanyListResponse, error) {
	companies, err := s.companyRepo.GetLargestByMarketCap(ctx, limit)
	if err != nil {
		s.logger.Error(ctx, "Failed to get largest companies by market cap", err)
		return nil, response.InternalServerError("Failed to get companies")
	}

	// Convert to list responses
	listResponses := make([]*response.CompanyListResponse, len(companies))
	for i, company := range companies {
		listResponses[i] = s.convertToCompanyListResponse(company)
	}

	return listResponses, nil
}

// SearchCompanies searches companies by name or ticker
func (s *companyService) SearchCompanies(ctx context.Context, query string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	// For now, get all companies and filter manually
	// In production, implement search in repository
	companies, err := s.companyRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to search companies", err)
		return nil, response.InternalServerError("Failed to search companies")
	}

	// Simple text search
	query = strings.ToLower(query)
	var filteredCompanies []*entities.Company
	for _, company := range companies {
		if strings.Contains(strings.ToLower(company.Name), query) ||
			strings.Contains(strings.ToLower(company.Ticker), query) {
			filteredCompanies = append(filteredCompanies, company)
		}
	}

	// Apply pagination
	total := len(filteredCompanies)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginatedCompanies := filteredCompanies[start:end]

	// Convert to list responses
	listResponses := make([]*response.CompanyListResponse, len(paginatedCompanies))
	for i, company := range paginatedCompanies {
		listResponses[i] = s.convertToCompanyListResponse(company)
	}

	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, total), nil
}

// GetCompanyStats gets basic statistics for companies
func (s *companyService) GetCompanyStats(ctx context.Context) (map[string]interface{}, error) {
	totalCount, err := s.companyRepo.Count(ctx)
	if err != nil {
		return nil, response.InternalServerError("Failed to get company stats")
	}

	activeCount, err := s.companyRepo.CountActive(ctx)
	if err != nil {
		return nil, response.InternalServerError("Failed to get company stats")
	}

	stats := map[string]interface{}{
		"total_companies":    totalCount,
		"active_companies":   activeCount,
		"inactive_companies": totalCount - activeCount,
	}

	return stats, nil
}

// ActivateCompany activates a company
func (s *companyService) ActivateCompany(ctx context.Context, id uuid.UUID) error {
	if err := s.companyRepo.Activate(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to activate company", err,
			logger.String("company_id", id.String()))
		return response.InternalServerError("Failed to activate company")
	}

	s.logger.Info(ctx, "Company activated successfully",
		logger.String("company_id", id.String()))
	return nil
}

// DeactivateCompany deactivates a company
func (s *companyService) DeactivateCompany(ctx context.Context, id uuid.UUID) error {
	if err := s.companyRepo.Deactivate(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to deactivate company", err,
			logger.String("company_id", id.String()))
		return response.InternalServerError("Failed to deactivate company")
	}

	s.logger.Info(ctx, "Company deactivated successfully",
		logger.String("company_id", id.String()))
	return nil
}

// UpdateMarketCap updates a company's market cap
func (s *companyService) UpdateMarketCap(ctx context.Context, ticker string, marketCap float64) error {
	if err := s.companyRepo.UpdateMarketCap(ctx, strings.ToUpper(ticker), marketCap); err != nil {
		s.logger.Error(ctx, "Failed to update market cap", err,
			logger.String("ticker", ticker),
			logger.Float64("market_cap", marketCap))
		return response.InternalServerError("Failed to update market cap")
	}

	s.logger.Info(ctx, "Market cap updated successfully",
		logger.String("ticker", ticker),
		logger.Float64("market_cap", marketCap))
	return nil
}

// ListActiveCompanies lists only active companies
func (s *companyService) ListActiveCompanies(ctx context.Context, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	companies, err := s.companyRepo.GetAllActive(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get active companies", err)
		return nil, response.InternalServerError("Failed to get companies")
	}

	total := len(companies)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginatedCompanies := companies[start:end]

	// Convert to list responses
	listResponses := make([]*response.CompanyListResponse, len(paginatedCompanies))
	for i, company := range paginatedCompanies {
		listResponses[i] = s.convertToCompanyListResponse(company)
	}

	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, total), nil
}

// SearchCompaniesByName searches companies by name
func (s *companyService) SearchCompaniesByName(ctx context.Context, name string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error) {
	// For now, use the general search function
	return s.SearchCompanies(ctx, name, pagination)
}

// Helper methods

func (s *companyService) convertToCompanyResponse(company *entities.Company) *response.CompanyResponse {
	return &response.CompanyResponse{
		ID:        company.ID,
		Ticker:    company.Ticker,
		Name:      company.Name,
		Sector:    company.Sector,
		MarketCap: company.MarketCap,
		Exchange:  company.Exchange,
		Logo:      company.Logo,
		IsActive:  company.IsActive,
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
	}
}

func (s *companyService) convertToCompanyListResponse(company *entities.Company) *response.CompanyListResponse {
	return &response.CompanyListResponse{
		ID:       company.ID,
		Ticker:   company.Ticker,
		Name:     company.Name,
		Sector:   company.Sector,
		Exchange: company.Exchange,
		Logo:     company.Logo,
		IsActive: company.IsActive,
	}
}
