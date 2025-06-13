package entities

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Company represents a publicly traded company
type Company struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;not null"`
	Ticker string    `json:"ticker" gorm:"type:string;unique;not null" validate:"required,min=1,max=10"`
	Name   string    `json:"name" gorm:"type:string;not null" validate:"required,min=2,max=200"`
	
	// Auditoría - timestamps automáticos por la BD
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Control de estado
	IsActive bool `json:"is_active" gorm:"default:true;not null"`
	
	// Metadatos opcionales de la empresa
	Sector    string  `json:"sector,omitempty" gorm:"type:string;null"`
	MarketCap float64 `json:"market_cap,omitempty" gorm:"type:decimal(15,2);null"` // En millones USD
	Exchange  string  `json:"exchange,omitempty" gorm:"type:string;null"`          // NYSE, NASDAQ, etc.
	
	// Relationships
	StockRatings []StockRating `json:"stock_ratings,omitempty" gorm:"foreignKey:CompanyID"`
}

// TableName specifies the table name for GORM
func (Company) TableName() string {
	return "companies"
}

// BeforeCreate is a GORM hook that runs before creating a record
func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	// Solo normalización básica de datos
	c.normalizeTicker()
	c.normalizeName()
	c.normalizeExchange()
	c.normalizeSector()
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a record
func (c *Company) BeforeUpdate(tx *gorm.DB) error {
	c.normalizeTicker()
	c.normalizeName()
	c.normalizeExchange()
	c.normalizeSector()
	return nil
}

// Private normalization methods (domain logic)
func (c *Company) normalizeTicker() {
	c.Ticker = strings.ToUpper(strings.TrimSpace(c.Ticker))
}

func (c *Company) normalizeName() {
	c.Name = strings.TrimSpace(c.Name)
}

func (c *Company) normalizeExchange() {
	if c.Exchange != "" {
		c.Exchange = strings.ToUpper(strings.TrimSpace(c.Exchange))
	}
}

func (c *Company) normalizeSector() {
	if c.Sector != "" {
		// Capitalizar primera letra de cada palabra
		words := strings.Fields(strings.ToLower(strings.TrimSpace(c.Sector)))
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(string(word[0])) + word[1:]
			}
		}
		c.Sector = strings.Join(words, " ")
	}
}

// NewCompany creates a new Company instance with basic info
func NewCompany(ticker, name string) *Company {
	return &Company{
		ID:       uuid.New(),
		Ticker:   strings.ToUpper(strings.TrimSpace(ticker)),
		Name:     strings.TrimSpace(name),
		IsActive: true,
	}
}

// NewCompanyWithDetails creates a new Company with additional details
func NewCompanyWithDetails(ticker, name, sector, exchange string, marketCap float64) *Company {
	company := NewCompany(ticker, name)
	if sector != "" {
		company.Sector = sector
	}
	if exchange != "" {
		company.Exchange = exchange
	}
	if marketCap > 0 {
		company.MarketCap = marketCap
	}
	return company
}

// IsValid validates the Company entity (business rules)
func (c *Company) IsValid() bool {
	if c.Ticker == "" || len(c.Ticker) < 1 || len(c.Ticker) > 10 {
		return false
	}
	if c.Name == "" || len(c.Name) < 2 || len(c.Name) > 200 {
		return false
	}
	// Market cap no puede ser negativo (business rule)
	if c.MarketCap < 0 {
		return false
	}
	return true
}

// Activate marks the company as active (state change - domain logic)
func (c *Company) Activate() {
	c.IsActive = true
}

// Deactivate marks the company as inactive (state change - domain logic)
func (c *Company) Deactivate() {
	c.IsActive = false
}

// UpdateMarketCap updates the market capitalization (business rule: must be >= 0)
func (c *Company) UpdateMarketCap(marketCap float64) {
	if marketCap >= 0 {
		c.MarketCap = marketCap
	}
}

// String returns a string representation of the Company
func (c *Company) String() string {
	return c.Ticker + " - " + c.Name
}
