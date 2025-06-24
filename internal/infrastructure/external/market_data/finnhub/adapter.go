package finnhub

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// Adapter converts Finnhub API responses to domain entities
type Adapter struct {
	logger logger.Logger
}

// NewAdapter creates a new Finnhub adapter
func NewAdapter(logger logger.Logger) *Adapter {
	return &Adapter{
		logger: logger,
	}
}

// QuoteToMarketData converts Finnhub QuoteResponse to MarketData entity
func (a *Adapter) QuoteToMarketData(ctx context.Context, quote *QuoteResponse, symbol string, companyID uuid.UUID) (*entities.MarketData, error) {
	if quote == nil {
		return nil, fmt.Errorf("quote response is nil")
	}

	if !quote.IsValid() {
		return nil, fmt.Errorf("invalid quote data")
	}

	marketData := &entities.MarketData{
		ID:              uuid.New(),
		CompanyID:       companyID,
		Symbol:          symbol,
		CurrentPrice:    quote.CurrentPrice,
		OpenPrice:       quote.OpenPrice,
		HighPrice:       quote.HighPrice,
		LowPrice:        quote.LowPrice,
		PreviousClose:   quote.PreviousClose,
		PriceChange:     quote.Change,
		PriceChangePerc: quote.PercentChange,
		Currency:        "USD",
		Exchange:        "US",
		MarketTimestamp: quote.GetTimestamp(),
		IsMarketOpen:    a.isMarketOpenNow(),
	}

	a.logger.Debug(ctx, "Converted quote to market data",
		logger.String("symbol", symbol),
		logger.Float64("price", marketData.CurrentPrice),
	)

	return marketData, nil
}

// ProfileToCompanyProfile converts Finnhub CompanyProfileResponse to CompanyProfile entity
func (a *Adapter) ProfileToCompanyProfile(ctx context.Context, profile *CompanyProfileResponse) (*entities.CompanyProfile, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile response is nil")
	}

	if !profile.IsValid() {
		return nil, fmt.Errorf("invalid profile data")
	}

	ipoDate, err := profile.GetIPODate()
	if err != nil {
		a.logger.Warn(ctx, "Failed to parse IPO date",
			logger.String("symbol", profile.Ticker),
			logger.String("ipo_string", profile.IPO),
		)
		ipoDate = time.Time{} // Set to zero value if parsing fails
	}

	companyProfile := &entities.CompanyProfile{
		ID:                uuid.New(),
		Symbol:            profile.Ticker,
		Name:              profile.Name,
		Industry:          profile.Industry,
		Sector:            profile.Industry, // Map finnhubIndustry to both Industry and Sector
		Country:           profile.Country,
		Currency:          profile.Currency,
		MarketCap:         int64(profile.MarketCap * 1_000_000),        // Convert from millions to actual value
		SharesOutstanding: int64(profile.ShareOutstanding * 1_000_000), // Convert from millions
		Website:           profile.Website,
		Logo:              profile.Logo,
		IPODate:           ipoDate,
		DataSource:        "finnhub",
		LastUpdated:       time.Now(),
	}

	a.logger.Debug(ctx, "Converted profile to company profile",
		logger.String("symbol", profile.Ticker),
		logger.String("name", profile.Name),
	)

	return companyProfile, nil
}

// NewsToNewsItems converts Finnhub NewsResponse to NewsItem entities
func (a *Adapter) NewsToNewsItems(ctx context.Context, news NewsResponse, symbol string) ([]*entities.NewsItem, error) {
	if len(news) == 0 {
		return nil, nil
	}

	newsItems := make([]*entities.NewsItem, 0, len(news))

	for _, item := range news {
		if !item.IsValid() {
			a.logger.Warn(ctx, "Skipping invalid news item",
				logger.String("symbol", symbol),
				logger.String("headline", item.Headline),
			)
			continue
		}

		newsItem := &entities.NewsItem{
			ID:          uuid.New(),
			Symbol:      symbol,
			Title:       item.Headline,
			Summary:     item.Summary,
			URL:         item.URL,
			ImageURL:    item.Image,
			Source:      item.Source,
			Category:    item.Category,
			Language:    "en",
			PublishedAt: item.GetPublishedTime(),
		}

		// Add basic sentiment analysis (can be enhanced later)
		newsItem.SentimentScore, newsItem.SentimentLabel = a.calculateBasicSentiment(item.Headline, item.Summary)

		newsItems = append(newsItems, newsItem)
	}

	a.logger.Debug(ctx, "Converted news to news items",
		logger.String("symbol", symbol),
		logger.Int("total_items", len(news)),
		logger.Int("valid_items", len(newsItems)),
	)

	return newsItems, nil
}

// FinancialsToBasicFinancials converts Finnhub BasicFinancialsResponse to BasicFinancials entity
func (a *Adapter) FinancialsToBasicFinancials(ctx context.Context, financials *BasicFinancialsResponse) (*entities.BasicFinancials, error) {
	if financials == nil {
		return nil, fmt.Errorf("financials response is nil")
	}

	if !financials.IsValid() {
		return nil, fmt.Errorf("invalid financials data")
	}

	basicFinancials := &entities.BasicFinancials{
		ID:     uuid.New(),
		Symbol: financials.Symbol,

		// Valuation Metrics
		PERatio:         financials.Metric.PE,
		PEGRatio:        financials.Metric.PEG,
		PriceToBook:     financials.Metric.PB,
		PriceToSales:    financials.Metric.PS,
		PriceToCashFlow: financials.Metric.PCF,

		// Profitability Metrics
		ROE:             financials.Metric.ROE,
		ROA:             financials.Metric.ROA,
		ROI:             financials.Metric.ROI,
		GrossMargin:     financials.Metric.GrossMargin,
		OperatingMargin: financials.Metric.OperatingMargin,
		NetMargin:       financials.Metric.NetMargin,

		// Growth Metrics
		RevenueGrowth:  financials.Metric.RevenueGrowthTTM,
		EarningsGrowth: financials.Metric.EpsGrowthTTM,

		// Financial Health
		DebtToEquity: financials.Metric.DebtEquity,
		CurrentRatio: financials.Metric.CurrentRatio,
		QuickRatio:   financials.Metric.QuickRatio,
		// Per Share Metrics
		BookValuePerShare: financials.Metric.BookValue,
		CashPerShare:      financials.Metric.CashPerShare,
		DividendPerShare:  financials.Metric.DividendPerShare,

		// Period information
		Period:     "annual",
		FiscalYear: time.Now().Year(),

		// Data source
		DataSource:  "finnhub",
		LastUpdated: time.Now(),
	}

	// Extract EPS from time series data if available
	if len(financials.Series.Annual.SalesPerShare) > 0 {
		// Use latest available data
		latest := financials.Series.Annual.SalesPerShare[len(financials.Series.Annual.SalesPerShare)-1]
		basicFinancials.EPS = latest.V
	}

	a.logger.Debug(ctx, "Converted financials to basic financials",
		logger.String("symbol", financials.Symbol),
		logger.Float64("pe_ratio", basicFinancials.PERatio),
		logger.Float64("roe", basicFinancials.ROE),
	)

	return basicFinancials, nil
}

// MarketNewsToNewsItems converts market news to NewsItem entities
func (a *Adapter) MarketNewsToNewsItems(ctx context.Context, news MarketNewsResponse) ([]*entities.NewsItem, error) {
	if len(news) == 0 {
		return nil, nil
	}

	newsItems := make([]*entities.NewsItem, 0, len(news))

	for _, item := range news {
		newsItem := &entities.NewsItem{
			ID:          uuid.New(),
			Symbol:      "MARKET", // General market news
			Title:       item.Headline,
			Summary:     item.Summary,
			URL:         item.URL,
			ImageURL:    item.Image,
			Source:      item.Source,
			Category:    item.Category,
			Language:    "en",
			PublishedAt: item.GetPublishedTime(),
		}

		// Add basic sentiment analysis
		newsItem.SentimentScore, newsItem.SentimentLabel = a.calculateBasicSentiment(item.Headline, item.Summary)

		newsItems = append(newsItems, newsItem)
	}

	a.logger.Debug(ctx, "Converted market news to news items",
		logger.Int("total_items", len(news)),
		logger.Int("converted_items", len(newsItems)),
	)

	return newsItems, nil
}

// Helper methods

// isMarketOpenNow checks if US market is currently open
func (a *Adapter) isMarketOpenNow() bool {
	now := time.Now().In(time.FixedZone("EST", -5*3600)) // Eastern Time

	// Check if it's a weekday
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}

	// Market hours: 9:30 AM to 4:00 PM EST
	marketOpen := time.Date(now.Year(), now.Month(), now.Day(), 9, 30, 0, 0, now.Location())
	marketClose := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, now.Location())

	return now.After(marketOpen) && now.Before(marketClose)
}

// calculateBasicSentiment provides basic sentiment analysis
// This is a simple implementation - in production, you'd use a proper sentiment analysis service
func (a *Adapter) calculateBasicSentiment(headline, summary string) (float64, string) {
	text := headline + " " + summary

	// Simple keyword-based sentiment analysis
	positiveWords := []string{"gain", "up", "rise", "bull", "positive", "strong", "good", "better", "profit", "growth", "upgrade", "buy"}
	negativeWords := []string{"loss", "down", "fall", "bear", "negative", "weak", "bad", "worse", "decline", "drop", "downgrade", "sell"}

	positiveCount := 0
	negativeCount := 0

	textLower := strings.ToLower(text)

	for _, word := range positiveWords {
		if strings.Contains(textLower, word) {
			positiveCount++
		}
	}

	for _, word := range negativeWords {
		if strings.Contains(textLower, word) {
			negativeCount++
		}
	}

	// Calculate sentiment score (-1 to 1)
	totalWords := positiveCount + negativeCount
	if totalWords == 0 {
		return 0.0, "neutral"
	}

	score := float64(positiveCount-negativeCount) / float64(totalWords)

	// Normalize to [-1, 1] range and determine label
	if score > 0.2 {
		return score, "positive"
	} else if score < -0.2 {
		return score, "negative"
	} else {
		return score, "neutral"
	}
}

// ValidateMarketData validates market data before saving
func (a *Adapter) ValidateMarketData(md *entities.MarketData) error {
	if md.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if md.CurrentPrice <= 0 {
		return fmt.Errorf("current price must be positive")
	}

	if md.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}

	// Check if price change percentage is reasonable (within -50% to +50%)
	if md.PriceChangePerc < -50 || md.PriceChangePerc > 50 {
		a.logger.Warn(context.Background(), "Unusual price change percentage detected",
			logger.String("symbol", md.Symbol),
			logger.Float64("change_perc", md.PriceChangePerc),
		)
	}

	return nil
}

// ValidateCompanyProfile validates company profile before saving
func (a *Adapter) ValidateCompanyProfile(cp *entities.CompanyProfile) error {
	if cp.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if cp.Name == "" {
		return fmt.Errorf("company name is required")
	}

	if cp.MarketCap < 0 {
		return fmt.Errorf("market cap cannot be negative")
	}

	return nil
}
