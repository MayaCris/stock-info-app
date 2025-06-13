package cache

import (
	"context"
	"sync"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/google/uuid"
)

// memoryCacheService implements CacheService using in-memory storage
// This is used as a fallback when Redis is unavailable or for testing
type memoryCacheService struct {
	companies    map[string]*cacheItem
	brokerages   map[string]*cacheItem
	stockRatings map[string]*cacheItem
	config       services.CacheConfiguration
	stats        *cacheStats
	mutex        sync.RWMutex
}

// cacheItem represents an item in the memory cache with expiration
type cacheItem struct {
	data      interface{}
	expiresAt time.Time
}

// NewMemoryCacheService creates a new memory cache service
func NewMemoryCacheService() services.CacheService {
	service := &memoryCacheService{
		companies:    make(map[string]*cacheItem),
		brokerages:   make(map[string]*cacheItem),
		stockRatings: make(map[string]*cacheItem),
		config:       services.DefaultCacheConfiguration(),
		stats: &cacheStats{
			startTime: time.Now(),
		},
	}

	// Start cleanup goroutine to remove expired items
	go service.startCleanupRoutine()

	return service
}

// startCleanupRoutine removes expired items periodically
func (m *memoryCacheService) startCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute) // Cleanup every 10 minutes
	defer ticker.Stop()

	for range ticker.C {
		m.cleanup()
	}
}

// cleanup removes expired items from memory
func (m *memoryCacheService) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()

	// Cleanup companies
	for key, item := range m.companies {
		if now.After(item.expiresAt) {
			delete(m.companies, key)
		}
	}
	// Cleanup brokerages
	for key, item := range m.brokerages {
		if now.After(item.expiresAt) {
			delete(m.brokerages, key)
		}
	}

	// Cleanup stock ratings
	for key, item := range m.stockRatings {
		if now.After(item.expiresAt) {
			delete(m.stockRatings, key)
		}
	}
}

// isExpired checks if a cache item has expired
func (item *cacheItem) isExpired() bool {
	return time.Now().After(item.expiresAt)
}

// ========================================
// COMPANY CACHE OPERATIONS
// ========================================

// GetCompany retrieves a company from memory cache
func (m *memoryCacheService) GetCompany(ctx context.Context, ticker string) (*entities.Company, error) {
	m.stats.lastAccess = time.Now()

	key := services.GenerateCompanyKey(m.config.CompanyPrefix, ticker)

	m.mutex.RLock()
	item, exists := m.companies[key]
	m.mutex.RUnlock()

	if !exists || item.isExpired() {
		if exists && item.isExpired() {
			// Remove expired item
			m.mutex.Lock()
			delete(m.companies, key)
			m.mutex.Unlock()
		}
		m.stats.missCount++
		return nil, nil // Cache miss
	}

	company, ok := item.data.(*entities.Company)
	if !ok {
		m.mutex.Lock()
		delete(m.companies, key) // Remove corrupted data
		m.mutex.Unlock()
		m.stats.missCount++
		return nil, nil
	}

	m.stats.hitCount++
	return company, nil
}

// SetCompany stores a company in memory cache
func (m *memoryCacheService) SetCompany(ctx context.Context, ticker string, company *entities.Company, ttl time.Duration) error {
	key := services.GenerateCompanyKey(m.config.CompanyPrefix, ticker)

	if ttl == 0 {
		ttl = m.config.CompanyTTL
	}

	// Create a copy of the company to avoid external modifications
	companyCopy := *company

	item := &cacheItem{
		data:      &companyCopy,
		expiresAt: time.Now().Add(ttl),
	}

	m.mutex.Lock()
	m.companies[key] = item
	m.mutex.Unlock()

	return nil
}

// DeleteCompany removes a company from memory cache
func (m *memoryCacheService) DeleteCompany(ctx context.Context, ticker string) error {
	key := services.GenerateCompanyKey(m.config.CompanyPrefix, ticker)

	m.mutex.Lock()
	delete(m.companies, key)
	m.mutex.Unlock()

	return nil
}

// ========================================
// BROKERAGE CACHE OPERATIONS
// ========================================

// GetBrokerage retrieves a brokerage from memory cache
func (m *memoryCacheService) GetBrokerage(ctx context.Context, name string) (*entities.Brokerage, error) {
	m.stats.lastAccess = time.Now()

	key := services.GenerateBrokerageKey(m.config.BrokeragePrefix, name)

	m.mutex.RLock()
	item, exists := m.brokerages[key]
	m.mutex.RUnlock()

	if !exists || item.isExpired() {
		if exists && item.isExpired() {
			// Remove expired item
			m.mutex.Lock()
			delete(m.brokerages, key)
			m.mutex.Unlock()
		}
		m.stats.missCount++
		return nil, nil // Cache miss
	}

	brokerage, ok := item.data.(*entities.Brokerage)
	if !ok {
		m.mutex.Lock()
		delete(m.brokerages, key) // Remove corrupted data
		m.mutex.Unlock()
		m.stats.missCount++
		return nil, nil
	}

	m.stats.hitCount++
	return brokerage, nil
}

// SetBrokerage stores a brokerage in memory cache
func (m *memoryCacheService) SetBrokerage(ctx context.Context, name string, brokerage *entities.Brokerage, ttl time.Duration) error {
	key := services.GenerateBrokerageKey(m.config.BrokeragePrefix, name)

	if ttl == 0 {
		ttl = m.config.BrokerageTTL
	}

	// Create a copy of the brokerage to avoid external modifications
	brokerageCopy := *brokerage

	item := &cacheItem{
		data:      &brokerageCopy,
		expiresAt: time.Now().Add(ttl),
	}

	m.mutex.Lock()
	m.brokerages[key] = item
	m.mutex.Unlock()

	return nil
}

// DeleteBrokerage removes a brokerage from memory cache
func (m *memoryCacheService) DeleteBrokerage(ctx context.Context, name string) error {
	key := services.GenerateBrokerageKey(m.config.BrokeragePrefix, name)

	m.mutex.Lock()
	delete(m.brokerages, key)
	m.mutex.Unlock()

	return nil
}

// ========================================
// BULK OPERATIONS
// ========================================

// GetCompanies retrieves multiple companies from memory cache
func (m *memoryCacheService) GetCompanies(ctx context.Context, tickers []string) (map[string]*entities.Company, error) {
	if len(tickers) == 0 {
		return make(map[string]*entities.Company), nil
	}

	m.stats.lastAccess = time.Now()
	companies := make(map[string]*entities.Company)

	for _, ticker := range tickers {
		company, _ := m.GetCompany(ctx, ticker)
		if company != nil {
			companies[ticker] = company
		}
	}

	return companies, nil
}

// SetCompanies stores multiple companies in memory cache
func (m *memoryCacheService) SetCompanies(ctx context.Context, companies map[string]*entities.Company, ttl time.Duration) error {
	for ticker, company := range companies {
		if err := m.SetCompany(ctx, ticker, company, ttl); err != nil {
			return err
		}
	}
	return nil
}

// GetBrokerages retrieves multiple brokerages from memory cache
func (m *memoryCacheService) GetBrokerages(ctx context.Context, names []string) (map[string]*entities.Brokerage, error) {
	if len(names) == 0 {
		return make(map[string]*entities.Brokerage), nil
	}

	m.stats.lastAccess = time.Now()
	brokerages := make(map[string]*entities.Brokerage)

	for _, name := range names {
		brokerage, _ := m.GetBrokerage(ctx, name)
		if brokerage != nil {
			brokerages[name] = brokerage
		}
	}

	return brokerages, nil
}

// SetBrokerages stores multiple brokerages in memory cache
func (m *memoryCacheService) SetBrokerages(ctx context.Context, brokerages map[string]*entities.Brokerage, ttl time.Duration) error {
	for name, brokerage := range brokerages {
		if err := m.SetBrokerage(ctx, name, brokerage, ttl); err != nil {
			return err
		}
	}
	return nil
}

// ========================================
// CACHE MANAGEMENT OPERATIONS
// ========================================

// Clear removes all cache entries
func (m *memoryCacheService) Clear(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.companies = make(map[string]*cacheItem)
	m.brokerages = make(map[string]*cacheItem)

	return nil
}

// ClearCompanies removes all company cache entries
func (m *memoryCacheService) ClearCompanies(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.companies = make(map[string]*cacheItem)

	return nil
}

// ClearBrokerages removes all brokerage cache entries
func (m *memoryCacheService) ClearBrokerages(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.brokerages = make(map[string]*cacheItem)

	return nil
}

// ========================================
// CACHE STATISTICS AND HEALTH
// ========================================

// GetStats returns cache statistics
func (m *memoryCacheService) GetStats(ctx context.Context) (services.CacheStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	hitRate := float64(0)
	total := m.stats.hitCount + m.stats.missCount
	if total > 0 {
		hitRate = float64(m.stats.hitCount) / float64(total) * 100
	}

	// Count non-expired items
	activeCompanies := int64(0)
	activeBrokerages := int64(0)
	now := time.Now()

	for _, item := range m.companies {
		if !now.After(item.expiresAt) {
			activeCompanies++
		}
	}

	for _, item := range m.brokerages {
		if !now.After(item.expiresAt) {
			activeBrokerages++
		}
	}

	return services.CacheStats{
		IsConnected:    true,
		Backend:        "memory",
		HitCount:       m.stats.hitCount,
		MissCount:      m.stats.missCount,
		HitRate:        hitRate,
		KeyCount:       activeCompanies + activeBrokerages,
		CompanyCount:   activeCompanies,
		BrokerageCount: activeBrokerages,
		LastAccess:     m.stats.lastAccess,
		Uptime:         time.Since(m.stats.startTime).String(),
	}, nil
}

// Ping tests the cache connection (always succeeds for memory cache)
func (m *memoryCacheService) Ping(ctx context.Context) error {
	return nil
}

// ========================================
// KEY OPERATIONS
// ========================================

// Exists checks if a key exists in cache
func (m *memoryCacheService) Exists(ctx context.Context, key string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check in companies
	if item, exists := m.companies[key]; exists && !item.isExpired() {
		return true, nil
	}

	// Check in brokerages
	if item, exists := m.brokerages[key]; exists && !item.isExpired() {
		return true, nil
	}

	return false, nil
}

// TTL returns the remaining time to live for a key
func (m *memoryCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	now := time.Now()

	// Check in companies
	if item, exists := m.companies[key]; exists {
		if now.After(item.expiresAt) {
			return -1, nil // Expired
		}
		return item.expiresAt.Sub(now), nil
	}

	// Check in brokerages
	if item, exists := m.brokerages[key]; exists {
		if now.After(item.expiresAt) {
			return -1, nil // Expired
		}
		return item.expiresAt.Sub(now), nil
	}

	return -2, nil // Key doesn't exist
}

// Expire sets a timeout on a key
func (m *memoryCacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	newExpiresAt := time.Now().Add(ttl)

	// Check in companies
	if item, exists := m.companies[key]; exists {
		item.expiresAt = newExpiresAt
		return nil
	}

	// Check in brokerages
	if item, exists := m.brokerages[key]; exists {
		item.expiresAt = newExpiresAt
		return nil
	}

	return &services.CacheError{
		Operation: "expire",
		Key:       key,
		Message:   "key not found",
	}
}

// ========================================
// STOCK RATING OPERATIONS
// ========================================

// GetStockRating retrieves a stock rating from memory cache
func (m *memoryCacheService) GetStockRating(ctx context.Context, companyID, brokerageID uuid.UUID) (*entities.StockRating, error) {
	key := services.GenerateStockRatingKey(m.config.StockRatingPrefix, companyID, brokerageID)

	m.mutex.RLock()
	item, exists := m.stockRatings[key]
	m.mutex.RUnlock()
	if !exists {
		m.stats.missCount++
		return nil, &services.CacheError{
			Operation: "get",
			Key:       key,
			Message:   "stock rating not found",
		}
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		m.mutex.Lock()
		delete(m.stockRatings, key)
		m.mutex.Unlock()
		m.stats.missCount++
		return nil, &services.CacheError{
			Operation: "get",
			Key:       key,
			Message:   "stock rating expired",
		}
	}

	stockRating, ok := item.data.(*entities.StockRating)
	if !ok {
		m.stats.missCount++
		return nil, &services.CacheError{
			Operation: "get",
			Key:       key,
			Message:   "invalid stock rating data type",
		}
	}

	m.stats.hitCount++
	m.stats.lastAccess = time.Now()
	return stockRating, nil
}

// SetStockRating stores a stock rating in memory cache
func (m *memoryCacheService) SetStockRating(ctx context.Context, stockRating *entities.StockRating, ttl time.Duration) error {
	key := services.GenerateStockRatingKey(m.config.StockRatingPrefix, stockRating.CompanyID, stockRating.BrokerageID)

	if ttl == 0 {
		ttl = m.config.StockRatingTTL
	}

	item := &cacheItem{
		data:      stockRating,
		expiresAt: time.Now().Add(ttl),
	}

	m.mutex.Lock()
	m.stockRatings[key] = item
	m.mutex.Unlock()

	return nil
}

// DeleteStockRating removes a stock rating from memory cache
func (m *memoryCacheService) DeleteStockRating(ctx context.Context, companyID, brokerageID uuid.UUID) error {
	key := services.GenerateStockRatingKey(m.config.StockRatingPrefix, companyID, brokerageID)

	m.mutex.Lock()
	delete(m.stockRatings, key)
	m.mutex.Unlock()

	return nil
}

// ========================================
