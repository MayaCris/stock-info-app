package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// analysisService implements the AnalysisService interface
type analysisService struct {
	stockRatingRepo repoInterfaces.StockRatingRepository
	companyRepo     repoInterfaces.CompanyRepository
	brokerageRepo   repoInterfaces.BrokerageRepository
	logger          logger.Logger
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(
	stockRatingRepo repoInterfaces.StockRatingRepository,
	companyRepo repoInterfaces.CompanyRepository,
	brokerageRepo repoInterfaces.BrokerageRepository,
	logger logger.Logger,
) interfaces.AnalysisService {
	return &analysisService{
		stockRatingRepo: stockRatingRepo,
		companyRepo:     companyRepo,
		brokerageRepo:   brokerageRepo,
		logger:          logger,
	}
}

// GetCompanyAnalysis provides detailed analysis for a specific company
func (s *analysisService) GetCompanyAnalysis(ctx context.Context, companyID uuid.UUID) (*response.AnalysisResponse, error) {
	// Get company details
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, response.NotFound("Company")
	}

	// Get company ratings
	ratings, err := s.stockRatingRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get company ratings", err)
		return nil, response.InternalServerError("Failed to get company analysis")
	}

	// Calculate rating statistics
	ratingStats := s.calculateCompanyRatingStats(ratings)

	// Get recent ratings for the response
	recentRatingResponses := make([]response.StockRatingListResponse, 0)
	recentLimit := 10
	if len(ratings) > 0 {
		limit := recentLimit
		if len(ratings) < limit {
			limit = len(ratings)
		}
		for i := len(ratings) - limit; i < len(ratings); i++ {
			rating := ratings[i]
			recentRatingResponses = append(recentRatingResponses, response.StockRatingListResponse{
				ID:        rating.ID,
				CompanyID: rating.CompanyID,
				Ticker:    company.Ticker,
				Company:   company.Name,
				Action:    rating.Action,
				RatingTo:  rating.RatingTo,
				TargetTo:  rating.TargetTo,
				EventTime: rating.EventTime,
			})
		}
	}

	// Generate recommendation
	recommendation := s.generateSimpleRecommendation(ratings)

	// Create analysis response
	analysisResp := &response.AnalysisResponse{
		CompanyID:      companyID,
		CompanyName:    company.Name,
		Ticker:         company.Ticker,
		TotalRatings:   len(ratings),
		RecentRatings:  recentRatingResponses,
		Recommendation: recommendation,
		Summary:        ratingStats,
		GeneratedAt:    time.Now(),
	}

	return analysisResp, nil
}

// GetCompanyAnalysisByTicker provides detailed analysis for a company by ticker
func (s *analysisService) GetCompanyAnalysisByTicker(ctx context.Context, ticker string) (*response.AnalysisResponse, error) {
	// Get company by ticker
	company, err := s.companyRepo.GetByTicker(ctx, ticker)
	if err != nil {
		return nil, response.NotFound("Company with ticker " + ticker)
	}

	return s.GetCompanyAnalysis(ctx, company.ID)
}

// GetMarketOverview provides market overview statistics
func (s *analysisService) GetMarketOverview(ctx context.Context) (map[string]interface{}, error) {
	// Get total counts
	totalCompanies, err := s.companyRepo.Count(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get company count", err)
		return nil, response.InternalServerError("Failed to get market overview")
	}

	activeCompanies, err := s.companyRepo.CountActive(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get active company count", err)
		return nil, response.InternalServerError("Failed to get market overview")
	}

	totalRatings, err := s.stockRatingRepo.Count(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get rating count", err)
		return nil, response.InternalServerError("Failed to get market overview")
	}

	totalBrokerages, err := s.brokerageRepo.Count(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get brokerage count", err)
		return nil, response.InternalServerError("Failed to get market overview")
	}

	activeBrokerages, err := s.brokerageRepo.CountActive(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get active brokerage count", err)
		return nil, response.InternalServerError("Failed to get market overview")
	}

	overview := map[string]interface{}{
		"timestamp": time.Now(),
		"companies": map[string]interface{}{
			"total":  totalCompanies,
			"active": activeCompanies,
		},
		"brokerages": map[string]interface{}{
			"total":  totalBrokerages,
			"active": activeBrokerages,
		},
		"ratings": map[string]interface{}{
			"total": totalRatings,
		},
	}

	return overview, nil
}

// GetSectorAnalysis provides analysis by sector
func (s *analysisService) GetSectorAnalysis(ctx context.Context, sector string) (map[string]interface{}, error) {
	// Get companies in this sector
	companies, err := s.companyRepo.GetBySector(ctx, sector)
	if err != nil {
		s.logger.Error(ctx, "Failed to get companies by sector", err)
		return nil, response.InternalServerError("Failed to get sector analysis")
	}

	analysis := map[string]interface{}{
		"sector":        sector,
		"company_count": len(companies),
		"companies":     companies,
		"generated_at":  time.Now(),
	}

	return analysis, nil
}

// GetTopRatedCompanies gets top rated companies
func (s *analysisService) GetTopRatedCompanies(ctx context.Context, limit int) ([]*response.CompanyListResponse, error) {
	// Get top companies by rating count
	topCompanies, err := s.stockRatingRepo.GetTopCompaniesByRatingCount(ctx, 30, limit)
	if err != nil {
		s.logger.Error(ctx, "Failed to get top rated companies", err)
		return nil, response.InternalServerError("Failed to get top rated companies")
	}

	// Convert to company list responses
	responses := make([]*response.CompanyListResponse, 0, len(topCompanies))
	for _, companyCount := range topCompanies {
		// Get full company details
		company, err := s.companyRepo.GetByID(ctx, companyCount.CompanyID)
		if err != nil {
			continue // Skip if company not found
		}

		responses = append(responses, &response.CompanyListResponse{
			ID:       company.ID,
			Ticker:   company.Ticker,
			Name:     company.Name,
			Sector:   company.Sector,
			Exchange: company.Exchange,
			IsActive: company.IsActive,
		})
	}

	return responses, nil
}

// GetRatingTrends provides rating trends over time
func (s *analysisService) GetRatingTrends(ctx context.Context, period string) (map[string]interface{}, error) {
	days := 30 // Default
	switch period {
	case "week":
		days = 7
	case "month":
		days = 30
	case "quarter":
		days = 90
	case "year":
		days = 365
	}

	// Get action type distribution
	actionDistribution, err := s.stockRatingRepo.GetActionTypeDistribution(ctx, days)
	if err != nil {
		s.logger.Error(ctx, "Failed to get rating trends", err)
		return nil, response.InternalServerError("Failed to get rating trends")
	}

	trends := map[string]interface{}{
		"period":       period,
		"days":         days,
		"actions":      actionDistribution,
		"generated_at": time.Now(),
	}

	return trends, nil
}

// GetBrokerageActivity provides brokerage activity analysis
func (s *analysisService) GetBrokerageActivity(ctx context.Context, period string) (map[string]interface{}, error) {
	days := 30 // Default
	switch period {
	case "week":
		days = 7
	case "month":
		days = 30
	case "quarter":
		days = 90
	case "year":
		days = 365
	}

	// Get top brokerages by activity
	topBrokerages, err := s.stockRatingRepo.GetTopBrokeragesByRatingCount(ctx, days, 10)
	if err != nil {
		s.logger.Error(ctx, "Failed to get brokerage activity", err)
		return nil, response.InternalServerError("Failed to get brokerage activity")
	}

	activity := map[string]interface{}{
		"period":         period,
		"days":           days,
		"top_brokerages": topBrokerages,
		"generated_at":   time.Now(),
	}

	return activity, nil
}

// GenerateRecommendation generates a recommendation for a company
func (s *analysisService) GenerateRecommendation(ctx context.Context, companyID uuid.UUID) (string, error) {
	// Get company ratings
	ratings, err := s.stockRatingRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get company ratings for recommendation", err)
		return "", response.InternalServerError("Failed to generate recommendation")
	}

	if len(ratings) == 0 {
		return "No data available", nil
	}

	return s.generateSimpleRecommendation(ratings), nil
}

// GetRecommendationsByRating gets recommendations by rating type
func (s *analysisService) GetRecommendationsByRating(ctx context.Context, rating string, limit int) ([]*response.CompanyListResponse, error) {
	// Get all ratings of the specified type
	ratings, err := s.stockRatingRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get ratings", err)
		return nil, response.InternalServerError("Failed to get recommendations")
	}

	// Filter by rating type and get unique companies
	companyIDs := make(map[uuid.UUID]bool)
	for _, r := range ratings {
		if r.RatingTo == rating {
			companyIDs[r.CompanyID] = true
		}
	}

	// Convert to company list responses
	responses := make([]*response.CompanyListResponse, 0)
	count := 0
	for companyID := range companyIDs {
		if count >= limit {
			break
		}

		company, err := s.companyRepo.GetByID(ctx, companyID)
		if err != nil {
			continue // Skip if company not found
		}

		responses = append(responses, &response.CompanyListResponse{
			ID:       company.ID,
			Ticker:   company.Ticker,
			Name:     company.Name,
			Sector:   company.Sector,
			Exchange: company.Exchange,
			IsActive: company.IsActive,
		})
		count++
	}

	return responses, nil
}

// Helper methods

func (s *analysisService) calculateCompanyRatingStats(ratings []*entities.StockRating) map[string]interface{} {
	if len(ratings) == 0 {
		return map[string]interface{}{
			"total":            0,
			"action_breakdown": map[string]int{},
			"rating_breakdown": map[string]int{},
		}
	}

	actionBreakdown := make(map[string]int)
	ratingBreakdown := make(map[string]int)

	for _, rating := range ratings {
		// Count by action
		actionBreakdown[rating.Action]++

		// Count by rating
		if rating.RatingTo != "" {
			ratingBreakdown[rating.RatingTo]++
		}
	}

	return map[string]interface{}{
		"total":            len(ratings),
		"action_breakdown": actionBreakdown,
		"rating_breakdown": ratingBreakdown,
	}
}

// Helper method to generate simple recommendations
func (s *analysisService) generateSimpleRecommendation(ratings []*entities.StockRating) string {
	if len(ratings) == 0 {
		return "No data available"
	}

	// Count recent ratings by type
	buyCount, holdCount, sellCount := 0, 0, 0

	// Look at recent ratings (last 5 or all if less than 5)
	recentCount := 5
	if len(ratings) < recentCount {
		recentCount = len(ratings)
	}

	recentRatings := ratings[len(ratings)-recentCount:]

	for _, rating := range recentRatings {
		switch rating.RatingTo {
		case "Buy", "Strong Buy", "Outperform":
			buyCount++
		case "Hold", "Neutral":
			holdCount++
		case "Sell", "Strong Sell", "Underperform":
			sellCount++
		}
	}

	// Generate recommendation based on majority
	if buyCount > holdCount && buyCount > sellCount {
		return "Buy"
	} else if sellCount > buyCount && sellCount > holdCount {
		return "Sell"
	} else {
		return "Hold"
	}
}
