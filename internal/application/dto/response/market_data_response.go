package response

import (
	"time"

	"github.com/google/uuid"
)

// MarketDataResponse represents real-time market data response
type MarketDataResponse struct {
	ID        uuid.UUID `json:"id"`
	CompanyID uuid.UUID `json:"company_id"`
	Symbol    string    `json:"symbol"`

	// Price Information
	CurrentPrice  float64 `json:"current_price"`
	OpenPrice     float64 `json:"open_price"`
	HighPrice     float64 `json:"high_price"`
	LowPrice      float64 `json:"low_price"`
	PreviousClose float64 `json:"previous_close"`

	// Change Information
	PriceChange     float64 `json:"price_change"`
	PriceChangePerc float64 `json:"price_change_perc"`

	// Volume and Trading
	Volume    int64 `json:"volume"`
	AvgVolume int64 `json:"avg_volume"`
	MarketCap int64 `json:"market_cap"`

	// Market Status
	IsMarketOpen bool   `json:"is_market_open"`
	Currency     string `json:"currency"`
	Exchange     string `json:"exchange"`

	// Timestamps
	MarketTimestamp time.Time `json:"market_timestamp"`
	LastUpdated     time.Time `json:"last_updated"`
}

// CompanyProfileResponse represents detailed company information
type CompanyProfileResponse struct {
	ID          uuid.UUID `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Industry    string    `json:"industry"`
	Sector      string    `json:"sector"`
	Country     string    `json:"country"`
	Currency    string    `json:"currency"`

	// Financial Metrics
	MarketCap         int64   `json:"market_cap"`
	SharesOutstanding int64   `json:"shares_outstanding"`
	PERatio           float64 `json:"pe_ratio"`
	PEGRatio          float64 `json:"peg_ratio"`
	PriceToBook       float64 `json:"price_to_book"`
	DividendYield     float64 `json:"dividend_yield"`
	EPS               float64 `json:"eps"`
	Beta              float64 `json:"beta"`

	// Company Details
	Website       string    `json:"website"`
	Logo          string    `json:"logo"`
	IPODate       time.Time `json:"ipo_date"`
	EmployeeCount int32     `json:"employee_count"`

	// Timestamps
	LastUpdated time.Time `json:"last_updated"`
}

// NewsResponse represents news article response
type NewsResponse struct {
	ID       uuid.UUID `json:"id"`
	Symbol   string    `json:"symbol"`
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	URL      string    `json:"url"`
	ImageURL string    `json:"image_url"`
	Source   string    `json:"source"`
	Category string    `json:"category"`
	Language string    `json:"language"`

	// Sentiment Analysis
	SentimentScore float64 `json:"sentiment_score"`
	SentimentLabel string  `json:"sentiment_label"`

	// Timestamps
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// BasicFinancialsResponse represents basic financial metrics
type BasicFinancialsResponse struct {
	ID     uuid.UUID `json:"id"`
	Symbol string    `json:"symbol"`

	// Valuation Metrics
	PERatio         float64 `json:"pe_ratio"`
	PEGRatio        float64 `json:"peg_ratio"`
	PriceToSales    float64 `json:"price_to_sales"`
	PriceToBook     float64 `json:"price_to_book"`
	PriceToCashFlow float64 `json:"price_to_cash_flow"`

	// Profitability Metrics
	ROE             float64 `json:"roe"`
	ROA             float64 `json:"roa"`
	ROI             float64 `json:"roi"`
	GrossMargin     float64 `json:"gross_margin"`
	OperatingMargin float64 `json:"operating_margin"`
	NetMargin       float64 `json:"net_margin"`

	// Growth Metrics
	RevenueGrowth  float64 `json:"revenue_growth"`
	EarningsGrowth float64 `json:"earnings_growth"`
	DividendGrowth float64 `json:"dividend_growth"`

	// Financial Health
	DebtToEquity float64 `json:"debt_to_equity"`
	CurrentRatio float64 `json:"current_ratio"`
	QuickRatio   float64 `json:"quick_ratio"`

	// Per Share Metrics
	EPS               float64 `json:"eps"`
	BookValuePerShare float64 `json:"book_value_per_share"`
	CashPerShare      float64 `json:"cash_per_share"`
	DividendPerShare  float64 `json:"dividend_per_share"`

	// Period Information
	Period        string `json:"period"`
	FiscalYear    int    `json:"fiscal_year"`
	FiscalQuarter int    `json:"fiscal_quarter"`

	// Timestamps
	LastUpdated time.Time `json:"last_updated"`
}

// MarketOverviewResponse represents market overview statistics
type MarketOverviewResponse struct {
	TotalStocks    int       `json:"total_stocks"`
	TotalGainers   int       `json:"total_gainers"`
	TotalLosers    int       `json:"total_losers"`
	AvgPriceChange float64   `json:"avg_price_change"`
	TotalVolume    int64     `json:"total_volume"`
	LastUpdated    time.Time `json:"last_updated"`
}

// MarketDataSummaryResponse represents aggregated market data
type MarketDataSummaryResponse struct {
	Symbol          string    `json:"symbol"`
	CompanyName     string    `json:"company_name"`
	CurrentPrice    float64   `json:"current_price"`
	PriceChange     float64   `json:"price_change"`
	PriceChangePerc float64   `json:"price_change_perc"`
	Volume          int64     `json:"volume"`
	MarketCap       int64     `json:"market_cap"`
	LastUpdated     time.Time `json:"last_updated"`
}

// NewsListResponse represents a list of news with pagination
type NewsListResponse struct {
	News       []*NewsResponse `json:"news"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	HasMore    bool            `json:"has_more"`
}

// SentimentAnalysisResponse represents sentiment analysis results
type SentimentAnalysisResponse struct {
	Symbol         string  `json:"symbol"`
	PositiveCount  int     `json:"positive_count"`
	NegativeCount  int     `json:"negative_count"`
	NeutralCount   int     `json:"neutral_count"`
	TotalCount     int     `json:"total_count"`
	AvgSentiment   float64 `json:"avg_sentiment"`
	SentimentTrend string  `json:"sentiment_trend"` // "positive", "negative", "neutral"
}

// FinancialRatiosResponse represents key financial ratios
type FinancialRatiosResponse struct {
	Symbol          string                `json:"symbol"`
	Valuation       ValuationRatios       `json:"valuation"`
	Profitability   ProfitabilityRatios   `json:"profitability"`
	Growth          GrowthRatios          `json:"growth"`
	FinancialHealth FinancialHealthRatios `json:"financial_health"`
}

// ValuationRatios represents valuation metrics
type ValuationRatios struct {
	PERatio         float64 `json:"pe_ratio"`
	PEGRatio        float64 `json:"peg_ratio"`
	PriceToSales    float64 `json:"price_to_sales"`
	PriceToBook     float64 `json:"price_to_book"`
	PriceToCashFlow float64 `json:"price_to_cash_flow"`
}

// ProfitabilityRatios represents profitability metrics
type ProfitabilityRatios struct {
	ROE             float64 `json:"roe"`
	ROA             float64 `json:"roa"`
	ROI             float64 `json:"roi"`
	GrossMargin     float64 `json:"gross_margin"`
	OperatingMargin float64 `json:"operating_margin"`
	NetMargin       float64 `json:"net_margin"`
}

// GrowthRatios represents growth metrics
type GrowthRatios struct {
	RevenueGrowth  float64 `json:"revenue_growth"`
	EarningsGrowth float64 `json:"earnings_growth"`
	DividendGrowth float64 `json:"dividend_growth"`
}

// FinancialHealthRatios represents financial health metrics
type FinancialHealthRatios struct {
	DebtToEquity float64 `json:"debt_to_equity"`
	CurrentRatio float64 `json:"current_ratio"`
	QuickRatio   float64 `json:"quick_ratio"`
}

// MarketDataHealthResponse represents API health status
type MarketDataHealthResponse struct {
	Status       string            `json:"status"` // "healthy", "degraded", "unhealthy"
	LastUpdate   time.Time         `json:"last_update"`
	DataSources  map[string]string `json:"data_sources"` // source -> status
	ResponseTime time.Duration     `json:"response_time"`
	Message      string            `json:"message,omitempty"`
}
