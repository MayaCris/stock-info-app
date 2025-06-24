package interfaces

import (
	"context"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
)

// MarketDataService defines the interface for market data operations
type MarketDataService interface {
	// Real-time data
	GetRealTimeQuote(ctx context.Context, symbol string) (*response.MarketDataResponse, error)

	// Company information
	GetCompanyProfile(ctx context.Context, symbol string) (*response.CompanyProfileResponse, error)

	// News and sentiment
	GetCompanyNews(ctx context.Context, symbol string, days int) ([]*response.NewsResponse, error)

	// Financial data
	GetBasicFinancials(ctx context.Context, symbol string) (*response.BasicFinancialsResponse, error)

	// Market overview
	GetMarketOverview(ctx context.Context) (*response.MarketOverviewResponse, error)

	// Bulk operations
	RefreshMarketData(ctx context.Context, symbols []string) error

	// Alpha Vantage specific methods
	GetHistoricalData(ctx context.Context, symbol, period, outputSize string) (*response.HistoricalDataResponse, error)
	GetTechnicalIndicators(ctx context.Context, symbol, indicator, interval, timePeriod string) (*response.TechnicalIndicatorsResponse, error)
	GetFundamentalData(ctx context.Context, symbol string) (*response.FundamentalDataResponse, error)
	GetEarningsData(ctx context.Context, symbol string) (*response.EarningsDataResponse, error)
	AlphaVantageHealthCheck(ctx context.Context) (bool, error)
}
