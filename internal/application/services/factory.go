package services

import (
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/alphavantage"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// ServiceFactory creates and manages service instances
type ServiceFactory struct {
	// Repositories
	stockRatingRepo         repoInterfaces.StockRatingRepository
	companyRepo             repoInterfaces.CompanyRepository
	brokerageRepo           repoInterfaces.BrokerageRepository
	financialMetricsRepo    repoInterfaces.FinancialMetricsRepository
	technicalIndicatorsRepo repoInterfaces.TechnicalIndicatorsRepository
	historicalDataRepo      repoInterfaces.HistoricalDataRepository

	// External clients
	alphaVantageClient  *alphavantage.Client
	alphaVantageAdapter *alphavantage.Adapter

	// Services (lazy initialization)
	stockService               interfaces.StockRatingService
	companyService             interfaces.CompanyService
	brokerageService           interfaces.BrokerageService
	analysisService            interfaces.AnalysisService
	financialMetricsService    *FinancialMetricsService
	technicalIndicatorsService *TechnicalIndicatorsService
	alphaVantageService        interfaces.AlphaVantageService

	// Infrastructure
	logger logger.Logger
}

// ServiceFactoryConfig holds configuration for service factory
type ServiceFactoryConfig struct {
	StockRatingRepo         repoInterfaces.StockRatingRepository
	CompanyRepo             repoInterfaces.CompanyRepository
	BrokerageRepo           repoInterfaces.BrokerageRepository
	FinancialMetricsRepo    repoInterfaces.FinancialMetricsRepository
	TechnicalIndicatorsRepo repoInterfaces.TechnicalIndicatorsRepository
	HistoricalDataRepo      repoInterfaces.HistoricalDataRepository
	AlphaVantageClient      *alphavantage.Client
	AlphaVantageAdapter     *alphavantage.Adapter
	Logger                  logger.Logger
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(config ServiceFactoryConfig) *ServiceFactory {
	return &ServiceFactory{
		stockRatingRepo:         config.StockRatingRepo,
		companyRepo:             config.CompanyRepo,
		brokerageRepo:           config.BrokerageRepo,
		financialMetricsRepo:    config.FinancialMetricsRepo,
		technicalIndicatorsRepo: config.TechnicalIndicatorsRepo,
		historicalDataRepo:      config.HistoricalDataRepo,
		alphaVantageClient:      config.AlphaVantageClient,
		alphaVantageAdapter:     config.AlphaVantageAdapter,
		logger:                  config.Logger,
	}
}

// GetStockRatingService returns the stock rating service instance
func (f *ServiceFactory) GetStockRatingService() interfaces.StockRatingService {
	if f.stockService == nil {
		f.stockService = NewStockRatingService(
			f.stockRatingRepo,
			f.companyRepo,
			f.brokerageRepo,
			f.logger,
		)
	}
	return f.stockService
}

// GetCompanyService returns the company service instance
func (f *ServiceFactory) GetCompanyService() interfaces.CompanyService {
	if f.companyService == nil {
		f.companyService = NewCompanyService(
			f.companyRepo,
			f.logger,
		)
	}
	return f.companyService
}

// GetBrokerageService returns the brokerage service instance
func (f *ServiceFactory) GetBrokerageService() interfaces.BrokerageService {
	if f.brokerageService == nil {
		f.brokerageService = NewBrokerageService(
			f.brokerageRepo,
			f.logger,
		)
	}
	return f.brokerageService
}

// GetAnalysisService returns the analysis service instance
func (f *ServiceFactory) GetAnalysisService() interfaces.AnalysisService {
	if f.analysisService == nil {
		f.analysisService = NewAnalysisService(
			f.stockRatingRepo,
			f.companyRepo,
			f.brokerageRepo,
			f.logger,
		)
	}
	return f.analysisService
}

// GetFinancialMetricsService returns the financial metrics service instance
func (f *ServiceFactory) GetFinancialMetricsService() *FinancialMetricsService {
	if f.financialMetricsService == nil {
		f.financialMetricsService = NewFinancialMetricsService(
			f.financialMetricsRepo,
			f.companyRepo,
		)
	}
	return f.financialMetricsService
}

// GetTechnicalIndicatorsService returns the technical indicators service instance
func (f *ServiceFactory) GetTechnicalIndicatorsService() *TechnicalIndicatorsService {
	if f.technicalIndicatorsService == nil {
		f.technicalIndicatorsService = NewTechnicalIndicatorsService(
			f.technicalIndicatorsRepo,
			f.companyRepo,
		)
	}
	return f.technicalIndicatorsService
}

// GetAlphaVantageService returns the Alpha Vantage service instance
func (f *ServiceFactory) GetAlphaVantageService() interfaces.AlphaVantageService {
	if f.alphaVantageService == nil {
		f.alphaVantageService = NewAlphaVantageService(
			f.alphaVantageClient,
			f.alphaVantageAdapter,
			f.financialMetricsRepo,
			f.technicalIndicatorsRepo,
			f.historicalDataRepo,
			f.companyRepo,
			f.logger,
		)
	}
	return f.alphaVantageService
}

// GetAllServices returns all service instances
func (f *ServiceFactory) GetAllServices() (
	interfaces.StockRatingService,
	interfaces.CompanyService,
	interfaces.BrokerageService,
	interfaces.AnalysisService,
	*FinancialMetricsService,
	*TechnicalIndicatorsService,
	interfaces.AlphaVantageService,
) {
	return f.GetStockRatingService(),
		f.GetCompanyService(),
		f.GetBrokerageService(),
		f.GetAnalysisService(),
		f.GetFinancialMetricsService(),
		f.GetTechnicalIndicatorsService(),
		f.GetAlphaVantageService()
}

// Reset clears all service instances (useful for testing)
func (f *ServiceFactory) Reset() {
	f.stockService = nil
	f.companyService = nil
	f.brokerageService = nil
	f.analysisService = nil
	f.financialMetricsService = nil
	f.technicalIndicatorsService = nil
	f.alphaVantageService = nil
}
