package response

import (
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
)

// FinancialAnalysisResponse represents comprehensive financial analysis response
type FinancialAnalysisResponse struct {
	Symbol           string   `json:"symbol"`
	FinancialScore   float64  `json:"financial_score"`
	StockType        string   `json:"stock_type"`
	AnalystConsensus string   `json:"analyst_consensus"`
	Insights         []string `json:"insights"`

	// Valuation Metrics
	PERatio      float64 `json:"pe_ratio"`
	PEGRatio     float64 `json:"peg_ratio"`
	PriceToBook  float64 `json:"price_to_book"`
	PriceToSales float64 `json:"price_to_sales"`

	// Profitability
	ROE       float64 `json:"roe"`
	ROA       float64 `json:"roa"`
	NetMargin float64 `json:"net_margin"`

	// Financial Health
	DebtToEquity float64 `json:"debt_to_equity"`
	CurrentRatio float64 `json:"current_ratio"`

	// Growth
	RevenueGrowthTTM  float64 `json:"revenue_growth_ttm"`
	EarningsGrowthTTM float64 `json:"earnings_growth_ttm"`

	LastUpdated time.Time `json:"last_updated"`
}

// SectorAnalysisResponse represents sector analysis response
type SectorAnalysisResponse struct {
	Sector      string                       `json:"sector"`
	TotalStocks int                          `json:"total_stocks"`
	Averages    map[string]float64           `json:"averages"`
	TopStocks   []*entities.FinancialMetrics `json:"top_stocks"`
}

// StockScreenCriteria represents criteria for stock screening
type StockScreenCriteria struct {
	MaxPE            float64 `json:"max_pe"`
	MinROE           float64 `json:"min_roe"`
	MinGrowth        float64 `json:"min_growth"`
	MaxDebtToEquity  float64 `json:"max_debt_to_equity"`
	MinDividendYield float64 `json:"min_dividend_yield"`
	Sector           string  `json:"sector"`
	Industry         string  `json:"industry"`
}

// TechnicalAnalysisResponse represents comprehensive technical analysis response
type TechnicalAnalysisResponse struct {
	Symbol         string            `json:"symbol"`
	TechnicalScore float64           `json:"technical_score"`
	Signals        map[string]string `json:"signals"`
	Insights       []string          `json:"insights"`

	// Key Indicators
	RSI        float64 `json:"rsi"`
	MACD       float64 `json:"macd"`
	MACDSignal float64 `json:"macd_signal"`

	// Moving Averages
	SMA20  float64 `json:"sma_20"`
	SMA50  float64 `json:"sma_50"`
	SMA200 float64 `json:"sma_200"`

	// Bollinger Bands
	BBUpper    float64 `json:"bb_upper"`
	BBMiddle   float64 `json:"bb_middle"`
	BBLower    float64 `json:"bb_lower"`
	BBPercentB float64 `json:"bb_percent_b"`

	// Volume
	Volume     int64 `json:"volume"`
	VolumeMA20 int64 `json:"volume_ma_20"`
	OBV        int64 `json:"obv"`

	// Support/Resistance
	Support1    float64 `json:"support_1"`
	Support2    float64 `json:"support_2"`
	Resistance1 float64 `json:"resistance_1"`
	Resistance2 float64 `json:"resistance_2"`

	// Volatility
	ATR       float64 `json:"atr"`
	BandWidth float64 `json:"band_width"`

	LastUpdated time.Time `json:"last_updated"`
}

// StockScreeningResult represents the result of stock screening
type StockScreeningResult struct {
	TotalMatched int                          `json:"total_matched"`
	Criteria     StockScreenCriteria          `json:"criteria"`
	Stocks       []*entities.FinancialMetrics `json:"stocks"`
	GeneratedAt  time.Time                    `json:"generated_at"`
}
