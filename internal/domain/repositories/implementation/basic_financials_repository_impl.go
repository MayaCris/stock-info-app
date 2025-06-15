package implementation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// basicFinancialsRepositoryImpl implements the BasicFinancialsRepository interface using GORM
type basicFinancialsRepositoryImpl struct {
	db *gorm.DB
}

// NewBasicFinancialsRepository creates a new basic financials repository implementation
func NewBasicFinancialsRepository(db *gorm.DB) interfaces.BasicFinancialsRepository {
	return &basicFinancialsRepositoryImpl{
		db: db,
	}
}

// ========================================
// CREATE OPERATIONS
// ========================================

// Create creates new basic financials in the database
func (r *basicFinancialsRepositoryImpl) Create(ctx context.Context, financials *entities.BasicFinancials) error {
	if err := r.db.WithContext(ctx).Create(financials).Error; err != nil {
		return fmt.Errorf("failed to create basic financials: %w", err)
	}
	return nil
}

// ========================================
// READ OPERATIONS
// ========================================

// GetByID retrieves basic financials by its unique ID
func (r *basicFinancialsRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.BasicFinancials, error) {
	var financials entities.BasicFinancials
	if err := r.db.WithContext(ctx).First(&financials, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("basic financials not found with id %s", id.String())
		}
		return nil, fmt.Errorf("failed to get basic financials by id: %w", err)
	}
	return &financials, nil
}

// GetBySymbol retrieves the latest basic financials for a stock symbol
func (r *basicFinancialsRepositoryImpl) GetBySymbol(ctx context.Context, symbol string) (*entities.BasicFinancials, error) {
	var financials entities.BasicFinancials
	if err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("created_at DESC").
		First(&financials).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("basic financials not found for symbol %s", symbol)
		}
		return nil, fmt.Errorf("failed to get basic financials by symbol: %w", err)
	}
	return &financials, nil
}

// GetBySymbolAndPeriod retrieves basic financials for a specific symbol, period and fiscal year
func (r *basicFinancialsRepositoryImpl) GetBySymbolAndPeriod(ctx context.Context, symbol string, period string, fiscalYear int) (*entities.BasicFinancials, error) {
	var financials entities.BasicFinancials
	if err := r.db.WithContext(ctx).
		Where("symbol = ? AND period = ? AND fiscal_year = ?", symbol, period, fiscalYear).
		First(&financials).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("basic financials not found for symbol %s, period %s, year %d", symbol, period, fiscalYear)
		}
		return nil, fmt.Errorf("failed to get basic financials by symbol, period and year: %w", err)
	}
	return &financials, nil
}

// GetLatestBySymbol retrieves the latest basic financials for a symbol
func (r *basicFinancialsRepositoryImpl) GetLatestBySymbol(ctx context.Context, symbol string) (*entities.BasicFinancials, error) {
	return r.GetBySymbol(ctx, symbol)
}

// GetHistoricalBySymbol retrieves historical basic financials for a symbol
func (r *basicFinancialsRepositoryImpl) GetHistoricalBySymbol(ctx context.Context, symbol string, years int) ([]*entities.BasicFinancials, error) {
	var financialsList []*entities.BasicFinancials
	cutoffYear := time.Now().Year() - years

	if err := r.db.WithContext(ctx).
		Where("symbol = ? AND fiscal_year >= ?", symbol, cutoffYear).
		Order("fiscal_year DESC, period DESC").
		Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get historical basic financials: %w", err)
	}

	return financialsList, nil
}

// ========================================
// FINANCIAL ANALYSIS OPERATIONS
// ========================================

// GetByPERatioRange retrieves basic financials within a P/E ratio range
func (r *basicFinancialsRepositoryImpl) GetByPERatioRange(ctx context.Context, minPE, maxPE float64) ([]*entities.BasicFinancials, error) {
	var financialsList []*entities.BasicFinancials

	// Get latest financials for each symbol within PE range
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	if err := r.db.WithContext(ctx).
		Table("basic_financials").
		Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
		Where("pe_ratio BETWEEN ? AND ?", minPE, maxPE).
		Order("pe_ratio ASC").
		Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get basic financials by PE ratio range: %w", err)
	}

	return financialsList, nil
}

// GetByROERange retrieves basic financials within a ROE range
func (r *basicFinancialsRepositoryImpl) GetByROERange(ctx context.Context, minROE, maxROE float64) ([]*entities.BasicFinancials, error) {
	var financialsList []*entities.BasicFinancials

	// Get latest financials for each symbol within ROE range
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	if err := r.db.WithContext(ctx).
		Table("basic_financials").
		Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
		Where("roe BETWEEN ? AND ?", minROE, maxROE).
		Order("roe DESC").
		Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get basic financials by ROE range: %w", err)
	}

	return financialsList, nil
}

// GetByDebtToEquityRange retrieves basic financials within a debt-to-equity range
func (r *basicFinancialsRepositoryImpl) GetByDebtToEquityRange(ctx context.Context, minDE, maxDE float64) ([]*entities.BasicFinancials, error) {
	var financialsList []*entities.BasicFinancials

	// Get latest financials for each symbol within D/E range
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	if err := r.db.WithContext(ctx).
		Table("basic_financials").
		Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
		Where("debt_to_equity BETWEEN ? AND ?", minDE, maxDE).
		Order("debt_to_equity ASC").
		Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get basic financials by debt-to-equity range: %w", err)
	}

	return financialsList, nil
}

// ========================================
// SCREENING AND FILTERING OPERATIONS
// ========================================

// GetValueStocks retrieves value stocks based on criteria
func (r *basicFinancialsRepositoryImpl) GetValueStocks(ctx context.Context, maxPE float64, minROE float64, limit int) ([]*entities.BasicFinancials, error) {
	var financialsList []*entities.BasicFinancials

	// Get latest financials for each symbol
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	query := r.db.WithContext(ctx).
		Table("basic_financials").
		Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
		Where("pe_ratio <= ? AND pe_ratio > 0 AND roe >= ?", maxPE, minROE).
		Order("pe_ratio ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get value stocks: %w", err)
	}

	return financialsList, nil
}

// GetGrowthStocks retrieves growth stocks based on criteria
func (r *basicFinancialsRepositoryImpl) GetGrowthStocks(ctx context.Context, minRevenueGrowth, minEarningsGrowth float64, limit int) ([]*entities.BasicFinancials, error) {
	var financialsList []*entities.BasicFinancials

	// Get latest financials for each symbol
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	query := r.db.WithContext(ctx).
		Table("basic_financials").
		Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
		Where("revenue_growth >= ? AND earnings_growth >= ?", minRevenueGrowth, minEarningsGrowth).
		Order("earnings_growth DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get growth stocks: %w", err)
	}

	return financialsList, nil
}

// GetDividendStocks retrieves dividend-paying stocks
func (r *basicFinancialsRepositoryImpl) GetDividendStocks(ctx context.Context, minDividendYield float64, limit int) ([]*entities.BasicFinancials, error) {
	var financialsList []*entities.BasicFinancials

	// Get latest financials for each symbol
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	query := r.db.WithContext(ctx).
		Table("basic_financials").
		Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
		Where("dividend_yield >= ?", minDividendYield).
		Order("dividend_yield DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get dividend stocks: %w", err)
	}

	return financialsList, nil
}

// ========================================
// DATA FRESHNESS OPERATIONS
// ========================================

// GetStaleFinancials retrieves financial data older than maxAge
func (r *basicFinancialsRepositoryImpl) GetStaleFinancials(ctx context.Context, maxAge time.Duration) ([]*entities.BasicFinancials, error) {
	staleTime := time.Now().Add(-maxAge)
	var financialsList []*entities.BasicFinancials

	if err := r.db.WithContext(ctx).
		Where("updated_at < ?", staleTime).
		Order("updated_at ASC").
		Find(&financialsList).Error; err != nil {
		return nil, fmt.Errorf("failed to get stale basic financials: %w", err)
	}

	return financialsList, nil
}

// ========================================
// UPDATE OPERATIONS
// ========================================

// Update updates existing basic financials
func (r *basicFinancialsRepositoryImpl) Update(ctx context.Context, financials *entities.BasicFinancials) error {
	if err := r.db.WithContext(ctx).Save(financials).Error; err != nil {
		return fmt.Errorf("failed to update basic financials: %w", err)
	}
	return nil
}

// ========================================
// DELETE OPERATIONS
// ========================================

// Delete removes basic financials by ID
func (r *basicFinancialsRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.BasicFinancials{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete basic financials: %w", err)
	}
	return nil
}

// ========================================
// BULK OPERATIONS
// ========================================

// BulkCreate creates multiple basic financials
func (r *basicFinancialsRepositoryImpl) BulkCreate(ctx context.Context, financials []*entities.BasicFinancials) error {
	if len(financials) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(financials, 100).Error; err != nil {
		return fmt.Errorf("failed to create basic financials in bulk: %w", err)
	}
	return nil
}

// BulkUpdate updates multiple basic financials
func (r *basicFinancialsRepositoryImpl) BulkUpdate(ctx context.Context, financials []*entities.BasicFinancials) error {
	if len(financials) == 0 {
		return nil
	}

	for i := 0; i < len(financials); i += 100 {
		end := i + 100
		if end > len(financials) {
			end = len(financials)
		}

		batch := financials[i:end]
		for _, financial := range batch {
			if err := r.db.WithContext(ctx).Save(financial).Error; err != nil {
				return fmt.Errorf("failed to update basic financials in bulk: %w", err)
			}
		}
	}

	return nil
}

// UpsertBySymbol creates or updates basic financials for a symbol
func (r *basicFinancialsRepositoryImpl) UpsertBySymbol(ctx context.Context, financials *entities.BasicFinancials) error {
	var existing entities.BasicFinancials
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND period = ? AND fiscal_year = ?", financials.Symbol, financials.Period, financials.FiscalYear).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record
		if err := r.db.WithContext(ctx).Create(financials).Error; err != nil {
			return fmt.Errorf("failed to create basic financials during upsert: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing basic financials during upsert: %w", err)
	} else {
		// Update existing record
		financials.ID = existing.ID
		financials.CreatedAt = existing.CreatedAt
		if err := r.db.WithContext(ctx).Save(financials).Error; err != nil {
			return fmt.Errorf("failed to update basic financials during upsert: %w", err)
		}
	}

	return nil
}

// ========================================
// STATISTICS AND ANALYTICS OPERATIONS
// ========================================

// Count returns the total number of basic financials records
func (r *basicFinancialsRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.BasicFinancials{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count basic financials: %w", err)
	}
	return count, nil
}

// GetAverageMetrics returns average values for key financial metrics
func (r *basicFinancialsRepositoryImpl) GetAverageMetrics(ctx context.Context) (map[string]float64, error) {
	var result struct {
		AvgPE            float64
		AvgROE           float64
		AvgROA           float64
		AvgDebtToEquity  float64
		AvgCurrentRatio  float64
		AvgDividendYield float64
	}

	// Get latest financials for each symbol and calculate averages
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	if err := r.db.WithContext(ctx).
		Table("basic_financials").
		Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
		Select(`
			AVG(NULLIF(pe_ratio, 0)) as avg_pe,
			AVG(NULLIF(roe, 0)) as avg_roe,
			AVG(NULLIF(roa, 0)) as avg_roa,
			AVG(NULLIF(debt_to_equity, 0)) as avg_debt_to_equity,
			AVG(NULLIF(current_ratio, 0)) as avg_current_ratio,
			AVG(NULLIF(dividend_yield, 0)) as avg_dividend_yield
		`).
		Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to get average metrics: %w", err)
	}

	metrics := map[string]float64{
		"avg_pe_ratio":       result.AvgPE,
		"avg_roe":            result.AvgROE,
		"avg_roa":            result.AvgROA,
		"avg_debt_to_equity": result.AvgDebtToEquity,
		"avg_current_ratio":  result.AvgCurrentRatio,
		"avg_dividend_yield": result.AvgDividendYield,
	}

	return metrics, nil
}

// GetMetricDistribution returns the distribution of a specific metric
func (r *basicFinancialsRepositoryImpl) GetMetricDistribution(ctx context.Context, metric string) (map[string]int64, error) {
	var results []struct {
		Range string
		Count int64
	}

	// Get latest financials for each symbol
	subQuery := r.db.Select("symbol, MAX(created_at) as max_created").
		Group("symbol")

	var query *gorm.DB

	switch metric {
	case "pe_ratio":
		query = r.db.WithContext(ctx).
			Table("basic_financials").
			Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
			Select(`
				CASE 
					WHEN pe_ratio <= 10 THEN '0-10'
					WHEN pe_ratio <= 20 THEN '10-20'
					WHEN pe_ratio <= 30 THEN '20-30'
					WHEN pe_ratio <= 50 THEN '30-50'
					ELSE '50+'
				END as range,
				COUNT(*) as count
			`).
			Where("pe_ratio > 0").
			Group("range")
	case "roe":
		query = r.db.WithContext(ctx).
			Table("basic_financials").
			Joins("JOIN (?) as latest ON basic_financials.symbol = latest.symbol AND basic_financials.created_at = latest.max_created", subQuery).
			Select(`
				CASE 
					WHEN roe < 0 THEN 'Negative'
					WHEN roe <= 5 THEN '0-5%'
					WHEN roe <= 10 THEN '5-10%'
					WHEN roe <= 15 THEN '10-15%'
					WHEN roe <= 20 THEN '15-20%'
					ELSE '20%+'
				END as range,
				COUNT(*) as count
			`).
			Group("range")
	default:
		return nil, fmt.Errorf("unsupported metric: %s", metric)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get metric distribution for %s: %w", metric, err)
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.Range] = result.Count
	}

	return distribution, nil
}

// ========================================
// HEALTH CHECK OPERATIONS
// ========================================

// Health performs a health check on the repository
func (r *basicFinancialsRepositoryImpl) Health(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.BasicFinancials{}).Limit(1).Count(&count).Error; err != nil {
		return fmt.Errorf("basic financials repository health check failed: %w", err)
	}
	return nil
}
