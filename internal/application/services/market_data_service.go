package services

import (
	"context"
	"strconv"
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/alphavantage"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/finnhub"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// marketDataService implements MarketDataService interface
type marketDataService struct {
	// Repositories
	marketDataRepo      repoInterfaces.MarketDataRepository
	companyProfileRepo  repoInterfaces.CompanyProfileRepository
	newsRepo            repoInterfaces.NewsRepository
	basicFinancialsRepo repoInterfaces.BasicFinancialsRepository
	companyRepo         repoInterfaces.CompanyRepository
	// External API clients
	finnhubClient       *finnhub.Client
	finnhubAdapter      *finnhub.Adapter
	alphavantageClient  *alphavantage.Client
	alphavantageAdapter *alphavantage.Adapter

	// Logger
	logger logger.Logger
}

// MarketDataServiceConfig represents configuration for market data service
type MarketDataServiceConfig struct {
	MarketDataRepo      repoInterfaces.MarketDataRepository
	CompanyProfileRepo  repoInterfaces.CompanyProfileRepository
	NewsRepo            repoInterfaces.NewsRepository
	BasicFinancialsRepo repoInterfaces.BasicFinancialsRepository
	CompanyRepo         repoInterfaces.CompanyRepository
	FinnhubClient       *finnhub.Client
	FinnhubAdapter      *finnhub.Adapter
	AlphaVantageClient  *alphavantage.Client
	AlphaVantageAdapter *alphavantage.Adapter
	Logger              logger.Logger
}

// NewMarketDataService creates a new market data service
func NewMarketDataService(config MarketDataServiceConfig) interfaces.MarketDataService {
	return &marketDataService{
		marketDataRepo:      config.MarketDataRepo,
		companyProfileRepo:  config.CompanyProfileRepo,
		newsRepo:            config.NewsRepo,
		basicFinancialsRepo: config.BasicFinancialsRepo,
		companyRepo:         config.CompanyRepo,
		finnhubClient:       config.FinnhubClient,
		finnhubAdapter:      config.FinnhubAdapter,
		alphavantageClient:  config.AlphaVantageClient,
		alphavantageAdapter: config.AlphaVantageAdapter,
		logger:              config.Logger,
	}
}

// GetRealTimeQuote gets real-time quote for a symbol
func (s *marketDataService) GetRealTimeQuote(ctx context.Context, symbol string) (*response.MarketDataResponse, error) {
	// First, try to get from cache/database (recent data)
	existingData, err := s.marketDataRepo.GetBySymbol(ctx, symbol)
	if err == nil && !existingData.IsStale(5*time.Minute) {
		s.logger.Debug(ctx, "Returning cached market data",
			logger.String("symbol", symbol),
		)
		return s.convertToMarketDataResponse(existingData), nil
	}

	// Get company info to link market data
	company, err := s.companyRepo.GetByTicker(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Company not found for symbol", err,
			logger.String("symbol", symbol),
		)
		return nil, response.NotFound("Company with symbol " + symbol)
	}

	// Fetch fresh data from Finnhub
	quote, err := s.finnhubClient.GetRealTimeQuote(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch real-time quote from Finnhub", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to fetch real-time data")
	}

	// Convert to domain entity
	marketData, err := s.finnhubAdapter.QuoteToMarketData(ctx, quote, symbol, company.ID)
	if err != nil {
		s.logger.Error(ctx, "Failed to convert quote to market data", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to process market data")
	}

	// Validate data
	if err := s.finnhubAdapter.ValidateMarketData(marketData); err != nil {
		s.logger.Error(ctx, "Invalid market data", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Invalid market data")
	}

	// Save to database
	if err := s.marketDataRepo.UpsertBySymbol(ctx, marketData); err != nil {
		s.logger.Error(ctx, "Failed to save market data", err,
			logger.String("symbol", symbol),
		)
		// Don't return error here, we can still return the data
	}

	s.logger.Info(ctx, "Successfully retrieved and saved real-time quote",
		logger.String("symbol", symbol),
		logger.Float64("price", marketData.CurrentPrice),
	)

	return s.convertToMarketDataResponse(marketData), nil
}

// GetCompanyProfile gets detailed company profile
func (s *marketDataService) GetCompanyProfile(ctx context.Context, symbol string) (*response.CompanyProfileResponse, error) {
	// Try to get from database first
	existingProfile, err := s.companyProfileRepo.GetBySymbol(ctx, symbol)
	if err == nil && time.Since(existingProfile.LastUpdated).Hours() < 24 {
		s.logger.Debug(ctx, "Returning cached company profile",
			logger.String("symbol", symbol),
		)
		return s.convertToCompanyProfileResponse(existingProfile), nil
	}

	// Fetch fresh data from Finnhub
	profile, err := s.finnhubClient.GetCompanyProfile(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch company profile from Finnhub", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to fetch company profile")
	}

	// Convert to domain entity
	companyProfile, err := s.finnhubAdapter.ProfileToCompanyProfile(ctx, profile)
	if err != nil {
		s.logger.Error(ctx, "Failed to convert profile to company profile", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to process company profile")
	}

	// Validate data
	if err := s.finnhubAdapter.ValidateCompanyProfile(companyProfile); err != nil {
		s.logger.Error(ctx, "Invalid company profile", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Invalid company profile")
	}

	// Save to database
	if err := s.companyProfileRepo.UpsertBySymbol(ctx, companyProfile); err != nil {
		s.logger.Error(ctx, "Failed to save company profile", err,
			logger.String("symbol", symbol),
		)
		// Don't return error here, we can still return the data
	}

	s.logger.Info(ctx, "Successfully retrieved and saved company profile",
		logger.String("symbol", symbol),
		logger.String("company_name", companyProfile.Name),
	)

	return s.convertToCompanyProfileResponse(companyProfile), nil
}

// GetCompanyNews gets recent news for a company
func (s *marketDataService) GetCompanyNews(ctx context.Context, symbol string, days int) ([]*response.NewsResponse, error) {
	if days <= 0 {
		days = 7 // Default to 7 days
	}

	// Calculate date range
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	// Fetch news from Finnhub
	news, err := s.finnhubClient.GetCompanyNews(ctx, symbol, from, to)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch company news from Finnhub", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to fetch company news")
	}

	// Convert to domain entities
	newsItems, err := s.finnhubAdapter.NewsToNewsItems(ctx, news, symbol)
	if err != nil {
		s.logger.Error(ctx, "Failed to convert news to news items", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to process news data")
	}

	// Save news items to database
	if len(newsItems) > 0 {
		if err := s.newsRepo.BulkCreate(ctx, newsItems); err != nil {
			s.logger.Error(ctx, "Failed to save news items", err,
				logger.String("symbol", symbol),
			)
			// Don't return error here, we can still return the data
		}
	}

	s.logger.Info(ctx, "Successfully retrieved and saved company news",
		logger.String("symbol", symbol),
		logger.Int("news_count", len(newsItems)),
	)

	// Convert to response DTOs
	newsResponses := make([]*response.NewsResponse, len(newsItems))
	for i, newsItem := range newsItems {
		newsResponses[i] = s.convertToNewsResponse(newsItem)
	}

	return newsResponses, nil
}

// GetBasicFinancials gets basic financial metrics for a company
func (s *marketDataService) GetBasicFinancials(ctx context.Context, symbol string) (*response.BasicFinancialsResponse, error) {
	// Try to get from database first
	existingFinancials, err := s.basicFinancialsRepo.GetLatestBySymbol(ctx, symbol)
	if err == nil && time.Since(existingFinancials.LastUpdated).Hours() < 24 {
		s.logger.Debug(ctx, "Returning cached basic financials",
			logger.String("symbol", symbol),
		)
		return s.convertToBasicFinancialsResponse(existingFinancials), nil
	}

	// Fetch fresh data from Finnhub
	financials, err := s.finnhubClient.GetBasicFinancials(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch basic financials from Finnhub", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to fetch financial data")
	}

	// Convert to domain entity
	basicFinancials, err := s.finnhubAdapter.FinancialsToBasicFinancials(ctx, financials)
	if err != nil {
		s.logger.Error(ctx, "Failed to convert financials to basic financials", err,
			logger.String("symbol", symbol),
		)
		return nil, response.InternalServerError("Failed to process financial data")
	}

	// Save to database
	if err := s.basicFinancialsRepo.UpsertBySymbol(ctx, basicFinancials); err != nil {
		s.logger.Error(ctx, "Failed to save basic financials", err,
			logger.String("symbol", symbol),
		)
		// Don't return error here, we can still return the data
	}

	s.logger.Info(ctx, "Successfully retrieved and saved basic financials",
		logger.String("symbol", symbol),
	)

	return s.convertToBasicFinancialsResponse(basicFinancials), nil
}

// GetMarketOverview gets general market overview
func (s *marketDataService) GetMarketOverview(ctx context.Context) (*response.MarketOverviewResponse, error) {
	// Get recent market data
	recentData, err := s.marketDataRepo.GetLatest(ctx, 100)
	if err != nil {
		s.logger.Error(ctx, "Failed to get recent market data", err)
		return nil, response.InternalServerError("Failed to get market overview")
	}

	// Calculate market statistics
	var totalVolume int64
	var totalGainers, totalLosers int
	var avgPriceChange float64
	var priceChangeSum float64

	for _, data := range recentData {
		totalVolume += data.Volume
		priceChangeSum += data.PriceChangePerc

		if data.PriceChange > 0 {
			totalGainers++
		} else if data.PriceChange < 0 {
			totalLosers++
		}
	}

	if len(recentData) > 0 {
		avgPriceChange = priceChangeSum / float64(len(recentData))
	}

	overview := &response.MarketOverviewResponse{
		TotalStocks:    len(recentData),
		TotalGainers:   totalGainers,
		TotalLosers:    totalLosers,
		AvgPriceChange: avgPriceChange,
		TotalVolume:    totalVolume,
		LastUpdated:    time.Now(),
	}

	s.logger.Info(ctx, "Successfully generated market overview",
		logger.Int("total_stocks", overview.TotalStocks),
		logger.Int("gainers", overview.TotalGainers),
		logger.Int("losers", overview.TotalLosers),
	)

	return overview, nil
}

// GetHistoricalData gets historical price data from Alpha Vantage
func (s *marketDataService) GetHistoricalData(ctx context.Context, symbol, period, outputSize string) (*response.HistoricalDataResponse, error) {
	s.logger.Info(ctx, "Fetching historical data from Alpha Vantage",
		logger.String("symbol", symbol),
		logger.String("period", period),
		logger.String("output_size", outputSize))

	var alphaVantageResp interface{}
	var err error

	switch period {
	case "daily":
		alphaVantageResp, err = s.alphavantageClient.GetTimeSeriesDaily(ctx, symbol, outputSize)
	case "weekly":
		alphaVantageResp, err = s.alphavantageClient.GetTimeSeriesWeekly(ctx, symbol)
	case "monthly":
		alphaVantageResp, err = s.alphavantageClient.GetTimeSeriesMonthly(ctx, symbol)
	default:
		return nil, response.BadRequest("Invalid period. Supported: daily, weekly, monthly")
	}
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch historical data from Alpha Vantage", err,
			logger.String("symbol", symbol),
			logger.String("period", period))
		return nil, response.InternalServerError("Failed to fetch historical data")
	}

	// Use the response data (placeholder to avoid unused variable error)
	_ = alphaVantageResp

	// Convert to our response format using adapter
	// For now, create a simple response with the raw data
	historicalData := &response.HistoricalDataResponse{
		Success: true,
		Message: "Historical data retrieved successfully",
		Data: &response.HistoricalDataPayload{
			Symbol:      symbol,
			Period:      period,
			OutputSize:  outputSize,
			DataSource:  "alphavantage",
			LastUpdated: time.Now(),
			// Note: Full conversion would need implementation of TimeSeriesDataToResponse method
			// For now, endpoint will return metadata only
		},
	}

	return historicalData, nil
}

// GetTechnicalIndicators gets technical indicators from Alpha Vantage
func (s *marketDataService) GetTechnicalIndicators(ctx context.Context, symbol, indicator, interval, timePeriod string) (*response.TechnicalIndicatorsResponse, error) {
	s.logger.Info(ctx, "Fetching technical indicators from Alpha Vantage",
		logger.String("symbol", symbol),
		logger.String("indicator", indicator),
		logger.String("interval", interval))

	var alphaVantageResp interface{}
	var err error

	switch indicator {
	case "RSI":
		alphaVantageResp, err = s.alphavantageClient.GetRSI(ctx, symbol, interval, timePeriod, "close")
	case "MACD":
		alphaVantageResp, err = s.alphavantageClient.GetMACD(ctx, symbol, interval, "12", "26", "9", "close")
	case "SMA":
		alphaVantageResp, err = s.alphavantageClient.GetSMA(ctx, symbol, interval, timePeriod, "close")
	case "EMA":
		alphaVantageResp, err = s.alphavantageClient.GetEMA(ctx, symbol, interval, timePeriod, "close")
	case "BBANDS":
		alphaVantageResp, err = s.alphavantageClient.GetBollingerBands(ctx, symbol, interval, timePeriod, "close", "2", "2")
	case "STOCH":
		alphaVantageResp, err = s.alphavantageClient.GetSTOCH(ctx, symbol, interval, "5", "3", "0", "0", "0")
	case "ADX":
		alphaVantageResp, err = s.alphavantageClient.GetADX(ctx, symbol, interval, timePeriod)
	case "CCI":
		alphaVantageResp, err = s.alphavantageClient.GetCCI(ctx, symbol, interval, timePeriod)
	case "AROON":
		alphaVantageResp, err = s.alphavantageClient.GetAROON(ctx, symbol, interval, timePeriod)
	default:
		return nil, response.BadRequest("Unsupported indicator. Supported: RSI, MACD, SMA, EMA, BBANDS, STOCH, ADX, CCI, AROON")
	}
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch technical indicators from Alpha Vantage", err,
			logger.String("symbol", symbol),
			logger.String("indicator", indicator))
		return nil, response.InternalServerError("Failed to fetch technical indicators")
	}

	// Use the response data (placeholder to avoid unused variable error)
	_ = alphaVantageResp

	// Convert to our response format using adapter
	// For now, create a simple response with the metadata
	indicators := &response.TechnicalIndicatorsResponse{
		Success: true,
		Message: "Technical indicators retrieved successfully",
		Data: &response.TechnicalIndicatorsPayload{
			Symbol:      symbol,
			Indicator:   indicator,
			Interval:    interval,
			TimePeriod:  timePeriod,
			DataSource:  "alphavantage",
			LastUpdated: time.Now(),
			// Note: Full conversion would need implementation of specific indicator response methods
			// For now, endpoint will return metadata only
		},
	}

	return indicators, nil
}

// GetFundamentalData gets fundamental financial data from Alpha Vantage
func (s *marketDataService) GetFundamentalData(ctx context.Context, symbol string) (*response.FundamentalDataResponse, error) {
	s.logger.Info(ctx, "Fetching fundamental data from Alpha Vantage",
		logger.String("symbol", symbol))

	// Get company overview
	overview, err := s.alphavantageClient.GetCompanyOverview(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch company overview from Alpha Vantage", err,
			logger.String("symbol", symbol))
		return nil, response.InternalServerError("Failed to fetch fundamental data")
	}
	// Get income statement
	_, err = s.alphavantageClient.GetIncomeStatement(ctx, symbol)
	if err != nil {
		s.logger.Warn(ctx, "Failed to fetch income statement, continuing with overview only",
			logger.String("symbol", symbol))
	}

	// Get balance sheet
	_, err = s.alphavantageClient.GetBalanceSheet(ctx, symbol)
	if err != nil {
		s.logger.Warn(ctx, "Failed to fetch balance sheet, continuing with overview only",
			logger.String("symbol", symbol))
	}
	// Get cash flow
	_, err = s.alphavantageClient.GetCashFlow(ctx, symbol)
	if err != nil {
		s.logger.Warn(ctx, "Failed to fetch cash flow, continuing with overview only",
			logger.String("symbol", symbol))
	}
	// Convert to our response format using adapter
	// For now, create a simple response with basic company overview data
	fundamentalData := &response.FundamentalDataResponse{
		Success: true,
		Message: "Fundamental data retrieved successfully",
		Data: &response.FundamentalDataPayload{
			Symbol:      symbol,
			CompanyName: overview.Name,
			Sector:      overview.Sector,
			Industry:    overview.Industry,
			DataSource:  "alphavantage",
			LastUpdated: time.Now(),
			// Note: Full conversion would need implementation of comprehensive fundamental response method
			// For now, endpoint will return basic metadata only
		},
	}

	return fundamentalData, nil
}

// GetEarningsData gets earnings data using Alpha Vantage
func (s *marketDataService) GetEarningsData(ctx context.Context, symbol string) (*response.EarningsDataResponse, error) {
	// Get company info to validate symbol
	_, err := s.companyRepo.GetByTicker(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Company not found for symbol", err,
			logger.String("symbol", symbol))
		return nil, response.NotFound("Company with symbol " + symbol)
	}

	// Fetch earnings data from Alpha Vantage
	earnings, err := s.alphavantageClient.GetEarnings(ctx, symbol)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch earnings from Alpha Vantage", err,
			logger.String("symbol", symbol))
		return nil, response.InternalServerError("Failed to fetch earnings data")
	}

	// Convert to response format
	var annualEarnings []*response.AnnualEarning
	for _, ae := range earnings.AnnualEarnings {
		eps, _ := strconv.ParseFloat(ae.ReportedEPS, 64)
		annualEarnings = append(annualEarnings, &response.AnnualEarning{
			FiscalDateEnding: ae.FiscalDateEnding,
			ReportedEPS:      eps,
		})
	}

	var quarterlyEarnings []*response.QuarterlyEarning
	for _, qe := range earnings.QuarterlyEarnings {
		reportedEPS, _ := strconv.ParseFloat(qe.ReportedEPS, 64)
		estimatedEPS, _ := strconv.ParseFloat(qe.EstimatedEPS, 64)
		surprise, _ := strconv.ParseFloat(qe.Surprise, 64)
		surprisePercentage, _ := strconv.ParseFloat(qe.SurprisePercentage, 64)

		quarterlyEarnings = append(quarterlyEarnings, &response.QuarterlyEarning{
			FiscalDateEnding:   qe.FiscalDateEnding,
			ReportedDate:       qe.ReportedDate,
			ReportedEPS:        reportedEPS,
			EstimatedEPS:       estimatedEPS,
			Surprise:           surprise,
			SurprisePercentage: surprisePercentage,
		})
	}

	earningsResponse := &response.EarningsDataResponse{
		Success: true,
		Message: "Earnings data retrieved successfully",
		Data: &response.EarningsDataPayload{
			Symbol:            symbol,
			DataSource:        "alphavantage",
			LastUpdated:       time.Now(),
			AnnualEarnings:    annualEarnings,
			QuarterlyEarnings: quarterlyEarnings,
		},
	}

	s.logger.Info(ctx, "Successfully retrieved earnings data",
		logger.String("symbol", symbol),
		logger.Int("annual_count", len(annualEarnings)),
		logger.Int("quarterly_count", len(quarterlyEarnings)))

	return earningsResponse, nil
}

// AlphaVantageHealthCheck checks Alpha Vantage API connectivity
func (s *marketDataService) AlphaVantageHealthCheck(ctx context.Context) (bool, error) {
	err := s.alphavantageClient.HealthCheck(ctx)
	if err != nil {
		s.logger.Error(ctx, "Alpha Vantage health check failed", err)
		return false, err
	}

	s.logger.Info(ctx, "Alpha Vantage health check passed")
	return true, nil
}

// RefreshMarketData refreshes market data for multiple symbols
func (s *marketDataService) RefreshMarketData(ctx context.Context, symbols []string) error {
	if len(symbols) == 0 {
		return nil
	}

	s.logger.Info(ctx, "Starting bulk market data refresh",
		logger.Int("symbol_count", len(symbols)))

	var errors []string
	successCount := 0

	for _, symbol := range symbols {
		_, err := s.GetRealTimeQuote(ctx, symbol)
		if err != nil {
			s.logger.Error(ctx, "Failed to refresh data for symbol", err,
				logger.String("symbol", symbol))
			errors = append(errors, symbol+": "+err.Error())
		} else {
			successCount++
		}
	}

	s.logger.Info(ctx, "Bulk market data refresh completed",
		logger.Int("success_count", successCount),
		logger.Int("error_count", len(errors)),
		logger.Int("total_symbols", len(symbols)))

	if len(errors) > 0 && successCount == 0 {
		return response.InternalServerError("Failed to refresh data for all symbols")
	}

	return nil
}

// Helper conversion methods

func (s *marketDataService) convertToMarketDataResponse(md *entities.MarketData) *response.MarketDataResponse {
	return &response.MarketDataResponse{
		ID:              md.ID,
		CompanyID:       md.CompanyID,
		Symbol:          md.Symbol,
		CurrentPrice:    md.CurrentPrice,
		OpenPrice:       md.OpenPrice,
		HighPrice:       md.HighPrice,
		LowPrice:        md.LowPrice,
		PreviousClose:   md.PreviousClose,
		PriceChange:     md.PriceChange,
		PriceChangePerc: md.PriceChangePerc,
		Volume:          md.Volume,
		AvgVolume:       md.AvgVolume,
		MarketCap:       md.MarketCap,
		IsMarketOpen:    md.IsMarketOpen,
		Currency:        md.Currency,
		Exchange:        md.Exchange,
		MarketTimestamp: md.MarketTimestamp,
		LastUpdated:     md.UpdatedAt,
	}
}

func (s *marketDataService) convertToCompanyProfileResponse(cp *entities.CompanyProfile) *response.CompanyProfileResponse {
	return &response.CompanyProfileResponse{
		ID:                cp.ID,
		Symbol:            cp.Symbol,
		Name:              cp.Name,
		Description:       cp.Description,
		Industry:          cp.Industry,
		Sector:            cp.Sector,
		Country:           cp.Country,
		Currency:          cp.Currency,
		MarketCap:         cp.MarketCap,
		SharesOutstanding: cp.SharesOutstanding,
		PERatio:           cp.PERatio,
		PEGRatio:          cp.PEGRatio,
		PriceToBook:       cp.PriceToBook,
		DividendYield:     cp.DividendYield,
		EPS:               cp.EPS,
		Beta:              cp.Beta,
		Website:           cp.Website,
		Logo:              cp.Logo,
		IPODate:           cp.IPODate,
		EmployeeCount:     cp.EmployeeCount,
		LastUpdated:       cp.LastUpdated,
	}
}

func (s *marketDataService) convertToNewsResponse(ni *entities.NewsItem) *response.NewsResponse {
	return &response.NewsResponse{
		ID:             ni.ID,
		Symbol:         ni.Symbol,
		Title:          ni.Title,
		Summary:        ni.Summary,
		URL:            ni.URL,
		ImageURL:       ni.ImageURL,
		Source:         ni.Source,
		Category:       ni.Category,
		Language:       ni.Language,
		SentimentScore: ni.SentimentScore,
		SentimentLabel: ni.SentimentLabel,
		PublishedAt:    ni.PublishedAt,
		CreatedAt:      ni.CreatedAt,
	}
}

func (s *marketDataService) convertToBasicFinancialsResponse(bf *entities.BasicFinancials) *response.BasicFinancialsResponse {
	return &response.BasicFinancialsResponse{
		ID:                bf.ID,
		Symbol:            bf.Symbol,
		PERatio:           bf.PERatio,
		PEGRatio:          bf.PEGRatio,
		PriceToSales:      bf.PriceToSales,
		PriceToBook:       bf.PriceToBook,
		PriceToCashFlow:   bf.PriceToCashFlow,
		ROE:               bf.ROE,
		ROA:               bf.ROA,
		ROI:               bf.ROI,
		GrossMargin:       bf.GrossMargin,
		OperatingMargin:   bf.OperatingMargin,
		NetMargin:         bf.NetMargin,
		RevenueGrowth:     bf.RevenueGrowth,
		EarningsGrowth:    bf.EarningsGrowth,
		DividendGrowth:    bf.DividendGrowth,
		DebtToEquity:      bf.DebtToEquity,
		CurrentRatio:      bf.CurrentRatio,
		QuickRatio:        bf.QuickRatio,
		EPS:               bf.EPS,
		BookValuePerShare: bf.BookValuePerShare,
		CashPerShare:      bf.CashPerShare,
		DividendPerShare:  bf.DividendPerShare,
		Period:            bf.Period,
		FiscalYear:        bf.FiscalYear,
		FiscalQuarter:     bf.FiscalQuarter,
		LastUpdated:       bf.LastUpdated,
	}
}
