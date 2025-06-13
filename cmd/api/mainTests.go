package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	"github.com/MayaCris/stock-info-app/scripts"
)

func main() {
	// Command line flags
	var (
		testAPI               = flag.Bool("test-api", false, "Test API connection")
		testAPIDetailed       = flag.Bool("test-api-detailed", false, "Test API connection with detailed analysis")
		validateAPI           = flag.Bool("validate-api", false, "Validate API response structure")
		apiStats              = flag.Bool("api-stats", false, "Get API statistics")
		help                  = flag.Bool("help", false, "Show help message")
		mapperTests           = flag.Bool("test-mapper", false, "Run mapper tests")
		mapperTestsApi        = flag.Bool("test-mapper-api", false, "Run API mapper tests")
		mapperExample         = flag.Bool("mapper-example", false, "Run example mapper test")
		testCache             = flag.Bool("test-cache", false, "Test cache operations")
		testCacheFallback     = flag.Bool("test-cache-fallback", false, "Test cache fallback mechanism")
		testCachePerf         = flag.Bool("test-cache-performance", false, "Run cache performance tests")
		testCachePerfDetailed = flag.Bool("test-cache-performance-detailed", false, "Run detailed cache performance analysis")
	)
	flag.Parse()

	// Show help
	if *help {
		showHelp()
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Execute API testing scripts if requested
	if *testAPI {
		if err := scripts.TestAPIConnection(cfg); err != nil {
			log.Fatalf("❌ API test failed: %v", err)
		}
		return
	}

	if *testAPIDetailed {
		if err := scripts.TestAPIConnectionDetailed(cfg); err != nil {
			log.Fatalf("❌ Detailed API test failed: %v", err)
		}
		return
	}

	if *validateAPI {
		if err := scripts.ValidateAPIResponse(cfg); err != nil {
			log.Fatalf("❌ API validation failed: %v", err)
		}
		return
	}

	if *apiStats {
		if err := scripts.GetAPIStats(cfg); err != nil {
			log.Fatalf("❌ API stats failed: %v", err)
		}
		return
	}

	if *mapperTests {
		if err := scripts.TestMapper(cfg); err != nil {
			log.Fatalf("❌ Mapper tests failed: %v", err)
		}
		return
	}

	if *mapperTestsApi {
		if err := scripts.TestMapperWithRealAPI(cfg); err != nil {
			log.Fatalf("❌ Mapper API tests failed: %v", err)
		}
		return
	}
	if *mapperExample {
		if err := scripts.ShowMappingExample(cfg); err != nil {
			log.Fatalf("❌ Mapper example failed: %v", err)
		}
		return
	}

	if *testCache {
		if err := scripts.TestCacheOperations(cfg); err != nil {
			log.Fatalf("❌ Cache operations test failed: %v", err)
		}
		return
	}

	if *testCacheFallback {
		if err := scripts.TestCacheWithFallback(cfg); err != nil {
			log.Fatalf("❌ Cache fallback test failed: %v", err)
		}
		return
	}
	if *testCachePerf {
		if err := scripts.TestCachePerformance(cfg); err != nil {
			log.Fatalf("❌ Cache performance test failed: %v", err)
		}
		return
	}

	if *testCachePerfDetailed {
		if err := scripts.TestCachePerformanceDetailed(cfg); err != nil {
			log.Fatalf("❌ Detailed cache performance test failed: %v", err)
		}
		return
	}

	// Default behavior: Start the application
	startApplication(cfg)
}

func startApplication(cfg *config.Config) {
	log.Printf("🚀 Starting %s in %s mode on port %s",
		cfg.App.Name, cfg.App.Env, cfg.App.Port)

	// Log API configuration
	log.Printf("📊 Primary API: %s (%s)", cfg.External.Primary.Name, cfg.External.Primary.BaseURL)
	log.Printf("📈 Secondary API: %s (%s)", cfg.External.Secondary.Name, cfg.External.Secondary.BaseURL)

	// Initialize database connection
	db, err := cockroachdb.NewConnection(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("❌ Error closing database: %v", err)
		}
	}()

	// Test database health
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if !db.IsHealthy(ctx) {
		log.Fatal("❌ Database health check failed")
	}

	// Print database stats
	if stats, err := db.GetStats(); err == nil {
		log.Printf("📊 Database stats: %+v", stats)
	}

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✅ Application started successfully")
	log.Println("💡 Use --help to see available commands")
	log.Println("Press Ctrl+C to shutdown...")

	// Wait for shutdown signal
	<-quit
	log.Println("🔄 Shutting down gracefully...")

	// Here you would add cleanup for your HTTP server, workers, etc.

	log.Println("👋 Application stopped")
}

func showHelp() {
	log.Println("📖 Stock System Backend - Available Commands:")
	log.Println("")
	log.Println("🚀 Application:")
	log.Println("  go run cmd/api/main.go                    Start the application")
	log.Println("")
	log.Println("🔧 API Testing:")
	log.Println("  go run cmd/api/main.go --test-api         Test basic API connection")
	log.Println("  go run cmd/api/main.go --test-api-detailed Test API with detailed analysis")
	log.Println("  go run cmd/api/main.go --validate-api     Validate API response structure")
	log.Println("  go run cmd/api/main.go --api-stats        Get API statistics")
	log.Println("")
	log.Println("🗃️  Mapper Testing:")
	log.Println("  go run cmd/api/main.go --test-mapper      Test mapper functionality")
	log.Println("  go run cmd/api/main.go --test-mapper-api  Test mapper with real API data")
	log.Println("  go run cmd/api/main.go --mapper-example   Show mapper usage example")
	log.Println("")
	log.Println("💾 Cache Testing:")
	log.Println("  go run cmd/api/main.go --test-cache       Test cache operations (Redis + Memory)")
	log.Println("  go run cmd/api/main.go --test-cache-fallback Test cache fallback mechanism")
	log.Println("  go run cmd/api/main.go --test-cache-performance Run cache performance tests")
	log.Println("  go run cmd/api/main.go --test-cache-performance-detailed Run detailed cache performance analysis")
	log.Println("")
	log.Println("❓ Help:")
	log.Println("  go run cmd/api/main.go --help             Show this help message")
	log.Println("")
	log.Println("📋 Examples:")
	log.Println("  # Test API connection")
	log.Println("  go run cmd/api/main.go --test-api")
	log.Println("")
	log.Println("  # Test cache operations")
	log.Println("  go run cmd/api/main.go --test-cache")
	log.Println("")
	log.Println("  # Test complete mapper functionality")
	log.Println("  go run cmd/api/main.go --test-mapper")
	log.Println("")
	log.Println("  # Start application normally")
	log.Println("  go run cmd/api/main.go")
}
