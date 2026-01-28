package errors

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts  int              // Maximum number of retry attempts
	InitialDelay time.Duration    // Initial delay before first retry
	MaxDelay     time.Duration    // Maximum delay between retries
	Multiplier   float64          // Multiplier for exponential backoff
	ShouldRetry  func(error) bool // Custom function to determine if error is retryable
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		ShouldRetry:  nil, // Will use IsRetryable by default
	}
}

// RetryWithBackoff executes a function with retry logic and exponential backoff
func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute function
		err := fn()

		// Success
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		shouldRetry := config.ShouldRetry
		if shouldRetry == nil {
			// Use default retry logic
			shouldRetry = func(e error) bool {
				var appErr *AppError
				if errors.As(e, &appErr) {
					return appErr.IsRetryable()
				}
				return false
			}
		}

		if !shouldRetry(err) {
			return err
		}

		// Last attempt, don't wait
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt, config)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// calculateBackoff calculates the backoff delay for a given attempt
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	if delay > float64(config.MaxDelay) {
		return config.MaxDelay
	}

	return time.Duration(delay)
}

// RetryLinear executes a function with linear backoff
func RetryLinear(ctx context.Context, maxAttempts int, delay time.Duration, fn func() error) error {
	config := RetryConfig{
		MaxAttempts:  maxAttempts,
		InitialDelay: delay,
		MaxDelay:     delay * time.Duration(maxAttempts),
		Multiplier:   1.0, // Linear backoff
	}
	return RetryWithBackoff(ctx, config, fn)
}

// RetryExponential executes a function with exponential backoff
func RetryExponential(ctx context.Context, maxAttempts int, initialDelay time.Duration, fn func() error) error {
	config := RetryConfig{
		MaxAttempts:  maxAttempts,
		InitialDelay: initialDelay,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
	return RetryWithBackoff(ctx, config, fn)
}

// RetryImmediate executes a function with immediate retry (no delay)
func RetryImmediate(ctx context.Context, maxAttempts int, fn func() error) error {
	config := RetryConfig{
		MaxAttempts:  maxAttempts,
		InitialDelay: 0,
		MaxDelay:     0,
		Multiplier:   1.0,
	}
	return RetryWithBackoff(ctx, config, fn)
}

// WithRetry is a helper that wraps a function with automatic retry based on error metadata
func WithRetry(ctx context.Context, fn func() error) error {
	// First attempt
	err := fn()
	if err == nil {
		return nil
	}

	// Check if error is retryable
	var appErr *AppError
	if !errors.As(err, &appErr) || !appErr.IsRetryable() {
		return err
	}

	// Get retry strategy from error metadata
	meta := appErr.GetMetadata()

	var config RetryConfig
	switch meta.RetryStrategy {
	case RetryStrategyImmediate:
		config = RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 0,
			MaxDelay:     0,
			Multiplier:   1.0,
		}
	case RetryStrategyLinear:
		config = RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 2 * time.Second,
			MaxDelay:     10 * time.Second,
			Multiplier:   1.0,
		}
	case RetryStrategyExponential:
		config = DefaultRetryConfig()
	default:
		return err // Don't retry
	}

	// Retry with appropriate strategy
	return RetryWithBackoff(ctx, config, fn)
}

// RecoverFromPanic recovers from panic and converts it to an AppError
func RecoverFromPanic(ctx context.Context) *AppError {
	if r := recover(); r != nil {
		return New(ErrInternal, "Panic recovered").
			WithDetails(fmt.Sprintf("%v", r)).
			WithTag("panic")
	}
	return nil
}

// SafeExecute executes a function and recovers from panics
func SafeExecute(ctx context.Context, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = New(ErrInternal, "Panic recovered").
				WithDetails(fmt.Sprintf("%v", r)).
				WithTag("panic")
		}
	}()

	return fn()
}

// SafeExecuteWithRetry executes a function with panic recovery and retry logic
func SafeExecuteWithRetry(ctx context.Context, fn func() error) error {
	safeFn := func() error {
		return SafeExecute(ctx, fn)
	}

	return WithRetry(ctx, safeFn)
}
