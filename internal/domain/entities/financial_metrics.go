package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FinancialMetrics represents comprehensive fundamental financial data from Alpha Vantage
type FinancialMetrics struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	CompanyID uuid.UUID `json:"company_id" gorm:"type:uuid;not null" validate:"required"`
	Symbol    string    `json:"symbol" gorm:"type:string;not null;index" validate:"required"`

	// Valuation Metrics
	PERatio         float64 `json:"pe_ratio" gorm:"type:decimal(10,4)"`       // Price to Earnings
	PEGRatio        float64 `json:"peg_ratio" gorm:"type:decimal(10,4)"`      // PEG Ratio
	PriceToBook     float64 `json:"price_to_book" gorm:"type:decimal(10,4)"`  // P/B Ratio
	PriceToSales    float64 `json:"price_to_sales" gorm:"type:decimal(10,4)"` // P/S Ratio
	EVToRevenue     float64 `json:"ev_to_revenue" gorm:"type:decimal(10,4)"`  // EV/Revenue
	EVToEBITDA      float64 `json:"ev_to_ebitda" gorm:"type:decimal(10,4)"`   // EV/EBITDA
	EnterpriseValue int64   `json:"enterprise_value" gorm:"type:bigint"`      // Enterprise Value

	// Profitability Metrics
	ROE             float64 `json:"roe" gorm:"type:decimal(8,4)"`              // Return on Equity
	ROA             float64 `json:"roa" gorm:"type:decimal(8,4)"`              // Return on Assets
	ROIC            float64 `json:"roic" gorm:"type:decimal(8,4)"`             // Return on Invested Capital
	GrossMargin     float64 `json:"gross_margin" gorm:"type:decimal(8,4)"`     // Gross Profit Margin
	OperatingMargin float64 `json:"operating_margin" gorm:"type:decimal(8,4)"` // Operating Margin
	NetMargin       float64 `json:"net_margin" gorm:"type:decimal(8,4)"`       // Net Profit Margin

	// Financial Health
	DebtToEquity      float64 `json:"debt_to_equity" gorm:"type:decimal(8,4)"`        // D/E Ratio
	CurrentRatio      float64 `json:"current_ratio" gorm:"type:decimal(8,4)"`         // Current Ratio
	QuickRatio        float64 `json:"quick_ratio" gorm:"type:decimal(8,4)"`           // Quick Ratio
	InterestCoverage  float64 `json:"interest_coverage" gorm:"type:decimal(8,4)"`     // Interest Coverage
	BookValuePerShare float64 `json:"book_value_per_share" gorm:"type:decimal(10,4)"` // Book Value per Share

	// Growth Metrics
	RevenueGrowthTTM  float64 `json:"revenue_growth_ttm" gorm:"type:decimal(8,4)"`  // TTM Revenue Growth
	EarningsGrowthTTM float64 `json:"earnings_growth_ttm" gorm:"type:decimal(8,4)"` // TTM Earnings Growth
	RevenueGrowth3Y   float64 `json:"revenue_growth_3y" gorm:"type:decimal(8,4)"`   // 3Y Revenue Growth
	EarningsGrowth3Y  float64 `json:"earnings_growth_3y" gorm:"type:decimal(8,4)"`  // 3Y Earnings Growth
	RevenueGrowth5Y   float64 `json:"revenue_growth_5y" gorm:"type:decimal(8,4)"`   // 5Y Revenue Growth
	EarningsGrowth5Y  float64 `json:"earnings_growth_5y" gorm:"type:decimal(8,4)"`  // 5Y Earnings Growth

	// Per Share Metrics
	EPS                  float64 `json:"eps" gorm:"type:decimal(10,4)"`                      // Earnings per Share
	EPSGrowthTTM         float64 `json:"eps_growth_ttm" gorm:"type:decimal(8,4)"`            // EPS Growth TTM
	DividendPerShare     float64 `json:"dividend_per_share" gorm:"type:decimal(10,4)"`       // Dividend per Share
	DividendYield        float64 `json:"dividend_yield" gorm:"type:decimal(8,4)"`            // Dividend Yield
	FreeCashFlowPerShare float64 `json:"free_cash_flow_per_share" gorm:"type:decimal(10,4)"` // FCF per Share

	// Analyst Data
	AnalystTargetPrice      float64 `json:"analyst_target_price" gorm:"type:decimal(15,4)"` // Target Price
	AnalystRatingStrong     int32   `json:"analyst_rating_strong_buy" gorm:"type:integer"`  // Strong Buy Count
	AnalystRatingBuy        int32   `json:"analyst_rating_buy" gorm:"type:integer"`         // Buy Count
	AnalystRatingHold       int32   `json:"analyst_rating_hold" gorm:"type:integer"`        // Hold Count
	AnalystRatingSell       int32   `json:"analyst_rating_sell" gorm:"type:integer"`        // Sell Count
	AnalystRatingStrongSell int32   `json:"analyst_rating_strong_sell" gorm:"type:integer"` // Strong Sell Count

	// Data Quality and Metadata
	DataSource    string    `json:"data_source" gorm:"type:string;default:'alphavantage'"`
	LastUpdated   time.Time `json:"last_updated" gorm:"not null"`
	ReportingDate time.Time `json:"reporting_date" gorm:"type:date"` // When the financials were reported
	Currency      string    `json:"currency" gorm:"type:string;size:3;default:'USD'"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (FinancialMetrics) TableName() string {
	return "financial_metrics"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (fm *FinancialMetrics) BeforeCreate(tx *gorm.DB) error {
	if fm.ID == uuid.Nil {
		fm.ID = uuid.New()
	}
	return nil
}

// IsHealthy returns true if the company shows signs of financial health
func (fm *FinancialMetrics) IsHealthy() bool {
	return fm.CurrentRatio >= 1.0 &&
		fm.DebtToEquity <= 2.0 &&
		fm.ROE > 0 &&
		fm.NetMargin > 0
}

// IsGrowthStock returns true if the company shows strong growth characteristics
func (fm *FinancialMetrics) IsGrowthStock() bool {
	return fm.RevenueGrowthTTM > 10.0 &&
		fm.EarningsGrowthTTM > 15.0
}

// IsValueStock returns true if the company appears undervalued
func (fm *FinancialMetrics) IsValueStock() bool {
	return fm.PERatio > 0 && fm.PERatio < 15.0 &&
		fm.PriceToBook > 0 && fm.PriceToBook < 1.5
}

// GetAnalystConsensus returns the consensus recommendation based on analyst ratings
func (fm *FinancialMetrics) GetAnalystConsensus() string {
	total := fm.AnalystRatingStrong + fm.AnalystRatingBuy + fm.AnalystRatingHold +
		fm.AnalystRatingSell + fm.AnalystRatingStrongSell

	if total == 0 {
		return "NO_CONSENSUS"
	}

	buyRatings := fm.AnalystRatingStrong + fm.AnalystRatingBuy
	sellRatings := fm.AnalystRatingSell + fm.AnalystRatingStrongSell

	if float64(buyRatings)/float64(total) >= 0.6 {
		return "BUY"
	} else if float64(sellRatings)/float64(total) >= 0.6 {
		return "SELL"
	} else {
		return "HOLD"
	}
}

// CalculateFinancialScore returns a score from 0-100 based on financial health
func (fm *FinancialMetrics) CalculateFinancialScore() float64 {
	score := 0.0

	// Profitability (30%)
	if fm.ROE > 15 {
		score += 10
	} else if fm.ROE > 10 {
		score += 7
	} else if fm.ROE > 5 {
		score += 4
	}
	if fm.NetMargin > 10 {
		score += 10
	} else if fm.NetMargin > 5 {
		score += 7
	} else if fm.NetMargin > 0 {
		score += 4
	}
	if fm.ROA > 10 {
		score += 10
	} else if fm.ROA > 5 {
		score += 7
	} else if fm.ROA > 0 {
		score += 4
	}

	// Financial Health (25%)
	if fm.CurrentRatio > 2 {
		score += 8
	} else if fm.CurrentRatio > 1.5 {
		score += 6
	} else if fm.CurrentRatio > 1 {
		score += 3
	}
	if fm.DebtToEquity < 0.3 {
		score += 8
	} else if fm.DebtToEquity < 0.6 {
		score += 6
	} else if fm.DebtToEquity < 1 {
		score += 3
	}
	if fm.InterestCoverage > 5 {
		score += 9
	} else if fm.InterestCoverage > 2.5 {
		score += 6
	} else if fm.InterestCoverage > 1 {
		score += 3
	}

	// Growth (25%)
	if fm.RevenueGrowthTTM > 20 {
		score += 8
	} else if fm.RevenueGrowthTTM > 10 {
		score += 6
	} else if fm.RevenueGrowthTTM > 0 {
		score += 3
	}
	if fm.EarningsGrowthTTM > 25 {
		score += 8
	} else if fm.EarningsGrowthTTM > 15 {
		score += 6
	} else if fm.EarningsGrowthTTM > 0 {
		score += 3
	}
	if fm.EPSGrowthTTM > 20 {
		score += 9
	} else if fm.EPSGrowthTTM > 10 {
		score += 6
	} else if fm.EPSGrowthTTM > 0 {
		score += 3
	}

	// Valuation (20%)
	if fm.PERatio > 0 && fm.PERatio < 15 {
		score += 10
	} else if fm.PERatio < 25 {
		score += 6
	} else if fm.PERatio < 40 {
		score += 3
	}
	if fm.PriceToBook > 0 && fm.PriceToBook < 2 {
		score += 10
	} else if fm.PriceToBook < 4 {
		score += 6
	} else if fm.PriceToBook < 6 {
		score += 3
	}

	return score
}
