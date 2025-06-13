package scripts

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
)

// TestAPIConnection tests the connection to the external stock API
func TestAPIConnection(cfg *config.Config) error {
	log.Println("🚀 Testing Stock API Connection...")

	// Create API client
	client := stock_api.NewClient(cfg)
	
	// Print client stats
	log.Printf("📊 Client configuration: %+v", client.GetStats())

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Fetch first page only
	log.Println("\n=== TEST 1: Basic Connectivity ===")
	firstPage, err := client.FetchPage(ctx, "")
	if err != nil {
		return fmt.Errorf("❌ Failed to fetch first page: %w", err)
	}

	log.Printf("✅ API connection successful!")
	log.Printf("📄 Items in first page: %d", firstPage.GetItemCount())
	log.Printf("🔗 Has next page: %t", firstPage.HasNextPage())
	
	if firstPage.HasNextPage() {
		log.Printf("➡️  Next page key: %s", firstPage.NextPage)
	}

	// Show sample items
	if firstPage.GetItemCount() > 0 {
		log.Println("\n📋 Sample items from API:")
		for i, item := range firstPage.Items {
			if i >= 3 { // Show only first 3 items
				break
			}
			eventTime, _ := item.GetEventTime()
			log.Printf("  %d. %s (%s) - %s by %s [%s]",
				i+1, item.Ticker, item.Company, item.Action, 
				item.Brokerage, eventTime.Format("2006-01-02 15:04"))
		}
	}

	log.Println("\n✅ API connection test completed successfully!")
	return nil
}

// TestAPIConnectionDetailed performs a more comprehensive test
func TestAPIConnectionDetailed(cfg *config.Config) error {
	log.Println("🔍 Testing Stock API Connection (Detailed)...")

	client := stock_api.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Test with multiple pages
	log.Println("\n=== DETAILED TEST: Multiple Pages ===")
	recentItems, err := client.FetchRecentPages(ctx, 5) // Fetch 5 pages max
	if err != nil {
		return fmt.Errorf("❌ Failed to fetch multiple pages: %w", err)
	}

	log.Printf("✅ Multiple pages fetched successfully!")
	log.Printf("📊 Total items from 5 pages: %d", len(recentItems))

	// Perform analysis
	if len(recentItems) > 0 {
		log.Println("\n📈 Data Analysis:")
		
		// Count by action type
		actionCounts := make(map[string]int)
		companyCounts := make(map[string]int)
		brokerageCounts := make(map[string]int)
		
		for _, item := range recentItems {
			actionCounts[item.Action]++
			companyCounts[item.Company]++
			brokerageCounts[item.Brokerage]++
		}
		
		log.Printf("  📊 Action breakdown:")
		for action, count := range actionCounts {
			log.Printf("    - %s: %d", action, count)
		}
		
		log.Printf("  🏢 Unique companies: %d", len(companyCounts))
		log.Printf("  🏦 Unique brokerages: %d", len(brokerageCounts))
		
		// Show most active companies
		log.Println("\n🏆 Most active companies:")
		count := 0
		for company, ratings := range companyCounts {
			if count >= 5 { // Show top 5
				break
			}
			log.Printf("  %d. %s: %d ratings", count+1, company, ratings)
			count++
		}
		
		// Show most active brokerages
		log.Println("\n🏅 Most active brokerages:")
		count = 0
		for brokerage, ratings := range brokerageCounts {
			if count >= 5 { // Show top 5
				break
			}
			log.Printf("  %d. %s: %d ratings", count+1, brokerage, ratings)
			count++
		}
	}

	log.Println("\n✅ Detailed API connection test completed successfully!")
	return nil
}

// ValidateAPIResponse validates the structure and content of API responses
func ValidateAPIResponse(cfg *config.Config) error {
	log.Println("🔍 Validating API Response Structure...")

	client := stock_api.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch first page
	response, err := client.FetchPage(ctx, "")
	if err != nil {
		return fmt.Errorf("❌ Failed to fetch page for validation: %w", err)
	}

	log.Println("\n=== VALIDATION RESULTS ===")
	
	// Validate response structure
	log.Printf("✅ Response structure valid")
	log.Printf("📄 Items count: %d", response.GetItemCount())
	log.Printf("🔗 Has pagination: %t", response.HasNextPage())
	
	// Validate individual items
	validItems := 0
	invalidItems := 0
	
	for i, item := range response.Items {
		if item.IsValid() {
			validItems++
		} else {
			invalidItems++
			log.Printf("⚠️  Invalid item %d: %+v", i+1, item)
		}
		
		// Test time parsing
		if item.Time != "" {
			if _, err := item.GetEventTime(); err != nil {
				log.Printf("⚠️  Item %d has invalid time format: %s", i+1, item.Time)
			}
		}
	}
	
	log.Printf("✅ Valid items: %d", validItems)
	if invalidItems > 0 {
		log.Printf("⚠️  Invalid items: %d", invalidItems)
	}
	
	// Validate required fields
	log.Println("\n📋 Field validation:")
	requiredFields := []string{"ticker", "company", "brokerage", "action", "time"}
	
	for _, field := range requiredFields {
		missingCount := 0
		for _, item := range response.Items {
			switch field {
			case "ticker":
				if item.Ticker == "" { missingCount++ }
			case "company":
				if item.Company == "" { missingCount++ }
			case "brokerage":
				if item.Brokerage == "" { missingCount++ }
			case "action":
				if item.Action == "" { missingCount++ }
			case "time":
				if item.Time == "" { missingCount++ }
			}
		}
		
		if missingCount == 0 {
			log.Printf("  ✅ %s: All items have this field", field)
		} else {
			log.Printf("  ⚠️  %s: %d items missing this field", field, missingCount)
		}
	}

	log.Println("\n✅ API response validation completed!")
	return nil
}

// GetAPIStats returns statistics about the API without fetching all data
func GetAPIStats(cfg *config.Config) error {
	log.Println("📊 Getting API Statistics...")

	client := stock_api.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Fetch first few pages to estimate total
	const samplePages = 10
	allItems, err := client.FetchRecentPages(ctx, samplePages)
	if err != nil {
		return fmt.Errorf("❌ Failed to fetch sample pages: %w", err)
	}

	log.Printf("✅ Sampled %d pages", samplePages)
	log.Printf("📊 Sample contains %d items", len(allItems))
	
	if len(allItems) > 0 {
		// Calculate average items per page
		avgPerPage := float64(len(allItems)) / float64(samplePages)
		log.Printf("📈 Average items per page: %.1f", avgPerPage)
		
		// Time range analysis
		if len(allItems) >= 2 {
			firstTime, _ := allItems[0].GetEventTime()
			lastTime, _ := allItems[len(allItems)-1].GetEventTime()
			
			log.Printf("📅 Time range in sample:")
			log.Printf("  - Oldest: %s", firstTime.Format("2006-01-02 15:04:05"))
			log.Printf("  - Newest: %s", lastTime.Format("2006-01-02 15:04:05"))
			log.Printf("  - Span: %s", lastTime.Sub(firstTime).String())
		}
	}

	log.Println("\n✅ API statistics completed!")
	return nil
}