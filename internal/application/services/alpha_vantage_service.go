package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/alphavantage"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// AlphaVantageService provides business logic for Alpha Vantage API integration
type AlphaVantageService struct {
	client                     *alphavantage.Client
	adapter                    *alphavantage.Adapter
	financialRepo              interfaces.FinancialMetricsRepository
	technicalRepo              interfaces.TechnicalIndicatorsRepository
	historicalRepo             interfaces.HistoricalDataRepository
	companyRepo                interfaces.CompanyRepository
	logger                     logger.Logger
	technicalIndicatorsService *AlphaVantageTechnicalIndicatorsService
	historicalDataService      *AlphaVantageHistoricalDataService
}

// NewAlphaVantageService creates a new instance of AlphaVantageService
func NewAlphaVantageService(
	client *alphavantage.Client,
	adapter *alphavantage.Adapter,
	financialRepo interfaces.FinancialMetricsRepository,
	technicalRepo interfaces.TechnicalIndicatorsRepository,
	historicalRepo interfaces.HistoricalDataRepository,
	companyRepo interfaces.CompanyRepository, logger logger.Logger,
) *AlphaVantageService {
	// Create specialized technical indicators service
	technicalIndicatorsService := NewAlphaVantageTechnicalIndicatorsService(
		client,
		adapter,
		technicalRepo,
		logger,
	)

	// Create specialized historical data service
	historicalDataService := NewAlphaVantageHistoricalDataService(
		client,
		adapter,
		historicalRepo,
		logger,
	)

	return &AlphaVantageService{
		client:                     client,
		adapter:                    adapter,
		financialRepo:              financialRepo,
		technicalRepo:              technicalRepo,
		historicalRepo:             historicalRepo,
		companyRepo:                companyRepo,
		logger:                     logger,
		technicalIndicatorsService: technicalIndicatorsService,
		historicalDataService:      historicalDataService,
	}
}

// GetFinancialMetricsFromAPI fetches financial metrics from Alpha Vantage API and saves to database
func (s *AlphaVantageService) GetFinancialMetricsFromAPI(ctx context.Context, symbol string) (*entities.FinancialMetrics, error) {
	s.logger.Info(ctx, "Fetching financial metrics from Alpha Vantage API",
		logger.String("symbol", symbol))

	// Get company for the symbol
	company, err := s.getOrCreateCompany(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get company for symbol %s: %w", symbol, err)
	}
	// Fetch overview data from Alpha Vantage
	overviewData, err := s.client.GetCompanyOverview(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get company overview from Alpha Vantage: %w", err)
	}
	// Update company with real information from AlphaVantage
	if err := s.updateCompanyWithOverviewData(ctx, company, overviewData); err != nil {
		s.logger.Warn(ctx, "Failed to update company with overview data",
			logger.String("symbol", symbol))
	}

	// Convert to financial metrics entity
	financialMetrics, err := s.adapter.CompanyOverviewToFinancialMetrics(ctx, overviewData, company.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert overview data to financial metrics: %w", err)
	}
	// Save to database
	if err := s.financialRepo.Create(ctx, financialMetrics); err != nil {
		s.logger.Error(ctx, "Failed to save financial metrics to database", err,
			logger.String("symbol", symbol))
		// Return the metrics even if save fails, as we have the data
		return financialMetrics, nil
	}

	// Retrieve the saved entity with company relationship loaded
	savedMetrics, err := s.financialRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Failed to retrieve saved financial metrics with company data", err,
			logger.String("symbol", symbol))
		// Return the original metrics if retrieval fails
		return financialMetrics, nil
	}

	s.logger.Info(ctx, "Successfully fetched and saved financial metrics",
		logger.String("symbol", symbol),
		logger.String("financial_metrics_id", savedMetrics.ID.String()))

	return savedMetrics, nil
}

// GetTechnicalIndicatorsFromAPI fetches technical indicators from Alpha Vantage API and saves to database
func (s *AlphaVantageService) GetTechnicalIndicatorsFromAPI(ctx context.Context, symbol string) ([]*entities.TechnicalIndicators, error) {
	s.logger.Info(ctx, "Fetching technical indicators from Alpha Vantage API",
		logger.String("symbol", symbol))

	// Get company for the symbol
	company, err := s.getOrCreateCompany(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get company for symbol %s: %w", symbol, err)
	}

	// Delegate to specialized technical indicators service
	return s.technicalIndicatorsService.GetTechnicalIndicatorsFromAPI(ctx, symbol, company.ID)
}

// GetTechnicalIndicatorFromAPI fetches a specific technical indicator from Alpha Vantage API
func (s *AlphaVantageService) GetTechnicalIndicatorFromAPI(ctx context.Context, symbol, indicator, interval, timePeriod, seriesType string) ([]*entities.TechnicalIndicators, error) {
	s.logger.Info(ctx, "Fetching specific technical indicator from Alpha Vantage API",
		logger.String("symbol", symbol),
		logger.String("indicator", indicator),
		logger.String("interval", interval),
		logger.String("time_period", timePeriod),
		logger.String("series_type", seriesType))

	// Get company for the symbol
	company, err := s.getOrCreateCompany(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get company for symbol %s: %w", symbol, err)
	}

	// Delegate to specialized technical indicators service
	return s.technicalIndicatorsService.GetTechnicalIndicatorFromAPI(ctx, symbol, indicator, interval, timePeriod, seriesType, company.ID)
}

// GetHistoricalDataFromAPI fetches historical data from Alpha Vantage API and saves to database
func (s *AlphaVantageService) GetHistoricalDataFromAPI(ctx context.Context, symbol, period, outputSize, interval, adjusted string) ([]*entities.HistoricalData, error) {
	s.logger.Info(ctx, "Fetching historical data from Alpha Vantage API",
		logger.String("symbol", symbol),
		logger.String("period", period),
		logger.String("outputSize", outputSize),
		logger.String("interval", interval),
		logger.String("adjusted", adjusted))

	// Get company for the symbol
	company, err := s.getOrCreateCompany(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get company for symbol %s: %w", symbol, err)
	}

	// Delegate to specialized historical data service
	return s.historicalDataService.GetHistoricalDataFromAPI(ctx, symbol, period, outputSize, interval, adjusted, company.ID)
}

// RefreshStockData refreshes all data for a single stock symbol
func (s *AlphaVantageService) RefreshStockData(ctx context.Context, symbol string) error {
	s.logger.Info(ctx, "Refreshing all stock data from Alpha Vantage",
		logger.String("symbol", symbol))

	var errors []error

	// Refresh financial metrics
	if _, err := s.GetFinancialMetricsFromAPI(ctx, symbol); err != nil {
		errors = append(errors, fmt.Errorf("financial metrics refresh failed: %w", err))
	}

	// Refresh technical indicators
	if _, err := s.GetTechnicalIndicatorsFromAPI(ctx, symbol); err != nil {
		errors = append(errors, fmt.Errorf("technical indicators refresh failed: %w", err))
	}

	// Refresh historical data (daily)
	if _, err := s.GetHistoricalDataFromAPI(ctx, symbol, "daily", "compact", "", ""); err != nil {
		errors = append(errors, fmt.Errorf("historical data refresh failed: %w", err))
	}

	if len(errors) > 0 {
		s.logger.Error(ctx, "Some data refresh operations failed",
			fmt.Errorf("refresh errors: %v", errors),
			logger.String("symbol", symbol))
		return fmt.Errorf("refresh completed with %d errors: %v", len(errors), errors)
	}

	s.logger.Info(ctx, "Successfully refreshed all stock data",
		logger.String("symbol", symbol))

	return nil
}

// BulkRefreshStockData refreshes data for multiple stock symbols with concurrency
func (s *AlphaVantageService) BulkRefreshStockData(ctx context.Context, symbols []string) error {
	s.logger.Info(ctx, "Starting bulk refresh of stock data",
		logger.Int("symbols_count", len(symbols)))

	const maxWorkers = 5 // Limit concurrency to respect API rate limits
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	errors := make(chan error, len(symbols))

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := s.RefreshStockData(ctx, sym); err != nil {
				errors <- fmt.Errorf("failed to refresh %s: %w", sym, err)
			}

			// Add delay to respect API rate limits
			time.Sleep(200 * time.Millisecond)
		}(symbol)
	}

	wg.Wait()
	close(errors)

	var refreshErrors []error
	for err := range errors {
		refreshErrors = append(refreshErrors, err)
	}

	if len(refreshErrors) > 0 {
		s.logger.Error(ctx, "Some bulk refresh operations failed",
			fmt.Errorf("bulk refresh errors: %v", refreshErrors))
		return fmt.Errorf("bulk refresh completed with %d errors: %v", len(refreshErrors), refreshErrors)
	}

	s.logger.Info(ctx, "Successfully completed bulk refresh of stock data",
		logger.Int("symbols_count", len(symbols)))

	return nil
}

// SyncWithDatabase synchronizes Alpha Vantage data with local database for a symbol
func (s *AlphaVantageService) SyncWithDatabase(ctx context.Context, symbol string) error {
	s.logger.Info(ctx, "Synchronizing Alpha Vantage data with database",
		logger.String("symbol", symbol))

	// Check if we have recent data in database
	existingMetrics, err := s.financialRepo.GetBySymbol(ctx, symbol)
	if err == nil && existingMetrics != nil {
		// Check if data is recent (less than 24 hours old)
		if time.Since(existingMetrics.LastUpdated) < 24*time.Hour {
			s.logger.Info(ctx, "Recent data found in database, skipping API call",
				logger.String("symbol", symbol),
				logger.Time("last_updated", existingMetrics.LastUpdated))
			return nil
		}
	}

	// Data is stale or doesn't exist, refresh from API
	return s.RefreshStockData(ctx, symbol)
}

// getOrCreateCompany gets existing company or creates a new one for the symbol
func (s *AlphaVantageService) getOrCreateCompany(ctx context.Context, symbol string) (*entities.Company, error) {
	// Try to get existing company by ticker
	company, err := s.companyRepo.GetByTicker(ctx, symbol)
	if err == nil && company != nil {
		return company, nil
	}

	// Company doesn't exist, create a new one
	company = &entities.Company{
		ID:        uuid.New(),
		Name:      symbol, // Will be updated with real name from API
		Ticker:    symbol,
		Exchange:  "UNKNOWN",
		Sector:    "UNKNOWN",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	s.logger.Info(ctx, "Created new company record",
		logger.String("symbol", symbol),
		logger.String("company_id", company.ID.String()))

	return company, nil
}

// updateCompanyWithOverviewData updates company entity with data from AlphaVantage overview
func (s *AlphaVantageService) updateCompanyWithOverviewData(ctx context.Context, company *entities.Company, overview *alphavantage.CompanyOverviewResponse) error {
	if overview == nil || overview.Symbol == "" {
		return fmt.Errorf("invalid overview data")
	}

	// Update company fields with real data from AlphaVantage
	company.Name = overview.Name
	company.Sector = overview.Sector
	company.Exchange = overview.Exchange
	// Parse market cap if available
	if overview.MarketCapitalization != "" && overview.MarketCapitalization != "None" {
		// Remove any non-numeric characters except decimal point
		cleanValue := strings.ReplaceAll(overview.MarketCapitalization, ",", "")
		if marketCap, err := strconv.ParseFloat(cleanValue, 64); err == nil {
			company.MarketCap = marketCap
		}
	}

	// Save updated company
	if err := s.companyRepo.Update(ctx, company); err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	s.logger.Info(ctx, "Updated company with overview data",
		logger.String("symbol", company.Ticker),
		logger.String("name", company.Name),
		logger.String("sector", company.Sector))

	return nil
}
