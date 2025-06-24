package implementation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// companyRepositoryImpl implements the CompanyRepository interface using GORM
type companyRepositoryImpl struct {
	db *gorm.DB
}

// NewCompanyRepository creates a new company repository implementation
func NewCompanyRepository(db *gorm.DB) interfaces.CompanyRepository {
	return &companyRepositoryImpl{
		db: db,
	}
}

// NewTransactionalCompanyRepository creates a new transactional company repository implementation
func NewTransactionalCompanyRepository(db *gorm.DB) interfaces.TransactionalCompanyRepository {
	return &companyRepositoryImpl{
		db: db,
	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// Create creates a new company in the database
func (r *companyRepositoryImpl) Create(ctx context.Context, company *entities.Company) error {
	if err := r.db.WithContext(ctx).Create(company).Error; err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}
	return nil
}

// CreateMany creates multiple companies in a single transaction
func (r *companyRepositoryImpl) CreateMany(ctx context.Context, companies []*entities.Company) error {
	if len(companies) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, company := range companies {
			if err := tx.Create(company).Error; err != nil {
				return fmt.Errorf("failed to create company %s: %w", company.Ticker, err)
			}
		}
		return nil
	})
}

// ========================================
// READ OPERATIONS
// ========================================

// GetByID retrieves a company by its ID
func (r *companyRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Company, error) {
	var company entities.Company

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("company with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get company by id: %w", err)
	}

	return &company, nil
}

// GetByTicker retrieves a company by its ticker symbol (CRITICAL for API sync)
func (r *companyRepositoryImpl) GetByTicker(ctx context.Context, ticker string) (*entities.Company, error) {
	var company entities.Company

	err := r.db.WithContext(ctx).Where("ticker = ?", strings.ToUpper(ticker)).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("company with ticker %s not found", ticker)
		}
		return nil, fmt.Errorf("failed to get company by ticker: %w", err)
	}

	return &company, nil
}

// GetByName retrieves a company by its name
func (r *companyRepositoryImpl) GetByName(ctx context.Context, name string) (*entities.Company, error) {
	var company entities.Company

	err := r.db.WithContext(ctx).Where("name = ?", name).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("company with name %s not found", name)
		}
		return nil, fmt.Errorf("failed to get company by name: %w", err)
	}

	return &company, nil
}

// GetAll retrieves all companies (including inactive)
func (r *companyRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Company, error) {
	var companies []*entities.Company

	err := r.db.WithContext(ctx).Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all companies: %w", err)
	}

	return companies, nil
}

// GetAllActive retrieves only active companies
func (r *companyRepositoryImpl) GetAllActive(ctx context.Context) ([]*entities.Company, error) {
	var companies []*entities.Company

	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active companies: %w", err)
	}

	return companies, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// Update updates an existing company
func (r *companyRepositoryImpl) Update(ctx context.Context, company *entities.Company) error {
	result := r.db.WithContext(ctx).Save(company)
	if result.Error != nil {
		return fmt.Errorf("failed to update company: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("company with id %s not found for update", company.ID)
	}

	return nil
}

// UpdateMarketCap updates only the market cap of a company by ticker
func (r *companyRepositoryImpl) UpdateMarketCap(ctx context.Context, ticker string, marketCap float64) error {
	result := r.db.WithContext(ctx).Model(&entities.Company{}).
		Where("ticker = ?", strings.ToUpper(ticker)).
		Update("market_cap", marketCap)

	if result.Error != nil {
		return fmt.Errorf("failed to update market cap for %s: %w", ticker, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("company with ticker %s not found for market cap update", ticker)
	}

	return nil
}

// Activate activates a company by ID
func (r *companyRepositoryImpl) Activate(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entities.Company{}).Where("id = ?", id).Update("is_active", true)
	if result.Error != nil {
		return fmt.Errorf("failed to activate company: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("company with id %s not found for activation", id)
	}

	return nil
}

// Deactivate deactivates a company by ID
func (r *companyRepositoryImpl) Deactivate(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entities.Company{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to deactivate company: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("company with id %s not found for deactivation", id)
	}

	return nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// Delete performs a soft delete on a company
func (r *companyRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entities.Company{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete company: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("company with id %s not found for deletion", id)
	}

	return nil
}

// HardDelete permanently deletes a company from the database
func (r *companyRepositoryImpl) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entities.Company{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete company: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("company with id %s not found for hard deletion", id)
	}

	return nil
}

// ========================================
// QUERY OPERATIONS - BASIC
// ========================================

// ExistsByTicker checks if a company with the given ticker exists
func (r *companyRepositoryImpl) ExistsByTicker(ctx context.Context, ticker string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.Company{}).Where("ticker = ?", strings.ToUpper(ticker)).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if company exists by ticker: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if a company with the given name exists
func (r *companyRepositoryImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.Company{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if company exists by name: %w", err)
	}

	return count > 0, nil
}

// Count returns the total number of companies (including inactive)
func (r *companyRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.Company{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count companies: %w", err)
	}

	return count, nil
}

// CountActive returns the number of active companies
func (r *companyRepositoryImpl) CountActive(ctx context.Context) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.Company{}).Where("is_active = ?", true).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count active companies: %w", err)
	}

	return count, nil
}

// ========================================
// QUERY OPERATIONS - FINANCIAL
// ========================================

// GetBySector retrieves companies by sector
func (r *companyRepositoryImpl) GetBySector(ctx context.Context, sector string) ([]*entities.Company, error) {
	var companies []*entities.Company

	err := r.db.WithContext(ctx).Where("sector = ? AND is_active = ?", sector, true).Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get companies by sector: %w", err)
	}

	return companies, nil
}

// GetByExchange retrieves companies by exchange
func (r *companyRepositoryImpl) GetByExchange(ctx context.Context, exchange string) ([]*entities.Company, error) {
	var companies []*entities.Company

	err := r.db.WithContext(ctx).Where("exchange = ? AND is_active = ?", strings.ToUpper(exchange), true).Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get companies by exchange: %w", err)
	}

	return companies, nil
}

// GetByMarketCapRange retrieves companies within a market cap range
func (r *companyRepositoryImpl) GetByMarketCapRange(ctx context.Context, minCap, maxCap float64) ([]*entities.Company, error) {
	var companies []*entities.Company

	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if minCap > 0 {
		query = query.Where("market_cap >= ?", minCap)
	}
	if maxCap > 0 {
		query = query.Where("market_cap <= ?", maxCap)
	}

	err := query.Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get companies by market cap range: %w", err)
	}

	return companies, nil
}

// GetLargestByMarketCap retrieves companies ordered by market cap (largest first)
func (r *companyRepositoryImpl) GetLargestByMarketCap(ctx context.Context, limit int) ([]*entities.Company, error) {
	var companies []*entities.Company

	query := r.db.WithContext(ctx).
		Where("is_active = ? AND market_cap > 0", true).
		Order("market_cap DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get largest companies by market cap: %w", err)
	}

	return companies, nil
}

// ========================================
// BUSINESS OPERATIONS (CRITICAL FOR API SYNC)
// ========================================

// FindOrCreateByTicker finds a company by ticker or creates it if it doesn't exist
func (r *companyRepositoryImpl) FindOrCreateByTicker(ctx context.Context, ticker, name string) (*entities.Company, error) {
	// First, try to find existing company
	company, err := r.GetByTicker(ctx, ticker)
	if err == nil {
		return company, nil
	}

	// If not found, create a new one
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		(err != nil && strings.Contains(fmt.Sprintf("%v", err), "not found")) {

		newCompany := entities.NewCompany(ticker, name)
		if err := r.Create(ctx, newCompany); err != nil {
			return nil, fmt.Errorf("failed to create new company %s: %w", ticker, err)
		}

		return newCompany, nil
	}

	// If there was a different error, return it
	return nil, fmt.Errorf("failed to find or create company: %w", err)
}

// FindOrCreateWithDetails finds or creates a company with additional details
func (r *companyRepositoryImpl) FindOrCreateWithDetails(ctx context.Context, ticker, name, sector, exchange string, marketCap float64) (*entities.Company, error) {
	// First, try to find existing company
	company, err := r.GetByTicker(ctx, ticker)
	if err == nil {
		// Update details if they're empty and we have new data
		updated := false
		if company.Sector == "" && sector != "" {
			company.Sector = sector
			updated = true
		}
		if company.Exchange == "" && exchange != "" {
			company.Exchange = exchange
			updated = true
		}
		if company.MarketCap == 0 && marketCap > 0 {
			company.MarketCap = marketCap
			updated = true
		}

		// Save updates if any
		if updated {
			if err := r.Update(ctx, company); err != nil {
				return nil, fmt.Errorf("failed to update company details: %w", err)
			}
		}

		return company, nil
	}

	// If not found, create a new one with details
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		(err != nil && strings.Contains(fmt.Sprintf("%v", err), "not found")) {

		newCompany := entities.NewCompanyWithDetails(ticker, name, sector, exchange, marketCap)
		if err := r.Create(ctx, newCompany); err != nil {
			return nil, fmt.Errorf("failed to create new company with details %s: %w", ticker, err)
		}

		return newCompany, nil
	}

	// If there was a different error, return it
	return nil, fmt.Errorf("failed to find or create company with details: %w", err)
}

// UpsertMany performs batch upsert operations for companies
func (r *companyRepositoryImpl) UpsertMany(ctx context.Context, companies []*entities.Company) error {
	if len(companies) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, company := range companies {
			// Try to find existing company by ticker
			var existing entities.Company
			err := tx.Where("ticker = ?", company.Ticker).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new company
				if err := tx.Create(company).Error; err != nil {
					return fmt.Errorf("failed to create company %s in batch: %w", company.Ticker, err)
				}
			} else if err == nil {
				// Update existing company
				company.ID = existing.ID // Preserve ID
				if err := tx.Save(company).Error; err != nil {
					return fmt.Errorf("failed to update company %s in batch: %w", company.Ticker, err)
				}
			} else {
				return fmt.Errorf("failed to check existing company %s: %w", company.Ticker, err)
			}
		}
		return nil
	})
}

// ========================================
// RELATIONSHIP OPERATIONS
// ========================================

// GetWithRatings retrieves a company with its stock ratings preloaded
func (r *companyRepositoryImpl) GetWithRatings(ctx context.Context, id uuid.UUID) (*entities.Company, error) {
	var company entities.Company

	err := r.db.WithContext(ctx).Preload("StockRatings").Where("id = ?", id).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("company with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get company with ratings: %w", err)
	}

	return &company, nil
}

// GetByRatingCount retrieves companies ordered by their rating count (most active first)
func (r *companyRepositoryImpl) GetByRatingCount(ctx context.Context, limit int) ([]*entities.Company, error) {
	var companies []*entities.Company

	query := r.db.WithContext(ctx).
		Select("companies.*, COUNT(stock_ratings.id) as rating_count").
		Joins("LEFT JOIN stock_ratings ON companies.id = stock_ratings.company_id").
		Where("companies.is_active = ?", true).
		Group("companies.id").
		Order("rating_count DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get companies by rating count: %w", err)
	}

	return companies, nil
}

// GetMostActiveCompanies retrieves companies with most ratings in the last N days
func (r *companyRepositoryImpl) GetMostActiveCompanies(ctx context.Context, days int, limit int) ([]*entities.Company, error) {
	var companies []*entities.Company

	query := r.db.WithContext(ctx).
		Select("companies.*, COUNT(stock_ratings.id) as recent_rating_count").
		Joins("LEFT JOIN stock_ratings ON companies.id = stock_ratings.company_id").
		Where("companies.is_active = ? AND stock_ratings.event_time >= NOW() - INTERVAL ? DAY", true, days).
		Group("companies.id").
		Order("recent_rating_count DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get most active companies: %w", err)
	}

	return companies, nil
}

// ========================================
// SEARCH OPERATIONS
// ========================================

// SearchByName searches companies by name using partial matching
func (r *companyRepositoryImpl) SearchByName(ctx context.Context, query string, limit int) ([]*entities.Company, error) {
	var companies []*entities.Company

	searchQuery := r.db.WithContext(ctx).
		Where("name ILIKE ? AND is_active = ?", "%"+query+"%", true).
		Order("name ASC")

	if limit > 0 {
		searchQuery = searchQuery.Limit(limit)
	}

	err := searchQuery.Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search companies by name: %w", err)
	}

	return companies, nil
}

// SearchByTicker searches companies by ticker using partial matching
func (r *companyRepositoryImpl) SearchByTicker(ctx context.Context, query string, limit int) ([]*entities.Company, error) {
	var companies []*entities.Company

	searchQuery := r.db.WithContext(ctx).
		Where("ticker ILIKE ? AND is_active = ?", "%"+strings.ToUpper(query)+"%", true).
		Order("ticker ASC")

	if limit > 0 {
		searchQuery = searchQuery.Limit(limit)
	}

	err := searchQuery.Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search companies by ticker: %w", err)
	}

	return companies, nil
}

// ========================================
// ANALYTICS OPERATIONS
// ========================================

// GetSectorDistribution returns the count of companies per sector
func (r *companyRepositoryImpl) GetSectorDistribution(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Sector string
		Count  int64
	}

	err := r.db.WithContext(ctx).
		Model(&entities.Company{}).
		Select("sector, COUNT(*) as count").
		Where("is_active = ? AND sector IS NOT NULL AND sector != ''", true).
		Group("sector").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get sector distribution: %w", err)
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Sector] = result.Count
	}

	return distribution, nil
}

// GetExchangeDistribution returns the count of companies per exchange
func (r *companyRepositoryImpl) GetExchangeDistribution(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Exchange string
		Count    int64
	}

	err := r.db.WithContext(ctx).
		Model(&entities.Company{}).
		Select("exchange, COUNT(*) as count").
		Where("is_active = ? AND exchange IS NOT NULL AND exchange != ''", true).
		Group("exchange").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get exchange distribution: %w", err)
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Exchange] = result.Count
	}

	return distribution, nil
}

// GetMarketCapStats returns market cap statistics (min, max, avg)
func (r *companyRepositoryImpl) GetMarketCapStats(ctx context.Context) (map[string]float64, error) {
	var result struct {
		MinCap float64
		MaxCap float64
		AvgCap float64
		Count  int64
	}

	err := r.db.WithContext(ctx).
		Model(&entities.Company{}).
		Select("MIN(market_cap) as min_cap, MAX(market_cap) as max_cap, AVG(market_cap) as avg_cap, COUNT(*) as count").
		Where("is_active = ? AND market_cap > 0", true).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get market cap stats: %w", err)
	}

	stats := map[string]float64{
		"min":   result.MinCap,
		"max":   result.MaxCap,
		"avg":   result.AvgCap,
		"count": float64(result.Count),
	}

	return stats, nil
}

// ========================================
// TRANSACTIONAL OPERATIONS
// ========================================

// CreateWithTx creates a new company using the provided transaction
func (r *companyRepositoryImpl) CreateWithTx(ctx context.Context, tx *gorm.DB, company *entities.Company) error {
	if err := tx.WithContext(ctx).Create(company).Error; err != nil {
		return fmt.Errorf("failed to create company with transaction: %w", err)
	}
	return nil
}

// CreateManyWithTx creates multiple companies using the provided transaction
func (r *companyRepositoryImpl) CreateManyWithTx(ctx context.Context, tx *gorm.DB, companies []*entities.Company) error {
	if len(companies) == 0 {
		return nil
	}

	for _, company := range companies {
		if err := tx.WithContext(ctx).Create(company).Error; err != nil {
			return fmt.Errorf("failed to create company %s with transaction: %w", company.Ticker, err)
		}
	}
	return nil
}

// GetByTickerWithTx retrieves a company by ticker using the provided transaction
func (r *companyRepositoryImpl) GetByTickerWithTx(ctx context.Context, tx *gorm.DB, ticker string) (*entities.Company, error) {
	var company entities.Company

	err := tx.WithContext(ctx).Where("ticker = ?", strings.ToUpper(ticker)).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("company with ticker %s not found", ticker)
		}
		return nil, fmt.Errorf("failed to get company by ticker with transaction: %w", err)
	}

	return &company, nil
}

// FindOrCreateByTickerWithTx finds or creates a company by ticker using the provided transaction
func (r *companyRepositoryImpl) FindOrCreateByTickerWithTx(ctx context.Context, tx *gorm.DB, ticker, name string) (*entities.Company, error) {
	// Try to find existing company
	company, err := r.GetByTickerWithTx(ctx, tx, ticker)
	if err == nil {
		return company, nil // Found existing
	}

	// Create new company if not found
	newCompany := entities.NewCompany(ticker, name)
	if err := r.CreateWithTx(ctx, tx, newCompany); err != nil {
		return nil, fmt.Errorf("failed to create company %s: %w", ticker, err)
	}

	return newCompany, nil
}

// CreateIgnoreDuplicatesWithTx creates a company ignoring duplicates without aborting transaction
func (r *companyRepositoryImpl) CreateIgnoreDuplicatesWithTx(ctx context.Context, tx *gorm.DB, company *entities.Company) (*entities.Company, error) {
	// Use raw SQL with ON CONFLICT DO NOTHING to avoid transaction aborts
	query := `
		INSERT INTO companies (id, ticker, name, sector, exchange, market_cap, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (ticker) DO NOTHING
		RETURNING id, ticker, name, sector, exchange, market_cap, is_active, created_at, updated_at
	`

	var result entities.Company
	err := tx.WithContext(ctx).Raw(query,
		company.ID,
		company.Ticker,
		company.Name,
		company.Sector,
		company.Exchange,
		company.MarketCap,
		company.IsActive,
	).Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to create company with duplicate handling: %w", err)
	}

	// If no rows were returned (conflict occurred), fetch the existing company
	if result.ID == uuid.Nil {
		existing, err := r.GetByTickerWithTx(ctx, tx, company.Ticker)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing company after conflict: %w", err)
		}
		return existing, nil
	}

	return &result, nil
}
