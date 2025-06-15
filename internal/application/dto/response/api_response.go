package response

import (
	"time"
)

// APIResponse is a standard wrapper for all API responses
type APIResponse[T any] struct {
	Success   bool      `json:"success"`
	Data      T         `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Meta      *Meta     `json:"meta,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Meta contains metadata for the response
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Success creates a successful API response
func Success[T any](data T) *APIResponse[T] {
	return &APIResponse[T]{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// SuccessWithMeta creates a successful API response with metadata
func SuccessWithMeta[T any](data T, meta *Meta) *APIResponse[T] {
	return &APIResponse[T]{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
	}
}

// Error creates an error API response
func Error(code, message string) *APIResponse[interface{}] {
	return &APIResponse[interface{}]{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
	}
}

// ErrorWithDetails creates an error API response with additional details
func ErrorWithDetails(code, message string, details map[string]interface{}) *APIResponse[interface{}] {
	return &APIResponse[interface{}]{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

// WithRequestID adds a request ID to the response
func (r *APIResponse[T]) WithRequestID(requestID string) *APIResponse[T] {
	r.RequestID = requestID
	return r
}

// NewMeta creates a new Meta instance for pagination
func NewMeta(page, perPage, total int) *Meta {
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
