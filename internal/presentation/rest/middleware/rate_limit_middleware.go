package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
)

// RateLimiter interface for different rate limiting implementations
type RateLimiter interface {
	Allow(key string) bool
	Remaining(key string) int
	Reset(key string) time.Time
}

// InMemoryRateLimiter implements a simple in-memory rate limiter
type InMemoryRateLimiter struct {
	requests map[string]*requestTracker
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

type requestTracker struct {
	count     int
	resetTime time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(limit int, window time.Duration) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		requests: make(map[string]*requestTracker),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request is allowed for the given key
func (r *InMemoryRateLimiter) Allow(key string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	tracker, exists := r.requests[key]

	if !exists || now.After(tracker.resetTime) {
		// First request or window has reset
		r.requests[key] = &requestTracker{
			count:     1,
			resetTime: now.Add(r.window),
		}
		return true
	}

	if tracker.count >= r.limit {
		return false
	}

	tracker.count++
	return true
}

// Remaining returns the number of requests remaining for the key
func (r *InMemoryRateLimiter) Remaining(key string) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tracker, exists := r.requests[key]
	if !exists {
		return r.limit
	}

	if time.Now().After(tracker.resetTime) {
		return r.limit
	}

	remaining := r.limit - tracker.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Reset returns the time when the rate limit will reset for the key
func (r *InMemoryRateLimiter) Reset(key string) time.Time {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tracker, exists := r.requests[key]
	if !exists {
		return time.Now().Add(r.window)
	}

	return tracker.resetTime
}

// cleanup removes expired entries
func (r *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(r.window)
	defer ticker.Stop()

	for range ticker.C {
		r.mutex.Lock()
		now := time.Now()
		for key, tracker := range r.requests {
			if now.After(tracker.resetTime) {
				delete(r.requests, key)
			}
		}
		r.mutex.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rateLimitConfig config.RateLimitConfig) gin.HandlerFunc {
	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	limiter := NewInMemoryRateLimiter(rateLimitConfig.Limit, rateLimitConfig.RequestsPer)

	return func(c *gin.Context) {
		// Get key based on configuration
		key := getKeyForRequest(c, rateLimitConfig.KeyFunc)

		// Check if request is allowed
		if !limiter.Allow(key) {
			// Add rate limit headers
			remaining := limiter.Remaining(key)
			reset := limiter.Reset(key)

			c.Header("X-RateLimit-Limit", strconv.Itoa(rateLimitConfig.Limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(reset).Seconds()), 10))

			// Return rate limit exceeded error
			errorResp := response.NewErrorResponse(
				response.ErrCodeRateLimitExceeded,
				"Rate limit exceeded. Please try again later.",
				http.StatusTooManyRequests,
			).WithDetails(map[string]interface{}{
				"limit":          rateLimitConfig.Limit,
				"window_seconds": int(rateLimitConfig.RequestsPer.Seconds()),
				"retry_after":    int(time.Until(reset).Seconds()),
			})

			c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
			c.Abort()
			return
		}

		// Add rate limit headers for successful requests
		remaining := limiter.Remaining(key)
		reset := limiter.Reset(key)

		c.Header("X-RateLimit-Limit", strconv.Itoa(rateLimitConfig.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

		c.Next()
	}
}

// getKeyForRequest extracts the key for rate limiting based on configuration
func getKeyForRequest(c *gin.Context, keyFunc string) string {
	switch keyFunc {
	case "user_id":
		// Try to get user ID from context (would be set by auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			return fmt.Sprintf("user:%v", userID)
		}
		// Fallback to IP if no user ID
		return fmt.Sprintf("ip:%s", c.ClientIP())
	case "ip":
		return fmt.Sprintf("ip:%s", c.ClientIP())
	default:
		// Default to IP-based limiting
		return fmt.Sprintf("ip:%s", c.ClientIP())
	}
}

// PerEndpointRateLimitMiddleware creates endpoint-specific rate limiting
func PerEndpointRateLimitMiddleware(limits map[string]config.RateLimitConfig) gin.HandlerFunc {
	limiters := make(map[string]*InMemoryRateLimiter)

	// Create limiters for each endpoint
	for endpoint, rateLimitConfig := range limits {
		if rateLimitConfig.Enabled {
			limiters[endpoint] = NewInMemoryRateLimiter(rateLimitConfig.Limit, rateLimitConfig.RequestsPer)
		}
	}

	return func(c *gin.Context) {
		endpoint := c.Request.Method + ":" + c.FullPath()

		limiter, exists := limiters[endpoint]
		if !exists {
			// No rate limit configured for this endpoint
			c.Next()
			return
		}

		rateLimitConfig := limits[endpoint]
		key := getKeyForRequest(c, rateLimitConfig.KeyFunc) + ":" + endpoint

		if !limiter.Allow(key) {
			remaining := limiter.Remaining(key)
			reset := limiter.Reset(key)

			c.Header("X-RateLimit-Limit", strconv.Itoa(rateLimitConfig.Limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(reset).Seconds()), 10))

			errorResp := response.NewErrorResponse(
				response.ErrCodeRateLimitExceeded,
				fmt.Sprintf("Rate limit exceeded for endpoint %s", endpoint),
				http.StatusTooManyRequests,
			)

			c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
			c.Abort()
			return
		}

		// Add headers for successful requests
		remaining := limiter.Remaining(key)
		reset := limiter.Reset(key)

		c.Header("X-RateLimit-Limit", strconv.Itoa(rateLimitConfig.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

		c.Next()
	}
}
