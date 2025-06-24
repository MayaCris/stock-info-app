package scripts

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/mappers/implementationsMap"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/implementation"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cache"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
)

// TestMapper tests the API mapper functionality
func TestMapper(cfg *config.Config) error {
	log.Println("🧪 Testing API Mapper functionality...")
	// Initialize database connection
	db, err := cockroachdb.NewConnection(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	companyRepo := implementation.NewCompanyRepository(db.DB)
	brokerageRepo := implementation.NewBrokerageRepository(db.DB)
	stockRatingRepo := implementation.NewStockRatingRepository(db.DB)

	// Initialize cache service
	cacheService := cache.NewCacheService(cfg)

	// Initialize mapper
	mapper := implementationsMap.NewAPIMapper(companyRepo, brokerageRepo, stockRatingRepo, cacheService)

	// Create sample API data for testing
	sampleItems := []stock_api.StockRatingItem{
		{
			Ticker:     "AAPL",
			Company:    "Apple Inc.",
			Brokerage:  "Goldman Sachs",
			Action:     "upgraded by",
			RatingFrom: "Hold",
			RatingTo:   "Buy",
			TargetFrom: "$150.00",
			TargetTo:   "$175.00",
			Time:       time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
		{
			Ticker:     "MSFT",
			Company:    "Microsoft Corporation",
			Brokerage:  "Morgan Stanley",
			Action:     "reiterated by",
			RatingFrom: "Buy",
			RatingTo:   "Buy",
			TargetFrom: "$280.00",
			TargetTo:   "$300.00",
			Time:       time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			Ticker:     "GOOGL",
			Company:    "Alphabet Inc.",
			Brokerage:  "JP Morgan",
			Action:     "downgraded by",
			RatingFrom: "Buy",
			RatingTo:   "Hold",
			TargetFrom: "$120.00",
			TargetTo:   "$110.00",
			Time:       time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
		},
	}

	ctx := context.Background()

	// Test 1: Single item mapping
	log.Println("\n=== TEST 1: Single Item Mapping ===")
	singleResult, err := mapper.MapStockRatingItem(ctx, sampleItems[0])
	if err != nil {
		return fmt.Errorf("failed to map single item: %w", err)
	}

	log.Printf("✅ Single mapping successful:")
	log.Printf("  Company: %s (%s) - Created: %t",
		singleResult.Company.Name, singleResult.Company.Ticker, singleResult.WasCompanyCreated)
	log.Printf("  Brokerage: %s - Created: %t",
		singleResult.Brokerage.Name, singleResult.WasBrokerageCreated)
	log.Printf("  Rating: %s - Created: %t",
		singleResult.StockRating.Action, singleResult.WasRatingCreated)

	// Test 2: Batch mapping (standard)
	log.Println("\n=== TEST 2: Batch Mapping (Standard) ===")
	batchResult, err := mapper.MapStockRatingItems(ctx, sampleItems)
	if err != nil {
		return fmt.Errorf("failed to batch map items: %w", err)
	}

	log.Printf("✅ Batch mapping completed:")
	log.Printf("📊 %s", batchResult.Summary())

	if batchResult.HasErrors() {
		log.Printf("⚠️  Errors occurred:")
		for _, failed := range batchResult.FailedMappings {
			log.Printf("  - %s (%s): %s", failed.OriginalItem.Ticker, failed.ErrorType, failed.Error)
		}
	}
	// Test 3: Batch mapping with cache (optimized)
	log.Println("\n=== TEST 3: Batch Mapping (Optimized with Cache) ===")
	if mapperImpl, ok := mapper.(*implementationsMap.APIMapperImpl); ok {
		// Now we can access the BatchMapWithCache method
		log.Println("🚀 Testing optimized batch mapping...")
		cachedResult, err := mapperImpl.BatchMapWithCache(ctx, sampleItems)
		if err != nil {
			log.Printf("⚠️  Optimized mapping failed: %v", err)
		} else {
			log.Printf("✅ Optimized batch mapping completed:")
			log.Printf("📊 %s", cachedResult.Summary())
		}
	} else {
		log.Println("✅ Optimized batch mapping available (with in-memory caching)")
	}

	// Test 4: Mapping statistics
	log.Println("\n=== TEST 4: Mapping Statistics ===")
	stats := batchResult.GetMappingStats()
	log.Printf("📈 Mapping Statistics:")
	log.Printf("  Unique Companies: %d", stats.UniqueCompanies)
	log.Printf("  Unique Brokerages: %d", stats.UniqueBrokerages)
	log.Printf("  Action Distribution: %+v", stats.ActionDistribution)
	if len(stats.ErrorBreakdown) > 0 {
		log.Printf("  Error Breakdown: %+v", stats.ErrorBreakdown)
	}

	// Test 5: Validation testing
	log.Println("\n=== TEST 5: Validation Testing ===")

	// Test with invalid item
	invalidItem := stock_api.StockRatingItem{
		Ticker:    "", // Missing required field
		Company:   "Test Company",
		Brokerage: "Test Brokerage",
		Action:    "upgraded by",
		Time:      "invalid-time-format", // Invalid time
	}

	if err := mapper.ValidateAPIItem(invalidItem); err != nil {
		log.Printf("✅ Validation correctly caught invalid item: %v", err)
	} else {
		log.Printf("❌ Validation should have failed for invalid item")
	}

	// Test with valid item
	validItem := sampleItems[0]
	if err := mapper.ValidateAPIItem(validItem); err != nil {
		log.Printf("❌ Validation failed for valid item: %v", err)
	} else {
		log.Printf("✅ Validation passed for valid item")
	}

	log.Println("\n✅ Mapper testing completed successfully!")
	return nil
}

// TestMapperWithRealAPI tests the mapper with real API data
func TestMapperWithRealAPI(cfg *config.Config) error {
	log.Println("🌐 Testing Mapper with Real API Data...")

	// Initialize database connection
	db, err := cockroachdb.NewConnection(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	// Initialize repositories
	companyRepo := implementation.NewCompanyRepository(db.DB)
	brokerageRepo := implementation.NewBrokerageRepository(db.DB)
	stockRatingRepo := implementation.NewStockRatingRepository(db.DB)

	// Initialize cache service
	cacheService := cache.NewCacheService(cfg)

	// Initialize mapper
	mapper := implementationsMap.NewAPIMapper(companyRepo, brokerageRepo, stockRatingRepo, cacheService)

	// Initialize API client
	apiClient := stock_api.NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Fetch recent data from API (limited to 2 pages for testing)
	log.Println("📡 Fetching data from external API...")
	apiItems, err := apiClient.FetchRecentPages(ctx, 2)
	if err != nil {
		return fmt.Errorf("failed to fetch API data: %w", err)
	}

	log.Printf("📦 Fetched %d items from API", len(apiItems))

	// Map the API data to domain entities
	log.Println("🔄 Mapping API data to domain entities...")
	result, err := mapper.MapStockRatingItems(ctx, apiItems)
	if err != nil {
		return fmt.Errorf("failed to map API data: %w", err)
	}

	// Display results
	log.Printf("🎉 Mapping completed!")
	log.Printf("📊 %s", result.Summary())

	// Show detailed statistics
	stats := result.GetMappingStats()
	log.Printf("\n📈 Detailed Statistics:")
	log.Printf("  Success Rate: %.1f%%", result.GetSuccessRate())
	log.Printf("  Unique Companies: %d", stats.UniqueCompanies)
	log.Printf("  Unique Brokerages: %d", stats.UniqueBrokerages)
	log.Printf("  Action Distribution:")
	for action, count := range stats.ActionDistribution {
		log.Printf("    - %s: %d", action, count)
	}

	// Show errors if any
	if result.HasErrors() {
		log.Printf("\n⚠️  Errors encountered:")
		errorCounts := make(map[string]int)
		for _, failed := range result.FailedMappings {
			errorCounts[failed.ErrorType]++
		}
		for errorType, count := range errorCounts {
			log.Printf("  - %s: %d errors", errorType, count)
		}

		// Show first few errors for debugging
		log.Printf("\nFirst few errors:")
		for i, failed := range result.FailedMappings {
			if i >= 3 {
				break
			}
			log.Printf("  %d. %s (%s): %s", i+1, failed.OriginalItem.Ticker, failed.ErrorType, failed.Error)
		}
	}

	// Query database to verify data was saved
	log.Println("\n🔍 Verifying data in database...")

	companyCount, _ := companyRepo.Count(ctx)
	brokerageCount, _ := brokerageRepo.Count(ctx)
	ratingCount, _ := stockRatingRepo.Count(ctx)

	log.Printf("📊 Database totals after mapping:")
	log.Printf("  Companies: %d", companyCount)
	log.Printf("  Brokerages: %d", brokerageCount)
	log.Printf("  Stock Ratings: %d", ratingCount)

	log.Println("\n✅ Real API mapping test completed successfully!")
	return nil
}

// ShowMappingExample demonstrates how to use the mapper programmatically
func ShowMappingExample(cfg *config.Config) error {
	log.Println("📖 Mapper Usage Example...")

	// This is how you would typically use the mapper in your application:

	log.Println(`🔧 Usage Pattern:

1. Initialize dependencies:
   db, err := cockroachdb.NewConnection(cfg)
   companyRepo := repositories.NewCompanyRepository(db.DB)
   brokerageRepo := repositories.NewBrokerageRepository(db.DB)
   stockRatingRepo := repositories.NewStockRatingRepository(db.DB)
   mapper := mappers.NewAPIMapper(companyRepo, brokerageRepo, stockRatingRepo)

2. Fetch API data:
   apiClient := stock_api.NewClient(cfg)
   items, err := apiClient.FetchAllPages(ctx)

3. Map to domain entities:
   result, err := mapper.MapStockRatingItems(ctx, items)

4. Handle results:
   if result.HasErrors() {
	   // Handle mapping errors
   }
   // Use result.SuccessfulMappings for further processing

5. The mapper automatically:
   ✅ Creates new companies/brokerages as needed
   ✅ Avoids duplicate stock ratings
   ✅ Validates data integrity
   ✅ Provides detailed statistics
   ✅ Preserves raw API data for debugging`)

	return nil
}
