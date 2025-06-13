package implementationsMap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/mappers/interfacesMap"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
)

// APIMapperImpl implements the APIMapper interface
type APIMapperImpl struct {
	companyRepo     interfaces.CompanyRepository
	brokerageRepo   interfaces.BrokerageRepository
	stockRatingRepo interfaces.StockRatingRepository
	cacheService    services.CacheService
}

// NewAPIMapper creates a new API mapper implementation
func NewAPIMapper(
	companyRepo interfaces.CompanyRepository,
	brokerageRepo interfaces.BrokerageRepository,
	stockRatingRepo interfaces.StockRatingRepository,
	cacheService services.CacheService,
) interfacesMap.APIMapper {
	return &APIMapperImpl{
		companyRepo:     companyRepo,
		brokerageRepo:   brokerageRepo,
		stockRatingRepo: stockRatingRepo,
		cacheService:    cacheService,
	}
}

// ========================================
// SINGLE ITEM MAPPING
// ========================================

// MapStockRatingItem maps a single API item to domain entities
func (m *APIMapperImpl) MapStockRatingItem(ctx context.Context, item stock_api.StockRatingItem) (*interfacesMap.MappingResult, error) {
	// Validate the API item first
	if err := m.ValidateAPIItem(item); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	result := &interfacesMap.MappingResult{
		OriginalItem: item,
	}

	// Generate raw JSON for debugging
	if rawJSON, err := json.Marshal(item); err == nil {
		result.RawJSON = rawJSON
	}

	// Step 1: Find or create Company
	company, wasCompanyCreated, err := m.findOrCreateCompany(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to process company: %w", err)
	}
	result.Company = company
	result.WasCompanyCreated = wasCompanyCreated

	// Step 2: Find or create Brokerage
	brokerage, wasBrokerageCreated, err := m.findOrCreateBrokerage(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to process brokerage: %w", err)
	}
	result.Brokerage = brokerage
	result.WasBrokerageCreated = wasBrokerageCreated

	// Step 3: Parse event time
	eventTime, err := m.parseEventTime(item.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event time: %w", err)
	}

	// Step 4: Find or create StockRating
	stockRating, wasRatingCreated, err := m.findOrCreateStockRating(ctx, company, brokerage, item, eventTime, result.RawJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to process stock rating: %w", err)
	}
	result.StockRating = stockRating
	result.WasRatingCreated = wasRatingCreated

	return result, nil
}

// ========================================
// BATCH MAPPING
// ========================================

// MapStockRatingItems maps multiple API items to domain entities
func (m *APIMapperImpl) MapStockRatingItems(ctx context.Context, items []stock_api.StockRatingItem) (*interfacesMap.BatchMappingResult, error) {
	startTime := time.Now()

	result := &interfacesMap.BatchMappingResult{
		TotalItems: len(items),
	}

	log.Printf("🔄 Starting batch mapping of %d items...", len(items))

	// Process each item
	for i, item := range items {
		if i > 0 && i%100 == 0 {
			log.Printf("📈 Processed %d/%d items (%.1f%%)", i, len(items), float64(i)/float64(len(items))*100)
		}

		mappingResult, err := m.MapStockRatingItem(ctx, item)
		if err != nil {
			// Handle failed mapping
			failedMapping := interfacesMap.FailedMapping{
				OriginalItem: item,
				Error:        err.Error(),
				ErrorType:    m.categorizeError(err),
			}
			result.FailedMappings = append(result.FailedMappings, failedMapping)
			result.FailureCount++

			log.Printf("⚠️  Failed to map item %d: %v", i+1, err)
			continue
		}

		// Handle successful mapping
		result.SuccessfulMappings = append(result.SuccessfulMappings, *mappingResult)
		result.SuccessCount++

		// Update statistics
		if mappingResult.WasCompanyCreated {
			result.CompaniesCreated++
		}
		if mappingResult.WasBrokerageCreated {
			result.BrokeragesCreated++
		}
		if mappingResult.WasRatingCreated {
			result.RatingsCreated++
		} else {
			result.DuplicateRatings++
		}
	}

	processingTime := time.Since(startTime)
	log.Printf("✅ Batch mapping completed in %v", processingTime)
	log.Printf("📊 %s", result.Summary())

	return result, nil
}

// ========================================
// VALIDATION
// ========================================

// ValidateAPIItem validates an API item before mapping
func (m *APIMapperImpl) ValidateAPIItem(item stock_api.StockRatingItem) error {
	var errors []interfacesMap.ValidationError

	// Required fields validation
	if strings.TrimSpace(item.Ticker) == "" {
		errors = append(errors, interfacesMap.ValidationError{
			Field: "ticker", Value: item.Ticker, Message: "ticker is required",
		})
	}

	if strings.TrimSpace(item.Company) == "" {
		errors = append(errors, interfacesMap.ValidationError{
			Field: "company", Value: item.Company, Message: "company name is required",
		})
	}

	if strings.TrimSpace(item.Brokerage) == "" {
		errors = append(errors, interfacesMap.ValidationError{
			Field: "brokerage", Value: item.Brokerage, Message: "brokerage name is required",
		})
	}

	if strings.TrimSpace(item.Action) == "" {
		errors = append(errors, interfacesMap.ValidationError{
			Field: "action", Value: item.Action, Message: "action is required",
		})
	}

	if strings.TrimSpace(item.Time) == "" {
		errors = append(errors, interfacesMap.ValidationError{
			Field: "time", Value: item.Time, Message: "time is required",
		})
	}

	// Time format validation
	if item.Time != "" {
		if _, err := time.Parse(time.RFC3339, item.Time); err != nil {
			errors = append(errors, interfacesMap.ValidationError{
				Field: "time", Value: item.Time, Message: "invalid time format",
			})
		}
	}

	// Ticker format validation (basic)
	if item.Ticker != "" {
		ticker := strings.ToUpper(strings.TrimSpace(item.Ticker))
		if len(ticker) < 1 || len(ticker) > 10 {
			errors = append(errors, interfacesMap.ValidationError{
				Field: "ticker", Value: item.Ticker, Message: "ticker must be 1-10 characters",
			})
		}
	}

	// Business logic validation
	if item.RatingFrom != "" && item.RatingTo != "" && item.RatingFrom == item.RatingTo {
		// This might be valid for reiterations, so just warn
		log.Printf("⚠️  Rating from/to are the same for %s: %s", item.Ticker, item.RatingFrom)
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %+v", errors)
	}

	return nil
}

// ========================================
// HELPER METHODS - ENTITY CREATION
// ========================================

// findOrCreateCompany finds or creates a company from API data with Redis cache
func (m *APIMapperImpl) findOrCreateCompany(ctx context.Context, item stock_api.StockRatingItem) (*entities.Company, bool, error) {
	ticker := strings.ToUpper(strings.TrimSpace(item.Ticker))
	name := strings.TrimSpace(item.Company)

	// First, check Redis cache
	if cachedCompany, err := m.cacheService.GetCompany(ctx, ticker); err == nil && cachedCompany != nil {
		// Cache hit - update name if necessary
		if cachedCompany.Name != name && len(name) > len(cachedCompany.Name) {
			cachedCompany.Name = name
			if updateErr := m.companyRepo.Update(ctx, cachedCompany); updateErr != nil {
				log.Printf("⚠️  Failed to update company name for %s: %v", ticker, updateErr)
			} else {
				// Update cache with new data
				m.cacheService.SetCompany(ctx, ticker, cachedCompany, 0)
			}
		}
		return cachedCompany, false, nil
	}

	// Cache miss or error - check database
	existingCompany, err := m.companyRepo.GetByTicker(ctx, ticker)
	if err == nil {
		// Company exists, optionally update name if it's different/better
		if existingCompany.Name != name && len(name) > len(existingCompany.Name) {
			existingCompany.Name = name
			if updateErr := m.companyRepo.Update(ctx, existingCompany); updateErr != nil {
				log.Printf("⚠️  Failed to update company name for %s: %v", ticker, updateErr)
			}
		}
		// Cache the company for future use
		if cacheErr := m.cacheService.SetCompany(ctx, ticker, existingCompany, 0); cacheErr != nil {
			log.Printf("⚠️  Failed to cache company %s: %v", ticker, cacheErr)
		}

		return existingCompany, false, nil
	}

	// Check if error is "not found" vs actual error
	if !strings.Contains(err.Error(), "not found") {
		return nil, false, fmt.Errorf("failed to check existing company: %w", err)
	}

	// Company doesn't exist, create new one
	newCompany := entities.NewCompany(ticker, name)
	if err := m.companyRepo.Create(ctx, newCompany); err != nil {
		return nil, false, fmt.Errorf("failed to create company: %w", err)
	}

	// Cache the new company
	if cacheErr := m.cacheService.SetCompany(ctx, ticker, newCompany, 0); cacheErr != nil {
		log.Printf("⚠️  Failed to cache new company %s: %v", ticker, cacheErr)
	}

	log.Printf("✨ Created new company: %s (%s)", ticker, name)
	return newCompany, true, nil
}

// findOrCreateBrokerage finds or creates a brokerage from API data
func (m *APIMapperImpl) findOrCreateBrokerage(ctx context.Context, item stock_api.StockRatingItem) (*entities.Brokerage, bool, error) {
	name := strings.TrimSpace(item.Brokerage)

	// First, check Redis cache
	if cachedBrokerage, err := m.cacheService.GetBrokerage(ctx, name); err == nil && cachedBrokerage != nil {
		// Cache hit - return existing brokerage
		if cachedBrokerage.Name != name && len(name) > len(cachedBrokerage.Name) {
			cachedBrokerage.Name = name
			if updateErr := m.brokerageRepo.Update(ctx, cachedBrokerage); updateErr != nil {
				log.Printf("⚠️  Failed to update brokerage name for %s: %v", name, updateErr)
			} else {
				// Update cache with new data
				m.cacheService.SetBrokerage(ctx, name, cachedBrokerage, 0)
			}
		}
		return cachedBrokerage, false, nil
	}

	// Try to find existing brokerage by name
	existingBrokerage, err := m.brokerageRepo.GetByName(ctx, name)
	if err == nil {
		if cacheErr := m.cacheService.SetBrokerage(ctx, name, existingBrokerage, 0); cacheErr != nil {
			log.Printf("⚠️  Failed to cache brokerage %s: %v", name, cacheErr)
		}

		return existingBrokerage, false, nil
	}

	// Check if error is "not found" vs actual error
	if !strings.Contains(err.Error(), "not found") {
		return nil, false, fmt.Errorf("failed to check existing brokerage: %w", err)
	}

	// Brokerage doesn't exist, create new one
	newBrokerage := entities.NewBrokerage(name)
	if err := m.brokerageRepo.Create(ctx, newBrokerage); err != nil {
		return nil, false, fmt.Errorf("failed to create brokerage: %w", err)
	}

	// Cache the new brokerage
	if cacheErr := m.cacheService.SetBrokerage(ctx, name, newBrokerage, 0); cacheErr != nil {
		log.Printf("⚠️  Failed to cache new brokerage %s: %v", name, cacheErr)
	}

	log.Printf("✨ Created new brokerage: %s", name)
	return newBrokerage, true, nil
}

// findOrCreateStockRating finds or creates a stock rating from API data
func (m *APIMapperImpl) findOrCreateStockRating(ctx context.Context, company *entities.Company, brokerage *entities.Brokerage,
	item stock_api.StockRatingItem, eventTime time.Time, rawJSON json.RawMessage) (*entities.StockRating, bool, error) {

	//First, validate that company and brokerage are not nil
	if company == nil {
		return nil, false, fmt.Errorf("company is nil")
	}
	if brokerage == nil {
		return nil, false, fmt.Errorf("brokerage is nil")
	}

	//Check Redis cache for existing stock rating
	if cachedRating, err := m.cacheService.GetStockRating(ctx, company.ID, brokerage.ID); err == nil && cachedRating != nil {

		// Cache hit - return existing rating
		log.Printf("🔍 Found existing stock rating in cache for %s by %s at %s", company.Ticker, brokerage.Name, eventTime)
		return cachedRating, false, nil
	}

	// Check if rating already exists (to avoid duplicates)
	existingRating, err := m.stockRatingRepo.FindExisting(ctx, company.ID, brokerage.ID, eventTime)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check existing rating: %w", err)
	}

	if existingRating != nil {
		// Rating already exists, return it
		return existingRating, false, nil
	}

	// Create new rating
	newRating := entities.NewStockRating(company.ID, brokerage.ID, item.Action, eventTime)

	// Set optional fields
	newRating.RatingFrom = strings.TrimSpace(item.RatingFrom)
	newRating.RatingTo = strings.TrimSpace(item.RatingTo)
	newRating.TargetFrom = strings.TrimSpace(item.TargetFrom)
	newRating.TargetTo = strings.TrimSpace(item.TargetTo)
	newRating.Source = "api"
	newRating.RawData = rawJSON

	// Create the rating
	if err := m.stockRatingRepo.Create(ctx, newRating); err != nil {
		return nil, false, fmt.Errorf("failed to create stock rating: %w", err)
	}

	 // Cache the new rating
	if cacheErr := m.cacheService.SetStockRating(ctx, newRating, 0); cacheErr != nil {
		log.Printf("⚠️  Failed to cache new stock rating for %s by %s at %s: %v", company.Ticker, brokerage.Name, eventTime, cacheErr)
	}
	return newRating, true, nil
}

// ========================================
// HELPER METHODS - DATA TRANSFORMATION
// ========================================

// parseEventTime parses the event time from API format
func (m *APIMapperImpl) parseEventTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// Parse RFC3339 format (API format)
	eventTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time '%s': %w", timeStr, err)
	}

	// Validate time is not in the future (with 1 hour tolerance)
	now := time.Now()
	if eventTime.After(now.Add(1 * time.Hour)) {
		return time.Time{}, fmt.Errorf("event time is in the future: %s", eventTime)
	}

	// Validate time is not too old (more than 10 years ago)
	tenYearsAgo := now.AddDate(-10, 0, 0)
	if eventTime.Before(tenYearsAgo) {
		return time.Time{}, fmt.Errorf("event time is too old: %s", eventTime)
	}

	return eventTime, nil
}

// categorizeError categorizes mapping errors for statistics
func (m *APIMapperImpl) categorizeError(err error) string {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "validation"):
		return "validation"
	case strings.Contains(errStr, "company"):
		return "company"
	case strings.Contains(errStr, "brokerage"):
		return "brokerage"
	case strings.Contains(errStr, "rating"):
		return "rating"
	case strings.Contains(errStr, "time"):
		return "time_parsing"
	default:
		return "unknown"
	}
}

// ========================================
// BATCH OPTIMIZATION METHODS
// ========================================

// BatchMapWithCache maps items using in-memory caching for performance
func (m *APIMapperImpl) BatchMapWithCache(ctx context.Context, items []stock_api.StockRatingItem) (*interfacesMap.BatchMappingResult, error) {
	// Create caches for companies and brokerages to avoid repeated DB calls
	companyCache := make(map[string]*entities.Company)
	brokerageCache := make(map[string]*entities.Brokerage)

	startTime := time.Now()
	result := &interfacesMap.BatchMappingResult{
		TotalItems: len(items),
	}

	log.Printf("🚀 Starting optimized batch mapping with cache for %d items...", len(items))

	for i, item := range items {
		if i > 0 && i%100 == 0 {
			log.Printf("📈 Processed %d/%d items (%.1f%%) - Cache: %d companies, %d brokerages",
				i, len(items), float64(i)/float64(len(items))*100, len(companyCache), len(brokerageCache))
		}

		// Use cached versions if available
		mappingResult, err := m.mapItemWithCache(ctx, item, companyCache, brokerageCache)
		if err != nil {
			failedMapping := interfacesMap.FailedMapping{
				OriginalItem: item,
				Error:        err.Error(),
				ErrorType:    m.categorizeError(err),
			}
			result.FailedMappings = append(result.FailedMappings, failedMapping)
			result.FailureCount++
			continue
		}

		result.SuccessfulMappings = append(result.SuccessfulMappings, *mappingResult)
		result.SuccessCount++

		// Update statistics
		if mappingResult.WasCompanyCreated {
			result.CompaniesCreated++
		}
		if mappingResult.WasBrokerageCreated {
			result.BrokeragesCreated++
		}
		if mappingResult.WasRatingCreated {
			result.RatingsCreated++
		} else {
			result.DuplicateRatings++
		}
	}

	processingTime := time.Since(startTime)
	log.Printf("✅ Optimized batch mapping completed in %v", processingTime)
	log.Printf("📊 %s", result.Summary())
	log.Printf("🗄️  Cache efficiency: %d companies, %d brokerages cached", len(companyCache), len(brokerageCache))

	return result, nil
}

// mapItemWithCache maps a single item using caches
func (m *APIMapperImpl) mapItemWithCache(ctx context.Context, item stock_api.StockRatingItem,
	companyCache map[string]*entities.Company, brokerageCache map[string]*entities.Brokerage) (*interfacesMap.MappingResult, error) {

	// Validate first
	if err := m.ValidateAPIItem(item); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	result := &interfacesMap.MappingResult{
		OriginalItem: item,
	}

	// Generate raw JSON
	if rawJSON, err := json.Marshal(item); err == nil {
		result.RawJSON = rawJSON
	}

	// Get or create company (with cache)
	ticker := strings.ToUpper(strings.TrimSpace(item.Ticker))
	company, ok := companyCache[ticker]
	if !ok {
		var wasCreated bool
		var err error
		company, wasCreated, err = m.findOrCreateCompany(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to process company: %w", err)
		}
		companyCache[ticker] = company
		result.WasCompanyCreated = wasCreated
	}
	result.Company = company

	// Get or create brokerage (with cache)
	brokerageName := strings.TrimSpace(item.Brokerage)
	brokerage, ok := brokerageCache[brokerageName]
	if !ok {
		var wasCreated bool
		var err error
		brokerage, wasCreated, err = m.findOrCreateBrokerage(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("failed to process brokerage: %w", err)
		}
		brokerageCache[brokerageName] = brokerage
		result.WasBrokerageCreated = wasCreated
	}
	result.Brokerage = brokerage

	// Parse time and create rating
	eventTime, err := m.parseEventTime(item.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event time: %w", err)
	}

	stockRating, wasRatingCreated, err := m.findOrCreateStockRating(ctx, company, brokerage, item, eventTime, result.RawJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to process stock rating: %w", err)
	}
	result.StockRating = stockRating
	result.WasRatingCreated = wasRatingCreated

	return result, nil
}
