package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
)

// CacheService defines the contract for caching operations
type CacheService interface {
	// Company cache operations
	GetCompany(ctx context.Context, ticker string) (*entities.Company, error)
	SetCompany(ctx context.Context, ticker string, company *entities.Company, ttl time.Duration) error
	DeleteCompany(ctx context.Context, ticker string) error
	
	// Brokerage cache operations
	GetBrokerage(ctx context.Context, name string) (*entities.Brokerage, error)
	SetBrokerage(ctx context.Context, name string, brokerage *entities.Brokerage, ttl time.Duration) error
	DeleteBrokerage(ctx context.Context, name string) error

	// Stock rating cache operations
	GetStockRating(ctx context.Context, companyID uuid.UUID, brokerageID uuid.UUID) (*entities.StockRating, error)
	SetStockRating(ctx context.Context, stockRating *entities.StockRating, ttl time.Duration) error
	DeleteStockRating(ctx context.Context, companyID uuid.UUID, brokerageID uuid.UUID) error
	
	// Bulk operations for performance
	GetCompanies(ctx context.Context, tickers []string) (map[string]*entities.Company, error)
	SetCompanies(ctx context.Context, companies map[string]*entities.Company, ttl time.Duration) error
	GetBrokerages(ctx context.Context, names []string) (map[string]*entities.Brokerage, error)
	SetBrokerages(ctx context.Context, brokerages map[string]*entities.Brokerage, ttl time.Duration) error
	
	// Cache management operations
	Clear(ctx context.Context) error
	ClearCompanies(ctx context.Context) error
	ClearBrokerages(ctx context.Context) error
	
	// Cache statistics and health
	GetStats(ctx context.Context) (CacheStats, error)
	Ping(ctx context.Context) error
	
	// Key operations
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// CacheStats represents cache statistics
type CacheStats struct {
	// Connection info
	IsConnected bool   `json:"is_connected"`
	Backend     string `json:"backend"` // "redis", "memory", etc.
	
	// Performance metrics
	HitCount       int64   `json:"hit_count"`
	MissCount      int64   `json:"miss_count"`
	HitRate        float64 `json:"hit_rate"` // HitCount / (HitCount + MissCount)
	
	// Memory usage (if available)
	MemoryUsage    int64 `json:"memory_usage"`    // bytes
	KeyCount       int64 `json:"key_count"`       // total keys
	
	// Cache specific metrics
	CompanyCount   int64 `json:"company_count"`   // cached companies
	BrokerageCount int64 `json:"brokerage_count"` // cached brokerages
	
	// Timing
	LastAccess     time.Time `json:"last_access"`
	Uptime         string    `json:"uptime"`
}

// CacheKey represents cache key information
type CacheKey struct {
	Key        string        `json:"key"`
	Type       string        `json:"type"`        // "company", "brokerage"
	Identifier string        `json:"identifier"`  // ticker or name
	TTL        time.Duration `json:"ttl"`
	CreatedAt  time.Time     `json:"created_at"`
}

// CacheConfiguration holds cache configuration options
type CacheConfiguration struct {
	// TTL settings
	DefaultTTL      time.Duration `json:"default_ttl"`
	CompanyTTL      time.Duration `json:"company_ttl"`
	BrokerageTTL    time.Duration `json:"brokerage_ttl"`
	StockRatingTTL  time.Duration `json:"stock_rating_ttl"` // e.g. 1 day for stock ratings
	
	// Key prefixes
	CompanyPrefix   string `json:"company_prefix"`
	BrokeragePrefix string `json:"brokerage_prefix"`
	StockRatingPrefix string `json:"stock_rating_prefix"` // e.g. "stock_rating:"
	
	// Behavior settings
	EnableCompression bool `json:"enable_compression"`
	MaxRetries       int  `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	
	// Fallback behavior
	FailSilently     bool `json:"fail_silently"`     // Don't fail if cache unavailable
	EnableFallback   bool `json:"enable_fallback"`   // Use memory cache if Redis fails
}

// DefaultCacheConfiguration returns default cache configuration
func DefaultCacheConfiguration() CacheConfiguration {
	return CacheConfiguration{
		DefaultTTL:      1 * time.Hour,
		CompanyTTL:      2 * time.Hour,     // Companies change rarely
		BrokerageTTL:    4 * time.Hour,     // Brokerages change very rarely
		StockRatingTTL:  24 * time.Hour,    // Stock ratings are updated daily
		CompanyPrefix:   "company:ticker:",
		BrokeragePrefix: "brokerage:name:",
		StockRatingPrefix: "stock_rating:",
		EnableCompression: false,
		MaxRetries:      3,
		RetryDelay:      100 * time.Millisecond,
		FailSilently:    true,              // Don't break app if cache fails
		EnableFallback:  true,              // Use memory fallback
	}
}

// CacheError represents cache-specific errors
type CacheError struct {
	Operation string // "get", "set", "delete", etc.
	Key       string
	Message   string
	Err       error
}

func (e *CacheError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("cache %s operation failed for key '%s': %s (%v)", 
			e.Operation, e.Key, e.Message, e.Err)
	}
	return fmt.Sprintf("cache %s operation failed for key '%s': %s", 
		e.Operation, e.Key, e.Message)
}

// Utility functions for cache key generation

// GenerateCompanyKey generates a cache key for a company
func GenerateCompanyKey(prefix, ticker string) string {
	return prefix + normalizeKey(ticker)
}

// GenerateBrokerageKey generates a cache key for a brokerage
func GenerateBrokerageKey(prefix, name string) string {
	return prefix + normalizeKey(name)
}

// GenerateStockRatingKey generates a cache key for a stock rating
func GenerateStockRatingKey(prefix string, companyID, brokerageID uuid.UUID) string {
	return fmt.Sprintf("%s%s:%s", prefix, companyID.String(), brokerageID.String())
}

// normalizeKey normalizes cache keys (uppercase, replace spaces with underscores)
func normalizeKey(key string) string {
	normalized := strings.ToUpper(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")
	return normalized
}