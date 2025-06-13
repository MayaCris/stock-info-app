package stock_api

import "time"

// APIResponse represents the complete response from the stock API
type APIResponse struct {
	Items    []StockRatingItem `json:"items"`
	NextPage string            `json:"next_page,omitempty"` // Key for pagination
}

// StockRatingItem represents a single stock rating item from the API
type StockRatingItem struct {
	Ticker     string `json:"ticker"`      // "BSBR"
	Company    string `json:"company"`     // "Banco Santander (Brasil)"
	Brokerage  string `json:"brokerage"`   // "The Goldman Sachs Group"
	Action     string `json:"action"`      // "upgraded by"
	RatingFrom string `json:"rating_from"` // "Sell"
	RatingTo   string `json:"rating_to"`   // "Neutral"
	TargetFrom string `json:"target_from"` // "$4.20"
	TargetTo   string `json:"target_to"`   // "$4.70"
	Time       string `json:"time"`        // "2025-01-13T00:30:05.813548892Z"
}

// GetEventTime parses the time string and returns a time.Time
func (item *StockRatingItem) GetEventTime() (time.Time, error) {
	return time.Parse(time.RFC3339, item.Time)
}

// IsValid checks if the stock rating item has required fields
func (item *StockRatingItem) IsValid() bool {
	return item.Ticker != "" && 
		   item.Company != "" && 
		   item.Brokerage != "" && 
		   item.Action != "" && 
		   item.Time != ""
}

// HasNextPage checks if there are more pages to fetch
func (response *APIResponse) HasNextPage() bool {
	return response.NextPage != ""
}

// GetItemCount returns the number of items in this response
func (response *APIResponse) GetItemCount() int {
	return len(response.Items)
}