package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// RepairMinorIssues attempts to fix minor integrity issues automatically
func (s *IntegrityValidationServiceImpl) RepairMinorIssues(ctx context.Context, dryRun bool) (*RepairReport, error) {
	log.Printf("🔧 Starting automatic repair of minor issues (dry_run: %t)...", dryRun)

	report := &RepairReport{
		UnrepairableIssues: make([]UnrepairableIssue, 0),
		DryRun:             dryRun,
	}

	// First, get a full integrity report to understand what needs fixing
	integrityReport, err := s.ValidateFullIntegrity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrity report for repair: %w", err)
	}

	// Repair orphaned records (by deletion - they're invalid anyway)
	if err := s.repairOrphanedRecords(ctx, integrityReport.OrphanReport, report, dryRun); err != nil {
		log.Printf("⚠️ Failed to repair orphaned records: %v", err)
	}

	// Repair duplicate records (keep the oldest, remove newer ones)
	if err := s.repairDuplicateRecords(ctx, integrityReport.DuplicateReport, report, dryRun); err != nil {
		log.Printf("⚠️ Failed to repair duplicate records: %v", err)
	}

	// Repair minor consistency issues
	if err := s.repairConsistencyIssues(ctx, integrityReport.ConsistencyReport, report, dryRun); err != nil {
		log.Printf("⚠️ Failed to repair consistency issues: %v", err)
	}

	// Calculate totals
	report.TotalRepairs = report.RepairedOrphans + report.RemovedDuplicates + report.FixedInconsistencies

	// Determine overall status
	if len(report.UnrepairableIssues) == 0 && report.TotalRepairs > 0 {
		report.Status = IntegrityStatusHealthy
		log.Printf("✅ Successfully repaired %d issues", report.TotalRepairs)
	} else if len(report.UnrepairableIssues) > 0 && report.TotalRepairs > 0 {
		report.Status = IntegrityStatusWarning
		log.Printf("⚠️ Repaired %d issues, but %d remain unrepairable",
			report.TotalRepairs, len(report.UnrepairableIssues))
	} else if len(report.UnrepairableIssues) > 0 {
		report.Status = IntegrityStatusCritical
		log.Printf("❌ Could not repair %d issues", len(report.UnrepairableIssues))
	} else {
		report.Status = IntegrityStatusHealthy
		log.Println("✅ No issues found to repair")
	}

	return report, nil
}

// repairOrphanedRecords removes orphaned stock rating records
func (s *IntegrityValidationServiceImpl) repairOrphanedRecords(ctx context.Context, orphanReport *OrphanReport, repairReport *RepairReport, dryRun bool) error {
	if orphanReport == nil || len(orphanReport.OrphanedStockRatings) == 0 {
		return nil
	}

	log.Printf("🔧 Repairing %d orphaned stock rating records...", len(orphanReport.OrphanedStockRatings))

	for _, orphan := range orphanReport.OrphanedStockRatings {
		if dryRun {
			log.Printf("🔍 DRY RUN: Would delete orphaned stock rating %s", orphan.ID)
			repairReport.RepairedOrphans++
		} else {
			// Attempt to delete the orphaned record
			if err := s.stockRatingRepo.Delete(ctx, orphan.ID); err != nil {
				log.Printf("❌ Failed to delete orphaned stock rating %s: %v", orphan.ID, err)
				repairReport.UnrepairableIssues = append(repairReport.UnrepairableIssues, UnrepairableIssue{
					Type:        "orphaned_stock_rating",
					ID:          orphan.ID,
					Description: fmt.Sprintf("Orphaned stock rating: %s", orphan.Reason),
					Reason:      fmt.Sprintf("Delete failed: %v", err),
				})
			} else {
				log.Printf("✅ Deleted orphaned stock rating %s", orphan.ID)
				repairReport.RepairedOrphans++
			}
		}
	}

	return nil
}

// repairDuplicateRecords removes duplicate records, keeping the oldest
func (s *IntegrityValidationServiceImpl) repairDuplicateRecords(ctx context.Context, duplicateReport *DuplicateReport, repairReport *RepairReport, dryRun bool) error {
	if duplicateReport == nil {
		return nil
	}

	// Repair duplicate companies
	for _, duplicate := range duplicateReport.DuplicateCompanies {
		if err := s.repairDuplicateCompanies(ctx, duplicate, repairReport, dryRun); err != nil {
			log.Printf("❌ Failed to repair duplicate companies for ticker %s: %v", duplicate.Ticker, err)
		}
	}

	// Repair duplicate brokerages
	for _, duplicate := range duplicateReport.DuplicateBrokerages {
		if err := s.repairDuplicateBrokerages(ctx, duplicate, repairReport, dryRun); err != nil {
			log.Printf("❌ Failed to repair duplicate brokerages for name %s: %v", duplicate.Name, err)
		}
	}

	// Repair duplicate stock ratings
	for _, duplicate := range duplicateReport.DuplicateStockRatings {
		if err := s.repairDuplicateStockRatings(ctx, duplicate, repairReport, dryRun); err != nil {
			log.Printf("❌ Failed to repair duplicate stock ratings: %v", err)
		}
	}

	return nil
}

// repairConsistencyIssues fixes minor data consistency problems
func (s *IntegrityValidationServiceImpl) repairConsistencyIssues(ctx context.Context, consistencyReport *ConsistencyReport, repairReport *RepairReport, dryRun bool) error {
	if consistencyReport == nil {
		return nil
	}

	// Fix company consistency issues
	for _, issue := range consistencyReport.InconsistentCompanies {
		if err := s.repairCompanyConsistency(ctx, issue, repairReport, dryRun); err != nil {
			log.Printf("❌ Failed to repair company consistency for %s: %v", issue.Ticker, err)
		}
	}

	// Fix brokerage consistency issues
	for _, issue := range consistencyReport.InconsistentBrokerages {
		if err := s.repairBrokerageConsistency(ctx, issue, repairReport, dryRun); err != nil {
			log.Printf("❌ Failed to repair brokerage consistency for %s: %v", issue.Name, err)
		}
	}

	return nil
}

// Helper methods for specific repairs

func (s *IntegrityValidationServiceImpl) repairDuplicateCompanies(ctx context.Context, duplicate DuplicateCompany, repairReport *RepairReport, dryRun bool) error {
	if len(duplicate.IDs) <= 1 {
		return nil
	}

	// Get all duplicate companies to find the oldest
	var oldestID = duplicate.IDs[0]
	var oldestCreatedAt = time.Now()

	for _, id := range duplicate.IDs {
		company, err := s.companyRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}

		if company.CreatedAt.Before(oldestCreatedAt) {
			oldestCreatedAt = company.CreatedAt
			oldestID = id
		}
	}

	// Delete all duplicates except the oldest
	for _, id := range duplicate.IDs {
		if id == oldestID {
			continue // Keep the oldest
		}

		if dryRun {
			log.Printf("🔍 DRY RUN: Would delete duplicate company %s (ticker: %s)", id, duplicate.Ticker)
			repairReport.RemovedDuplicates++
		} else {
			if err := s.companyRepo.Delete(ctx, id); err != nil {
				repairReport.UnrepairableIssues = append(repairReport.UnrepairableIssues, UnrepairableIssue{
					Type:        "duplicate_company",
					ID:          id,
					Description: fmt.Sprintf("Duplicate company with ticker %s", duplicate.Ticker),
					Reason:      fmt.Sprintf("Delete failed: %v", err),
				})
			} else {
				log.Printf("✅ Deleted duplicate company %s (ticker: %s)", id, duplicate.Ticker)
				repairReport.RemovedDuplicates++
			}
		}
	}

	return nil
}

func (s *IntegrityValidationServiceImpl) repairDuplicateBrokerages(ctx context.Context, duplicate DuplicateBrokerage, repairReport *RepairReport, dryRun bool) error {
	if len(duplicate.IDs) <= 1 {
		return nil
	}

	// Similar logic to companies - keep the oldest
	var oldestID = duplicate.IDs[0]
	var oldestCreatedAt = time.Now()

	for _, id := range duplicate.IDs {
		brokerage, err := s.brokerageRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}

		if brokerage.CreatedAt.Before(oldestCreatedAt) {
			oldestCreatedAt = brokerage.CreatedAt
			oldestID = id
		}
	}

	// Delete all duplicates except the oldest
	for _, id := range duplicate.IDs {
		if id == oldestID {
			continue
		}

		if dryRun {
			log.Printf("🔍 DRY RUN: Would delete duplicate brokerage %s (name: %s)", id, duplicate.Name)
			repairReport.RemovedDuplicates++
		} else {
			if err := s.brokerageRepo.Delete(ctx, id); err != nil {
				repairReport.UnrepairableIssues = append(repairReport.UnrepairableIssues, UnrepairableIssue{
					Type:        "duplicate_brokerage",
					ID:          id,
					Description: fmt.Sprintf("Duplicate brokerage with name %s", duplicate.Name),
					Reason:      fmt.Sprintf("Delete failed: %v", err),
				})
			} else {
				log.Printf("✅ Deleted duplicate brokerage %s (name: %s)", id, duplicate.Name)
				repairReport.RemovedDuplicates++
			}
		}
	}

	return nil
}

func (s *IntegrityValidationServiceImpl) repairDuplicateStockRatings(ctx context.Context, duplicate DuplicateStockRating, repairReport *RepairReport, dryRun bool) error {
	if len(duplicate.IDs) <= 1 {
		return nil
	}

	// Keep the first one, delete the rest
	for i, id := range duplicate.IDs {
		if i == 0 {
			continue // Keep the first
		}

		if dryRun {
			log.Printf("🔍 DRY RUN: Would delete duplicate stock rating %s", id)
			repairReport.RemovedDuplicates++
		} else {
			if err := s.stockRatingRepo.Delete(ctx, id); err != nil {
				repairReport.UnrepairableIssues = append(repairReport.UnrepairableIssues, UnrepairableIssue{
					Type:        "duplicate_stock_rating",
					ID:          id,
					Description: "Duplicate stock rating",
					Reason:      fmt.Sprintf("Delete failed: %v", err),
				})
			} else {
				log.Printf("✅ Deleted duplicate stock rating %s", id)
				repairReport.RemovedDuplicates++
			}
		}
	}

	return nil
}

func (s *IntegrityValidationServiceImpl) repairCompanyConsistency(ctx context.Context, issue InconsistentCompany, repairReport *RepairReport, dryRun bool) error {
	// Only repair minor formatting issues, not critical business rule violations
	if !strings.Contains(issue.Issue, "uppercase") && !strings.Contains(issue.Issue, "whitespace") {
		return nil // Skip non-repairable issues
	}

	company, err := s.companyRepo.GetByID(ctx, issue.ID)
	if err != nil {
		return err
	}

	fixed := false

	// Fix ticker case
	if strings.Contains(issue.Issue, "uppercase") {
		company.Ticker = strings.ToUpper(company.Ticker)
		fixed = true
	}

	// Fix name whitespace
	if strings.Contains(issue.Issue, "whitespace") {
		company.Name = strings.TrimSpace(company.Name)
		fixed = true
	}

	if fixed {
		if dryRun {
			log.Printf("🔍 DRY RUN: Would fix consistency issues for company %s", company.Ticker)
			repairReport.FixedInconsistencies++
		} else {
			if err := s.companyRepo.Update(ctx, company); err != nil {
				repairReport.UnrepairableIssues = append(repairReport.UnrepairableIssues, UnrepairableIssue{
					Type:        "inconsistent_company",
					ID:          company.ID,
					Description: fmt.Sprintf("Company consistency issue: %s", issue.Issue),
					Reason:      fmt.Sprintf("Update failed: %v", err),
				})
			} else {
				log.Printf("✅ Fixed consistency issues for company %s", company.Ticker)
				repairReport.FixedInconsistencies++
			}
		}
	}

	return nil
}

func (s *IntegrityValidationServiceImpl) repairBrokerageConsistency(ctx context.Context, issue InconsistentBrokerage, repairReport *RepairReport, dryRun bool) error {
	// Only repair whitespace issues
	if !strings.Contains(issue.Issue, "whitespace") {
		return nil
	}

	brokerage, err := s.brokerageRepo.GetByID(ctx, issue.ID)
	if err != nil {
		return err
	}

	// Fix name whitespace
	originalName := brokerage.Name
	brokerage.Name = strings.TrimSpace(brokerage.Name)

	if brokerage.Name != originalName {
		if dryRun {
			log.Printf("🔍 DRY RUN: Would fix consistency issues for brokerage %s", brokerage.Name)
			repairReport.FixedInconsistencies++
		} else {
			if err := s.brokerageRepo.Update(ctx, brokerage); err != nil {
				repairReport.UnrepairableIssues = append(repairReport.UnrepairableIssues, UnrepairableIssue{
					Type:        "inconsistent_brokerage",
					ID:          brokerage.ID,
					Description: fmt.Sprintf("Brokerage consistency issue: %s", issue.Issue),
					Reason:      fmt.Sprintf("Update failed: %v", err),
				})
			} else {
				log.Printf("✅ Fixed consistency issues for brokerage %s", brokerage.Name)
				repairReport.FixedInconsistencies++
			}
		}
	}

	return nil
}
