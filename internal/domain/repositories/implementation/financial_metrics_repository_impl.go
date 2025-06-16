package implementation

import (
	"context"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FinancialMetricsRepositoryImpl implements the FinancialMetricsRepository interface
type FinancialMetricsRepositoryImpl struct {
	db *gorm.DB
}

// NewFinancialMetricsRepository creates a new instance of FinancialMetricsRepositoryImpl
func NewFinancialMetricsRepository(db *gorm.DB) interfaces.FinancialMetricsRepository {
	return &FinancialMetricsRepositoryImpl{
		db: db,
	}
}

// Create creates a new financial metrics record
func (r *FinancialMetricsRepositoryImpl) Create(ctx context.Context, metrics *entities.FinancialMetrics) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

// GetByID retrieves financial metrics by ID
func (r *FinancialMetricsRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.FinancialMetrics, error) {
	var metrics entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		First(&metrics, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}

// GetBySymbol retrieves financial metrics by symbol
func (r *FinancialMetricsRepositoryImpl) GetBySymbol(ctx context.Context, symbol string) (*entities.FinancialMetrics, error) {
	var metrics entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		First(&metrics, "symbol = ?", symbol).Error
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}

// GetByCompanyID retrieves financial metrics by company ID
func (r *FinancialMetricsRepositoryImpl) GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*entities.FinancialMetrics, error) {
	var metrics entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		First(&metrics, "company_id = ?", companyID).Error
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}

// Update updates an existing financial metrics record
func (r *FinancialMetricsRepositoryImpl) Update(ctx context.Context, metrics *entities.FinancialMetrics) error {
	return r.db.WithContext(ctx).Save(metrics).Error
}

// Delete soft deletes a financial metrics record
func (r *FinancialMetricsRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.FinancialMetrics{}, "id = ?", id).Error
}

// List retrieves a paginated list of financial metrics
func (r *FinancialMetricsRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Limit(limit).
		Offset(offset).
		Order("updated_at DESC").
		Find(&metrics).Error
	return metrics, err
}

// GetBySymbols retrieves financial metrics for multiple symbols
func (r *FinancialMetricsRepositoryImpl) GetBySymbols(ctx context.Context, symbols []string) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("symbol IN ?", symbols).
		Find(&metrics).Error
	return metrics, err
}

// GetBySector retrieves financial metrics by sector
func (r *FinancialMetricsRepositoryImpl) GetBySector(ctx context.Context, sector string) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Joins("JOIN companies ON companies.id = financial_metrics.company_id").
		Where("companies.sector = ?", sector).
		Find(&metrics).Error
	return metrics, err
}

// GetByIndustry retrieves financial metrics by industry
func (r *FinancialMetricsRepositoryImpl) GetByIndustry(ctx context.Context, industry string) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Joins("JOIN companies ON companies.id = financial_metrics.company_id").
		Where("companies.industry = ?", industry).
		Find(&metrics).Error
	return metrics, err
}

// GetByPERatio retrieves financial metrics within PE ratio range
func (r *FinancialMetricsRepositoryImpl) GetByPERatio(ctx context.Context, minPE, maxPE float64) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("pe_ratio BETWEEN ? AND ?", minPE, maxPE).
		Find(&metrics).Error
	return metrics, err
}

// GetByROE retrieves financial metrics with ROE above minimum
func (r *FinancialMetricsRepositoryImpl) GetByROE(ctx context.Context, minROE float64) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("roe >= ?", minROE).
		Order("roe DESC").
		Find(&metrics).Error
	return metrics, err
}

// GetByMarketCap retrieves financial metrics within market cap range
func (r *FinancialMetricsRepositoryImpl) GetByMarketCap(ctx context.Context, minCap, maxCap int64) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	query := r.db.WithContext(ctx).
		Preload("Company").
		Joins("JOIN companies ON companies.id = financial_metrics.company_id")

	if minCap > 0 {
		query = query.Where("companies.market_cap >= ?", minCap)
	}
	if maxCap > 0 {
		query = query.Where("companies.market_cap <= ?", maxCap)
	}

	err := query.Find(&metrics).Error
	return metrics, err
}

// GetByGrowthRate retrieves financial metrics with growth above minimum
func (r *FinancialMetricsRepositoryImpl) GetByGrowthRate(ctx context.Context, minGrowth float64) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("revenue_growth_ttm >= ? OR earnings_growth_ttm >= ?", minGrowth, minGrowth).
		Order("revenue_growth_ttm DESC").
		Find(&metrics).Error
	return metrics, err
}

// GetByDebtToEquity retrieves financial metrics with debt to equity below maximum
func (r *FinancialMetricsRepositoryImpl) GetByDebtToEquity(ctx context.Context, maxDebtToEquity float64) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("debt_to_equity <= ?", maxDebtToEquity).
		Order("debt_to_equity ASC").
		Find(&metrics).Error
	return metrics, err
}

// GetTopByROE retrieves top companies by ROE
func (r *FinancialMetricsRepositoryImpl) GetTopByROE(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("roe > 0").
		Order("roe DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetTopByGrowth retrieves top companies by growth rate
func (r *FinancialMetricsRepositoryImpl) GetTopByGrowth(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("earnings_growth_ttm > 0").
		Order("earnings_growth_ttm DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetValueStocks retrieves value stocks based on financial criteria
func (r *FinancialMetricsRepositoryImpl) GetValueStocks(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("pe_ratio > 0 AND pe_ratio < 15 AND price_to_book > 0 AND price_to_book < 1.5").
		Order("pe_ratio ASC, price_to_book ASC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetGrowthStocks retrieves growth stocks based on financial criteria
func (r *FinancialMetricsRepositoryImpl) GetGrowthStocks(ctx context.Context, limit int) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("revenue_growth_ttm > 10 AND earnings_growth_ttm > 15").
		Order("earnings_growth_ttm DESC, revenue_growth_ttm DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetDividendStocks retrieves dividend-paying stocks with minimum yield
func (r *FinancialMetricsRepositoryImpl) GetDividendStocks(ctx context.Context, minYield float64, limit int) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("dividend_yield >= ?", minYield).
		Order("dividend_yield DESC").
		Limit(limit).
		Find(&metrics).Error
	return metrics, err
}

// GetStaleData retrieves financial metrics that haven't been updated recently
func (r *FinancialMetricsRepositoryImpl) GetStaleData(ctx context.Context, olderThan time.Time) ([]*entities.FinancialMetrics, error) {
	var metrics []*entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("last_updated < ?", olderThan).
		Find(&metrics).Error
	return metrics, err
}

// GetLastUpdated retrieves the last update time for a symbol
func (r *FinancialMetricsRepositoryImpl) GetLastUpdated(ctx context.Context, symbol string) (time.Time, error) {
	var metrics entities.FinancialMetrics
	err := r.db.WithContext(ctx).
		Select("last_updated").
		First(&metrics, "symbol = ?", symbol).Error
	if err != nil {
		return time.Time{}, err
	}
	return metrics.LastUpdated, nil
}

// BulkCreate creates multiple financial metrics records
func (r *FinancialMetricsRepositoryImpl) BulkCreate(ctx context.Context, metrics []*entities.FinancialMetrics) error {
	return r.db.WithContext(ctx).CreateInBatches(metrics, 100).Error
}

// BulkUpdate updates multiple financial metrics records
func (r *FinancialMetricsRepositoryImpl) BulkUpdate(ctx context.Context, metrics []*entities.FinancialMetrics) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, metric := range metrics {
			if err := tx.Save(metric).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetSectorAverages calculates average financial metrics for a sector
func (r *FinancialMetricsRepositoryImpl) GetSectorAverages(ctx context.Context, sector string) (map[string]float64, error) {
	var result struct {
		AvgPE     float64 `json:"avg_pe"`
		AvgROE    float64 `json:"avg_roe"`
		AvgROA    float64 `json:"avg_roa"`
		AvgGrowth float64 `json:"avg_growth"`
	}

	err := r.db.WithContext(ctx).
		Model(&entities.FinancialMetrics{}).
		Joins("JOIN companies ON companies.id = financial_metrics.company_id").
		Where("companies.sector = ?", sector).
		Select(`
			AVG(CASE WHEN pe_ratio > 0 THEN pe_ratio END) as avg_pe,
			AVG(CASE WHEN roe > 0 THEN roe END) as avg_roe,
			AVG(CASE WHEN roa > 0 THEN roa END) as avg_roa,
			AVG(CASE WHEN earnings_growth_ttm > 0 THEN earnings_growth_ttm END) as avg_growth
		`).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"pe_ratio":            result.AvgPE,
		"roe":                 result.AvgROE,
		"roa":                 result.AvgROA,
		"earnings_growth_ttm": result.AvgGrowth,
	}, nil
}

// GetIndustryAverages calculates average financial metrics for an industry
func (r *FinancialMetricsRepositoryImpl) GetIndustryAverages(ctx context.Context, industry string) (map[string]float64, error) {
	var result struct {
		AvgPE     float64 `json:"avg_pe"`
		AvgROE    float64 `json:"avg_roe"`
		AvgROA    float64 `json:"avg_roa"`
		AvgGrowth float64 `json:"avg_growth"`
	}

	err := r.db.WithContext(ctx).
		Model(&entities.FinancialMetrics{}).
		Joins("JOIN companies ON companies.id = financial_metrics.company_id").
		Where("companies.industry = ?", industry).
		Select(`
			AVG(CASE WHEN pe_ratio > 0 THEN pe_ratio END) as avg_pe,
			AVG(CASE WHEN roe > 0 THEN roe END) as avg_roe,
			AVG(CASE WHEN roa > 0 THEN roa END) as avg_roa,
			AVG(CASE WHEN earnings_growth_ttm > 0 THEN earnings_growth_ttm END) as avg_growth
		`).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"pe_ratio":            result.AvgPE,
		"roe":                 result.AvgROE,
		"roa":                 result.AvgROA,
		"earnings_growth_ttm": result.AvgGrowth,
	}, nil
}

// Count returns the total number of financial metrics records
func (r *FinancialMetricsRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.FinancialMetrics{}).Count(&count).Error
	return count, err
}
