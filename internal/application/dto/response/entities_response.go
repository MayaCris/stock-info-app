package response

import (
	"time"

	"github.com/google/uuid"
)

// CompanyResponse represents a company in API responses
type CompanyResponse struct {
	ID        uuid.UUID `json:"id"`
	Ticker    string    `json:"ticker"`
	Name      string    `json:"name"`
	Sector    string    `json:"sector,omitempty"`
	MarketCap float64   `json:"market_cap,omitempty"`
	Exchange  string    `json:"exchange,omitempty"`
	Logo      string    `json:"logo,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CompanyListResponse represents a simplified company for list views
type CompanyListResponse struct {
	ID       uuid.UUID `json:"id"`
	Ticker   string    `json:"ticker"`
	Name     string    `json:"name"`
	Sector   string    `json:"sector,omitempty"`
	Exchange string    `json:"exchange,omitempty"`
	Logo     string    `json:"logo,omitempty"`
	IsActive bool      `json:"is_active"`
}

// BrokerageResponse represents a brokerage in API responses
type BrokerageResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Website     string    `json:"website,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StockRatingResponse represents a stock rating in API responses
type StockRatingResponse struct {
	ID          uuid.UUID          `json:"id"`
	CompanyID   uuid.UUID          `json:"company_id"`
	BrokerageID uuid.UUID          `json:"brokerage_id"`
	Company     *CompanyResponse   `json:"company,omitempty"`
	Brokerage   *BrokerageResponse `json:"brokerage,omitempty"`
	Action      string             `json:"action"`
	RatingFrom  string             `json:"rating_from,omitempty"`
	RatingTo    string             `json:"rating_to,omitempty"`
	TargetFrom  string             `json:"target_from,omitempty"`
	TargetTo    string             `json:"target_to,omitempty"`
	EventTime   time.Time          `json:"event_time"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// StockRatingListResponse represents a simplified stock rating for list views
type StockRatingListResponse struct {
	ID        uuid.UUID `json:"id"`
	CompanyID uuid.UUID `json:"company_id"`
	Ticker    string    `json:"ticker"`
	Company   string    `json:"company_name"`
	Brokerage string    `json:"brokerage_name"`
	Action    string    `json:"action"`
	RatingTo  string    `json:"rating_to,omitempty"`
	TargetTo  string    `json:"target_to,omitempty"`
	EventTime time.Time `json:"event_time"`
}

// HealthCheckResponse represents health check status
type HealthCheckResponse struct {
	Status    string                       `json:"status"`
	Timestamp time.Time                    `json:"timestamp"`
	Version   string                       `json:"version"`
	Checks    map[string]HealthCheckDetail `json:"checks"`
}

// HealthCheckDetail represents individual health check details
type HealthCheckDetail struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// AnalysisResponse represents analysis results
type AnalysisResponse struct {
	CompanyID      uuid.UUID                 `json:"company_id"`
	Ticker         string                    `json:"ticker"`
	CompanyName    string                    `json:"company_name"`
	TotalRatings   int                       `json:"total_ratings"`
	RecentRatings  []StockRatingListResponse `json:"recent_ratings"`
	Recommendation string                    `json:"recommendation"`
	Summary        map[string]interface{}    `json:"summary"`
	GeneratedAt    time.Time                 `json:"generated_at"`
}
