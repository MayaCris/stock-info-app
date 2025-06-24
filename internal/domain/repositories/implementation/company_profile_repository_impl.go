package implementation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// companyProfileRepositoryImpl implements the CompanyProfileRepository interface using GORM
type companyProfileRepositoryImpl struct {
	db *gorm.DB
}

// NewCompanyProfileRepository creates a new company profile repository implementation
func NewCompanyProfileRepository(db *gorm.DB) interfaces.CompanyProfileRepository {
	return &companyProfileRepositoryImpl{
		db: db,
	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// Create creates a new company profile in the database
func (r *companyProfileRepositoryImpl) Create(ctx context.Context, profile *entities.CompanyProfile) error {
	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		return fmt.Errorf("failed to create company profile: %w", err)
	}
	return nil
}

// ========================================
// READ OPERATIONS
// ========================================

// GetByID retrieves company profile by its unique ID
func (r *companyProfileRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.CompanyProfile, error) {
	var profile entities.CompanyProfile
	if err := r.db.WithContext(ctx).First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("company profile not found with id %s", id.String())
		}
		return nil, fmt.Errorf("failed to get company profile by id: %w", err)
	}
	return &profile, nil
}

// GetBySymbol retrieves company profile by stock symbol
func (r *companyProfileRepositoryImpl) GetBySymbol(ctx context.Context, symbol string) (*entities.CompanyProfile, error) {
	var profile entities.CompanyProfile
	if err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("company profile not found for symbol %s", symbol)
		}
		return nil, fmt.Errorf("failed to get company profile by symbol: %w", err)
	}
	return &profile, nil
}

// GetAll retrieves all company profiles with pagination
func (r *companyProfileRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	query := r.db.WithContext(ctx).Order("name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get all company profiles: %w", err)
	}

	return profiles, nil
}

// GetBySector retrieves company profiles by sector
func (r *companyProfileRepositoryImpl) GetBySector(ctx context.Context, sector string) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	if err := r.db.WithContext(ctx).
		Where("sector = ?", sector).
		Order("name ASC").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get company profiles by sector: %w", err)
	}
	return profiles, nil
}

// GetByIndustry retrieves company profiles by industry
func (r *companyProfileRepositoryImpl) GetByIndustry(ctx context.Context, industry string) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	if err := r.db.WithContext(ctx).
		Where("industry = ?", industry).
		Order("name ASC").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get company profiles by industry: %w", err)
	}
	return profiles, nil
}

// GetByCountry retrieves company profiles by country
func (r *companyProfileRepositoryImpl) GetByCountry(ctx context.Context, country string) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	if err := r.db.WithContext(ctx).
		Where("country = ?", country).
		Order("name ASC").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get company profiles by country: %w", err)
	}
	return profiles, nil
}

// GetByMarketCapRange retrieves company profiles within a market cap range
func (r *companyProfileRepositoryImpl) GetByMarketCapRange(ctx context.Context, minCap, maxCap int64) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	if err := r.db.WithContext(ctx).
		Where("market_cap BETWEEN ? AND ?", minCap, maxCap).
		Order("market_cap DESC").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get company profiles by market cap range: %w", err)
	}
	return profiles, nil
}

// GetLargeCapStocks retrieves large cap stocks (market cap > $10B)
func (r *companyProfileRepositoryImpl) GetLargeCapStocks(ctx context.Context, limit int) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	query := r.db.WithContext(ctx).
		Where("market_cap > ?", 10000000000). // $10B
		Order("market_cap DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get large cap stocks: %w", err)
	}
	return profiles, nil
}

// GetSmallCapStocks retrieves small cap stocks (market cap < $2B)
func (r *companyProfileRepositoryImpl) GetSmallCapStocks(ctx context.Context, limit int) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	query := r.db.WithContext(ctx).
		Where("market_cap < ?", 2000000000). // $2B
		Order("market_cap DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get small cap stocks: %w", err)
	}
	return profiles, nil
}

// GetStaleProfiles retrieves profiles older than maxAge
func (r *companyProfileRepositoryImpl) GetStaleProfiles(ctx context.Context, maxAge time.Duration) ([]*entities.CompanyProfile, error) {
	staleTime := time.Now().Add(-maxAge)
	var profiles []*entities.CompanyProfile
	if err := r.db.WithContext(ctx).
		Where("updated_at < ?", staleTime).
		Order("updated_at ASC").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get stale company profiles: %w", err)
	}
	return profiles, nil
}

// GetRecentlyUpdated retrieves profiles updated since the specified time
func (r *companyProfileRepositoryImpl) GetRecentlyUpdated(ctx context.Context, since time.Time) ([]*entities.CompanyProfile, error) {
	var profiles []*entities.CompanyProfile
	if err := r.db.WithContext(ctx).
		Where("updated_at >= ?", since).
		Order("updated_at DESC").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get recently updated company profiles: %w", err)
	}
	return profiles, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// Update updates an existing company profile
func (r *companyProfileRepositoryImpl) Update(ctx context.Context, profile *entities.CompanyProfile) error {
	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		return fmt.Errorf("failed to update company profile: %w", err)
	}
	return nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// Delete removes company profile by ID
func (r *companyProfileRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.CompanyProfile{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete company profile: %w", err)
	}
	return nil
}

// ========================================
// BULK OPERATIONS
// ========================================

// BulkCreate creates multiple company profiles
func (r *companyProfileRepositoryImpl) BulkCreate(ctx context.Context, profiles []*entities.CompanyProfile) error {
	if len(profiles) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(profiles, 100).Error; err != nil {
		return fmt.Errorf("failed to create company profiles in bulk: %w", err)
	}
	return nil
}

// BulkUpdate updates multiple company profiles
func (r *companyProfileRepositoryImpl) BulkUpdate(ctx context.Context, profiles []*entities.CompanyProfile) error {
	if len(profiles) == 0 {
		return nil
	}

	for i := 0; i < len(profiles); i += 100 {
		end := i + 100
		if end > len(profiles) {
			end = len(profiles)
		}

		batch := profiles[i:end]
		for _, profile := range batch {
			if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
				return fmt.Errorf("failed to update company profile in bulk: %w", err)
			}
		}
	}

	return nil
}

// UpsertBySymbol creates or updates company profile for a symbol
func (r *companyProfileRepositoryImpl) UpsertBySymbol(ctx context.Context, profile *entities.CompanyProfile) error {
	var existing entities.CompanyProfile
	err := r.db.WithContext(ctx).
		Where("symbol = ?", profile.Symbol).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record
		if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
			return fmt.Errorf("failed to create company profile during upsert: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing company profile during upsert: %w", err)
	} else {
		// Update existing record
		profile.ID = existing.ID
		profile.CreatedAt = existing.CreatedAt
		if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
			return fmt.Errorf("failed to update company profile during upsert: %w", err)
		}
	}

	return nil
}

// ========================================
// STATISTICS OPERATIONS
// ========================================

// Count returns the total number of company profiles
func (r *companyProfileRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.CompanyProfile{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count company profiles: %w", err)
	}
	return count, nil
}

// GetSectorDistribution returns the distribution of companies by sector
func (r *companyProfileRepositoryImpl) GetSectorDistribution(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Sector string
		Count  int64
	}

	if err := r.db.WithContext(ctx).
		Model(&entities.CompanyProfile{}).
		Select("sector, COUNT(*) as count").
		Group("sector").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get sector distribution: %w", err)
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Sector] = result.Count
	}

	return distribution, nil
}

// GetCountryDistribution returns the distribution of companies by country
func (r *companyProfileRepositoryImpl) GetCountryDistribution(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Country string
		Count   int64
	}

	if err := r.db.WithContext(ctx).
		Model(&entities.CompanyProfile{}).
		Select("country, COUNT(*) as count").
		Group("country").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get country distribution: %w", err)
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Country] = result.Count
	}

	return distribution, nil
}

// ========================================
// HEALTH CHECK OPERATIONS
// ========================================

// Health performs a health check on the repository
func (r *companyProfileRepositoryImpl) Health(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.CompanyProfile{}).Limit(1).Count(&count).Error; err != nil {
		return fmt.Errorf("company profile repository health check failed: %w", err)
	}
	return nil
}
