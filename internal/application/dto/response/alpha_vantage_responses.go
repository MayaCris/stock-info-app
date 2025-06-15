package response

import (
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
)

// HistoricalDataResponse represents historical price data response
type HistoricalDataResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    *HistoricalDataPayload `json:"data"`
}

// HistoricalDataPayload contains historical data details
type HistoricalDataPayload struct {
	Symbol      string                     `json:"symbol"`
	Period      string                     `json:"period"`
	OutputSize  string                     `json:"output_size"`
	DataSource  string                     `json:"data_source"`
	LastUpdated time.Time                  `json:"last_updated"`
	TotalPoints int                        `json:"total_points"`
	Historical  []*entities.HistoricalData `json:"historical_data"`
	Summary     *HistoricalDataSummary     `json:"summary,omitempty"`
}

// HistoricalDataSummary contains summary statistics
type HistoricalDataSummary struct {
	HighestPrice  float64 `json:"highest_price"`
	LowestPrice   float64 `json:"lowest_price"`
	AveragePrice  float64 `json:"average_price"`
	PriceChange   float64 `json:"price_change"`
	PercentChange float64 `json:"percent_change"`
	TotalVolume   int64   `json:"total_volume"`
	AverageVolume int64   `json:"average_volume"`
}

// TechnicalIndicatorsResponse represents technical indicators response
type TechnicalIndicatorsResponse struct {
	Success bool                        `json:"success"`
	Message string                      `json:"message"`
	Data    *TechnicalIndicatorsPayload `json:"data"`
}

// TechnicalIndicatorsPayload contains technical indicators details
type TechnicalIndicatorsPayload struct {
	Symbol      string                          `json:"symbol"`
	Indicator   string                          `json:"indicator"`
	Interval    string                          `json:"interval"`
	TimePeriod  string                          `json:"time_period"`
	DataSource  string                          `json:"data_source"`
	LastUpdated time.Time                       `json:"last_updated"`
	TotalPoints int                             `json:"total_points"`
	Indicators  []*entities.TechnicalIndicators `json:"technical_indicators"`
	LatestValue *TechnicalIndicatorLatestValue  `json:"latest_value,omitempty"`
}

// TechnicalIndicatorLatestValue contains the latest indicator value
type TechnicalIndicatorLatestValue struct {
	Date   time.Time `json:"date"`
	Value  float64   `json:"value"`
	Signal string    `json:"signal,omitempty"` // "BUY", "SELL", "HOLD", "OVERSOLD", "OVERBOUGHT"
}

// FundamentalDataResponse represents fundamental financial data response
type FundamentalDataResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    *FundamentalDataPayload `json:"data"`
}

// FundamentalDataPayload contains fundamental data details
type FundamentalDataPayload struct {
	Symbol         string                     `json:"symbol"`
	CompanyName    string                     `json:"company_name"`
	Sector         string                     `json:"sector"`
	Industry       string                     `json:"industry"`
	DataSource     string                     `json:"data_source"`
	LastUpdated    time.Time                  `json:"last_updated"`
	Financials     *entities.FinancialMetrics `json:"financial_metrics"`
	CompanyProfile *CompanyFundamentalProfile `json:"company_profile"`
	Valuation      *ValuationMetrics          `json:"valuation"`
}

// CompanyFundamentalProfile contains company fundamental profile
type CompanyFundamentalProfile struct {
	AssetType         string `json:"asset_type"`
	Description       string `json:"description"`
	CIK               string `json:"cik"`
	Exchange          string `json:"exchange"`
	Currency          string `json:"currency"`
	Country           string `json:"country"`
	Address           string `json:"address"`
	FiscalYearEnd     string `json:"fiscal_year_end"`
	LatestQuarter     string `json:"latest_quarter"`
	SharesOutstanding int64  `json:"shares_outstanding"`
}

// ValuationMetrics contains valuation metrics
type ValuationMetrics struct {
	MarketCap         int64   `json:"market_cap"`
	EnterpriseValue   int64   `json:"enterprise_value"`
	Week52High        float64 `json:"week_52_high"`
	Week52Low         float64 `json:"week_52_low"`
	MovingAverage50   float64 `json:"moving_average_50"`
	MovingAverage200  float64 `json:"moving_average_200"`
	Beta              float64 `json:"beta"`
	EVToRevenue       float64 `json:"ev_to_revenue"`
	EVToEBITDA        float64 `json:"ev_to_ebitda"`
	PriceToSalesRatio float64 `json:"price_to_sales_ratio"`
	PriceToBookRatio  float64 `json:"price_to_book_ratio"`
}

// EarningsDataResponse represents earnings data response
type EarningsDataResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    *EarningsDataPayload `json:"data"`
}

// EarningsDataPayload contains earnings data details
type EarningsDataPayload struct {
	Symbol            string              `json:"symbol"`
	DataSource        string              `json:"data_source"`
	LastUpdated       time.Time           `json:"last_updated"`
	AnnualEarnings    []*AnnualEarning    `json:"annual_earnings"`
	QuarterlyEarnings []*QuarterlyEarning `json:"quarterly_earnings"`
	EarningsTrend     *EarningsTrend      `json:"earnings_trend,omitempty"`
}

// AnnualEarning represents annual earnings data
type AnnualEarning struct {
	FiscalDateEnding string  `json:"fiscal_date_ending"`
	ReportedEPS      float64 `json:"reported_eps"`
}

// QuarterlyEarning represents quarterly earnings data
type QuarterlyEarning struct {
	FiscalDateEnding   string  `json:"fiscal_date_ending"`
	ReportedDate       string  `json:"reported_date"`
	ReportedEPS        float64 `json:"reported_eps"`
	EstimatedEPS       float64 `json:"estimated_eps"`
	Surprise           float64 `json:"surprise"`
	SurprisePercentage float64 `json:"surprise_percentage"`
}

// EarningsTrend contains earnings trend analysis
type EarningsTrend struct {
	EPSGrowthYoY        float64 `json:"eps_growth_yoy"`
	EPSGrowthQoQ        float64 `json:"eps_growth_qoq"`
	AverageEPSGrowth    float64 `json:"average_eps_growth"`
	EarningsConsistency string  `json:"earnings_consistency"` // "HIGH", "MEDIUM", "LOW"
	NextEarningsDate    string  `json:"next_earnings_date,omitempty"`
}

// AlphaVantageHealthResponse represents health check response
type AlphaVantageHealthResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    *HealthCheckPayload `json:"data"`
}

// HealthCheckPayload contains health check details
type HealthCheckPayload struct {
	Service      string    `json:"service"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime string    `json:"response_time,omitempty"`
}

// Helper functions to create responses

// NewHistoricalDataResponse creates a new historical data response
func NewHistoricalDataResponse(data []*entities.HistoricalData, symbol, period, outputSize string) *HistoricalDataResponse {
	summary := calculateHistoricalSummary(data)

	return &HistoricalDataResponse{
		Success: true,
		Message: "Historical data retrieved successfully",
		Data: &HistoricalDataPayload{
			Symbol:      symbol,
			Period:      period,
			OutputSize:  outputSize,
			DataSource:  "alphavantage",
			LastUpdated: time.Now(),
			TotalPoints: len(data),
			Historical:  data,
			Summary:     summary,
		},
	}
}

// NewTechnicalIndicatorsResponse creates a new technical indicators response
func NewTechnicalIndicatorsResponse(indicators []*entities.TechnicalIndicators, symbol, indicator, interval, timePeriod string) *TechnicalIndicatorsResponse {
	var latestValue *TechnicalIndicatorLatestValue
	if len(indicators) > 0 {
		latest := indicators[0]
		latestValue = &TechnicalIndicatorLatestValue{
			Date:  latest.MarketDate,
			Value: latest.RSI, // This would be dynamic based on indicator type
		}
	}

	return &TechnicalIndicatorsResponse{
		Success: true,
		Message: "Technical indicators retrieved successfully",
		Data: &TechnicalIndicatorsPayload{
			Symbol:      symbol,
			Indicator:   indicator,
			Interval:    interval,
			TimePeriod:  timePeriod,
			DataSource:  "alphavantage",
			LastUpdated: time.Now(),
			TotalPoints: len(indicators),
			Indicators:  indicators,
			LatestValue: latestValue,
		},
	}
}

// NewFundamentalDataResponse creates a new fundamental data response
func NewFundamentalDataResponse(financials *entities.FinancialMetrics, companyName, sector, industry string) *FundamentalDataResponse {
	// Create valuation metrics from financial data
	valuation := &ValuationMetrics{
		EnterpriseValue:   financials.EnterpriseValue,
		EVToRevenue:       financials.EVToRevenue,
		EVToEBITDA:        financials.EVToEBITDA,
		PriceToSalesRatio: financials.PriceToSales,
		PriceToBookRatio:  financials.PriceToBook,
	}

	return &FundamentalDataResponse{
		Success: true,
		Message: "Fundamental data retrieved successfully",
		Data: &FundamentalDataPayload{
			Symbol:      financials.Symbol,
			CompanyName: companyName,
			Sector:      sector,
			Industry:    industry,
			DataSource:  "alphavantage",
			LastUpdated: time.Now(),
			Financials:  financials,
			Valuation:   valuation,
		},
	}
}

// calculateHistoricalSummary calculates summary statistics for historical data
func calculateHistoricalSummary(data []*entities.HistoricalData) *HistoricalDataSummary {
	if len(data) == 0 {
		return nil
	}

	var highestPrice, lowestPrice, totalPrice float64
	var totalVolume int64

	highestPrice = data[0].HighPrice
	lowestPrice = data[0].LowPrice

	for _, d := range data {
		if d.HighPrice > highestPrice {
			highestPrice = d.HighPrice
		}
		if d.LowPrice < lowestPrice {
			lowestPrice = d.LowPrice
		}
		totalPrice += d.ClosePrice
		totalVolume += d.Volume
	}

	averagePrice := totalPrice / float64(len(data))
	averageVolume := totalVolume / int64(len(data))

	// Calculate price change (first to last)
	priceChange := data[len(data)-1].ClosePrice - data[0].ClosePrice
	percentChange := (priceChange / data[0].ClosePrice) * 100

	return &HistoricalDataSummary{
		HighestPrice:  highestPrice,
		LowestPrice:   lowestPrice,
		AveragePrice:  averagePrice,
		PriceChange:   priceChange,
		PercentChange: percentChange,
		TotalVolume:   totalVolume,
		AverageVolume: averageVolume,
	}
}
