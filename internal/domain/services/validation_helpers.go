package services

import (
	"fmt"
	"time"
)

// calculateSeverityByViolations calcula severidad basada en número de violaciones
func (s *IntegrityValidationServiceImpl) calculateSeverityByViolations(violations []string, entityType string) string {
	violationCount := len(violations)

	switch entityType {
	case "company":
		if violationCount > s.config.Rules.Company.ViolationsForCritical {
			return "critical"
		}
	case "brokerage":
		if violationCount > s.config.Rules.Brokerage.ViolationsForCritical {
			return "critical"
		}
	case "stock_rating":
		if violationCount > s.config.Rules.StockRating.ViolationsForCritical {
			return "critical"
		}
	}

	if violationCount > 0 {
		return "WARNING"
	}

	return "GOOD"
}

// isWithinValidTimeRange verifica si un tiempo está dentro del rango válido
func (s *IntegrityValidationServiceImpl) isWithinValidTimeRange(eventTime time.Time, context string) bool {
	now := time.Now()

	switch context {
	case "consistency":
		return eventTime.After(now.AddDate(-s.config.Rules.StockRating.MaxAgeYearsConsistency, 0, 0))
	case "business":
		return eventTime.After(now.AddDate(-s.config.Rules.StockRating.MaxAgeYearsBusiness, 0, 0))
	case "future":
		return !eventTime.After(now)
	default:
		return true
	}
}

// isValidLength verifica si un string tiene longitud válida según configuración
func (s *IntegrityValidationServiceImpl) isValidLength(text string, entityType, fieldType string) bool {
	length := len(text)

	switch entityType {
	case "company":
		switch fieldType {
		case "ticker":
			return length >= s.config.Rules.Company.TickerMinLength &&
				length <= s.config.Rules.Company.TickerMaxLength
		case "name":
			return length >= s.config.Rules.Company.NameMinLength &&
				length <= s.config.Rules.Company.NameMaxLength
		}
	case "brokerage":
		switch fieldType {
		case "name":
			return length >= s.config.Rules.Brokerage.NameMinLength &&
				length <= s.config.Rules.Brokerage.NameMaxLength
		}
	}

	return true
}

// getLengthViolationMessage genera mensaje descriptivo para violaciones de longitud
func (s *IntegrityValidationServiceImpl) getLengthViolationMessage(entityType, fieldType string) string {
	switch entityType {
	case "company":
		switch fieldType {
		case "ticker":
			return fmt.Sprintf("Ticker length must be %d-%d characters",
				s.config.Rules.Company.TickerMinLength,
				s.config.Rules.Company.TickerMaxLength)
		case "name":
			return fmt.Sprintf("Company name length must be %d-%d characters",
				s.config.Rules.Company.NameMinLength,
				s.config.Rules.Company.NameMaxLength)
		}
	case "brokerage":
		switch fieldType {
		case "name":
			return fmt.Sprintf("Brokerage name length must be %d-%d characters",
				s.config.Rules.Brokerage.NameMinLength,
				s.config.Rules.Brokerage.NameMaxLength)
		}
	}
	return "Invalid length"
}

// getTimeRangeViolationMessage genera mensaje descriptivo para violaciones temporales
func (s *IntegrityValidationServiceImpl) getTimeRangeViolationMessage(context string) string {
	switch context {
	case "consistency":
		return fmt.Sprintf("Event time is too old for consistency validation (max %d years)",
			s.config.Rules.StockRating.MaxAgeYearsConsistency)
	case "business":
		return fmt.Sprintf("Event time cannot be older than %d years",
			s.config.Rules.StockRating.MaxAgeYearsBusiness)
	case "future":
		return "Event time cannot be in the future"
	}
	return "Invalid event time"
}

// validateEntityIntegrity realiza validación integral de una entidad usando configuración
func (s *IntegrityValidationServiceImpl) validateEntityIntegrity(entityType string, violations []string) (string, bool) {
	if len(violations) == 0 {
		return "GOOD", true
	}

	severity := s.calculateSeverityByViolations(violations, entityType)
	return severity, false
}
