package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
)

// MarketDataRepository defines the interface for market data operations
type MarketDataRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, marketData *entities.MarketData) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.MarketData, error)
	GetBySymbol(ctx context.Context, symbol string) (*entities.MarketData, error)
	Update(ctx context.Context, marketData *entities.MarketData) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Company-related queries
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*entities.MarketData, error)
	GetByCompanyIDs(ctx context.Context, companyIDs []uuid.UUID) ([]*entities.MarketData, error)

	// Time-based queries
	GetLatest(ctx context.Context, limit int) ([]*entities.MarketData, error)
	GetByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*entities.MarketData, error)
	GetStaleData(ctx context.Context, maxAge time.Duration) ([]*entities.MarketData, error)

	// Market analysis
	GetTopGainers(ctx context.Context, limit int) ([]*entities.MarketData, error)
	GetTopLosers(ctx context.Context, limit int) ([]*entities.MarketData, error)
	GetMostActive(ctx context.Context, limit int) ([]*entities.MarketData, error)

	// Bulk operations
	BulkCreate(ctx context.Context, marketData []*entities.MarketData) error
	BulkUpdate(ctx context.Context, marketData []*entities.MarketData) error
	UpsertBySymbol(ctx context.Context, marketData *entities.MarketData) error

	// Data management
	CleanupOldData(ctx context.Context, olderThan time.Time) (int64, error)
	Count(ctx context.Context) (int64, error)

	// Health check
	Health(ctx context.Context) error
}

// CompanyProfileRepository defines the interface for company profile operations
type CompanyProfileRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, profile *entities.CompanyProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.CompanyProfile, error)
	GetBySymbol(ctx context.Context, symbol string) (*entities.CompanyProfile, error)
	Update(ctx context.Context, profile *entities.CompanyProfile) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and filtering
	GetAll(ctx context.Context, limit, offset int) ([]*entities.CompanyProfile, error)
	GetBySector(ctx context.Context, sector string) ([]*entities.CompanyProfile, error)
	GetByIndustry(ctx context.Context, industry string) ([]*entities.CompanyProfile, error)
	GetByCountry(ctx context.Context, country string) ([]*entities.CompanyProfile, error)

	// Market cap analysis
	GetByMarketCapRange(ctx context.Context, minCap, maxCap int64) ([]*entities.CompanyProfile, error)
	GetLargeCapStocks(ctx context.Context, limit int) ([]*entities.CompanyProfile, error)
	GetSmallCapStocks(ctx context.Context, limit int) ([]*entities.CompanyProfile, error)

	// Data freshness
	GetStaleProfiles(ctx context.Context, maxAge time.Duration) ([]*entities.CompanyProfile, error)
	GetRecentlyUpdated(ctx context.Context, since time.Time) ([]*entities.CompanyProfile, error)

	// Bulk operations
	BulkCreate(ctx context.Context, profiles []*entities.CompanyProfile) error
	BulkUpdate(ctx context.Context, profiles []*entities.CompanyProfile) error
	UpsertBySymbol(ctx context.Context, profile *entities.CompanyProfile) error

	// Statistics
	Count(ctx context.Context) (int64, error)
	GetSectorDistribution(ctx context.Context) (map[string]int64, error)
	GetCountryDistribution(ctx context.Context) (map[string]int64, error)

	// Health check
	Health(ctx context.Context) error
}

// NewsRepository defines the interface for news operations
type NewsRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, news *entities.NewsItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.NewsItem, error)
	Update(ctx context.Context, news *entities.NewsItem) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Symbol-based queries
	GetBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*entities.NewsItem, error)
	GetLatestBySymbol(ctx context.Context, symbol string, limit int) ([]*entities.NewsItem, error)

	// Time-based queries
	GetByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*entities.NewsItem, error)
	GetRecent(ctx context.Context, hours int, limit int) ([]*entities.NewsItem, error)
	GetToday(ctx context.Context, limit int) ([]*entities.NewsItem, error)

	// Category and source filtering
	GetByCategory(ctx context.Context, category string, limit, offset int) ([]*entities.NewsItem, error)
	GetBySource(ctx context.Context, source string, limit, offset int) ([]*entities.NewsItem, error)

	// Sentiment analysis
	GetBySentiment(ctx context.Context, sentiment string, limit, offset int) ([]*entities.NewsItem, error)
	GetPositiveNews(ctx context.Context, limit int) ([]*entities.NewsItem, error)
	GetNegativeNews(ctx context.Context, limit int) ([]*entities.NewsItem, error)

	// Market news
	GetMarketNews(ctx context.Context, limit, offset int) ([]*entities.NewsItem, error)
	GetLatestMarketNews(ctx context.Context, limit int) ([]*entities.NewsItem, error)

	// Bulk operations
	BulkCreate(ctx context.Context, news []*entities.NewsItem) error
	BulkUpdate(ctx context.Context, news []*entities.NewsItem) error

	// Data management
	CleanupOldNews(ctx context.Context, olderThan time.Time) (int64, error)
	RemoveDuplicates(ctx context.Context) (int64, error)

	// Statistics
	Count(ctx context.Context) (int64, error)
	CountBySymbol(ctx context.Context, symbol string) (int64, error)
	GetSentimentDistribution(ctx context.Context, symbol string) (map[string]int64, error)

	// Health check
	Health(ctx context.Context) error
}

// BasicFinancialsRepository defines the interface for basic financials operations
type BasicFinancialsRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, financials *entities.BasicFinancials) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.BasicFinancials, error)
	GetBySymbol(ctx context.Context, symbol string) (*entities.BasicFinancials, error)
	Update(ctx context.Context, financials *entities.BasicFinancials) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Period-based queries
	GetBySymbolAndPeriod(ctx context.Context, symbol string, period string, fiscalYear int) (*entities.BasicFinancials, error)
	GetLatestBySymbol(ctx context.Context, symbol string) (*entities.BasicFinancials, error)
	GetHistoricalBySymbol(ctx context.Context, symbol string, years int) ([]*entities.BasicFinancials, error)

	// Financial analysis
	GetByPERatioRange(ctx context.Context, minPE, maxPE float64) ([]*entities.BasicFinancials, error)
	GetByROERange(ctx context.Context, minROE, maxROE float64) ([]*entities.BasicFinancials, error)
	GetByDebtToEquityRange(ctx context.Context, minDE, maxDE float64) ([]*entities.BasicFinancials, error)

	// Screening and filtering
	GetValueStocks(ctx context.Context, maxPE float64, minROE float64, limit int) ([]*entities.BasicFinancials, error)
	GetGrowthStocks(ctx context.Context, minRevenueGrowth, minEarningsGrowth float64, limit int) ([]*entities.BasicFinancials, error)
	GetDividendStocks(ctx context.Context, minDividendYield float64, limit int) ([]*entities.BasicFinancials, error)

	// Data freshness
	GetStaleFinancials(ctx context.Context, maxAge time.Duration) ([]*entities.BasicFinancials, error)

	// Bulk operations
	BulkCreate(ctx context.Context, financials []*entities.BasicFinancials) error
	BulkUpdate(ctx context.Context, financials []*entities.BasicFinancials) error
	UpsertBySymbol(ctx context.Context, financials *entities.BasicFinancials) error

	// Statistics and analytics
	Count(ctx context.Context) (int64, error)
	GetAverageMetrics(ctx context.Context) (map[string]float64, error)
	GetMetricDistribution(ctx context.Context, metric string) (map[string]int64, error)

	// Health check
	Health(ctx context.Context) error
}

// MarketDataSummary represents aggregated market data
type MarketDataSummary struct {
	TotalStocks    int64     `json:"total_stocks"`
	TotalGainers   int64     `json:"total_gainers"`
	TotalLosers    int64     `json:"total_losers"`
	AvgPriceChange float64   `json:"avg_price_change"`
	TotalVolume    int64     `json:"total_volume"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

// SentimentSummary represents sentiment analysis summary
type SentimentSummary struct {
	PositiveCount int64   `json:"positive_count"`
	NegativeCount int64   `json:"negative_count"`
	NeutralCount  int64   `json:"neutral_count"`
	AvgSentiment  float64 `json:"avg_sentiment"`
	TotalNews     int64   `json:"total_news"`
}
