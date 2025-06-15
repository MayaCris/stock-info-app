package request

import (
	"strings"

	"github.com/google/uuid"
)

// CreateCompanyRequest represents request to create a company
type CreateCompanyRequest struct {
	Ticker    string  `json:"ticker" binding:"required,min=1,max=10"`
	Name      string  `json:"name" binding:"required,min=2,max=200"`
	Sector    string  `json:"sector,omitempty"`
	MarketCap float64 `json:"market_cap,omitempty"`
	Exchange  string  `json:"exchange,omitempty"`
}

// UpdateCompanyRequest represents request to update a company
type UpdateCompanyRequest struct {
	Name      *string  `json:"name,omitempty" binding:"omitempty,min=2,max=200"`
	Sector    *string  `json:"sector,omitempty"`
	MarketCap *float64 `json:"market_cap,omitempty"`
	Exchange  *string  `json:"exchange,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
}

// CreateBrokerageRequest represents request to create a brokerage
type CreateBrokerageRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty" binding:"omitempty,url"`
}

// UpdateBrokerageRequest represents request to update a brokerage
type UpdateBrokerageRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty"`
	Website     *string `json:"website,omitempty" binding:"omitempty,url"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// CreateStockRatingRequest represents request to create a stock rating
type CreateStockRatingRequest struct {
	CompanyID   uuid.UUID `json:"company_id" binding:"required"`
	BrokerageID uuid.UUID `json:"brokerage_id" binding:"required"`
	Action      string    `json:"action" binding:"required"`
	RatingFrom  string    `json:"rating_from,omitempty"`
	RatingTo    string    `json:"rating_to,omitempty"`
	TargetFrom  string    `json:"target_from,omitempty"`
	TargetTo    string    `json:"target_to,omitempty"`
}

// StockRatingFilterRequest represents filters for stock ratings
type StockRatingFilterRequest struct {
	CompanyID   *uuid.UUID `form:"company_id"`
	BrokerageID *uuid.UUID `form:"brokerage_id"`
	Ticker      string     `form:"ticker"`
	Action      string     `form:"action"`
	RatingTo    string     `form:"rating_to"`
	DateFrom    string     `form:"date_from" binding:"omitempty,datetime=2006-01-02"`
	DateTo      string     `form:"date_to" binding:"omitempty,datetime=2006-01-02"`
}

// CompanyFilterRequest represents filters for companies
type CompanyFilterRequest struct {
	Ticker   string `form:"ticker"`
	Name     string `form:"name"`
	Sector   string `form:"sector"`
	Exchange string `form:"exchange"`
	IsActive *bool  `form:"is_active"`
}

// BrokerageFilterRequest represents filters for brokerages
type BrokerageFilterRequest struct {
	Name     string `form:"name"`
	IsActive *bool  `form:"is_active"`
}

// PopulateDatabaseRequest represents request to populate database
type PopulateDatabaseRequest struct {
	Mode       string `json:"mode" binding:"required,oneof=quick full incremental"`
	Pages      *int   `json:"pages,omitempty" binding:"omitempty,min=1,max=50"`
	BatchSize  *int   `json:"batch_size,omitempty" binding:"omitempty,min=10,max=1000"`
	DryRun     bool   `json:"dry_run,omitempty"`
	ClearFirst bool   `json:"clear_first,omitempty"`
}

// Validate validates the request and normalizes data
func (r *CreateCompanyRequest) Validate() error {
	r.Ticker = strings.ToUpper(strings.TrimSpace(r.Ticker))
	r.Name = strings.TrimSpace(r.Name)
	r.Sector = strings.TrimSpace(r.Sector)
	r.Exchange = strings.ToUpper(strings.TrimSpace(r.Exchange))
	return nil
}

// Validate validates the update request
func (r *UpdateCompanyRequest) Validate() error {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
	if r.Sector != nil {
		trimmed := strings.TrimSpace(*r.Sector)
		r.Sector = &trimmed
	}
	if r.Exchange != nil {
		upper := strings.ToUpper(strings.TrimSpace(*r.Exchange))
		r.Exchange = &upper
	}
	return nil
}

// Validate validates the brokerage request
func (r *CreateBrokerageRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.Website = strings.TrimSpace(r.Website)
	return nil
}

// Validate validates the update brokerage request
func (r *UpdateBrokerageRequest) Validate() error {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
	if r.Description != nil {
		trimmed := strings.TrimSpace(*r.Description)
		r.Description = &trimmed
	}
	if r.Website != nil {
		trimmed := strings.TrimSpace(*r.Website)
		r.Website = &trimmed
	}
	return nil
}
