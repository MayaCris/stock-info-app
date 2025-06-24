package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/application/dto/request"
	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// stockRatingService implements the StockRatingService interface
type stockRatingService struct {
	stockRatingRepo repoInterfaces.StockRatingRepository
	companyRepo     repoInterfaces.CompanyRepository
	brokerageRepo   repoInterfaces.BrokerageRepository
	logger          logger.Logger
}

// NewStockRatingService creates a new stock rating service
func NewStockRatingService(
	stockRatingRepo repoInterfaces.StockRatingRepository,
	companyRepo repoInterfaces.CompanyRepository,
	brokerageRepo repoInterfaces.BrokerageRepository,
	logger logger.Logger,
) interfaces.StockRatingService {
	return &stockRatingService{
		stockRatingRepo: stockRatingRepo,
		companyRepo:     companyRepo,
		brokerageRepo:   brokerageRepo,
		logger:          logger,
	}
}

// CreateStockRating creates a new stock rating
func (s *stockRatingService) CreateStockRating(ctx context.Context, req *request.CreateStockRatingRequest) (*response.StockRatingResponse, error) {
	// Validate that company exists
	company, err := s.companyRepo.GetByID(ctx, req.CompanyID)
	if err != nil {
		s.logger.Error(ctx, "Failed to find company for stock rating", err,
			logger.String("company_id", req.CompanyID.String()))
		return nil, response.NotFound("Company")
	}

	// Validate that brokerage exists
	brokerage, err := s.brokerageRepo.GetByID(ctx, req.BrokerageID)
	if err != nil {
		s.logger.Error(ctx, "Failed to find brokerage for stock rating", err,
			logger.String("brokerage_id", req.BrokerageID.String()))
		return nil, response.NotFound("Brokerage")
	}

	// Create stock rating entity
	stockRating := &entities.StockRating{
		ID:          uuid.New(),
		CompanyID:   req.CompanyID,
		BrokerageID: req.BrokerageID,
		Action:      req.Action,
		RatingFrom:  req.RatingFrom,
		RatingTo:    req.RatingTo,
		TargetFrom:  req.TargetFrom,
		TargetTo:    req.TargetTo,
	}

	// Save to repository
	if err := s.stockRatingRepo.Create(ctx, stockRating); err != nil {
		s.logger.Error(ctx, "Failed to create stock rating", err,
			logger.String("company_id", req.CompanyID.String()),
			logger.String("brokerage_id", req.BrokerageID.String()))
		return nil, response.InternalServerError("Failed to create stock rating")
	}

	s.logger.Info(ctx, "Stock rating created successfully",
		logger.String("stock_rating_id", stockRating.ID.String()),
		logger.String("company_ticker", company.Ticker),
		logger.String("brokerage_name", brokerage.Name))

	// Convert to response
	return s.convertToStockRatingResponse(stockRating, company, brokerage), nil
}

// GetStockRatingByID retrieves a stock rating by ID
func (s *stockRatingService) GetStockRatingByID(ctx context.Context, id uuid.UUID) (*response.StockRatingResponse, error) {
	stockRating, err := s.stockRatingRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get stock rating by ID", err,
			logger.String("stock_rating_id", id.String()))
		return nil, response.NotFound("Stock rating")
	}

	// Get related entities
	company, _ := s.companyRepo.GetByID(ctx, stockRating.CompanyID)
	brokerage, _ := s.brokerageRepo.GetByID(ctx, stockRating.BrokerageID)

	return s.convertToStockRatingResponse(stockRating, company, brokerage), nil
}

// DeleteStockRating deletes a stock rating
func (s *stockRatingService) DeleteStockRating(ctx context.Context, id uuid.UUID) error {
	// Check if exists
	_, err := s.stockRatingRepo.GetByID(ctx, id)
	if err != nil {
		return response.NotFound("Stock rating")
	}

	if err := s.stockRatingRepo.Delete(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to delete stock rating", err,
			logger.String("stock_rating_id", id.String()))
		return response.InternalServerError("Failed to delete stock rating")
	}

	s.logger.Info(ctx, "Stock rating deleted successfully",
		logger.String("stock_rating_id", id.String()))
	return nil
}

// ListStockRatings lists stock ratings with filters and pagination
func (s *stockRatingService) ListStockRatings(ctx context.Context, filter *request.StockRatingFilterRequest, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	// Get total count for pagination
	total, err := s.stockRatingRepo.Count(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to count stock ratings", err)
		return nil, response.InternalServerError("Failed to count stock ratings")
	}
	// Get stock ratings using GetAll with pagination logic
	allRatings, err := s.stockRatingRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get stock ratings", err)
		return nil, response.InternalServerError("Failed to get stock ratings")
	}

	// Apply pagination manually (in production, implement GetWithPagination in repository)
	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if start > len(allRatings) {
		start = len(allRatings)
	}
	if end > len(allRatings) {
		end = len(allRatings)
	}
	stockRatings := allRatings[start:end]

	// Convert to list responses
	listResponses := make([]*response.StockRatingListResponse, len(stockRatings))
	for i, rating := range stockRatings {
		// Get company and brokerage info
		company, _ := s.companyRepo.GetByID(ctx, rating.CompanyID)
		brokerage, _ := s.brokerageRepo.GetByID(ctx, rating.BrokerageID)
		listResponses[i] = s.convertToStockRatingListResponse(rating, company, brokerage)
	}

	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, int(total)), nil
}

// GetRatingsByCompany gets ratings for a specific company
func (s *stockRatingService) GetRatingsByCompany(ctx context.Context, companyID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	// Check if company exists
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, response.NotFound("Company")
	}

	// Get ratings by company
	stockRatings, err := s.stockRatingRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get stock ratings by company", err,
			logger.String("company_id", companyID.String()))
		return nil, response.InternalServerError("Failed to get stock ratings")
	}

	// Convert to list responses
	listResponses := make([]*response.StockRatingListResponse, len(stockRatings))
	for i, rating := range stockRatings {
		brokerage, _ := s.brokerageRepo.GetByID(ctx, rating.BrokerageID)
		listResponses[i] = s.convertToStockRatingListResponse(rating, company, brokerage)
	}

	// For simplicity, we'll return all results (in production, implement proper pagination in repository)
	total := len(listResponses)
	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, total), nil
}

// GetRatingsByTicker gets ratings for a company by ticker
func (s *stockRatingService) GetRatingsByTicker(ctx context.Context, ticker string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error) {
	// Get company by ticker
	company, err := s.companyRepo.GetByTicker(ctx, ticker)
	if err != nil {
		return nil, response.NotFound("Company with ticker " + ticker)
	}

	return s.GetRatingsByCompany(ctx, company.ID, pagination)
}

// GetRatingsByBrokerage gets ratings by brokerage
func (s *stockRatingService) GetRatingsByBrokerage(ctx context.Context, brokerageID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error) {
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		return nil, response.BadRequest("Invalid pagination parameters")
	}

	// Check if brokerage exists
	brokerage, err := s.brokerageRepo.GetByID(ctx, brokerageID)
	if err != nil {
		return nil, response.NotFound("Brokerage")
	}

	// Get ratings by brokerage
	stockRatings, err := s.stockRatingRepo.GetByBrokerageID(ctx, brokerageID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get stock ratings by brokerage", err,
			logger.String("brokerage_id", brokerageID.String()))
		return nil, response.InternalServerError("Failed to get stock ratings")
	}

	// Convert to list responses
	listResponses := make([]*response.StockRatingListResponse, len(stockRatings))
	for i, rating := range stockRatings {
		company, _ := s.companyRepo.GetByID(ctx, rating.CompanyID)
		listResponses[i] = s.convertToStockRatingListResponse(rating, company, brokerage)
	}

	total := len(listResponses)
	return response.NewPaginatedResponse(listResponses, pagination.Page, pagination.PerPage, total), nil
}

// GetRecentRatings gets recent stock ratings
func (s *stockRatingService) GetRecentRatings(ctx context.Context, limit int) ([]*response.StockRatingListResponse, error) {
	stockRatings, err := s.stockRatingRepo.GetRecent(ctx, 30, limit) // Last 30 days
	if err != nil {
		s.logger.Error(ctx, "Failed to get recent stock ratings", err)
		return nil, response.InternalServerError("Failed to get recent ratings")
	}

	// Convert to list responses
	listResponses := make([]*response.StockRatingListResponse, len(stockRatings))
	for i, rating := range stockRatings {
		company, _ := s.companyRepo.GetByID(ctx, rating.CompanyID)
		brokerage, _ := s.brokerageRepo.GetByID(ctx, rating.BrokerageID)
		listResponses[i] = s.convertToStockRatingListResponse(rating, company, brokerage)
	}

	return listResponses, nil
}

// GetRatingsByDateRange gets ratings within a date range
func (s *stockRatingService) GetRatingsByDateRange(ctx context.Context, startDate, endDate string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error) {
	// Implementation would parse dates and query repository
	// For now, return empty result
	return response.NewPaginatedResponse([]*response.StockRatingListResponse{}, pagination.Page, pagination.PerPage, 0), nil
}

// GetRatingStatsByCompany gets statistics for a company's ratings
func (s *stockRatingService) GetRatingStatsByCompany(ctx context.Context, companyID uuid.UUID) (map[string]interface{}, error) {
	// Check if company exists
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, response.NotFound("Company")
	}

	// Get all ratings for the company
	ratings, err := s.stockRatingRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get company ratings")
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"company_id":    companyID,
		"company_name":  company.Name,
		"ticker":        company.Ticker,
		"total_ratings": len(ratings),
		"rating_breakdown": map[string]int{
			"buy":   0,
			"hold":  0,
			"sell":  0,
			"other": 0,
		},
	}

	// Count ratings by type
	ratingBreakdown := stats["rating_breakdown"].(map[string]int)
	for _, rating := range ratings {
		switch rating.RatingTo {
		case "Buy", "Strong Buy", "Outperform":
			ratingBreakdown["buy"]++
		case "Hold", "Neutral":
			ratingBreakdown["hold"]++
		case "Sell", "Strong Sell", "Underperform":
			ratingBreakdown["sell"]++
		default:
			ratingBreakdown["other"]++
		}
	}

	return stats, nil
}

// Helper methods

func (s *stockRatingService) convertToStockRatingResponse(rating *entities.StockRating, company *entities.Company, brokerage *entities.Brokerage) *response.StockRatingResponse {
	resp := &response.StockRatingResponse{
		ID:          rating.ID,
		CompanyID:   rating.CompanyID,
		BrokerageID: rating.BrokerageID,
		Action:      rating.Action,
		RatingFrom:  rating.RatingFrom,
		RatingTo:    rating.RatingTo,
		TargetFrom:  rating.TargetFrom,
		TargetTo:    rating.TargetTo,
		EventTime:   rating.EventTime,
		CreatedAt:   rating.CreatedAt,
		UpdatedAt:   rating.UpdatedAt,
	}

	if company != nil {
		resp.Company = &response.CompanyResponse{
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
	if brokerage != nil {
		resp.Brokerage = &response.BrokerageResponse{
			ID:        brokerage.ID,
			Name:      brokerage.Name,
			Website:   brokerage.Website,
			IsActive:  brokerage.IsActive,
			CreatedAt: brokerage.CreatedAt,
			UpdatedAt: brokerage.UpdatedAt,
		}
	}

	return resp
}

func (s *stockRatingService) convertToStockRatingListResponse(rating *entities.StockRating, company *entities.Company, brokerage *entities.Brokerage) *response.StockRatingListResponse {
	resp := &response.StockRatingListResponse{
		ID:        rating.ID,
		CompanyID: rating.CompanyID,
		Action:    rating.Action,
		RatingTo:  rating.RatingTo,
		TargetTo:  rating.TargetTo,
		EventTime: rating.EventTime,
	}

	if company != nil {
		resp.Ticker = company.Ticker
		resp.Company = company.Name
	}

	if brokerage != nil {
		resp.Brokerage = brokerage.Name
	}

	return resp
}
