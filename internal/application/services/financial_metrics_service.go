package services

import (
	"context"
	"fmt"
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// FinancialMetricsService provides business logic for financial metrics
type FinancialMetricsService struct {
	financialRepo interfaces.FinancialMetricsRepository
	companyRepo   interfaces.CompanyRepository
}

// NewFinancialMetricsService creates a new instance of FinancialMetricsService
func NewFinancialMetricsService(
	financialRepo interfaces.FinancialMetricsRepository,
	companyRepo interfaces.CompanyRepository,
) *FinancialMetricsService {
	return &FinancialMetricsService{
		financialRepo: financialRepo,
		companyRepo:   companyRepo,
	}
}

// GetFinancialMetricsBySymbol retrieves financial metrics for a specific symbol
func (s *FinancialMetricsService) GetFinancialMetricsBySymbol(ctx context.Context, symbol string) (*entities.FinancialMetrics, error) {
	metrics, err := s.financialRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial metrics for symbol %s: %w", symbol, err)
	}
	return metrics, nil
}

// CreateFinancialMetrics creates new financial metrics record
func (s *FinancialMetricsService) CreateFinancialMetrics(ctx context.Context, metrics *entities.FinancialMetrics) error {
	// Validate company exists
	if _, err := s.companyRepo.GetByID(ctx, metrics.CompanyID); err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Set timestamps
	now := time.Now()
	metrics.LastUpdated = now

	if err := s.financialRepo.Create(ctx, metrics); err != nil {
		return fmt.Errorf("failed to create financial metrics: %w", err)
	}

	return nil
}

// UpdateFinancialMetrics updates existing financial metrics
func (s *FinancialMetricsService) UpdateFinancialMetrics(ctx context.Context, metrics *entities.FinancialMetrics) error {
	// Check if record exists
	existing, err := s.financialRepo.GetByID(ctx, metrics.ID)
	if err != nil {
		return fmt.Errorf("financial metrics not found: %w", err)
	}

	// Update timestamp
	metrics.LastUpdated = time.Now()

	// Preserve creation timestamp
	metrics.CreatedAt = existing.CreatedAt

	if err := s.financialRepo.Update(ctx, metrics); err != nil {
		return fmt.Errorf("failed to update financial metrics: %w", err)
	}

	return nil
}

// GetFinancialAnalysis provides comprehensive financial analysis for a symbol
func (s *FinancialMetricsService) GetFinancialAnalysis(ctx context.Context, symbol string) (*response.FinancialAnalysisResponse, error) {
	metrics, err := s.financialRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial metrics: %w", err)
	}

	// Calculate financial score
	financialScore := metrics.CalculateFinancialScore()

	// Determine stock type
	stockType := s.determineStockType(metrics)

	// Get analyst consensus
	analystConsensus := metrics.GetAnalystConsensus()

	// Generate insights
	insights := s.generateFinancialInsights(metrics)

	analysis := &response.FinancialAnalysisResponse{
		Symbol:           metrics.Symbol,
		FinancialScore:   financialScore,
		StockType:        stockType,
		AnalystConsensus: analystConsensus,
		Insights:         insights,

		// Valuation Metrics
		PERatio:      metrics.PERatio,
		PEGRatio:     metrics.PEGRatio,
		PriceToBook:  metrics.PriceToBook,
		PriceToSales: metrics.PriceToSales,

		// Profitability
		ROE:       metrics.ROE,
		ROA:       metrics.ROA,
		NetMargin: metrics.NetMargin,

		// Financial Health
		DebtToEquity: metrics.DebtToEquity,
		CurrentRatio: metrics.CurrentRatio,

		// Growth
		RevenueGrowthTTM:  metrics.RevenueGrowthTTM,
		EarningsGrowthTTM: metrics.EarningsGrowthTTM,

		LastUpdated: metrics.LastUpdated,
	}

	return analysis, nil
}

// GetValueStocks retrieves stocks that appear undervalued
func (s *FinancialMetricsService) GetValueStocks(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error) {
	return s.financialRepo.GetValueStocks(ctx, limit)
}

// GetGrowthStocks retrieves stocks with strong growth characteristics
func (s *FinancialMetricsService) GetGrowthStocks(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error) {
	return s.financialRepo.GetGrowthStocks(ctx, limit)
}

// GetDividendStocks retrieves dividend-paying stocks
func (s *FinancialMetricsService) GetDividendStocks(ctx context.Context, minYield float64, limit int) ([]*entities.FinancialMetrics, error) {
	return s.financialRepo.GetDividendStocks(ctx, minYield, limit)
}

// GetTopPerformers retrieves top performing stocks by ROE
func (s *FinancialMetricsService) GetTopPerformers(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error) {
	return s.financialRepo.GetTopByROE(ctx, limit)
}

// GetSectorAnalysis provides financial analysis for a specific sector
func (s *FinancialMetricsService) GetSectorAnalysis(ctx context.Context, sector string) (*response.SectorAnalysisResponse, error) {
	// Get sector averages
	averages, err := s.financialRepo.GetSectorAverages(ctx, sector)
	if err != nil {
		return nil, fmt.Errorf("failed to get sector averages: %w", err)
	}

	// Get sector stocks
	stocks, err := s.financialRepo.GetBySector(ctx, sector)
	if err != nil {
		return nil, fmt.Errorf("failed to get sector stocks: %w", err)
	}

	analysis := &response.SectorAnalysisResponse{
		Sector:      sector,
		TotalStocks: len(stocks),
		Averages:    averages,
		TopStocks:   s.getTopStocksFromList(stocks, 10),
	}

	return analysis, nil
}

// ScreenStocks screens stocks based on financial criteria
func (s *FinancialMetricsService) ScreenStocks(ctx context.Context, criteria response.StockScreenCriteria) ([]*entities.FinancialMetrics, error) {
	var results []*entities.FinancialMetrics
	var err error

	// Apply PE ratio filter
	if criteria.MaxPE > 0 {
		results, err = s.financialRepo.GetByPERatio(ctx, 0, criteria.MaxPE)
		if err != nil {
			return nil, fmt.Errorf("failed to filter by PE ratio: %w", err)
		}
	}

	// Apply ROE filter
	if criteria.MinROE > 0 {
		if results == nil {
			results, err = s.financialRepo.GetByROE(ctx, criteria.MinROE)
		} else {
			results = s.filterByROE(results, criteria.MinROE)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to filter by ROE: %w", err)
		}
	}

	// Apply growth filter
	if criteria.MinGrowth > 0 {
		if results == nil {
			results, err = s.financialRepo.GetByGrowthRate(ctx, criteria.MinGrowth)
		} else {
			results = s.filterByGrowth(results, criteria.MinGrowth)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to filter by growth: %w", err)
		}
	}

	// Apply debt to equity filter
	if criteria.MaxDebtToEquity > 0 {
		if results == nil {
			results, err = s.financialRepo.GetByDebtToEquity(ctx, criteria.MaxDebtToEquity)
		} else {
			results = s.filterByDebtToEquity(results, criteria.MaxDebtToEquity)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to filter by debt to equity: %w", err)
		}
	}

	return results, nil
}

// RefreshStaleData refreshes financial metrics that are outdated
func (s *FinancialMetricsService) RefreshStaleData(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	staleData, err := s.financialRepo.GetStaleData(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("failed to get stale data: %w", err)
	}

	// This would trigger data refresh from external APIs
	// Implementation would depend on the specific data source service
	for _, metrics := range staleData {
		// Log which metrics need refresh
		fmt.Printf("Financial metrics for %s need refresh (last updated: %s)\n",
			metrics.Symbol, metrics.LastUpdated.Format(time.RFC3339))
	}

	return nil
}

// Helper methods

func (s *FinancialMetricsService) determineStockType(metrics *entities.FinancialMetrics) string {
	if metrics.IsGrowthStock() {
		return "GROWTH"
	} else if metrics.IsValueStock() {
		return "VALUE"
	} else if metrics.DividendYield > 3.0 {
		return "DIVIDEND"
	}
	return "BLEND"
}

func (s *FinancialMetricsService) generateFinancialInsights(metrics *entities.FinancialMetrics) []string {
	var insights []string

	// Profitability insights
	if metrics.ROE > 20 {
		insights = append(insights, "Excellent return on equity indicates efficient use of shareholder capital")
	} else if metrics.ROE < 10 {
		insights = append(insights, "Below-average return on equity may indicate operational inefficiencies")
	}

	// Valuation insights
	if metrics.PERatio > 0 && metrics.PERatio < 15 {
		insights = append(insights, "Trading at attractive valuation relative to earnings")
	} else if metrics.PERatio > 30 {
		insights = append(insights, "High P/E ratio suggests growth expectations or potential overvaluation")
	}

	// Growth insights
	if metrics.EarningsGrowthTTM > 20 {
		insights = append(insights, "Strong earnings growth demonstrates business momentum")
	} else if metrics.EarningsGrowthTTM < 0 {
		insights = append(insights, "Declining earnings warrant careful analysis")
	}

	// Financial health insights
	if metrics.IsHealthy() {
		insights = append(insights, "Strong balance sheet with healthy financial ratios")
	} else {
		if metrics.DebtToEquity > 2.0 {
			insights = append(insights, "High debt levels may indicate financial risk")
		}
		if metrics.CurrentRatio < 1.0 {
			insights = append(insights, "Current ratio below 1.0 suggests potential liquidity concerns")
		}
	}

	return insights
}

func (s *FinancialMetricsService) getTopStocksFromList(stocks []*entities.FinancialMetrics, limit int) []*entities.FinancialMetrics {
	if len(stocks) <= limit {
		return stocks
	}

	// Sort by financial score
	for i := 0; i < len(stocks)-1; i++ {
		for j := i + 1; j < len(stocks); j++ {
			if stocks[i].CalculateFinancialScore() < stocks[j].CalculateFinancialScore() {
				stocks[i], stocks[j] = stocks[j], stocks[i]
			}
		}
	}

	return stocks[:limit]
}

func (s *FinancialMetricsService) filterByROE(stocks []*entities.FinancialMetrics, minROE float64) []*entities.FinancialMetrics {
	var filtered []*entities.FinancialMetrics
	for _, stock := range stocks {
		if stock.ROE >= minROE {
			filtered = append(filtered, stock)
		}
	}
	return filtered
}

func (s *FinancialMetricsService) filterByGrowth(stocks []*entities.FinancialMetrics, minGrowth float64) []*entities.FinancialMetrics {
	var filtered []*entities.FinancialMetrics
	for _, stock := range stocks {
		if stock.EarningsGrowthTTM >= minGrowth || stock.RevenueGrowthTTM >= minGrowth {
			filtered = append(filtered, stock)
		}
	}
	return filtered
}

func (s *FinancialMetricsService) filterByDebtToEquity(stocks []*entities.FinancialMetrics, maxDebtToEquity float64) []*entities.FinancialMetrics {
	var filtered []*entities.FinancialMetrics
	for _, stock := range stocks {
		if stock.DebtToEquity <= maxDebtToEquity {
			filtered = append(filtered, stock)
		}
	}
	return filtered
}
