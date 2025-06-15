package services

// ValidationThresholds - Umbrales para determinar severidad
type ValidationThresholds struct {
	// Umbrales de consistencia
	ConsistencyWarningLimit  int
	ConsistencyCriticalLimit int

	// Umbrales de duplicados
	DuplicatesWarningLimit  int
	DuplicatesCriticalLimit int

	// Umbrales de reglas de negocio
	BusinessRulesWarningLimit  int
	BusinessRulesCriticalLimit int
}

// EntityValidationRules - Reglas de validación por entidad
type EntityValidationRules struct {
	Company     CompanyValidationRules
	Brokerage   BrokerageValidationRules
	StockRating StockRatingValidationRules
}

type CompanyValidationRules struct {
	TickerMinLength       int
	TickerMaxLength       int
	NameMinLength         int
	NameMaxLength         int
	ViolationsForCritical int
}

type BrokerageValidationRules struct {
	NameMinLength         int
	NameMaxLength         int
	ViolationsForCritical int
}

type StockRatingValidationRules struct {
	MaxAgeYearsConsistency int
	MaxAgeYearsBusiness    int
	ViolationsForCritical  int
}

// ValidationConfig - Configuración completa de validaciones
type ValidationConfig struct {
	Thresholds ValidationThresholds
	Rules      EntityValidationRules
}

// DefaultValidationConfig - Configuración por defecto basada en valores actuales del código
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Thresholds: ValidationThresholds{
			// Basado en línea 116: report.TotalInconsistencies <= 5
			ConsistencyWarningLimit:  5,
			ConsistencyCriticalLimit: 15,
			// Basado en línea 159: report.TotalDuplicates <= 3
			DuplicatesWarningLimit:  3,
			DuplicatesCriticalLimit: 10,
			// Basado en línea 202: report.TotalViolations <= 5
			BusinessRulesWarningLimit:  5,
			BusinessRulesCriticalLimit: 20,
		},
		Rules: EntityValidationRules{
			Company: CompanyValidationRules{
				// Basado en línea 435: len(company.Ticker) < 1 || len(company.Ticker) > 10
				TickerMinLength: 1,
				TickerMaxLength: 10,
				// Basado en línea 440: len(company.Name) < 2 || len(company.Name) > 200
				NameMinLength: 2,
				NameMaxLength: 200,
				// Basado en línea 456: len(violations) > 2
				ViolationsForCritical: 2,
			},
			Brokerage: BrokerageValidationRules{
				// Basado en línea 484: len(brokerage.Name) < 2 || len(brokerage.Name) > 100
				NameMinLength: 2,
				NameMaxLength: 100,
				// Basado en línea 495: len(violations) > 1
				ViolationsForCritical: 1,
			},
			StockRating: StockRatingValidationRules{
				// Basado en línea 315: time.Now().AddDate(-10, 0, 0)
				MaxAgeYearsConsistency: 10,
				// Basado en línea 527: time.Now().AddDate(-20, 0, 0)
				MaxAgeYearsBusiness: 20,
				// Basado en línea 547: len(violations) > 2
				ViolationsForCritical: 2,
			},
		},
	}
}

// GetThresholdMessage - Genera mensajes descriptivos para violaciones de límites
func (c *ValidationConfig) GetLengthViolationMessage(entityType, fieldType string) string {
	switch entityType {
	case "company":
		switch fieldType {
		case "ticker":
			return "Ticker length must be 1-10 characters"
		case "name":
			return "Company name length must be 2-200 characters"
		}
	case "brokerage":
		switch fieldType {
		case "name":
			return "Brokerage name length must be 2-100 characters"
		}
	}
	return "Invalid length"
}

// GetTimeRangeViolationMessage - Genera mensajes para violaciones temporales
func (c *ValidationConfig) GetTimeRangeViolationMessage(context string) string {
	switch context {
	case "consistency":
		return "Event time is too old for consistency validation"
	case "business":
		return "Event time cannot be older than 20 years"
	case "future":
		return "Event time cannot be in the future"
	}
	return "Invalid event time"
}
