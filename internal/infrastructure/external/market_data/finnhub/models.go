package finnhub

import (
	"encoding/json"
	"time"
)

// QuoteResponse represents the real-time quote response from Finnhub
type QuoteResponse struct {
	CurrentPrice  float64 `json:"c"`  // Current price
	Change        float64 `json:"d"`  // Change
	PercentChange float64 `json:"dp"` // Percent change
	HighPrice     float64 `json:"h"`  // High price of the day
	LowPrice      float64 `json:"l"`  // Low price of the day
	OpenPrice     float64 `json:"o"`  // Open price of the day
	PreviousClose float64 `json:"pc"` // Previous close price
	Timestamp     int64   `json:"t"`  // Timestamp
}

// GetTimestamp converts Unix timestamp to time.Time
func (q *QuoteResponse) GetTimestamp() time.Time {
	return time.Unix(q.Timestamp, 0)
}

// IsValid checks if the quote response has valid data
func (q *QuoteResponse) IsValid() bool {
	return q.CurrentPrice > 0 && q.Timestamp > 0
}

// CompanyProfileResponse represents company profile from Finnhub
type CompanyProfileResponse struct {
	Country          string  `json:"country"`
	Currency         string  `json:"currency"`
	Exchange         string  `json:"exchange"`
	Industry         string  `json:"finnhubIndustry"`
	IPO              string  `json:"ipo"`
	Logo             string  `json:"logo"`
	MarketCap        float64 `json:"marketCapitalization"`
	Name             string  `json:"name"`
	Phone            string  `json:"phone"`
	ShareOutstanding float64 `json:"shareOutstanding"`
	Ticker           string  `json:"ticker"`
	Website          string  `json:"weburl"`
}

// GetIPODate parses IPO date string to time.Time
func (cp *CompanyProfileResponse) GetIPODate() (time.Time, error) {
	if cp.IPO == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", cp.IPO)
}

// GetMarketCapInBillions returns market cap in billions
func (cp *CompanyProfileResponse) GetMarketCapInBillions() float64 {
	return cp.MarketCap / 1000 // Finnhub returns in millions
}

// IsValid checks if company profile has required fields
func (cp *CompanyProfileResponse) IsValid() bool {
	return cp.Name != "" && cp.Ticker != ""
}

// NewsResponse represents news response from Finnhub
type NewsResponse []NewsItemResponse

// NewsItemResponse represents a single news item
type NewsItemResponse struct {
	Category string `json:"category"`
	DateTime int64  `json:"datetime"`
	Headline string `json:"headline"`
	ID       int64  `json:"id"`
	Image    string `json:"image"`
	Related  string `json:"related"`
	Source   string `json:"source"`
	Summary  string `json:"summary"`
	URL      string `json:"url"`
}

// GetPublishedTime converts Unix timestamp to time.Time
func (ni *NewsItemResponse) GetPublishedTime() time.Time {
	return time.Unix(ni.DateTime, 0)
}

// IsValid checks if news item has required fields
func (ni *NewsItemResponse) IsValid() bool {
	return ni.Headline != "" && ni.URL != "" && ni.DateTime > 0
}

// BasicFinancialsResponse represents basic financials from Finnhub
type BasicFinancialsResponse struct {
	Series struct {
		Annual    BasicFinancialsData `json:"annual"`
		Quarterly BasicFinancialsData `json:"quarterly"`
	} `json:"series"`
	Metric     BasicFinancialsMetric `json:"metric"`
	MetricType string                `json:"metricType"`
	Symbol     string                `json:"symbol"`
}

// BasicFinancialsData contains time series financial data
type BasicFinancialsData struct {
	CurrentRatio       FinancialDataSeries `json:"currentRatio"`
	SalesPerShare      FinancialDataSeries `json:"salesPerShare"`
	NetMargin          FinancialDataSeries `json:"netMargin"`
	ROE                FinancialDataSeries `json:"roe"`
	ROA                FinancialDataSeries `json:"roa"`
	DebtEquity         FinancialDataSeries `json:"totalDebt/totalEquity"`
	RevenueGrowth      FinancialDataSeries `json:"revenueGrowthTTM"`
	EpsGrowth          FinancialDataSeries `json:"epsGrowthTTM"`
	OperatingMargin    FinancialDataSeries `json:"operatingMargin"`
	PretaxMargin       FinancialDataSeries `json:"pretaxMargin"`
	SalesGrowth        FinancialDataSeries `json:"salesGrowth"`
	GrossMargin        FinancialDataSeries `json:"grossMargin"`
	DividendGrowthRate FinancialDataSeries `json:"dividendGrowthRate5Y"`
}

// BasicFinancialsMetric contains current metric values
type BasicFinancialsMetric struct {
	// Valuation Metrics
	PE  float64 `json:"peBasicExclExtraTTM"`
	PEG float64 `json:"pegRatio"`
	PB  float64 `json:"pbQuarterly"`
	PS  float64 `json:"psQuarterly"`
	PCF float64 `json:"pcfShareTTM"`

	// Profitability
	ROE             float64 `json:"roeRfy"`
	ROA             float64 `json:"roaRfy"`
	ROI             float64 `json:"roiTTM"`
	GrossMargin     float64 `json:"grossMarginTTM"`
	OperatingMargin float64 `json:"operatingMarginTTM"`
	NetMargin       float64 `json:"netMarginTTM"`

	// Growth
	RevenueGrowthTTM float64 `json:"revenueGrowthTTM"`
	EpsGrowthTTM     float64 `json:"epsGrowthTTM"`

	// Financial Health
	DebtEquity   float64 `json:"totalDebt/totalEquityQuarterly"`
	CurrentRatio float64 `json:"currentRatioQuarterly"`
	QuickRatio   float64 `json:"quickRatioQuarterly"`

	// Per Share Data
	BookValue        float64 `json:"bookValuePerShareQuarterly"`
	CashPerShare     float64 `json:"cashPerSharePerShareTTM"`
	DividendPerShare float64 `json:"dividendPerShare"`

	// Other Metrics
	Beta           float64 `json:"beta"`
	Week52High     float64 `json:"52WeekHigh"`
	Week52Low      float64 `json:"52WeekLow"`
	Week52HighDate string  `json:"52WeekHighDate"`
	Week52LowDate  string  `json:"52WeekLowDate"`
	DividendYield  float64 `json:"dividendYieldIndicatedAnnual"`
}

// FinancialDataPoint represents a time-series data point
type FinancialDataPoint struct {
	Period string  `json:"period"`
	V      float64 `json:"v"`
}

// FinancialDataSeries represents a series of financial data points
type FinancialDataSeries []FinancialDataPoint

// GetLatestValue returns the most recent value from financial data series
func (data FinancialDataSeries) GetLatestValue() float64 {
	if len(data) == 0 {
		return 0
	}
	return data[len(data)-1].V
}

// IsValid checks if basic financials response is valid
func (bf *BasicFinancialsResponse) IsValid() bool {
	return bf.Symbol != ""
}

// MarketNewsResponse represents market news response
type MarketNewsResponse []MarketNewsItem

// MarketNewsItem represents a market news item
type MarketNewsItem struct {
	Category string `json:"category"`
	DateTime int64  `json:"datetime"`
	Headline string `json:"headline"`
	ID       int64  `json:"id"`
	Image    string `json:"image"`
	Related  string `json:"related"`
	Source   string `json:"source"`
	Summary  string `json:"summary"`
	URL      string `json:"url"`
}

// GetPublishedTime converts Unix timestamp to time.Time
func (mn *MarketNewsItem) GetPublishedTime() time.Time {
	return time.Unix(mn.DateTime, 0)
}

// StockSymbolsResponse represents stock symbols response
type StockSymbolsResponse []StockSymbol

// StockSymbol represents a stock symbol
type StockSymbol struct {
	Currency       string `json:"currency"`
	Description    string `json:"description"`
	DisplaySymbol  string `json:"displaySymbol"`
	FIGI           string `json:"figi"`
	ISIN           string `json:"isin"`
	MIC            string `json:"mic"`
	ShareClassFIGI string `json:"shareClassFIGI"`
	Symbol         string `json:"symbol"`
	Symbol2        string `json:"symbol2"`
	Type           string `json:"type"`
}

// ErrorResponse represents an error response from Finnhub
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RecommendationTrendsResponse represents recommendation trends
type RecommendationTrendsResponse []RecommendationTrend

// RecommendationTrend represents recommendation trend data
type RecommendationTrend struct {
	Buy        int    `json:"buy"`
	Hold       int    `json:"hold"`
	Period     string `json:"period"`
	Sell       int    `json:"sell"`
	StrongBuy  int    `json:"strongBuy"`
	StrongSell int    `json:"strongSell"`
	Symbol     string `json:"symbol"`
}

// GetTotalRecommendations returns total number of recommendations
func (rt *RecommendationTrend) GetTotalRecommendations() int {
	return rt.Buy + rt.Hold + rt.Sell + rt.StrongBuy + rt.StrongSell
}

// GetBuyRatio returns ratio of buy recommendations
func (rt *RecommendationTrend) GetBuyRatio() float64 {
	total := rt.GetTotalRecommendations()
	if total == 0 {
		return 0
	}
	return float64(rt.Buy+rt.StrongBuy) / float64(total)
}

// EarningsResponse represents earnings data response
type EarningsResponse []EarningsData

// EarningsData represents earnings data for a period
type EarningsData struct {
	Actual          *float64 `json:"actual"`
	Estimate        *float64 `json:"estimate"`
	Period          string   `json:"period"`
	Symbol          string   `json:"symbol"`
	Surprise        *float64 `json:"surprise"`
	SurprisePercent *float64 `json:"surprisePercent"`
}

// HasPositiveSurprise checks if earnings had positive surprise
func (e *EarningsData) HasPositiveSurprise() bool {
	return e.Surprise != nil && *e.Surprise > 0
}

// Common response helper functions

// ToJSON converts any response to JSON string
func ToJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses JSON string to response struct
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
