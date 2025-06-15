package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarketData represents real-time market data for a stock
type MarketData struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	CompanyID uuid.UUID `json:"company_id" gorm:"type:uuid;not null" validate:"required"`
	Symbol    string    `json:"symbol" gorm:"type:string;not null;index" validate:"required"`

	// Price Information
	CurrentPrice  float64 `json:"current_price" gorm:"type:decimal(15,4);not null"`
	OpenPrice     float64 `json:"open_price" gorm:"type:decimal(15,4)"`
	HighPrice     float64 `json:"high_price" gorm:"type:decimal(15,4)"`
	LowPrice      float64 `json:"low_price" gorm:"type:decimal(15,4)"`
	PreviousClose float64 `json:"previous_close" gorm:"type:decimal(15,4)"`

	// Change Information
	PriceChange     float64 `json:"price_change" gorm:"type:decimal(15,4)"`     // Absolute change
	PriceChangePerc float64 `json:"price_change_perc" gorm:"type:decimal(8,4)"` // Percentage change

	// Volume and Trading
	Volume    int64 `json:"volume" gorm:"type:bigint"`
	AvgVolume int64 `json:"avg_volume" gorm:"type:bigint"`
	MarketCap int64 `json:"market_cap" gorm:"type:bigint"`

	// Market Status
	IsMarketOpen bool   `json:"is_market_open" gorm:"type:boolean;default:false"`
	Currency     string `json:"currency" gorm:"type:string;size:3;default:'USD'"`
	Exchange     string `json:"exchange" gorm:"type:string;size:10"`

	// Timestamps
	MarketTimestamp time.Time      `json:"market_timestamp" gorm:"not null"` // When data was generated
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (MarketData) TableName() string {
	return "market_data"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (md *MarketData) BeforeCreate(tx *gorm.DB) error {
	if md.ID == uuid.Nil {
		md.ID = uuid.New()
	}
	return nil
}

// CompanyProfile represents detailed company information from external APIs
type CompanyProfile struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	Symbol string    `json:"symbol" gorm:"type:string;not null;unique_index" validate:"required"`

	// Basic Information
	Name        string `json:"name" gorm:"type:string;not null"`
	Description string `json:"description" gorm:"type:text"`
	Industry    string `json:"industry" gorm:"type:string"`
	Sector      string `json:"sector" gorm:"type:string"`
	Country     string `json:"country" gorm:"type:string"`
	Currency    string `json:"currency" gorm:"type:string;size:3"`

	// Financial Metrics
	MarketCap         int64   `json:"market_cap" gorm:"type:bigint"`
	SharesOutstanding int64   `json:"shares_outstanding" gorm:"type:bigint"`
	PERatio           float64 `json:"pe_ratio" gorm:"type:decimal(10,4)"`
	PEGRatio          float64 `json:"peg_ratio" gorm:"type:decimal(10,4)"`
	PriceToBook       float64 `json:"price_to_book" gorm:"type:decimal(10,4)"`
	DividendYield     float64 `json:"dividend_yield" gorm:"type:decimal(8,4)"`
	EPS               float64 `json:"eps" gorm:"type:decimal(10,4)"`

	// Trading Information
	Beta       float64 `json:"beta" gorm:"type:decimal(8,4)"`
	Week52High float64 `json:"week_52_high" gorm:"type:decimal(15,4)"`
	Week52Low  float64 `json:"week_52_low" gorm:"type:decimal(15,4)"`

	// Company Details
	Website       string    `json:"website" gorm:"type:string"`
	Logo          string    `json:"logo" gorm:"type:string"`
	IPODate       time.Time `json:"ipo_date" gorm:"type:date"`
	EmployeeCount int32     `json:"employee_count" gorm:"type:integer"`

	// Data Source and Freshness
	DataSource  string    `json:"data_source" gorm:"type:string;default:'finnhub'"`
	LastUpdated time.Time `json:"last_updated" gorm:"not null"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for GORM
func (CompanyProfile) TableName() string {
	return "company_profiles"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (cp *CompanyProfile) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	return nil
}

// NewsItem represents news articles related to stocks
type NewsItem struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	Symbol string    `json:"symbol" gorm:"type:string;not null;index" validate:"required"`

	// Article Information
	Title    string `json:"title" gorm:"type:string;not null"`
	Summary  string `json:"summary" gorm:"type:text"`
	URL      string `json:"url" gorm:"type:string;not null"`
	ImageURL string `json:"image_url" gorm:"type:string"`

	// Source Information
	Source   string `json:"source" gorm:"type:string;not null"`
	Category string `json:"category" gorm:"type:string"`
	Language string `json:"language" gorm:"type:string;size:2;default:'en'"`

	// Sentiment Analysis
	SentimentScore float64 `json:"sentiment_score" gorm:"type:decimal(4,3)"` // -1 to 1
	SentimentLabel string  `json:"sentiment_label" gorm:"type:string"`       // positive, negative, neutral

	// Timestamps
	PublishedAt time.Time      `json:"published_at" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for GORM
func (NewsItem) TableName() string {
	return "news_items"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (ni *NewsItem) BeforeCreate(tx *gorm.DB) error {
	if ni.ID == uuid.Nil {
		ni.ID = uuid.New()
	}
	return nil
}

// BasicFinancials represents basic financial metrics
type BasicFinancials struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	Symbol string    `json:"symbol" gorm:"type:string;not null;index" validate:"required"`

	// Valuation Metrics
	MarketCap       float64 `json:"market_cap" gorm:"type:decimal(20,2)"`
	PERatio         float64 `json:"pe_ratio" gorm:"type:decimal(10,4)"`
	PEGRatio        float64 `json:"peg_ratio" gorm:"type:decimal(10,4)"`
	PriceToSales    float64 `json:"price_to_sales" gorm:"type:decimal(10,4)"`
	PriceToBook     float64 `json:"price_to_book" gorm:"type:decimal(10,4)"`
	PriceToCashFlow float64 `json:"price_to_cash_flow" gorm:"type:decimal(10,4)"`

	// Profitability Metrics
	ROE             float64 `json:"roe" gorm:"type:decimal(8,4)"`
	ROA             float64 `json:"roa" gorm:"type:decimal(8,4)"`
	ROI             float64 `json:"roi" gorm:"type:decimal(8,4)"`
	GrossMargin     float64 `json:"gross_margin" gorm:"type:decimal(8,4)"`
	OperatingMargin float64 `json:"operating_margin" gorm:"type:decimal(8,4)"`
	NetMargin       float64 `json:"net_margin" gorm:"type:decimal(8,4)"`

	// Growth Metrics
	RevenueGrowth  float64 `json:"revenue_growth" gorm:"type:decimal(8,4)"`
	EarningsGrowth float64 `json:"earnings_growth" gorm:"type:decimal(8,4)"`
	DividendGrowth float64 `json:"dividend_growth" gorm:"type:decimal(8,4)"`

	// Financial Health
	DebtToEquity float64 `json:"debt_to_equity" gorm:"type:decimal(8,4)"`
	CurrentRatio float64 `json:"current_ratio" gorm:"type:decimal(8,4)"`
	QuickRatio   float64 `json:"quick_ratio" gorm:"type:decimal(8,4)"`

	// Per Share Metrics
	EPS               float64 `json:"eps" gorm:"type:decimal(10,4)"`
	BookValuePerShare float64 `json:"book_value_per_share" gorm:"type:decimal(10,4)"`
	CashPerShare      float64 `json:"cash_per_share" gorm:"type:decimal(10,4)"`
	DividendPerShare  float64 `json:"dividend_per_share" gorm:"type:decimal(10,4)"`

	// Period Information
	Period        string `json:"period" gorm:"type:string"` // annual, quarterly
	FiscalYear    int    `json:"fiscal_year" gorm:"type:integer"`
	FiscalQuarter int    `json:"fiscal_quarter" gorm:"type:integer"`

	// Data Source and Freshness
	DataSource  string    `json:"data_source" gorm:"type:string;default:'finnhub'"`
	LastUpdated time.Time `json:"last_updated" gorm:"not null"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for GORM
func (BasicFinancials) TableName() string {
	return "basic_financials"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (bf *BasicFinancials) BeforeCreate(tx *gorm.DB) error {
	if bf.ID == uuid.Nil {
		bf.ID = uuid.New()
	}
	return nil
}

// IsStale checks if the market data is stale based on market hours
func (md *MarketData) IsStale(maxAge time.Duration) bool {
	return time.Since(md.MarketTimestamp) > maxAge
}

// GetPriceChangePercentage calculates price change percentage
func (md *MarketData) GetPriceChangePercentage() float64 {
	if md.PreviousClose == 0 {
		return 0
	}
	return ((md.CurrentPrice - md.PreviousClose) / md.PreviousClose) * 100
}

// IsPositiveGainer checks if stock is gaining
func (md *MarketData) IsPositiveGainer() bool {
	return md.PriceChange > 0
}

// GetMarketCapInBillions returns market cap in billions
func (cp *CompanyProfile) GetMarketCapInBillions() float64 {
	return float64(cp.MarketCap) / 1_000_000_000
}

// IsLargeCapStock checks if company is large cap (>10B market cap)
func (cp *CompanyProfile) IsLargeCapStock() bool {
	return cp.GetMarketCapInBillions() > 10
}

// GetSentimentCategory returns sentiment category for news
func (ni *NewsItem) GetSentimentCategory() string {
	if ni.SentimentScore > 0.1 {
		return "positive"
	} else if ni.SentimentScore < -0.1 {
		return "negative"
	}
	return "neutral"
}
