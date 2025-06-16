package interfaces

import (
	"context"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
)

// AlphaVantageService defines the interface for Alpha Vantage API integration
type AlphaVantageService interface {
	// API Data Fetching Methods
	GetFinancialMetricsFromAPI(ctx context.Context, symbol string) (*entities.FinancialMetrics, error)
	GetTechnicalIndicatorsFromAPI(ctx context.Context, symbol string) ([]*entities.TechnicalIndicators, error)
	GetTechnicalIndicatorFromAPI(ctx context.Context, symbol, indicator, interval, timePeriod, seriesType string) ([]*entities.TechnicalIndicators, error)
	GetHistoricalDataFromAPI(ctx context.Context, symbol, period, outputSize, interval, adjusted string) ([]*entities.HistoricalData, error)

	// Data Management Methods
	RefreshStockData(ctx context.Context, symbol string) error
	BulkRefreshStockData(ctx context.Context, symbols []string) error
	SyncWithDatabase(ctx context.Context, symbol string) error
}
