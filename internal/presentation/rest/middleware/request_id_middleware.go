package middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return requestid.New(
		requestid.WithGenerator(func() string {
			return uuid.New().String()
		}),
		requestid.WithCustomHeaderStrKey("X-Request-ID"),
	)
}

// CustomRequestIDMiddleware provides a custom implementation with more control
func CustomRequestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader("X-Request-ID")

		// If not provided, generate a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set the request ID in the context for use by other middlewares and handlers
		c.Set("request_id", requestID)

		// Add to response headers
		c.Header("X-Request-ID", requestID)

		// Continue processing
		c.Next()
	})
}

// RequestIDMiddlewareWithPrefix creates a request ID middleware with a custom prefix
func RequestIDMiddlewareWithPrefix(prefix string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader("X-Request-ID")

		// If not provided, generate a new one with prefix
		if requestID == "" {
			baseID := uuid.New().String()
			if prefix != "" {
				requestID = prefix + "-" + baseID
			} else {
				requestID = baseID
			}
		}

		// Set the request ID in the context
		c.Set("request_id", requestID)

		// Add to response headers
		c.Header("X-Request-ID", requestID)

		// Continue processing
		c.Next()
	})
}

// GetRequestID retrieves the request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	// Fallback to header
	return c.GetHeader("X-Request-ID")
}

// RequestTraceMiddleware adds both request ID and trace ID for distributed tracing
func RequestTraceMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Handle Request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Handle Trace ID (for distributed tracing)
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		// Handle Span ID (for distributed tracing)
		spanID := c.GetHeader("X-Span-ID")
		if spanID == "" {
			spanID = uuid.New().String()
		}
		c.Set("span_id", spanID)
		c.Header("X-Span-ID", spanID)

		// Continue processing
		c.Next()
	})
}
