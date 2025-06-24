package factory

import (
	"fmt"

	"github.com/MayaCris/stock-info-app/internal/application/usecases/population"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/implementation"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/adapters"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cache"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// PopulationUseCaseFactory crea instancias del caso de uso de población
type PopulationUseCaseFactory struct {
	config *config.Config
	// Cached dependencies for reuse
	cachedUseCase      *population.PopulateDatabaseUseCase
	cachedDependencies *PopulationDependencies
}

// PopulationDependencies representa todas las dependencias necesarias para el uso de población
type PopulationDependencies struct {
	// Repositories
	CompanyRepo     interfaces.TransactionalCompanyRepository
	BrokerageRepo   interfaces.TransactionalBrokerageRepository
	StockRatingRepo interfaces.TransactionalStockRatingRepository

	// Services
	CacheService       services.CacheService
	TransactionService services.TransactionService
	IntegrityService   services.IntegrityValidationService

	// External dependencies
	DataProvider     population.StockDataProvider
	PopulationLogger logger.PopulationLogger
	IntegrityLogger  logger.IntegrityLogger
}

// NewPopulationUseCaseFactory crea una nueva factory
func NewPopulationUseCaseFactory(cfg *config.Config) *PopulationUseCaseFactory {
	return &PopulationUseCaseFactory{
		config: cfg,
	}
}

// CreatePopulateDatabaseUseCase crea una instancia completa del caso de uso
func (f *PopulationUseCaseFactory) CreatePopulateDatabaseUseCase() (*population.PopulateDatabaseUseCase, error) {
	if f.cachedUseCase != nil {
		return f.cachedUseCase, nil
	}

	dependencies, err := f.createDependencies(true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create population dependencies: %w", err)
	}

	// Create use case
	useCase := population.NewPopulateDatabaseUseCase(
		dependencies.CompanyRepo,
		dependencies.BrokerageRepo,
		dependencies.StockRatingRepo,
		dependencies.CacheService,
		dependencies.DataProvider,
		dependencies.TransactionService,
		dependencies.IntegrityService,
		dependencies.PopulationLogger,
	)

	// Cache for reuse
	f.cachedUseCase = useCase
	return useCase, nil
}

// createDependencies crea todas las dependencias necesarias para el caso de uso de población
func (f *PopulationUseCaseFactory) createDependencies(enableCache bool, customDataProvider population.StockDataProvider) (*PopulationDependencies, error) {
	if f.cachedDependencies != nil {
		return f.cachedDependencies, nil
	}

	// 1. Database connection
	db, err := cockroachdb.NewConnection(f.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// 2. Transaction service
	transactionService := services.NewTransactionService(db.DB)

	// 3. Repositories (Transactional)
	companyRepo := implementation.NewTransactionalCompanyRepository(db.DB)
	brokerageRepo := implementation.NewTransactionalBrokerageRepository(db.DB)
	stockRatingRepo := implementation.NewTransactionalStockRatingRepository(db.DB)

	// 4. Cache service
	var cacheService services.CacheService
	if enableCache && f.config.Cache.Host != "" {
		cacheService = cache.NewCacheService(f.config)
	}

	// 5. Data provider (custom or default)
	var dataProvider population.StockDataProvider
	if customDataProvider != nil {
		dataProvider = customDataProvider
	} else {
		apiClient := stock_api.NewClient(f.config)
		dataProvider = adapters.NewStockAPIDataProvider(apiClient)
	}

	// 6. Logger setup
	populationLogger, err := logger.InitializePopulationLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize population logger: %w", err)
	}

	// Create base logger for specialized loggers
	baseLogger, err := logger.InitializeGlobalLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize global logger: %w", err)
	}

	// Create specialized loggers
	integrityLogger := logger.NewIntegrityLogger(baseLogger, &logger.LogConfig{})

	// 7. Integrity validation service
	var integrityService services.IntegrityValidationService
	if f.config.App.IsProduction() {
		// Production: Use custom strict configuration
		integrityService = CreateIntegrityValidationServiceWithCustomConfig(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			baseLogger,
			true, // isProduction
		)
	} else {
		// Development: Use default configuration
		integrityService = services.NewIntegrityValidationServiceWithDefaults(
			companyRepo,
			brokerageRepo,
			stockRatingRepo,
			integrityLogger,
		)
	}

	// Cache dependencies
	f.cachedDependencies = &PopulationDependencies{
		CompanyRepo:        companyRepo,
		BrokerageRepo:      brokerageRepo,
		StockRatingRepo:    stockRatingRepo,
		CacheService:       cacheService,
		TransactionService: transactionService,
		IntegrityService:   integrityService,
		DataProvider:       dataProvider,
		PopulationLogger:   populationLogger,
		IntegrityLogger:    integrityLogger,
	}

	return f.cachedDependencies, nil
}

// CreateIntegrityValidationServiceWithCustomConfig creates an integrity validation service with custom configuration
// This is an example of how to use custom validation rules per environment
func CreateIntegrityValidationServiceWithCustomConfig(
	companyRepo interfaces.TransactionalCompanyRepository,
	brokerageRepo interfaces.TransactionalBrokerageRepository,
	stockRatingRepo interfaces.TransactionalStockRatingRepository,
	baseLogger logger.Logger,
	isProduction bool,
) services.IntegrityValidationService {

	integrityLogger := logger.NewIntegrityLogger(baseLogger, &logger.LogConfig{})

	// Create custom configuration based on environment
	config := services.DefaultValidationConfig()

	if isProduction {
		// Production: More strict validation
		config.Rules.Company.ViolationsForCritical = 1    // Stricter for production
		config.Rules.StockRating.MaxAgeYearsBusiness = 15 // Less tolerance for old data
		config.Thresholds.BusinessRulesWarningLimit = 3   // Lower threshold
	} else {
		// Development/Testing: More lenient validation
		config.Rules.Company.ViolationsForCritical = 3    // More lenient
		config.Rules.StockRating.MaxAgeYearsBusiness = 25 // More tolerance
		config.Thresholds.BusinessRulesWarningLimit = 10  // Higher threshold
	}

	return services.NewIntegrityValidationService(
		companyRepo,
		brokerageRepo,
		stockRatingRepo,
		integrityLogger,
		config,
	)
}

// CreatePopulateDatabaseUseCaseWithOptions crea el caso de uso con opciones personalizadas
func (f *PopulationUseCaseFactory) CreatePopulateDatabaseUseCaseWithOptions(
	enableCache bool,
	customDataProvider population.StockDataProvider,
) (*population.PopulateDatabaseUseCase, error) {
	dependencies, err := f.createDependencies(enableCache, customDataProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create population dependencies with options: %w", err)
	}

	// Create use case with custom dependencies (don't cache this variant)
	useCase := population.NewPopulateDatabaseUseCase(
		dependencies.CompanyRepo,
		dependencies.BrokerageRepo,
		dependencies.StockRatingRepo,
		dependencies.CacheService,
		dependencies.DataProvider,
		dependencies.TransactionService,
		dependencies.IntegrityService,
		dependencies.PopulationLogger,
	)

	return useCase, nil
}

// CreateDevelopmentUseCase crea un caso de uso optimizado para desarrollo
func (f *PopulationUseCaseFactory) CreateDevelopmentUseCase() (*population.PopulateDatabaseUseCase, error) {
	return f.CreatePopulateDatabaseUseCaseWithOptions(false, nil) // No cache for development
}

// CreateProductionUseCase crea un caso de uso optimizado para producción
func (f *PopulationUseCaseFactory) CreateProductionUseCase() (*population.PopulateDatabaseUseCase, error) {
	return f.CreatePopulateDatabaseUseCaseWithOptions(true, nil) // With cache for production
}

// GetUseCaseByEnvironment retorna el caso de uso apropiado según el entorno
func (f *PopulationUseCaseFactory) GetUseCaseByEnvironment() (*population.PopulateDatabaseUseCase, error) {
	if f.config.App.IsDevelopment() {
		return f.CreateDevelopmentUseCase()
	}
	return f.CreateProductionUseCase()
}

// GetDependencies retorna las dependencias cacheadas o las crea si no existen
func (f *PopulationUseCaseFactory) GetDependencies() (*PopulationDependencies, error) {
	return f.createDependencies(true, nil)
}

// ResetCache limpia las dependencias en caché (útil para testing)
func (f *PopulationUseCaseFactory) ResetCache() {
	f.cachedUseCase = nil
	f.cachedDependencies = nil
}

// HealthCheck verifica que todas las dependencias estén funcionando correctamente
func (f *PopulationUseCaseFactory) HealthCheck() error {
	dependencies, err := f.createDependencies(true, nil)
	if err != nil {
		return fmt.Errorf("failed to create dependencies for health check: %w", err)
	}

	// Basic check that dependencies are created
	if dependencies.CompanyRepo == nil {
		return fmt.Errorf("company repository is nil")
	}
	if dependencies.BrokerageRepo == nil {
		return fmt.Errorf("brokerage repository is nil")
	}
	if dependencies.StockRatingRepo == nil {
		return fmt.Errorf("stock rating repository is nil")
	}
	if dependencies.TransactionService == nil {
		return fmt.Errorf("transaction service is nil")
	}
	if dependencies.IntegrityService == nil {
		return fmt.Errorf("integrity service is nil")
	}
	if dependencies.DataProvider == nil {
		return fmt.Errorf("data provider is nil")
	}
	if dependencies.PopulationLogger == nil {
		return fmt.Errorf("population logger is nil")
	}

	return nil
}
