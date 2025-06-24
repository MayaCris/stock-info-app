package entities

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Brokerage represents a brokerage firm or financial institution
type Brokerage struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	Name   string    `json:"name" gorm:"type:string;unique;not null" validate:"required,min=2,max=100"`
	
	// Auditoría - timestamps automáticos por la BD
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Control de estado
	IsActive bool `json:"is_active" gorm:"default:true;not null"`
	
	// Metadatos opcionales
	Website string `json:"website,omitempty" gorm:"type:string;null"`
	Country string `json:"country,omitempty" gorm:"type:string;null" validate:"omitempty,len=3"`
	
	// Relationships
	StockRatings []StockRating `json:"stock_ratings,omitempty" gorm:"foreignKey:BrokerageID"`
}

// TableName specifies the table name for GORM
func (Brokerage) TableName() string {
	return "brokerages"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (b *Brokerage) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	// Solo normalización básica de datos
	b.normalizeName()
	b.normalizeWebsite()
	b.normalizeCountry()
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a record  
func (b *Brokerage) BeforeUpdate(tx *gorm.DB) error {
	b.normalizeName()
	b.normalizeWebsite()
	b.normalizeCountry()
	return nil
}

// Private normalization methods (domain logic)
func (b *Brokerage) normalizeName() {
	b.Name = strings.TrimSpace(b.Name)
}

func (b *Brokerage) normalizeWebsite() {
	if b.Website != "" {
		b.Website = strings.TrimSpace(strings.ToLower(b.Website))
		if !strings.HasPrefix(b.Website, "http://") && !strings.HasPrefix(b.Website, "https://") {
			b.Website = "https://" + b.Website
		}
	}
}

func (b *Brokerage) normalizeCountry() {
	if b.Country != "" {
		b.Country = strings.ToUpper(strings.TrimSpace(b.Country))
	}
}

// NewBrokerage creates a new Brokerage instance
func NewBrokerage(name string) *Brokerage {
	return &Brokerage{
		ID:       uuid.New(),
		Name:     strings.TrimSpace(name),
		IsActive: true,
	}
}

// NewBrokerageWithDetails creates a new Brokerage with additional details
func NewBrokerageWithDetails(name, website, country string) *Brokerage {
	brokerage := NewBrokerage(name)
	if website != "" {
		brokerage.Website = website
	}
	if country != "" {
		brokerage.Country = country
	}
	return brokerage
}

// IsValid validates the Brokerage entity (business rules)
func (b *Brokerage) IsValid() bool {
	if b.Name == "" || len(b.Name) < 2 || len(b.Name) > 100 {
		return false
	}
	// Validate country code if is present (ISO 3166-1 alpha-3)
	if b.Country != "" && len(b.Country) != 3 {
		return false
	}
	return true
}

// Activate marks the brokerage as active (state change - domain logic)
func (b *Brokerage) Activate() {
	b.IsActive = true
}

// Deactivate marks the brokerage as inactive (state change - domain logic)
func (b *Brokerage) Deactivate() {
	b.IsActive = false
}

// HasWebsite checks if the brokerage has a website configured (simple property check)
func (b *Brokerage) HasWebsite() bool {
	return strings.TrimSpace(b.Website) != ""
}

// String returns a string representation of the Brokerage
func (b *Brokerage) String() string {
	return b.Name
}