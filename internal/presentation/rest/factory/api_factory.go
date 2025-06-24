package factory

import (
	"fmt"

	"github.com/MayaCris/stock-info-app/internal/application/services"
	serviceInterfaces "github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/implementation"
	domainServices "github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cache"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	infraFactory "github.com/MayaCris/stock-info-app/internal/infrastructure/factory"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// APIFactory crea instancias de servicios y dependencias para handlers REST
type APIFactory struct {
	config *config.Config
	// Cached dependencies for reuse
	serviceFactory *services.ServiceFactory
	dependencies   *Dependencies
}

// NewAPIFactory crea una nueva factory para la API
func NewAPIFactory(cfg *config.Config) *APIFactory {
	return &APIFactory{
		config: cfg,
	}
}

// Dependencies representa todas las dependencias necesarias para los handlers
type Dependencies struct {
	CompanyService      serviceInterfaces.CompanyService
	BrokerageService    serviceInterfaces.BrokerageService
	StockService        serviceInterfaces.StockRatingService
	AnalysisService     serviceInterfaces.AnalysisService
	MarketDataService   serviceInterfaces.MarketDataService
	AlphaVantageService serviceInterfaces.AlphaVantageService
	Logger              logger.Logger
	CacheService        domainServices.CacheService
	TransactionService  domainServices.TransactionService
}

// CreateDependencies crea todas las dependencias necesarias para los handlers
func (f *APIFactory) CreateDependencies() (*Dependencies, error) {
	if f.dependencies != nil {
		return f.dependencies, nil
	}

	// 1. Database connection
	db, err := cockroachdb.NewConnection(f.config)
	if err != nil {
		return nil, err
	}

	// 2. Transaction service
	transactionService := domainServices.NewTransactionService(db.DB)
	// 3. Repositories
	companyRepo := implementation.NewCompanyRepository(db.DB)
	brokerageRepo := implementation.NewBrokerageRepository(db.DB)
	stockRatingRepo := implementation.NewStockRatingRepository(db.DB)
	// Market data repositories
	marketDataRepo := implementation.NewMarketDataRepository(db.DB)
	companyProfileRepo := implementation.NewCompanyProfileRepository(db.DB)
	newsRepo := implementation.NewNewsRepository(db.DB)
	basicFinancialsRepo := implementation.NewBasicFinancialsRepository(db.DB)

	// Alpha Vantage specific repositories
	historicalDataRepo := implementation.NewHistoricalDataRepository(db.DB)
	financialMetricsRepo := implementation.NewFinancialMetricsRepository(db.DB)
	technicalIndicatorsRepo := implementation.NewTechnicalIndicatorsRepository(db.DB)

	// 4. Cache service
	var cacheService domainServices.CacheService
	if f.config.Cache.Host != "" {
		cacheService = cache.NewCacheService(f.config)
	}

	// 5. Logger
	appLogger, err := logger.InitializeGlobalLogger()
	if err != nil {
		return nil, err
	}
	// 6. Create market data service using market data factory
	marketDataFactory := infraFactory.NewMarketDataFactory(infraFactory.MarketDataFactoryConfig{
		Config:              f.config,
		Logger:              appLogger,
		MarketDataRepo:      marketDataRepo,
		CompanyProfileRepo:  companyProfileRepo,
		NewsRepo:            newsRepo,
		BasicFinancialsRepo: basicFinancialsRepo,
		CompanyRepo:         companyRepo,
	})
	marketDataService := marketDataFactory.CreateMarketDataService()
	// 7. Service factory with Alpha Vantage components
	if f.serviceFactory == nil {
		f.serviceFactory = services.NewServiceFactory(services.ServiceFactoryConfig{
			CompanyRepo:             companyRepo,
			BrokerageRepo:           brokerageRepo,
			StockRatingRepo:         stockRatingRepo,
			HistoricalDataRepo:      historicalDataRepo,
			FinancialMetricsRepo:    financialMetricsRepo,
			TechnicalIndicatorsRepo: technicalIndicatorsRepo,
			AlphaVantageClient:      marketDataFactory.GetAlphaVantageClient(),
			AlphaVantageAdapter:     marketDataFactory.GetAlphaVantageAdapter(),
			Logger:                  appLogger,
		})
	}
	// 8. Create services using factory methods
	companyService := f.serviceFactory.GetCompanyService()
	brokerageService := f.serviceFactory.GetBrokerageService()
	stockService := f.serviceFactory.GetStockRatingService()
	analysisService := f.serviceFactory.GetAnalysisService()

	// 9. Create Alpha Vantage service using service factory
	alphaVantageService := f.serviceFactory.GetAlphaVantageService()

	// 10. Cache dependencies
	f.dependencies = &Dependencies{
		CompanyService:      companyService,
		BrokerageService:    brokerageService,
		StockService:        stockService,
		AnalysisService:     analysisService,
		MarketDataService:   marketDataService,
		AlphaVantageService: alphaVantageService,
		Logger:              appLogger,
		CacheService:        cacheService,
		TransactionService:  transactionService,
	}

	return f.dependencies, nil
}

// GetCompanyService retorna el servicio de companies
func (f *APIFactory) GetCompanyService() (serviceInterfaces.CompanyService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.CompanyService, nil
}

// GetBrokerageService retorna el servicio de brokerages
func (f *APIFactory) GetBrokerageService() (serviceInterfaces.BrokerageService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.BrokerageService, nil
}

// GetStockService retorna el servicio de stock ratings
func (f *APIFactory) GetStockService() (serviceInterfaces.StockRatingService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.StockService, nil
}

// GetMarketDataService retorna el servicio de market data
func (f *APIFactory) GetMarketDataService() (serviceInterfaces.MarketDataService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.MarketDataService, nil
}

// GetAnalysisService retorna el servicio de análisis
func (f *APIFactory) GetAnalysisService() (serviceInterfaces.AnalysisService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.AnalysisService, nil
}

// GetLogger retorna el logger configurado
func (f *APIFactory) GetLogger() (logger.Logger, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.Logger, nil
}

// GetCacheService retorna el servicio de cache
func (f *APIFactory) GetCacheService() (domainServices.CacheService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.CacheService, nil
}

// GetTransactionService retorna el servicio de transacciones
func (f *APIFactory) GetTransactionService() (domainServices.TransactionService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.TransactionService, nil
}

// GetAlphaVantageService retorna el servicio de Alpha Vantage
func (f *APIFactory) GetAlphaVantageService() (serviceInterfaces.AlphaVantageService, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return deps.AlphaVantageService, nil
}

// GetAllServices retorna todos los servicios principales
func (f *APIFactory) GetAllServices() (*APIServices, error) {
	deps, err := f.CreateDependencies()
	if err != nil {
		return nil, err
	}
	return &APIServices{
		Company:      deps.CompanyService,
		Brokerage:    deps.BrokerageService,
		Stock:        deps.StockService,
		Analysis:     deps.AnalysisService,
		MarketData:   deps.MarketDataService,
		AlphaVantage: deps.AlphaVantageService,
	}, nil
}

// APIServices contiene todos los servicios principales de la API
type APIServices struct {
	Company      serviceInterfaces.CompanyService
	Brokerage    serviceInterfaces.BrokerageService
	Stock        serviceInterfaces.StockRatingService
	Analysis     serviceInterfaces.AnalysisService
	MarketData   serviceInterfaces.MarketDataService
	AlphaVantage serviceInterfaces.AlphaVantageService
}

// Cleanup libera recursos de la factory
func (f *APIFactory) Cleanup() error {
	// Reset cached dependencies to force recreation on next use
	f.dependencies = nil
	f.serviceFactory = nil

	return nil
}

// GetConfig retorna la configuración actual
func (f *APIFactory) GetConfig() *config.Config {
	return f.config
}

// UpdateConfig actualiza la configuración y limpia la cache
func (f *APIFactory) UpdateConfig(newConfig *config.Config) error {
	f.config = newConfig
	return f.Cleanup()
}

// ValidateConfiguration valida que todas las configuraciones necesarias estén presentes
func (f *APIFactory) ValidateConfiguration() error {
	if f.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate database configuration
	if f.config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if f.config.Database.Port == "" {
		return fmt.Errorf("database port is required")
	}

	if f.config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Note: Cache configuration is optional

	return nil
}
