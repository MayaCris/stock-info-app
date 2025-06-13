package interfacesMap

import (
	"fmt"
	"context"
	"encoding/json"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
)

// APIMapper defines the contract for mapping API responses to domain entities
type APIMapper interface {
	// MapStockRatingItem maps a single API item to domain entities
	// Returns the created/found entities and any mapping errors
	MapStockRatingItem(ctx context.Context, item stock_api.StockRatingItem) (*MappingResult, error)
	
	// MapStockRatingItems maps multiple API items to domain entities
	// Returns results for each item, including successes and failures
	MapStockRatingItems(ctx context.Context, items []stock_api.StockRatingItem) (*BatchMappingResult, error)
	
	// ValidateAPIItem validates an API item before mapping
	ValidateAPIItem(item stock_api.StockRatingItem) error
}

// MappingResult holds the result of mapping a single API item
type MappingResult struct {
	Company      *entities.Company      `json:"company"`
	Brokerage    *entities.Brokerage    `json:"brokerage"`
	StockRating  *entities.StockRating  `json:"stock_rating"`
	
	// Metadata
	WasCompanyCreated   bool `json:"was_company_created"`
	WasBrokerageCreated bool `json:"was_brokerage_created"`
	WasRatingCreated    bool `json:"was_rating_created"`
	
	// Original data for debugging
	OriginalItem stock_api.StockRatingItem `json:"original_item"`
	RawJSON      json.RawMessage           `json:"raw_json,omitempty"`
}

// BatchMappingResult holds the results of mapping multiple API items
type BatchMappingResult struct {
	SuccessfulMappings []MappingResult `json:"successful_mappings"`
	FailedMappings     []FailedMapping `json:"failed_mappings"`
	
	// Statistics
	TotalItems       int `json:"total_items"`
	SuccessCount     int `json:"success_count"`
	FailureCount     int `json:"failure_count"`
	
	// Entity statistics
	CompaniesCreated   int `json:"companies_created"`
	BrokeragesCreated  int `json:"brokerages_created"`
	RatingsCreated     int `json:"ratings_created"`
	DuplicateRatings   int `json:"duplicate_ratings"`
}

// FailedMapping holds information about a failed mapping attempt
type FailedMapping struct {
	OriginalItem stock_api.StockRatingItem `json:"original_item"`
	Error        string                    `json:"error"`
	ErrorType    string                    `json:"error_type"` // "validation", "company", "brokerage", "rating"
}

// MappingStats provides detailed statistics about the mapping process
type MappingStats struct {
	ProcessingTime      string            `json:"processing_time"`
	ItemsPerSecond      float64           `json:"items_per_second"`
	ErrorBreakdown      map[string]int    `json:"error_breakdown"`
	UniqueCompanies     int               `json:"unique_companies"`
	UniqueBrokerages    int               `json:"unique_brokerages"`
	ActionDistribution  map[string]int    `json:"action_distribution"`
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// GetMappingStats returns detailed statistics from a batch mapping result
func (bmr *BatchMappingResult) GetMappingStats() MappingStats {
	stats := MappingStats{
		ErrorBreakdown:     make(map[string]int),
		ActionDistribution: make(map[string]int),
	}
	
	// Count error types
	for _, failed := range bmr.FailedMappings {
		stats.ErrorBreakdown[failed.ErrorType]++
	}
	
	// Count action types
	for _, success := range bmr.SuccessfulMappings {
		action := success.StockRating.Action
		stats.ActionDistribution[action]++
	}
	
	// Count unique entities
	uniqueCompanies := make(map[string]bool)
	uniqueBrokerages := make(map[string]bool)
	
	for _, success := range bmr.SuccessfulMappings {
		uniqueCompanies[success.Company.Ticker] = true
		uniqueBrokerages[success.Brokerage.Name] = true
	}
	
	stats.UniqueCompanies = len(uniqueCompanies)
	stats.UniqueBrokerages = len(uniqueBrokerages)
	
	return stats
}

// HasErrors checks if there were any mapping errors
func (bmr *BatchMappingResult) HasErrors() bool {
	return bmr.FailureCount > 0
}

// GetSuccessRate returns the success rate as a percentage
func (bmr *BatchMappingResult) GetSuccessRate() float64 {
	if bmr.TotalItems == 0 {
		return 0.0
	}
	return (float64(bmr.SuccessCount) / float64(bmr.TotalItems)) * 100
}

// Summary returns a human-readable summary of the mapping results
func (bmr *BatchMappingResult) Summary() string {
	return fmt.Sprintf(
		"Mapped %d/%d items (%.1f%% success). Created: %d companies, %d brokerages, %d ratings. Duplicates: %d",
		bmr.SuccessCount, bmr.TotalItems, bmr.GetSuccessRate(),
		bmr.CompaniesCreated, bmr.BrokeragesCreated, bmr.RatingsCreated, bmr.DuplicateRatings,
	)
}