package interfaces

import (
	"context"

	"github.com/MayaCris/stock-info-app/internal/application/dto/request"
	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/google/uuid"
)

// CompanyService defines the interface for company business logic
type CompanyService interface {
	// CRUD operations
	CreateCompany(ctx context.Context, req *request.CreateCompanyRequest) (*response.CompanyResponse, error)
	GetCompanyByID(ctx context.Context, id uuid.UUID) (*response.CompanyResponse, error)
	GetCompanyByTicker(ctx context.Context, ticker string) (*response.CompanyResponse, error)
	UpdateCompany(ctx context.Context, id uuid.UUID, req *request.UpdateCompanyRequest) (*response.CompanyResponse, error)
	DeleteCompany(ctx context.Context, id uuid.UUID) error
	// List operations
	ListCompanies(ctx context.Context, filter *request.CompanyFilterRequest, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error)
	ListActiveCompanies(ctx context.Context, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error)

	// Business operations
	ActivateCompany(ctx context.Context, id uuid.UUID) error
	DeactivateCompany(ctx context.Context, id uuid.UUID) error
	UpdateMarketCap(ctx context.Context, ticker string, marketCap float64) error

	// Search operations
	SearchCompaniesByName(ctx context.Context, name string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error)
	GetCompaniesBySector(ctx context.Context, sector string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.CompanyListResponse], error)
}

// BrokerageService defines the interface for brokerage business logic
type BrokerageService interface {
	// CRUD operations
	CreateBrokerage(ctx context.Context, req *request.CreateBrokerageRequest) (*response.BrokerageResponse, error)
	GetBrokerageByID(ctx context.Context, id uuid.UUID) (*response.BrokerageResponse, error)
	UpdateBrokerage(ctx context.Context, id uuid.UUID, req *request.UpdateBrokerageRequest) (*response.BrokerageResponse, error)
	DeleteBrokerage(ctx context.Context, id uuid.UUID) error
	// List operations
	ListBrokerages(ctx context.Context, filter *request.BrokerageFilterRequest, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.BrokerageResponse], error)
	ListActiveBrokerages(ctx context.Context, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.BrokerageResponse], error)

	// Business operations
	ActivateBrokerage(ctx context.Context, id uuid.UUID) error
	DeactivateBrokerage(ctx context.Context, id uuid.UUID) error

	// Search operations
	SearchBrokeragesByName(ctx context.Context, name string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.BrokerageResponse], error)
}

// StockRatingService defines the interface for stock rating business logic
type StockRatingService interface {
	// CRUD operations
	CreateStockRating(ctx context.Context, req *request.CreateStockRatingRequest) (*response.StockRatingResponse, error)
	GetStockRatingByID(ctx context.Context, id uuid.UUID) (*response.StockRatingResponse, error)
	DeleteStockRating(ctx context.Context, id uuid.UUID) error
	// List operations
	ListStockRatings(ctx context.Context, filter *request.StockRatingFilterRequest, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error)
	GetRatingsByCompany(ctx context.Context, companyID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error)
	GetRatingsByTicker(ctx context.Context, ticker string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error)
	GetRatingsByBrokerage(ctx context.Context, brokerageID uuid.UUID, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error)

	// Analytics operations
	GetRecentRatings(ctx context.Context, limit int) ([]*response.StockRatingListResponse, error)
	GetRatingsByDateRange(ctx context.Context, startDate, endDate string, pagination *response.PaginationRequest) (*response.PaginatedResponse[*response.StockRatingListResponse], error)
	GetRatingStatsByCompany(ctx context.Context, companyID uuid.UUID) (map[string]interface{}, error)
}

// AnalysisService defines the interface for analysis and recommendation business logic
type AnalysisService interface {
	// Company analysis
	GetCompanyAnalysis(ctx context.Context, companyID uuid.UUID) (*response.AnalysisResponse, error)
	GetCompanyAnalysisByTicker(ctx context.Context, ticker string) (*response.AnalysisResponse, error)

	// Market analysis
	GetMarketOverview(ctx context.Context) (map[string]interface{}, error)
	GetSectorAnalysis(ctx context.Context, sector string) (map[string]interface{}, error)
	GetTopRatedCompanies(ctx context.Context, limit int) ([]*response.CompanyListResponse, error)

	// Trend analysis
	GetRatingTrends(ctx context.Context, period string) (map[string]interface{}, error)
	GetBrokerageActivity(ctx context.Context, period string) (map[string]interface{}, error)

	// Recommendations
	GenerateRecommendation(ctx context.Context, companyID uuid.UUID) (string, error)
	GetRecommendationsByRating(ctx context.Context, rating string, limit int) ([]*response.CompanyListResponse, error)
}

// AdminService defines the interface for administrative operations
type AdminService interface {
	// Database operations
	PopulateDatabase(ctx context.Context, req *request.PopulateDatabaseRequest) (map[string]interface{}, error)
	ValidateDatabase(ctx context.Context) (map[string]interface{}, error)

	// Cache operations
	ClearCache(ctx context.Context) error
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)

	// System operations
	GetSystemHealth(ctx context.Context) (*response.HealthCheckResponse, error)
	GetSystemStats(ctx context.Context) (map[string]interface{}, error)
}
