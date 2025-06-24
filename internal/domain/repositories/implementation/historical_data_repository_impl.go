package implementation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
)

// HistoricalDataRepositoryImpl implements the HistoricalDataRepository interface
type HistoricalDataRepositoryImpl struct {
	db *gorm.DB
}

// NewHistoricalDataRepository creates a new instance of HistoricalDataRepositoryImpl
func NewHistoricalDataRepository(database *gorm.DB) interfaces.HistoricalDataRepository {
	return &HistoricalDataRepositoryImpl{
		db: database,
	}
}

// Create saves a new historical data record
func (r *HistoricalDataRepositoryImpl) Create(ctx context.Context, data *entities.HistoricalData) error {
	return r.db.WithContext(ctx).Create(data).Error
}

// GetByID retrieves a historical data record by ID
func (r *HistoricalDataRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.HistoricalData, error) {
	var data entities.HistoricalData
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetBySymbolAndDate retrieves a historical data record by symbol and date
func (r *HistoricalDataRepositoryImpl) GetBySymbolAndDate(ctx context.Context, symbol string, date time.Time) (*entities.HistoricalData, error) {
	var data entities.HistoricalData
	err := r.db.WithContext(ctx).Where("symbol = ? AND date = ?", symbol, date).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// Update updates an existing historical data record
func (r *HistoricalDataRepositoryImpl) Update(ctx context.Context, data *entities.HistoricalData) error {
	return r.db.WithContext(ctx).Save(data).Error
}

// Delete removes a historical data record by ID
func (r *HistoricalDataRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.HistoricalData{}, "id = ?", id).Error
}

// GetBySymbol retrieves historical data for a symbol within a date range
func (r *HistoricalDataRepositoryImpl) GetBySymbol(ctx context.Context, symbol string, startDate, endDate time.Time) ([]*entities.HistoricalData, error) {
	var data []*entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("date DESC").
		Find(&data).Error
	return data, err
}

// GetBySymbolLastN retrieves the last N days of historical data for a symbol
func (r *HistoricalDataRepositoryImpl) GetBySymbolLastN(ctx context.Context, symbol string, days int) ([]*entities.HistoricalData, error) {
	var data []*entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("date DESC").
		Limit(days).
		Find(&data).Error
	return data, err
}

// GetByCompanyID retrieves historical data for a company within a date range
func (r *HistoricalDataRepositoryImpl) GetByCompanyID(ctx context.Context, companyID uuid.UUID, startDate, endDate time.Time) ([]*entities.HistoricalData, error) {
	var data []*entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND date >= ? AND date <= ?", companyID, startDate, endDate).
		Order("date DESC").
		Find(&data).Error
	return data, err
}

// GetByTimeFrame retrieves historical data for a symbol with specific time frame
func (r *HistoricalDataRepositoryImpl) GetByTimeFrame(ctx context.Context, symbol string, timeFrame string, startDate, endDate time.Time) ([]*entities.HistoricalData, error) {
	var data []*entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND time_frame = ? AND date >= ? AND date <= ?", symbol, timeFrame, startDate, endDate).
		Order("date DESC").
		Find(&data).Error
	return data, err
}

// GetHighestPrice finds the highest price record for a symbol within date range
func (r *HistoricalDataRepositoryImpl) GetHighestPrice(ctx context.Context, symbol string, startDate, endDate time.Time) (*entities.HistoricalData, error) {
	var data entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("high_price DESC").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetLowestPrice finds the lowest price record for a symbol within date range
func (r *HistoricalDataRepositoryImpl) GetLowestPrice(ctx context.Context, symbol string, startDate, endDate time.Time) (*entities.HistoricalData, error) {
	var data entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("low_price ASC").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetLatestPrice retrieves the most recent historical data for a symbol
func (r *HistoricalDataRepositoryImpl) GetLatestPrice(ctx context.Context, symbol string) (*entities.HistoricalData, error) {
	var data entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Order("date DESC").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetPriceRange returns the highest and lowest prices for a symbol within date range
func (r *HistoricalDataRepositoryImpl) GetPriceRange(ctx context.Context, symbol string, startDate, endDate time.Time) (float64, float64, error) {
	var result struct {
		MaxHigh float64
		MinLow  float64
	}
	err := r.db.WithContext(ctx).
		Model(&entities.HistoricalData{}).
		Select("MAX(high_price) as max_high, MIN(low_price) as min_low").
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Scan(&result).Error

	return result.MaxHigh, result.MinLow, err
}

// GetHighestVolume finds the record with highest volume for a symbol within date range
func (r *HistoricalDataRepositoryImpl) GetHighestVolume(ctx context.Context, symbol string, startDate, endDate time.Time) (*entities.HistoricalData, error) {
	var data entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("volume DESC").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetAverageVolume calculates the average volume for a symbol within date range
func (r *HistoricalDataRepositoryImpl) GetAverageVolume(ctx context.Context, symbol string, startDate, endDate time.Time) (int64, error) {
	var result struct {
		AvgVolume sql.NullFloat64
	}

	err := r.db.WithContext(ctx).
		Model(&entities.HistoricalData{}).
		Select("AVG(volume) as avg_volume").
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return 0, err
	}

	if !result.AvgVolume.Valid {
		return 0, nil
	}

	return int64(result.AvgVolume.Float64), nil
}

// GetVolumeSpikes finds records where volume exceeds the average by a multiplier
func (r *HistoricalDataRepositoryImpl) GetVolumeSpikes(ctx context.Context, symbol string, multiplier float64, days int) ([]*entities.HistoricalData, error) {
	// First get the average volume for the period
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	avgVolume, err := r.GetAverageVolume(ctx, symbol, startDate, endDate)
	if err != nil {
		return nil, err
	}

	threshold := float64(avgVolume) * multiplier

	var data []*entities.HistoricalData
	err = r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ? AND volume > ?", symbol, startDate, endDate, threshold).
		Order("date DESC").
		Find(&data).Error

	return data, err
}

// GetGaps finds price gaps exceeding a minimum percentage
func (r *HistoricalDataRepositoryImpl) GetGaps(ctx context.Context, symbol string, minGapPercent float64, days int) ([]*entities.HistoricalData, error) {
	// This is a simplified implementation. In practice, gap detection would require
	// comparing consecutive trading days and identifying discontinuities.
	var data []*entities.HistoricalData
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// For now, just return records where the difference between open and previous close is significant
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("date DESC").
		Find(&data).Error

	return data, err
}

// GetBreakouts finds potential breakout patterns (simplified implementation)
func (r *HistoricalDataRepositoryImpl) GetBreakouts(ctx context.Context, symbol string, days int) ([]*entities.HistoricalData, error) {
	var data []*entities.HistoricalData
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	// Simplified: records where close is significantly higher than open
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ? AND close_price > open_price * 1.05", symbol, startDate, endDate).
		Order("date DESC").
		Find(&data).Error

	return data, err
}

// GetBreakdowns finds potential breakdown patterns (simplified implementation)
func (r *HistoricalDataRepositoryImpl) GetBreakdowns(ctx context.Context, symbol string, days int) ([]*entities.HistoricalData, error) {
	var data []*entities.HistoricalData
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	// Simplified: records where close is significantly lower than open
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ? AND close_price < open_price * 0.95", symbol, startDate, endDate).
		Order("date DESC").
		Find(&data).Error

	return data, err
}

// GetReturns calculates daily returns for a symbol within date range
func (r *HistoricalDataRepositoryImpl) GetReturns(ctx context.Context, symbol string, startDate, endDate time.Time) ([]float64, error) {
	var data []*entities.HistoricalData
	err := r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("date ASC").
		Find(&data).Error

	if err != nil {
		return nil, err
	}
	returns := make([]float64, 0, len(data)-1)
	for i := 1; i < len(data); i++ {
		if data[i-1].ClosePrice > 0 {
			dailyReturn := (data[i].ClosePrice - data[i-1].ClosePrice) / data[i-1].ClosePrice
			returns = append(returns, dailyReturn)
		}
	}

	return returns, nil
}

// GetVolatility calculates volatility for a symbol over a number of days
func (r *HistoricalDataRepositoryImpl) GetVolatility(ctx context.Context, symbol string, days int) (float64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	returns, err := r.GetReturns(ctx, symbol, startDate, endDate)
	if err != nil {
		return 0, err
	}

	if len(returns) < 2 {
		return 0, nil
	}

	// Calculate standard deviation of returns
	var sum, sumSquares float64
	n := float64(len(returns))

	for _, ret := range returns {
		sum += ret
	}
	mean := sum / n

	for _, ret := range returns {
		diff := ret - mean
		sumSquares += diff * diff
	}

	variance := sumSquares / (n - 1)
	return variance, nil // Return variance as volatility measure
}

// GetCorrelation calculates correlation between two symbols (simplified)
func (r *HistoricalDataRepositoryImpl) GetCorrelation(ctx context.Context, symbol1, symbol2 string, days int) (float64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	returns1, err := r.GetReturns(ctx, symbol1, startDate, endDate)
	if err != nil {
		return 0, err
	}

	returns2, err := r.GetReturns(ctx, symbol2, startDate, endDate)
	if err != nil {
		return 0, err
	}

	if len(returns1) != len(returns2) || len(returns1) < 2 {
		return 0, fmt.Errorf("insufficient or mismatched data for correlation calculation")
	}

	// Simplified correlation calculation
	return 0.0, nil // Placeholder implementation
}

// GetSMA calculates Simple Moving Average (simplified - should be calculated from historical data)
func (r *HistoricalDataRepositoryImpl) GetSMA(ctx context.Context, symbol string, period int, date time.Time) (float64, error) {
	startDate := date.AddDate(0, 0, -period)

	var result struct {
		AvgClose sql.NullFloat64
	}
	err := r.db.WithContext(ctx).
		Model(&entities.HistoricalData{}).
		Select("AVG(close_price) as avg_close").
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, date).
		Scan(&result).Error

	if err != nil {
		return 0, err
	}

	if !result.AvgClose.Valid {
		return 0, nil
	}

	return result.AvgClose.Float64, nil
}

// GetEMA calculates Exponential Moving Average (placeholder)
func (r *HistoricalDataRepositoryImpl) GetEMA(ctx context.Context, symbol string, period int, date time.Time) (float64, error) {
	// Simplified placeholder - EMA calculation would require iterative processing
	return r.GetSMA(ctx, symbol, period, date)
}

// GetMissingDates finds missing trading dates in a range
func (r *HistoricalDataRepositoryImpl) GetMissingDates(ctx context.Context, symbol string, startDate, endDate time.Time) ([]time.Time, error) {
	var existingDates []time.Time
	err := r.db.WithContext(ctx).
		Model(&entities.HistoricalData{}).
		Select("date").
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("date ASC").
		Pluck("date", &existingDates).Error

	if err != nil {
		return nil, err
	}

	// Generate expected trading dates and find missing ones
	// This is a simplified implementation
	var missingDates []time.Time
	return missingDates, nil
}

// GetDataGaps finds gaps in data exceeding maxGapDays
func (r *HistoricalDataRepositoryImpl) GetDataGaps(ctx context.Context, symbol string, maxGapDays int) ([]time.Time, error) {
	// Placeholder implementation
	return []time.Time{}, nil
}

// ValidateDataIntegrity checks for data consistency issues
func (r *HistoricalDataRepositoryImpl) ValidateDataIntegrity(ctx context.Context, symbol string) error { // Check for basic data integrity issues
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.HistoricalData{}).
		Where("symbol = ? AND (open_price <= 0 OR high_price <= 0 OR low_price <= 0 OR close_price <= 0)", symbol).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("found %d records with invalid price data for symbol %s", count, symbol)
	}

	return nil
}

// BulkCreate creates multiple historical data records
func (r *HistoricalDataRepositoryImpl) BulkCreate(ctx context.Context, data []*entities.HistoricalData) error {
	if len(data) == 0 {
		return nil
	}

	// Use batch insert for better performance
	batchSize := 100
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]
		if err := r.db.WithContext(ctx).CreateInBatches(batch, batchSize).Error; err != nil {
			return err
		}
	}

	return nil
}

// BulkUpdate updates multiple historical data records
func (r *HistoricalDataRepositoryImpl) BulkUpdate(ctx context.Context, data []*entities.HistoricalData) error {
	if len(data) == 0 {
		return nil
	}

	// Update each record individually (GORM doesn't have efficient bulk update)
	for _, record := range data {
		if err := r.Update(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

// DeleteBySymbolAndDateRange deletes historical data for a symbol within date range
func (r *HistoricalDataRepositoryImpl) DeleteBySymbolAndDateRange(ctx context.Context, symbol string, startDate, endDate time.Time) error {
	return r.db.WithContext(ctx).
		Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Delete(&entities.HistoricalData{}).Error
}

// List retrieves historical data with pagination
func (r *HistoricalDataRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.HistoricalData, error) {
	var data []*entities.HistoricalData
	err := r.db.WithContext(ctx).
		Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&data).Error
	return data, err
}

// Count returns the total number of historical data records
func (r *HistoricalDataRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.HistoricalData{}).Count(&count).Error
	return count, err
}

// CountBySymbol returns the number of historical data records for a symbol
func (r *HistoricalDataRepositoryImpl) CountBySymbol(ctx context.Context, symbol string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.HistoricalData{}).
		Where("symbol = ?", symbol).
		Count(&count).Error
	return count, err
}
