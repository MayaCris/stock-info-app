package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/repositories/implementation"
	"github.com/MayaCris/stock-info-app/internal/domain/services"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// IntegrityValidationOptions configura las opciones de validación de integridad
type IntegrityValidationOptions struct {
	GenerateReport bool   // Si generar un archivo de reporte
	ReportPath     string // Ruta donde guardar el reporte
	AutoRepair     bool   // Si intentar reparación automática
	DryRun         bool   // Solo simular reparaciones
	ShowDetails    bool   // Mostrar detalles completos
	OutputFormat   string // json, text
}

// DefaultIntegrityOptions devuelve opciones por defecto para validación
func DefaultIntegrityOptions() IntegrityValidationOptions {
	return IntegrityValidationOptions{
		GenerateReport: true,
		ReportPath:     "./integrity_report.json",
		AutoRepair:     false,
		DryRun:         true,
		ShowDetails:    true,
		OutputFormat:   "json",
	}
}

// QuickIntegrityCheck devuelve opciones para verificación rápida
func QuickIntegrityCheck() IntegrityValidationOptions {
	return IntegrityValidationOptions{
		GenerateReport: false,
		ReportPath:     "",
		AutoRepair:     false,
		DryRun:         false,
		ShowDetails:    false,
		OutputFormat:   "text",
	}
}

// FullIntegrityValidationWithRepair devuelve opciones para validación completa con reparación
func FullIntegrityValidationWithRepair() IntegrityValidationOptions {
	return IntegrityValidationOptions{
		GenerateReport: true,
		ReportPath:     fmt.Sprintf("./integrity_report_%s.json", time.Now().Format("20060102_150405")),
		AutoRepair:     true,
		DryRun:         false,
		ShowDetails:    true,
		OutputFormat:   "json",
	}
}

// RunDatabaseIntegrityValidation ejecuta validación completa de integridad de base de datos
func RunDatabaseIntegrityValidation(cfg *config.Config, options IntegrityValidationOptions) error {
	log.Println("🔍 Starting comprehensive database integrity validation...")

	// 1. Database connection
	db, err := cockroachdb.NewConnection(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// 2. Create repositories
	companyRepo := implementation.NewCompanyRepository(db.DB)
	brokerageRepo := implementation.NewBrokerageRepository(db.DB)
	stockRatingRepo := implementation.NewStockRatingRepository(db.DB)
	// 3. Create logger
	baseLogger, err := logger.InitializeGlobalLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	integrityLogger := logger.NewIntegrityLogger(baseLogger, &logger.LogConfig{})
	// 4. Create integrity validation service with default configuration
	integrityService := services.NewIntegrityValidationServiceWithDefaults(
		companyRepo,
		brokerageRepo,
		stockRatingRepo,
		integrityLogger,
	)

	// 5. Execute full integrity validation
	ctx := context.Background()
	integrityReport, err := integrityService.ValidateFullIntegrity(ctx)
	if err != nil {
		return fmt.Errorf("integrity validation failed: %w", err)
	}

	// 5. Show results
	showIntegrityResults(integrityReport, options)

	// 6. Auto-repair if requested
	var repairReport *services.RepairReport
	if options.AutoRepair {
		log.Println("🔧 Starting automatic repair...")
		repairReport, err = integrityService.RepairMinorIssues(ctx, options.DryRun)
		if err != nil {
			log.Printf("❌ Repair failed: %v", err)
		} else {
			showRepairResults(repairReport, options)

			// Re-validate after repair if actual changes were made
			if !options.DryRun && repairReport.TotalRepairs > 0 {
				log.Println("🔍 Re-validating after repair...")
				postRepairReport, err := integrityService.ValidateFullIntegrity(ctx)
				if err == nil {
					log.Println("\n📊 POST-REPAIR VALIDATION RESULTS:")
					showIntegrityResults(postRepairReport, options)
				}
			}
		}
	}

	// 7. Generate report file if requested
	if options.GenerateReport && options.ReportPath != "" {
		if err := generateIntegrityReportFile(integrityReport, repairReport, options); err != nil {
			log.Printf("⚠️ Failed to generate report file: %v", err)
		} else {
			log.Printf("📄 Integrity report saved to: %s", options.ReportPath)
		}
	}

	// 8. Return appropriate exit code based on status
	if integrityReport.OverallStatus == services.IntegrityStatusCritical {
		return fmt.Errorf("critical integrity issues found: %d critical issues", integrityReport.CriticalIssues)
	}

	log.Println("✅ Database integrity validation completed successfully")
	return nil
}

// showIntegrityResults muestra los resultados de validación de integridad
func showIntegrityResults(report *services.IntegrityReport, options IntegrityValidationOptions) {
	if options.OutputFormat == "json" && options.ShowDetails {
		// Show JSON format for detailed analysis
		jsonData, err := json.MarshalIndent(report, "", "  ")
		if err == nil {
			fmt.Println(string(jsonData))
		}
		return
	}
	// Text format output
	log.Println("\n" + strings.Repeat("=", 70))
	log.Println("🔍 DATABASE INTEGRITY VALIDATION SUMMARY")
	log.Println(strings.Repeat("=", 70))
	log.Printf("📊 Overall Status: %s", report.OverallStatus)
	log.Printf("⏱️  Validation Duration: %v", report.Duration)
	log.Printf("🚨 Total Issues: %d (Critical: %d, Warning: %d)",
		report.TotalIssues, report.CriticalIssues, report.WarningIssues)

	if options.ShowDetails {
		// Detailed breakdown
		if report.OrphanReport != nil {
			log.Printf("\n🔗 ORPHANED RECORDS:")
			log.Printf("   Total: %d (Status: %s)",
				report.OrphanReport.TotalOrphans, report.OrphanReport.Status)
			if len(report.OrphanReport.OrphanedStockRatings) > 0 {
				log.Printf("   Examples: %d orphaned stock ratings found",
					len(report.OrphanReport.OrphanedStockRatings))
			}
		}

		if report.ConsistencyReport != nil {
			log.Printf("\n⚠️  CONSISTENCY ISSUES:")
			log.Printf("   Total: %d (Status: %s)",
				report.ConsistencyReport.TotalInconsistencies, report.ConsistencyReport.Status)
			log.Printf("   Companies: %d, Brokerages: %d, Ratings: %d",
				len(report.ConsistencyReport.InconsistentCompanies),
				len(report.ConsistencyReport.InconsistentBrokerages),
				len(report.ConsistencyReport.InconsistentRatings))
		}

		if report.DuplicateReport != nil {
			log.Printf("\n🔄 DUPLICATE RECORDS:")
			log.Printf("   Total: %d (Status: %s)",
				report.DuplicateReport.TotalDuplicates, report.DuplicateReport.Status)
			log.Printf("   Companies: %d, Brokerages: %d, Ratings: %d",
				len(report.DuplicateReport.DuplicateCompanies),
				len(report.DuplicateReport.DuplicateBrokerages),
				len(report.DuplicateReport.DuplicateStockRatings))
		}

		if report.BusinessReport != nil {
			log.Printf("\n📋 BUSINESS RULE VIOLATIONS:")
			log.Printf("   Total: %d (Status: %s)",
				report.BusinessReport.TotalViolations, report.BusinessReport.Status)
			log.Printf("   Companies: %d, Brokerages: %d, Ratings: %d",
				len(report.BusinessReport.InvalidCompanies),
				len(report.BusinessReport.InvalidBrokerages),
				len(report.BusinessReport.InvalidStockRatings))
		}

		// Show recommendations
		if recommendations, ok := report.Summary["recommendations"].([]string); ok && len(recommendations) > 0 {
			log.Println("\n💡 RECOMMENDATIONS:")
			for i, rec := range recommendations {
				log.Printf("   %d. %s", i+1, rec)
			}
		}
	}

	log.Println(strings.Repeat("=", 70))
}

// showRepairResults muestra los resultados de reparación
func showRepairResults(report *services.RepairReport, options IntegrityValidationOptions) {
	log.Println("\n" + strings.Repeat("=", 70))
	log.Println("🔧 AUTOMATIC REPAIR SUMMARY")
	log.Println(strings.Repeat("=", 70))
	log.Printf("📊 Repair Status: %s", report.Status)
	log.Printf("🔧 Total Repairs: %d", report.TotalRepairs)
	log.Printf("🗑️  Orphans Removed: %d", report.RepairedOrphans)
	log.Printf("🔄 Duplicates Removed: %d", report.RemovedDuplicates)
	log.Printf("✅ Consistency Fixed: %d", report.FixedInconsistencies)
	log.Printf("❌ Unrepairable: %d", len(report.UnrepairableIssues))

	if report.DryRun {
		log.Println("🔍 DRY RUN: No actual changes were made")
	}

	if options.ShowDetails && len(report.UnrepairableIssues) > 0 {
		log.Println("\n❌ UNREPAIRABLE ISSUES:")
		for i, issue := range report.UnrepairableIssues {
			if i < 10 { // Show first 10 issues
				log.Printf("   %d. %s: %s", i+1, issue.Type, issue.Description)
			}
		}
		if len(report.UnrepairableIssues) > 10 {
			log.Printf("   ... and %d more issues", len(report.UnrepairableIssues)-10)
		}
	}

	log.Println(strings.Repeat("=", 70))
}

// generateIntegrityReportFile genera un archivo de reporte de integridad
func generateIntegrityReportFile(integrityReport *services.IntegrityReport, repairReport *services.RepairReport, options IntegrityValidationOptions) error {
	reportData := map[string]interface{}{
		"timestamp":         time.Now(),
		"validation_config": options,
		"integrity_report":  integrityReport,
	}

	if repairReport != nil {
		reportData["repair_report"] = repairReport
	}

	// Create report file
	file, err := os.Create(options.ReportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// Write JSON data
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(reportData); err != nil {
		return fmt.Errorf("failed to write report data: %w", err)
	}

	return nil
}
