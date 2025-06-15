package services

import (
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// ServiceFactory creates and manages service instances
type ServiceFactory struct {
	// Repositories
	stockRatingRepo repoInterfaces.StockRatingRepository
	companyRepo     repoInterfaces.CompanyRepository
	brokerageRepo   repoInterfaces.BrokerageRepository

	// Services (lazy initialization)
	stockService     interfaces.StockRatingService
	companyService   interfaces.CompanyService
	brokerageService interfaces.BrokerageService
	analysisService  interfaces.AnalysisService

	// Infrastructure
	logger logger.Logger
}

// ServiceFactoryConfig holds configuration for service factory
type ServiceFactoryConfig struct {
	StockRatingRepo repoInterfaces.StockRatingRepository
	CompanyRepo     repoInterfaces.CompanyRepository
	BrokerageRepo   repoInterfaces.BrokerageRepository
	Logger          logger.Logger
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(config ServiceFactoryConfig) *ServiceFactory {
	return &ServiceFactory{
		stockRatingRepo: config.StockRatingRepo,
		companyRepo:     config.CompanyRepo,
		brokerageRepo:   config.BrokerageRepo,
		logger:          config.Logger,
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

// GetAllServices returns all service instances
func (f *ServiceFactory) GetAllServices() (
	interfaces.StockRatingService,
	interfaces.CompanyService,
	interfaces.BrokerageService,
	interfaces.AnalysisService,
) {
	return f.GetStockRatingService(),
		f.GetCompanyService(),
		f.GetBrokerageService(),
		f.GetAnalysisService()
}

// Reset clears all service instances (useful for testing)
func (f *ServiceFactory) Reset() {
	f.stockService = nil
	f.companyService = nil
	f.brokerageService = nil
	f.analysisService = nil
}
