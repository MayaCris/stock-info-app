package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/google/uuid"
)

// ValidateOrphanedRecords checks for orphaned records across tables using optimized JOINs
func (s *IntegrityValidationServiceImpl) ValidateOrphanedRecords(ctx context.Context) (*OrphanReport, error) {
	s.logger.LogValidationStart(ctx, "orphaned_records", 0)

	report := &OrphanReport{
		OrphanedStockRatings: make([]OrphanedStockRating, 0),
	}
	// Use optimized query with JOINs to get orphaned stock ratings directly
	orphanedResults, err := s.stockRatingRepo.GetOrphanedStockRatingsWithReasons(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get orphaned stock ratings", err,
			logger.String("operation", "get_orphaned_ratings"))
		return nil, fmt.Errorf("failed to get orphaned stock ratings: %w", err)
	}

	s.logger.Info(ctx, "Found orphaned stock ratings using optimized query",
		logger.String("operation", "orphan_check"),
		logger.Int("orphaned_count", len(orphanedResults)))

	// Convert results to report format
	for _, orphanResult := range orphanedResults {
		orphanRecord := OrphanedStockRating{
			ID:          orphanResult.ID,
			CompanyID:   orphanResult.CompanyID,
			BrokerageID: orphanResult.BrokerageID,
			Reason:      orphanResult.Reason,
			Severity:    "critical",
		}
		report.OrphanedStockRatings = append(report.OrphanedStockRatings, orphanRecord)

		s.logger.LogOrphanDetected(ctx, "stock_rating", orphanResult.ID.String(),
			orphanResult.Reason, "critical")
	}

	report.TotalOrphans = len(report.OrphanedStockRatings)
	// Determine status
	if report.TotalOrphans == 0 {
		report.Status = IntegrityStatusHealthy
		s.logger.Info(ctx, "✅ No orphaned records found",
			logger.String("operation", "orphan_validation_complete"))
	} else {
		report.Status = IntegrityStatusCritical
		s.logger.Warn(ctx, "❌ Found orphaned stock rating records",
			logger.String("operation", "orphan_validation_complete"),
			logger.Int("orphaned_count", report.TotalOrphans))
	}

	s.logger.LogValidationEnd(ctx, "orphaned_records", report.TotalOrphans, time.Since(time.Now()))
	return report, nil
}

// ValidateDataConsistency checks for data consistency issues
func (s *IntegrityValidationServiceImpl) ValidateDataConsistency(ctx context.Context) (*ConsistencyReport, error) {
	s.logger.LogValidationStart(ctx, "data_consistency", 0)

	report := &ConsistencyReport{
		InconsistentCompanies:  make([]InconsistentCompany, 0),
		InconsistentBrokerages: make([]InconsistentBrokerage, 0),
		InconsistentRatings:    make([]InconsistentRating, 0),
	}

	// Validate company consistency
	if err := s.validateCompanyConsistency(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate company consistency: %w", err)
	}

	// Validate brokerage consistency
	if err := s.validateBrokerageConsistency(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate brokerage consistency: %w", err)
	}

	// Validate stock rating consistency
	if err := s.validateStockRatingConsistency(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate stock rating consistency: %w", err)
	}

	report.TotalInconsistencies = len(report.InconsistentCompanies) +
		len(report.InconsistentBrokerages) + len(report.InconsistentRatings)
	// Determine status
	if report.TotalInconsistencies == 0 {
		report.Status = IntegrityStatusHealthy
		log.Println("✅ No data consistency issues found")
	} else if report.TotalInconsistencies <= s.config.Thresholds.ConsistencyWarningLimit {
		report.Status = IntegrityStatusWarning
		log.Printf("⚠️ Found %d minor consistency issues", report.TotalInconsistencies)
	} else {
		report.Status = IntegrityStatusCritical
		log.Printf("❌ Found %d significant consistency issues", report.TotalInconsistencies)
	}

	return report, nil
}

// ValidateDuplicates checks for duplicate records
func (s *IntegrityValidationServiceImpl) ValidateDuplicates(ctx context.Context) (*DuplicateReport, error) {
	log.Println("🔍 Validating duplicate records...")

	report := &DuplicateReport{
		DuplicateCompanies:    make([]DuplicateCompany, 0),
		DuplicateBrokerages:   make([]DuplicateBrokerage, 0),
		DuplicateStockRatings: make([]DuplicateStockRating, 0),
	}

	// Check for duplicate companies (by ticker)
	if err := s.validateCompanyDuplicates(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate company duplicates: %w", err)
	}

	// Check for duplicate brokerages (by name)
	if err := s.validateBrokerageDuplicates(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate brokerage duplicates: %w", err)
	}

	// Check for duplicate stock ratings
	if err := s.validateStockRatingDuplicates(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate stock rating duplicates: %w", err)
	}

	report.TotalDuplicates = len(report.DuplicateCompanies) +
		len(report.DuplicateBrokerages) + len(report.DuplicateStockRatings)

	// Determine status
	if report.TotalDuplicates == 0 {
		report.Status = IntegrityStatusHealthy
		log.Println("✅ No duplicate records found")
	} else if report.TotalDuplicates <= s.config.Thresholds.DuplicatesWarningLimit {
		report.Status = IntegrityStatusWarning
		log.Printf("⚠️ Found %d minor duplicate issues", report.TotalDuplicates)
	} else {
		report.Status = IntegrityStatusCritical
		log.Printf("❌ Found %d significant duplicate issues", report.TotalDuplicates)
	}

	return report, nil
}

// ValidateBusinessRules checks business logic constraints
func (s *IntegrityValidationServiceImpl) ValidateBusinessRules(ctx context.Context) (*BusinessRuleReport, error) {
	log.Println("🔍 Validating business rules...")

	report := &BusinessRuleReport{
		InvalidCompanies:    make([]InvalidCompany, 0),
		InvalidBrokerages:   make([]InvalidBrokerage, 0),
		InvalidStockRatings: make([]InvalidStockRating, 0),
	}

	// Validate company business rules
	if err := s.validateCompanyBusinessRules(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate company business rules: %w", err)
	}

	// Validate brokerage business rules
	if err := s.validateBrokerageBusinessRules(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate brokerage business rules: %w", err)
	}

	// Validate stock rating business rules
	if err := s.validateStockRatingBusinessRules(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to validate stock rating business rules: %w", err)
	}

	report.TotalViolations = len(report.InvalidCompanies) +
		len(report.InvalidBrokerages) + len(report.InvalidStockRatings)
	// Determine status
	if report.TotalViolations == 0 {
		report.Status = IntegrityStatusHealthy
		log.Println("✅ All business rules validated successfully")
	} else if report.TotalViolations <= s.config.Thresholds.BusinessRulesWarningLimit {
		report.Status = IntegrityStatusWarning
		log.Printf("⚠️ Found %d minor business rule violations", report.TotalViolations)
	} else {
		report.Status = IntegrityStatusCritical
		log.Printf("❌ Found %d significant business rule violations", report.TotalViolations)
	}

	return report, nil
}

// Helper methods for specific validations

func (s *IntegrityValidationServiceImpl) validateCompanyConsistency(ctx context.Context, report *ConsistencyReport) error {
	companies, err := s.companyRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, company := range companies {
		issues := make([]string, 0)

		// Check if company is valid according to entity rules
		if !company.IsValid() {
			issues = append(issues, "Entity validation failed")
		}

		// Check ticker format (should be uppercase, 1-10 chars)
		if company.Ticker != strings.ToUpper(company.Ticker) {
			issues = append(issues, "Ticker should be uppercase")
		}

		// Check if name is properly trimmed
		if company.Name != strings.TrimSpace(company.Name) {
			issues = append(issues, "Name has extra whitespace")
		}

		// Check market cap consistency
		if company.MarketCap < 0 {
			issues = append(issues, "Market cap cannot be negative")
		}

		if len(issues) > 0 {
			report.InconsistentCompanies = append(report.InconsistentCompanies, InconsistentCompany{
				ID:       company.ID,
				Ticker:   company.Ticker,
				Name:     company.Name,
				Issue:    strings.Join(issues, "; "),
				Severity: "warning",
			})
		}
	}

	return nil
}

func (s *IntegrityValidationServiceImpl) validateBrokerageConsistency(ctx context.Context, report *ConsistencyReport) error {
	brokerages, err := s.brokerageRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, brokerage := range brokerages {
		issues := make([]string, 0)

		// Check if brokerage is valid according to entity rules
		if !brokerage.IsValid() {
			issues = append(issues, "Entity validation failed")
		}

		// Check if name is properly trimmed
		if brokerage.Name != strings.TrimSpace(brokerage.Name) {
			issues = append(issues, "Name has extra whitespace")
		}

		// Check minimum name length
		if len(strings.TrimSpace(brokerage.Name)) < 2 {
			issues = append(issues, "Name too short")
		}

		if len(issues) > 0 {
			report.InconsistentBrokerages = append(report.InconsistentBrokerages, InconsistentBrokerage{
				ID:       brokerage.ID,
				Name:     brokerage.Name,
				Issue:    strings.Join(issues, "; "),
				Severity: "warning",
			})
		}
	}

	return nil
}

func (s *IntegrityValidationServiceImpl) validateStockRatingConsistency(ctx context.Context, report *ConsistencyReport) error {
	ratings, err := s.stockRatingRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, rating := range ratings {
		issues := make([]string, 0)

		// Check if rating is valid according to entity rules
		if !rating.IsValid() {
			issues = append(issues, "Entity validation failed")
		}
		// Check event time (shouldn't be in the future)
		if !s.isWithinValidTimeRange(rating.EventTime, "future") {
			issues = append(issues, s.getTimeRangeViolationMessage("future"))
		}

		// Check if event time is too old for consistency validation
		if !s.isWithinValidTimeRange(rating.EventTime, "consistency") {
			issues = append(issues, s.getTimeRangeViolationMessage("consistency"))
		}

		if len(issues) > 0 {
			report.InconsistentRatings = append(report.InconsistentRatings, InconsistentRating{
				ID:       rating.ID,
				Issue:    strings.Join(issues, "; "),
				Severity: "warning",
			})
		}
	}

	return nil
}

// validateCompanyDuplicates checks for duplicate companies by ticker
func (s *IntegrityValidationServiceImpl) validateCompanyDuplicates(ctx context.Context, report *DuplicateReport) error {
	companies, err := s.companyRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	// Group companies by ticker
	tickerGroups := make(map[string][]uuid.UUID)
	for _, company := range companies {
		tickerGroups[company.Ticker] = append(tickerGroups[company.Ticker], company.ID)
	}

	// Find duplicates
	for ticker, ids := range tickerGroups {
		if len(ids) > 1 {
			report.DuplicateCompanies = append(report.DuplicateCompanies, DuplicateCompany{
				IDs:      ids,
				Ticker:   ticker,
				Count:    len(ids),
				Severity: "critical",
			})
		}
	}

	return nil
}

// validateBrokerageDuplicates checks for duplicate brokerages by name
func (s *IntegrityValidationServiceImpl) validateBrokerageDuplicates(ctx context.Context, report *DuplicateReport) error {
	brokerages, err := s.brokerageRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	// Group brokerages by name
	nameGroups := make(map[string][]uuid.UUID)
	for _, brokerage := range brokerages {
		nameGroups[brokerage.Name] = append(nameGroups[brokerage.Name], brokerage.ID)
	}

	// Find duplicates
	for name, ids := range nameGroups {
		if len(ids) > 1 {
			report.DuplicateBrokerages = append(report.DuplicateBrokerages, DuplicateBrokerage{
				IDs:      ids,
				Name:     name,
				Count:    len(ids),
				Severity: "critical",
			})
		}
	}

	return nil
}

// validateStockRatingDuplicates checks for duplicate stock ratings
func (s *IntegrityValidationServiceImpl) validateStockRatingDuplicates(ctx context.Context, report *DuplicateReport) error {
	ratings, err := s.stockRatingRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	// Group ratings by company_id + brokerage_id + event_time
	ratingGroups := make(map[string][]uuid.UUID)
	for _, rating := range ratings {
		key := fmt.Sprintf("%s_%s_%d", rating.CompanyID, rating.BrokerageID, rating.EventTime.Unix())
		ratingGroups[key] = append(ratingGroups[key], rating.ID)
	}

	// Find duplicates
	for key, ids := range ratingGroups {
		if len(ids) > 1 {
			// Parse the key to get company and brokerage IDs
			parts := strings.Split(key, "_")
			if len(parts) >= 3 {
				companyID, _ := uuid.Parse(parts[0])
				brokerageID, _ := uuid.Parse(parts[1])

				report.DuplicateStockRatings = append(report.DuplicateStockRatings, DuplicateStockRating{
					IDs:         ids,
					CompanyID:   companyID,
					BrokerageID: brokerageID,
					Count:       len(ids),
					Severity:    "critical",
				})
			}
		}
	}

	return nil
}

// validateCompanyBusinessRules validates company business rules
func (s *IntegrityValidationServiceImpl) validateCompanyBusinessRules(ctx context.Context, report *BusinessRuleReport) error {
	companies, err := s.companyRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, company := range companies {
		violations := make([]string, 0) // Rule: Ticker must be within configured length limits
		if !s.isValidLength(company.Ticker, "company", "ticker") {
			violations = append(violations, s.getLengthViolationMessage("company", "ticker"))
		}

		// Rule: Company name must be within configured length limits
		if !s.isValidLength(company.Name, "company", "name") {
			violations = append(violations, s.getLengthViolationMessage("company", "name"))
		}

		// Rule: Market cap cannot be negative
		if company.MarketCap < 0 {
			violations = append(violations, "Market cap cannot be negative")
		}
		// Rule: Active companies should have valid data
		if company.IsActive && (company.Ticker == "" || company.Name == "") {
			violations = append(violations, "Active companies must have ticker and name")
		}
		if len(violations) > 0 {
			severity := s.calculateSeverityByViolations(violations, "company")

			report.InvalidCompanies = append(report.InvalidCompanies, InvalidCompany{
				ID:       company.ID,
				Ticker:   company.Ticker,
				Name:     company.Name,
				Rule:     strings.Join(violations, "; "),
				Severity: severity,
			})
		}
	}

	return nil
}

// validateBrokerageBusinessRules validates brokerage business rules
func (s *IntegrityValidationServiceImpl) validateBrokerageBusinessRules(ctx context.Context, report *BusinessRuleReport) error {
	brokerages, err := s.brokerageRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, brokerage := range brokerages {
		violations := make([]string, 0) // Rule: Brokerage name must be within configured length limits
		if !s.isValidLength(brokerage.Name, "brokerage", "name") {
			violations = append(violations, s.getLengthViolationMessage("brokerage", "name"))
		}
		// Rule: Active brokerages should have valid names
		if brokerage.IsActive && strings.TrimSpace(brokerage.Name) == "" {
			violations = append(violations, "Active brokerages must have a valid name")
		}

		if len(violations) > 0 {
			severity := s.calculateSeverityByViolations(violations, "brokerage")

			report.InvalidBrokerages = append(report.InvalidBrokerages, InvalidBrokerage{
				ID:       brokerage.ID,
				Name:     brokerage.Name,
				Rule:     strings.Join(violations, "; "),
				Severity: severity,
			})
		}
	}

	return nil
}

// validateStockRatingBusinessRules validates stock rating business rules
func (s *IntegrityValidationServiceImpl) validateStockRatingBusinessRules(ctx context.Context, report *BusinessRuleReport) error {
	ratings, err := s.stockRatingRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, rating := range ratings {
		violations := make([]string, 0)
		// Rule: Event time cannot be in the future
		if !s.isWithinValidTimeRange(rating.EventTime, "future") {
			violations = append(violations, s.getTimeRangeViolationMessage("future"))
		}

		// Rule: Event time cannot be older than configured business limit
		if !s.isWithinValidTimeRange(rating.EventTime, "business") {
			violations = append(violations, s.getTimeRangeViolationMessage("business"))
		}

		// Rule: Action must not be empty
		if strings.TrimSpace(rating.Action) == "" {
			violations = append(violations, "Action cannot be empty")
		}

		// Rule: Company and Brokerage IDs must be valid UUIDs
		if rating.CompanyID == uuid.Nil {
			violations = append(violations, "Company ID cannot be nil")
		}

		if rating.BrokerageID == uuid.Nil {
			violations = append(violations, "Brokerage ID cannot be nil")
		}

		if len(violations) > 0 {
			severity := s.calculateSeverityByViolations(violations, "stock_rating")

			report.InvalidStockRatings = append(report.InvalidStockRatings, InvalidStockRating{
				ID:       rating.ID,
				Rule:     strings.Join(violations, "; "),
				Severity: severity,
			})
		}
	}

	return nil
}
