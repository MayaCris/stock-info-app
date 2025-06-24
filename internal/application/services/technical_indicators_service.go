package services

import (
	"context"
	"fmt"
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// TechnicalIndicatorsService provides business logic for technical indicators
type TechnicalIndicatorsService struct {
	technicalRepo interfaces.TechnicalIndicatorsRepository
	companyRepo   interfaces.CompanyRepository
}

// NewTechnicalIndicatorsService creates a new instance of TechnicalIndicatorsService
func NewTechnicalIndicatorsService(
	technicalRepo interfaces.TechnicalIndicatorsRepository,
	companyRepo interfaces.CompanyRepository,
) *TechnicalIndicatorsService {
	return &TechnicalIndicatorsService{
		technicalRepo: technicalRepo,
		companyRepo:   companyRepo,
	}
}

// GetTechnicalIndicatorsBySymbol retrieves technical indicators for a specific symbol
func (s *TechnicalIndicatorsService) GetTechnicalIndicatorsBySymbol(ctx context.Context, symbol string) (*entities.TechnicalIndicators, error) {
	indicators, err := s.technicalRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get technical indicators for symbol %s: %w", symbol, err)
	}
	return indicators, nil
}

// CreateTechnicalIndicators creates new technical indicators record
func (s *TechnicalIndicatorsService) CreateTechnicalIndicators(ctx context.Context, indicators *entities.TechnicalIndicators) error {
	// Validate company exists
	if _, err := s.companyRepo.GetByID(ctx, indicators.CompanyID); err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Set timestamps
	now := time.Now()
	indicators.LastUpdated = now

	// Calculate signals based on indicators
	s.calculateSignals(indicators)

	if err := s.technicalRepo.Create(ctx, indicators); err != nil {
		return fmt.Errorf("failed to create technical indicators: %w", err)
	}

	return nil
}

// UpdateTechnicalIndicators updates existing technical indicators
func (s *TechnicalIndicatorsService) UpdateTechnicalIndicators(ctx context.Context, indicators *entities.TechnicalIndicators) error {
	// Check if record exists
	existing, err := s.technicalRepo.GetByID(ctx, indicators.ID)
	if err != nil {
		return fmt.Errorf("technical indicators not found: %w", err)
	}

	// Update timestamp
	indicators.LastUpdated = time.Now()

	// Preserve creation timestamp
	indicators.CreatedAt = existing.CreatedAt

	// Recalculate signals
	s.calculateSignals(indicators)

	if err := s.technicalRepo.Update(ctx, indicators); err != nil {
		return fmt.Errorf("failed to update technical indicators: %w", err)
	}

	return nil
}

// GetTechnicalAnalysis provides comprehensive technical analysis for a symbol
func (s *TechnicalIndicatorsService) GetTechnicalAnalysis(ctx context.Context, symbol string) (*response.TechnicalAnalysisResponse, error) {
	indicators, err := s.technicalRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get technical indicators: %w", err)
	}

	// Calculate technical score
	technicalScore := indicators.CalculateTechnicalScore()

	// Get technical signals
	signals := indicators.GetTechnicalSignals()

	// Generate insights
	insights := s.generateTechnicalInsights(indicators)

	analysis := &response.TechnicalAnalysisResponse{
		Symbol:         indicators.Symbol,
		TechnicalScore: technicalScore,
		Signals:        signals,
		Insights:       insights,

		// Key Indicators
		RSI:        indicators.RSI,
		MACD:       indicators.MACD,
		MACDSignal: indicators.MACDSignal,

		// Moving Averages
		SMA20:  indicators.SMA20,
		SMA50:  indicators.SMA50,
		SMA200: indicators.SMA200,

		// Bollinger Bands
		BBUpper:    indicators.BBUpper,
		BBMiddle:   indicators.BBMiddle,
		BBLower:    indicators.BBLower,
		BBPercentB: indicators.BBPercentB,

		// Volume
		VolumeMA20: indicators.VolumeMA20,
		OBV:        indicators.OBV,

		// Support/Resistance
		Support1:    indicators.Support1,
		Support2:    indicators.Support2,
		Resistance1: indicators.Resistance1,
		Resistance2: indicators.Resistance2,

		// Volatility
		ATR:       indicators.ATR,
		BandWidth: indicators.BandWidth,

		LastUpdated: indicators.LastUpdated,
	}

	return analysis, nil
}

// GetBullishStocks retrieves stocks with bullish technical signals
func (s *TechnicalIndicatorsService) GetBullishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetBullishStocks(ctx)
}

// GetBearishStocks retrieves stocks with bearish technical signals
func (s *TechnicalIndicatorsService) GetBearishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetBearishStocks(ctx)
}

// GetOverboughtStocks retrieves overbought stocks
func (s *TechnicalIndicatorsService) GetOverboughtStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetOverboughtStocks(ctx)
}

// GetOversoldStocks retrieves oversold stocks
func (s *TechnicalIndicatorsService) GetOversoldStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetOversoldStocks(ctx)
}

// GetGoldenCrossStocks retrieves stocks with golden cross pattern
func (s *TechnicalIndicatorsService) GetGoldenCrossStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetGoldenCross(ctx)
}

// GetDeathCrossStocks retrieves stocks with death cross pattern
func (s *TechnicalIndicatorsService) GetDeathCrossStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetDeathCross(ctx)
}

// GetMACDBullishStocks retrieves stocks with bullish MACD signals
func (s *TechnicalIndicatorsService) GetMACDBullishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetMACDBullish(ctx)
}

// GetMACDBearishStocks retrieves stocks with bearish MACD signals
func (s *TechnicalIndicatorsService) GetMACDBearishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetMACDBearish(ctx)
}

// GetTopTechnicalScores retrieves stocks with highest technical scores
func (s *TechnicalIndicatorsService) GetTopTechnicalScores(ctx context.Context, limit int) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetTopByScore(ctx, limit)
}

// GetStrongestSignals retrieves stocks with strongest technical signals
func (s *TechnicalIndicatorsService) GetStrongestSignals(ctx context.Context, limit int) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetStrongestSignals(ctx, limit)
}

// GetBollingerBreakouts retrieves stocks with Bollinger Band breakouts
func (s *TechnicalIndicatorsService) GetBollingerBreakouts(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetBollingerBreakout(ctx)
}

// GetBollingerSqueezes retrieves stocks with Bollinger Band squeezes
func (s *TechnicalIndicatorsService) GetBollingerSqueezes(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetBollingerSqueeze(ctx)
}

// GetHighVolumeStocks retrieves stocks with high trading volume
func (s *TechnicalIndicatorsService) GetHighVolumeStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetHighVolumeStocks(ctx)
}

// GetHighVolatilityStocks retrieves stocks with high volatility
func (s *TechnicalIndicatorsService) GetHighVolatilityStocks(ctx context.Context, threshold float64) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetHighVolatility(ctx, threshold)
}

// GetLowVolatilityStocks retrieves stocks with low volatility
func (s *TechnicalIndicatorsService) GetLowVolatilityStocks(ctx context.Context, threshold float64) ([]*entities.TechnicalIndicators, error) {
	return s.technicalRepo.GetLowVolatility(ctx, threshold)
}

// RefreshStaleData refreshes technical indicators that are outdated
func (s *TechnicalIndicatorsService) RefreshStaleData(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	staleData, err := s.technicalRepo.GetStaleData(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("failed to get stale data: %w", err)
	}

	// This would trigger data refresh from external APIs
	for _, indicators := range staleData {
		fmt.Printf("Technical indicators for %s need refresh (last updated: %s)\n",
			indicators.Symbol, indicators.LastUpdated.Format(time.RFC3339))
	}

	return nil
}

// AnalyzeMarketSentiment analyzes overall market sentiment based on technical indicators
func (s *TechnicalIndicatorsService) AnalyzeMarketSentiment(ctx context.Context) (map[string]interface{}, error) {
	// Get all recent indicators
	indicators, err := s.technicalRepo.List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get technical indicators: %w", err)
	}

	var bullishCount, bearishCount, neutralCount int
	var totalScore float64

	for _, indicator := range indicators {
		score := indicator.CalculateTechnicalScore()
		totalScore += score

		if indicator.IsBullish() {
			bullishCount++
		} else if indicator.IsBearish() {
			bearishCount++
		} else {
			neutralCount++
		}
	}

	total := len(indicators)
	if total == 0 {
		return map[string]interface{}{
			"sentiment": "UNKNOWN",
			"message":   "No data available",
		}, nil
	}

	avgScore := totalScore / float64(total)
	bullishPercent := float64(bullishCount) / float64(total) * 100
	bearishPercent := float64(bearishCount) / float64(total) * 100

	var sentiment string
	if bullishPercent > 60 {
		sentiment = "BULLISH"
	} else if bearishPercent > 60 {
		sentiment = "BEARISH"
	} else {
		sentiment = "NEUTRAL"
	}

	return map[string]interface{}{
		"sentiment":       sentiment,
		"average_score":   avgScore,
		"bullish_percent": bullishPercent,
		"bearish_percent": bearishPercent,
		"neutral_percent": float64(neutralCount) / float64(total) * 100,
		"total_stocks":    total,
		"bullish_count":   bullishCount,
		"bearish_count":   bearishCount,
		"neutral_count":   neutralCount,
	}, nil
}

// Helper methods

func (s *TechnicalIndicatorsService) calculateSignals(indicators *entities.TechnicalIndicators) {
	// Calculate trend signal
	if indicators.IsBullish() {
		indicators.TrendSignal = "BULLISH"
	} else if indicators.IsBearish() {
		indicators.TrendSignal = "BEARISH"
	} else {
		indicators.TrendSignal = "NEUTRAL"
	}

	// Calculate momentum signal
	if indicators.RSI > 60 && indicators.MACD > indicators.MACDSignal {
		indicators.MomentumSignal = "STRONG"
	} else if indicators.RSI > 50 {
		indicators.MomentumSignal = "WEAK"
	} else {
		indicators.MomentumSignal = "NEUTRAL"
	}

	// Calculate volume signal
	if indicators.OBV > 0 {
		indicators.VolumeSignal = "HIGH"
	} else {
		indicators.VolumeSignal = "LOW"
	}

	// Calculate overall signal
	score := indicators.CalculateTechnicalScore()
	indicators.SignalStrength = score

	if score >= 70 {
		indicators.OverallSignal = "BUY"
	} else if score >= 60 {
		indicators.OverallSignal = "WEAK_BUY"
	} else if score >= 40 {
		indicators.OverallSignal = "HOLD"
	} else if score >= 30 {
		indicators.OverallSignal = "WEAK_SELL"
	} else {
		indicators.OverallSignal = "SELL"
	}
}

func (s *TechnicalIndicatorsService) generateTechnicalInsights(indicators *entities.TechnicalIndicators) []string {
	var insights []string

	// RSI insights
	if indicators.IsOverbought() {
		insights = append(insights, "Stock is in overbought territory - potential selling opportunity")
	} else if indicators.IsOversold() {
		insights = append(insights, "Stock is in oversold territory - potential buying opportunity")
	} else if indicators.RSI > 50 && indicators.RSI < 70 {
		insights = append(insights, "RSI indicates healthy bullish momentum")
	}

	// MACD insights
	if indicators.MACD > indicators.MACDSignal && indicators.MACDHistogram > 0 {
		insights = append(insights, "MACD shows bullish momentum with increasing strength")
	} else if indicators.MACD < indicators.MACDSignal && indicators.MACDHistogram < 0 {
		insights = append(insights, "MACD shows bearish momentum with decreasing strength")
	}

	// Moving average insights
	if indicators.SMA20 > indicators.SMA50 && indicators.SMA50 > indicators.SMA200 {
		insights = append(insights, "All moving averages align bullishly - strong uptrend")
	} else if indicators.SMA20 < indicators.SMA50 && indicators.SMA50 < indicators.SMA200 {
		insights = append(insights, "All moving averages align bearishly - strong downtrend")
	}

	// Volume insights
	if indicators.OBV > 0 {
		insights = append(insights, "Volume flow supports current price direction")
	} else {
		insights = append(insights, "Volume flow diverges from price - potential reversal signal")
	}

	// Bollinger Bands insights
	if indicators.BBPercentB > 0.8 {
		insights = append(insights, "Price near upper Bollinger Band - potential resistance")
	} else if indicators.BBPercentB < 0.2 {
		insights = append(insights, "Price near lower Bollinger Band - potential support")
	}

	// Volatility insights
	if indicators.BandWidth < 0.1 {
		insights = append(insights, "Low volatility (Bollinger squeeze) - potential breakout ahead")
	} else if indicators.ATR > 0 {
		insights = append(insights, "Higher volatility indicates increased trading activity")
	}

	// Support/Resistance insights
	if indicators.Support1 > 0 && indicators.Resistance1 > 0 {
		insights = append(insights, fmt.Sprintf("Key support at %.2f, resistance at %.2f",
			indicators.Support1, indicators.Resistance1))
	}

	return insights
}
