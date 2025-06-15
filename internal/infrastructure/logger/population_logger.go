package logger

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PopulationFileLogger implementa PopulationLogger con logging específico para población
type PopulationFileLogger struct {
	Logger
	config *LogConfig
}

// NewPopulationLogger crea un nuevo logger específico para población
func NewPopulationLogger(baseLogger Logger, config *LogConfig) PopulationLogger {
	return &PopulationFileLogger{
		Logger: baseLogger.WithContext("POPULATION"),
		config: config,
	}
}

// LogPopulationStart registra el inicio del proceso de población
func (p *PopulationFileLogger) LogPopulationStart(ctx context.Context, config PopulationConfig) {
	if !p.config.ShouldLogProgress() {
		return
	}

	p.Info(ctx, "🚀 Starting database population process",
		String("operation", "population_start"),
		Int("batch_size", config.BatchSize),
		Int("max_pages", config.MaxPages),
		Duration("delay_between", config.DelayBetween),
		Bool("clear_first", config.ClearFirst),
		Bool("use_cache", config.UseCache),
		Bool("dry_run", config.DryRun),
		Bool("validate_after", config.ValidateAfter),
	)

	// Log de configuración detallada en debug
	p.Debug(ctx, "Population configuration details",
		String("operation", "config_detail"),
		Any("config", config),
	)
}

// LogPopulationEnd registra el fin del proceso de población
func (p *PopulationFileLogger) LogPopulationEnd(ctx context.Context, result PopulationResult, duration time.Duration) {
	successRate := float64(0)
	if result.TotalItems > 0 {
		successRate = float64(result.ProcessedItems) / float64(result.TotalItems) * 100
	}

	level := InfoLevel
	message := "✅ Database population completed successfully"

	if result.ErrorCount > 0 {
		if successRate < 50 {
			level = ErrorLevel
			message = "❌ Database population completed with significant errors"
		} else {
			level = WarnLevel
			message = "⚠️ Database population completed with some errors"
		}
	}

	fields := []Field{
		String("operation", "population_end"), Duration("total_duration", duration),
		Int("total_pages", result.TotalPages),
		Int("pages_requested", result.PagesRequested),
		Int("total_items", result.TotalItems),
		Int("processed_items", result.ProcessedItems),
		Int("skipped_items", result.SkippedItems),
		Int("error_count", result.ErrorCount),
		Int("companies_created", result.Companies),
		Int("brokerages_created", result.Brokerages),
		Int("stock_ratings_created", result.StockRatings),
		Float64("success_rate", successRate),
	}

	switch level {
	case InfoLevel:
		p.Info(ctx, message, fields...)
	case WarnLevel:
		p.Warn(ctx, message, fields...)
	case ErrorLevel:
		p.Error(ctx, message, nil, fields...)
	}

	// Log errores si existen
	if result.ErrorCount > 0 && len(result.Errors) > 0 {
		errorSummary := p.buildErrorSummary(result.Errors)
		p.Warn(ctx, "Population errors summary",
			String("operation", "error_summary"),
			Int("total_errors", len(result.Errors)),
			String("error_summary", errorSummary),
		)
	}

	// Log resumen detallado
	p.logPopulationSummary(ctx, result, duration)
}

// LogPageProcessing registra el procesamiento de páginas
func (p *PopulationFileLogger) LogPageProcessing(ctx context.Context, pageNum, totalPages int, itemCount int) {
	if !p.config.ShouldLogProgress() {
		return
	}

	progress := float64(pageNum) / float64(totalPages) * 100

	p.Info(ctx, fmt.Sprintf("📖 Processing page %d/%d (%.1f%%)", pageNum, totalPages, progress),
		String("operation", "page_processing"),
		Int("page_number", pageNum),
		Int("total_pages", totalPages),
		Int("item_count", itemCount),
		Float64("progress_percent", progress),
	)
}

// LogBatchProcessing registra el procesamiento de lotes
func (p *PopulationFileLogger) LogBatchProcessing(ctx context.Context, batchSize int, operation string) {
	if !p.config.ShouldLogProgress() {
		return
	}

	p.Debug(ctx, fmt.Sprintf("🔄 Processing batch: %s", operation),
		String("operation", "batch_processing"),
		String("batch_operation", operation),
		Int("batch_size", batchSize),
	)
}

// LogEntityCreated registra la creación de una entidad
func (p *PopulationFileLogger) LogEntityCreated(ctx context.Context, entityType string, identifier string, fields ...Field) {
	if !p.config.ShouldLogProgress() {
		return
	}

	logFields := []Field{
		String("operation", "entity_created"),
		String("entity_type", entityType),
		String("identifier", identifier),
	}
	logFields = append(logFields, fields...)

	p.Debug(ctx, fmt.Sprintf("✅ Created %s: %s", entityType, identifier), logFields...)
}

// LogEntitySkipped registra cuando una entidad es omitida
func (p *PopulationFileLogger) LogEntitySkipped(ctx context.Context, entityType string, identifier string, reason string) {
	if !p.config.ShouldLogProgress() {
		return
	}

	p.Debug(ctx, fmt.Sprintf("⏭️ Skipped %s: %s", entityType, identifier),
		String("operation", "entity_skipped"),
		String("entity_type", entityType),
		String("identifier", identifier),
		String("skip_reason", reason),
	)
}

// LogEntityError registra errores en el procesamiento de entidades
func (p *PopulationFileLogger) LogEntityError(ctx context.Context, entityType string, identifier string, err error, fields ...Field) {
	logFields := []Field{
		String("operation", "entity_error"),
		String("entity_type", entityType),
		String("identifier", identifier),
	}
	logFields = append(logFields, fields...)

	p.Error(ctx, fmt.Sprintf("❌ Failed to process %s: %s", entityType, identifier), err, logFields...)
}

// LogIntegrityValidation registra resultados de validación de integridad
func (p *PopulationFileLogger) LogIntegrityValidation(ctx context.Context, status string, issues int, duration time.Duration) {
	if !p.config.ShouldLogIntegrityCheck() {
		return
	}

	message := fmt.Sprintf("🔍 Integrity validation completed: %s", status)

	fields := []Field{
		String("operation", "integrity_validation"),
		String("validation_status", status),
		Int("issues_found", issues),
		Duration("validation_duration", duration),
	}

	if issues == 0 {
		p.Info(ctx, message, fields...)
	} else if issues < 10 {
		p.Warn(ctx, message, fields...)
	} else {
		p.Error(ctx, message, nil, fields...)
	}
}

// LogTransactionOperation registra operaciones transaccionales
func (p *PopulationFileLogger) LogTransactionOperation(ctx context.Context, operation string, retry int, success bool, duration time.Duration) {
	if !p.config.ShouldLogTransactions() {
		return
	}

	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	fields := []Field{
		String("operation", "transaction"),
		String("tx_operation", operation),
		Int("retry_count", retry),
		Bool("success", success),
		String("status", status),
		Duration("operation_duration", duration),
	}

	if success {
		if retry > 0 {
			p.Warn(ctx, fmt.Sprintf("🔄 Transaction %s succeeded after %d retries", operation, retry), fields...)
		} else {
			p.Debug(ctx, fmt.Sprintf("✅ Transaction %s completed", operation), fields...)
		}
	} else {
		p.Error(ctx, fmt.Sprintf("❌ Transaction %s failed after %d retries", operation, retry), nil, fields...)
	}
}

// buildErrorSummary construye un resumen de errores
func (p *PopulationFileLogger) buildErrorSummary(errors []string) string {
	if len(errors) == 0 {
		return "No errors"
	}

	// Agrupar errores similares
	errorTypes := make(map[string]int)
	for _, err := range errors {
		errorType := p.extractErrorType(err)
		errorTypes[errorType]++
	}

	var summary []string
	for errType, count := range errorTypes {
		summary = append(summary, fmt.Sprintf("%s: %d", errType, count))
	}

	return strings.Join(summary, ", ")
}

// extractErrorType extrae el tipo de error de un mensaje
func (p *PopulationFileLogger) extractErrorType(errorMsg string) string {
	errorMsg = strings.ToLower(errorMsg)

	if strings.Contains(errorMsg, "company") {
		return "company_error"
	}
	if strings.Contains(errorMsg, "brokerage") {
		return "brokerage_error"
	}
	if strings.Contains(errorMsg, "stock_rating") {
		return "stock_rating_error"
	}
	if strings.Contains(errorMsg, "fetch") || strings.Contains(errorMsg, "api") {
		return "api_error"
	}
	if strings.Contains(errorMsg, "database") || strings.Contains(errorMsg, "sql") {
		return "database_error"
	}
	if strings.Contains(errorMsg, "validation") {
		return "validation_error"
	}
	if strings.Contains(errorMsg, "cache") {
		return "cache_error"
	}

	return "unknown_error"
}

// logPopulationSummary registra un resumen detallado de la población
func (p *PopulationFileLogger) logPopulationSummary(ctx context.Context, result PopulationResult, duration time.Duration) {
	var summaryLines []string
	summaryLines = append(summaryLines, strings.Repeat("=", 60))
	summaryLines = append(summaryLines, "📊 DATABASE POPULATION SUMMARY")
	summaryLines = append(summaryLines, strings.Repeat("=", 60))
	summaryLines = append(summaryLines, fmt.Sprintf("📄 Pages with data processed: %d", result.TotalPages))
	summaryLines = append(summaryLines, fmt.Sprintf("📋 Total pages requested: %d", result.PagesRequested))
	summaryLines = append(summaryLines, fmt.Sprintf("📊 Total items fetched: %d", result.TotalItems))
	summaryLines = append(summaryLines, fmt.Sprintf("✅ Successfully processed: %d", result.ProcessedItems))
	summaryLines = append(summaryLines, fmt.Sprintf("⏭️  Skipped (already exists): %d", result.SkippedItems))
	summaryLines = append(summaryLines, fmt.Sprintf("❌ Errors: %d", result.ErrorCount))
	summaryLines = append(summaryLines, fmt.Sprintf("🏢 Companies created: %d", result.Companies))
	summaryLines = append(summaryLines, fmt.Sprintf("🏦 Brokerages created: %d", result.Brokerages))
	summaryLines = append(summaryLines, fmt.Sprintf("⭐ Stock ratings created: %d", result.StockRatings))
	summaryLines = append(summaryLines, fmt.Sprintf("⏱️  Total duration: %v", duration))

	if result.ProcessedItems > 0 {
		successRate := float64(result.ProcessedItems) / float64(result.TotalItems) * 100
		summaryLines = append(summaryLines, fmt.Sprintf("📈 Success rate: %.2f%%", successRate))
	}

	summaryLines = append(summaryLines, strings.Repeat("=", 60))

	summary := strings.Join(summaryLines, "\n")

	p.Info(ctx, "Population summary",
		String("operation", "population_summary"),
		String("summary", summary),
	)
}
