package adapters

import (
	"context"
	"fmt"

	"github.com/MayaCris/stock-info-app/internal/application/usecases/population"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/external/stock_api"
)

// StockAPIDataProvider adapta la stock API a la interfaz StockDataProvider
type StockAPIDataProvider struct {
	client *stock_api.Client
}

// NewStockAPIDataProvider crea un nuevo adapter para la stock API
func NewStockAPIDataProvider(client *stock_api.Client) *StockAPIDataProvider {
	return &StockAPIDataProvider{
		client: client,
	}
}

// FetchPage implementa StockDataProvider.FetchPage
func (p *StockAPIDataProvider) FetchPage(ctx context.Context, page string) (*population.StockDataPage, error) {
	// Fetch from external API
	apiResponse, err := p.client.FetchPage(ctx, page)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page from stock API: %w", err)
	}

	// Convert API response to domain model
	items := make([]population.StockDataItem, 0, len(apiResponse.Items))
	for _, apiItem := range apiResponse.Items {
		// Validate item
		if !apiItem.IsValid() {
			continue // Skip invalid items
		}

		// Parse event time
		eventTime, err := apiItem.GetEventTime()
		if err != nil {
			continue // Skip items with invalid time
		}

		// Convert to domain model
		domainItem := population.StockDataItem{
			Ticker:     apiItem.Ticker,
			Company:    apiItem.Company,
			Brokerage:  apiItem.Brokerage,
			Action:     apiItem.Action,
			RatingFrom: apiItem.RatingFrom,
			RatingTo:   apiItem.RatingTo,
			TargetFrom: apiItem.TargetFrom,
			TargetTo:   apiItem.TargetTo,
			EventTime:  eventTime,
		}

		items = append(items, domainItem)
	}

	return &population.StockDataPage{
		Items:    items,
		NextPage: apiResponse.NextPage,
		HasMore:  apiResponse.HasNextPage(),
	}, nil
}

// GetNextPageToken implementa StockDataProvider.GetNextPageToken
func (p *StockAPIDataProvider) GetNextPageToken(currentPage string) string {
	// Para la primera página, devolver string vacío
	if currentPage == "" {
		return ""
	}
	return currentPage
}

// HasMorePages implementa StockDataProvider.HasMorePages
func (p *StockAPIDataProvider) HasMorePages(response *population.StockDataPage) bool {
	return response.HasMore
}
