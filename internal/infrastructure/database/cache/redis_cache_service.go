package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/redis/go-redis/v9"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

// redisCacheService implements CacheService using Redis
type redisCacheService struct {
	client *redis.Client
	config services.CacheConfiguration
	stats  *cacheStats
}

// cacheStats tracks cache performance metrics
type cacheStats struct {
	hitCount   int64
	missCount  int64
	lastAccess time.Time
	startTime  time.Time
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(cfg *config.Config) (services.CacheService, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Cache.GetRedisAddr(),
		Password:     cfg.Cache.Password,
		DB:           cfg.Cache.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Redis cache connected successfully to %s", cfg.Cache.GetRedisAddr())

	return &redisCacheService{
		client: client,
		config: services.DefaultCacheConfiguration(),
		stats: &cacheStats{
			startTime: time.Now(),
		},
	}, nil
}

// ========================================
// COMPANY CACHE OPERATIONS
// ========================================

// GetCompany retrieves a company from cache
func (r *redisCacheService) GetCompany(ctx context.Context, ticker string) (*entities.Company, error) {
	r.stats.lastAccess = time.Now()

	key := services.GenerateCompanyKey(r.config.CompanyPrefix, ticker)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.stats.missCount++
			return nil, nil // Cache miss, not an error
		}
		r.stats.missCount++
		return nil, r.wrapError("get", key, "Redis get failed", err)
	}

	var company entities.Company
	if err := json.Unmarshal([]byte(data), &company); err != nil {
		r.stats.missCount++
		// Delete corrupted data
		r.client.Del(ctx, key)
		return nil, r.wrapError("get", key, "JSON unmarshal failed", err)
	}

	r.stats.hitCount++
	return &company, nil
}

// SetCompany stores a company in cache
func (r *redisCacheService) SetCompany(ctx context.Context, ticker string, company *entities.Company, ttl time.Duration) error {
	key := services.GenerateCompanyKey(r.config.CompanyPrefix, ticker)

	data, err := json.Marshal(company)
	if err != nil {
		return r.wrapError("set", key, "JSON marshal failed", err)
	}

	if ttl == 0 {
		ttl = r.config.CompanyTTL
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return r.wrapError("set", key, "Redis set failed", err)
	}

	return nil
}

// DeleteCompany removes a company from cache
func (r *redisCacheService) DeleteCompany(ctx context.Context, ticker string) error {
	key := services.GenerateCompanyKey(r.config.CompanyPrefix, ticker)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return r.wrapError("delete", key, "Redis delete failed", err)
	}

	return nil
}

// ========================================
// BROKERAGE CACHE OPERATIONS
// ========================================

// GetBrokerage retrieves a brokerage from cache
func (r *redisCacheService) GetBrokerage(ctx context.Context, name string) (*entities.Brokerage, error) {
	r.stats.lastAccess = time.Now()

	key := services.GenerateBrokerageKey(r.config.BrokeragePrefix, name)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.stats.missCount++
			return nil, nil // Cache miss, not an error
		}
		r.stats.missCount++
		return nil, r.wrapError("get", key, "Redis get failed", err)
	}

	var brokerage entities.Brokerage
	if err := json.Unmarshal([]byte(data), &brokerage); err != nil {
		r.stats.missCount++
		// Delete corrupted data
		r.client.Del(ctx, key)
		return nil, r.wrapError("get", key, "JSON unmarshal failed", err)
	}

	r.stats.hitCount++
	return &brokerage, nil
}

// SetBrokerage stores a brokerage in cache
func (r *redisCacheService) SetBrokerage(ctx context.Context, name string, brokerage *entities.Brokerage, ttl time.Duration) error {
	key := services.GenerateBrokerageKey(r.config.BrokeragePrefix, name)

	data, err := json.Marshal(brokerage)
	if err != nil {
		return r.wrapError("set", key, "JSON marshal failed", err)
	}

	if ttl == 0 {
		ttl = r.config.BrokerageTTL
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return r.wrapError("set", key, "Redis set failed", err)
	}

	return nil
}

// DeleteBrokerage removes a brokerage from cache
func (r *redisCacheService) DeleteBrokerage(ctx context.Context, name string) error {
	key := services.GenerateBrokerageKey(r.config.BrokeragePrefix, name)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return r.wrapError("delete", key, "Redis delete failed", err)
	}

	return nil
}

// Stock Rating cache operations
// GetStockRating retrieves a stock rating from cache

func (r *redisCacheService) GetStockRating(ctx context.Context, companyID, brokerageID uuid.UUID) (*entities.StockRating, error) {
	r.stats.lastAccess = time.Now()

	key := services.GenerateStockRatingKey(r.config.StockRatingPrefix, companyID, brokerageID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.stats.missCount++
			return nil, nil // Cache miss, not an error
		}
		r.stats.missCount++
		return nil, r.wrapError("get", key, "Redis get failed", err)
	}

	var stockRating entities.StockRating
	if err := json.Unmarshal([]byte(data), &stockRating); err != nil {
		r.stats.missCount++
		// Delete corrupted data
		r.client.Del(ctx, key)
		return nil, r.wrapError("get", key, "JSON unmarshal failed", err)
	}

	r.stats.hitCount++
	return &stockRating, nil
}

// SetStockRating stores a stock rating in cache
func (r *redisCacheService) SetStockRating(ctx context.Context, stockRating *entities.StockRating, ttl time.Duration) error {
	key := services.GenerateStockRatingKey(r.config.StockRatingPrefix, stockRating.CompanyID, stockRating.BrokerageID)

	data, err := json.Marshal(stockRating)
	if err != nil {
		return r.wrapError("set", key, "JSON marshal failed", err)
	}

	if ttl == 0 {
		ttl = r.config.StockRatingTTL
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return r.wrapError("set", key, "Redis set failed", err)
	}

	return nil
}

// DeleteStockRating removes a stock rating from cache
func (r *redisCacheService) DeleteStockRating(ctx context.Context, companyID, brokerageID uuid.UUID) error {
	key := services.GenerateStockRatingKey(r.config.StockRatingPrefix, companyID, brokerageID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return r.wrapError("delete", key, "Redis delete failed", err)
	}

	log.Printf("🗑️  Deleted stock rating from cache: %s", key)
	return nil
}

// ========================================
// BULK OPERATIONS
// ========================================

// GetCompanies retrieves multiple companies from cache
func (r *redisCacheService) GetCompanies(ctx context.Context, tickers []string) (map[string]*entities.Company, error) {
	if len(tickers) == 0 {
		return make(map[string]*entities.Company), nil
	}

	r.stats.lastAccess = time.Now()

	// Generate keys for all tickers
	keys := make([]string, len(tickers))
	keyToTicker := make(map[string]string)

	for i, ticker := range tickers {
		key := services.GenerateCompanyKey(r.config.CompanyPrefix, ticker)
		keys[i] = key
		keyToTicker[key] = ticker
	}

	// Execute MGET command
	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		r.stats.missCount += int64(len(tickers))
		return nil, r.wrapError("mget", fmt.Sprintf("%d keys", len(keys)), "Redis MGET failed", err)
	}

	// Process results
	companies := make(map[string]*entities.Company)

	for i, result := range results {
		key := keys[i]
		ticker := keyToTicker[key]

		if result == nil {
			r.stats.missCount++
			continue // Cache miss for this ticker
		}

		data, ok := result.(string)
		if !ok {
			r.stats.missCount++
			continue
		}

		var company entities.Company
		if err := json.Unmarshal([]byte(data), &company); err != nil {
			r.stats.missCount++
			// Delete corrupted data
			r.client.Del(ctx, key)
			continue
		}

		companies[ticker] = &company
		r.stats.hitCount++
	}

	return companies, nil
}

// SetCompanies stores multiple companies in cache
func (r *redisCacheService) SetCompanies(ctx context.Context, companies map[string]*entities.Company, ttl time.Duration) error {
	if len(companies) == 0 {
		return nil
	}

	if ttl == 0 {
		ttl = r.config.CompanyTTL
	}

	// Use pipeline for efficiency
	pipe := r.client.Pipeline()

	for ticker, company := range companies {
		key := services.GenerateCompanyKey(r.config.CompanyPrefix, ticker)

		data, err := json.Marshal(company)
		if err != nil {
			return r.wrapError("mset", key, "JSON marshal failed", err)
		}

		pipe.Set(ctx, key, data, ttl)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return r.wrapError("mset", fmt.Sprintf("%d companies", len(companies)), "Redis pipeline failed", err)
	}

	return nil
}

// GetBrokerages retrieves multiple brokerages from cache
func (r *redisCacheService) GetBrokerages(ctx context.Context, names []string) (map[string]*entities.Brokerage, error) {
	if len(names) == 0 {
		return make(map[string]*entities.Brokerage), nil
	}

	r.stats.lastAccess = time.Now()

	// Generate keys for all names
	keys := make([]string, len(names))
	keyToName := make(map[string]string)

	for i, name := range names {
		key := services.GenerateBrokerageKey(r.config.BrokeragePrefix, name)
		keys[i] = key
		keyToName[key] = name
	}

	// Execute MGET command
	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		r.stats.missCount += int64(len(names))
		return nil, r.wrapError("mget", fmt.Sprintf("%d keys", len(keys)), "Redis MGET failed", err)
	}

	// Process results
	brokerages := make(map[string]*entities.Brokerage)

	for i, result := range results {
		key := keys[i]
		name := keyToName[key]

		if result == nil {
			r.stats.missCount++
			continue // Cache miss for this name
		}

		data, ok := result.(string)
		if !ok {
			r.stats.missCount++
			continue
		}

		var brokerage entities.Brokerage
		if err := json.Unmarshal([]byte(data), &brokerage); err != nil {
			r.stats.missCount++
			// Delete corrupted data
			r.client.Del(ctx, key)
			continue
		}

		brokerages[name] = &brokerage
		r.stats.hitCount++
	}

	return brokerages, nil
}

// SetBrokerages stores multiple brokerages in cache
func (r *redisCacheService) SetBrokerages(ctx context.Context, brokerages map[string]*entities.Brokerage, ttl time.Duration) error {
	if len(brokerages) == 0 {
		return nil
	}

	if ttl == 0 {
		ttl = r.config.BrokerageTTL
	}

	// Use pipeline for efficiency
	pipe := r.client.Pipeline()

	for name, brokerage := range brokerages {
		key := services.GenerateBrokerageKey(r.config.BrokeragePrefix, name)

		data, err := json.Marshal(brokerage)
		if err != nil {
			return r.wrapError("mset", key, "JSON marshal failed", err)
		}

		pipe.Set(ctx, key, data, ttl)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return r.wrapError("mset", fmt.Sprintf("%d brokerages", len(brokerages)), "Redis pipeline failed", err)
	}

	return nil
}

// ========================================
// CACHE MANAGEMENT OPERATIONS
// ========================================

// Clear removes all cache entries
func (r *redisCacheService) Clear(ctx context.Context) error {
	// Get all keys with our prefixes
	companyKeys, err := r.client.Keys(ctx, r.config.CompanyPrefix+"*").Result()
	if err != nil {
		return r.wrapError("clear", "company keys", "Failed to get company keys", err)
	}

	brokerageKeys, err := r.client.Keys(ctx, r.config.BrokeragePrefix+"*").Result()
	if err != nil {
		return r.wrapError("clear", "brokerage keys", "Failed to get brokerage keys", err)
	}

	allKeys := append(companyKeys, brokerageKeys...)

	if len(allKeys) > 0 {
		if err := r.client.Del(ctx, allKeys...).Err(); err != nil {
			return r.wrapError("clear", fmt.Sprintf("%d keys", len(allKeys)), "Redis delete failed", err)
		}
	}

	return nil
}

// ClearCompanies removes all company cache entries
func (r *redisCacheService) ClearCompanies(ctx context.Context) error {
	keys, err := r.client.Keys(ctx, r.config.CompanyPrefix+"*").Result()
	if err != nil {
		return r.wrapError("clear_companies", "company keys", "Failed to get company keys", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return r.wrapError("clear_companies", fmt.Sprintf("%d keys", len(keys)), "Redis delete failed", err)
		}
	}

	return nil
}

// ClearBrokerages removes all brokerage cache entries
func (r *redisCacheService) ClearBrokerages(ctx context.Context) error {
	keys, err := r.client.Keys(ctx, r.config.BrokeragePrefix+"*").Result()
	if err != nil {
		return r.wrapError("clear_brokerages", "brokerage keys", "Failed to get brokerage keys", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return r.wrapError("clear_brokerages", fmt.Sprintf("%d keys", len(keys)), "Redis delete failed", err)
		}
	}

	return nil
}

// ========================================
// CACHE STATISTICS AND HEALTH
// ========================================

// GetStats returns cache statistics
func (r *redisCacheService) GetStats(ctx context.Context) (services.CacheStats, error) {
	// Get Redis memory info
	memoryInfo, err := r.client.Info(ctx, "memory").Result()
	var memoryUsage int64 = 0
	if err != nil {
		log.Printf("⚠️  Failed to get Redis memory info: %v", err)
	} else {
		// Parse memory usage from info if needed
		// For now, we'll skip parsing and use 0
		_ = memoryInfo
	}

	// Count our keys
	companyKeys, _ := r.client.Keys(ctx, r.config.CompanyPrefix+"*").Result()
	brokerageKeys, _ := r.client.Keys(ctx, r.config.BrokeragePrefix+"*").Result()

	hitRate := float64(0)
	total := r.stats.hitCount + r.stats.missCount
	if total > 0 {
		hitRate = float64(r.stats.hitCount) / float64(total) * 100
	}
	return services.CacheStats{
		IsConnected:    r.client.Ping(ctx).Err() == nil,
		Backend:        "redis",
		HitCount:       r.stats.hitCount,
		MissCount:      r.stats.missCount,
		HitRate:        hitRate,
		MemoryUsage:    memoryUsage,
		KeyCount:       int64(len(companyKeys) + len(brokerageKeys)),
		CompanyCount:   int64(len(companyKeys)),
		BrokerageCount: int64(len(brokerageKeys)),
		LastAccess:     r.stats.lastAccess,
		Uptime:         time.Since(r.stats.startTime).String(),
	}, nil
}

// Ping tests the cache connection
func (r *redisCacheService) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return r.wrapError("ping", "redis", "Ping failed", err)
	}
	return nil
}

// ========================================
// KEY OPERATIONS
// ========================================

// Exists checks if a key exists in cache
func (r *redisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, r.wrapError("exists", key, "Redis exists failed", err)
	}
	return count > 0, nil
}

// TTL returns the remaining time to live for a key
func (r *redisCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, r.wrapError("ttl", key, "Redis TTL failed", err)
	}
	return ttl, nil
}

// Expire sets a timeout on a key
func (r *redisCacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return r.wrapError("expire", key, "Redis expire failed", err)
	}
	return nil
}

// ========================================
// HELPER METHODS
// ========================================

// wrapError creates a standardized cache error
func (r *redisCacheService) wrapError(operation, key, message string, err error) error {
	// In production, you might want to log these errors
	if !r.config.FailSilently {
		log.Printf("❌ Cache %s error for key '%s': %s (%v)", operation, key, message, err)
	}

	return &services.CacheError{
		Operation: operation,
		Key:       key,
		Message:   message,
		Err:       err,
	}
}

// Close closes the Redis connection
func (r *redisCacheService) Close() error {
	return r.client.Close()
}
