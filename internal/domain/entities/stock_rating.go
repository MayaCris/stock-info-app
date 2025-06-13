package entities

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockRating represents a stock rating/recommendation from a brokerage
type StockRating struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	CompanyID   uuid.UUID `json:"company_id" gorm:"type:uuid;not null" validate:"required"`
	BrokerageID uuid.UUID `json:"brokerage_id" gorm:"type:uuid;not null" validate:"required"`
	
	// Rating data (from API)
	Action     string `json:"action" gorm:"type:string;not null" validate:"required"`         // "upgraded by", "downgraded by", "reiterated by"
	RatingFrom string `json:"rating_from,omitempty" gorm:"type:string;null"`                 // "Buy", "Sell", "Hold", etc.
	RatingTo   string `json:"rating_to,omitempty" gorm:"type:string;null"`                   // "Buy", "Sell", "Hold", etc.
	TargetFrom string `json:"target_from,omitempty" gorm:"type:string;null"`                 // "$4.20"
	TargetTo   string `json:"target_to,omitempty" gorm:"type:string;null"`                   // "$4.70"
	
	// Timestamps
	EventTime time.Time `json:"event_time" gorm:"not null" validate:"required"`              // When the rating occurred (from API)
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;not null"`                   // When saved to our DB
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`                   // Last modification
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                                        // Soft delete
	
	// Processing metadata
	Source      string          `json:"source" gorm:"type:string;default:'api';not null"`     // Data source
	RawData     json.RawMessage `json:"raw_data,omitempty" gorm:"type:jsonb;null"`            // Original API response
	IsProcessed bool            `json:"is_processed" gorm:"default:false;not null"`           // Processing status
	
	// Relationships
	Company   Company   `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Brokerage Brokerage `json:"brokerage,omitempty" gorm:"foreignKey:BrokerageID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (StockRating) TableName() string {
	return "stock_ratings"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (sr *StockRating) BeforeCreate(tx *gorm.DB) error {
	if sr.ID == uuid.Nil {
		sr.ID = uuid.New()
	}
	// Solo normalización básica de datos
	sr.normalizeAction()
	sr.normalizeRatings()
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a record
func (sr *StockRating) BeforeUpdate(tx *gorm.DB) error {
	sr.normalizeAction()
	sr.normalizeRatings()
	return nil
}

// Private normalization methods (domain logic)
func (sr *StockRating) normalizeAction() {
	sr.Action = strings.ToLower(strings.TrimSpace(sr.Action))
}

func (sr *StockRating) normalizeRatings() {
	if sr.RatingFrom != "" {
		sr.RatingFrom = strings.TrimSpace(sr.RatingFrom)
	}
	if sr.RatingTo != "" {
		sr.RatingTo = strings.TrimSpace(sr.RatingTo)
	}
}

// NewStockRating creates a new StockRating instance
func NewStockRating(companyID, brokerageID uuid.UUID, action string, eventTime time.Time) *StockRating {
	return &StockRating{
		ID:          uuid.New(),
		CompanyID:   companyID,
		BrokerageID: brokerageID,
		Action:      action,
		EventTime:   eventTime,
		Source:      "api",
		IsProcessed: false,
	}
}

// IsValid validates the StockRating entity (business rules)
func (sr *StockRating) IsValid() bool {
	if sr.CompanyID == uuid.Nil || sr.BrokerageID == uuid.Nil {
		return false
	}
	if sr.Action == "" {
		return false
	}
	if sr.EventTime.IsZero() {
		return false
	}
	return true
}

// MarkAsProcessed marks the rating as processed (state change - domain logic)
func (sr *StockRating) MarkAsProcessed() {
	sr.IsProcessed = true
}

// MarkAsUnprocessed marks the rating as unprocessed (state change - domain logic)
func (sr *StockRating) MarkAsUnprocessed() {
	sr.IsProcessed = false
}

// Basic domain logic for action classification
func (sr *StockRating) IsUpgrade() bool {
	return strings.Contains(strings.ToLower(sr.Action), "upgrade")
}

func (sr *StockRating) IsDowngrade() bool {
	return strings.Contains(strings.ToLower(sr.Action), "downgrade")
}

func (sr *StockRating) IsReiteration() bool {
	return strings.Contains(strings.ToLower(sr.Action), "reiterat")
}

// HasRatingChange checks if the rating changed (simple comparison - domain logic)
func (sr *StockRating) HasRatingChange() bool {
	return sr.RatingFrom != "" && sr.RatingTo != "" && sr.RatingFrom != sr.RatingTo
}

// HasTargetChange checks if the price target changed (simple comparison - domain logic)
func (sr *StockRating) HasTargetChange() bool {
	return sr.TargetFrom != "" && sr.TargetTo != "" && sr.TargetFrom != sr.TargetTo
}

// String returns a string representation of the StockRating
func (sr *StockRating) String() string {
	return sr.Action
}
