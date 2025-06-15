package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// Client represents the Alpha Vantage API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     logger.Logger
}

// NewClient creates a new Alpha Vantage API client
func NewClient(cfg *config.Config, log logger.Logger) *Client {
	return &Client{
		baseURL: cfg.External.Secondary.BaseURL,
		apiKey:  cfg.External.Secondary.Key,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: log,
	}
}

// makeRequest makes an HTTP request to the Alpha Vantage API
func (c *Client) makeRequest(ctx context.Context, function string, params map[string]string) ([]byte, error) {
	// Build URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		c.logger.Error(ctx, "Failed to parse base URL", err, logger.String("baseURL", c.baseURL))
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	// Add query parameters
	query := u.Query()
	query.Set("function", function)
	query.Set("apikey", c.apiKey)

	for key, value := range params {
		query.Set(key, value)
	}

	u.RawQuery = query.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error(ctx, "Failed to create request", err, logger.String("url", u.String()))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Stock-Info-App/1.0")

	// Log request details
	c.logger.Debug(ctx, "Making Alpha Vantage API request",
		logger.String("function", function),
		logger.String("url", u.String()))

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(ctx, "Request failed", err, logger.String("url", u.String()))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(ctx, "Failed to read response body", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		c.logger.Error(ctx, "API request failed", fmt.Errorf("status code %d", resp.StatusCode),
			logger.Int("statusCode", resp.StatusCode),
			logger.String("status", resp.Status),
			logger.String("body", string(body)))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	// Check for API errors in response
	var errorCheck AlphaVantageResponse
	if err := json.Unmarshal(body, &errorCheck); err == nil {
		if errorCheck.ErrorMessage != "" {
			c.logger.Error(ctx, "Alpha Vantage API error", fmt.Errorf("API error: %s", errorCheck.ErrorMessage))
			return nil, fmt.Errorf("alpha Vantage API error: %s", errorCheck.ErrorMessage)
		}
		if errorCheck.Information != "" && errorCheck.Information == "Thank you for using Alpha Vantage! Our standard API call frequency is 5 calls per minute and 500 calls per day." {
			c.logger.Warn(ctx, "Alpha Vantage rate limit information", logger.String("info", errorCheck.Information))
		}
		if errorCheck.Note != "" {
			c.logger.Warn(ctx, "Alpha Vantage API note", logger.String("note", errorCheck.Note))
			return nil, fmt.Errorf("alpha Vantage API note: %s", errorCheck.Note)
		}
	}

	c.logger.Debug(ctx, "Alpha Vantage API request successful",
		logger.String("function", function),
		logger.Int("responseSize", len(body)))

	return body, nil
}

// GetTimeSeriesDaily retrieves daily historical data for a symbol
func (c *Client) GetTimeSeriesDaily(ctx context.Context, symbol string, outputSize string) (*TimeSeriesDailyResponse, error) {
	params := map[string]string{
		"symbol":     symbol,
		"outputsize": outputSize, // "compact" or "full"
	}

	body, err := c.makeRequest(ctx, "TIME_SERIES_DAILY_ADJUSTED", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily time series for %s: %w", symbol, err)
	}
	var response TimeSeriesDailyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal daily time series response", err,
			logger.String("symbol", symbol),
			logger.String("responsePreview", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved daily time series",
		logger.String("symbol", symbol),
		logger.Int("dataPoints", len(response.TimeSeries)))

	return &response, nil
}

// GetTimeSeriesWeekly retrieves weekly historical data for a symbol
func (c *Client) GetTimeSeriesWeekly(ctx context.Context, symbol string) (*TimeSeriesWeeklyResponse, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	body, err := c.makeRequest(ctx, "TIME_SERIES_WEEKLY_ADJUSTED", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly time series for %s: %w", symbol, err)
	}
	var response TimeSeriesWeeklyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal weekly time series response", err,
			logger.String("symbol", symbol),
			logger.String("responsePreview", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved weekly time series",
		logger.String("symbol", symbol),
		logger.Int("dataPoints", len(response.TimeSeries)))

	return &response, nil
}

// GetTimeSeriesMonthly retrieves monthly historical data for a symbol
func (c *Client) GetTimeSeriesMonthly(ctx context.Context, symbol string) (*TimeSeriesMonthlyResponse, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	body, err := c.makeRequest(ctx, "TIME_SERIES_MONTHLY_ADJUSTED", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly time series for %s: %w", symbol, err)
	}
	var response TimeSeriesMonthlyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal monthly time series response", err,
			logger.String("symbol", symbol),
			logger.String("responsePreview", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved monthly time series",
		logger.String("symbol", symbol),
		logger.Int("dataPoints", len(response.TimeSeries)))

	return &response, nil
}

// GetCompanyOverview retrieves fundamental data for a symbol
func (c *Client) GetCompanyOverview(ctx context.Context, symbol string) (*CompanyOverviewResponse, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	body, err := c.makeRequest(ctx, "OVERVIEW", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get company overview for %s: %w", symbol, err)
	}
	var response CompanyOverviewResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal company overview response", err,
			logger.String("symbol", symbol),
			logger.String("responsePreview", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved company overview",
		logger.String("symbol", symbol),
		logger.String("company", response.Name))

	return &response, nil
}

// GetRSI retrieves RSI indicator for a symbol
func (c *Client) GetRSI(ctx context.Context, symbol, interval, timePeriod, seriesType string) (*RSIResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,   // "1min", "5min", "15min", "30min", "60min", "daily", "weekly", "monthly"
		"time_period": timePeriod, // typically "14"
		"series_type": seriesType, // "close", "open", "high", "low"
	}

	body, err := c.makeRequest(ctx, "RSI", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get RSI for %s: %w", symbol, err)
	}
	var response RSIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal RSI response", err,
			logger.String("symbol", symbol),
			logger.String("responsePreview", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved RSI",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.Int("dataPoints", len(response.RSI)))

	return &response, nil
}

// GetMACD retrieves MACD indicator for a symbol
func (c *Client) GetMACD(ctx context.Context, symbol, interval, fastPeriod, slowPeriod, signalPeriod, seriesType string) (*MACDResponse, error) {
	params := map[string]string{
		"symbol":       symbol,
		"interval":     interval,
		"fastperiod":   fastPeriod,   // typically "12"
		"slowperiod":   slowPeriod,   // typically "26"
		"signalperiod": signalPeriod, // typically "9"
		"series_type":  seriesType,
	}

	body, err := c.makeRequest(ctx, "MACD", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get MACD for %s: %w", symbol, err)
	}
	var response MACDResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal MACD response", err,
			logger.String("symbol", symbol),
			logger.String("responsePreview", string(body[:min(500, len(body))])))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved MACD",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.Int("dataPoints", len(response.MACD)))

	return &response, nil
}

// GetBollingerBands retrieves Bollinger Bands for a symbol
func (c *Client) GetBollingerBands(ctx context.Context, symbol, interval, timePeriod, seriesType, nbdevup, nbdevdn string) (*BollingerBandsResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": timePeriod, // typically "20"
		"series_type": seriesType,
		"nbdevup":     nbdevup, // typically "2"
		"nbdevdn":     nbdevdn, // typically "2"
	}

	body, err := c.makeRequest(ctx, "BBANDS", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bollinger Bands for %s: %w", symbol, err)
	}

	var response BollingerBandsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal Bollinger Bands response", err,
			logger.String("error", err.Error()),
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved Bollinger Bands",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.Int("dataPoints", len(response.Bands)))

	return &response, nil
}

// GetSMA retrieves Simple Moving Average for a symbol
func (c *Client) GetSMA(ctx context.Context, symbol, interval, timePeriod, seriesType string) (*SMAResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": timePeriod,
		"series_type": seriesType,
	}

	body, err := c.makeRequest(ctx, "SMA", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMA for %s: %w", symbol, err)
	}

	var response SMAResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal SMA response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved SMA",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.String("period", timePeriod),
		logger.Int("dataPoints", len(response.SMA)))

	return &response, nil
}

// GetEMA retrieves Exponential Moving Average for a symbol
func (c *Client) GetEMA(ctx context.Context, symbol, interval, timePeriod, seriesType string) (*EMAResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": timePeriod,
		"series_type": seriesType,
	}

	body, err := c.makeRequest(ctx, "EMA", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get EMA for %s: %w", symbol, err)
	}

	var response EMAResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal EMA response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved EMA",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.String("period", timePeriod),
		logger.Int("dataPoints", len(response.EMA)))

	return &response, nil
}

// GetSTOCH retrieves Stochastic Oscillator for a symbol
func (c *Client) GetSTOCH(ctx context.Context, symbol, interval, fastkperiod, slowkperiod, slowdperiod, slowkmatype, slowdmatype string) (*STOCHResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"fastkperiod": fastkperiod, // typically "5"
		"slowkperiod": slowkperiod, // typically "3"
		"slowdperiod": slowdperiod, // typically "3"
		"slowkmatype": slowkmatype, // 0=SMA, 1=EMA, 2=WMA, 3=DEMA, 4=TEMA, etc.
		"slowdmatype": slowdmatype,
	}

	body, err := c.makeRequest(ctx, "STOCH", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get STOCH for %s: %w", symbol, err)
	}

	var response STOCHResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal STOCH response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved STOCH",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.Int("dataPoints", len(response.STOCH)))

	return &response, nil
}

// GetADX retrieves ADX indicator for a symbol
func (c *Client) GetADX(ctx context.Context, symbol, interval, timePeriod string) (*ADXResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": timePeriod, // typically "14"
	}

	body, err := c.makeRequest(ctx, "ADX", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get ADX for %s: %w", symbol, err)
	}

	var response ADXResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal ADX response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved ADX",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.Int("dataPoints", len(response.ADX)))

	return &response, nil
}

// GetCCI retrieves CCI indicator for a symbol
func (c *Client) GetCCI(ctx context.Context, symbol, interval, timePeriod string) (*CCIResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": timePeriod, // typically "20"
	}

	body, err := c.makeRequest(ctx, "CCI", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get CCI for %s: %w", symbol, err)
	}

	var response CCIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal CCI response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved CCI",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.Int("dataPoints", len(response.CCI)))

	return &response, nil
}

// GetAROON retrieves AROON indicator for a symbol
func (c *Client) GetAROON(ctx context.Context, symbol, interval, timePeriod string) (*AROONResponse, error) {
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": timePeriod, // typically "14"
	}

	body, err := c.makeRequest(ctx, "AROON", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get AROON for %s: %w", symbol, err)
	}

	var response AROONResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal AROON response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved AROON",
		logger.String("symbol", symbol),
		logger.String("interval", interval),
		logger.Int("dataPoints", len(response.AROON)))

	return &response, nil
}

// GetEarnings retrieves earnings data for a symbol
func (c *Client) GetEarnings(ctx context.Context, symbol string) (*EarningsResponse, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	body, err := c.makeRequest(ctx, "EARNINGS", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get earnings for %s: %w", symbol, err)
	}

	var response EarningsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal earnings response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved earnings",
		logger.String("symbol", symbol),
		logger.Int("annualEarnings", len(response.AnnualEarnings)),
		logger.Int("quarterlyEarnings", len(response.QuarterlyEarnings)))

	return &response, nil
}

// GetIncomeStatement retrieves income statement for a symbol
func (c *Client) GetIncomeStatement(ctx context.Context, symbol string) (*IncomeStatementResponse, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	body, err := c.makeRequest(ctx, "INCOME_STATEMENT", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get income statement for %s: %w", symbol, err)
	}

	var response IncomeStatementResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal income statement response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved income statement",
		logger.String("symbol", symbol),
		logger.Int("annualReports", len(response.AnnualReports)),
		logger.Int("quarterlyReports", len(response.QuarterlyReports)))

	return &response, nil
}

// GetBalanceSheet retrieves balance sheet for a symbol
func (c *Client) GetBalanceSheet(ctx context.Context, symbol string) (*BalanceSheetResponse, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	body, err := c.makeRequest(ctx, "BALANCE_SHEET", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance sheet for %s: %w", symbol, err)
	}

	var response BalanceSheetResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal balance sheet response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved balance sheet",
		logger.String("symbol", symbol),
		logger.Int("annualReports", len(response.AnnualReports)),
		logger.Int("quarterlyReports", len(response.QuarterlyReports)))

	return &response, nil
}

// GetCashFlow retrieves cash flow statement for a symbol
func (c *Client) GetCashFlow(ctx context.Context, symbol string) (*CashFlowResponse, error) {
	params := map[string]string{
		"symbol": symbol,
	}

	body, err := c.makeRequest(ctx, "CASH_FLOW", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash flow for %s: %w", symbol, err)
	}

	var response CashFlowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal cash flow response", err,
			logger.String("symbol", symbol))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info(ctx, "Successfully retrieved cash flow",
		logger.String("symbol", symbol),
		logger.Int("annualReports", len(response.AnnualReports)),
		logger.Int("quarterlyReports", len(response.QuarterlyReports)))

	return &response, nil
}

// HealthCheck verifies the API is accessible
func (c *Client) HealthCheck(ctx context.Context) error {
	// Use a known symbol for health check
	_, err := c.GetCompanyOverview(ctx, "AAPL")
	if err != nil {
		c.logger.Error(ctx, "Alpha Vantage health check failed", err)
		return fmt.Errorf("alpha vantage health check failed: %w", err)
	}

	c.logger.Info(ctx, "Alpha Vantage health check passed")
	return nil
}

// min returns the minimum of two integers (Go 1.21+ has min function in stdlib)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
