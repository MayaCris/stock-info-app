package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/alphavantage"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// TechnicalIndicatorStrategy defines the interface for technical indicator processing
type TechnicalIndicatorStrategy interface {
	GetIndicatorName() string
	FetchData(ctx context.Context, client *alphavantage.Client, symbol, interval, timePeriod, seriesType string) (interface{}, error)
	ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID, timePeriod, interval string) ([]*entities.TechnicalIndicators, error)
}

// RSIStrategy implements RSI technical indicator processing
type RSIStrategy struct{}

func (r *RSIStrategy) GetIndicatorName() string {
	return "RSI"
}

func (r *RSIStrategy) FetchData(ctx context.Context, client *alphavantage.Client, symbol, interval, timePeriod, seriesType string) (interface{}, error) {
	// Use provided parameters or defaults
	if interval == "" {
		interval = "daily"
	}
	if timePeriod == "" {
		timePeriod = "14"
	}
	if seriesType == "" {
		seriesType = "close"
	}
	return client.GetRSI(ctx, symbol, interval, timePeriod, seriesType)
}

func (r *RSIStrategy) ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID, timePeriod, interval string) ([]*entities.TechnicalIndicators, error) {
	rsiResp, ok := response.(*alphavantage.RSIResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type for RSI indicator: expected *RSIResponse")
	}

	// Convert timePeriod string to int
	period := 14 // default
	if timePeriod != "" {
		if p, err := strconv.Atoi(timePeriod); err == nil {
			period = p
		}
	}

	return adapter.RSIResponseToTechnicalIndicators(ctx, rsiResp, symbol, companyID, period)
}

// MACDStrategy implements MACD technical indicator processing
type MACDStrategy struct{}

func (m *MACDStrategy) GetIndicatorName() string {
	return "MACD"
}

func (m *MACDStrategy) FetchData(ctx context.Context, client *alphavantage.Client, symbol, interval, timePeriod, seriesType string) (interface{}, error) {
	// Use provided parameters or defaults
	if interval == "" {
		interval = "daily"
	}
	if seriesType == "" {
		seriesType = "close"
	}
	// MACD has specific parameters: fast_period=12, slow_period=26, signal_period=9
	// For now we'll use defaults, but could be made configurable later
	return client.GetMACD(ctx, symbol, interval, "12", "26", "9", seriesType)
}

func (m *MACDStrategy) ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID, timePeriod, interval string) ([]*entities.TechnicalIndicators, error) {
	macdResp, ok := response.(*alphavantage.MACDResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type for MACD indicator: expected *MACDResponse")
	}
	return adapter.MACDResponseToTechnicalIndicators(ctx, macdResp, symbol, companyID)
}

// SMAStrategy implements SMA technical indicator processing
type SMAStrategy struct {
	Period int
}

func (s *SMAStrategy) GetIndicatorName() string {
	return "SMA"
}

func (s *SMAStrategy) FetchData(ctx context.Context, client *alphavantage.Client, symbol, interval, timePeriod, seriesType string) (interface{}, error) {
	// Use provided parameters or defaults
	if interval == "" {
		interval = "daily"
	}
	if timePeriod == "" {
		timePeriod = "20"
	}
	if seriesType == "" {
		seriesType = "close"
	}
	return client.GetSMA(ctx, symbol, interval, timePeriod, seriesType)
}

func (s *SMAStrategy) ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID, timePeriod, interval string) ([]*entities.TechnicalIndicators, error) {
	smaResp, ok := response.(*alphavantage.SMAResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type for SMA indicator: expected *SMAResponse")
	}

	// Use the provided timePeriod or fallback to struct field
	period := s.Period
	if timePeriod != "" {
		if p, err := strconv.Atoi(timePeriod); err == nil {
			period = p
		}
	}

	return adapter.SMAResponseToTechnicalIndicators(ctx, smaResp, symbol, companyID, period)
}

// EMAStrategy implements EMA technical indicator processing
type EMAStrategy struct {
	Period int
}

func (e *EMAStrategy) GetIndicatorName() string {
	return "EMA"
}

func (e *EMAStrategy) FetchData(ctx context.Context, client *alphavantage.Client, symbol, interval, timePeriod, seriesType string) (interface{}, error) {
	// Use provided parameters or defaults
	if interval == "" {
		interval = "daily"
	}
	if timePeriod == "" {
		timePeriod = "20"
	}
	if seriesType == "" {
		seriesType = "close"
	}
	return client.GetEMA(ctx, symbol, interval, timePeriod, seriesType)
}

func (e *EMAStrategy) ConvertToEntity(ctx context.Context, adapter *alphavantage.Adapter, response interface{}, symbol string, companyID uuid.UUID, timePeriod, interval string) ([]*entities.TechnicalIndicators, error) {
	emaResp, ok := response.(*alphavantage.EMAResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type for EMA indicator: expected *EMAResponse")
	}

	// Use the provided timePeriod or fallback to struct field
	period := e.Period
	if timePeriod != "" {
		if p, err := strconv.Atoi(timePeriod); err == nil {
			period = p
		}
	}

	return adapter.EMAResponseToTechnicalIndicators(ctx, emaResp, symbol, companyID, period)
}

// AlphaVantageTechnicalIndicatorsService provides business logic for Alpha Vantage technical indicators
type AlphaVantageTechnicalIndicatorsService struct {
	client     *alphavantage.Client
	adapter    *alphavantage.Adapter
	repository interfaces.TechnicalIndicatorsRepository
	logger     logger.Logger
	strategies []TechnicalIndicatorStrategy
}

// NewAlphaVantageTechnicalIndicatorsService creates a new instance
func NewAlphaVantageTechnicalIndicatorsService(
	client *alphavantage.Client,
	adapter *alphavantage.Adapter,
	repository interfaces.TechnicalIndicatorsRepository,
	logger logger.Logger,
) *AlphaVantageTechnicalIndicatorsService {
	strategies := []TechnicalIndicatorStrategy{
		&RSIStrategy{},
		&MACDStrategy{},
		&SMAStrategy{Period: 20},
		&EMAStrategy{Period: 20},
	}

	return &AlphaVantageTechnicalIndicatorsService{
		client:     client,
		adapter:    adapter,
		repository: repository,
		logger:     logger,
		strategies: strategies,
	}
}

// GetTechnicalIndicatorsFromAPI fetches technical indicators using strategy pattern
func (s *AlphaVantageTechnicalIndicatorsService) GetTechnicalIndicatorsFromAPI(ctx context.Context, symbol string, companyID uuid.UUID) ([]*entities.TechnicalIndicators, error) {
	s.logger.Info(ctx, "Fetching technical indicators from Alpha Vantage API",
		logger.String("symbol", symbol))

	var allIndicators []*entities.TechnicalIndicators
	for _, strategy := range s.strategies {
		indicators, err := s.processIndicatorStrategy(ctx, strategy, symbol, "daily", "", "close", companyID)
		if err != nil {
			s.logger.Error(ctx, "Failed to process technical indicator strategy", err,
				logger.String("symbol", symbol),
				logger.String("indicator", strategy.GetIndicatorName()))
			continue
		}
		allIndicators = append(allIndicators, indicators...)
	}

	s.logger.Info(ctx, "Successfully fetched technical indicators",
		logger.String("symbol", symbol),
		logger.Int("indicators_count", len(allIndicators)))

	return allIndicators, nil
}

// processIndicatorStrategy processes a single technical indicator strategy
func (s *AlphaVantageTechnicalIndicatorsService) processIndicatorStrategy(ctx context.Context, strategy TechnicalIndicatorStrategy, symbol, interval, timePeriod, seriesType string, companyID uuid.UUID) ([]*entities.TechnicalIndicators, error) {
	// Fetch data using strategy
	response, err := strategy.FetchData(ctx, s.client, symbol, interval, timePeriod, seriesType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s data: %w", strategy.GetIndicatorName(), err)
	} // Convert to entities using strategy
	indicators, err := strategy.ConvertToEntity(ctx, s.adapter, response, symbol, companyID, timePeriod, interval)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s response: %w", strategy.GetIndicatorName(), err)
	}

	// Check if we have data to save
	if len(indicators) == 0 {
		s.logger.Warn(ctx, "No technical indicators data to save",
			logger.String("symbol", symbol),
			logger.String("indicator", strategy.GetIndicatorName()))
		return indicators, nil
	}

	// Test database connectivity with a single record before attempting bulk save
	testErr := s.repository.Create(ctx, indicators[0])
	if testErr != nil {
		// Check if error indicates table doesn't exist
		if strings.Contains(strings.ToLower(testErr.Error()), "does not exist") ||
			strings.Contains(strings.ToLower(testErr.Error()), "relation") {
			// Table doesn't exist - log single warning and return data without persistence
			s.logger.Warn(ctx, "Database table does not exist, skipping all persistence operations and returning data",
				logger.String("symbol", symbol),
				logger.String("indicator", strategy.GetIndicatorName()),
				logger.String("table_error", testErr.Error()),
				logger.Int("data_count", len(indicators)))
			return indicators, nil
		} else {
			// Other type of error - log but still return data
			s.logger.Error(ctx, "Failed to save technical indicator (non-table error)", testErr,
				logger.String("symbol", symbol),
				logger.String("indicator", strategy.GetIndicatorName()))
			return indicators, nil
		}
	}

	// First save was successful, continue with the rest
	savedCount := 1
	for i := 1; i < len(indicators); i++ {
		if err := s.repository.Create(ctx, indicators[i]); err != nil {
			s.logger.Error(ctx, "Failed to save technical indicator", err,
				logger.String("symbol", symbol),
				logger.String("indicator", strategy.GetIndicatorName()),
				logger.Int("record_index", i))
			continue
		}
		savedCount++
	}

	s.logger.Info(ctx, "Successfully saved technical indicators to database",
		logger.String("symbol", symbol),
		logger.String("indicator", strategy.GetIndicatorName()),
		logger.Int("saved_count", savedCount),
		logger.Int("total_count", len(indicators)))

	return indicators, nil
}

// AddStrategy allows adding new technical indicator strategies (Open/Closed Principle)
func (s *AlphaVantageTechnicalIndicatorsService) AddStrategy(strategy TechnicalIndicatorStrategy) {
	s.strategies = append(s.strategies, strategy)
}

// GetTechnicalIndicatorFromAPI fetches a specific technical indicator with parameters
func (s *AlphaVantageTechnicalIndicatorsService) GetTechnicalIndicatorFromAPI(ctx context.Context, symbol, indicator, interval, timePeriod, seriesType string, companyID uuid.UUID) ([]*entities.TechnicalIndicators, error) {
	s.logger.Info(ctx, "Fetching specific technical indicator from Alpha Vantage API",
		logger.String("symbol", symbol),
		logger.String("indicator", indicator),
		logger.String("interval", interval),
		logger.String("time_period", timePeriod),
		logger.String("series_type", seriesType))

	// Find the strategy for the requested indicator
	var targetStrategy TechnicalIndicatorStrategy
	for _, strategy := range s.strategies {
		if strings.EqualFold(strategy.GetIndicatorName(), indicator) {
			targetStrategy = strategy
			break
		}
	}

	if targetStrategy == nil {
		return nil, fmt.Errorf("unsupported technical indicator: %s", indicator)
	}
	// Process the specific indicator
	indicators, err := s.processIndicatorStrategy(ctx, targetStrategy, symbol, interval, timePeriod, seriesType, companyID)
	if err != nil {
		s.logger.Error(ctx, "Failed to process specific technical indicator", err,
			logger.String("symbol", symbol),
			logger.String("indicator", indicator))
		return nil, err
	}

	s.logger.Info(ctx, "Successfully fetched specific technical indicator",
		logger.String("symbol", symbol),
		logger.String("indicator", indicator),
		logger.Int("indicators_count", len(indicators)))

	return indicators, nil
}
