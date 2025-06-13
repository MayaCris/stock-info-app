package stock_api

import (
	"time"
)

// Config holds configuration for the stock API client
type Config struct {
	BaseURL    string
	AuthToken  string
	Timeout    time.Duration
	MaxRetries int
	RateLimit  time.Duration // Delay between requests
}

// DefaultConfig returns default configuration for the stock API
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list",
		AuthToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdHRlbXB0cyI6MjUsImVtYWlsIjoiIiwiZXhwIjoxNzQ5NjkyNzEyLCJpZCI6IjAiLCJwYXNzd29yZCI6IicgT1IgJzEnPScxJyBBTkQgJyc9JyJ9.tugpIfgM-bUuX6hsa4mqfFb2DpEyRiAcoVzcsKcpv1U",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RateLimit:  100 * time.Millisecond,
	}
}

// IsValid validates the configuration
func (c *Config) IsValid() bool {
	return c.BaseURL != "" && c.AuthToken != ""
}