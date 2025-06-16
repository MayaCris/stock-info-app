package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/alphavantage"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// HistoricalDataStrategy defines the interface for historical data processing
type HistoricalDataStrategy interface {
	GetPeriodName() string
	FetchData(ctx context.Context, client *alphavantage.Client, symbol, outputSize, interval, adjusted string) (interface{}, error)
	ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID) ([]*entities.HistoricalData, error)
}

// DailyDataStrategy implements daily historical data processing
type DailyDataStrategy struct{}

func (d *DailyDataStrategy) GetPeriodName() string {
	return "daily"
}

func (d *DailyDataStrategy) FetchData(ctx context.Context, client *alphavantage.Client, symbol, outputSize, interval, adjusted string) (interface{}, error) {
	// Use provided outputSize or default
	if outputSize == "" {
		outputSize = "compact"
	}

	// For daily data, interval and adjusted parameters are not typically used by AlphaVantage
	// But we could extend this later for intraday data
	return client.GetTimeSeriesDaily(ctx, symbol, outputSize)
}

func (d *DailyDataStrategy) ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID) ([]*entities.HistoricalData, error) {
	dailyResponse, ok := response.(*alphavantage.TimeSeriesDailyResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type for daily data: expected *TimeSeriesDailyResponse")
	}
	return adapter.TimeSeriesDataToHistoricalData(ctx, dailyResponse, symbol, companyID)
}

// WeeklyDataStrategy implements weekly historical data processing
type WeeklyDataStrategy struct{}

func (w *WeeklyDataStrategy) GetPeriodName() string {
	return "weekly"
}

func (w *WeeklyDataStrategy) FetchData(ctx context.Context, client *alphavantage.Client, symbol, outputSize, interval, adjusted string) (interface{}, error) {
	// Weekly data typically doesn't use outputSize parameter in AlphaVantage API
	// But we could extend this later if needed
	return client.GetTimeSeriesWeekly(ctx, symbol)
}

func (w *WeeklyDataStrategy) ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID) ([]*entities.HistoricalData, error) {
	// TODO: Implement weekly data conversion in adapter
	return nil, fmt.Errorf("weekly data conversion not yet implemented")
}

// MonthlyDataStrategy implements monthly historical data processing
type MonthlyDataStrategy struct{}

func (m *MonthlyDataStrategy) GetPeriodName() string {
	return "monthly"
}

func (m *MonthlyDataStrategy) FetchData(ctx context.Context, client *alphavantage.Client, symbol, outputSize, interval, adjusted string) (interface{}, error) {
	// Monthly data typically doesn't use outputSize parameter in AlphaVantage API
	// But we could extend this later if needed
	return client.GetTimeSeriesMonthly(ctx, symbol)
}

func (m *MonthlyDataStrategy) ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID) ([]*entities.HistoricalData, error) {
	// TODO: Implement monthly data conversion in adapter
	return nil, fmt.Errorf("monthly data conversion not yet implemented")
}

// AlphaVantageHistoricalDataService provides business logic for Alpha Vantage historical data
type AlphaVantageHistoricalDataService struct {
	client     *alphavantage.Client
	adapter    *alphavantage.Adapter
	repository interfaces.HistoricalDataRepository
	logger     logger.Logger
	strategies map[string]HistoricalDataStrategy
}

// NewAlphaVantageHistoricalDataService creates a new instance
func NewAlphaVantageHistoricalDataService(
	client *alphavantage.Client,
	adapter *alphavantage.Adapter,
	repository interfaces.HistoricalDataRepository,
	logger logger.Logger,
) *AlphaVantageHistoricalDataService {
	strategies := map[string]HistoricalDataStrategy{
		"daily":   &DailyDataStrategy{},
		"weekly":  &WeeklyDataStrategy{},
		"monthly": &MonthlyDataStrategy{},
	}

	return &AlphaVantageHistoricalDataService{
		client:     client,
		adapter:    adapter,
		repository: repository,
		logger:     logger,
		strategies: strategies,
	}
}

// GetHistoricalDataFromAPI fetches historical data using strategy pattern
func (s *AlphaVantageHistoricalDataService) GetHistoricalDataFromAPI(ctx context.Context, symbol, period, outputSize, interval, adjusted string, companyID uuid.UUID) ([]*entities.HistoricalData, error) {
	s.logger.Info(ctx, "Fetching historical data from Alpha Vantage API",
		logger.String("symbol", symbol),
		logger.String("period", period),
		logger.String("outputSize", outputSize),
		logger.String("interval", interval),
		logger.String("adjusted", adjusted))

	// Get strategy for the requested period
	strategy, exists := s.strategies[period]
	if !exists {
		return nil, fmt.Errorf("unsupported period: %s", period)
	}

	// Fetch data using strategy
	response, err := strategy.FetchData(ctx, s.client, symbol, outputSize, interval, adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s data from Alpha Vantage: %w", period, err)
	}

	// Convert to entities using strategy
	historicalData, err := strategy.ConvertToEntity(ctx, s.adapter, response, symbol, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s time series data: %w", period, err)
	}

	// Save to database in batches
	if err := s.saveHistoricalDataBatch(ctx, historicalData, symbol); err != nil {
		s.logger.Error(ctx, "Failed to save historical data to database", err,
			logger.String("symbol", symbol))
		// Return the data even if save fails
	}

	s.logger.Info(ctx, "Successfully fetched and saved historical data",
		logger.String("symbol", symbol),
		logger.String("period", period),
		logger.Int("records_count", len(historicalData)))

	return historicalData, nil
}

// saveHistoricalDataBatch saves historical data in batches to optimize database performance
func (s *AlphaVantageHistoricalDataService) saveHistoricalDataBatch(ctx context.Context, historicalData []*entities.HistoricalData, symbol string) error {
	// Check if we have data to save
	if len(historicalData) == 0 {
		s.logger.Warn(ctx, "No historical data to save",
			logger.String("symbol", symbol))
		return nil
	}

	// Test database connectivity with a single record before attempting bulk save
	testErr := s.repository.Create(ctx, historicalData[0])
	if testErr != nil {
		// Check if error indicates table doesn't exist
		if strings.Contains(strings.ToLower(testErr.Error()), "does not exist") ||
			strings.Contains(strings.ToLower(testErr.Error()), "relation") {
			// Table doesn't exist - log single warning and return without persistence
			s.logger.Warn(ctx, "Database table does not exist, skipping all historical data persistence operations",
				logger.String("symbol", symbol),
				logger.String("table_error", testErr.Error()),
				logger.Int("data_count", len(historicalData)))
			return nil
		} else {
			// Other type of error - log but continue trying to save the rest
			s.logger.Error(ctx, "Failed to save first historical data record (non-table error)", testErr,
				logger.String("symbol", symbol))
		}
	}

	// First save was successful, continue with the rest
	const batchSize = 100
	savedCount := 1 // First record already saved

	for i := 1; i < len(historicalData); i += batchSize {
		end := i + batchSize
		if end > len(historicalData) {
			end = len(historicalData)
		}

		batch := historicalData[i:end]
		for _, data := range batch {
			if err := s.repository.Create(ctx, data); err != nil {
				s.logger.Error(ctx, "Failed to save historical data record", err,
					logger.String("symbol", symbol),
					logger.Time("date", data.Date))
				continue
			}
			savedCount++
		}
	}

	if savedCount > 0 {
		s.logger.Info(ctx, "Successfully saved historical data to database",
			logger.String("symbol", symbol),
			logger.Int("saved_count", savedCount),
			logger.Int("total_count", len(historicalData)))
	}

	return nil
}

// AddStrategy allows adding new historical data strategies (Open/Closed Principle)
func (s *AlphaVantageHistoricalDataService) AddStrategy(period string, strategy HistoricalDataStrategy) {
	s.strategies[period] = strategy
}

// GetSupportedPeriods returns the list of supported periods
func (s *AlphaVantageHistoricalDataService) GetSupportedPeriods() []string {
	periods := make([]string, 0, len(s.strategies))
	for period := range s.strategies {
		periods = append(periods, period)
	}
	return periods
}
