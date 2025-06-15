package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// TransactionService defines the contract for managing database transactions
type TransactionService interface {
	// ExecuteInTransaction executes a function within a database transaction
	// If the function returns an error, the transaction is rolled back
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error

	// ExecuteWithRetry executes a function with retry logic for transient failures
	ExecuteWithRetry(ctx context.Context, maxRetries int, fn func(ctx context.Context) error) error
}

// TransactionServiceImpl implements TransactionService using GORM
type TransactionServiceImpl struct {
	db *gorm.DB
}

// NewTransactionService creates a new transaction service instance
func NewTransactionService(db *gorm.DB) TransactionService {
	return &TransactionServiceImpl{
		db: db,
	}
}

// ExecuteInTransaction executes a function within a database transaction
func (ts *TransactionServiceImpl) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	return ts.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}

// ExecuteWithRetry executes a function with exponential backoff retry logic
func (ts *TransactionServiceImpl) ExecuteWithRetry(ctx context.Context, maxRetries int, fn func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry based on error type
		if !isRetriableError(err) {
			return fmt.Errorf("non-retriable error on attempt %d: %w", attempt+1, err)
		}

		// If this was the last attempt, return the error
		if attempt == maxRetries {
			break
		}

		// Calculate backoff duration (exponential backoff: 100ms, 200ms, 400ms, 800ms, ...)
		backoffDuration := time.Duration(100*(1<<attempt)) * time.Millisecond

		log.Printf("⚠️ Attempt %d failed, retrying in %v: %v", attempt+1, backoffDuration, err)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry attempt %d: %w", attempt+1, ctx.Err())
		case <-time.After(backoffDuration):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("failed after %d attempts, last error: %w", maxRetries+1, lastErr)
}

// isRetriableError determines if an error is worth retrying
func isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Database connection errors (retriable)
	retriablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"deadlock",
		"serialization failure",
		"could not serialize access",
		"restart transaction",
	}

	for _, pattern := range retriablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	// Non-retriable errors (business logic, constraint violations, etc.)
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					containsMiddle(str, substr))))
}

// containsMiddle checks if substr is in the middle of str
func containsMiddle(str, substr string) bool {
	for i := 1; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
