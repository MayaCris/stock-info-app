package stock_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

// Client represents the stock API client
type Client struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
	config     *config.Config
}

// NewClient creates a new stock API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:   cfg.ThirdStockAPI.BaseURL,
		authToken: cfg.ThirdStockAPI.Auth,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: cfg,
	}
}

// FetchPage fetches a single page of stock ratings
func (c *Client) FetchPage(ctx context.Context, nextPage string) (*APIResponse, error) {
	// Build URL with query parameters
	reqURL, err := c.buildURL(nextPage)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "stock-system-backend/1.0")

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	log.Printf("✅ Fetched page with %d items, next_page: %s", 
		apiResponse.GetItemCount(), 
		apiResponse.NextPage)

	return &apiResponse, nil
}

// FetchAllPages fetches all pages using pagination until no more pages
func (c *Client) FetchAllPages(ctx context.Context) ([]StockRatingItem, error) {
	var allItems []StockRatingItem
	var nextPage string
	pageCount := 0

	log.Println("🔄 Starting to fetch all pages from API...")

	for {
		pageCount++
		log.Printf("📄 Fetching page %d (next_page: %s)", pageCount, nextPage)

		// Fetch current page
		response, err := c.FetchPage(ctx, nextPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", pageCount, err)
		}

		// Add items to our collection
		validItems := 0
		for _, item := range response.Items {
			if item.IsValid() {
				allItems = append(allItems, item)
				validItems++
			} else {
				log.Printf("⚠️  Skipping invalid item: %+v", item)
			}
		}

		log.Printf("✅ Page %d: %d total items, %d valid items", 
			pageCount, len(response.Items), validItems)

		// Check if there are more pages
		if !response.HasNextPage() {
			log.Println("🏁 No more pages to fetch")
			break
		}

		// Set next page for next iteration
		nextPage = response.NextPage

		// Optional: Add delay between requests to be respectful to the API
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond): // 100ms delay
		}
	}

	log.Printf("🎉 Successfully fetched %d total items from %d pages", 
		len(allItems), pageCount)

	return allItems, nil
}

// FetchRecentPages fetches only recent pages (useful for incremental updates)
func (c *Client) FetchRecentPages(ctx context.Context, maxPages int) ([]StockRatingItem, error) {
	var allItems []StockRatingItem
	var nextPage string
	pageCount := 0

	log.Printf("🔄 Fetching recent pages (max %d pages)...", maxPages)

	for pageCount < maxPages {
		pageCount++
		log.Printf("📄 Fetching recent page %d/%d", pageCount, maxPages)

		response, err := c.FetchPage(ctx, nextPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", pageCount, err)
		}

		// Add valid items
		for _, item := range response.Items {
			if item.IsValid() {
				allItems = append(allItems, item)
			}
		}

		// Break if no more pages
		if !response.HasNextPage() {
			log.Printf("🏁 No more pages (stopped at page %d)", pageCount)
			break
		}

		nextPage = response.NextPage

		// Add delay
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	log.Printf("✅ Fetched %d items from %d recent pages", len(allItems), pageCount)
	return allItems, nil
}

// buildURL constructs the request URL with query parameters
func (c *Client) buildURL(nextPage string) (string, error) {
	if nextPage == "" {
		return c.baseURL, nil
	}

	// Parse base URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	// Add query parameter
	query := u.Query()
	query.Set("next_page", nextPage)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

// GetStats returns client statistics (useful for monitoring)
func (c *Client) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"base_url":     c.baseURL,
		"timeout":      c.httpClient.Timeout,
		"has_auth":     c.authToken != "",
		"user_agent":   "stock-system-backend/1.0",
	}
}