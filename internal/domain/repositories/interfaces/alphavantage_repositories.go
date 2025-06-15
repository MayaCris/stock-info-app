package interfaces

import (
	"context"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/google/uuid"
)

// FinancialMetricsRepository defines the interface for financial metrics data access
type FinancialMetricsRepository interface {
	// CRUD Operations
	Create(ctx context.Context, metrics *entities.FinancialMetrics) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.FinancialMetrics, error)
	GetBySymbol(ctx context.Context, symbol string) (*entities.FinancialMetrics, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*entities.FinancialMetrics, error)
	Update(ctx context.Context, metrics *entities.FinancialMetrics) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query Operations
	List(ctx context.Context, limit, offset int) ([]*entities.FinancialMetrics, error)
	GetBySymbols(ctx context.Context, symbols []string) ([]*entities.FinancialMetrics, error)
	GetBySector(ctx context.Context, sector string) ([]*entities.FinancialMetrics, error)
	GetByIndustry(ctx context.Context, industry string) ([]*entities.FinancialMetrics, error)

	// Filtering and Analysis
	GetByPERatio(ctx context.Context, minPE, maxPE float64) ([]*entities.FinancialMetrics, error)
	GetByROE(ctx context.Context, minROE float64) ([]*entities.FinancialMetrics, error)
	GetByMarketCap(ctx context.Context, minCap, maxCap int64) ([]*entities.FinancialMetrics, error)
	GetByGrowthRate(ctx context.Context, minGrowth float64) ([]*entities.FinancialMetrics, error)
	GetByDebtToEquity(ctx context.Context, maxDebtToEquity float64) ([]*entities.FinancialMetrics, error)

	// Screening and Rankings
	GetTopByROE(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error)
	GetTopByGrowth(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error)
	GetValueStocks(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error)
	GetGrowthStocks(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error)
	GetDividendStocks(ctx context.Context, minYield float64, limit int) ([]*entities.FinancialMetrics, error)

	// Data Freshness
	GetStaleData(ctx context.Context, olderThan time.Time) ([]*entities.FinancialMetrics, error)
	GetLastUpdated(ctx context.Context, symbol string) (time.Time, error)

	// Bulk Operations
	BulkCreate(ctx context.Context, metrics []*entities.FinancialMetrics) error
	BulkUpdate(ctx context.Context, metrics []*entities.FinancialMetrics) error

	// Statistics
	GetSectorAverages(ctx context.Context, sector string) (map[string]float64, error)
	GetIndustryAverages(ctx context.Context, industry string) (map[string]float64, error)
	Count(ctx context.Context) (int64, error)
}

// TechnicalIndicatorsRepository defines the interface for technical indicators data access
type TechnicalIndicatorsRepository interface {
	// CRUD Operations
	Create(ctx context.Context, indicators *entities.TechnicalIndicators) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.TechnicalIndicators, error)
	GetBySymbol(ctx context.Context, symbol string) (*entities.TechnicalIndicators, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*entities.TechnicalIndicators, error)
	Update(ctx context.Context, indicators *entities.TechnicalIndicators) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query Operations
	List(ctx context.Context, limit, offset int) ([]*entities.TechnicalIndicators, error)
	GetBySymbols(ctx context.Context, symbols []string) ([]*entities.TechnicalIndicators, error)
	GetByTimeFrame(ctx context.Context, timeFrame string) ([]*entities.TechnicalIndicators, error)
	GetBySignal(ctx context.Context, signal string) ([]*entities.TechnicalIndicators, error)

	// Technical Analysis Filtering
	GetByRSI(ctx context.Context, minRSI, maxRSI float64) ([]*entities.TechnicalIndicators, error)
	GetOverboughtStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error)
	GetOversoldStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error)
	GetBullishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error)
	GetBearishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error)

	// MACD Analysis
	GetMACDBullish(ctx context.Context) ([]*entities.TechnicalIndicators, error)
	GetMACDBearish(ctx context.Context) ([]*entities.TechnicalIndicators, error)

	// Moving Average Analysis
	GetAboveMA(ctx context.Context, period int) ([]*entities.TechnicalIndicators, error)
	GetBelowMA(ctx context.Context, period int) ([]*entities.TechnicalIndicators, error)
	GetGoldenCross(ctx context.Context) ([]*entities.TechnicalIndicators, error)
	GetDeathCross(ctx context.Context) ([]*entities.TechnicalIndicators, error)

	// Volume Analysis
	GetHighVolumeStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error)
	GetVolumeBreakout(ctx context.Context) ([]*entities.TechnicalIndicators, error)

	// Volatility Analysis
	GetHighVolatility(ctx context.Context, threshold float64) ([]*entities.TechnicalIndicators, error)
	GetLowVolatility(ctx context.Context, threshold float64) ([]*entities.TechnicalIndicators, error)

	// Bollinger Bands Analysis
	GetBollingerBreakout(ctx context.Context) ([]*entities.TechnicalIndicators, error)
	GetBollingerSqueeze(ctx context.Context) ([]*entities.TechnicalIndicators, error)

	// Rankings
	GetTopByScore(ctx context.Context, limit int) ([]*entities.TechnicalIndicators, error)
	GetStrongestSignals(ctx context.Context, limit int) ([]*entities.TechnicalIndicators, error)

	// Data Freshness
	GetStaleData(ctx context.Context, olderThan time.Time) ([]*entities.TechnicalIndicators, error)
	GetLastUpdated(ctx context.Context, symbol string) (time.Time, error)

	// Bulk Operations
	BulkCreate(ctx context.Context, indicators []*entities.TechnicalIndicators) error
	BulkUpdate(ctx context.Context, indicators []*entities.TechnicalIndicators) error

	// Statistics
	Count(ctx context.Context) (int64, error)
}

// HistoricalDataRepository defines the interface for historical data access
type HistoricalDataRepository interface {
	// CRUD Operations
	Create(ctx context.Context, data *entities.HistoricalData) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.HistoricalData, error)
	GetBySymbolAndDate(ctx context.Context, symbol string, date time.Time) (*entities.HistoricalData, error)
	Update(ctx context.Context, data *entities.HistoricalData) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Historical Data Queries
	GetBySymbol(ctx context.Context, symbol string, startDate, endDate time.Time) ([]*entities.HistoricalData, error)
	GetBySymbolLastN(ctx context.Context, symbol string, days int) ([]*entities.HistoricalData, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID, startDate, endDate time.Time) ([]*entities.HistoricalData, error)
	GetByTimeFrame(ctx context.Context, symbol string, timeFrame string, startDate, endDate time.Time) ([]*entities.HistoricalData, error)

	// Price Analysis
	GetHighestPrice(ctx context.Context, symbol string, startDate, endDate time.Time) (*entities.HistoricalData, error)
	GetLowestPrice(ctx context.Context, symbol string, startDate, endDate time.Time) (*entities.HistoricalData, error)
	GetLatestPrice(ctx context.Context, symbol string) (*entities.HistoricalData, error)
	GetPriceRange(ctx context.Context, symbol string, startDate, endDate time.Time) (float64, float64, error)

	// Volume Analysis
	GetHighestVolume(ctx context.Context, symbol string, startDate, endDate time.Time) (*entities.HistoricalData, error)
	GetAverageVolume(ctx context.Context, symbol string, startDate, endDate time.Time) (int64, error)
	GetVolumeSpikes(ctx context.Context, symbol string, multiplier float64, days int) ([]*entities.HistoricalData, error)

	// Pattern Detection
	GetGaps(ctx context.Context, symbol string, minGapPercent float64, days int) ([]*entities.HistoricalData, error)
	GetBreakouts(ctx context.Context, symbol string, days int) ([]*entities.HistoricalData, error)
	GetBreakdowns(ctx context.Context, symbol string, days int) ([]*entities.HistoricalData, error)

	// Statistical Analysis
	GetReturns(ctx context.Context, symbol string, startDate, endDate time.Time) ([]float64, error)
	GetVolatility(ctx context.Context, symbol string, days int) (float64, error)
	GetCorrelation(ctx context.Context, symbol1, symbol2 string, days int) (float64, error)

	// Moving Averages (calculated on-the-fly)
	GetSMA(ctx context.Context, symbol string, period int, date time.Time) (float64, error)
	GetEMA(ctx context.Context, symbol string, period int, date time.Time) (float64, error)

	// Data Quality
	GetMissingDates(ctx context.Context, symbol string, startDate, endDate time.Time) ([]time.Time, error)
	GetDataGaps(ctx context.Context, symbol string, maxGapDays int) ([]time.Time, error)
	ValidateDataIntegrity(ctx context.Context, symbol string) error

	// Bulk Operations
	BulkCreate(ctx context.Context, data []*entities.HistoricalData) error
	BulkUpdate(ctx context.Context, data []*entities.HistoricalData) error
	DeleteBySymbolAndDateRange(ctx context.Context, symbol string, startDate, endDate time.Time) error

	// Pagination and Limits
	List(ctx context.Context, limit, offset int) ([]*entities.HistoricalData, error)
	Count(ctx context.Context) (int64, error)
	CountBySymbol(ctx context.Context, symbol string) (int64, error)
}

// HistoricalDataSummaryRepository defines the interface for historical data summary access
type HistoricalDataSummaryRepository interface {
	// CRUD Operations
	Create(ctx context.Context, summary *entities.HistoricalDataSummary) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.HistoricalDataSummary, error)
	GetBySymbolAndPeriod(ctx context.Context, symbol, period string) (*entities.HistoricalDataSummary, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entities.HistoricalDataSummary, error)
	Update(ctx context.Context, summary *entities.HistoricalDataSummary) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Summary Queries
	GetBySymbol(ctx context.Context, symbol string) ([]*entities.HistoricalDataSummary, error)
	GetByPeriod(ctx context.Context, period string) ([]*entities.HistoricalDataSummary, error)
	GetRecentSummaries(ctx context.Context, symbol string, limit int) ([]*entities.HistoricalDataSummary, error)

	// Performance Analysis
	GetTopPerformers(ctx context.Context, period string, limit int) ([]*entities.HistoricalDataSummary, error)
	GetWorstPerformers(ctx context.Context, period string, limit int) ([]*entities.HistoricalDataSummary, error)
	GetByPerformance(ctx context.Context, period string, minReturn, maxReturn float64) ([]*entities.HistoricalDataSummary, error)

	// Risk Analysis
	GetByRiskLevel(ctx context.Context, riskLevel string) ([]*entities.HistoricalDataSummary, error)
	GetByVolatility(ctx context.Context, minVol, maxVol float64) ([]*entities.HistoricalDataSummary, error)
	GetByBeta(ctx context.Context, minBeta, maxBeta float64) ([]*entities.HistoricalDataSummary, error)
	GetLowVolatilityStocks(ctx context.Context, period string, limit int) ([]*entities.HistoricalDataSummary, error)

	// Statistical Queries
	GetBySharpeRatio(ctx context.Context, minSharpe float64) ([]*entities.HistoricalDataSummary, error)
	GetByMaxDrawdown(ctx context.Context, maxDrawdown float64) ([]*entities.HistoricalDataSummary, error)
	GetByWinRate(ctx context.Context, minWinRate float64) ([]*entities.HistoricalDataSummary, error)

	// Data Maintenance
	GetStaleData(ctx context.Context, olderThan time.Time) ([]*entities.HistoricalDataSummary, error)
	RefreshSummary(ctx context.Context, symbol, period string) error

	// Bulk Operations
	BulkCreate(ctx context.Context, summaries []*entities.HistoricalDataSummary) error
	BulkUpdate(ctx context.Context, summaries []*entities.HistoricalDataSummary) error

	// Statistics
	List(ctx context.Context, limit, offset int) ([]*entities.HistoricalDataSummary, error)
	Count(ctx context.Context) (int64, error)
}
