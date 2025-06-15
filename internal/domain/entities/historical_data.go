package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HistoricalData represents historical price data from Alpha Vantage
type HistoricalData struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	CompanyID uuid.UUID `json:"company_id" gorm:"type:uuid;not null" validate:"required"`
	Symbol    string    `json:"symbol" gorm:"type:string;not null;index" validate:"required"`

	// OHLCV Data
	Date          time.Time `json:"date" gorm:"type:date;not null;index"`
	OpenPrice     float64   `json:"open_price" gorm:"type:decimal(15,4);not null"`
	HighPrice     float64   `json:"high_price" gorm:"type:decimal(15,4);not null"`
	LowPrice      float64   `json:"low_price" gorm:"type:decimal(15,4);not null"`
	ClosePrice    float64   `json:"close_price" gorm:"type:decimal(15,4);not null"`
	AdjustedClose float64   `json:"adjusted_close" gorm:"type:decimal(15,4);not null"`
	Volume        int64     `json:"volume" gorm:"type:bigint;not null"`

	// Daily Calculations
	DailyReturn     float64 `json:"daily_return" gorm:"type:decimal(8,4)"`     // (Close - PrevClose) / PrevClose
	DailyRange      float64 `json:"daily_range" gorm:"type:decimal(15,4)"`     // High - Low
	DailyVolatility float64 `json:"daily_volatility" gorm:"type:decimal(8,4)"` // Daily volatility measure

	// Technical Levels
	IsGapUp     bool    `json:"is_gap_up" gorm:"type:boolean;default:false"`
	IsGapDown   bool    `json:"is_gap_down" gorm:"type:boolean;default:false"`
	GapPercent  float64 `json:"gap_percent" gorm:"type:decimal(8,4)"`
	IsBreakout  bool    `json:"is_breakout" gorm:"type:boolean;default:false"`
	IsBreakdown bool    `json:"is_breakdown" gorm:"type:boolean;default:false"`

	// Time Frame
	TimeFrame string `json:"time_frame" gorm:"type:string;size:10;default:'1D'"` // 1D, 1W, 1M

	// Data Quality
	DataSource  string    `json:"data_source" gorm:"type:string;default:'alphavantage'"`
	LastUpdated time.Time `json:"last_updated" gorm:"not null"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (HistoricalData) TableName() string {
	return "historical_data"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (hd *HistoricalData) BeforeCreate(tx *gorm.DB) error {
	if hd.ID == uuid.Nil {
		hd.ID = uuid.New()
	}
	return nil
}

// CalculateDailyReturn calculates the daily return based on previous close
func (hd *HistoricalData) CalculateDailyReturn(previousClose float64) {
	if previousClose > 0 {
		hd.DailyReturn = ((hd.ClosePrice - previousClose) / previousClose) * 100
	}
}

// CalculateDailyRange calculates the daily trading range
func (hd *HistoricalData) CalculateDailyRange() {
	hd.DailyRange = hd.HighPrice - hd.LowPrice
}

// IsGreen returns true if the close price is higher than open price
func (hd *HistoricalData) IsGreen() bool {
	return hd.ClosePrice > hd.OpenPrice
}

// IsRed returns true if the close price is lower than open price
func (hd *HistoricalData) IsRed() bool {
	return hd.ClosePrice < hd.OpenPrice
}

// IsDoji returns true if open and close are approximately equal (within 0.1%)
func (hd *HistoricalData) IsDoji() bool {
	if hd.OpenPrice == 0 {
		return false
	}
	diff := ((hd.ClosePrice - hd.OpenPrice) / hd.OpenPrice) * 100
	return diff >= -0.1 && diff <= 0.1
}

// GetBodyPercent returns the percentage of the body relative to the full range
func (hd *HistoricalData) GetBodyPercent() float64 {
	if hd.DailyRange == 0 {
		return 0
	}
	bodySize := hd.ClosePrice - hd.OpenPrice
	if bodySize < 0 {
		bodySize = -bodySize
	}
	return (bodySize / hd.DailyRange) * 100
}

// GetUpperShadowPercent returns the percentage of upper shadow relative to full range
func (hd *HistoricalData) GetUpperShadowPercent() float64 {
	if hd.DailyRange == 0 {
		return 0
	}
	var upperShadow float64
	if hd.ClosePrice > hd.OpenPrice {
		upperShadow = hd.HighPrice - hd.ClosePrice
	} else {
		upperShadow = hd.HighPrice - hd.OpenPrice
	}
	return (upperShadow / hd.DailyRange) * 100
}

// GetLowerShadowPercent returns the percentage of lower shadow relative to full range
func (hd *HistoricalData) GetLowerShadowPercent() float64 {
	if hd.DailyRange == 0 {
		return 0
	}
	var lowerShadow float64
	if hd.ClosePrice > hd.OpenPrice {
		lowerShadow = hd.OpenPrice - hd.LowPrice
	} else {
		lowerShadow = hd.ClosePrice - hd.LowPrice
	}
	return (lowerShadow / hd.DailyRange) * 100
}

// IsSignificantVolume returns true if volume is above average
func (hd *HistoricalData) IsSignificantVolume(avgVolume int64) bool {
	if avgVolume == 0 {
		return false
	}
	return hd.Volume > int64(float64(avgVolume)*1.5) // 50% above average
}

// DetectGap detects if there's a gap from previous day
func (hd *HistoricalData) DetectGap(previousHigh, previousLow float64) {
	gapThreshold := 2.0 // 2% gap threshold

	if previousHigh > 0 && hd.LowPrice > previousHigh {
		gapPercent := ((hd.LowPrice - previousHigh) / previousHigh) * 100
		if gapPercent >= gapThreshold {
			hd.IsGapUp = true
			hd.GapPercent = gapPercent
		}
	}

	if previousLow > 0 && hd.HighPrice < previousLow {
		gapPercent := ((previousLow - hd.HighPrice) / previousLow) * 100
		if gapPercent >= gapThreshold {
			hd.IsGapDown = true
			hd.GapPercent = -gapPercent
		}
	}
}

// HistoricalDataSummary represents aggregated historical data statistics
type HistoricalDataSummary struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	CompanyID uuid.UUID `json:"company_id" gorm:"type:uuid;not null" validate:"required"`
	Symbol    string    `json:"symbol" gorm:"type:string;not null;index" validate:"required"`

	// Time Period
	StartDate time.Time `json:"start_date" gorm:"type:date;not null"`
	EndDate   time.Time `json:"end_date" gorm:"type:date;not null"`
	Period    string    `json:"period" gorm:"type:string;size:10"` // 1M, 3M, 6M, 1Y, 2Y, 5Y

	// Price Statistics
	HighestPrice    float64 `json:"highest_price" gorm:"type:decimal(15,4)"`
	LowestPrice     float64 `json:"lowest_price" gorm:"type:decimal(15,4)"`
	AveragePrice    float64 `json:"average_price" gorm:"type:decimal(15,4)"`
	StartPrice      float64 `json:"start_price" gorm:"type:decimal(15,4)"`
	EndPrice        float64 `json:"end_price" gorm:"type:decimal(15,4)"`
	PriceChange     float64 `json:"price_change" gorm:"type:decimal(15,4)"`
	PriceChangePerc float64 `json:"price_change_perc" gorm:"type:decimal(8,4)"`

	// Volatility Metrics
	Volatility  float64 `json:"volatility" gorm:"type:decimal(8,4)"`   // Standard deviation of returns
	Beta        float64 `json:"beta" gorm:"type:decimal(8,4)"`         // Beta relative to market
	SharpeRatio float64 `json:"sharpe_ratio" gorm:"type:decimal(8,4)"` // Risk-adjusted return
	MaxDrawdown float64 `json:"max_drawdown" gorm:"type:decimal(8,4)"` // Maximum decline from peak

	// Volume Statistics
	AverageVolume int64 `json:"average_volume" gorm:"type:bigint"`
	TotalVolume   int64 `json:"total_volume" gorm:"type:bigint"`
	HighestVolume int64 `json:"highest_volume" gorm:"type:bigint"`
	LowestVolume  int64 `json:"lowest_volume" gorm:"type:bigint"`

	// Trading Days
	TotalTradingDays int32 `json:"total_trading_days" gorm:"type:integer"`
	UpDays           int32 `json:"up_days" gorm:"type:integer"`
	DownDays         int32 `json:"down_days" gorm:"type:integer"`
	UnchangedDays    int32 `json:"unchanged_days" gorm:"type:integer"`

	// Performance Metrics
	WinRate       float64 `json:"win_rate" gorm:"type:decimal(8,4)"`        // Percentage of up days
	AverageGain   float64 `json:"average_gain" gorm:"type:decimal(8,4)"`    // Average gain on up days
	AverageLoss   float64 `json:"average_loss" gorm:"type:decimal(8,4)"`    // Average loss on down days
	GainLossRatio float64 `json:"gain_loss_ratio" gorm:"type:decimal(8,4)"` // Average gain / Average loss

	// Data Quality
	DataSource  string    `json:"data_source" gorm:"type:string;default:'alphavantage'"`
	LastUpdated time.Time `json:"last_updated" gorm:"not null"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (HistoricalDataSummary) TableName() string {
	return "historical_data_summaries"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (hds *HistoricalDataSummary) BeforeCreate(tx *gorm.DB) error {
	if hds.ID == uuid.Nil {
		hds.ID = uuid.New()
	}
	return nil
}

// CalculateMetrics calculates various performance metrics from historical data
func (hds *HistoricalDataSummary) CalculateMetrics() {
	if hds.TotalTradingDays > 0 {
		hds.WinRate = (float64(hds.UpDays) / float64(hds.TotalTradingDays)) * 100
	}

	if hds.StartPrice > 0 {
		hds.PriceChange = hds.EndPrice - hds.StartPrice
		hds.PriceChangePerc = (hds.PriceChange / hds.StartPrice) * 100
	}

	if hds.AverageLoss != 0 {
		hds.GainLossRatio = hds.AverageGain / (-hds.AverageLoss) // Make loss positive for ratio
	}
}

// IsOutperforming returns true if the stock outperformed in the period
func (hds *HistoricalDataSummary) IsOutperforming(marketReturn float64) bool {
	return hds.PriceChangePerc > marketReturn
}

// IsLowVolatility returns true if the stock has low volatility
func (hds *HistoricalDataSummary) IsLowVolatility() bool {
	return hds.Volatility < 20.0 // Less than 20% volatility
}

// IsHighVolatility returns true if the stock has high volatility
func (hds *HistoricalDataSummary) IsHighVolatility() bool {
	return hds.Volatility > 40.0 // More than 40% volatility
}

// GetRiskLevel returns the risk level based on volatility and beta
func (hds *HistoricalDataSummary) GetRiskLevel() string {
	if hds.Volatility < 15.0 && hds.Beta < 0.8 {
		return "LOW"
	} else if hds.Volatility < 25.0 && hds.Beta < 1.2 {
		return "MEDIUM"
	} else {
		return "HIGH"
	}
}
