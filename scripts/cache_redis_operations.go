package scripts

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cache"
	"github.com/google/uuid"
)

// TestCacheOperations tests all cache operations including Redis connectivity
func TestCacheOperations(cfg *config.Config) error {
	log.Println("🔄 Testing Cache Operations...")

	// Initialize cache service (will use Redis with memory fallback)
	cacheService := cache.NewCacheService(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Basic connectivity and health check
	log.Println("\n=== TEST 1: Cache Connectivity ===")
	if err := testCacheConnectivity(ctx, cacheService); err != nil {
		return fmt.Errorf("cache connectivity test failed: %w", err)
	}

	// Test 2: Company operations
	log.Println("\n=== TEST 2: Company Cache Operations ===")
	if err := testCompanyOperations(ctx, cacheService); err != nil {
		return fmt.Errorf("company cache operations failed: %w", err)
	}

	// Test 3: Brokerage operations
	log.Println("\n=== TEST 3: Brokerage Cache Operations ===")
	if err := testBrokerageOperations(ctx, cacheService); err != nil {
		return fmt.Errorf("brokerage cache operations failed: %w", err)
	}

	// Test 4: Stock Rating operations
	log.Println("\n=== TEST 4: Stock Rating Cache Operations ===")
	if err := testStockRatingOperations(ctx, cacheService); err != nil {
		return fmt.Errorf("stock rating cache operations failed: %w", err)
	}

	// Test 5: Bulk operations
	log.Println("\n=== TEST 5: Bulk Cache Operations ===")
	if err := testBulkOperations(ctx, cacheService); err != nil {
		return fmt.Errorf("bulk cache operations failed: %w", err)
	}

	// Test 6: Cache management
	log.Println("\n=== TEST 6: Cache Management Operations ===")
	if err := testCacheManagement(ctx, cacheService); err != nil {
		return fmt.Errorf("cache management operations failed: %w", err)
	}

	// Test 7: Performance and statistics
	log.Println("\n=== TEST 7: Cache Performance & Statistics ===")
	if err := testCacheStatistics(ctx, cacheService); err != nil {
		return fmt.Errorf("cache statistics test failed: %w", err)
	}

	log.Println("\n✅ All cache operations tests completed successfully!")
	return nil
}

// testCacheConnectivity tests basic connectivity and health
func testCacheConnectivity(ctx context.Context, cacheService services.CacheService) error {
	// Test ping
	if err := cacheService.Ping(ctx); err != nil {
		log.Printf("⚠️  Cache ping failed (might be using fallback): %v", err)
	} else {
		log.Println("✅ Cache ping successful")
	}

	// Get initial statistics
	stats, err := cacheService.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cache stats: %w", err)
	}

	log.Printf("📊 Cache Backend: %s", stats.Backend)
	log.Printf("📊 Cache Connected: %t", stats.IsConnected)
	log.Printf("📊 Initial Key Count: %d", stats.KeyCount)

	return nil
}

// testCompanyOperations tests company cache operations
func testCompanyOperations(ctx context.Context, cacheService services.CacheService) error {
	// Create test company
	company := entities.NewCompany("AAPL", "Apple Inc.")
	ticker := "AAPL"
	ttl := 5 * time.Minute

	// Test Set
	log.Printf("🔄 Setting company: %s", ticker)
	if err := cacheService.SetCompany(ctx, ticker, company, ttl); err != nil {
		return fmt.Errorf("failed to set company: %w", err)
	}
	log.Printf("✅ Company set successfully")

	// Test Get
	log.Printf("🔄 Getting company: %s", ticker)
	retrievedCompany, err := cacheService.GetCompany(ctx, ticker)
	if err != nil {
		return fmt.Errorf("failed to get company: %w", err)
	}

	// Validate data
	if retrievedCompany.Ticker != company.Ticker || retrievedCompany.Name != company.Name {
		return fmt.Errorf("retrieved company data mismatch: got %+v, want %+v", retrievedCompany, company)
	}
	log.Printf("✅ Company retrieved successfully: %s (%s)", retrievedCompany.Name, retrievedCompany.Ticker)

	// Test key existence
	key := "company:ticker:AAPL"
	exists, err := cacheService.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check key existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("company key should exist but doesn't")
	}
	log.Printf("✅ Key existence check passed")

	// Test TTL
	ttlRemaining, err := cacheService.TTL(ctx, key)
	if err != nil {
		log.Printf("⚠️  TTL check failed (might not be supported): %v", err)
	} else {
		log.Printf("⏰ TTL remaining: %v", ttlRemaining)
	}
	// Test Delete
	log.Printf("🔄 Deleting company: %s", ticker)
	if err := cacheService.DeleteCompany(ctx, ticker); err != nil {
		return fmt.Errorf("failed to delete company: %w", err)
	}
	log.Printf("✅ Company deleted successfully")
	// Verify deletion with more detailed check
	log.Printf("🔄 Verifying company deletion...")
	retrievedAfterDelete, err := cacheService.GetCompany(ctx, ticker)
	if err == nil && retrievedAfterDelete != nil {
		log.Printf("⚠️  Retrieved company after deletion: %+v", retrievedAfterDelete)
		return fmt.Errorf("company should not exist after deletion, but found: %s", retrievedAfterDelete.Ticker)
	}
	log.Printf("✅ Company deletion verified - no company found after deletion")
	// Also verify the key doesn't exist
	exists, err = cacheService.Exists(ctx, key)
	if err != nil {
		log.Printf("⚠️  Could not check key existence after deletion: %v", err)
	} else if exists {
		log.Printf("⚠️  Key still exists after deletion")
		return fmt.Errorf("cache key should not exist after deletion")
	} else {
		log.Printf("✅ Cache key confirmed deleted")
	}

	return nil
}

// testBrokerageOperations tests brokerage cache operations
func testBrokerageOperations(ctx context.Context, cacheService services.CacheService) error {
	// Create test brokerage
	brokerage := entities.NewBrokerage("Goldman Sachs")
	name := "Goldman Sachs"
	ttl := 10 * time.Minute

	// Test Set
	log.Printf("🔄 Setting brokerage: %s", name)
	if err := cacheService.SetBrokerage(ctx, name, brokerage, ttl); err != nil {
		return fmt.Errorf("failed to set brokerage: %w", err)
	}
	log.Printf("✅ Brokerage set successfully")

	// Test Get
	log.Printf("🔄 Getting brokerage: %s", name)
	retrievedBrokerage, err := cacheService.GetBrokerage(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get brokerage: %w", err)
	}

	// Validate data
	if retrievedBrokerage.Name != brokerage.Name {
		return fmt.Errorf("retrieved brokerage data mismatch: got %+v, want %+v", retrievedBrokerage, brokerage)
	}
	log.Printf("✅ Brokerage retrieved successfully: %s", retrievedBrokerage.Name)

	// Test Delete
	log.Printf("🔄 Deleting brokerage: %s", name)
	if err := cacheService.DeleteBrokerage(ctx, name); err != nil {
		return fmt.Errorf("failed to delete brokerage: %w", err)
	}
	log.Printf("✅ Brokerage deleted successfully")

	return nil
}

// testStockRatingOperations tests stock rating cache operations
func testStockRatingOperations(ctx context.Context, cacheService services.CacheService) error {
	// Create test entities
	companyID := uuid.New()
	brokerageID := uuid.New()
	eventTime := time.Now().Add(-1 * time.Hour)

	stockRating := entities.NewStockRating(companyID, brokerageID, "upgraded by", eventTime)
	stockRating.RatingFrom = "Hold"
	stockRating.RatingTo = "Buy"
	stockRating.TargetFrom = "$150.00"
	stockRating.TargetTo = "$175.00"

	ttl := 24 * time.Hour

	// Test Set
	log.Printf("🔄 Setting stock rating for company %s, brokerage %s", companyID, brokerageID)
	if err := cacheService.SetStockRating(ctx, stockRating, ttl); err != nil {
		return fmt.Errorf("failed to set stock rating: %w", err)
	}
	log.Printf("✅ Stock rating set successfully")

	// Test Get
	log.Printf("🔄 Getting stock rating")
	retrievedRating, err := cacheService.GetStockRating(ctx, companyID, brokerageID)
	if err != nil {
		return fmt.Errorf("failed to get stock rating: %w", err)
	}

	// Validate data
	if retrievedRating.CompanyID != stockRating.CompanyID ||
		retrievedRating.BrokerageID != stockRating.BrokerageID ||
		retrievedRating.Action != stockRating.Action {
		return fmt.Errorf("retrieved stock rating data mismatch")
	}
	log.Printf("✅ Stock rating retrieved successfully: %s", retrievedRating.Action)

	// Test Delete
	log.Printf("🔄 Deleting stock rating")
	if err := cacheService.DeleteStockRating(ctx, companyID, brokerageID); err != nil {
		return fmt.Errorf("failed to delete stock rating: %w", err)
	}
	log.Printf("✅ Stock rating deleted successfully")

	return nil
}

// testBulkOperations tests bulk cache operations
func testBulkOperations(ctx context.Context, cacheService services.CacheService) error {
	// Test bulk company operations
	companies := map[string]*entities.Company{
		"AAPL":  entities.NewCompany("AAPL", "Apple Inc."),
		"MSFT":  entities.NewCompany("MSFT", "Microsoft Corporation"),
		"GOOGL": entities.NewCompany("GOOGL", "Alphabet Inc."),
	}

	// Test SetCompanies
	log.Printf("🔄 Setting %d companies in bulk", len(companies))
	if err := cacheService.SetCompanies(ctx, companies, 5*time.Minute); err != nil {
		return fmt.Errorf("failed to set companies in bulk: %w", err)
	}
	log.Printf("✅ Bulk companies set successfully")

	// Test GetCompanies
	tickers := []string{"AAPL", "MSFT", "GOOGL", "NONEXISTENT"}
	log.Printf("🔄 Getting companies in bulk: %v", tickers)
	retrievedCompanies, err := cacheService.GetCompanies(ctx, tickers)
	if err != nil {
		return fmt.Errorf("failed to get companies in bulk: %w", err)
	}

	// Validate results
	expectedCount := 3 // NONEXISTENT should not be found
	if len(retrievedCompanies) != expectedCount {
		return fmt.Errorf("expected %d companies, got %d", expectedCount, len(retrievedCompanies))
	}
	log.Printf("✅ Bulk companies retrieved successfully: %d found", len(retrievedCompanies))

	// Test bulk brokerage operations
	brokerages := map[string]*entities.Brokerage{
		"Goldman Sachs":  entities.NewBrokerage("Goldman Sachs"),
		"Morgan Stanley": entities.NewBrokerage("Morgan Stanley"),
		"JP Morgan":      entities.NewBrokerage("JP Morgan"),
	}

	// Test SetBrokerages
	log.Printf("🔄 Setting %d brokerages in bulk", len(brokerages))
	if err := cacheService.SetBrokerages(ctx, brokerages, 10*time.Minute); err != nil {
		return fmt.Errorf("failed to set brokerages in bulk: %w", err)
	}
	log.Printf("✅ Bulk brokerages set successfully")

	// Test GetBrokerages
	names := []string{"Goldman Sachs", "Morgan Stanley", "JP Morgan"}
	log.Printf("🔄 Getting brokerages in bulk: %v", names)
	retrievedBrokerages, err := cacheService.GetBrokerages(ctx, names)
	if err != nil {
		return fmt.Errorf("failed to get brokerages in bulk: %w", err)
	}

	if len(retrievedBrokerages) != len(names) {
		return fmt.Errorf("expected %d brokerages, got %d", len(names), len(retrievedBrokerages))
	}
	log.Printf("✅ Bulk brokerages retrieved successfully: %d found", len(retrievedBrokerages))

	return nil
}

// testCacheManagement tests cache management operations
func testCacheManagement(ctx context.Context, cacheService services.CacheService) error {
	// First, add some test data
	company := entities.NewCompany("TEST", "Test Company")
	brokerage := entities.NewBrokerage("Test Brokerage")

	cacheService.SetCompany(ctx, "TEST", company, 5*time.Minute)
	cacheService.SetBrokerage(ctx, "Test Brokerage", brokerage, 5*time.Minute)

	// Test ClearCompanies
	log.Printf("🔄 Clearing companies cache")
	if err := cacheService.ClearCompanies(ctx); err != nil {
		return fmt.Errorf("failed to clear companies: %w", err)
	}
	log.Printf("✅ Companies cache cleared successfully")
	// Verify companies are cleared
	retrievedCompany, err := cacheService.GetCompany(ctx, "TEST")
	if err == nil && retrievedCompany != nil {
		return fmt.Errorf("company should not exist after clearing companies cache")
	} // Verify brokerages still exist
	retrievedBrokerage, err := cacheService.GetBrokerage(ctx, "Test Brokerage")
	if err != nil || retrievedBrokerage == nil {
		return fmt.Errorf("brokerage should still exist after clearing companies: %w", err)
	}
	log.Printf("✅ Selective cache clearing verified")

	// Test ClearBrokerages
	log.Printf("🔄 Clearing brokerages cache")
	if err := cacheService.ClearBrokerages(ctx); err != nil {
		return fmt.Errorf("failed to clear brokerages: %w", err)
	}
	log.Printf("✅ Brokerages cache cleared successfully")

	// Add some data again and test full clear
	cacheService.SetCompany(ctx, "TEST", company, 5*time.Minute)
	cacheService.SetBrokerage(ctx, "Test Brokerage", brokerage, 5*time.Minute)

	// Test Clear (full cache clear)
	log.Printf("🔄 Clearing entire cache")
	if err := cacheService.Clear(ctx); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	log.Printf("✅ Entire cache cleared successfully")

	return nil
}

// testCacheStatistics tests cache statistics and performance metrics
func testCacheStatistics(ctx context.Context, cacheService services.CacheService) error {
	// Perform operations to generate cache statistics
	log.Printf("🔄 Generating cache activity for statistics...")

	// Multiple sets and gets to create hit/miss patterns
	for i := 0; i < 5; i++ {
		ticker := fmt.Sprintf("TEST_%d", i)
		testCompany := entities.NewCompany(ticker, fmt.Sprintf("Test Company %d", i))
		cacheService.SetCompany(ctx, ticker, testCompany, 5*time.Minute)
	}

	// Some gets (hits)
	for i := 0; i < 5; i++ {
		ticker := fmt.Sprintf("TEST_%d", i)
		cacheService.GetCompany(ctx, ticker)
	}

	// Some gets for non-existent keys (misses)
	for i := 5; i < 8; i++ {
		ticker := fmt.Sprintf("NONEXISTENT_%d", i)
		cacheService.GetCompany(ctx, ticker)
	}

	// Get final statistics
	log.Printf("🔄 Retrieving cache statistics...")
	stats, err := cacheService.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cache stats: %w", err)
	}

	// Display comprehensive statistics
	log.Printf("📊 === CACHE PERFORMANCE STATISTICS ===")
	log.Printf("📊 Backend: %s", stats.Backend)
	log.Printf("📊 Connected: %t", stats.IsConnected)
	log.Printf("📊 Total Keys: %d", stats.KeyCount)
	log.Printf("📊 Companies Cached: %d", stats.CompanyCount)
	log.Printf("📊 Brokerages Cached: %d", stats.BrokerageCount)
	log.Printf("📊 Hit Count: %d", stats.HitCount)
	log.Printf("📊 Miss Count: %d", stats.MissCount)
	log.Printf("📊 Hit Rate: %.2f%%", stats.HitRate)
	log.Printf("📊 Memory Usage: %d bytes", stats.MemoryUsage)
	log.Printf("📊 Last Access: %s", stats.LastAccess.Format(time.RFC3339))
	log.Printf("📊 Uptime: %s", stats.Uptime)

	// Performance recommendations
	if stats.HitRate < 50.0 {
		log.Printf("⚠️  Low hit rate detected. Consider reviewing cache TTL settings.")
	} else {
		log.Printf("✅ Good hit rate performance")
	}

	if stats.KeyCount > 10000 {
		log.Printf("⚠️  High key count. Consider implementing cache cleanup strategies.")
	}

	return nil
}

// TestCacheWithFallback tests the fallback mechanism specifically
func TestCacheWithFallback(cfg *config.Config) error {
	log.Println("🔄 Testing Cache Fallback Mechanism...")

	// Create cache service with fallback
	cacheService := cache.NewRedisCacheServiceWithFallback(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test basic operations to see if fallback is working
	company := entities.NewCompany("FALLBACK_TEST", "Fallback Test Company")

	log.Printf("🔄 Testing fallback cache operations...")

	// Test Set operation
	if err := cacheService.SetCompany(ctx, "FALLBACK_TEST", company, 5*time.Minute); err != nil {
		return fmt.Errorf("fallback set operation failed: %w", err)
	}
	log.Printf("✅ Fallback set operation successful")

	// Test Get operation
	retrievedCompany, err := cacheService.GetCompany(ctx, "FALLBACK_TEST")
	if err != nil {
		return fmt.Errorf("fallback get operation failed: %w", err)
	}

	if retrievedCompany.Ticker != company.Ticker {
		return fmt.Errorf("fallback data mismatch")
	}
	log.Printf("✅ Fallback get operation successful")

	// Get statistics to see which backend is being used
	stats, err := cacheService.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get fallback stats: %w", err)
	}

	log.Printf("📊 Fallback Cache Backend: %s", stats.Backend)
	log.Printf("📊 Fallback Cache Connected: %t", stats.IsConnected)

	log.Println("✅ Cache fallback mechanism test completed successfully!")
	return nil
}

// TestCachePerformance runs performance benchmarks
func TestCachePerformance(cfg *config.Config) error {
	log.Println("🚀 Testing Cache Performance...")

	cacheService := cache.NewCacheService(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Performance test: Bulk operations
	startTime := time.Now()

	// Generate test data
	companies := make(map[string]*entities.Company)
	for i := 0; i < 100; i++ {
		ticker := fmt.Sprintf("PERF_%d", i)
		companies[ticker] = entities.NewCompany(ticker, fmt.Sprintf("Performance Test Company %d", i))
	}

	// Test bulk set performance
	log.Printf("🔄 Testing bulk set performance with %d companies...", len(companies))
	bulkSetStart := time.Now()
	if err := cacheService.SetCompanies(ctx, companies, 5*time.Minute); err != nil {
		return fmt.Errorf("bulk set performance test failed: %w", err)
	}
	bulkSetDuration := time.Since(bulkSetStart)
	log.Printf("✅ Bulk set completed in %v (%.2f ops/sec)", bulkSetDuration, float64(len(companies))/bulkSetDuration.Seconds())

	// Test bulk get performance
	tickers := make([]string, 0, len(companies))
	for ticker := range companies {
		tickers = append(tickers, ticker)
	}

	log.Printf("🔄 Testing bulk get performance...")
	bulkGetStart := time.Now()
	retrievedCompanies, err := cacheService.GetCompanies(ctx, tickers)
	if err != nil {
		return fmt.Errorf("bulk get performance test failed: %w", err)
	}
	bulkGetDuration := time.Since(bulkGetStart)
	log.Printf("✅ Bulk get completed in %v (%.2f ops/sec)", bulkGetDuration, float64(len(retrievedCompanies))/bulkGetDuration.Seconds())

	totalDuration := time.Since(startTime)
	log.Printf("📊 Total performance test duration: %v", totalDuration)

	// Cleanup
	cacheService.ClearCompanies(ctx)
	log.Println("✅ Cache performance test completed successfully!")
	return nil
}

// TestCachePerformanceDetailed runs comprehensive performance benchmarks with detailed analysis
func TestCachePerformanceDetailed(cfg *config.Config) error {
	log.Println("🚀 Testing Cache Performance (Detailed Analysis)...")

	cacheService := cache.NewCacheService(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Performance benchmarks with different dataset sizes
	dataSizes := []int{10, 50, 100, 500}

	log.Println("\n📊 === PERFORMANCE BENCHMARK RESULTS ===")

	for _, size := range dataSizes {
		log.Printf("\n🔄 Testing with %d items...", size)

		// Generate test data
		companies := make(map[string]*entities.Company)
		for i := 0; i < size; i++ {
			ticker := fmt.Sprintf("PERF_%d_%d", size, i)
			companies[ticker] = entities.NewCompany(ticker, fmt.Sprintf("Perf Company %d", i))
		}

		// Test individual operations
		log.Printf("  📝 Individual SET operations:")
		individualSetStart := time.Now()
		for ticker, company := range companies {
			if err := cacheService.SetCompany(ctx, ticker, company, 5*time.Minute); err != nil {
				return fmt.Errorf("individual set failed: %w", err)
			}
		}
		individualSetDuration := time.Since(individualSetStart)
		individualSetOps := float64(size) / individualSetDuration.Seconds()
		log.Printf("    ⏱️  %v total (%.2f ops/sec, %.2fms per op)",
			individualSetDuration, individualSetOps, float64(individualSetDuration.Nanoseconds())/float64(size)/1e6)

		// Test bulk operations
		log.Printf("  📦 Bulk SET operations:")
		bulkSetStart := time.Now()
		if err := cacheService.SetCompanies(ctx, companies, 5*time.Minute); err != nil {
			return fmt.Errorf("bulk set failed: %w", err)
		}
		bulkSetDuration := time.Since(bulkSetStart)
		bulkSetOps := float64(size) / bulkSetDuration.Seconds()
		log.Printf("    ⏱️  %v total (%.2f ops/sec, %.2fms per op)",
			bulkSetDuration, bulkSetOps, float64(bulkSetDuration.Nanoseconds())/float64(size)/1e6)

		// Calculate bulk vs individual performance improvement
		improvement := (individualSetDuration.Seconds() - bulkSetDuration.Seconds()) / individualSetDuration.Seconds() * 100
		log.Printf("    🚀 Bulk operations are %.1f%% faster than individual", improvement)

		// Test GET operations
		tickers := make([]string, 0, len(companies))
		for ticker := range companies {
			tickers = append(tickers, ticker)
		}

		log.Printf("  📖 Individual GET operations:")
		individualGetStart := time.Now()
		for _, ticker := range tickers {
			if _, err := cacheService.GetCompany(ctx, ticker); err != nil {
				return fmt.Errorf("individual get failed: %w", err)
			}
		}
		individualGetDuration := time.Since(individualGetStart)
		individualGetOps := float64(size) / individualGetDuration.Seconds()
		log.Printf("    ⏱️  %v total (%.2f ops/sec, %.2fms per op)",
			individualGetDuration, individualGetOps, float64(individualGetDuration.Nanoseconds())/float64(size)/1e6)

		log.Printf("  📦 Bulk GET operations:")
		bulkGetStart := time.Now()
		retrievedCompanies, err := cacheService.GetCompanies(ctx, tickers)
		if err != nil {
			return fmt.Errorf("bulk get failed: %w", err)
		}
		bulkGetDuration := time.Since(bulkGetStart)
		bulkGetOps := float64(len(retrievedCompanies)) / bulkGetDuration.Seconds()
		log.Printf("    ⏱️  %v total (%.2f ops/sec, %.2fms per op)",
			bulkGetDuration, bulkGetOps, float64(bulkGetDuration.Nanoseconds())/float64(len(retrievedCompanies))/1e6)

		// Calculate GET bulk vs individual performance improvement
		getImprovement := (individualGetDuration.Seconds() - bulkGetDuration.Seconds()) / individualGetDuration.Seconds() * 100
		log.Printf("    🚀 Bulk GET operations are %.1f%% faster than individual", getImprovement)

		// Read/Write ratio analysis
		readWriteRatio := individualGetOps / individualSetOps
		log.Printf("  📊 Read/Write Performance Ratio: %.2fx (reads are %.1fx faster)",
			readWriteRatio, readWriteRatio)

		// Cleanup
		cacheService.ClearCompanies(ctx)
	}

	// Network latency test
	log.Printf("\n🌐 Network Latency Analysis:")
	latencyTests := 10
	var totalLatency time.Duration

	for i := 0; i < latencyTests; i++ {
		start := time.Now()
		cacheService.Ping(ctx)
		latency := time.Since(start)
		totalLatency += latency
	}

	avgLatency := totalLatency / time.Duration(latencyTests)
	log.Printf("  📡 Average ping latency: %v", avgLatency)
	log.Printf("  🔍 Theoretical max ops/sec (ping-limited): %.0f", 1000.0/float64(avgLatency.Milliseconds()))

	// Performance recommendations
	log.Printf("\n💡 Performance Analysis:")
	if avgLatency < 10*time.Millisecond {
		log.Printf("  ✅ Excellent network latency (<%v)", 10*time.Millisecond)
	} else if avgLatency < 50*time.Millisecond {
		log.Printf("  ⚠️  Moderate network latency (%v)", avgLatency)
	} else {
		log.Printf("  ❌ High network latency (%v) - consider using connection pooling", avgLatency)
	}

	log.Println("\n✅ Detailed cache performance analysis completed!")
	return nil
}
