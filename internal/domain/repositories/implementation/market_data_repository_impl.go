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

// marketDataRepositoryImpl implements the MarketDataRepository interface using GORM
type marketDataRepositoryImpl struct {
	db *gorm.DB
}

// NewMarketDataRepository creates a new market data repository implementation
func NewMarketDataRepository(db *gorm.DB) interfaces.MarketDataRepository {
	return &marketDataRepositoryImpl{
		db: db,
	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// Create creates a new market data record in the database
func (r *marketDataRepositoryImpl) Create(ctx context.Context, marketData *entities.MarketData) error {
	if err := r.db.WithContext(ctx).Create(marketData).Error; err != nil {
		return fmt.Errorf("failed to create market data: %w", err)
	}
	return nil
}

// CreateMany creates multiple market data records in a single transaction
func (r *marketDataRepositoryImpl) CreateMany(ctx context.Context, marketDataList []*entities.MarketData) error {
	if len(marketDataList) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(marketDataList, 100).Error; err != nil {
		return fmt.Errorf("failed to create market data in batch: %w", err)
	}
	return nil
}

// BulkCreate creates multiple market data records (alias for CreateMany)
func (r *marketDataRepositoryImpl) BulkCreate(ctx context.Context, marketData []*entities.MarketData) error {
	return r.CreateMany(ctx, marketData)
}

// ========================================
// READ OPERATIONS
// ========================================

// GetByID retrieves market data by its unique ID
func (r *marketDataRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.MarketData, error) {
	var marketData entities.MarketData
	if err := r.db.WithContext(ctx).First(&marketData, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("market data not found with id %s", id.String())
		}
		return nil, fmt.Errorf("failed to get market data by id: %w", err)
	}
	return &marketData, nil
}

// GetBySymbol retrieves the latest market data for a stock symbol
func (r *marketDataRepositoryImpl) GetBySymbol(ctx context.Context, symbol string) (*entities.MarketData, error) {
	var marketData entities.MarketData
	if err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("market_market_timestamp DESC").
		First(&marketData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("market data not found for symbol %s", symbol)
		}
		return nil, fmt.Errorf("failed to get market data by symbol: %w", err)
	}
	return &marketData, nil
}

// GetBySymbolAndTimeRange retrieves market data for a symbol within a time range
func (r *marketDataRepositoryImpl) GetBySymbolAndTimeRange(ctx context.Context, symbol string, startTime, endTime time.Time) ([]*entities.MarketData, error) {
	var marketDataList []*entities.MarketData
	if err := r.db.WithContext(ctx).
		Where("symbol = ? AND market_timestamp BETWEEN ? AND ?", symbol, startTime, endTime).
		Order("market_timestamp DESC").
		Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get market data by symbol and time range: %w", err)
	}
	return marketDataList, nil
}

// GetLatestForMultipleSymbols retrieves the latest market data for multiple symbols
func (r *marketDataRepositoryImpl) GetLatestForMultipleSymbols(ctx context.Context, symbols []string) ([]*entities.MarketData, error) {
	if len(symbols) == 0 {
		return []*entities.MarketData{}, nil
	}

	var marketDataList []*entities.MarketData

	// Using a subquery to get the latest record for each symbol
	subQuery := r.db.Select("symbol, MAX(market_timestamp) as max_market_timestamp").
		Where("symbol IN ?", symbols).
		Group("symbol")

	if err := r.db.WithContext(ctx).
		Table("market_data").
		Joins("JOIN (?) as latest ON market_data.symbol = latest.symbol AND market_data.market_timestamp = latest.max_market_timestamp", subQuery).
		Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest market data for multiple symbols: %w", err)
	}

	return marketDataList, nil
}

// GetAll retrieves all market data with pagination
func (r *marketDataRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.MarketData, error) {
	var marketDataList []*entities.MarketData
	query := r.db.WithContext(ctx).Order("market_timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get all market data: %w", err)
	}

	return marketDataList, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// Update updates an existing market data record
func (r *marketDataRepositoryImpl) Update(ctx context.Context, marketData *entities.MarketData) error {
	if err := r.db.WithContext(ctx).Save(marketData).Error; err != nil {
		return fmt.Errorf("failed to update market data: %w", err)
	}
	return nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// Delete removes market data by ID
func (r *marketDataRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.MarketData{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete market data: %w", err)
	}
	return nil
}

// DeleteBySymbol removes all market data for a specific symbol
func (r *marketDataRepositoryImpl) DeleteBySymbol(ctx context.Context, symbol string) error {
	if err := r.db.WithContext(ctx).Where("symbol = ?", symbol).Delete(&entities.MarketData{}).Error; err != nil {
		return fmt.Errorf("failed to delete market data by symbol: %w", err)
	}
	return nil
}

// DeleteOldRecords removes market data older than the specified time
func (r *marketDataRepositoryImpl) DeleteOldRecords(ctx context.Context, olderThan time.Time) error {
	if err := r.db.WithContext(ctx).Where("market_timestamp < ?", olderThan).Delete(&entities.MarketData{}).Error; err != nil {
		return fmt.Errorf("failed to delete old market data records: %w", err)
	}
	return nil
}

// ========================================
// COUNT OPERATIONS
// ========================================

// Count returns the total number of market data records
func (r *marketDataRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.MarketData{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count market data records: %w", err)
	}
	return count, nil
}

// CountBySymbol returns the number of market data records for a specific symbol
func (r *marketDataRepositoryImpl) CountBySymbol(ctx context.Context, symbol string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.MarketData{}).Where("symbol = ?", symbol).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count market data records by symbol: %w", err)
	}
	return count, nil
}

// ========================================
// EXISTS OPERATIONS
// ========================================

// Exists checks if market data exists by ID
func (r *marketDataRepositoryImpl) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.MarketData{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check if market data exists: %w", err)
	}
	return count > 0, nil
}

// ExistsBySymbol checks if market data exists for a specific symbol
func (r *marketDataRepositoryImpl) ExistsBySymbol(ctx context.Context, symbol string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.MarketData{}).Where("symbol = ?", symbol).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check if market data exists by symbol: %w", err)
	}
	return count > 0, nil
}

// ========================================
// ADDITIONAL READ OPERATIONS
// ========================================

// GetByCompanyID retrieves the latest market data for a company
func (r *marketDataRepositoryImpl) GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*entities.MarketData, error) {
	var marketData entities.MarketData
	if err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("market_timestamp DESC").
		First(&marketData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("market data not found for company id %s", companyID.String())
		}
		return nil, fmt.Errorf("failed to get market data by company id: %w", err)
	}
	return &marketData, nil
}

// GetByCompanyIDs retrieves the latest market data for multiple companies
func (r *marketDataRepositoryImpl) GetByCompanyIDs(ctx context.Context, companyIDs []uuid.UUID) ([]*entities.MarketData, error) {
	if len(companyIDs) == 0 {
		return []*entities.MarketData{}, nil
	}

	var marketDataList []*entities.MarketData

	// Using a subquery to get the latest record for each company
	subQuery := r.db.Select("company_id, MAX(market_timestamp) as max_market_timestamp").
		Where("company_id IN ?", companyIDs).
		Group("company_id")

	if err := r.db.WithContext(ctx).
		Table("market_data").
		Joins("JOIN (?) as latest ON market_data.company_id = latest.company_id AND market_data.market_timestamp = latest.max_market_timestamp", subQuery).
		Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest market data for multiple companies: %w", err)
	}

	return marketDataList, nil
}

// GetLatest retrieves the most recent market data records
func (r *marketDataRepositoryImpl) GetLatest(ctx context.Context, limit int) ([]*entities.MarketData, error) {
	var marketDataList []*entities.MarketData
	query := r.db.WithContext(ctx).Order("market_timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest market data: %w", err)
	}

	return marketDataList, nil
}

// GetByTimeRange retrieves market data within a time range
func (r *marketDataRepositoryImpl) GetByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*entities.MarketData, error) {
	var marketDataList []*entities.MarketData
	if err := r.db.WithContext(ctx).
		Where("market_timestamp BETWEEN ? AND ?", startTime, endTime).
		Order("market_timestamp DESC").
		Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get market data by time range: %w", err)
	}
	return marketDataList, nil
}

// GetStaleData retrieves market data older than maxAge
func (r *marketDataRepositoryImpl) GetStaleData(ctx context.Context, maxAge time.Duration) ([]*entities.MarketData, error) {
	staleTime := time.Now().Add(-maxAge)
	var marketDataList []*entities.MarketData
	if err := r.db.WithContext(ctx).
		Where("market_timestamp < ?", staleTime).
		Order("market_timestamp DESC").
		Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get stale market data: %w", err)
	}
	return marketDataList, nil
}

// ========================================
// MARKET ANALYSIS OPERATIONS
// ========================================

// GetTopGainers retrieves stocks with highest percentage gains
func (r *marketDataRepositoryImpl) GetTopGainers(ctx context.Context, limit int) ([]*entities.MarketData, error) {
	var marketDataList []*entities.MarketData

	// Get latest records and order by percentage change descending
	subQuery := r.db.Select("symbol, MAX(market_timestamp) as max_market_timestamp").
		Group("symbol")

	query := r.db.WithContext(ctx).
		Table("market_data").
		Joins("JOIN (?) as latest ON market_data.symbol = latest.symbol AND market_data.market_timestamp = latest.max_market_timestamp", subQuery).
		Where("price_change_perc > 0").
		Order("price_change_perc DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get top gainers: %w", err)
	}

	return marketDataList, nil
}

// GetTopLosers retrieves stocks with highest percentage losses
func (r *marketDataRepositoryImpl) GetTopLosers(ctx context.Context, limit int) ([]*entities.MarketData, error) {
	var marketDataList []*entities.MarketData

	// Get latest records and order by percentage change ascending
	subQuery := r.db.Select("symbol, MAX(market_timestamp) as max_market_timestamp").
		Group("symbol")

	query := r.db.WithContext(ctx).
		Table("market_data").
		Joins("JOIN (?) as latest ON market_data.symbol = latest.symbol AND market_data.market_timestamp = latest.max_market_timestamp", subQuery).
		Where("price_change_perc < 0").
		Order("price_change_perc ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get top losers: %w", err)
	}

	return marketDataList, nil
}

// GetMostActive retrieves stocks with highest trading volume
func (r *marketDataRepositoryImpl) GetMostActive(ctx context.Context, limit int) ([]*entities.MarketData, error) {
	var marketDataList []*entities.MarketData

	// Get latest records and order by volume descending
	subQuery := r.db.Select("symbol, MAX(market_timestamp) as max_market_timestamp").
		Group("symbol")

	query := r.db.WithContext(ctx).
		Table("market_data").
		Joins("JOIN (?) as latest ON market_data.symbol = latest.symbol AND market_data.market_timestamp = latest.max_market_timestamp", subQuery).
		Order("volume DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&marketDataList).Error; err != nil {
		return nil, fmt.Errorf("failed to get most active stocks: %w", err)
	}

	return marketDataList, nil
}

// ========================================
// BULK UPDATE OPERATIONS
// ========================================

// BulkUpdate updates multiple market data records
func (r *marketDataRepositoryImpl) BulkUpdate(ctx context.Context, marketData []*entities.MarketData) error {
	if len(marketData) == 0 {
		return nil
	}

	// Update in batches for better performance
	for i := 0; i < len(marketData); i += 100 {
		end := i + 100
		if end > len(marketData) {
			end = len(marketData)
		}

		batch := marketData[i:end]
		for _, md := range batch {
			if err := r.db.WithContext(ctx).Save(md).Error; err != nil {
				return fmt.Errorf("failed to update market data in bulk: %w", err)
			}
		}
	}

	return nil
}

// UpsertBySymbol creates or updates market data for a symbol
func (r *marketDataRepositoryImpl) UpsertBySymbol(ctx context.Context, marketData *entities.MarketData) error { // Check if record exists for the symbol and market_timestamp
	var existing entities.MarketData
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND market_timestamp = ?", marketData.Symbol, marketData.MarketTimestamp).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record
		if err := r.db.WithContext(ctx).Create(marketData).Error; err != nil {
			return fmt.Errorf("failed to create market data during upsert: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing market data during upsert: %w", err)
	} else {
		// Update existing record
		marketData.ID = existing.ID
		if err := r.db.WithContext(ctx).Save(marketData).Error; err != nil {
			return fmt.Errorf("failed to update market data during upsert: %w", err)
		}
	}

	return nil
}

// ========================================
// DATA MANAGEMENT OPERATIONS
// ========================================

// CleanupOldData removes market data older than the specified time and returns the count of deleted records
func (r *marketDataRepositoryImpl) CleanupOldData(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("market_timestamp < ?", olderThan).Delete(&entities.MarketData{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old market data: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ========================================
// HEALTH CHECK OPERATIONS
// ========================================

// Health performs a health check on the repository
func (r *marketDataRepositoryImpl) Health(ctx context.Context) error {
	// Simple query to test database connectivity
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.MarketData{}).Limit(1).Count(&count).Error; err != nil {
		return fmt.Errorf("market data repository health check failed: %w", err)
	}
	return nil
}
