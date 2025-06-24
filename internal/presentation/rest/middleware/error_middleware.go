package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/MayaCris/stock-info-app/internal/application/dto/response"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/logger"
)

// ErrorHandlingMiddleware provides centralized error handling
func ErrorHandlingMiddleware(appLogger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Handle panic
				stack := debug.Stack()
				requestID := c.GetString("request_id")

				ctx := context.Background()
				if requestID != "" {
					ctx = context.WithValue(ctx, "request_id", requestID)
				}

				appLogger.Error(ctx, "Panic recovered in HTTP handler",
					fmt.Errorf("panic: %v", err),
					logger.String("request_id", requestID),
					logger.String("method", c.Request.Method),
					logger.String("path", c.Request.URL.Path),
					logger.String("stack_trace", string(stack)),
				)

				// Return internal server error
				errorResp := response.InternalServerError("An unexpected error occurred")
				c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
				c.Abort()
			}
		}()

		// Process request
		c.Next()

		// Handle errors that occurred during request processing
		if len(c.Errors) > 0 {
			handleGinErrors(c, appLogger)
		}
	})
}

// handleGinErrors processes Gin errors and converts them to appropriate responses
func handleGinErrors(c *gin.Context, appLogger logger.Logger) {
	lastError := c.Errors.Last()
	err := lastError.Err

	requestID := c.GetString("request_id")
	ctx := context.Background()
	if requestID != "" {
		ctx = context.WithValue(ctx, "request_id", requestID)
	}

	// Check if it's already a structured error response
	var errorResp *response.ErrorResponse
	if errors.As(err, &errorResp) {
		// Log the structured error
		logStructuredError(ctx, appLogger, errorResp, c)
		c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
		return
	}

	// Handle validation errors
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		errorResp = handleValidationErrors(validationErrors)
		logStructuredError(ctx, appLogger, errorResp, c)
		c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
		return
	}

	// Handle binding errors
	if strings.Contains(err.Error(), "bind") || strings.Contains(err.Error(), "unmarshal") {
		errorResp = response.BadRequest("Invalid request format: " + err.Error())
		logStructuredError(ctx, appLogger, errorResp, c)
		c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
		return
	}

	// Handle generic errors based on error type
	switch lastError.Type {
	case gin.ErrorTypeBind:
		errorResp = response.BadRequest("Request binding failed: " + err.Error())
	case gin.ErrorTypeRender:
		errorResp = response.InternalServerError("Response rendering failed")
	case gin.ErrorTypePublic:
		errorResp = response.BadRequest(err.Error())
	default:
		// Log unexpected errors with more detail
		appLogger.Error(ctx, "Unhandled error in HTTP request",
			err,
			logger.String("request_id", requestID),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("error_type", fmt.Sprintf("%d", lastError.Type)),
		)
		errorResp = response.InternalServerError("An error occurred while processing your request")
	}

	logStructuredError(ctx, appLogger, errorResp, c)
	c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
}

// handleValidationErrors converts validation errors to structured response
func handleValidationErrors(validationErrors validator.ValidationErrors) *response.ErrorResponse {
	var fieldErrors []response.ValidationError

	for _, fieldError := range validationErrors {
		fieldErrors = append(fieldErrors, response.ValidationError{
			Field:   getJSONFieldName(fieldError),
			Message: getValidationErrorMessage(fieldError),
			Value:   fmt.Sprintf("%v", fieldError.Value()),
		})
	}

	return response.ValidationFailedWithFields(fieldErrors)
}

// getJSONFieldName extracts the JSON field name from validation error
func getJSONFieldName(fieldError validator.FieldError) string {
	// Convert struct field name to JSON field name
	// This is a simplified version - in production you might want to use struct tags
	fieldName := fieldError.Field()

	// Convert PascalCase to snake_case for JSON
	var result strings.Builder
	for i, r := range fieldName {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// getValidationErrorMessage returns a human-readable validation error message
func getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return fmt.Sprintf("Must be at least %s characters long", fieldError.Param())
	case "max":
		return fmt.Sprintf("Must be no more than %s characters long", fieldError.Param())
	case "len":
		return fmt.Sprintf("Must be exactly %s characters long", fieldError.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fieldError.Param())
	case "url":
		return "Must be a valid URL"
	case "uuid":
		return "Must be a valid UUID"
	case "gt":
		return fmt.Sprintf("Must be greater than %s", fieldError.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", fieldError.Param())
	case "lt":
		return fmt.Sprintf("Must be less than %s", fieldError.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", fieldError.Param())
	case "datetime":
		return fmt.Sprintf("Must be a valid datetime in format %s", fieldError.Param())
	default:
		return fmt.Sprintf("Invalid value for %s", fieldError.Field())
	}
}

// logStructuredError logs structured error with appropriate level
func logStructuredError(ctx context.Context, appLogger logger.Logger, errorResp *response.ErrorResponse, c *gin.Context) {
	requestID := c.GetString("request_id")

	fields := []logger.Field{
		logger.String("request_id", requestID),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
		logger.String("error_code", string(errorResp.Code)),
		logger.Int("status_code", errorResp.StatusCode),
		logger.String("client_ip", c.ClientIP()),
	}

	if errorResp.Details != nil {
		fields = append(fields, logger.Any("error_details", errorResp.Details))
	}

	// Log based on status code
	switch {
	case errorResp.StatusCode >= 500:
		appLogger.Error(ctx, fmt.Sprintf("HTTP %d: %s", errorResp.StatusCode, errorResp.Message),
			fmt.Errorf("error: %s", errorResp.Message), fields...)
	case errorResp.StatusCode >= 400:
		appLogger.Warn(ctx, fmt.Sprintf("HTTP %d: %s", errorResp.StatusCode, errorResp.Message), fields...)
	default:
		appLogger.Info(ctx, fmt.Sprintf("HTTP %d: %s", errorResp.StatusCode, errorResp.Message), fields...)
	}
}

// NotFoundMiddleware handles 404 errors
func NotFoundMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		errorResp := response.NotFound("Endpoint")
		c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
	})
}

// MethodNotAllowedMiddleware handles 405 errors
func MethodNotAllowedMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		errorResp := response.NewErrorResponse(
			response.ErrCodeBadRequest,
			fmt.Sprintf("Method %s not allowed for this endpoint", c.Request.Method),
			http.StatusMethodNotAllowed,
		)
		c.JSON(errorResp.StatusCode, errorResp.ToAPIResponse())
	})
}
