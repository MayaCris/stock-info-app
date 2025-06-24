//go:build integration
// +build integration

package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	services "github.com/MayaCris/stock-info-app/internal/application/services"
	repoImpl "github.com/MayaCris/stock-info-app/internal/domain/repositories/implementation"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	alphavantage "github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/alphavantage"
	finnhub "github.com/MayaCris/stock-info-app/internal/infrastructure/external/market_data/finnhub"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/handlers"
	"github.com/MayaCris/stock-info-app/internal/presentation/rest/routes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupRouterAndHandler() (*gin.Engine, *handlers.AlphaVantageHandler, *gorm.DB) {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	factory := logger.NewLoggerFactory()
	log, err := factory.CreateLogger(cfg.App.Env)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode,
	)), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DB: %v", err))
	}
	marketDataRepo := repoImpl.NewMarketDataRepository(db)
	companyProfileRepo := repoImpl.NewCompanyProfileRepository(db)
	newsRepo := repoImpl.NewNewsRepository(db)
	basicFinancialsRepo := repoImpl.NewBasicFinancialsRepository(db)
	companyRepo := repoImpl.NewCompanyRepository(db)
	finnhubClient := finnhub.NewClient(finnhub.ClientConfig{
		BaseURL: cfg.External.Primary.BaseURL,
		APIKey:  cfg.External.Primary.Key,
		Timeout: 30 * time.Second,
		Logger:  log,
	})
	finnhubAdapter := finnhub.NewAdapter(log)
	alphaVantageClient := alphavantage.NewClient(cfg, log)
	alphaVantageAdapter := alphavantage.NewAdapter(log)
	service := services.NewMarketDataService(services.MarketDataServiceConfig{
		MarketDataRepo:      marketDataRepo,
		CompanyProfileRepo:  companyProfileRepo,
		NewsRepo:            newsRepo,
		BasicFinancialsRepo: basicFinancialsRepo,
		CompanyRepo:         companyRepo,
		FinnhubClient:       finnhubClient,
		FinnhubAdapter:      finnhubAdapter,
		AlphaVantageClient:  alphaVantageClient,
		AlphaVantageAdapter: alphaVantageAdapter,
		Logger:              log,
	})
	handler := handlers.NewAlphaVantageHandler(service, log)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	v1 := r.Group("/api/v1")

	// Setup routes using the new handler pattern
	alphaVantageRoutes := routes.NewAlphaVantageRoutes(nil) // middleware manager not needed for tests
	alphaVantageRoutes.SetupAlphaVantageRoutes(v1, handler)

	return r, handler, db
}

func getRandomSymbol(db *gorm.DB) string {
	type result struct{ Ticker string }
	var res result
	db.Raw("SELECT ticker FROM companies ORDER BY RANDOM() LIMIT 1").Scan(&res)
	return res.Ticker
}

func debugAlphaVantageResponse(t *testing.T, cfg *config.Config, symbol string) {
	t.Helper()

	// Test direct Alpha Vantage call
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=OVERVIEW&symbol=%s&apikey=%s",
		symbol, cfg.External.Secondary.Key)

	resp, err := http.Get(url)
	if err != nil {
		t.Logf("Direct Alpha Vantage call failed: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Failed to read Alpha Vantage response: %v", err)
		return
	}

	t.Logf("Direct Alpha Vantage response for %s (size: %d bytes):", symbol, len(body))
	if len(body) < 1000 {
		t.Logf("Response body: %s", string(body))
	} else {
		t.Logf("Response body (first 500 chars): %s...", string(body[:500]))
	}
}

func TestAlphaVantageEndpoints_Integration(t *testing.T) {
	// Skip in production environment
	if os.Getenv("APP_ENV") == "production" {
		t.Skip("Skipping integration test in production environment")
	}

	// Skip if Alpha Vantage API key is not configured
	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load configuration")
	if cfg.External.Secondary.Key == "" {
		t.Skip("Alpha Vantage API key not configured, skipping integration tests")
	}
	router, _, db := setupRouterAndHandler()

	// Use a well-known symbol for consistent testing
	symbol := getRandomSymbol(db)
	if symbol == "" {
		symbol = "AAPL" // Fallback to Apple if no symbols in DB
		t.Logf("No symbols found in database, using fallback symbol: %s", symbol)
	} else {
		t.Logf("Using symbol from database: %s", symbol)
	}

	// Debug Alpha Vantage responses to understand the 258-byte issue
	t.Logf("=== Debugging Alpha Vantage API responses ===")
	debugAlphaVantageResponse(t, cfg, symbol)
	t.Logf("=== End debugging ===")

	testCases := []struct {
		name           string
		url            string
		expectedStatus int
		validateJSON   func(t *testing.T, body []byte)
	}{
		{
			name:           "HealthCheck",
			url:            "/api/v1/alpha/health",
			expectedStatus: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var healthResp map[string]interface{}
				err := json.Unmarshal(body, &healthResp)
				require.NoError(t, err, "Should parse health check response as JSON")
				assert.Equal(t, "healthy", healthResp["status"], "Health check should return healthy status")
				assert.Equal(t, "alpha", healthResp["service"], "Should identify alpha service")
			},
		},
		{
			name:           "HistoricalData",
			url:            fmt.Sprintf("/api/v1/alpha/historical/%s?period=daily&outputsize=compact", symbol),
			expectedStatus: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var histResp response.HistoricalDataResponse
				err := json.Unmarshal(body, &histResp)
				require.NoError(t, err, "Should parse historical data response as JSON")
				assert.True(t, histResp.Success, "Historical data request should be successful")
				assert.NotNil(t, histResp.Data, "Historical data should have data payload")
				assert.Equal(t, symbol, histResp.Data.Symbol, "Symbol should match request")
				assert.Equal(t, "daily", histResp.Data.Period, "Period should match request")
			},
		},
		{
			name:           "TechnicalIndicators",
			url:            fmt.Sprintf("/api/v1/alpha/technical/%s?indicator=RSI&interval=daily&time_period=14", symbol),
			expectedStatus: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var techResp response.TechnicalIndicatorsResponse
				err := json.Unmarshal(body, &techResp)
				require.NoError(t, err, "Should parse technical indicators response as JSON")
				assert.True(t, techResp.Success, "Technical indicators request should be successful")
				assert.NotNil(t, techResp.Data, "Technical indicators should have data payload")
				assert.Equal(t, symbol, techResp.Data.Symbol, "Symbol should match request")
				assert.Equal(t, "RSI", techResp.Data.Indicator, "Indicator should match request")
			},
		},
		{
			name:           "FundamentalData",
			url:            fmt.Sprintf("/api/v1/alpha/fundamental/%s", symbol),
			expectedStatus: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var fundResp response.FundamentalDataResponse
				err := json.Unmarshal(body, &fundResp)
				require.NoError(t, err, "Should parse fundamental data response as JSON")
				assert.True(t, fundResp.Success, "Fundamental data request should be successful")
				assert.NotNil(t, fundResp.Data, "Fundamental data should have data payload")
				assert.Equal(t, symbol, fundResp.Data.Symbol, "Symbol should match request")
			},
		},
		{
			name:           "EarningsData",
			url:            fmt.Sprintf("/api/v1/alpha/earnings/%s", symbol),
			expectedStatus: http.StatusOK,
			validateJSON: func(t *testing.T, body []byte) {
				var earnResp response.EarningsDataResponse
				err := json.Unmarshal(body, &earnResp)
				require.NoError(t, err, "Should parse earnings data response as JSON")
				assert.True(t, earnResp.Success, "Earnings data request should be successful")
				assert.NotNil(t, earnResp.Data, "Earnings data should have data payload")
				assert.Equal(t, symbol, earnResp.Data.Symbol, "Symbol should match request")
			},
		},
	}
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Add delay between requests to avoid rate limiting
			if i > 0 {
				time.Sleep(1 * time.Second)
				t.Logf("Added 1s delay before %s to avoid rate limiting", tc.name)
			}

			// Create request with proper context and headers
			req, err := http.NewRequestWithContext(context.Background(), "GET", tc.url, nil)
			require.NoError(t, err, "Should create HTTP request")
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Log response size and basic info first
			responseBody := w.Body.String()
			t.Logf("Endpoint %s: Status=%d, Size=%d bytes", tc.name, w.Code, len(responseBody))

			// Check for specific Alpha Vantage error patterns
			if strings.Contains(responseBody, "Thank you for using Alpha Vantage") ||
				strings.Contains(responseBody, "premium") ||
				strings.Contains(responseBody, "call frequency") {
				t.Logf("⚠️  Alpha Vantage rate limit detected in %s response", tc.name)
			}

			// Validate HTTP status
			assert.Equal(t, tc.expectedStatus, w.Code,
				"Expected status %d but got %d for endpoint %s. Response: %s",
				tc.expectedStatus, w.Code, tc.url, responseBody[:min(len(responseBody), 200)])

			// Validate content type
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json",
				"Response should have JSON content type")

			// Validate JSON structure if we have a validator function
			if tc.validateJSON != nil && w.Code == http.StatusOK {
				body := w.Body.Bytes()
				assert.NotEmpty(t, body, "Response body should not be empty")
				tc.validateJSON(t, body)
			}

			// Log response for debugging (only first 500 chars to avoid clutter)
			if len(responseBody) > 500 {
				responseBody = responseBody[:500] + "..."
			}
			t.Logf("Endpoint %s returned: %s", tc.name, responseBody)
		})
	}

	// Debugging: Directly call Alpha Vantage for the random symbol
	debugAlphaVantageResponse(t, cfg, symbol)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
