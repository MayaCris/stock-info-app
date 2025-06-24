package interfaces

import (
	"context"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionalCompanyRepository extends CompanyRepository with transactional operations
type TransactionalCompanyRepository interface {
	CompanyRepository

	// Transactional operations using provided transaction
	CreateWithTx(ctx context.Context, tx *gorm.DB, company *entities.Company) error
	CreateManyWithTx(ctx context.Context, tx *gorm.DB, companies []*entities.Company) error
	GetByTickerWithTx(ctx context.Context, tx *gorm.DB, ticker string) (*entities.Company, error)
	FindOrCreateByTickerWithTx(ctx context.Context, tx *gorm.DB, ticker, name string) (*entities.Company, error)
	// CreateIgnoreDuplicatesWithTx creates a company ignoring duplicates without aborting transaction
	CreateIgnoreDuplicatesWithTx(ctx context.Context, tx *gorm.DB, company *entities.Company) (*entities.Company, error)
}

// TransactionalBrokerageRepository extends BrokerageRepository with transactional operations
type TransactionalBrokerageRepository interface {
	BrokerageRepository

	// Transactional operations using provided transaction
	CreateWithTx(ctx context.Context, tx *gorm.DB, brokerage *entities.Brokerage) error
	CreateManyWithTx(ctx context.Context, tx *gorm.DB, brokerages []*entities.Brokerage) error
	GetByNameWithTx(ctx context.Context, tx *gorm.DB, name string) (*entities.Brokerage, error)
	FindOrCreateByNameWithTx(ctx context.Context, tx *gorm.DB, name string) (*entities.Brokerage, error)
	// CreateIgnoreDuplicatesWithTx creates a brokerage ignoring duplicates without aborting transaction
	CreateIgnoreDuplicatesWithTx(ctx context.Context, tx *gorm.DB, brokerage *entities.Brokerage) (*entities.Brokerage, error)
}

// TransactionalStockRatingRepository extends StockRatingRepository with transactional operations
type TransactionalStockRatingRepository interface {
	StockRatingRepository

	// Transactional operations using provided transaction
	CreateWithTx(ctx context.Context, tx *gorm.DB, stockRating *entities.StockRating) error
	CreateManyWithTx(ctx context.Context, tx *gorm.DB, stockRatings []*entities.StockRating) error
	GetByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.StockRating, error)
	BulkInsertIgnoreDuplicatesWithTx(ctx context.Context, tx *gorm.DB, stockRatings []*entities.StockRating) (int, error)
}
