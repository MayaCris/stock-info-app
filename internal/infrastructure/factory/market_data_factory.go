package factory

import (
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/services"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/alphavantage"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/finnhub"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// MarketDataFactory creates market data related services
type MarketDataFactory struct {
	config *config.Config
	logger logger.Logger

	// Repositories
	marketDataRepo      repoInterfaces.MarketDataRepository
	companyProfileRepo  repoInterfaces.CompanyProfileRepository
	newsRepo            repoInterfaces.NewsRepository
	basicFinancialsRepo repoInterfaces.BasicFinancialsRepository
	companyRepo         repoInterfaces.CompanyRepository

	// External clients
	finnhubClient       *finnhub.Client
	finnhubAdapter      *finnhub.Adapter
	alphavantageClient  *alphavantage.Client
	alphavantageAdapter *alphavantage.Adapter
}

// MarketDataFactoryConfig represents configuration for market data factory
type MarketDataFactoryConfig struct {
	Config              *config.Config
	Logger              logger.Logger
	MarketDataRepo      repoInterfaces.MarketDataRepository
	CompanyProfileRepo  repoInterfaces.CompanyProfileRepository
	NewsRepo            repoInterfaces.NewsRepository
	BasicFinancialsRepo repoInterfaces.BasicFinancialsRepository
	CompanyRepo         repoInterfaces.CompanyRepository
}

// NewMarketDataFactory creates a new market data factory
func NewMarketDataFactory(config MarketDataFactoryConfig) *MarketDataFactory {
	factory := &MarketDataFactory{
		config:              config.Config,
		logger:              config.Logger,
		marketDataRepo:      config.MarketDataRepo,
		companyProfileRepo:  config.CompanyProfileRepo,
		newsRepo:            config.NewsRepo,
		basicFinancialsRepo: config.BasicFinancialsRepo,
		companyRepo:         config.CompanyRepo,
	}

	// Initialize external clients
	factory.initializeFinnhubClient()
	factory.initializeAlphaVantageClient()

	return factory
}

// CreateMarketDataService creates a new market data service
func (f *MarketDataFactory) CreateMarketDataService() interfaces.MarketDataService {
	return services.NewMarketDataService(services.MarketDataServiceConfig{
		MarketDataRepo:      f.marketDataRepo,
		CompanyProfileRepo:  f.companyProfileRepo,
		NewsRepo:            f.newsRepo,
		BasicFinancialsRepo: f.basicFinancialsRepo,
		CompanyRepo:         f.companyRepo,
		FinnhubClient:       f.finnhubClient,
		FinnhubAdapter:      f.finnhubAdapter,
		AlphaVantageClient:  f.alphavantageClient,
		AlphaVantageAdapter: f.alphavantageAdapter,
		Logger:              f.logger,
	})
}

// GetFinnhubClient returns the Finnhub client
func (f *MarketDataFactory) GetFinnhubClient() *finnhub.Client {
	return f.finnhubClient
}

// GetFinnhubAdapter returns the Finnhub adapter
func (f *MarketDataFactory) GetFinnhubAdapter() *finnhub.Adapter {
	return f.finnhubAdapter
}

// GetAlphaVantageClient returns the Alpha Vantage client
func (f *MarketDataFactory) GetAlphaVantageClient() *alphavantage.Client {
	return f.alphavantageClient
}

// GetAlphaVantageAdapter returns the Alpha Vantage adapter
func (f *MarketDataFactory) GetAlphaVantageAdapter() *alphavantage.Adapter {
	return f.alphavantageAdapter
}

// initializeFinnhubClient initializes the Finnhub API client
func (f *MarketDataFactory) initializeFinnhubClient() {
	// Get configuration from environment
	apiKey := f.config.External.Primary.Key
	baseURL := f.config.External.Primary.BaseURL

	if apiKey == "" {
		f.logger.Warn(nil, "Finnhub API key not configured")
	}

	if baseURL == "" {
		baseURL = "https://finnhub.io/api/v1"
	}

	// Create Finnhub client
	f.finnhubClient = finnhub.NewClient(finnhub.ClientConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
		Logger:  f.logger,
	})

	// Create Finnhub adapter
	f.finnhubAdapter = finnhub.NewAdapter(f.logger)

	f.logger.Info(nil, "Finnhub API client initialized",
		logger.String("component", "finnhub_client"))
}

// initializeAlphaVantageClient initializes the Alpha Vantage API client
func (f *MarketDataFactory) initializeAlphaVantageClient() {
	// Get configuration from environment
	apiKey := f.config.External.Secondary.Key
	baseURL := f.config.External.Secondary.BaseURL

	if apiKey == "" {
		f.logger.Warn(nil, "Alpha Vantage API key not configured")
	}

	if baseURL == "" {
		baseURL = "https://www.alphavantage.co/query"
	}

	// Create Alpha Vantage client
	f.alphavantageClient = alphavantage.NewClient(f.config, f.logger)

	// Create Alpha Vantage adapter
	f.alphavantageAdapter = alphavantage.NewAdapter(f.logger)

	f.logger.Info(nil, "Alpha Vantage API client initialized",
		logger.String("component", "alphavantage_client"))
}

// HealthCheck checks the health of external APIs
func (f *MarketDataFactory) HealthCheck() map[string]string {
	results := make(map[string]string)

	// Check Finnhub API
	if f.finnhubClient != nil {
		if err := f.finnhubClient.Health(nil); err != nil {
			results["finnhub"] = "unhealthy: " + err.Error()
		} else {
			results["finnhub"] = "healthy"
		}
	} else {
		results["finnhub"] = "not_configured"
	}
	// Check Alpha Vantage API
	if f.alphavantageClient != nil {
		if err := f.alphavantageClient.HealthCheck(nil); err != nil {
			results["alphavantage"] = "unhealthy: " + err.Error()
		} else {
			results["alphavantage"] = "healthy"
		}
	} else {
		results["alphavantage"] = "not_configured"
	}

	return results
}

// RefreshConfiguration refreshes the configuration and reinitializes clients
func (f *MarketDataFactory) RefreshConfiguration(newConfig *config.Config) {
	f.config = newConfig
	f.initializeFinnhubClient()
	f.initializeAlphaVantageClient()

	f.logger.Info(nil, "Market data factory configuration refreshed")
}
