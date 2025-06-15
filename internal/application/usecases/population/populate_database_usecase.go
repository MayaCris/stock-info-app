package population

import (
	"context"
	"fmt"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"gorm.io/gorm"
)

// PopulationConfig configura las opciones de población
type PopulationConfig struct {
	BatchSize     int           // Tamaño del lote para procesamiento
	MaxPages      int           // Máximo número de páginas a procesar
	DelayBetween  time.Duration // Delay entre lotes para evitar saturar la API
	ClearFirst    bool          // Si limpiar la base de datos primero
	UseCache      bool          // Si usar cache durante la población
	DryRun        bool          // Solo mostrar qué se haría sin ejecutar
	ValidateAfter bool          // Validar integridad después de la población
}

// PopulationResult contiene los resultados de la población
type PopulationResult struct {
	TotalPages     int // Páginas con datos procesadas
	PagesRequested int // Total de páginas consultadas (incluyendo vacías)
	TotalItems     int
	ProcessedItems int
	SkippedItems   int
	ErrorCount     int
	Companies      int
	Brokerages     int
	StockRatings   int
	Duration       time.Duration
	Errors         []string
}

// StockDataProvider representa cualquier fuente de datos de stock
type StockDataProvider interface {
	FetchPage(ctx context.Context, page string) (*StockDataPage, error)
	GetNextPageToken(currentPage string) string
	HasMorePages(response *StockDataPage) bool
}

// StockDataPage representa una página de datos
type StockDataPage struct {
	Items    []StockDataItem
	NextPage string
	HasMore  bool
}

// StockDataItem representa un item de datos de stock
type StockDataItem struct {
	Ticker     string
	Company    string
	Brokerage  string
	Action     string
	RatingFrom string
	RatingTo   string
	TargetFrom string
	TargetTo   string
	EventTime  time.Time
}

// PopulateDatabaseUseCase implementa el caso de uso de población de base de datos
type PopulateDatabaseUseCase struct {
	companyRepo        interfaces.TransactionalCompanyRepository
	brokerageRepo      interfaces.TransactionalBrokerageRepository
	stockRatingRepo    interfaces.TransactionalStockRatingRepository
	cacheService       services.CacheService
	dataProvider       StockDataProvider
	transactionService services.TransactionService
	integrityService   services.IntegrityValidationService
	logger             logger.PopulationLogger
}

// NewPopulateDatabaseUseCase crea una nueva instancia del caso de uso
func NewPopulateDatabaseUseCase(
	companyRepo interfaces.TransactionalCompanyRepository,
	brokerageRepo interfaces.TransactionalBrokerageRepository,
	stockRatingRepo interfaces.TransactionalStockRatingRepository,
	cacheService services.CacheService,
	dataProvider StockDataProvider,
	transactionService services.TransactionService,
	integrityService services.IntegrityValidationService,
	logger logger.PopulationLogger,
) *PopulateDatabaseUseCase {
	return &PopulateDatabaseUseCase{
		companyRepo:        companyRepo,
		brokerageRepo:      brokerageRepo,
		stockRatingRepo:    stockRatingRepo,
		cacheService:       cacheService,
		dataProvider:       dataProvider,
		transactionService: transactionService,
		integrityService:   integrityService,
		logger:             logger,
	}
}

// Execute ejecuta el caso de uso de población
func (uc *PopulateDatabaseUseCase) Execute(ctx context.Context, config PopulationConfig) (*PopulationResult, error) {
	startTime := time.Now()

	// Convertir config a tipo compatible con logger
	logConfig := logger.PopulationConfig{
		BatchSize:     config.BatchSize,
		MaxPages:      config.MaxPages,
		DelayBetween:  config.DelayBetween,
		ClearFirst:    config.ClearFirst,
		UseCache:      config.UseCache,
		DryRun:        config.DryRun,
		ValidateAfter: config.ValidateAfter,
	}

	uc.logger.LogPopulationStart(ctx, logConfig)

	result := &PopulationResult{
		Errors: make([]string, 0),
	}

	// 1. Clear database if requested
	if config.ClearFirst && !config.DryRun {
		if err := uc.clearDatabase(ctx); err != nil {
			return nil, fmt.Errorf("failed to clear database: %w", err)
		}
		uc.logger.Info(ctx, "🧹 Database cleared successfully", logger.String("operation", "clear_database"))
	}

	// 2. Process pages
	if err := uc.processPages(ctx, config, result); err != nil {
		return nil, fmt.Errorf("failed to process pages: %w", err)
	}

	// 3. Validate after population if requested
	if config.ValidateAfter && !config.DryRun {
		if err := uc.validateIntegrityEnhanced(ctx, result); err != nil {
			uc.logger.Warn(ctx, "⚠️ Validation warnings encountered", logger.ErrorField(err))
		}
	}

	result.Duration = time.Since(startTime)

	// Convertir result a tipo compatible con logger
	logResult := logger.PopulationResult{
		TotalPages:     result.TotalPages,
		PagesRequested: result.PagesRequested,
		TotalItems:     result.TotalItems,
		ProcessedItems: result.ProcessedItems,
		SkippedItems:   result.SkippedItems,
		ErrorCount:     result.ErrorCount,
		Companies:      result.Companies,
		Brokerages:     result.Brokerages,
		StockRatings:   result.StockRatings,
		Duration:       result.Duration,
		Errors:         result.Errors,
	}

	uc.logger.LogPopulationEnd(ctx, logResult, result.Duration)

	return result, nil
}

// processPages procesa todas las páginas de datos
func (uc *PopulateDatabaseUseCase) processPages(ctx context.Context, config PopulationConfig, result *PopulationResult) error {
	currentPage := ""

	for pageNum := 1; pageNum <= config.MaxPages; pageNum++ {
		uc.logger.LogPageProcessing(ctx, pageNum, config.MaxPages,  0)

		// Increment pages requested (including empty ones) - count every page we attempt to fetch
		result.PagesRequested++

		// Fetch data
		dataPage, err := uc.dataProvider.FetchPage(ctx, currentPage)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to fetch page %d: %v", pageNum, err)
			result.Errors = append(result.Errors, errMsg)
			result.ErrorCount++
			uc.logger.Error(ctx, "❌ Failed to fetch page", err,
				logger.Int("page_number", pageNum),
				logger.String("operation", "fetch_page"))
			continue
		}

		if len(dataPage.Items) == 0 {
			uc.logger.Info(ctx, "📄 No more data available",
				logger.Int("page_number", pageNum),
				logger.String("operation", "page_complete"))
			break
		}

		// Update page processing log with actual item count
		uc.logger.LogPageProcessing(ctx, pageNum, config.MaxPages, len(dataPage.Items))

		// Only count pages with data
		result.TotalPages++
		result.TotalItems += len(dataPage.Items)

		if config.DryRun {
			uc.logger.Info(ctx, "🔍 DRY RUN: Would process items",
				logger.Int("item_count", len(dataPage.Items)),
				logger.String("operation", "dry_run"))
			result.ProcessedItems += len(dataPage.Items)
		} else {
			// Process batch
			if err := uc.processBatch(ctx, dataPage.Items, config, result); err != nil {
				errMsg := fmt.Sprintf("Failed to process batch on page %d: %v", pageNum, err)
				result.Errors = append(result.Errors, errMsg)
				result.ErrorCount++
				uc.logger.Error(ctx, "❌ Failed to process batch", err,
					logger.Int("page_number", pageNum),
					logger.String("operation", "process_batch"))
			}
		}

		// Check if there are more pages
		if !dataPage.HasMore {
			break
		}
		currentPage = dataPage.NextPage

		// Delay between pages to avoid overwhelming the API
		if config.DelayBetween > 0 {
			time.Sleep(config.DelayBetween)
		}
	}

	return nil
}

// processBatch procesa un lote de items de forma atómica con transacciones
func (uc *PopulateDatabaseUseCase) processBatch(ctx context.Context, items []StockDataItem, config PopulationConfig, result *PopulationResult) error {
	startTime := time.Now()
	uc.logger.LogBatchProcessing(ctx, len(items), "transactional_batch")

	// Usar el servicio transaccional para garantizar atomicidad
	err := uc.transactionService.ExecuteWithRetry(ctx, 3, func(ctx context.Context) error {
		return uc.transactionService.ExecuteInTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
			// Process companies and brokerages first (to ensure they exist)
			if err := uc.processCompaniesAndBrokeragesTransactional(ctx, tx, items, result); err != nil {
				return fmt.Errorf("failed to process companies and brokerages: %w", err)
			}

			// Then process stock ratings
			if err := uc.processStockRatingsTransactional(ctx, tx, items, result); err != nil {
				return fmt.Errorf("failed to process stock ratings: %w", err)
			}

			return nil
		})
	})

	duration := time.Since(startTime)
	uc.logger.LogTransactionOperation(ctx, "batch_processing", 0, err == nil, duration)

	return err
}

// processCompaniesAndBrokerages procesa companies y brokerages
func (uc *PopulateDatabaseUseCase) processCompaniesAndBrokerages(ctx context.Context, items []StockDataItem, result *PopulationResult) error {
	// Extract unique companies and brokerages
	companies := make(map[string]*entities.Company)
	brokerages := make(map[string]*entities.Brokerage)

	for _, item := range items {
		// Company
		if _, exists := companies[item.Ticker]; !exists {
			companies[item.Ticker] = entities.NewCompany(item.Ticker, item.Company)
		}

		// Brokerage
		if _, exists := brokerages[item.Brokerage]; !exists {
			brokerages[item.Brokerage] = entities.NewBrokerage(item.Brokerage)
		}
	}

	// Save companies
	for ticker, company := range companies {
		// Check if exists
		existing, err := uc.companyRepo.GetByTicker(ctx, ticker)
		if err == nil && existing != nil {
			result.SkippedItems++
			continue
		}

		if err := uc.companyRepo.Create(ctx, company); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create company %s: %v", ticker, err))
			continue
		}

		result.Companies++
		result.ProcessedItems++

		// Cache if enabled
		if uc.cacheService != nil {
			uc.cacheService.SetCompany(ctx, ticker, company, 5*time.Minute)
		}
	}

	// Save brokerages
	for name, brokerage := range brokerages {
		// Check if exists
		existing, err := uc.brokerageRepo.GetByName(ctx, name)
		if err == nil && existing != nil {
			result.SkippedItems++
			continue
		}

		if err := uc.brokerageRepo.Create(ctx, brokerage); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create brokerage %s: %v", name, err))
			continue
		}

		result.Brokerages++
		result.ProcessedItems++

		// Cache if enabled
		if uc.cacheService != nil {
			uc.cacheService.SetBrokerage(ctx, name, brokerage, 5*time.Minute)
		}
	}

	return nil
}

// processStockRatings procesa los stock ratings
func (uc *PopulateDatabaseUseCase) processStockRatings(ctx context.Context, items []StockDataItem, result *PopulationResult) error {
	for _, item := range items {
		// Get company
		company, err := uc.companyRepo.GetByTicker(ctx, item.Ticker)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Company not found for ticker %s: %v", item.Ticker, err))
			continue
		}

		// Get brokerage
		brokerage, err := uc.brokerageRepo.GetByName(ctx, item.Brokerage)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Brokerage not found %s: %v", item.Brokerage, err))
			continue
		}
		// Create stock rating
		stockRating := entities.NewStockRating(
			company.ID,
			brokerage.ID,
			item.Action,
			item.EventTime,
		)

		// Set additional fields
		stockRating.RatingFrom = item.RatingFrom
		stockRating.RatingTo = item.RatingTo
		stockRating.TargetFrom = item.TargetFrom
		stockRating.TargetTo = item.TargetTo

		if err := uc.stockRatingRepo.Create(ctx, stockRating); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create stock rating: %v", err))
			continue
		}

		result.StockRatings++
		result.ProcessedItems++

		// Cache if enabled
		if uc.cacheService != nil {
			uc.cacheService.SetStockRating(ctx, stockRating, 5*time.Minute)
		}
	}

	return nil
}

// clearDatabase limpia la base de datos
func (uc *PopulateDatabaseUseCase) clearDatabase(ctx context.Context) error {
	// Since DeleteAll might not be available, we'll implement a safer approach
	uc.logger.Warn(ctx, "⚠️ Clear database operation not fully implemented - would need DeleteAll methods in repositories",
		logger.String("operation", "clear_database"),
		logger.String("status", "not_implemented"))

	// For now, we'll just clear the cache
	if uc.cacheService != nil {
		uc.cacheService.Clear(ctx)
	}

	return nil
}

// validateIntegrity valida la integridad de los datos
func (uc *PopulateDatabaseUseCase) validateIntegrity(ctx context.Context, result *PopulationResult) error {
	uc.logger.Info(ctx, "🔍 Validating database integrity...",
		logger.String("operation", "integrity_validation"))

	// Get all stock ratings and check for orphaned records
	stockRatings, err := uc.stockRatingRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	orphanedCount := 0
	for _, rating := range stockRatings {
		// Check company exists
		if _, err := uc.companyRepo.GetByID(ctx, rating.CompanyID); err != nil {
			orphanedCount++
		}

		// Check brokerage exists
		if _, err := uc.brokerageRepo.GetByID(ctx, rating.BrokerageID); err != nil {
			orphanedCount++
		}
	}

	if orphanedCount > 0 {
		return fmt.Errorf("found %d orphaned stock rating records", orphanedCount)
	}

	uc.logger.Info(ctx, "✅ Database integrity validation passed",
		logger.String("operation", "integrity_validation"),
		logger.String("status", "passed"))
	return nil
}

// logResults registra los resultados finales
// ========================================
// TRANSACTIONAL BATCH PROCESSING METHODS
// ========================================

// processCompaniesAndBrokeragesTransactional procesa companies y brokerages usando transacciones
func (uc *PopulateDatabaseUseCase) processCompaniesAndBrokeragesTransactional(ctx context.Context, tx *gorm.DB, items []StockDataItem, result *PopulationResult) error {
	// Extract unique companies and brokerages
	companies := make(map[string]*entities.Company)
	brokerages := make(map[string]*entities.Brokerage)

	for _, item := range items {
		// Company
		if _, exists := companies[item.Ticker]; !exists {
			companies[item.Ticker] = entities.NewCompany(item.Ticker, item.Company)
		}

		// Brokerage
		if _, exists := brokerages[item.Brokerage]; !exists {
			brokerages[item.Brokerage] = entities.NewBrokerage(item.Brokerage)
		}
	}

	uc.logger.Debug(ctx, "Processing entities in transaction",
		logger.String("operation", "process_entities_tx"),
		logger.Int("unique_companies", len(companies)),
		logger.Int("unique_brokerages", len(brokerages)))
	// Process companies using transaction with duplicate handling
	for ticker, company := range companies {
		// Use CreateIgnoreDuplicatesWithTx to avoid transaction aborts on duplicates
		createdOrExisting, err := uc.companyRepo.CreateIgnoreDuplicatesWithTx(ctx, tx, company)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create company %s: %v", ticker, err))
			uc.logger.LogEntityError(ctx, "company", ticker, err)
			continue
		}

		// Check if it was created or already existed
		if createdOrExisting.ID == company.ID {
			// New company was created
			result.Companies++
			result.ProcessedItems++
			uc.logger.LogEntityCreated(ctx, "company", ticker,
				logger.String("company_name", createdOrExisting.Name),
				logger.String("company_id", createdOrExisting.ID.String()))
		} else {
			// Company already existed, was skipped
			result.SkippedItems++
			uc.logger.LogEntitySkipped(ctx, "company", ticker, "already_exists")
		}

		// Update company reference to use the returned one (created or existing)
		companies[ticker] = createdOrExisting

		// Cache if enabled (cache operations outside transaction for better performance)
		if uc.cacheService != nil {
			uc.cacheService.SetCompany(ctx, ticker, createdOrExisting, 5*time.Minute)
		}
	}
	// Process brokerages using transaction with duplicate handling
	for name, brokerage := range brokerages {
		// Use CreateIgnoreDuplicatesWithTx to avoid transaction aborts on duplicates
		createdOrExisting, err := uc.brokerageRepo.CreateIgnoreDuplicatesWithTx(ctx, tx, brokerage)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create brokerage %s: %v", name, err))
			uc.logger.LogEntityError(ctx, "brokerage", name, err)
			continue
		}

		// Check if it was created or already existed
		if createdOrExisting.ID == brokerage.ID {
			// New brokerage was created
			result.Brokerages++
			result.ProcessedItems++
			uc.logger.LogEntityCreated(ctx, "brokerage", name,
				logger.String("brokerage_id", createdOrExisting.ID.String()))
		} else {
			// Brokerage already existed, was skipped
			result.SkippedItems++
			uc.logger.LogEntitySkipped(ctx, "brokerage", name, "already_exists")
		}

		// Update brokerage reference to use the returned one (created or existing)
		brokerages[name] = createdOrExisting

		// Cache if enabled
		if uc.cacheService != nil {
			uc.cacheService.SetBrokerage(ctx, name, createdOrExisting, 5*time.Minute)
		}
	}

	return nil
}

// processStockRatingsTransactional procesa los stock ratings usando transacciones
func (uc *PopulateDatabaseUseCase) processStockRatingsTransactional(ctx context.Context, tx *gorm.DB, items []StockDataItem, result *PopulationResult) error {
	uc.logger.Debug(ctx, "Processing stock ratings in transaction",
		logger.String("operation", "process_stock_ratings_tx"),
		logger.Int("items_count", len(items)))

	// Collect all stock ratings to insert in bulk
	var stockRatings []*entities.StockRating

	for _, item := range items {
		// Get company (should exist from previous step within same transaction)
		company, err := uc.companyRepo.GetByTickerWithTx(ctx, tx, item.Ticker)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Company not found for ticker %s: %v", item.Ticker, err))
			uc.logger.LogEntityError(ctx, "stock_rating", fmt.Sprintf("%s-%s", item.Ticker, item.Brokerage), err,
				logger.String("ticker", item.Ticker),
				logger.String("reason", "company_not_found"))
			continue
		}

		// Get brokerage (should exist from previous step within same transaction)
		brokerage, err := uc.brokerageRepo.GetByNameWithTx(ctx, tx, item.Brokerage)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Brokerage not found %s: %v", item.Brokerage, err))
			uc.logger.LogEntityError(ctx, "stock_rating", fmt.Sprintf("%s-%s", item.Ticker, item.Brokerage), err,
				logger.String("brokerage", item.Brokerage),
				logger.String("reason", "brokerage_not_found"))
			continue
		}

		// Create stock rating entity
		stockRating := entities.NewStockRating(
			company.ID,
			brokerage.ID,
			item.Action,
			item.EventTime,
		)

		// Set additional fields
		stockRating.RatingFrom = item.RatingFrom
		stockRating.RatingTo = item.RatingTo
		stockRating.TargetFrom = item.TargetFrom
		stockRating.TargetTo = item.TargetTo

		// Add to bulk insert collection
		stockRatings = append(stockRatings, stockRating)
	}

	// Perform bulk insert ignoring duplicates
	if len(stockRatings) > 0 {
		insertedCount, err := uc.stockRatingRepo.BulkInsertIgnoreDuplicatesWithTx(ctx, tx, stockRatings)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to bulk insert stock ratings: %v", err))
			uc.logger.Error(ctx, "❌ Failed to bulk insert stock ratings", err,
				logger.String("operation", "bulk_insert_stock_ratings"))
			return err
		}

		// Update metrics
		result.StockRatings += insertedCount
		result.ProcessedItems += insertedCount
		skippedCount := len(stockRatings) - insertedCount
		result.SkippedItems += skippedCount

		// Log results
		uc.logger.Info(ctx, "✅ Bulk insert stock ratings completed",
			logger.String("operation", "bulk_insert_stock_ratings"),
			logger.Int("total_ratings", len(stockRatings)),
			logger.Int("inserted", insertedCount),
			logger.Int("skipped_duplicates", skippedCount))

		// Cache inserted ratings if enabled
		if uc.cacheService != nil {
			for _, stockRating := range stockRatings {
				uc.cacheService.SetStockRating(ctx, stockRating, 5*time.Minute)
			}
		}
	}

	return nil
}

// ========================================
// ENHANCED INTEGRITY VALIDATION METHODS
// ========================================

// validateIntegrityEnhanced utiliza el nuevo servicio de validación para verificar integridad
func (uc *PopulateDatabaseUseCase) validateIntegrityEnhanced(ctx context.Context, result *PopulationResult) error {
	uc.logger.Info(ctx, "🔍 Running enhanced database integrity validation...",
		logger.String("operation", "enhanced_integrity_validation"))

	// Usar el nuevo servicio de validación de integridad
	integrityReport, err := uc.integrityService.ValidateFullIntegrity(ctx)
	if err != nil {
		return fmt.Errorf("integrity validation failed: %w", err)
	}

	// Log usando el nuevo logger de integridad
	uc.logger.LogIntegrityValidation(ctx, string(integrityReport.OverallStatus),
		integrityReport.TotalIssues, integrityReport.Duration)

	// Si hay problemas críticos, intentar reparación automática
	if integrityReport.OverallStatus == services.IntegrityStatusCritical {
		uc.logger.Info(ctx, "🔧 Critical issues found, attempting automatic repair...",
			logger.String("operation", "auto_repair"),
			logger.Int("critical_issues", integrityReport.CriticalIssues))

		repairReport, err := uc.integrityService.RepairMinorIssues(ctx, false) // false = not dry run
		if err != nil {
			uc.logger.Error(ctx, "❌ Automatic repair failed", err,
				logger.String("operation", "auto_repair"))
		} else {
			uc.logger.Info(ctx, "🔧 Automatic repair completed",
				logger.String("operation", "auto_repair"),
				logger.Int("total_repairs", repairReport.TotalRepairs),
				logger.Int("orphans_removed", repairReport.RepairedOrphans),
				logger.Int("duplicates_removed", repairReport.RemovedDuplicates))

			// Re-validate after repair
			if repairReport.TotalRepairs > 0 {
				uc.logger.Info(ctx, "🔍 Re-validating after automatic repair...",
					logger.String("operation", "post_repair_validation"))
				postRepairReport, err := uc.integrityService.ValidateFullIntegrity(ctx)
				if err == nil {
					uc.logger.LogIntegrityValidation(ctx, string(postRepairReport.OverallStatus),
						postRepairReport.TotalIssues, postRepairReport.Duration)
				}
			}
		}
	}

	// Return error only for critical unresolved issues
	if integrityReport.OverallStatus == services.IntegrityStatusCritical && integrityReport.CriticalIssues > 0 {
		return fmt.Errorf("critical integrity issues remain: %d issues found", integrityReport.CriticalIssues)
	}

	return nil
}

// logIntegrityResults registra los resultados de validación de integridad
// ========================================
// LEGACY VALIDATION METHODS (for backward compatibility)
// ========================================
