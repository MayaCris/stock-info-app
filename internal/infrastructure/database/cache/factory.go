package cache

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

// fallbackCacheService implements a cache service that uses Redis as primary
// and falls back to memory cache when Redis operations fail
type fallbackCacheService struct {
	primary  services.CacheService // Redis cache
	fallback services.CacheService // Memory cache
	config   services.CacheConfiguration
}

// Company cache operations
func (f *fallbackCacheService) GetCompany(ctx context.Context, ticker string) (*entities.Company, error) {
	if company, err := f.primary.GetCompany(ctx, ticker); err == nil {
		return company, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for GetCompany(%s)", ticker)
	return f.fallback.GetCompany(ctx, ticker)
}

func (f *fallbackCacheService) SetCompany(ctx context.Context, ticker string, company *entities.Company, ttl time.Duration) error {
	if err := f.primary.SetCompany(ctx, ticker, company, ttl); err != nil {
		log.Printf("⚠️  Primary cache failed, using fallback for SetCompany(%s): %v", ticker, err)
		return f.fallback.SetCompany(ctx, ticker, company, ttl)
	}
	return nil
}

func (f *fallbackCacheService) DeleteCompany(ctx context.Context, ticker string) error {
	// Try to delete from both caches to ensure consistency
	primaryErr := f.primary.DeleteCompany(ctx, ticker)
	fallbackErr := f.fallback.DeleteCompany(ctx, ticker)

	// If primary succeeds, ignore fallback errors (data might not exist there)
	if primaryErr == nil {
		return nil
	}

	// If primary fails, log and return fallback result
	log.Printf("⚠️  Primary cache delete failed for company %s: %v", ticker, primaryErr)
	return fallbackErr
}

// Brokerage cache operations
func (f *fallbackCacheService) GetBrokerage(ctx context.Context, name string) (*entities.Brokerage, error) {
	if brokerage, err := f.primary.GetBrokerage(ctx, name); err == nil {
		return brokerage, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for GetBrokerage(%s)", name)
	return f.fallback.GetBrokerage(ctx, name)
}

func (f *fallbackCacheService) SetBrokerage(ctx context.Context, name string, brokerage *entities.Brokerage, ttl time.Duration) error {
	if err := f.primary.SetBrokerage(ctx, name, brokerage, ttl); err != nil {
		log.Printf("⚠️  Primary cache failed, using fallback for SetBrokerage(%s): %v", name, err)
		return f.fallback.SetBrokerage(ctx, name, brokerage, ttl)
	}
	return nil
}

func (f *fallbackCacheService) DeleteBrokerage(ctx context.Context, name string) error {
	// Try to delete from both caches to ensure consistency
	primaryErr := f.primary.DeleteBrokerage(ctx, name)
	fallbackErr := f.fallback.DeleteBrokerage(ctx, name)

	// If primary succeeds, ignore fallback errors (data might not exist there)
	if primaryErr == nil {
		return nil
	}

	// If primary fails, log and return fallback result
	log.Printf("⚠️  Primary cache delete failed for brokerage %s: %v", name, primaryErr)
	return fallbackErr
}

// Stock rating cache operations
func (f *fallbackCacheService) GetStockRating(ctx context.Context, companyID, brokerageID uuid.UUID) (*entities.StockRating, error) {
	if rating, err := f.primary.GetStockRating(ctx, companyID, brokerageID); err == nil {
		return rating, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for GetStockRating(%s, %s)", companyID, brokerageID)
	return f.fallback.GetStockRating(ctx, companyID, brokerageID)
}

func (f *fallbackCacheService) SetStockRating(ctx context.Context, stockRating *entities.StockRating, ttl time.Duration) error {
	if err := f.primary.SetStockRating(ctx, stockRating, ttl); err != nil {
		log.Printf("⚠️  Primary cache failed, using fallback for SetStockRating(%s, %s): %v",
			stockRating.CompanyID, stockRating.BrokerageID, err)
		return f.fallback.SetStockRating(ctx, stockRating, ttl)
	}
	return nil
}

func (f *fallbackCacheService) DeleteStockRating(ctx context.Context, companyID, brokerageID uuid.UUID) error {
	// Try to delete from both caches to ensure consistency
	primaryErr := f.primary.DeleteStockRating(ctx, companyID, brokerageID)
	fallbackErr := f.fallback.DeleteStockRating(ctx, companyID, brokerageID)

	// If primary succeeds, ignore fallback errors (data might not exist there)
	if primaryErr == nil {
		return nil
	}

	// If primary fails, log and return fallback result
	log.Printf("⚠️  Primary cache delete failed for stock rating (%s, %s): %v", companyID, brokerageID, primaryErr)
	return fallbackErr
}

// Bulk operations
func (f *fallbackCacheService) GetCompanies(ctx context.Context, tickers []string) (map[string]*entities.Company, error) {
	if companies, err := f.primary.GetCompanies(ctx, tickers); err == nil {
		return companies, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for GetCompanies")
	return f.fallback.GetCompanies(ctx, tickers)
}

func (f *fallbackCacheService) SetCompanies(ctx context.Context, companies map[string]*entities.Company, ttl time.Duration) error {
	if err := f.primary.SetCompanies(ctx, companies, ttl); err != nil {
		log.Printf("⚠️  Primary cache failed, using fallback for SetCompanies: %v", err)
		return f.fallback.SetCompanies(ctx, companies, ttl)
	}
	return nil
}

func (f *fallbackCacheService) GetBrokerages(ctx context.Context, names []string) (map[string]*entities.Brokerage, error) {
	if brokerages, err := f.primary.GetBrokerages(ctx, names); err == nil {
		return brokerages, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for GetBrokerages")
	return f.fallback.GetBrokerages(ctx, names)
}

func (f *fallbackCacheService) SetBrokerages(ctx context.Context, brokerages map[string]*entities.Brokerage, ttl time.Duration) error {
	if err := f.primary.SetBrokerages(ctx, brokerages, ttl); err != nil {
		log.Printf("⚠️  Primary cache failed, using fallback for SetBrokerages: %v", err)
		return f.fallback.SetBrokerages(ctx, brokerages, ttl)
	}
	return nil
}

// Cache management operations
func (f *fallbackCacheService) Clear(ctx context.Context) error {
	// Try to clear both caches to ensure consistency
	primaryErr := f.primary.Clear(ctx)
	fallbackErr := f.fallback.Clear(ctx)

	// If primary succeeds, ignore fallback errors
	if primaryErr == nil {
		return nil
	}

	// If primary fails, log and return fallback result
	log.Printf("⚠️  Primary cache clear failed: %v", primaryErr)
	return fallbackErr
}

func (f *fallbackCacheService) ClearCompanies(ctx context.Context) error {
	// Try to clear both caches to ensure consistency
	primaryErr := f.primary.ClearCompanies(ctx)
	fallbackErr := f.fallback.ClearCompanies(ctx)

	// If primary succeeds, ignore fallback errors
	if primaryErr == nil {
		return nil
	}

	// If primary fails, log and return fallback result
	log.Printf("⚠️  Primary cache ClearCompanies failed: %v", primaryErr)
	return fallbackErr
}

func (f *fallbackCacheService) ClearBrokerages(ctx context.Context) error {
	// Try to clear both caches to ensure consistency
	primaryErr := f.primary.ClearBrokerages(ctx)
	fallbackErr := f.fallback.ClearBrokerages(ctx)

	// If primary succeeds, ignore fallback errors
	if primaryErr == nil {
		return nil
	}

	// If primary fails, log and return fallback result
	log.Printf("⚠️  Primary cache ClearBrokerages failed: %v", primaryErr)
	return fallbackErr
}

// Cache statistics and health
func (f *fallbackCacheService) GetStats(ctx context.Context) (services.CacheStats, error) {
	if stats, err := f.primary.GetStats(ctx); err == nil {
		return stats, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for GetStats")
	return f.fallback.GetStats(ctx)
}

func (f *fallbackCacheService) Ping(ctx context.Context) error {
	if err := f.primary.Ping(ctx); err != nil {
		log.Printf("⚠️  Primary cache ping failed, using fallback: %v", err)
		return f.fallback.Ping(ctx)
	}
	return nil
}

// Key operations
func (f *fallbackCacheService) Exists(ctx context.Context, key string) (bool, error) {
	if exists, err := f.primary.Exists(ctx, key); err == nil {
		return exists, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for Exists(%s)", key)
	return f.fallback.Exists(ctx, key)
}

func (f *fallbackCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	if ttl, err := f.primary.TTL(ctx, key); err == nil {
		return ttl, nil
	}
	log.Printf("⚠️  Primary cache failed, using fallback for TTL(%s)", key)
	return f.fallback.TTL(ctx, key)
}

func (f *fallbackCacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	primaryErr := f.primary.Expire(ctx, key, ttl)
	fallbackErr := f.fallback.Expire(ctx, key, ttl)
	if primaryErr != nil {
		log.Printf("⚠️  Primary cache Expire failed for key %s: %v", key, primaryErr)
		return fallbackErr
	}
	return fallbackErr
}

// NewCacheService creates a cache service based on configuration
// It attempts to use Redis first, falling back to memory cache if Redis fails
func NewCacheService(cfg *config.Config) services.CacheService {
	// Try Redis first
	if redisService, err := NewRedisCacheService(cfg); err == nil {
		log.Println("✅ Using Redis cache service")
		return redisService
	} else {
		log.Printf("⚠️  Failed to connect to Redis: %v", err)
	}

	// Fallback to memory cache
	log.Println("🔄 Falling back to memory cache service")
	return NewMemoryCacheService()
}

// NewRedisCacheServiceWithFallback creates a Redis cache service with memory fallback
// This wrapper automatically falls back to memory cache for operations if Redis fails
func NewRedisCacheServiceWithFallback(cfg *config.Config) services.CacheService {
	redisService, err := NewRedisCacheService(cfg)
	if err != nil {
		log.Printf("⚠️  Redis unavailable, using memory cache: %v", err)
		return NewMemoryCacheService()
	}

	// Create a fallback wrapper
	return &fallbackCacheService{
		primary:  redisService,
		fallback: NewMemoryCacheService(),
		config:   services.DefaultCacheConfiguration(),
	}
}

// NewMemoryCacheServiceOnly creates a memory-only cache service
// Useful for testing or when Redis is not desired
func NewMemoryCacheServiceOnly() services.CacheService {
	log.Println("🧠 Using memory-only cache service")
	return NewMemoryCacheService()
}

// NewRedisCacheServiceOnly creates a Redis-only cache service
// Returns error if Redis is not available
func NewRedisCacheServiceOnly(cfg *config.Config) (services.CacheService, error) {
	service, err := NewRedisCacheService(cfg)
	if err != nil {
		return nil, err
	}
	log.Println("🔴 Using Redis-only cache service")
	return service, nil
}
