package services

import (
	"context"
	"fmt"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/google/uuid"
)

// IntegrityValidationService defines the contract for database integrity validation
type IntegrityValidationService interface {
	// ValidateFullIntegrity performs comprehensive integrity validation
	ValidateFullIntegrity(ctx context.Context) (*IntegrityReport, error)

	// ValidateOrphanedRecords checks for orphaned records across tables
	ValidateOrphanedRecords(ctx context.Context) (*OrphanReport, error)

	// ValidateDataConsistency checks for data consistency issues
	ValidateDataConsistency(ctx context.Context) (*ConsistencyReport, error)

	// ValidateDuplicates checks for duplicate records
	ValidateDuplicates(ctx context.Context) (*DuplicateReport, error)

	// ValidateBusinessRules checks business logic constraints
	ValidateBusinessRules(ctx context.Context) (*BusinessRuleReport, error)

	// RepairMinorIssues attempts to fix minor integrity issues automatically
	RepairMinorIssues(ctx context.Context, dryRun bool) (*RepairReport, error)
}

// IntegrityReport contains comprehensive integrity validation results
type IntegrityReport struct {
	Timestamp         time.Time              `json:"timestamp"`
	OverallStatus     IntegrityStatus        `json:"overall_status"`
	TotalIssues       int                    `json:"total_issues"`
	CriticalIssues    int                    `json:"critical_issues"`
	WarningIssues     int                    `json:"warning_issues"`
	OrphanReport      *OrphanReport          `json:"orphan_report"`
	ConsistencyReport *ConsistencyReport     `json:"consistency_report"`
	DuplicateReport   *DuplicateReport       `json:"duplicate_report"`
	BusinessReport    *BusinessRuleReport    `json:"business_report"`
	Summary           map[string]interface{} `json:"summary"`
	Duration          time.Duration          `json:"duration"`
}

// IntegrityStatus represents the overall integrity status
type IntegrityStatus string

const (
	IntegrityStatusHealthy  IntegrityStatus = "GOOD"
	IntegrityStatusWarning  IntegrityStatus = "WARNING"
	IntegrityStatusCritical IntegrityStatus = "CRITICAL"
	IntegrityStatusFailed   IntegrityStatus = "FAILED"
)

// OrphanReport contains information about orphaned records
type OrphanReport struct {
	OrphanedStockRatings []OrphanedStockRating `json:"orphaned_stock_ratings"`
	TotalOrphans         int                   `json:"total_orphans"`
	Status               IntegrityStatus       `json:"status"`
}

// ConsistencyReport contains data consistency validation results
type ConsistencyReport struct {
	InconsistentCompanies  []InconsistentCompany   `json:"inconsistent_companies"`
	InconsistentBrokerages []InconsistentBrokerage `json:"inconsistent_brokerages"`
	InconsistentRatings    []InconsistentRating    `json:"inconsistent_ratings"`
	TotalInconsistencies   int                     `json:"total_inconsistencies"`
	Status                 IntegrityStatus         `json:"status"`
}

// DuplicateReport contains duplicate record information
type DuplicateReport struct {
	DuplicateCompanies    []DuplicateCompany     `json:"duplicate_companies"`
	DuplicateBrokerages   []DuplicateBrokerage   `json:"duplicate_brokerages"`
	DuplicateStockRatings []DuplicateStockRating `json:"duplicate_stock_ratings"`
	TotalDuplicates       int                    `json:"total_duplicates"`
	Status                IntegrityStatus        `json:"status"`
}

// BusinessRuleReport contains business rule validation results
type BusinessRuleReport struct {
	InvalidCompanies    []InvalidCompany     `json:"invalid_companies"`
	InvalidBrokerages   []InvalidBrokerage   `json:"invalid_brokerages"`
	InvalidStockRatings []InvalidStockRating `json:"invalid_stock_ratings"`
	TotalViolations     int                  `json:"total_violations"`
	Status              IntegrityStatus      `json:"status"`
}

// RepairReport contains results of automatic repair operations
type RepairReport struct {
	RepairedOrphans      int                 `json:"repaired_orphans"`
	RemovedDuplicates    int                 `json:"removed_duplicates"`
	FixedInconsistencies int                 `json:"fixed_inconsistencies"`
	UnrepairableIssues   []UnrepairableIssue `json:"unrepairable_issues"`
	TotalRepairs         int                 `json:"total_repairs"`
	Status               IntegrityStatus     `json:"status"`
	DryRun               bool                `json:"dry_run"`
}

// Specific issue types
type OrphanedStockRating struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"company_id"`
	BrokerageID uuid.UUID `json:"brokerage_id"`
	Reason      string    `json:"reason"`
	Severity    string    `json:"severity"`
}

type InconsistentCompany struct {
	ID       uuid.UUID `json:"id"`
	Ticker   string    `json:"ticker"`
	Name     string    `json:"name"`
	Issue    string    `json:"issue"`
	Severity string    `json:"severity"`
}

type InconsistentBrokerage struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Issue    string    `json:"issue"`
	Severity string    `json:"severity"`
}

type InconsistentRating struct {
	ID       uuid.UUID `json:"id"`
	Issue    string    `json:"issue"`
	Severity string    `json:"severity"`
}

type DuplicateCompany struct {
	IDs      []uuid.UUID `json:"ids"`
	Ticker   string      `json:"ticker"`
	Count    int         `json:"count"`
	Severity string      `json:"severity"`
}

type DuplicateBrokerage struct {
	IDs      []uuid.UUID `json:"ids"`
	Name     string      `json:"name"`
	Count    int         `json:"count"`
	Severity string      `json:"severity"`
}

type DuplicateStockRating struct {
	IDs         []uuid.UUID `json:"ids"`
	CompanyID   uuid.UUID   `json:"company_id"`
	BrokerageID uuid.UUID   `json:"brokerage_id"`
	EventTime   time.Time   `json:"event_time"`
	Count       int         `json:"count"`
	Severity    string      `json:"severity"`
}

type InvalidCompany struct {
	ID       uuid.UUID `json:"id"`
	Ticker   string    `json:"ticker"`
	Name     string    `json:"name"`
	Rule     string    `json:"rule"`
	Severity string    `json:"severity"`
}

type InvalidBrokerage struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Rule     string    `json:"rule"`
	Severity string    `json:"severity"`
}

type InvalidStockRating struct {
	ID       uuid.UUID `json:"id"`
	Rule     string    `json:"rule"`
	Severity string    `json:"severity"`
}

type UnrepairableIssue struct {
	Type        string    `json:"type"`
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Reason      string    `json:"reason"`
}

// IntegrityValidationServiceImpl implements IntegrityValidationService
type IntegrityValidationServiceImpl struct {
	companyRepo     interfaces.CompanyRepository
	brokerageRepo   interfaces.BrokerageRepository
	stockRatingRepo interfaces.StockRatingRepository
	logger          logger.IntegrityLogger
	config          *ValidationConfig
}

// NewIntegrityValidationService creates a new integrity validation service
func NewIntegrityValidationService(
	companyRepo interfaces.CompanyRepository,
	brokerageRepo interfaces.BrokerageRepository,
	stockRatingRepo interfaces.StockRatingRepository,
	integrityLogger logger.IntegrityLogger,
	config *ValidationConfig,
) IntegrityValidationService {
	// Usar configuración por defecto si no se proporciona
	if config == nil {
		config = DefaultValidationConfig()
	}

	return &IntegrityValidationServiceImpl{
		companyRepo:     companyRepo,
		brokerageRepo:   brokerageRepo,
		stockRatingRepo: stockRatingRepo,
		logger:          integrityLogger,
		config:          config,
	}
}

// NewIntegrityValidationServiceWithDefaults creates a new integrity validation service with default configuration
// This maintains backward compatibility with existing code
func NewIntegrityValidationServiceWithDefaults(
	companyRepo interfaces.CompanyRepository,
	brokerageRepo interfaces.BrokerageRepository,
	stockRatingRepo interfaces.StockRatingRepository,
	integrityLogger logger.IntegrityLogger,
) IntegrityValidationService {
	return NewIntegrityValidationService(
		companyRepo,
		brokerageRepo,
		stockRatingRepo,
		integrityLogger,
		nil, // nil will use default configuration
	)
}

// ValidateFullIntegrity performs comprehensive integrity validation
func (s *IntegrityValidationServiceImpl) ValidateFullIntegrity(ctx context.Context) (*IntegrityReport, error) {
	startTime := time.Now()
	s.logger.LogValidationStart(ctx, "full_integrity", 0)

	report := &IntegrityReport{
		Timestamp: startTime,
		Summary:   make(map[string]interface{}),
	}

	// 1. Validate orphaned records
	orphanReport, err := s.ValidateOrphanedRecords(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to validate orphaned records", err,
			logger.String("operation", "orphan_validation"))
		return nil, fmt.Errorf("failed to validate orphaned records: %w", err)
	}
	report.OrphanReport = orphanReport

	// 2. Validate data consistency
	consistencyReport, err := s.ValidateDataConsistency(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to validate data consistency", err,
			logger.String("operation", "consistency_validation"))
		return nil, fmt.Errorf("failed to validate data consistency: %w", err)
	}
	report.ConsistencyReport = consistencyReport

	// 3. Validate duplicates
	duplicateReport, err := s.ValidateDuplicates(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to validate duplicates", err,
			logger.String("operation", "duplicate_validation"))
		return nil, fmt.Errorf("failed to validate duplicates: %w", err)
	}
	report.DuplicateReport = duplicateReport

	// 4. Validate business rules
	businessReport, err := s.ValidateBusinessRules(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to validate business rules", err,
			logger.String("operation", "business_rule_validation"))
		return nil, fmt.Errorf("failed to validate business rules: %w", err)
	}
	report.BusinessReport = businessReport

	// Calculate overall status and metrics
	report.Duration = time.Since(startTime)
	s.calculateOverallStatus(report)
	s.generateSummary(report)

	issuesFound := report.TotalIssues
	s.logger.LogValidationEnd(ctx, "full_integrity", issuesFound, report.Duration)

	return report, nil
}

// calculateOverallStatus determines the overall integrity status based on sub-reports
func (s *IntegrityValidationServiceImpl) calculateOverallStatus(report *IntegrityReport) {
	criticalIssues := 0
	warningIssues := 0

	// Count issues from each report
	if report.OrphanReport != nil {
		if report.OrphanReport.Status == IntegrityStatusCritical {
			criticalIssues += report.OrphanReport.TotalOrphans
		}
	}

	if report.ConsistencyReport != nil {
		for _, issue := range report.ConsistencyReport.InconsistentCompanies {
			if issue.Severity == "critical" {
				criticalIssues++
			} else {
				warningIssues++
			}
		}
		for _, issue := range report.ConsistencyReport.InconsistentBrokerages {
			if issue.Severity == "critical" {
				criticalIssues++
			} else {
				warningIssues++
			}
		}
		for _, issue := range report.ConsistencyReport.InconsistentRatings {
			if issue.Severity == "critical" {
				criticalIssues++
			} else {
				warningIssues++
			}
		}
	}

	if report.DuplicateReport != nil {
		criticalIssues += len(report.DuplicateReport.DuplicateCompanies)
		criticalIssues += len(report.DuplicateReport.DuplicateBrokerages)
		criticalIssues += len(report.DuplicateReport.DuplicateStockRatings)
	}

	if report.BusinessReport != nil {
		for _, issue := range report.BusinessReport.InvalidCompanies {
			if issue.Severity == "critical" {
				criticalIssues++
			} else {
				warningIssues++
			}
		}
		for _, issue := range report.BusinessReport.InvalidBrokerages {
			if issue.Severity == "critical" {
				criticalIssues++
			} else {
				warningIssues++
			}
		}
		for _, issue := range report.BusinessReport.InvalidStockRatings {
			if issue.Severity == "critical" {
				criticalIssues++
			} else {
				warningIssues++
			}
		}
	}

	report.CriticalIssues = criticalIssues
	report.WarningIssues = warningIssues
	report.TotalIssues = criticalIssues + warningIssues

	// Determine overall status
	if criticalIssues > 0 {
		report.OverallStatus = IntegrityStatusCritical
	} else if warningIssues > 0 {
		report.OverallStatus = IntegrityStatusWarning
	} else {
		report.OverallStatus = IntegrityStatusHealthy
	}
}

// generateSummary creates a summary of the integrity validation results
func (s *IntegrityValidationServiceImpl) generateSummary(report *IntegrityReport) {
	summary := make(map[string]interface{})

	// Overall metrics
	summary["total_issues"] = report.TotalIssues
	summary["critical_issues"] = report.CriticalIssues
	summary["warning_issues"] = report.WarningIssues
	summary["overall_status"] = string(report.OverallStatus)
	summary["validation_duration"] = report.Duration.String()

	// Detailed breakdown
	breakdown := make(map[string]interface{})

	if report.OrphanReport != nil {
		breakdown["orphaned_records"] = map[string]interface{}{
			"total":  report.OrphanReport.TotalOrphans,
			"status": string(report.OrphanReport.Status),
		}
	}

	if report.ConsistencyReport != nil {
		breakdown["consistency_issues"] = map[string]interface{}{
			"total":      report.ConsistencyReport.TotalInconsistencies,
			"companies":  len(report.ConsistencyReport.InconsistentCompanies),
			"brokerages": len(report.ConsistencyReport.InconsistentBrokerages),
			"ratings":    len(report.ConsistencyReport.InconsistentRatings),
			"status":     string(report.ConsistencyReport.Status),
		}
	}

	if report.DuplicateReport != nil {
		breakdown["duplicate_records"] = map[string]interface{}{
			"total":      report.DuplicateReport.TotalDuplicates,
			"companies":  len(report.DuplicateReport.DuplicateCompanies),
			"brokerages": len(report.DuplicateReport.DuplicateBrokerages),
			"ratings":    len(report.DuplicateReport.DuplicateStockRatings),
			"status":     string(report.DuplicateReport.Status),
		}
	}

	if report.BusinessReport != nil {
		breakdown["business_rule_violations"] = map[string]interface{}{
			"total":      report.BusinessReport.TotalViolations,
			"companies":  len(report.BusinessReport.InvalidCompanies),
			"brokerages": len(report.BusinessReport.InvalidBrokerages),
			"ratings":    len(report.BusinessReport.InvalidStockRatings),
			"status":     string(report.BusinessReport.Status),
		}
	}

	summary["breakdown"] = breakdown

	// Recommendations
	recommendations := make([]string, 0)

	if report.CriticalIssues > 0 {
		recommendations = append(recommendations, "Immediate action required - critical integrity issues found")
		if report.OrphanReport != nil && report.OrphanReport.TotalOrphans > 0 {
			recommendations = append(recommendations, "Remove orphaned stock rating records")
		}
		if report.DuplicateReport != nil && report.DuplicateReport.TotalDuplicates > 0 {
			recommendations = append(recommendations, "Resolve duplicate records")
		}
	}

	if report.WarningIssues > 0 {
		recommendations = append(recommendations, "Consider fixing minor consistency issues")
	}

	if report.TotalIssues == 0 {
		recommendations = append(recommendations, "Database integrity is healthy")
	} else {
		recommendations = append(recommendations, "Run automatic repair to fix minor issues")
	}

	summary["recommendations"] = recommendations

	report.Summary = summary
}
