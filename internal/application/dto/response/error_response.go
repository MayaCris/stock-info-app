package response

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standard error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest        ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeRequestTooLarge   ErrorCode = "REQUEST_TOO_LARGE"

	// Server errors (5xx)
	ErrCodeInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalAPIError   ErrorCode = "EXTERNAL_API_ERROR"
	ErrCodeCacheError         ErrorCode = "CACHE_ERROR"

	// Business logic errors
	ErrCodeResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeDuplicateResource    ErrorCode = "DUPLICATE_RESOURCE"
	ErrCodeBusinessRuleViolated ErrorCode = "BUSINESS_RULE_VIOLATED"
)

// ErrorResponse contains structured error information
type ErrorResponse struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"` // HTTP status code, not included in JSON
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewErrorResponse creates a new ErrorResponse
func NewErrorResponse(code ErrorCode, message string, statusCode int) *ErrorResponse {
	return &ErrorResponse{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error response
func (e *ErrorResponse) WithDetails(details map[string]interface{}) *ErrorResponse {
	e.Details = details
	return e
}

// ToAPIResponse converts ErrorResponse to APIResponse
func (e *ErrorResponse) ToAPIResponse() *APIResponse[interface{}] {
	return ErrorWithDetails(string(e.Code), e.Message, e.Details)
}

// Predefined error responses

// BadRequest creates a bad request error
func BadRequest(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeBadRequest, message, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *ErrorResponse {
	if message == "" {
		message = "Authentication required"
	}
	return NewErrorResponse(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *ErrorResponse {
	if message == "" {
		message = "Access denied"
	}
	return NewErrorResponse(ErrCodeForbidden, message, http.StatusForbidden)
}

// NotFound creates a not found error
func NotFound(resource string) *ErrorResponse {
	message := fmt.Sprintf("%s not found", resource)
	return NewErrorResponse(ErrCodeNotFound, message, http.StatusNotFound)
}

// Conflict creates a conflict error
func Conflict(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeConflict, message, http.StatusConflict)
}

// ValidationFailed creates a validation failed error
func ValidationFailed(message string) *ErrorResponse {
	return NewErrorResponse(ErrCodeValidationFailed, message, http.StatusBadRequest)
}

// InternalServerError creates an internal server error
func InternalServerError(message string) *ErrorResponse {
	if message == "" {
		message = "Internal server error"
	}
	return NewErrorResponse(ErrCodeInternalServer, message, http.StatusInternalServerError)
}

// ServiceUnavailable creates a service unavailable error
func ServiceUnavailable(message string) *ErrorResponse {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return NewErrorResponse(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

// DatabaseError creates a database error
func DatabaseError(message string) *ErrorResponse {
	if message == "" {
		message = "Database operation failed"
	}
	return NewErrorResponse(ErrCodeDatabaseError, message, http.StatusInternalServerError)
}

// ExternalAPIError creates an external API error
func ExternalAPIError(apiName, message string) *ErrorResponse {
	fullMessage := fmt.Sprintf("External API error (%s): %s", apiName, message)
	return NewErrorResponse(ErrCodeExternalAPIError, fullMessage, http.StatusBadGateway)
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationFailedWithFields creates a validation error with field details
func ValidationFailedWithFields(errors []ValidationError) *ErrorResponse {
	details := map[string]interface{}{
		"validation_errors": errors,
	}
	return ValidationFailed("Request validation failed").WithDetails(details)
}
