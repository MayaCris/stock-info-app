package alphavantage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// Adapter converts Alpha Vantage API responses to domain entities
type Adapter struct {
	logger logger.Logger
}

// NewAdapter creates a new Alpha Vantage adapter
func NewAdapter(logger logger.Logger) *Adapter {
	return &Adapter{
		logger: logger,
	}
}

// TimeSeriesDataToHistoricalData converts Alpha Vantage time series to HistoricalData entities
func (a *Adapter) TimeSeriesDataToHistoricalData(ctx context.Context, response *TimeSeriesDailyResponse, symbol string, companyID uuid.UUID) ([]*entities.HistoricalData, error) {
	if response == nil || len(response.TimeSeries) == 0 {
		return nil, fmt.Errorf("empty time series response")
	}

	var historicalData []*entities.HistoricalData

	for dateStr, data := range response.TimeSeries {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse date", err, logger.String("date", dateStr))
			continue
		}

		openPrice, err := strconv.ParseFloat(data.Open, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse open price", err, logger.String("price", data.Open))
			continue
		}

		highPrice, err := strconv.ParseFloat(data.High, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse high price", err, logger.String("price", data.High))
			continue
		}

		lowPrice, err := strconv.ParseFloat(data.Low, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse low price", err, logger.String("price", data.Low))
			continue
		}

		closePrice, err := strconv.ParseFloat(data.Close, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse close price", err, logger.String("price", data.Close))
			continue
		}

		adjustedClose, err := strconv.ParseFloat(data.AdjustedClose, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse adjusted close", err, logger.String("price", data.AdjustedClose))
			continue
		}
		volume, err := strconv.ParseInt(data.Volume, 10, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse volume", err, logger.String("volume", data.Volume))
			continue
		}

		historical := &entities.HistoricalData{
			ID:            uuid.New(),
			CompanyID:     companyID,
			Symbol:        symbol,
			Date:          date,
			OpenPrice:     openPrice,
			HighPrice:     highPrice,
			LowPrice:      lowPrice,
			ClosePrice:    closePrice,
			AdjustedClose: adjustedClose,
			Volume:        volume,
			TimeFrame:     "1D",
			DataSource:    "alphavantage",
			LastUpdated:   time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		historicalData = append(historicalData, historical)
	}

	a.logger.Info(ctx, "Converted time series to historical data",
		logger.String("symbol", symbol),
		logger.Int("dataPoints", len(historicalData)))

	return historicalData, nil
}

// CompanyOverviewToFinancialMetrics converts Alpha Vantage company overview to FinancialMetrics entity
func (a *Adapter) CompanyOverviewToFinancialMetrics(ctx context.Context, overview *CompanyOverviewResponse, companyID uuid.UUID) (*entities.FinancialMetrics, error) {
	if overview == nil || overview.Symbol == "" {
		return nil, fmt.Errorf("invalid company overview response")
	} // Parse required values
	marketCap, _ := a.parseNumericString(overview.MarketCapitalization)
	ebitda, _ := a.parseNumericString(overview.EBITDA)
	eps, _ := a.parseNumericString(overview.EPS)
	peRatio, _ := a.parseNumericString(overview.PERatio)
	pegRatio, _ := a.parseNumericString(overview.PEGRatio)
	bookValue, _ := a.parseNumericString(overview.BookValue)
	dividendPerShare, _ := a.parseNumericString(overview.DividendPerShare)
	dividendYield, _ := a.parseNumericString(overview.DividendYield)
	profitMargin, _ := a.parseNumericString(overview.ProfitMargin)
	operatingMarginTTM, _ := a.parseNumericString(overview.OperatingMarginTTM)
	returnOnAssetsTTM, _ := a.parseNumericString(overview.ReturnOnAssetsTTM)
	returnOnEquityTTM, _ := a.parseNumericString(overview.ReturnOnEquityTTM)
	quarterlyEarningsGrowthYOY, _ := a.parseNumericString(overview.QuarterlyEarningsGrowthYOY)
	quarterlyRevenueGrowthYOY, _ := a.parseNumericString(overview.QuarterlyRevenueGrowthYOY)
	analystTargetPrice, _ := a.parseNumericString(overview.AnalystTargetPrice)
	priceToSalesRatioTTM, _ := a.parseNumericString(overview.PriceToSalesRatioTTM)
	priceToBookRatio, _ := a.parseNumericString(overview.PriceToBookRatio)
	evToRevenue, _ := a.parseNumericString(overview.EVToRevenue)
	evToEBITDA, _ := a.parseNumericString(overview.EVToEBITDA)
	// Calculate enterprise value
	enterpriseValue := int64(marketCap)
	if ebitda > 0 {
		enterpriseValue = int64(marketCap + ebitda)
	}

	financialMetrics := &entities.FinancialMetrics{
		ID:                 uuid.New(),
		CompanyID:          companyID,
		Symbol:             overview.Symbol,
		PERatio:            peRatio,
		PEGRatio:           pegRatio,
		PriceToBook:        priceToBookRatio,
		PriceToSales:       priceToSalesRatioTTM,
		EVToRevenue:        evToRevenue,
		EVToEBITDA:         evToEBITDA,
		EnterpriseValue:    enterpriseValue,
		ROE:                returnOnEquityTTM,
		ROA:                returnOnAssetsTTM,
		GrossMargin:        profitMargin,
		OperatingMargin:    operatingMarginTTM,
		NetMargin:          profitMargin,
		EPS:                eps,
		EPSGrowthTTM:       quarterlyEarningsGrowthYOY,
		DividendPerShare:   dividendPerShare,
		DividendYield:      dividendYield,
		AnalystTargetPrice: analystTargetPrice,
		RevenueGrowthTTM:   quarterlyRevenueGrowthYOY,
		EarningsGrowthTTM:  quarterlyEarningsGrowthYOY,
		BookValuePerShare:  bookValue,
		DataSource:         "alphavantage",
		LastUpdated:        time.Now(),
		Currency:           "USD",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	a.logger.Info(ctx, "Converted company overview to financial metrics",
		logger.String("symbol", overview.Symbol),
		logger.String("company", overview.Name))

	return financialMetrics, nil
}

// RSIResponseToTechnicalIndicators converts RSI response to TechnicalIndicators entities
func (a *Adapter) RSIResponseToTechnicalIndicators(ctx context.Context, response *RSIResponse, symbol string, companyID uuid.UUID) ([]*entities.TechnicalIndicators, error) {
	if response == nil || len(response.RSI) == 0 {
		return nil, fmt.Errorf("empty RSI response")
	}

	var indicators []*entities.TechnicalIndicators

	// For each date, create a TechnicalIndicators record
	// Since we only have RSI data, we'll create one record per date with only RSI filled
	for dateStr, rsiValue := range response.RSI {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse date", err, logger.String("date", dateStr))
			continue
		}

		rsi, err := strconv.ParseFloat(rsiValue.RSI, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse RSI value", err, logger.String("value", rsiValue.RSI))
			continue
		}

		indicator := &entities.TechnicalIndicators{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Symbol:      symbol,
			RSI:         rsi,
			TimeFrame:   "1D", // Default timeframe based on interval
			Period:      14,   // Default RSI period
			DataSource:  "alphavantage",
			MarketDate:  date,
			LastUpdated: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Set timeframe based on interval
		if response.MetaData.Interval != "" {
			indicator.TimeFrame = response.MetaData.Interval
		}

		indicators = append(indicators, indicator)
	}

	a.logger.Info(ctx, "Converted RSI response to technical indicators",
		logger.String("symbol", symbol),
		logger.Int("dataPoints", len(indicators)))

	return indicators, nil
}

// MACDResponseToTechnicalIndicators converts MACD response to TechnicalIndicators entities
func (a *Adapter) MACDResponseToTechnicalIndicators(ctx context.Context, response *MACDResponse, symbol string, companyID uuid.UUID) ([]*entities.TechnicalIndicators, error) {
	if response == nil || len(response.MACD) == 0 {
		return nil, fmt.Errorf("empty MACD response")
	}

	var indicators []*entities.TechnicalIndicators

	for dateStr, macdValues := range response.MACD {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse date", err, logger.String("date", dateStr))
			continue
		}

		macd, err := strconv.ParseFloat(macdValues.MACD, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse MACD value", err, logger.String("value", macdValues.MACD))
			continue
		}

		macdSignal, err := strconv.ParseFloat(macdValues.MACDSignal, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse MACD Signal value", err, logger.String("value", macdValues.MACDSignal))
			continue
		}

		macdHist, err := strconv.ParseFloat(macdValues.MACDHist, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse MACD Histogram value", err, logger.String("value", macdValues.MACDHist))
			continue
		}

		// Create one indicator with all MACD components
		indicator := &entities.TechnicalIndicators{
			ID:            uuid.New(),
			CompanyID:     companyID,
			Symbol:        symbol,
			MACD:          macd,
			MACDSignal:    macdSignal,
			MACDHistogram: macdHist,
			TimeFrame:     "1D", // Default timeframe
			Period:        14,   // Default period
			DataSource:    "alphavantage",
			MarketDate:    date,
			LastUpdated:   time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Set timeframe based on interval
		if response.MetaData.Interval != "" {
			indicator.TimeFrame = response.MetaData.Interval
		}

		indicators = append(indicators, indicator)
	}

	a.logger.Info(ctx, "Converted MACD response to technical indicators",
		logger.String("symbol", symbol),
		logger.Int("dataPoints", len(indicators)))

	return indicators, nil
}

// SMAResponseToTechnicalIndicators converts SMA response to TechnicalIndicators entities
func (a *Adapter) SMAResponseToTechnicalIndicators(ctx context.Context, response *SMAResponse, symbol string, companyID uuid.UUID, period int) ([]*entities.TechnicalIndicators, error) {
	if response == nil || len(response.SMA) == 0 {
		return nil, fmt.Errorf("empty SMA response")
	}

	var indicators []*entities.TechnicalIndicators

	for dateStr, smaValue := range response.SMA {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse date", err, logger.String("date", dateStr))
			continue
		}

		sma, err := strconv.ParseFloat(smaValue.SMA, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse SMA value", err, logger.String("value", smaValue.SMA))
			continue
		}

		indicator := &entities.TechnicalIndicators{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Symbol:      symbol,
			TimeFrame:   "1D", // Default timeframe
			Period:      int32(period),
			DataSource:  "alphavantage",
			MarketDate:  date,
			LastUpdated: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Set the appropriate SMA field based on period
		switch period {
		case 20:
			indicator.SMA20 = sma
		case 50:
			indicator.SMA50 = sma
		case 200:
			indicator.SMA200 = sma
		default:
			// For other periods, we could use SMA20 as a generic field
			indicator.SMA20 = sma
		}

		// Set timeframe based on interval
		if response.MetaData.Interval != "" {
			indicator.TimeFrame = response.MetaData.Interval
		}

		indicators = append(indicators, indicator)
	}

	a.logger.Info(ctx, "Converted SMA response to technical indicators",
		logger.String("symbol", symbol),
		logger.Int("period", period),
		logger.Int("dataPoints", len(indicators)))

	return indicators, nil
}

// EMAResponseToTechnicalIndicators converts EMA response to TechnicalIndicators entities
func (a *Adapter) EMAResponseToTechnicalIndicators(ctx context.Context, response *EMAResponse, symbol string, companyID uuid.UUID, period int) ([]*entities.TechnicalIndicators, error) {
	if response == nil || len(response.EMA) == 0 {
		return nil, fmt.Errorf("empty EMA response")
	}

	var indicators []*entities.TechnicalIndicators

	for dateStr, emaValue := range response.EMA {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse date", err, logger.String("date", dateStr))
			continue
		}

		ema, err := strconv.ParseFloat(emaValue.EMA, 64)
		if err != nil {
			a.logger.Error(ctx, "Failed to parse EMA value", err, logger.String("value", emaValue.EMA))
			continue
		}

		indicator := &entities.TechnicalIndicators{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Symbol:      symbol,
			TimeFrame:   "1D", // Default timeframe
			Period:      int32(period),
			DataSource:  "alphavantage",
			MarketDate:  date,
			LastUpdated: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Set the appropriate EMA field based on period
		switch period {
		case 12:
			indicator.EMA12 = ema
		case 26:
			indicator.EMA26 = ema
		default:
			// For other periods, we could use EMA12 as a generic field
			indicator.EMA12 = ema
		}

		// Set timeframe based on interval
		if response.MetaData.Interval != "" {
			indicator.TimeFrame = response.MetaData.Interval
		}

		indicators = append(indicators, indicator)
	}

	a.logger.Info(ctx, "Converted EMA response to technical indicators",
		logger.String("symbol", symbol),
		logger.Int("period", period),
		logger.Int("dataPoints", len(indicators)))

	return indicators, nil
}

// ValidateHistoricalData validates historical data before saving
func (a *Adapter) ValidateHistoricalData(data *entities.HistoricalData) error {
	if data.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if data.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}

	if data.ClosePrice <= 0 {
		return fmt.Errorf("close price must be positive")
	}

	if data.Volume < 0 {
		return fmt.Errorf("volume cannot be negative")
	}

	if data.HighPrice < data.LowPrice {
		return fmt.Errorf("high price cannot be less than low price")
	}

	if data.OpenPrice <= 0 || data.HighPrice <= 0 || data.LowPrice <= 0 {
		return fmt.Errorf("all prices must be positive")
	}

	return nil
}

// ValidateFinancialMetrics validates financial metrics before saving
func (a *Adapter) ValidateFinancialMetrics(metrics *entities.FinancialMetrics) error {
	if metrics.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if metrics.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}
	// Check for reasonable PE ratio range
	if metrics.PERatio < 0 || metrics.PERatio > 1000 {
		a.logger.Warn(context.Background(), "Unusual PE ratio detected",
			logger.String("symbol", metrics.Symbol),
			logger.Float64("pe_ratio", metrics.PERatio),
		)
	}

	return nil
}

// ValidateTechnicalIndicators validates technical indicators before saving
func (a *Adapter) ValidateTechnicalIndicators(indicator *entities.TechnicalIndicators) error {
	if indicator.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if indicator.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}

	// Validate RSI range if it's set
	if indicator.RSI > 0 && (indicator.RSI < 0 || indicator.RSI > 100) {
		return fmt.Errorf("RSI value must be between 0 and 100")
	}

	return nil
}

// Helper function to parse numeric strings from Alpha Vantage API
func (a *Adapter) parseNumericString(value string) (float64, error) {
	if value == "" || value == "None" || value == "-" {
		return 0, nil
	}

	// Remove any commas or other formatting
	cleanValue := strings.ReplaceAll(value, ",", "")

	parsed, err := strconv.ParseFloat(cleanValue, 64)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}
