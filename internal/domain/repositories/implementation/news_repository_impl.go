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

// newsRepositoryImpl implements the NewsRepository interface using GORM
type newsRepositoryImpl struct {
	db *gorm.DB
}

// NewNewsRepository creates a new news repository implementation
func NewNewsRepository(db *gorm.DB) interfaces.NewsRepository {
	return &newsRepositoryImpl{
		db: db,
	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// Create creates a new news item in the database
func (r *newsRepositoryImpl) Create(ctx context.Context, news *entities.NewsItem) error {
	if err := r.db.WithContext(ctx).Create(news).Error; err != nil {
		return fmt.Errorf("failed to create news item: %w", err)
	}
	return nil
}

// ========================================
// READ OPERATIONS
// ========================================

// GetByID retrieves news item by its unique ID
func (r *newsRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.NewsItem, error) {
	var news entities.NewsItem
	if err := r.db.WithContext(ctx).First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("news item not found with id %s", id.String())
		}
		return nil, fmt.Errorf("failed to get news item by id: %w", err)
	}
	return &news, nil
}

// GetBySymbol retrieves news items for a specific stock symbol with pagination
func (r *newsRepositoryImpl) GetBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get news by symbol: %w", err)
	}

	return newsList, nil
}

// GetLatestBySymbol retrieves the latest news items for a stock symbol
func (r *newsRepositoryImpl) GetLatestBySymbol(ctx context.Context, symbol string, limit int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest news by symbol: %w", err)
	}

	return newsList, nil
}

// GetByTimeRange retrieves news items within a time range
func (r *newsRepositoryImpl) GetByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	if err := r.db.WithContext(ctx).
		Where("published_at BETWEEN ? AND ?", startTime, endTime).
		Order("published_at DESC").
		Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get news by time range: %w", err)
	}
	return newsList, nil
}

// GetRecent retrieves recent news items within specified hours
func (r *newsRepositoryImpl) GetRecent(ctx context.Context, hours int, limit int) ([]*entities.NewsItem, error) {
	since := time.Now().Add(time.Duration(-hours) * time.Hour)
	var newsList []*entities.NewsItem

	query := r.db.WithContext(ctx).
		Where("published_at >= ?", since).
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent news: %w", err)
	}

	return newsList, nil
}

// GetBySentiment retrieves news items by sentiment
func (r *newsRepositoryImpl) GetBySentiment(ctx context.Context, sentiment string, limit, offset int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("sentiment = ?", sentiment).
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get news by sentiment: %w", err)
	}

	return newsList, nil
}

// GetBySource retrieves news items by source
func (r *newsRepositoryImpl) GetBySource(ctx context.Context, source string, limit, offset int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("source = ?", source).
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get news by source: %w", err)
	}

	return newsList, nil
}

// GetTopSentimentNews retrieves news with the highest sentiment scores
func (r *newsRepositoryImpl) GetTopSentimentNews(ctx context.Context, limit int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Order("sentiment_score DESC").
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get top sentiment news: %w", err)
	}

	return newsList, nil
}

// GetTrendingNews retrieves trending news based on multiple symbols
func (r *newsRepositoryImpl) GetTrendingNews(ctx context.Context, symbols []string, limit int) ([]*entities.NewsItem, error) {
	if len(symbols) == 0 {
		return []*entities.NewsItem{}, nil
	}

	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("symbol IN ?", symbols).
		Where("published_at >= ?", time.Now().Add(-24*time.Hour)). // Last 24 hours
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get trending news: %w", err)
	}

	return newsList, nil
}

// GetToday retrieves today's news items
func (r *newsRepositoryImpl) GetToday(ctx context.Context, limit int) ([]*entities.NewsItem, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("published_at BETWEEN ? AND ?", today, tomorrow).
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get today's news: %w", err)
	}

	return newsList, nil
}

// GetByCategory retrieves news items by category
func (r *newsRepositoryImpl) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get news by category: %w", err)
	}

	return newsList, nil
}

// GetPositiveNews retrieves news with positive sentiment
func (r *newsRepositoryImpl) GetPositiveNews(ctx context.Context, limit int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("sentiment = ? OR sentiment_score > ?", "positive", 0.5).
		Order("sentiment_score DESC").
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get positive news: %w", err)
	}

	return newsList, nil
}

// GetNegativeNews retrieves news with negative sentiment
func (r *newsRepositoryImpl) GetNegativeNews(ctx context.Context, limit int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("sentiment = ? OR sentiment_score < ?", "negative", -0.5).
		Order("sentiment_score ASC").
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get negative news: %w", err)
	}

	return newsList, nil
}

// GetMarketNews retrieves general market news
func (r *newsRepositoryImpl) GetMarketNews(ctx context.Context, limit, offset int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("category = ? OR symbol = ?", "market", "").
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get market news: %w", err)
	}

	return newsList, nil
}

// GetLatestMarketNews retrieves the latest general market news
func (r *newsRepositoryImpl) GetLatestMarketNews(ctx context.Context, limit int) ([]*entities.NewsItem, error) {
	var newsList []*entities.NewsItem
	query := r.db.WithContext(ctx).
		Where("category = ? OR symbol = ?", "market", "").
		Order("published_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&newsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest market news: %w", err)
	}

	return newsList, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// Update updates an existing news item
func (r *newsRepositoryImpl) Update(ctx context.Context, news *entities.NewsItem) error {
	if err := r.db.WithContext(ctx).Save(news).Error; err != nil {
		return fmt.Errorf("failed to update news item: %w", err)
	}
	return nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// Delete removes news item by ID
func (r *newsRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.NewsItem{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete news item: %w", err)
	}
	return nil
}

// DeleteBySymbol removes all news items for a specific symbol
func (r *newsRepositoryImpl) DeleteBySymbol(ctx context.Context, symbol string) error {
	if err := r.db.WithContext(ctx).Where("symbol = ?", symbol).Delete(&entities.NewsItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete news by symbol: %w", err)
	}
	return nil
}

// ========================================
// BULK OPERATIONS
// ========================================

// BulkCreate creates multiple news items
func (r *newsRepositoryImpl) BulkCreate(ctx context.Context, newsList []*entities.NewsItem) error {
	if len(newsList) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(newsList, 100).Error; err != nil {
		return fmt.Errorf("failed to create news items in bulk: %w", err)
	}
	return nil
}

// BulkUpdate updates multiple news items
func (r *newsRepositoryImpl) BulkUpdate(ctx context.Context, newsList []*entities.NewsItem) error {
	if len(newsList) == 0 {
		return nil
	}

	for i := 0; i < len(newsList); i += 100 {
		end := i + 100
		if end > len(newsList) {
			end = len(newsList)
		}

		batch := newsList[i:end]
		for _, news := range batch {
			if err := r.db.WithContext(ctx).Save(news).Error; err != nil {
				return fmt.Errorf("failed to update news item in bulk: %w", err)
			}
		}
	}

	return nil
}

// UpsertByURL creates or updates news item by URL (to avoid duplicates)
func (r *newsRepositoryImpl) UpsertByURL(ctx context.Context, news *entities.NewsItem) error {
	var existing entities.NewsItem
	err := r.db.WithContext(ctx).
		Where("url = ?", news.URL).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record
		if err := r.db.WithContext(ctx).Create(news).Error; err != nil {
			return fmt.Errorf("failed to create news item during upsert: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing news item during upsert: %w", err)
	} else {
		// Update existing record
		news.ID = existing.ID
		news.CreatedAt = existing.CreatedAt
		if err := r.db.WithContext(ctx).Save(news).Error; err != nil {
			return fmt.Errorf("failed to update news item during upsert: %w", err)
		}
	}

	return nil
}

// ========================================
// DATA MANAGEMENT OPERATIONS
// ========================================

// CleanupOldNews removes news items older than the specified time
func (r *newsRepositoryImpl) CleanupOldNews(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("published_at < ?", olderThan).Delete(&entities.NewsItem{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old news: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// RemoveDuplicates removes duplicate news items based on URL
func (r *newsRepositoryImpl) RemoveDuplicates(ctx context.Context) (int64, error) {
	// Find duplicate URLs
	var duplicates []struct {
		URL   string
		Count int64
	}

	if err := r.db.WithContext(ctx).
		Model(&entities.NewsItem{}).
		Select("url, COUNT(*) as count").
		Group("url").
		Having("COUNT(*) > 1").
		Scan(&duplicates).Error; err != nil {
		return 0, fmt.Errorf("failed to find duplicate news: %w", err)
	}

	var totalDeleted int64

	// For each duplicate URL, keep the latest and delete the rest
	for _, dup := range duplicates {
		var newsItems []*entities.NewsItem
		if err := r.db.WithContext(ctx).
			Where("url = ?", dup.URL).
			Order("created_at DESC").
			Find(&newsItems).Error; err != nil {
			continue
		}

		// Keep the first (latest) one, delete the rest
		for i := 1; i < len(newsItems); i++ {
			if err := r.db.WithContext(ctx).Delete(newsItems[i]).Error; err != nil {
				continue
			}
			totalDeleted++
		}
	}

	return totalDeleted, nil
}

// ========================================
// STATISTICS OPERATIONS
// ========================================

// Count returns the total number of news items
func (r *newsRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.NewsItem{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count news items: %w", err)
	}
	return count, nil
}

// CountBySymbol returns the number of news items for a specific symbol
func (r *newsRepositoryImpl) CountBySymbol(ctx context.Context, symbol string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.NewsItem{}).Where("symbol = ?", symbol).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count news items by symbol: %w", err)
	}
	return count, nil
}

// GetSentimentDistribution returns the distribution of news by sentiment for a specific symbol
func (r *newsRepositoryImpl) GetSentimentDistribution(ctx context.Context, symbol string) (map[string]int64, error) {
	var results []struct {
		Sentiment string
		Count     int64
	}

	query := r.db.WithContext(ctx).
		Model(&entities.NewsItem{}).
		Select("sentiment, COUNT(*) as count").
		Group("sentiment")

	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get sentiment distribution: %w", err)
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Sentiment] = result.Count
	}

	return distribution, nil
}

// ========================================
// HEALTH CHECK OPERATIONS
// ========================================

// Health performs a health check on the repository
func (r *newsRepositoryImpl) Health(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.NewsItem{}).Limit(1).Count(&count).Error; err != nil {
		return fmt.Errorf("news repository health check failed: %w", err)
	}
	return nil
}
