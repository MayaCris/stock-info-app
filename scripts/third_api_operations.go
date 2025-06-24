package scripts

import (
	"context"
	"fmt"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// initScriptLogger initializes a logger for script operations
func initScriptLogger(context string) (logger.Logger, error) {
	scriptLogger, err := logger.InitializeGlobalLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return scriptLogger.WithContext(context), nil
}

// TestAPIConnection tests the connection to the external stock API
func TestAPIConnection(cfg *config.Config) error {
	apiLogger, err := initScriptLogger("API_TEST")
	if err != nil {
		return err
	}
	defer func() { _ = apiLogger.Close() }()

	ctx := context.Background()

	apiLogger.Info(ctx, "🚀 Testing Stock API Connection",
		logger.String("component", "api_test"))

	// Create API client
	client := stock_api.NewClient(cfg)

	// Print client stats
	stats := client.GetStats()
	apiLogger.Info(ctx, "📊 Client configuration loaded",
		logger.String("component", "api_client"),
		logger.Any("stats", stats))

	// Create context with timeout
	testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Fetch first page only
	apiLogger.Info(ctx, "=== TEST 1: Basic Connectivity ===",
		logger.String("test_phase", "basic_connectivity"))

	firstPage, err := client.FetchPage(testCtx, "")
	if err != nil {
		apiLogger.Error(ctx, "❌ Failed to fetch first page", err,
			logger.String("test_phase", "basic_connectivity"))
		return fmt.Errorf("❌ Failed to fetch first page: %w", err)
	}

	apiLogger.Info(ctx, "✅ API connection successful!",
		logger.String("test_phase", "basic_connectivity"),
		logger.Int("items_count", firstPage.GetItemCount()),
		logger.Bool("has_next_page", firstPage.HasNextPage()))

	if firstPage.HasNextPage() {
		apiLogger.Info(ctx, "➡️ Next page available",
			logger.String("next_page_key", firstPage.NextPage))
	}

	// Show sample items
	if firstPage.GetItemCount() > 0 {
		apiLogger.Info(ctx, "📋 Sample items from API",
			logger.String("operation", "sample_display"))

		for i, item := range firstPage.Items {
			if i >= 3 { // Show only first 3 items
				break
			}
			eventTime, _ := item.GetEventTime()
			apiLogger.Info(ctx, fmt.Sprintf("  %d. %s (%s) - %s by %s [%s]",
				i+1, item.Ticker, item.Company, item.Action,
				item.Brokerage, eventTime.Format("2006-01-02 15:04")),
				logger.String("operation", "sample_item"),
				logger.String("ticker", item.Ticker),
				logger.String("company", item.Company),
				logger.String("action", item.Action),
				logger.String("brokerage", item.Brokerage))
		}
	}

	apiLogger.Info(ctx, "✅ API connection test completed successfully!",
		logger.String("test_phase", "completion"))
	return nil
}

// TestAPIConnectionDetailed performs a more comprehensive test
func TestAPIConnectionDetailed(cfg *config.Config) error {
	apiLogger, err := initScriptLogger("API_TEST_DETAILED")
	if err != nil {
		return err
	}
	defer func() { _ = apiLogger.Close() }()

	ctx := context.Background()

	apiLogger.Info(ctx, "🔍 Testing Stock API Connection (Detailed)",
		logger.String("component", "api_test_detailed"))

	client := stock_api.NewClient(cfg)
	testCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Test with multiple pages
	apiLogger.Info(ctx, "=== DETAILED TEST: Multiple Pages ===",
		logger.String("test_phase", "multiple_pages"))

	recentItems, err := client.FetchRecentPages(testCtx, 5) // Fetch 5 pages max
	if err != nil {
		apiLogger.Error(ctx, "❌ Failed to fetch multiple pages", err,
			logger.String("test_phase", "multiple_pages"))
		return fmt.Errorf("❌ Failed to fetch multiple pages: %w", err)
	}

	apiLogger.Info(ctx, "✅ Multiple pages fetched successfully!",
		logger.String("test_phase", "multiple_pages"),
		logger.Int("total_items", len(recentItems)))

	// Perform analysis
	if len(recentItems) > 0 {
		apiLogger.Info(ctx, "📈 Data Analysis",
			logger.String("operation", "data_analysis"))

		// Count by action type
		actionCounts := make(map[string]int)
		companyCounts := make(map[string]int)
		brokerageCounts := make(map[string]int)

		for _, item := range recentItems {
			actionCounts[item.Action]++
			companyCounts[item.Company]++
			brokerageCounts[item.Brokerage]++
		}

		apiLogger.Info(ctx, "📊 Action breakdown",
			logger.String("operation", "action_analysis"),
			logger.Any("action_counts", actionCounts))

		apiLogger.Info(ctx, "🏢 Company statistics",
			logger.String("operation", "company_analysis"),
			logger.Int("unique_companies", len(companyCounts)))

		apiLogger.Info(ctx, "🏦 Brokerage statistics",
			logger.String("operation", "brokerage_analysis"),
			logger.Int("unique_brokerages", len(brokerageCounts)))

		// Show most active companies
		apiLogger.Info(ctx, "🏆 Most active companies",
			logger.String("operation", "top_companies"))
		count := 0
		for company, ratings := range companyCounts {
			if count >= 5 { // Show top 5
				break
			}
			apiLogger.Info(ctx, fmt.Sprintf("  %d. %s: %d ratings", count+1, company, ratings),
				logger.String("operation", "top_company"),
				logger.String("company", company),
				logger.Int("rating_count", ratings))
			count++
		}

		// Show most active brokerages
		apiLogger.Info(ctx, "🏅 Most active brokerages",
			logger.String("operation", "top_brokerages"))
		count = 0
		for brokerage, ratings := range brokerageCounts {
			if count >= 5 { // Show top 5
				break
			}
			apiLogger.Info(ctx, fmt.Sprintf("  %d. %s: %d ratings", count+1, brokerage, ratings),
				logger.String("operation", "top_brokerage"),
				logger.String("brokerage", brokerage),
				logger.Int("rating_count", ratings))
			count++
		}
	}

	apiLogger.Info(ctx, "✅ Detailed API connection test completed successfully!",
		logger.String("test_phase", "completion"))
	return nil
}

// ValidateAPIResponse validates the structure and content of API responses
func ValidateAPIResponse(cfg *config.Config) error {
	apiLogger, err := initScriptLogger("API_VALIDATION")
	if err != nil {
		return err
	}
	defer func() { _ = apiLogger.Close() }()

	ctx := context.Background()

	apiLogger.Info(ctx, "🔍 Validating API Response Structure",
		logger.String("component", "api_validation"))

	client := stock_api.NewClient(cfg)
	testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch first page
	response, err := client.FetchPage(testCtx, "")
	if err != nil {
		apiLogger.Error(ctx, "❌ Failed to fetch page for validation", err,
			logger.String("test_phase", "fetch"))
		return fmt.Errorf("❌ Failed to fetch page for validation: %w", err)
	}

	apiLogger.Info(ctx, "=== VALIDATION RESULTS ===",
		logger.String("test_phase", "validation"))

	// Validate response structure
	apiLogger.Info(ctx, "✅ Response structure valid",
		logger.String("validation", "structure"),
		logger.Int("items_count", response.GetItemCount()),
		logger.Bool("has_pagination", response.HasNextPage()))

	// Validate individual items
	validItems := 0
	invalidItems := 0

	for i, item := range response.Items {
		if item.IsValid() {
			validItems++
		} else {
			invalidItems++
			apiLogger.Warn(ctx, fmt.Sprintf("⚠️ Invalid item %d", i+1),
				logger.String("validation", "item_invalid"),
				logger.Any("item", item))
		}

		// Test time parsing
		if item.Time != "" {
			if _, err := item.GetEventTime(); err != nil {
				apiLogger.Warn(ctx, fmt.Sprintf("⚠️ Item %d has invalid time format", i+1),
					logger.String("validation", "time_format"),
					logger.String("time_value", item.Time),
					logger.String("error", err.Error()))
			}
		}
	}

	apiLogger.Info(ctx, "Item validation results",
		logger.String("validation", "items"),
		logger.Int("valid_items", validItems),
		logger.Int("invalid_items", invalidItems))

	// Validate required fields
	apiLogger.Info(ctx, "📋 Field validation",
		logger.String("operation", "field_validation"))

	requiredFields := []string{"ticker", "company", "brokerage", "action", "time"}

	for _, field := range requiredFields {
		missingCount := 0
		for _, item := range response.Items {
			switch field {
			case "ticker":
				if item.Ticker == "" {
					missingCount++
				}
			case "company":
				if item.Company == "" {
					missingCount++
				}
			case "brokerage":
				if item.Brokerage == "" {
					missingCount++
				}
			case "action":
				if item.Action == "" {
					missingCount++
				}
			case "time":
				if item.Time == "" {
					missingCount++
				}
			}
		}

		if missingCount == 0 {
			apiLogger.Info(ctx, fmt.Sprintf("✅ %s: All items have this field", field),
				logger.String("field_validation", field),
				logger.String("status", "complete"))
		} else {
			apiLogger.Warn(ctx, fmt.Sprintf("⚠️ %s: %d items missing this field", field, missingCount),
				logger.String("field_validation", field),
				logger.String("status", "incomplete"),
				logger.Int("missing_count", missingCount))
		}
	}

	apiLogger.Info(ctx, "✅ API response validation completed!",
		logger.String("test_phase", "completion"))
	return nil
}

// GetAPIStats returns statistics about the API without fetching all data
func GetAPIStats(cfg *config.Config) error {
	apiLogger, err := initScriptLogger("API_STATS")
	if err != nil {
		return err
	}
	defer func() { _ = apiLogger.Close() }()

	ctx := context.Background()

	apiLogger.Info(ctx, "📊 Getting API Statistics",
		logger.String("component", "api_stats"))

	client := stock_api.NewClient(cfg)
	testCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Fetch first few pages to estimate total
	const samplePages = 10
	allItems, err := client.FetchRecentPages(testCtx, samplePages)
	if err != nil {
		apiLogger.Error(ctx, "❌ Failed to fetch sample pages", err,
			logger.String("operation", "sample_fetch"))
		return fmt.Errorf("❌ Failed to fetch sample pages: %w", err)
	}

	apiLogger.Info(ctx, "✅ Sample data collected",
		logger.String("operation", "sample_complete"),
		logger.Int("sampled_pages", samplePages),
		logger.Int("sample_items", len(allItems)))

	if len(allItems) > 0 {
		// Calculate average items per page
		avgPerPage := float64(len(allItems)) / float64(samplePages)
		apiLogger.Info(ctx, "📈 API statistics",
			logger.String("operation", "statistics"),
			logger.Float64("avg_items_per_page", avgPerPage))

		// Time range analysis
		if len(allItems) >= 2 {
			firstTime, _ := allItems[0].GetEventTime()
			lastTime, _ := allItems[len(allItems)-1].GetEventTime()

			apiLogger.Info(ctx, "📅 Time range analysis",
				logger.String("operation", "time_analysis"),
				logger.Time("earliest_event", firstTime),
				logger.Time("latest_event", lastTime))
		}

		// Data quality metrics
		completeItems := 0
		for _, item := range allItems {
			if item.IsValid() && item.Ticker != "" && item.Company != "" &&
				item.Brokerage != "" && item.Action != "" && item.Time != "" {
				completeItems++
			}
		}

		completenessRate := float64(completeItems) / float64(len(allItems)) * 100
		apiLogger.Info(ctx, "📊 Data quality metrics",
			logger.String("operation", "quality_analysis"),
			logger.Int("complete_items", completeItems),
			logger.Int("total_items", len(allItems)),
			logger.Float64("completeness_rate", completenessRate))
	}

	apiLogger.Info(ctx, "✅ API statistics collection completed!",
		logger.String("operation", "completion"))
	return nil
}
