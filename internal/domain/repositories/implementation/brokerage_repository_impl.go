package implementation

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// brokerageRepositoryImpl implements the BrokerageRepository interface using GORM
type brokerageRepositoryImpl struct {
	db *gorm.DB
}

// NewBrokerageRepository creates a new brokerage repository implementation
func NewBrokerageRepository(db *gorm.DB) interfaces.BrokerageRepository {
	return &brokerageRepositoryImpl{
		db: db,
	}
}

// NewTransactionalBrokerageRepository creates a new transactional brokerage repository implementation
func NewTransactionalBrokerageRepository(db *gorm.DB) interfaces.TransactionalBrokerageRepository {
	return &brokerageRepositoryImpl{
		db: db,
	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// Create creates a new brokerage in the database
func (r *brokerageRepositoryImpl) Create(ctx context.Context, brokerage *entities.Brokerage) error {
	if err := r.db.WithContext(ctx).Create(brokerage).Error; err != nil {
		return fmt.Errorf("failed to create brokerage: %w", err)
	}
	return nil
}

// CreateMany creates multiple brokerages in a single transaction
func (r *brokerageRepositoryImpl) CreateMany(ctx context.Context, brokerages []*entities.Brokerage) error {
	if len(brokerages) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, brokerage := range brokerages {
			if err := tx.Create(brokerage).Error; err != nil {
				return fmt.Errorf("failed to create brokerage %s: %w", brokerage.Name, err)
			}
		}
		return nil
	})
}

// ========================================
// READ OPERATIONS
// ========================================

// GetByID retrieves a brokerage by its ID
func (r *brokerageRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Brokerage, error) {
	var brokerage entities.Brokerage

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&brokerage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("brokerage with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get brokerage by id: %w", err)
	}

	return &brokerage, nil
}

// GetByName retrieves a brokerage by its name
func (r *brokerageRepositoryImpl) GetByName(ctx context.Context, name string) (*entities.Brokerage, error) {
	var brokerage entities.Brokerage

	err := r.db.WithContext(ctx).Where("name = ?", name).First(&brokerage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("brokerage with name %s not found", name)
		}
		return nil, fmt.Errorf("failed to get brokerage by name: %w", err)
	}

	return &brokerage, nil
}

// GetAll retrieves all brokerages (including inactive)
func (r *brokerageRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Brokerage, error) {
	var brokerages []*entities.Brokerage

	err := r.db.WithContext(ctx).Find(&brokerages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all brokerages: %w", err)
	}

	return brokerages, nil
}

// GetAllActive retrieves only active brokerages
func (r *brokerageRepositoryImpl) GetAllActive(ctx context.Context) ([]*entities.Brokerage, error) {
	var brokerages []*entities.Brokerage

	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&brokerages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active brokerages: %w", err)
	}

	return brokerages, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// Update updates an existing brokerage
func (r *brokerageRepositoryImpl) Update(ctx context.Context, brokerage *entities.Brokerage) error {
	result := r.db.WithContext(ctx).Save(brokerage)
	if result.Error != nil {
		return fmt.Errorf("failed to update brokerage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("brokerage with id %s not found for update", brokerage.ID)
	}

	return nil
}

// Activate activates a brokerage by ID
func (r *brokerageRepositoryImpl) Activate(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entities.Brokerage{}).Where("id = ?", id).Update("is_active", true)
	if result.Error != nil {
		return fmt.Errorf("failed to activate brokerage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("brokerage with id %s not found for activation", id)
	}

	return nil
}

// Deactivate deactivates a brokerage by ID
func (r *brokerageRepositoryImpl) Deactivate(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entities.Brokerage{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to deactivate brokerage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("brokerage with id %s not found for deactivation", id)
	}

	return nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// Delete performs a soft delete on a brokerage
func (r *brokerageRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entities.Brokerage{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete brokerage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("brokerage with id %s not found for deletion", id)
	}

	return nil
}

// HardDelete permanently deletes a brokerage from the database
func (r *brokerageRepositoryImpl) HardDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entities.Brokerage{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to hard delete brokerage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("brokerage with id %s not found for hard deletion", id)
	}

	return nil
}

// ========================================
// QUERY OPERATIONS
// ========================================

// Exists checks if a brokerage with the given name exists
func (r *brokerageRepositoryImpl) Exists(ctx context.Context, name string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.Brokerage{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if brokerage exists: %w", err)
	}

	return count > 0, nil
}

// Count returns the total number of brokerages (including inactive)
func (r *brokerageRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.Brokerage{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count brokerages: %w", err)
	}

	return count, nil
}

// CountActive returns the number of active brokerages
func (r *brokerageRepositoryImpl) CountActive(ctx context.Context) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entities.Brokerage{}).Where("is_active = ?", true).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count active brokerages: %w", err)
	}

	return count, nil
}

// ========================================
// BUSINESS OPERATIONS (CRITICAL FOR API SYNC)
// ========================================

// FindOrCreate finds a brokerage by name or creates it if it doesn't exist
func (r *brokerageRepositoryImpl) FindOrCreate(ctx context.Context, name string) (*entities.Brokerage, error) {
	// First, try to find existing brokerage
	brokerage, err := r.GetByName(ctx, name)
	if err == nil {
		return brokerage, nil
	}

	// If not found, create a new one
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		(err != nil && fmt.Sprintf("%v", err) == fmt.Sprintf("brokerage with name %s not found", name)) {

		newBrokerage := entities.NewBrokerage(name)
		if err := r.Create(ctx, newBrokerage); err != nil {
			return nil, fmt.Errorf("failed to create new brokerage %s: %w", name, err)
		}

		return newBrokerage, nil
	}

	// If there was a different error, return it
	return nil, fmt.Errorf("failed to find or create brokerage: %w", err)
}

// FindOrCreateWithDetails finds or creates a brokerage with additional details
func (r *brokerageRepositoryImpl) FindOrCreateWithDetails(ctx context.Context, name, website, country string) (*entities.Brokerage, error) {
	// First, try to find existing brokerage
	brokerage, err := r.GetByName(ctx, name)
	if err == nil {
		// Update details if they're empty and we have new data
		updated := false
		if brokerage.Website == "" && website != "" {
			brokerage.Website = website
			updated = true
		}
		if brokerage.Country == "" && country != "" {
			brokerage.Country = country
			updated = true
		}

		// Save updates if any
		if updated {
			if err := r.Update(ctx, brokerage); err != nil {
				return nil, fmt.Errorf("failed to update brokerage details: %w", err)
			}
		}

		return brokerage, nil
	}

	// If not found, create a new one with details
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		(err != nil && fmt.Sprintf("%v", err) == fmt.Sprintf("brokerage with name %s not found", name)) {

		newBrokerage := entities.NewBrokerageWithDetails(name, website, country)
		if err := r.Create(ctx, newBrokerage); err != nil {
			return nil, fmt.Errorf("failed to create new brokerage with details %s: %w", name, err)
		}

		return newBrokerage, nil
	}

	// If there was a different error, return it
	return nil, fmt.Errorf("failed to find or create brokerage with details: %w", err)
}

// ========================================
// RELATIONSHIP OPERATIONS
// ========================================

// GetWithRatings retrieves a brokerage with its stock ratings preloaded
func (r *brokerageRepositoryImpl) GetWithRatings(ctx context.Context, id uuid.UUID) (*entities.Brokerage, error) {
	var brokerage entities.Brokerage

	err := r.db.WithContext(ctx).Preload("StockRatings").Where("id = ?", id).First(&brokerage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("brokerage with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get brokerage with ratings: %w", err)
	}

	return &brokerage, nil
}

// GetByRatingCount retrieves brokerages ordered by their rating count (most active first)
func (r *brokerageRepositoryImpl) GetByRatingCount(ctx context.Context, limit int) ([]*entities.Brokerage, error) {
	var brokerages []*entities.Brokerage

	query := r.db.WithContext(ctx).
		Select("brokerages.*, COUNT(stock_ratings.id) as rating_count").
		Joins("LEFT JOIN stock_ratings ON brokerages.id = stock_ratings.brokerage_id").
		Group("brokerages.id").
		Order("rating_count DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&brokerages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get brokerages by rating count: %w", err)
	}

	return brokerages, nil
}

// ========================================
// TRANSACTIONAL OPERATIONS
// ========================================

// CreateWithTx creates a new brokerage using the provided transaction
func (r *brokerageRepositoryImpl) CreateWithTx(ctx context.Context, tx *gorm.DB, brokerage *entities.Brokerage) error {
	if err := tx.WithContext(ctx).Create(brokerage).Error; err != nil {
		return fmt.Errorf("failed to create brokerage with transaction: %w", err)
	}
	return nil
}

// CreateManyWithTx creates multiple brokerages using the provided transaction
func (r *brokerageRepositoryImpl) CreateManyWithTx(ctx context.Context, tx *gorm.DB, brokerages []*entities.Brokerage) error {
	if len(brokerages) == 0 {
		return nil
	}

	for _, brokerage := range brokerages {
		if err := tx.WithContext(ctx).Create(brokerage).Error; err != nil {
			return fmt.Errorf("failed to create brokerage %s with transaction: %w", brokerage.Name, err)
		}
	}
	return nil
}

// GetByNameWithTx retrieves a brokerage by name using the provided transaction
func (r *brokerageRepositoryImpl) GetByNameWithTx(ctx context.Context, tx *gorm.DB, name string) (*entities.Brokerage, error) {
	var brokerage entities.Brokerage

	err := tx.WithContext(ctx).Where("name = ?", name).First(&brokerage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("brokerage with name %s not found", name)
		}
		return nil, fmt.Errorf("failed to get brokerage by name with transaction: %w", err)
	}

	return &brokerage, nil
}

// FindOrCreateByNameWithTx finds or creates a brokerage by name using the provided transaction
func (r *brokerageRepositoryImpl) FindOrCreateByNameWithTx(ctx context.Context, tx *gorm.DB, name string) (*entities.Brokerage, error) {
	// Try to find existing brokerage
	brokerage, err := r.GetByNameWithTx(ctx, tx, name)
	if err == nil {
		return brokerage, nil // Found existing
	}

	// Create new brokerage if not found
	newBrokerage := entities.NewBrokerage(name)
	if err := r.CreateWithTx(ctx, tx, newBrokerage); err != nil {
		return nil, fmt.Errorf("failed to create brokerage %s: %w", name, err)
	}

	return newBrokerage, nil
}

// CreateIgnoreDuplicatesWithTx creates a brokerage ignoring duplicates without aborting transaction
func (r *brokerageRepositoryImpl) CreateIgnoreDuplicatesWithTx(ctx context.Context, tx *gorm.DB, brokerage *entities.Brokerage) (*entities.Brokerage, error) {
	// Use raw SQL with ON CONFLICT DO NOTHING to avoid transaction aborts
	query := `
		INSERT INTO brokerages (id, name, website, country, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (name) DO NOTHING
		RETURNING id, name, website, country, is_active, created_at, updated_at
	`

	var result entities.Brokerage
	err := tx.WithContext(ctx).Raw(query,
		brokerage.ID,
		brokerage.Name,
		brokerage.Website,
		brokerage.Country,
		brokerage.IsActive,
	).Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to create brokerage with duplicate handling: %w", err)
	}

	// If no rows were returned (conflict occurred), fetch the existing brokerage
	if result.ID == uuid.Nil {
		existing, err := r.GetByNameWithTx(ctx, tx, brokerage.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing brokerage after conflict: %w", err)
		}
		return existing, nil
	}

	return &result, nil
}
