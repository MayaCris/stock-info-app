package implementation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// stockRatingRepositoryImpl implements the StockRatingRepository interface using GORM
type stockRatingRepositoryImpl struct {
	db *gorm.DB
}

// NewStockRatingRepository creates a new stock rating repository implementation
func NewStockRatingRepository(db *gorm.DB) interfaces.StockRatingRepository {
	return &stockRatingRepositoryImpl{
		db: db,
	}
}

// NewTransactionalStockRatingRepository creates a new transactional stock rating repository implementation
func NewTransactionalStockRatingRepository(db *gorm.DB) interfaces.TransactionalStockRatingRepository {
	return &stockRatingRepositoryImpl{
		db: db,
	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// Create creates a new stock rating in the database
func (r *stockRatingRepositoryImpl) Create(ctx context.Context, rating *entities.StockRating) error {
	if err := r.db.WithContext(ctx).Create(rating).Error; err != nil {
		// Handle unique constraint violation
		if strings.Contains(err.Error(), "unique_rating_per_company_brokerage_time") {
			return fmt.Errorf("rating already exists for company %s, brokerage %s at time %s",
				rating.CompanyID, rating.BrokerageID, rating.EventTime)
		}
		return fmt.Errorf("failed to create stock rating: %w", err)
	}
	return nil
}

// CreateMany creates multiple stock ratings in a single transaction
func (r *stockRatingRepositoryImpl) CreateMany(ctx context.Context, ratings []*entities.StockRating) error {
	if len(ratings) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rating := range ratings {
			if err := tx.Create(rating).Error; err != nil {
				// Skip duplicates but log them
				if strings.Contains(err.Error(), "unique_rating_per_company_brokerage_time") {
					continue // Skip duplicate, don't fail the entire batch
				}
				return fmt.Errorf("failed to create stock rating in batch: %w", err)
			}
		}
		return nil
	})
}

// ========================================
// READ OPERATIONS - BASIC
// ========================================

// GetByID retrieves a stock rating by its ID
func (r *stockRatingRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.StockRating, error) {
	var rating entities.StockRating

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rating).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("stock rating with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get stock rating by id: %w", err)
	}

	return &rating, nil
}

// GetAll retrieves all stock ratings
func (r *stockRatingRepositoryImpl) GetAll(ctx context.Context) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	err := r.db.WithContext(ctx).Order("event_time DESC").Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all stock ratings: %w", err)
	}

	return ratings, nil
}

// GetByCompanyID retrieves all ratings for a specific company
func (r *stockRatingRepositoryImpl) GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("event_time DESC").
		Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings by company id: %w", err)
	}

	return ratings, nil
}

// GetByBrokerageID retrieves all ratings from a specific brokerage
func (r *stockRatingRepositoryImpl) GetByBrokerageID(ctx context.Context, brokerageID uuid.UUID) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	err := r.db.WithContext(ctx).
		Where("brokerage_id = ?", brokerageID).
		Order("event_time DESC").
		Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings by brokerage id: %w", err)
	}

	return ratings, nil
}

// ========================================
// READ OPERATIONS - ADVANCED FILTERING
// ========================================

// GetByCompanyAndBrokerage retrieves ratings for specific company and brokerage combination
func (r *stockRatingRepositoryImpl) GetByCompanyAndBrokerage(ctx context.Context, companyID, brokerageID uuid.UUID) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	err := r.db.WithContext(ctx).
		Where("company_id = ? AND brokerage_id = ?", companyID, brokerageID).
		Order("event_time DESC").
		Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings by company and brokerage: %w", err)
	}

	return ratings, nil
}

// GetByEventTimeRange retrieves ratings within a specific time range
func (r *stockRatingRepositoryImpl) GetByEventTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	err := r.db.WithContext(ctx).
		Where("event_time >= ? AND event_time <= ?", startTime, endTime).
		Order("event_time DESC").
		Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings by time range: %w", err)
	}

	return ratings, nil
}

// GetByCompanyAndDateRange retrieves ratings for a company within a date range
func (r *stockRatingRepositoryImpl) GetByCompanyAndDateRange(ctx context.Context, companyID uuid.UUID, startTime, endTime time.Time) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	err := r.db.WithContext(ctx).
		Where("company_id = ? AND event_time >= ? AND event_time <= ?", companyID, startTime, endTime).
		Order("event_time DESC").
		Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings by company and date range: %w", err)
	}

	return ratings, nil
}

// GetRecent retrieves recent ratings from the last N days
func (r *stockRatingRepositoryImpl) GetRecent(ctx context.Context, days int, limit int) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	cutoffTime := time.Now().AddDate(0, 0, -days)

	query := r.db.WithContext(ctx).
		Where("event_time >= ?", cutoffTime).
		Order("event_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recent ratings: %w", err)
	}

	return ratings, nil
}

// ========================================
// READ OPERATIONS - BY ACTION TYPE
// ========================================

// GetUpgrades retrieves upgrade ratings
func (r *stockRatingRepositoryImpl) GetUpgrades(ctx context.Context, limit int) ([]*entities.StockRating, error) {
	return r.GetByActionType(ctx, "upgraded by", limit)
}

// GetDowngrades retrieves downgrade ratings
func (r *stockRatingRepositoryImpl) GetDowngrades(ctx context.Context, limit int) ([]*entities.StockRating, error) {
	return r.GetByActionType(ctx, "downgraded by", limit)
}

// GetReiterations retrieves reiteration ratings
func (r *stockRatingRepositoryImpl) GetReiterations(ctx context.Context, limit int) ([]*entities.StockRating, error) {
	return r.GetByActionType(ctx, "reiterated by", limit)
}

// GetByActionType retrieves ratings by action type
func (r *stockRatingRepositoryImpl) GetByActionType(ctx context.Context, actionType string, limit int) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	query := r.db.WithContext(ctx).
		Where("action ILIKE ?", "%"+actionType+"%").
		Order("event_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings by action type: %w", err)
	}

	return ratings, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// Update updates an existing stock rating
func (r *stockRatingRepositoryImpl) Update(ctx context.Context, rating *entities.StockRating) error {
	result := r.db.WithContext(ctx).Save(rating)
	if result.Error != nil {
		return fmt.Errorf("failed to update stock rating: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("stock rating with id %s not found for update", rating.ID)
	}

	return nil
}

// MarkAsProcessed marks a rating as processed
func (r *stockRatingRepositoryImpl) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("id = ?", id).
		Update("is_processed", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark rating as processed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("stock rating with id %s not found for processing", id)
	}

	return nil
}

// MarkAsUnprocessed marks a rating as unprocessed
func (r *stockRatingRepositoryImpl) MarkAsUnprocessed(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("id = ?", id).
		Update("is_processed", false)

	if result.Error != nil {
		return fmt.Errorf("failed to mark rating as unprocessed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("stock rating with id %s not found for unprocessing", id)
	}

	return nil
}

// MarkManyAsProcessed marks multiple ratings as processed
func (r *stockRatingRepositoryImpl) MarkManyAsProcessed(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("id IN ?", ids).
		Update("is_processed", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark ratings as processed: %w", result.Error)
	}

	return nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// Delete performs a soft delete on a stock rating
func (r *stockRatingRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entities.StockRating{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete stock rating: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("stock rating with id %s not found for deletion", id)
	}

	return nil
}

// HardDelete permanently deletes a stock rating from the database
func (r *stockRatingRepositoryImpl) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entities.StockRating{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete stock rating: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("stock rating with id %s not found for hard deletion", id)
	}

	return nil
}

// ========================================
// QUERY OPERATIONS - BASIC STATS
// ========================================

// Count returns the total number of stock ratings
func (r *stockRatingRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.StockRating{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count stock ratings: %w", err)
	}

	return count, nil
}

// CountByCompany returns the number of ratings for a specific company
func (r *stockRatingRepositoryImpl) CountByCompany(ctx context.Context, companyID uuid.UUID) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("company_id = ?", companyID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count ratings by company: %w", err)
	}

	return count, nil
}

// CountByBrokerage returns the number of ratings from a specific brokerage
func (r *stockRatingRepositoryImpl) CountByBrokerage(ctx context.Context, brokerageID uuid.UUID) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("brokerage_id = ?", brokerageID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count ratings by brokerage: %w", err)
	}

	return count, nil
}

// CountByActionType returns the number of ratings by action type
func (r *stockRatingRepositoryImpl) CountByActionType(ctx context.Context, actionType string) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("action ILIKE ?", "%"+actionType+"%").Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count ratings by action type: %w", err)
	}

	return count, nil
}

// ========================================
// BUSINESS OPERATIONS - CRITICAL FOR API SYNC
// ========================================

// FindExisting finds an existing rating with exact match
func (r *stockRatingRepositoryImpl) FindExisting(ctx context.Context, companyID, brokerageID uuid.UUID, eventTime time.Time) (*entities.StockRating, error) {
	var rating entities.StockRating

	err := r.db.WithContext(ctx).
		Where("company_id = ? AND brokerage_id = ? AND event_time = ?", companyID, brokerageID, eventTime).
		First(&rating).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found, but not an error
		}
		return nil, fmt.Errorf("failed to find existing rating: %w", err)
	}

	return &rating, nil
}

// FindOrCreateRating finds or creates a stock rating (CRITICAL for API sync)
func (r *stockRatingRepositoryImpl) FindOrCreateRating(ctx context.Context, companyID, brokerageID uuid.UUID, eventTime time.Time,
	action, ratingFrom, ratingTo, targetFrom, targetTo string, rawData []byte) (*entities.StockRating, error) {

	// First, try to find existing rating
	existing, err := r.FindExisting(ctx, companyID, brokerageID, eventTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing rating: %w", err)
	}

	if existing != nil {
		return existing, nil
	}

	// Create new rating
	newRating := entities.NewStockRating(companyID, brokerageID, action, eventTime)
	newRating.RatingFrom = ratingFrom
	newRating.RatingTo = ratingTo
	newRating.TargetFrom = targetFrom
	newRating.TargetTo = targetTo
	if rawData != nil {
		newRating.RawData = rawData
	}

	if err := r.Create(ctx, newRating); err != nil {
		return nil, fmt.Errorf("failed to create new rating: %w", err)
	}

	return newRating, nil
}

// UpsertMany performs batch upsert operations for stock ratings
func (r *stockRatingRepositoryImpl) UpsertMany(ctx context.Context, ratings []*entities.StockRating) error {
	if len(ratings) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rating := range ratings {
			// Try to find existing rating
			var existing entities.StockRating
			err := tx.Where("company_id = ? AND brokerage_id = ? AND event_time = ?",
				rating.CompanyID, rating.BrokerageID, rating.EventTime).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new rating
				if err := tx.Create(rating).Error; err != nil {
					return fmt.Errorf("failed to create rating in upsert: %w", err)
				}
			} else if err == nil {
				// Update existing rating
				rating.ID = existing.ID // Preserve ID
				if err := tx.Save(rating).Error; err != nil {
					return fmt.Errorf("failed to update rating in upsert: %w", err)
				}
			} else {
				return fmt.Errorf("failed to check existing rating in upsert: %w", err)
			}
		}
		return nil
	})
}

// BulkInsertIgnoreDuplicates inserts ratings ignoring duplicates, returns count inserted
func (r *stockRatingRepositoryImpl) BulkInsertIgnoreDuplicates(ctx context.Context, ratings []*entities.StockRating) (int, error) {
	if len(ratings) == 0 {
		return 0, nil
	}

	insertedCount := 0

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rating := range ratings {
			err := tx.Create(rating).Error
			if err != nil {
				// Skip if duplicate
				if strings.Contains(err.Error(), "unique_rating_per_company_brokerage_time") {
					continue
				}
				return fmt.Errorf("failed to insert rating: %w", err)
			}
			insertedCount++
		}
		return nil
	})

	return insertedCount, err
}

// BulkInsertIgnoreDuplicatesWithTx inserts ratings ignoring duplicates using provided transaction
func (r *stockRatingRepositoryImpl) BulkInsertIgnoreDuplicatesWithTx(ctx context.Context, tx *gorm.DB, ratings []*entities.StockRating) (int, error) {
	if len(ratings) == 0 {
		return 0, nil
	}

	insertedCount := 0

	for _, rating := range ratings {
		// Use raw SQL with ON CONFLICT DO NOTHING to avoid transaction aborts
		query := `
			INSERT INTO stock_ratings (
				id, company_id, brokerage_id, action, rating_from, rating_to, 
				target_from, target_to, event_time, created_at, updated_at, 
				source, is_processed
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?, ?)
			ON CONFLICT (company_id, brokerage_id, event_time) DO NOTHING
		`

		result := tx.WithContext(ctx).Exec(query,
			rating.ID,
			rating.CompanyID,
			rating.BrokerageID,
			rating.Action,
			rating.RatingFrom,
			rating.RatingTo,
			rating.TargetFrom,
			rating.TargetTo,
			rating.EventTime,
			rating.Source,
			rating.IsProcessed,
		)

		if result.Error != nil {
			return insertedCount, fmt.Errorf("failed to insert rating: %w", result.Error)
		}

		// Count rows affected (1 = inserted, 0 = skipped due to conflict)
		insertedCount += int(result.RowsAffected)
	}

	return insertedCount, nil
}

// ========================================
// PROCESSING OPERATIONS (FOR BACKGROUND JOBS)
// ========================================

// GetUnprocessed retrieves unprocessed ratings
func (r *stockRatingRepositoryImpl) GetUnprocessed(ctx context.Context, limit int) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	query := r.db.WithContext(ctx).
		Where("is_processed = ?", false).
		Order("created_at ASC") // Process oldest first

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed ratings: %w", err)
	}

	return ratings, nil
}

// GetUnprocessedBySource retrieves unprocessed ratings from a specific source
func (r *stockRatingRepositoryImpl) GetUnprocessedBySource(ctx context.Context, source string, limit int) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	query := r.db.WithContext(ctx).
		Where("is_processed = ? AND source = ?", false, source).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed ratings by source: %w", err)
	}

	return ratings, nil
}

// GetProcessingBatch retrieves a batch of ratings for processing
func (r *stockRatingRepositoryImpl) GetProcessingBatch(ctx context.Context, batchSize int) ([]*entities.StockRating, error) {
	return r.GetUnprocessed(ctx, batchSize)
}

// ========================================
// RELATIONSHIP OPERATIONS - WITH PRELOADING
// ========================================

// GetWithCompany retrieves a rating with company preloaded
func (r *stockRatingRepositoryImpl) GetWithCompany(ctx context.Context, id uuid.UUID) (*entities.StockRating, error) {
	var rating entities.StockRating

	err := r.db.WithContext(ctx).Preload("Company").Where("id = ?", id).First(&rating).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("stock rating with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get rating with company: %w", err)
	}

	return &rating, nil
}

// GetWithBrokerage retrieves a rating with brokerage preloaded
func (r *stockRatingRepositoryImpl) GetWithBrokerage(ctx context.Context, id uuid.UUID) (*entities.StockRating, error) {
	var rating entities.StockRating

	err := r.db.WithContext(ctx).Preload("Brokerage").Where("id = ?", id).First(&rating).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("stock rating with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get rating with brokerage: %w", err)
	}

	return &rating, nil
}

// GetWithRelations retrieves a rating with both company and brokerage preloaded
func (r *stockRatingRepositoryImpl) GetWithRelations(ctx context.Context, id uuid.UUID) (*entities.StockRating, error) {
	var rating entities.StockRating

	err := r.db.WithContext(ctx).Preload("Company").Preload("Brokerage").Where("id = ?", id).First(&rating).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("stock rating with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get rating with relations: %w", err)
	}

	return &rating, nil
}

// GetAllWithRelations retrieves ratings with relations preloaded
func (r *stockRatingRepositoryImpl) GetAllWithRelations(ctx context.Context, limit int) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	query := r.db.WithContext(ctx).
		Preload("Company").
		Preload("Brokerage").
		Order("event_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings with relations: %w", err)
	}

	return ratings, nil
}

// ========================================
// ANALYTICS OPERATIONS
// ========================================

// GetActionTypeDistribution returns count of each action type in the last N days
func (r *stockRatingRepositoryImpl) GetActionTypeDistribution(ctx context.Context, days int) (map[string]int64, error) {
	var results []struct {
		Action string
		Count  int64
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)

	err := r.db.WithContext(ctx).
		Model(&entities.StockRating{}).
		Select("action, COUNT(*) as count").
		Where("event_time >= ?", cutoffTime).
		Group("action").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get action type distribution: %w", err)
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Action] = result.Count
	}

	return distribution, nil
}

// GetTopCompaniesByRatingCount returns companies with most ratings in last N days
func (r *stockRatingRepositoryImpl) GetTopCompaniesByRatingCount(ctx context.Context, days int, limit int) ([]interfaces.CompanyRatingCount, error) {
	var results []interfaces.CompanyRatingCount

	cutoffTime := time.Now().AddDate(0, 0, -days)

	query := r.db.WithContext(ctx).
		Select("companies.id as company_id, companies.name as company_name, companies.ticker, COUNT(stock_ratings.id) as rating_count").
		Table("stock_ratings").
		Joins("JOIN companies ON stock_ratings.company_id = companies.id").
		Where("stock_ratings.event_time >= ?", cutoffTime).
		Group("companies.id, companies.name, companies.ticker").
		Order("rating_count DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top companies by rating count: %w", err)
	}

	return results, nil
}

// GetTopBrokeragesByRatingCount returns brokerages with most ratings in last N days
func (r *stockRatingRepositoryImpl) GetTopBrokeragesByRatingCount(ctx context.Context, days int, limit int) ([]interfaces.BrokerageRatingCount, error) {
	var results []interfaces.BrokerageRatingCount

	cutoffTime := time.Now().AddDate(0, 0, -days)

	query := r.db.WithContext(ctx).
		Select("brokerages.id as brokerage_id, brokerages.name as brokerage_name, COUNT(stock_ratings.id) as rating_count").
		Table("stock_ratings").
		Joins("JOIN brokerages ON stock_ratings.brokerage_id = brokerages.id").
		Where("stock_ratings.event_time >= ?", cutoffTime).
		Group("brokerages.id, brokerages.name").
		Order("rating_count DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top brokerages by rating count: %w", err)
	}

	return results, nil
}

// GetRatingTrend returns daily rating counts for a company over last N days
func (r *stockRatingRepositoryImpl) GetRatingTrend(ctx context.Context, companyID uuid.UUID, days int) ([]interfaces.DailyRatingCount, error) {
	var results []interfaces.DailyRatingCount

	cutoffTime := time.Now().AddDate(0, 0, -days)

	err := r.db.WithContext(ctx).
		Select(`
			DATE(event_time) as date,
			COUNT(*) as rating_count,
			COUNT(CASE WHEN action ILIKE '%upgrade%' THEN 1 END) as upgrades,
			COUNT(CASE WHEN action ILIKE '%downgrade%' THEN 1 END) as downgrades,
			COUNT(CASE WHEN action ILIKE '%reiterat%' THEN 1 END) as reiterations
		`).
		Model(&entities.StockRating{}).
		Where("company_id = ? AND event_time >= ?", companyID, cutoffTime).
		Group("DATE(event_time)").
		Order("date DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get rating trend: %w", err)
	}

	return results, nil
}

// ========================================
// TIME-BASED QUERIES
// ========================================

// GetTodaysRatings retrieves ratings from today
func (r *stockRatingRepositoryImpl) GetTodaysRatings(ctx context.Context) ([]*entities.StockRating, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return r.GetByEventTimeRange(ctx, startOfDay, endOfDay)
}

// GetThisWeeksRatings retrieves ratings from this week
func (r *stockRatingRepositoryImpl) GetThisWeeksRatings(ctx context.Context) ([]*entities.StockRating, error) {
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	return r.GetByEventTimeRange(ctx, startOfWeek, now)
}

// GetThisMonthsRatings retrieves ratings from this month
func (r *stockRatingRepositoryImpl) GetThisMonthsRatings(ctx context.Context) ([]*entities.StockRating, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	return r.GetByEventTimeRange(ctx, startOfMonth, now)
}

// ========================================
// DUPLICATE DETECTION AND CLEANUP
// ========================================

// FindDuplicates finds groups of duplicate ratings
func (r *stockRatingRepositoryImpl) FindDuplicates(ctx context.Context) ([]interfaces.DuplicateGroup, error) {
	var results []interfaces.DuplicateGroup

	err := r.db.WithContext(ctx).
		Select("company_id, brokerage_id, event_time, COUNT(*) as count").
		Model(&entities.StockRating{}).
		Group("company_id, brokerage_id, event_time").
		Having("COUNT(*) > 1").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find duplicates: %w", err)
	}

	// Get the actual IDs for each duplicate group
	for i := range results {
		var ids []uuid.UUID
		err := r.db.WithContext(ctx).
			Model(&entities.StockRating{}).
			Select("id").
			Where("company_id = ? AND brokerage_id = ? AND event_time = ?",
				results[i].CompanyID, results[i].BrokerageID, results[i].EventTime).
			Pluck("id", &ids).Error

		if err != nil {
			return nil, fmt.Errorf("failed to get duplicate IDs: %w", err)
		}

		results[i].RatingIDs = ids
	}

	return results, nil
}

// RemoveDuplicates removes duplicate ratings, keeping either newest or oldest
func (r *stockRatingRepositoryImpl) RemoveDuplicates(ctx context.Context, keepNewest bool) (int, error) {
	duplicates, err := r.FindDuplicates(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to find duplicates: %w", err)
	}

	removedCount := 0

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, group := range duplicates {
			if len(group.RatingIDs) <= 1 {
				continue
			}

			// Get the detailed ratings to decide which to keep
			var ratings []*entities.StockRating
			err := tx.Where("id IN ?", group.RatingIDs).Find(&ratings).Error
			if err != nil {
				return fmt.Errorf("failed to get duplicate ratings: %w", err)
			}

			// Find which one to keep
			var keepID uuid.UUID
			if keepNewest {
				var newest time.Time
				for _, rating := range ratings {
					if rating.CreatedAt.After(newest) {
						newest = rating.CreatedAt
						keepID = rating.ID
					}
				}
			} else {
				oldest := time.Now()
				for _, rating := range ratings {
					if rating.CreatedAt.Before(oldest) {
						oldest = rating.CreatedAt
						keepID = rating.ID
					}
				}
			}

			// Delete all except the one to keep
			var idsToDelete []uuid.UUID
			for _, id := range group.RatingIDs {
				if id != keepID {
					idsToDelete = append(idsToDelete, id)
				}
			}

			if len(idsToDelete) > 0 {
				result := tx.Where("id IN ?", idsToDelete).Delete(&entities.StockRating{})
				if result.Error != nil {
					return fmt.Errorf("failed to delete duplicates: %w", result.Error)
				}
				removedCount += int(result.RowsAffected)
			}
		}
		return nil
	})

	return removedCount, err
}

// ========================================
// DATA QUALITY OPERATIONS
// ========================================

// GetRatingsWithMissingData retrieves ratings that have missing required data
func (r *stockRatingRepositoryImpl) GetRatingsWithMissingData(ctx context.Context) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	err := r.db.WithContext(ctx).
		Where("action = '' OR action IS NULL OR company_id IS NULL OR brokerage_id IS NULL").
		Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings with missing data: %w", err)
	}

	return ratings, nil
}

// GetRatingsWithInvalidDates retrieves ratings with invalid or future dates
func (r *stockRatingRepositoryImpl) GetRatingsWithInvalidDates(ctx context.Context) ([]*entities.StockRating, error) {
	var ratings []*entities.StockRating

	futureDate := time.Now().Add(24 * time.Hour) // Tomorrow

	err := r.db.WithContext(ctx).
		Where("event_time > ? OR event_time < ?", futureDate, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		Find(&ratings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings with invalid dates: %w", err)
	}

	return ratings, nil
}

// ValidateDataIntegrity performs comprehensive data integrity validation
func (r *stockRatingRepositoryImpl) ValidateDataIntegrity(ctx context.Context) (interfaces.DataIntegrityReport, error) {
	var report interfaces.DataIntegrityReport

	// Get total count
	if count, err := r.Count(ctx); err == nil {
		report.TotalRatings = count
	}

	// Count missing company references
	r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("company_id IS NULL").Count(&report.MissingCompany)

	// Count missing brokerage references
	r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("brokerage_id IS NULL").Count(&report.MissingBrokerage)

	// Count invalid event times
	futureDate := time.Now().Add(24 * time.Hour)
	r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("event_time > ? OR event_time < ?", futureDate, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		Count(&report.InvalidEventTime)

	// Count empty actions
	r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("action = '' OR action IS NULL").Count(&report.EmptyAction)

	// Count duplicates
	duplicates, _ := r.FindDuplicates(ctx)
	report.DuplicateCount = int64(len(duplicates))

	// Count orphaned ratings (company or brokerage not found)
	r.db.WithContext(ctx).
		Select("COUNT(*)").
		Table("stock_ratings").
		Joins("LEFT JOIN companies ON stock_ratings.company_id = companies.id").
		Joins("LEFT JOIN brokerages ON stock_ratings.brokerage_id = brokerages.id").
		Where("companies.id IS NULL OR brokerages.id IS NULL").
		Scan(&report.OrphanedRatings)

	// Count processed vs unprocessed
	r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("is_processed = ?", true).Count(&report.ProcessedRatings)

	r.db.WithContext(ctx).Model(&entities.StockRating{}).
		Where("is_processed = ?", false).Count(&report.UnprocessedRatings)

	return report, nil
}

// GetOrphanedStockRatings efficiently finds orphaned stock ratings using JOINs
func (r *stockRatingRepositoryImpl) GetOrphanedStockRatings(ctx context.Context) ([]*entities.StockRating, error) {
	var orphanedRatings []*entities.StockRating

	// Use LEFT JOINs to find stock ratings where company_id or brokerage_id don't exist
	err := r.db.WithContext(ctx).
		Select("stock_ratings.*").
		Table("stock_ratings").
		Joins("LEFT JOIN companies ON stock_ratings.company_id = companies.id AND companies.deleted_at IS NULL").
		Joins("LEFT JOIN brokerages ON stock_ratings.brokerage_id = brokerages.id AND brokerages.deleted_at IS NULL").
		Where("companies.id IS NULL OR brokerages.id IS NULL").
		Where("stock_ratings.deleted_at IS NULL").
		Find(&orphanedRatings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get orphaned stock ratings: %w", err)
	}

	return orphanedRatings, nil
}

// GetOrphanedStockRatingsWithReasons efficiently finds orphaned stock ratings with specific reasons
func (r *stockRatingRepositoryImpl) GetOrphanedStockRatingsWithReasons(ctx context.Context) ([]interfaces.OrphanedRatingResult, error) {
	var results []interfaces.OrphanedRatingResult

	// Query to get orphaned ratings with reason details
	query := `
		SELECT 
			sr.id,
			sr.company_id,
			sr.brokerage_id,
			sr.event_time,
			sr.action,
			CASE 
				WHEN c.id IS NULL AND b.id IS NULL THEN 'Both company and brokerage not found'
				WHEN c.id IS NULL THEN 'Company not found'
				WHEN b.id IS NULL THEN 'Brokerage not found'
			END as reason
		FROM stock_ratings sr
		LEFT JOIN companies c ON sr.company_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN brokerages b ON sr.brokerage_id = b.id AND b.deleted_at IS NULL
		WHERE (c.id IS NULL OR b.id IS NULL) AND sr.deleted_at IS NULL
		ORDER BY sr.event_time DESC
	`

	err := r.db.WithContext(ctx).Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get orphaned stock ratings with reasons: %w", err)
	}
	return results, nil
}

// ========================================
// TRANSACTIONAL OPERATIONS
// ========================================

// CreateWithTx creates a new stock rating using the provided transaction
func (r *stockRatingRepositoryImpl) CreateWithTx(ctx context.Context, tx *gorm.DB, rating *entities.StockRating) error {
	if err := tx.WithContext(ctx).Create(rating).Error; err != nil {
		// Handle unique constraint violation
		if strings.Contains(err.Error(), "unique_rating_per_company_brokerage_time") {
			return fmt.Errorf("rating already exists for company %s, brokerage %s at time %s",
				rating.CompanyID, rating.BrokerageID, rating.EventTime)
		}
		return fmt.Errorf("failed to create stock rating with transaction: %w", err)
	}
	return nil
}

// CreateManyWithTx creates multiple stock ratings using the provided transaction
func (r *stockRatingRepositoryImpl) CreateManyWithTx(ctx context.Context, tx *gorm.DB, ratings []*entities.StockRating) error {
	if len(ratings) == 0 {
		return nil
	}

	for _, rating := range ratings {
		if err := r.CreateWithTx(ctx, tx, rating); err != nil {
			return fmt.Errorf("failed to create stock rating in batch: %w", err)
		}
	}
	return nil
}

// GetByIDWithTx retrieves a stock rating by ID using the provided transaction
func (r *stockRatingRepositoryImpl) GetByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.StockRating, error) {
	var rating entities.StockRating

	err := tx.WithContext(ctx).Where("id = ?", id).First(&rating).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("stock rating with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get stock rating by id with transaction: %w", err)
	}

	return &rating, nil
}
