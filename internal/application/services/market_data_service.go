package services

import (
	"context"
	"time"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/application/services/interfaces"
	"github.com/MayaCris/stock-info-app/internal/domain/entities"
	repoInterfaces "github.com/MayaCris/stock-info-app/internal/domain/repositories/interfaces"
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
	finnhubClient  *finnhub.Client
	finnhubAdapter *finnhub.Adapter

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

// RefreshMarketData refreshes market data for multiple symbols
func (s *marketDataService) RefreshMarketData(ctx context.Context, symbols []string) error {
	s.logger.Info(ctx, "Starting market data refresh",
		logger.Int("symbol_count", len(symbols)),
	)

	for _, symbol := range symbols {
		// Refresh with some delay to avoid rate limiting
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}

		_, err := s.GetRealTimeQuote(ctx, symbol)
		if err != nil {
			s.logger.Warn(ctx, "Failed to refresh market data for symbol",
				logger.String("symbol", symbol),
				logger.String("error", err.Error()),
			)
			continue
		}
	}

	s.logger.Info(ctx, "Completed market data refresh",
		logger.Int("symbol_count", len(symbols)),
	)

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
