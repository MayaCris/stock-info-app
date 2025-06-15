package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// Client represents Finnhub API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     logger.Logger
}

// ClientConfig represents configuration for Finnhub client
type ClientConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
	Logger  logger.Logger
}

// NewClient creates a new Finnhub API client
func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: config.Logger,
	}
}

// GetRealTimeQuote gets real-time quote for a symbol
func (c *Client) GetRealTimeQuote(ctx context.Context, symbol string) (*QuoteResponse, error) {
	endpoint := "/quote"
	params := url.Values{
		"symbol": {symbol},
	}

	var quote QuoteResponse
	if err := c.makeRequest(ctx, endpoint, params, &quote); err != nil {
		c.logger.Error(ctx, "Failed to get real-time quote", err,
			logger.String("symbol", symbol),
		)
		return nil, fmt.Errorf("failed to get real-time quote for %s: %w", symbol, err)
	}

	if !quote.IsValid() {
		return nil, fmt.Errorf("invalid quote data for symbol %s", symbol)
	}

	c.logger.Info(ctx, "Successfully retrieved real-time quote",
		logger.String("symbol", symbol),
		logger.Float64("price", quote.CurrentPrice),
	)

	return &quote, nil
}

// GetCompanyProfile gets company profile information
func (c *Client) GetCompanyProfile(ctx context.Context, symbol string) (*CompanyProfileResponse, error) {
	endpoint := "/stock/profile2"
	params := url.Values{
		"symbol": {symbol},
	}

	var profile CompanyProfileResponse
	if err := c.makeRequest(ctx, endpoint, params, &profile); err != nil {
		c.logger.Error(ctx, "Failed to get company profile", err,
			logger.String("symbol", symbol),
		)
		return nil, fmt.Errorf("failed to get company profile for %s: %w", symbol, err)
	}

	if !profile.IsValid() {
		return nil, fmt.Errorf("invalid company profile data for symbol %s", symbol)
	}

	c.logger.Info(ctx, "Successfully retrieved company profile",
		logger.String("symbol", symbol),
		logger.String("company_name", profile.Name),
	)

	return &profile, nil
}

// GetCompanyNews gets company news for a symbol
func (c *Client) GetCompanyNews(ctx context.Context, symbol string, from, to time.Time) (NewsResponse, error) {
	endpoint := "/company-news"
	params := url.Values{
		"symbol": {symbol},
		"from":   {from.Format("2006-01-02")},
		"to":     {to.Format("2006-01-02")},
	}

	var news NewsResponse
	if err := c.makeRequest(ctx, endpoint, params, &news); err != nil {
		c.logger.Error(ctx, "Failed to get company news", err,
			logger.String("symbol", symbol),
			logger.String("from", from.Format("2006-01-02")),
			logger.String("to", to.Format("2006-01-02")),
		)
		return nil, fmt.Errorf("failed to get company news for %s: %w", symbol, err)
	}

	c.logger.Info(ctx, "Successfully retrieved company news",
		logger.String("symbol", symbol),
		logger.Int("news_count", len(news)),
	)

	return news, nil
}

// GetBasicFinancials gets basic financial metrics for a symbol
func (c *Client) GetBasicFinancials(ctx context.Context, symbol string) (*BasicFinancialsResponse, error) {
	endpoint := "/stock/metric"
	params := url.Values{
		"symbol": {symbol},
		"metric": {"all"},
	}

	var financials BasicFinancialsResponse
	if err := c.makeRequest(ctx, endpoint, params, &financials); err != nil {
		c.logger.Error(ctx, "Failed to get basic financials", err,
			logger.String("symbol", symbol),
		)
		return nil, fmt.Errorf("failed to get basic financials for %s: %w", symbol, err)
	}

	if !financials.IsValid() {
		return nil, fmt.Errorf("invalid basic financials data for symbol %s", symbol)
	}

	c.logger.Info(ctx, "Successfully retrieved basic financials",
		logger.String("symbol", symbol),
	)

	return &financials, nil
}

// GetMarketNews gets general market news
func (c *Client) GetMarketNews(ctx context.Context, category string, minID int64) (MarketNewsResponse, error) {
	endpoint := "/news"
	params := url.Values{
		"category": {category},
	}

	if minID > 0 {
		params.Set("minId", strconv.FormatInt(minID, 10))
	}

	var news MarketNewsResponse
	if err := c.makeRequest(ctx, endpoint, params, &news); err != nil {
		c.logger.Error(ctx, "Failed to get market news", err,
			logger.String("category", category),
		)
		return nil, fmt.Errorf("failed to get market news for category %s: %w", category, err)
	}

	c.logger.Info(ctx, "Successfully retrieved market news",
		logger.String("category", category),
		logger.Int("news_count", len(news)),
	)

	return news, nil
}

// GetRecommendationTrends gets recommendation trends for a symbol
func (c *Client) GetRecommendationTrends(ctx context.Context, symbol string) (RecommendationTrendsResponse, error) {
	endpoint := "/stock/recommendation"
	params := url.Values{
		"symbol": {symbol},
	}

	var trends RecommendationTrendsResponse
	if err := c.makeRequest(ctx, endpoint, params, &trends); err != nil {
		c.logger.Error(ctx, "Failed to get recommendation trends", err,
			logger.String("symbol", symbol),
		)
		return nil, fmt.Errorf("failed to get recommendation trends for %s: %w", symbol, err)
	}

	c.logger.Info(ctx, "Successfully retrieved recommendation trends",
		logger.String("symbol", symbol),
		logger.Int("periods_count", len(trends)),
	)

	return trends, nil
}

// GetEarnings gets earnings data for a symbol
func (c *Client) GetEarnings(ctx context.Context, symbol string) (EarningsResponse, error) {
	endpoint := "/stock/earnings"
	params := url.Values{
		"symbol": {symbol},
	}

	var earnings EarningsResponse
	if err := c.makeRequest(ctx, endpoint, params, &earnings); err != nil {
		c.logger.Error(ctx, "Failed to get earnings", err,
			logger.String("symbol", symbol),
		)
		return nil, fmt.Errorf("failed to get earnings for %s: %w", symbol, err)
	}

	c.logger.Info(ctx, "Successfully retrieved earnings",
		logger.String("symbol", symbol),
		logger.Int("earnings_count", len(earnings)),
	)

	return earnings, nil
}

// GetStockSymbols gets list of supported stock symbols for an exchange
func (c *Client) GetStockSymbols(ctx context.Context, exchange string) (StockSymbolsResponse, error) {
	endpoint := "/stock/symbol"
	params := url.Values{
		"exchange": {exchange},
	}

	var symbols StockSymbolsResponse
	if err := c.makeRequest(ctx, endpoint, params, &symbols); err != nil {
		c.logger.Error(ctx, "Failed to get stock symbols", err,
			logger.String("exchange", exchange),
		)
		return nil, fmt.Errorf("failed to get stock symbols for exchange %s: %w", exchange, err)
	}

	c.logger.Info(ctx, "Successfully retrieved stock symbols",
		logger.String("exchange", exchange),
		logger.Int("symbols_count", len(symbols)),
	)

	return symbols, nil
}

// GetMarketStatus gets current market status
func (c *Client) GetMarketStatus(ctx context.Context, exchange string) (map[string]interface{}, error) {
	endpoint := "/stock/market-status"
	params := url.Values{
		"exchange": {exchange},
	}

	var status map[string]interface{}
	if err := c.makeRequest(ctx, endpoint, params, &status); err != nil {
		c.logger.Error(ctx, "Failed to get market status", err,
			logger.String("exchange", exchange),
		)
		return nil, fmt.Errorf("failed to get market status for exchange %s: %w", exchange, err)
	}

	c.logger.Info(ctx, "Successfully retrieved market status",
		logger.String("exchange", exchange),
	)

	return status, nil
}

// makeRequest makes HTTP request to Finnhub API
func (c *Client) makeRequest(ctx context.Context, endpoint string, params url.Values, result interface{}) error {
	// Add API key to parameters
	params.Set("token", c.apiKey)

	// Build URL
	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, params.Encode())

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "stock-info-app/1.0")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if json.Unmarshal(body, &errorResp) == nil && errorResp.Error != "" {
			return fmt.Errorf("API error: %s", errorResp.Error)
		}
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	// Check for rate limiting
	if resp.Header.Get("X-Ratelimit-Remaining") == "0" {
		c.logger.Warn(ctx, "Finnhub API rate limit approaching",
			logger.String("reset_time", resp.Header.Get("X-Ratelimit-Reset")),
		)
	}

	// Parse JSON response
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}

// Health checks if the Finnhub API is accessible
func (c *Client) Health(ctx context.Context) error {
	// Try to get market status for a known exchange
	_, err := c.GetMarketStatus(ctx, "US")
	if err != nil {
		return fmt.Errorf("Finnhub API health check failed: %w", err)
	}
	return nil
}

// BatchQuotes gets real-time quotes for multiple symbols
func (c *Client) BatchQuotes(ctx context.Context, symbols []string) (map[string]*QuoteResponse, error) {
	results := make(map[string]*QuoteResponse)
	errors := make([]error, 0)

	// Finnhub doesn't support batch requests, so we make individual requests
	// In production, you might want to implement request pooling and rate limiting
	for _, symbol := range symbols {
		quote, err := c.GetRealTimeQuote(ctx, symbol)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get quote for %s: %w", symbol, err))
			continue
		}
		results[symbol] = quote

		// Small delay to avoid rate limiting
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	if len(errors) > 0 {
		c.logger.Warn(ctx, "Some batch quote requests failed",
			logger.Int("failed_count", len(errors)),
			logger.Int("success_count", len(results)),
		)
	}

	return results, nil
}

// GetQuoteWithRetry gets quote with retry logic
func (c *Client) GetQuoteWithRetry(ctx context.Context, symbol string, maxRetries int) (*QuoteResponse, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			// Exponential backoff
			delay := time.Duration(i*i) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		quote, err := c.GetRealTimeQuote(ctx, symbol)
		if err == nil {
			return quote, nil
		}

		lastErr = err
		c.logger.Warn(ctx, "Quote request failed, retrying",
			logger.String("symbol", symbol),
			logger.Int("attempt", i+1),
			logger.Int("max_retries", maxRetries),
			logger.String("error", err.Error()),
		)
	}

	return nil, fmt.Errorf("failed to get quote after %d retries: %w", maxRetries, lastErr)
}
