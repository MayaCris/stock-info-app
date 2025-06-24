package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
)

// BrokerageRepository defines the contract for brokerage data access
type BrokerageRepository interface {
	// Create operations
	Create(ctx context.Context, brokerage *entities.Brokerage) error
	CreateMany(ctx context.Context, brokerages []*entities.Brokerage) error

	// Read operations
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Brokerage, error)
	GetByName(ctx context.Context, name string) (*entities.Brokerage, error)
	GetAll(ctx context.Context) ([]*entities.Brokerage, error)
	GetAllActive(ctx context.Context) ([]*entities.Brokerage, error)

	// Update operations
	Update(ctx context.Context, brokerage *entities.Brokerage) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error

	// Delete operations
	Delete(ctx context.Context, id uuid.UUID) error // Soft delete
	HardDelete(ctx context.Context, id uuid.UUID) error // Permanent delete

	// Query operations
	Exists(ctx context.Context, name string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)

	// Business operations (for API sync)
	FindOrCreate(ctx context.Context, name string) (*entities.Brokerage, error)
	FindOrCreateWithDetails(ctx context.Context, name, website, country string) (*entities.Brokerage, error)

	// Relationship operations
	GetWithRatings(ctx context.Context, id uuid.UUID) (*entities.Brokerage, error)
	GetByRatingCount(ctx context.Context, limit int) ([]*entities.Brokerage, error)
}