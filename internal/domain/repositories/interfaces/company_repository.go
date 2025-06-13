package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
)

// CompanyRepository defines the contract for company data access
type CompanyRepository interface {
	// Create operations
	Create(ctx context.Context, company *entities.Company) error
	CreateMany(ctx context.Context, companies []*entities.Company) error

	// Read operations
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Company, error)
	GetByTicker(ctx context.Context, ticker string) (*entities.Company, error)
	GetByName(ctx context.Context, name string) (*entities.Company, error)
	GetAll(ctx context.Context) ([]*entities.Company, error)
	GetAllActive(ctx context.Context) ([]*entities.Company, error)

	// Update operations
	Update(ctx context.Context, company *entities.Company) error
	UpdateMarketCap(ctx context.Context, ticker string, marketCap float64) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error

	// Delete operations
	Delete(ctx context.Context, id uuid.UUID) error // Soft delete
	HardDelete(ctx context.Context, id uuid.UUID) error // Permanent delete

	// Query operations - Basic
	ExistsByTicker(ctx context.Context, ticker string) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)

	// Query operations - Financial
	GetBySector(ctx context.Context, sector string) ([]*entities.Company, error)
	GetByExchange(ctx context.Context, exchange string) ([]*entities.Company, error)
	GetByMarketCapRange(ctx context.Context, minCap, maxCap float64) ([]*entities.Company, error)
	GetLargestByMarketCap(ctx context.Context, limit int) ([]*entities.Company, error)

	// Business operations (for API sync) - CRITICAL
	FindOrCreateByTicker(ctx context.Context, ticker, name string) (*entities.Company, error)
	FindOrCreateWithDetails(ctx context.Context, ticker, name, sector, exchange string, marketCap float64) (*entities.Company, error)
	
	// Batch operations for sync
	UpsertMany(ctx context.Context, companies []*entities.Company) error
	
	// Relationship operations
	GetWithRatings(ctx context.Context, id uuid.UUID) (*entities.Company, error)
	GetByRatingCount(ctx context.Context, limit int) ([]*entities.Company, error)
	GetMostActiveCompanies(ctx context.Context, days int, limit int) ([]*entities.Company, error)

	// Search operations
	SearchByName(ctx context.Context, query string, limit int) ([]*entities.Company, error)
	SearchByTicker(ctx context.Context, query string, limit int) ([]*entities.Company, error)

	// Analytics operations
	GetSectorDistribution(ctx context.Context) (map[string]int64, error)
	GetExchangeDistribution(ctx context.Context) (map[string]int64, error)
	GetMarketCapStats(ctx context.Context) (map[string]float64, error) // min, max, avg, median
}