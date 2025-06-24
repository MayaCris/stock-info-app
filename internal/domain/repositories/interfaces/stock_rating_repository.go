package interfaces

import (
	"context"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/google/uuid"
)

// StockRatingRepository defines the contract for stock rating data access
// This is the most complex repository due to relationships and sync operations
type StockRatingRepository interface {
	// Create operations
	Create(ctx context.Context, rating *entities.StockRating) error
	CreateMany(ctx context.Context, ratings []*entities.StockRating) error

	// Read operations - Basic
	GetByID(ctx context.Context, id uuid.UUID) (*entities.StockRating, error)
	GetAll(ctx context.Context) ([]*entities.StockRating, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*entities.StockRating, error)
	GetByBrokerageID(ctx context.Context, brokerageID uuid.UUID) ([]*entities.StockRating, error)

	// Read operations - Advanced filtering
	GetByCompanyAndBrokerage(ctx context.Context, companyID, brokerageID uuid.UUID) ([]*entities.StockRating, error)
	GetByEventTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*entities.StockRating, error)
	GetByCompanyAndDateRange(ctx context.Context, companyID uuid.UUID, startTime, endTime time.Time) ([]*entities.StockRating, error)
	GetRecent(ctx context.Context, days int, limit int) ([]*entities.StockRating, error)

	// Read operations - By action type
	GetUpgrades(ctx context.Context, limit int) ([]*entities.StockRating, error)
	GetDowngrades(ctx context.Context, limit int) ([]*entities.StockRating, error)
	GetReiterations(ctx context.Context, limit int) ([]*entities.StockRating, error)
	GetByActionType(ctx context.Context, actionType string, limit int) ([]*entities.StockRating, error)

	// Update operations
	Update(ctx context.Context, rating *entities.StockRating) error
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
	MarkAsUnprocessed(ctx context.Context, id uuid.UUID) error
	MarkManyAsProcessed(ctx context.Context, ids []uuid.UUID) error

	// Delete operations
	Delete(ctx context.Context, id uuid.UUID) error     // Soft delete
	HardDelete(ctx context.Context, id uuid.UUID) error // Permanent delete

	// Query operations - Basic stats
	Count(ctx context.Context) (int64, error)
	CountByCompany(ctx context.Context, companyID uuid.UUID) (int64, error)
	CountByBrokerage(ctx context.Context, brokerageID uuid.UUID) (int64, error)
	CountByActionType(ctx context.Context, actionType string) (int64, error)

	// Business operations - CRITICAL for API sync
	FindExisting(ctx context.Context, companyID, brokerageID uuid.UUID, eventTime time.Time) (*entities.StockRating, error)
	FindOrCreateRating(ctx context.Context, companyID, brokerageID uuid.UUID, eventTime time.Time,
		action, ratingFrom, ratingTo, targetFrom, targetTo string, rawData []byte) (*entities.StockRating, error)
	UpsertMany(ctx context.Context, ratings []*entities.StockRating) error
	BulkInsertIgnoreDuplicates(ctx context.Context, ratings []*entities.StockRating) (int, error) // Returns count inserted

	// Processing operations (for background jobs)
	GetUnprocessed(ctx context.Context, limit int) ([]*entities.StockRating, error)
	GetUnprocessedBySource(ctx context.Context, source string, limit int) ([]*entities.StockRating, error)
	GetProcessingBatch(ctx context.Context, batchSize int) ([]*entities.StockRating, error)

	// Relationship operations - with preloading
	GetWithCompany(ctx context.Context, id uuid.UUID) (*entities.StockRating, error)
	GetWithBrokerage(ctx context.Context, id uuid.UUID) (*entities.StockRating, error)
	GetWithRelations(ctx context.Context, id uuid.UUID) (*entities.StockRating, error) // Both Company and Brokerage
	GetAllWithRelations(ctx context.Context, limit int) ([]*entities.StockRating, error)

	// Analytics operations
	GetActionTypeDistribution(ctx context.Context, days int) (map[string]int64, error)
	GetTopCompaniesByRatingCount(ctx context.Context, days int, limit int) ([]CompanyRatingCount, error)
	GetTopBrokeragesByRatingCount(ctx context.Context, days int, limit int) ([]BrokerageRatingCount, error)
	GetRatingTrend(ctx context.Context, companyID uuid.UUID, days int) ([]DailyRatingCount, error)

	// Time-based queries
	GetTodaysRatings(ctx context.Context) ([]*entities.StockRating, error)
	GetThisWeeksRatings(ctx context.Context) ([]*entities.StockRating, error)
	GetThisMonthsRatings(ctx context.Context) ([]*entities.StockRating, error)

	// Duplicate detection and cleanup
	FindDuplicates(ctx context.Context) ([]DuplicateGroup, error)
	RemoveDuplicates(ctx context.Context, keepNewest bool) (int, error) // Returns count removed

	// Data quality operations
	GetRatingsWithMissingData(ctx context.Context) ([]*entities.StockRating, error)
	GetRatingsWithInvalidDates(ctx context.Context) ([]*entities.StockRating, error)
	ValidateDataIntegrity(ctx context.Context) (DataIntegrityReport, error)

	// Orphan detection operations
	GetOrphanedStockRatings(ctx context.Context) ([]*entities.StockRating, error)
	GetOrphanedStockRatingsWithReasons(ctx context.Context) ([]OrphanedRatingResult, error)
}

// Supporting types for analytics operations

// CompanyRatingCount represents rating count per company
type CompanyRatingCount struct {
	CompanyID   uuid.UUID `json:"company_id"`
	CompanyName string    `json:"company_name"`
	Ticker      string    `json:"ticker"`
	RatingCount int64     `json:"rating_count"`
}

// BrokerageRatingCount represents rating count per brokerage
type BrokerageRatingCount struct {
	BrokerageID   uuid.UUID `json:"brokerage_id"`
	BrokerageName string    `json:"brokerage_name"`
	RatingCount   int64     `json:"rating_count"`
}

// DailyRatingCount represents rating count per day
type DailyRatingCount struct {
	Date         time.Time `json:"date"`
	RatingCount  int64     `json:"rating_count"`
	Upgrades     int64     `json:"upgrades"`
	Downgrades   int64     `json:"downgrades"`
	Reiterations int64     `json:"reiterations"`
}

// DuplicateGroup represents a group of duplicate ratings
type DuplicateGroup struct {
	CompanyID   uuid.UUID   `json:"company_id"`
	BrokerageID uuid.UUID   `json:"brokerage_id"`
	EventTime   time.Time   `json:"event_time"`
	RatingIDs   []uuid.UUID `json:"rating_ids"`
	Count       int         `json:"count"`
}

// DataIntegrityReport represents data quality metrics
type DataIntegrityReport struct {
	TotalRatings       int64 `json:"total_ratings"`
	MissingCompany     int64 `json:"missing_company"`
	MissingBrokerage   int64 `json:"missing_brokerage"`
	InvalidEventTime   int64 `json:"invalid_event_time"`
	EmptyAction        int64 `json:"empty_action"`
	DuplicateCount     int64 `json:"duplicate_count"`
	OrphanedRatings    int64 `json:"orphaned_ratings"` // Company/Brokerage not found
	ProcessedRatings   int64 `json:"processed_ratings"`
	UnprocessedRatings int64 `json:"unprocessed_ratings"`
}

// OrphanedRatingResult represents a stock rating with orphan details
type OrphanedRatingResult struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"company_id"`
	BrokerageID uuid.UUID `json:"brokerage_id"`
	EventTime   time.Time `json:"event_time"`
	Action      string    `json:"action"`
	Reason      string    `json:"reason"`
}
