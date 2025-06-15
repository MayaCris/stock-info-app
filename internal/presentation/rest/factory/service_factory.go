package factory

import (
	"fmt"

	applicationServices "github.com/MayaCris/stock-info-app/internal/application/services"
	serviceInterfaces "github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/implementation"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	domainServices "github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cache"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// ServiceFactory crea instancias de servicios de aplicación para handlers REST
type ServiceFactory struct {
	config *config.Config
	logger logger.Logger
	// Cached dependencies for reuse
	applicationServiceFactory *applicationServices.ServiceFactory
	dependencies              *ServiceDependencies
}

// NewServiceFactory crea una nueva factory para servicios
func NewServiceFactory(cfg *config.Config, appLogger logger.Logger) *ServiceFactory {
	return &ServiceFactory{
		config: cfg,
		logger: appLogger,
	}
}

// ServiceDependencies representa todas las dependencias de servicios necesarias
type ServiceDependencies struct {
	// Application Services
	CompanyService   serviceInterfaces.CompanyService
	BrokerageService serviceInterfaces.BrokerageService
	StockService     serviceInterfaces.StockRatingService
	AnalysisService  serviceInterfaces.AnalysisService

	// Domain Services
	CacheService       domainServices.CacheService
	TransactionService domainServices.TransactionService

	// Infrastructure
	Logger logger.Logger
}

// ServiceConfiguration representa la configuración para la creación de servicios
type ServiceConfiguration struct {
	EnableCache       bool
	EnableTransaction bool
	Environment       string
}

// CreateServiceDependencies crea todas las dependencias de servicios necesarias
func (f *ServiceFactory) CreateServiceDependencies() (*ServiceDependencies, error) {
	if f.dependencies != nil {
		return f.dependencies, nil
	}

	// 1. Database connection
	db, err := cockroachdb.NewConnection(f.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// 2. Create repositories
	companyRepo := implementation.NewCompanyRepository(db.DB)
	brokerageRepo := implementation.NewBrokerageRepository(db.DB)
	stockRatingRepo := implementation.NewStockRatingRepository(db.DB)

	// 3. Create domain services
	transactionService := domainServices.NewTransactionService(db.DB)

	var cacheService domainServices.CacheService
	if f.config.Cache.Host != "" {
		cacheService = cache.NewCacheService(f.config)
	}

	// 4. Create application service factory if not exists
	if f.applicationServiceFactory == nil {
		f.applicationServiceFactory = applicationServices.NewServiceFactory(
			applicationServices.ServiceFactoryConfig{
				StockRatingRepo: stockRatingRepo,
				CompanyRepo:     companyRepo,
				BrokerageRepo:   brokerageRepo,
				Logger:          f.logger,
			},
		)
	}

	// 5. Create application services using the factory
	companyService := f.applicationServiceFactory.GetCompanyService()
	brokerageService := f.applicationServiceFactory.GetBrokerageService()
	stockService := f.applicationServiceFactory.GetStockRatingService()
	analysisService := f.applicationServiceFactory.GetAnalysisService()

	// Cache dependencies
	f.dependencies = &ServiceDependencies{
		CompanyService:     companyService,
		BrokerageService:   brokerageService,
		StockService:       stockService,
		AnalysisService:    analysisService,
		CacheService:       cacheService,
		TransactionService: transactionService,
		Logger:             f.logger,
	}

	return f.dependencies, nil
}

// CreateServiceDependenciesWithConfig crea servicios con configuración personalizada
func (f *ServiceFactory) CreateServiceDependenciesWithConfig(config ServiceConfiguration) (*ServiceDependencies, error) {
	// 1. Database connection
	db, err := cockroachdb.NewConnection(f.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	} // 2. Create repositories based on environment
	var companyRepo repoInterfaces.CompanyRepository
	var brokerageRepo repoInterfaces.BrokerageRepository
	var stockRatingRepo repoInterfaces.StockRatingRepository

	if config.EnableTransaction {
		// Use transactional repositories for production
		companyRepo = implementation.NewTransactionalCompanyRepository(db.DB)
		brokerageRepo = implementation.NewTransactionalBrokerageRepository(db.DB)
		stockRatingRepo = implementation.NewTransactionalStockRatingRepository(db.DB)
	} else {
		// Use regular repositories for development/testing
		companyRepo = implementation.NewCompanyRepository(db.DB)
		brokerageRepo = implementation.NewBrokerageRepository(db.DB)
		stockRatingRepo = implementation.NewStockRatingRepository(db.DB)
	}

	// 3. Create domain services
	transactionService := domainServices.NewTransactionService(db.DB)

	var cacheService domainServices.CacheService
	if config.EnableCache && f.config.Cache.Host != "" {
		cacheService = cache.NewCacheService(f.config)
	}

	// 4. Create application services
	applicationServiceFactory := applicationServices.NewServiceFactory(
		applicationServices.ServiceFactoryConfig{
			StockRatingRepo: stockRatingRepo,
			CompanyRepo:     companyRepo,
			BrokerageRepo:   brokerageRepo,
			Logger:          f.logger,
		},
	)

	return &ServiceDependencies{
		CompanyService:     applicationServiceFactory.GetCompanyService(),
		BrokerageService:   applicationServiceFactory.GetBrokerageService(),
		StockService:       applicationServiceFactory.GetStockRatingService(),
		AnalysisService:    applicationServiceFactory.GetAnalysisService(),
		CacheService:       cacheService,
		TransactionService: transactionService,
		Logger:             f.logger,
	}, nil
}

// GetCompanyService retorna el servicio de companies
func (f *ServiceFactory) GetCompanyService() (serviceInterfaces.CompanyService, error) {
	deps, err := f.CreateServiceDependencies()
	if err != nil {
		return nil, err
	}
	return deps.CompanyService, nil
}

// GetBrokerageService retorna el servicio de brokerages
func (f *ServiceFactory) GetBrokerageService() (serviceInterfaces.BrokerageService, error) {
	deps, err := f.CreateServiceDependencies()
	if err != nil {
		return nil, err
	}
	return deps.BrokerageService, nil
}

// GetStockService retorna el servicio de stocks
func (f *ServiceFactory) GetStockService() (serviceInterfaces.StockRatingService, error) {
	deps, err := f.CreateServiceDependencies()
	if err != nil {
		return nil, err
	}
	return deps.StockService, nil
}

// GetAnalysisService retorna el servicio de análisis
func (f *ServiceFactory) GetAnalysisService() (serviceInterfaces.AnalysisService, error) {
	deps, err := f.CreateServiceDependencies()
	if err != nil {
		return nil, err
	}
	return deps.AnalysisService, nil
}

// GetCacheService retorna el servicio de cache
func (f *ServiceFactory) GetCacheService() (domainServices.CacheService, error) {
	deps, err := f.CreateServiceDependencies()
	if err != nil {
		return nil, err
	}
	return deps.CacheService, nil
}

// GetTransactionService retorna el servicio de transacciones
func (f *ServiceFactory) GetTransactionService() (domainServices.TransactionService, error) {
	deps, err := f.CreateServiceDependencies()
	if err != nil {
		return nil, err
	}
	return deps.TransactionService, nil
}

// CreateDevelopmentServices crea servicios optimizados para desarrollo
func (f *ServiceFactory) CreateDevelopmentServices() (*ServiceDependencies, error) {
	return f.CreateServiceDependenciesWithConfig(ServiceConfiguration{
		EnableCache:       false, // Disable cache for faster development
		EnableTransaction: false, // Use simpler repositories
		Environment:       "development",
	})
}

// CreateProductionServices crea servicios optimizados para producción
func (f *ServiceFactory) CreateProductionServices() (*ServiceDependencies, error) {
	return f.CreateServiceDependenciesWithConfig(ServiceConfiguration{
		EnableCache:       true, // Enable cache for performance
		EnableTransaction: true, // Use transactional repositories
		Environment:       "production",
	})
}

// GetServicesByEnvironment retorna servicios según el entorno
func (f *ServiceFactory) GetServicesByEnvironment() (*ServiceDependencies, error) {
	if f.config.App.IsDevelopment() {
		return f.CreateDevelopmentServices()
	}
	return f.CreateProductionServices()
}

// CreateAllServices crea todos los servicios de aplicación
func (f *ServiceFactory) CreateAllServices() (*ServiceDependencies, error) {
	return f.CreateServiceDependencies()
}

// ResetCache limpia las dependencias en caché (útil para testing)
func (f *ServiceFactory) ResetCache() {
	f.applicationServiceFactory = nil
	f.dependencies = nil
}

// HealthCheck verifica que todos los servicios estén funcionando correctamente
func (f *ServiceFactory) HealthCheck() error {
	deps, err := f.CreateServiceDependencies()
	if err != nil {
		return fmt.Errorf("failed to create service dependencies: %w", err)
	}

	// Basic check that services are created
	if deps.CompanyService == nil {
		return fmt.Errorf("company service is nil")
	}
	if deps.BrokerageService == nil {
		return fmt.Errorf("brokerage service is nil")
	}
	if deps.StockService == nil {
		return fmt.Errorf("stock service is nil")
	}
	if deps.AnalysisService == nil {
		return fmt.Errorf("analysis service is nil")
	}

	return nil
}
