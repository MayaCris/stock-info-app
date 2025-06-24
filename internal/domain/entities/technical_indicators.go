package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TechnicalIndicators represents technical analysis indicators from Alpha Vantage
type TechnicalIndicators struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	CompanyID uuid.UUID `json:"company_id" gorm:"type:uuid;not null" validate:"required"`
	Symbol    string    `json:"symbol" gorm:"type:string;not null;index" validate:"required"`

	// Moving Averages
	SMA20  float64 `json:"sma_20" gorm:"type:decimal(15,4)"`  // Simple Moving Average 20 days
	SMA50  float64 `json:"sma_50" gorm:"type:decimal(15,4)"`  // Simple Moving Average 50 days
	SMA200 float64 `json:"sma_200" gorm:"type:decimal(15,4)"` // Simple Moving Average 200 days
	EMA12  float64 `json:"ema_12" gorm:"type:decimal(15,4)"`  // Exponential Moving Average 12 days
	EMA26  float64 `json:"ema_26" gorm:"type:decimal(15,4)"`  // Exponential Moving Average 26 days

	// Oscillators
	RSI       float64 `json:"rsi" gorm:"type:decimal(8,4)"`        // Relative Strength Index
	StochK    float64 `json:"stoch_k" gorm:"type:decimal(8,4)"`    // Stochastic %K
	StochD    float64 `json:"stoch_d" gorm:"type:decimal(8,4)"`    // Stochastic %D
	WilliamsR float64 `json:"williams_r" gorm:"type:decimal(8,4)"` // Williams %R
	CCI       float64 `json:"cci" gorm:"type:decimal(8,4)"`        // Commodity Channel Index

	// MACD (Moving Average Convergence Divergence)
	MACD          float64 `json:"macd" gorm:"type:decimal(15,4)"`           // MACD Line
	MACDSignal    float64 `json:"macd_signal" gorm:"type:decimal(15,4)"`    // MACD Signal Line
	MACDHistogram float64 `json:"macd_histogram" gorm:"type:decimal(15,4)"` // MACD Histogram

	// Bollinger Bands
	BBUpper  float64 `json:"bb_upper" gorm:"type:decimal(15,4)"`  // Bollinger Bands Upper
	BBMiddle float64 `json:"bb_middle" gorm:"type:decimal(15,4)"` // Bollinger Bands Middle (SMA20)
	BBLower  float64 `json:"bb_lower" gorm:"type:decimal(15,4)"`  // Bollinger Bands Lower

	// Volume Indicators
	VWAP       float64 `json:"vwap" gorm:"type:decimal(15,4)"`    // Volume Weighted Average Price
	VolumeMA20 int64   `json:"volume_ma_20" gorm:"type:bigint"`   // 20-day Volume Moving Average
	OBV        int64   `json:"obv" gorm:"type:bigint"`            // On Balance Volume
	ADLINE     float64 `json:"ad_line" gorm:"type:decimal(15,4)"` // Accumulation/Distribution Line

	// Trend Indicators
	ADX        float64 `json:"adx" gorm:"type:decimal(8,4)"`        // Average Directional Index
	AROON_UP   float64 `json:"aroon_up" gorm:"type:decimal(8,4)"`   // Aroon Up
	AROON_DOWN float64 `json:"aroon_down" gorm:"type:decimal(8,4)"` // Aroon Down
	SAR        float64 `json:"sar" gorm:"type:decimal(15,4)"`       // Parabolic SAR

	// Volatility Indicators
	ATR        float64 `json:"atr" gorm:"type:decimal(15,4)"`         // Average True Range
	BandWidth  float64 `json:"band_width" gorm:"type:decimal(8,4)"`   // Bollinger Bands Width
	BBPercentB float64 `json:"bb_percent_b" gorm:"type:decimal(8,4)"` // Bollinger %B

	// Support and Resistance Levels
	Resistance1 float64 `json:"resistance_1" gorm:"type:decimal(15,4)"` // First resistance level
	Resistance2 float64 `json:"resistance_2" gorm:"type:decimal(15,4)"` // Second resistance level
	Support1    float64 `json:"support_1" gorm:"type:decimal(15,4)"`    // First support level
	Support2    float64 `json:"support_2" gorm:"type:decimal(15,4)"`    // Second support level
	PivotPoint  float64 `json:"pivot_point" gorm:"type:decimal(15,4)"`  // Pivot Point

	// Period and Time Frame
	TimeFrame string `json:"time_frame" gorm:"type:string;size:10;default:'1D'"` // 1D, 1W, 1M
	Period    int32  `json:"period" gorm:"type:integer;default:14"`              // Period for calculations

	// Signals and Interpretation
	TrendSignal    string  `json:"trend_signal" gorm:"type:string;size:20"`    // BULLISH, BEARISH, NEUTRAL
	MomentumSignal string  `json:"momentum_signal" gorm:"type:string;size:20"` // STRONG, WEAK, NEUTRAL
	VolumeSignal   string  `json:"volume_signal" gorm:"type:string;size:20"`   // HIGH, LOW, NORMAL
	OverallSignal  string  `json:"overall_signal" gorm:"type:string;size:20"`  // BUY, SELL, HOLD
	SignalStrength float64 `json:"signal_strength" gorm:"type:decimal(8,4)"`   // 0-100 confidence

	// Data Quality and Metadata
	DataSource  string    `json:"data_source" gorm:"type:string;default:'alphavantage'"`
	LastUpdated time.Time `json:"last_updated" gorm:"not null"`
	MarketDate  time.Time `json:"market_date" gorm:"type:date"` // Date of the market data used

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (TechnicalIndicators) TableName() string {
	return "technical_indicators"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (ti *TechnicalIndicators) BeforeCreate(tx *gorm.DB) error {
	if ti.ID == uuid.Nil {
		ti.ID = uuid.New()
	}
	return nil
}

// IsBullish returns true if technical indicators suggest a bullish trend
func (ti *TechnicalIndicators) IsBullish() bool {
	conditions := 0

	// RSI in healthy range (not overbought)
	if ti.RSI > 30 && ti.RSI < 70 {
		conditions++
	}

	// MACD bullish signal
	if ti.MACD > ti.MACDSignal {
		conditions++
	}

	// Price above key moving averages
	if ti.SMA20 > ti.SMA50 && ti.SMA50 > ti.SMA200 {
		conditions++
	}

	// ADX shows strong trend
	if ti.ADX > 25 {
		conditions++
	}

	return conditions >= 3
}

// IsBearish returns true if technical indicators suggest a bearish trend
func (ti *TechnicalIndicators) IsBearish() bool {
	conditions := 0

	// RSI oversold or showing weakness
	if ti.RSI < 30 || ti.RSI > 70 {
		conditions++
	}

	// MACD bearish signal
	if ti.MACD < ti.MACDSignal {
		conditions++
	}

	// Price below key moving averages
	if ti.SMA20 < ti.SMA50 {
		conditions++
	}

	// Aroon showing bearish trend
	if ti.AROON_DOWN > ti.AROON_UP {
		conditions++
	}

	return conditions >= 3
}

// IsOverbought returns true if the stock appears overbought
func (ti *TechnicalIndicators) IsOverbought() bool {
	return ti.RSI > 70 || ti.StochK > 80 || ti.WilliamsR > -20
}

// IsOversold returns true if the stock appears oversold
func (ti *TechnicalIndicators) IsOversold() bool {
	return ti.RSI < 30 || ti.StochK < 20 || ti.WilliamsR < -80
}

// CalculateTechnicalScore returns a score from 0-100 based on technical analysis
func (ti *TechnicalIndicators) CalculateTechnicalScore() float64 {
	score := 50.0 // Start with neutral score

	// Trend Analysis (40%)
	if ti.SMA20 > ti.SMA50 && ti.SMA50 > ti.SMA200 {
		score += 15 // Strong uptrend
	} else if ti.SMA20 > ti.SMA50 {
		score += 8 // Short-term uptrend
	} else if ti.SMA20 < ti.SMA50 && ti.SMA50 < ti.SMA200 {
		score -= 15 // Strong downtrend
	} else if ti.SMA20 < ti.SMA50 {
		score -= 8 // Short-term downtrend
	}

	if ti.ADX > 25 {
		if ti.IsBullish() {
			score += 10 // Strong bullish trend
		} else if ti.IsBearish() {
			score -= 10 // Strong bearish trend
		}
	}

	// Momentum Analysis (30%)
	if ti.RSI > 50 && ti.RSI < 70 {
		score += 8 // Healthy bullish momentum
	} else if ti.RSI > 30 && ti.RSI < 50 {
		score += 3 // Weak momentum
	} else if ti.RSI > 70 {
		score -= 5 // Overbought
	} else if ti.RSI < 30 {
		score -= 8 // Oversold
	}

	if ti.MACD > ti.MACDSignal {
		score += 7 // MACD bullish
	} else {
		score -= 7 // MACD bearish
	}

	// Volume Analysis (20%)
	if ti.OBV > 0 {
		score += 5 // Positive volume flow
	} else {
		score -= 5 // Negative volume flow
	}

	// Volatility Analysis (10%)
	if ti.BBPercentB > 0.2 && ti.BBPercentB < 0.8 {
		score += 5 // Normal volatility
	} else {
		score -= 3 // High volatility (risky)
	}

	// Ensure score stays within bounds
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	return score
}

// GetTechnicalSignals returns a summary of technical signals
func (ti *TechnicalIndicators) GetTechnicalSignals() map[string]string {
	signals := make(map[string]string)

	// Trend Signal
	if ti.IsBullish() {
		signals["trend"] = "BULLISH"
	} else if ti.IsBearish() {
		signals["trend"] = "BEARISH"
	} else {
		signals["trend"] = "NEUTRAL"
	}

	// Momentum Signal
	if ti.RSI > 60 && ti.MACD > ti.MACDSignal {
		signals["momentum"] = "STRONG_BULLISH"
	} else if ti.RSI > 50 {
		signals["momentum"] = "BULLISH"
	} else if ti.RSI < 40 && ti.MACD < ti.MACDSignal {
		signals["momentum"] = "STRONG_BEARISH"
	} else if ti.RSI < 50 {
		signals["momentum"] = "BEARISH"
	} else {
		signals["momentum"] = "NEUTRAL"
	}

	// Support/Resistance Signal
	if ti.Support1 > 0 && ti.Resistance1 > 0 {
		signals["support_resistance"] = "DEFINED"
	} else {
		signals["support_resistance"] = "UNDEFINED"
	}

	// Volume Signal
	if ti.OBV > 0 {
		signals["volume"] = "ACCUMULATION"
	} else {
		signals["volume"] = "DISTRIBUTION"
	}

	// Overall Signal
	score := ti.CalculateTechnicalScore()
	if score >= 70 {
		signals["overall"] = "STRONG_BUY"
	} else if score >= 60 {
		signals["overall"] = "BUY"
	} else if score >= 40 {
		signals["overall"] = "HOLD"
	} else if score >= 30 {
		signals["overall"] = "SELL"
	} else {
		signals["overall"] = "STRONG_SELL"
	}

	return signals
}
