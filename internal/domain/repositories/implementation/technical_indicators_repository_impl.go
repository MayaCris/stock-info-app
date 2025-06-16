package implementation

import (
	"context"
	"time"

	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	"github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TechnicalIndicatorsRepositoryImpl implements the TechnicalIndicatorsRepository interface
type TechnicalIndicatorsRepositoryImpl struct {
	db *gorm.DB
}

// NewTechnicalIndicatorsRepository creates a new instance of TechnicalIndicatorsRepositoryImpl
func NewTechnicalIndicatorsRepository(db *gorm.DB) interfaces.TechnicalIndicatorsRepository {
	return &TechnicalIndicatorsRepositoryImpl{
		db: db,
	}
}

// Create creates a new technical indicators record
func (r *TechnicalIndicatorsRepositoryImpl) Create(ctx context.Context, indicators *entities.TechnicalIndicators) error {
	return r.db.WithContext(ctx).Create(indicators).Error
}

// GetByID retrieves technical indicators by ID
func (r *TechnicalIndicatorsRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.TechnicalIndicators, error) {
	var indicators entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		First(&indicators, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &indicators, nil
}

// GetBySymbol retrieves technical indicators by symbol
func (r *TechnicalIndicatorsRepositoryImpl) GetBySymbol(ctx context.Context, symbol string) (*entities.TechnicalIndicators, error) {
	var indicators entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		First(&indicators, "symbol = ?", symbol).Error
	if err != nil {
		return nil, err
	}
	return &indicators, nil
}

// GetByCompanyID retrieves technical indicators by company ID
func (r *TechnicalIndicatorsRepositoryImpl) GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*entities.TechnicalIndicators, error) {
	var indicators entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		First(&indicators, "company_id = ?", companyID).Error
	if err != nil {
		return nil, err
	}
	return &indicators, nil
}

// Update updates an existing technical indicators record
func (r *TechnicalIndicatorsRepositoryImpl) Update(ctx context.Context, indicators *entities.TechnicalIndicators) error {
	return r.db.WithContext(ctx).Save(indicators).Error
}

// Delete soft deletes a technical indicators record
func (r *TechnicalIndicatorsRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.TechnicalIndicators{}, "id = ?", id).Error
}

// List retrieves a paginated list of technical indicators
func (r *TechnicalIndicatorsRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Limit(limit).
		Offset(offset).
		Order("updated_at DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetBySymbols retrieves technical indicators for multiple symbols
func (r *TechnicalIndicatorsRepositoryImpl) GetBySymbols(ctx context.Context, symbols []string) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("symbol IN ?", symbols).
		Find(&indicators).Error
	return indicators, err
}

// GetByTimeFrame retrieves technical indicators by time frame
func (r *TechnicalIndicatorsRepositoryImpl) GetByTimeFrame(ctx context.Context, timeFrame string) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("time_frame = ?", timeFrame).
		Find(&indicators).Error
	return indicators, err
}

// GetBySignal retrieves technical indicators by signal type
func (r *TechnicalIndicatorsRepositoryImpl) GetBySignal(ctx context.Context, signal string) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("overall_signal = ?", signal).
		Find(&indicators).Error
	return indicators, err
}

// GetByRSI retrieves technical indicators within RSI range
func (r *TechnicalIndicatorsRepositoryImpl) GetByRSI(ctx context.Context, minRSI, maxRSI float64) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("rsi BETWEEN ? AND ?", minRSI, maxRSI).
		Find(&indicators).Error
	return indicators, err
}

// GetOverboughtStocks retrieves overbought stocks
func (r *TechnicalIndicatorsRepositoryImpl) GetOverboughtStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("rsi > 70 OR stoch_k > 80 OR williams_r > -20").
		Order("rsi DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetOversoldStocks retrieves oversold stocks
func (r *TechnicalIndicatorsRepositoryImpl) GetOversoldStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("rsi < 30 OR stoch_k < 20 OR williams_r < -80").
		Order("rsi ASC").
		Find(&indicators).Error
	return indicators, err
}

// GetBullishStocks retrieves stocks with bullish technical signals
func (r *TechnicalIndicatorsRepositoryImpl) GetBullishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where(`(rsi > 30 AND rsi < 70) AND 
			   (macd > macd_signal) AND 
			   (sma_20 > sma_50 AND sma_50 > sma_200) AND 
			   (adx > 25)`).
		Order("signal_strength DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetBearishStocks retrieves stocks with bearish technical signals
func (r *TechnicalIndicatorsRepositoryImpl) GetBearishStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where(`(rsi < 30 OR rsi > 70) AND 
			   (macd < macd_signal) AND 
			   (sma_20 < sma_50) AND 
			   (aroon_down > aroon_up)`).
		Order("signal_strength DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetMACDBullish retrieves stocks with bullish MACD signals
func (r *TechnicalIndicatorsRepositoryImpl) GetMACDBullish(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("macd > macd_signal AND macd_histogram > 0").
		Order("macd_histogram DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetMACDBearish retrieves stocks with bearish MACD signals
func (r *TechnicalIndicatorsRepositoryImpl) GetMACDBearish(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("macd < macd_signal AND macd_histogram < 0").
		Order("macd_histogram ASC").
		Find(&indicators).Error
	return indicators, err
}

// GetAboveMA retrieves stocks trading above moving average
func (r *TechnicalIndicatorsRepositoryImpl) GetAboveMA(ctx context.Context, period int) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	var query *gorm.DB

	switch period {
	case 20:
		query = r.db.WithContext(ctx).Where("sma_20 > 0") // Assuming current price comparison would be done at service level
	case 50:
		query = r.db.WithContext(ctx).Where("sma_50 > 0")
	case 200:
		query = r.db.WithContext(ctx).Where("sma_200 > 0")
	default:
		query = r.db.WithContext(ctx).Where("sma_20 > 0")
	}

	err := query.Preload("Company").Find(&indicators).Error
	return indicators, err
}

// GetBelowMA retrieves stocks trading below moving average
func (r *TechnicalIndicatorsRepositoryImpl) GetBelowMA(ctx context.Context, period int) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	var query *gorm.DB

	switch period {
	case 20:
		query = r.db.WithContext(ctx).Where("sma_20 > 0") // Service layer should implement price comparison
	case 50:
		query = r.db.WithContext(ctx).Where("sma_50 > 0")
	case 200:
		query = r.db.WithContext(ctx).Where("sma_200 > 0")
	default:
		query = r.db.WithContext(ctx).Where("sma_20 > 0")
	}

	err := query.Preload("Company").Find(&indicators).Error
	return indicators, err
}

// GetGoldenCross retrieves stocks with golden cross pattern (SMA20 > SMA50)
func (r *TechnicalIndicatorsRepositoryImpl) GetGoldenCross(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("sma_20 > sma_50 AND sma_50 > sma_200").
		Find(&indicators).Error
	return indicators, err
}

// GetDeathCross retrieves stocks with death cross pattern (SMA20 < SMA50)
func (r *TechnicalIndicatorsRepositoryImpl) GetDeathCross(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("sma_20 < sma_50 AND sma_50 < sma_200").
		Find(&indicators).Error
	return indicators, err
}

// GetHighVolumeStocks retrieves stocks with high volume
func (r *TechnicalIndicatorsRepositoryImpl) GetHighVolumeStocks(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("volume_ma_20 > 0 AND obv > 0").
		Order("volume_ma_20 DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetVolumeBreakout retrieves stocks with volume breakout
func (r *TechnicalIndicatorsRepositoryImpl) GetVolumeBreakout(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("volume_ma_20 > 0").
		Order("obv DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetHighVolatility retrieves stocks with high volatility
func (r *TechnicalIndicatorsRepositoryImpl) GetHighVolatility(ctx context.Context, threshold float64) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("atr > ? OR band_width > ?", threshold, threshold).
		Order("atr DESC").
		Find(&indicators).Error
	return indicators, err
}

// GetLowVolatility retrieves stocks with low volatility
func (r *TechnicalIndicatorsRepositoryImpl) GetLowVolatility(ctx context.Context, threshold float64) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("atr < ? AND band_width < ?", threshold, threshold).
		Order("atr ASC").
		Find(&indicators).Error
	return indicators, err
}

// GetBollingerBreakout retrieves stocks with Bollinger Band breakout
func (r *TechnicalIndicatorsRepositoryImpl) GetBollingerBreakout(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("bb_percent_b > 1.0 OR bb_percent_b < 0.0").
		Find(&indicators).Error
	return indicators, err
}

// GetBollingerSqueeze retrieves stocks with Bollinger Band squeeze
func (r *TechnicalIndicatorsRepositoryImpl) GetBollingerSqueeze(ctx context.Context) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("band_width < 0.1").
		Order("band_width ASC").
		Find(&indicators).Error
	return indicators, err
}

// GetTopByScore retrieves top stocks by technical score
func (r *TechnicalIndicatorsRepositoryImpl) GetTopByScore(ctx context.Context, limit int) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Order("signal_strength DESC").
		Limit(limit).
		Find(&indicators).Error
	return indicators, err
}

// GetStrongestSignals retrieves stocks with strongest signals
func (r *TechnicalIndicatorsRepositoryImpl) GetStrongestSignals(ctx context.Context, limit int) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("signal_strength > 70").
		Order("signal_strength DESC").
		Limit(limit).
		Find(&indicators).Error
	return indicators, err
}

// GetStaleData retrieves technical indicators that haven't been updated recently
func (r *TechnicalIndicatorsRepositoryImpl) GetStaleData(ctx context.Context, olderThan time.Time) ([]*entities.TechnicalIndicators, error) {
	var indicators []*entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("last_updated < ?", olderThan).
		Find(&indicators).Error
	return indicators, err
}

// GetLastUpdated retrieves the last update time for a symbol
func (r *TechnicalIndicatorsRepositoryImpl) GetLastUpdated(ctx context.Context, symbol string) (time.Time, error) {
	var indicators entities.TechnicalIndicators
	err := r.db.WithContext(ctx).
		Select("last_updated").
		First(&indicators, "symbol = ?", symbol).Error
	if err != nil {
		return time.Time{}, err
	}
	return indicators.LastUpdated, nil
}

// BulkCreate creates multiple technical indicators records
func (r *TechnicalIndicatorsRepositoryImpl) BulkCreate(ctx context.Context, indicators []*entities.TechnicalIndicators) error {
	return r.db.WithContext(ctx).CreateInBatches(indicators, 100).Error
}

// BulkUpdate updates multiple technical indicators records
func (r *TechnicalIndicatorsRepositoryImpl) BulkUpdate(ctx context.Context, indicators []*entities.TechnicalIndicators) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, indicator := range indicators {
			if err := tx.Save(indicator).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Count returns the total number of technical indicators records
func (r *TechnicalIndicatorsRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.TechnicalIndicators{}).Count(&count).Error
	return count, err
}
