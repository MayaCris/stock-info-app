package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// LoggingMiddleware creates a logging middleware using the existing logger system
func LoggingMiddleware(appLogger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Start time
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Get request ID if available
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}

		// Create context with request ID
		ctx := context.Background()
		if requestID != "" {
			ctx = context.WithValue(ctx, "request_id", requestID)
		}

		// Log request start
		if raw != "" {
			path = path + "?" + raw
		}

		appLogger.Info(ctx, "HTTP Request Started",
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
			logger.String("request_id", requestID),
			logger.Int("content_length", int(c.Request.ContentLength)),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response data
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Determine log level based on status code
		logLevel := getLogLevelForStatus(statusCode)

		// Common fields
		fields := []logger.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.Int("status_code", statusCode),
			logger.Duration("latency", latency),
			logger.String("client_ip", c.ClientIP()),
			logger.String("request_id", requestID),
			logger.Int("body_size", bodySize),
		}
		// Add error information if present
		if len(c.Errors) > 0 {
			errorMsgs := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errorMsgs[i] = err.Error()
			}
			fields = append(fields, logger.Any("errors", errorMsgs))
		}

		// Log based on status code
		switch logLevel {
		case logger.InfoLevel:
			appLogger.Info(ctx, "HTTP Request Completed", fields...)
		case logger.WarnLevel:
			appLogger.Warn(ctx, "HTTP Request Completed with Warning", fields...)
		case logger.ErrorLevel:
			var err error
			if len(c.Errors) > 0 {
				err = c.Errors.Last()
			}
			appLogger.Error(ctx, "HTTP Request Failed", err, fields...)
		}
	})
}

// DetailedLoggingMiddleware provides more detailed logging including request/response bodies
func DetailedLoggingMiddleware(appLogger logger.Logger, logBodies bool) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}

		ctx := context.Background()
		if requestID != "" {
			ctx = context.WithValue(ctx, "request_id", requestID)
		}

		// Log request details
		if raw != "" {
			path = path + "?" + raw
		}

		fields := []logger.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
			logger.String("request_id", requestID),
			logger.Int("content_length", int(c.Request.ContentLength)),
			logger.String("content_type", c.GetHeader("Content-Type")),
			logger.String("accept", c.GetHeader("Accept")),
			logger.String("referer", c.GetHeader("Referer")),
		}

		// Add request headers for debugging
		if appLogger != nil {
			headerFields := make(map[string]interface{})
			for key, values := range c.Request.Header {
				if len(values) == 1 {
					headerFields[key] = values[0]
				} else {
					headerFields[key] = values
				}
			}
			fields = append(fields, logger.Any("headers", headerFields))
		}

		appLogger.Info(ctx, "Detailed HTTP Request Started", fields...)

		// Process request
		c.Next()

		// Log response
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		responseFields := []logger.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.Int("status_code", statusCode),
			logger.Duration("latency", latency),
			logger.String("latency_human", latency.String()),
			logger.String("client_ip", c.ClientIP()),
			logger.String("request_id", requestID),
			logger.Int("body_size", bodySize),
		}

		// Add response headers
		responseHeaders := make(map[string]interface{})
		for key, values := range c.Writer.Header() {
			if len(values) == 1 {
				responseHeaders[key] = values[0]
			} else {
				responseHeaders[key] = values
			}
		}
		responseFields = append(responseFields, logger.Any("response_headers", responseHeaders))

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				responseFields = append(responseFields, logger.String("error", err.Error()))
			}
		}

		logLevel := getLogLevelForStatus(statusCode)
		message := fmt.Sprintf("Detailed HTTP Request Completed - %s %s [%d]",
			c.Request.Method, path, statusCode)

		switch logLevel {
		case logger.InfoLevel:
			appLogger.Info(ctx, message, responseFields...)
		case logger.WarnLevel:
			appLogger.Warn(ctx, message, responseFields...)
		case logger.ErrorLevel:
			var err error
			if len(c.Errors) > 0 {
				err = c.Errors.Last()
			}
			appLogger.Error(ctx, message, err, responseFields...)
		}
	})
}

// getLogLevelForStatus determines the appropriate log level based on HTTP status code
func getLogLevelForStatus(statusCode int) logger.LogLevel {
	switch {
	case statusCode >= 200 && statusCode < 400:
		return logger.InfoLevel
	case statusCode >= 400 && statusCode < 500:
		return logger.WarnLevel
	case statusCode >= 500:
		return logger.ErrorLevel
	default:
		return logger.InfoLevel
	}
}

// AccessLogMiddleware provides a simple access log format
func AccessLogMiddleware(appLogger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log in common log format style
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		message := fmt.Sprintf("%s - %s \"%s %s %s\" %d %d \"%s\" \"%s\" %v",
			c.ClientIP(),
			"-", // remote user (not implemented)
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Proto,
			statusCode,
			c.Writer.Size(),
			c.GetHeader("Referer"),
			c.Request.UserAgent(),
			latency,
		)

		ctx := context.Background()
		requestID := c.GetString("request_id")
		if requestID != "" {
			ctx = context.WithValue(ctx, "request_id", requestID)
		}

		logLevel := getLogLevelForStatus(statusCode)
		switch logLevel {
		case logger.InfoLevel:
			appLogger.Info(ctx, message)
		case logger.WarnLevel:
			appLogger.Warn(ctx, message)
		case logger.ErrorLevel:
			var err error
			if len(c.Errors) > 0 {
				err = c.Errors.Last()
			}
			appLogger.Error(ctx, message, err)
		}
	})
}
