package errors

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay 1s, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay 30s, got %v", cfg.MaxDelay)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %f", cfg.Multiplier)
	}
	if cfg.ShouldRetry != nil {
		t.Error("expected ShouldRetry to be nil")
	}
}

func TestRetryWithBackoff_SucceedsImmediately(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := RetryWithBackoff(ctx, RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 0,
		MaxDelay:     0,
		Multiplier:   1.0,
		ShouldRetry:  func(error) bool { return true },
	}, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetryWithBackoff_RetriesOnError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := RetryWithBackoff(ctx, RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 0,
		MaxDelay:     0,
		Multiplier:   1.0,
		ShouldRetry:  func(error) bool { return true },
	}, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryWithBackoff_StopsWhenNotRetryable(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	permanentErr := errors.New("permanent error")
	err := RetryWithBackoff(ctx, RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 0,
		MaxDelay:     0,
		Multiplier:   1.0,
		ShouldRetry:  func(error) bool { return false },
	}, func() error {
		callCount++
		return permanentErr
	})

	if err != permanentErr {
		t.Fatalf("expected permanent error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestRetryWithBackoff_ExhaustsMaxAttempts(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := RetryWithBackoff(ctx, RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 0,
		MaxDelay:     0,
		Multiplier:   1.0,
		ShouldRetry:  func(error) bool { return true },
	}, func() error {
		callCount++
		return errors.New("always fails")
	})

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := RetryWithBackoff(ctx, RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 1 * time.Second,
		MaxDelay:     5 * time.Second,
		Multiplier:   1.0,
		ShouldRetry:  func(error) bool { return true },
	}, func() error {
		callCount++
		return errors.New("retry me")
	})

	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestRetryWithBackoff_DefaultShouldRetry(t *testing.T) {
	ctx := context.Background()

	t.Run("retries AppError with retryable type", func(t *testing.T) {
		callCount := 0
		err := RetryWithBackoff(ctx, RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 0,
			MaxDelay:     0,
			Multiplier:   1.0,
			ShouldRetry:  nil, // use default
		}, func() error {
			callCount++
			if callCount < 3 {
				return New(ErrTimeout, "timed out") // ErrTimeout is retryable
			}
			return nil
		})

		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if callCount != 3 {
			t.Errorf("expected 3 calls, got %d", callCount)
		}
	})

	t.Run("does not retry non-retryable AppError", func(t *testing.T) {
		callCount := 0
		err := RetryWithBackoff(ctx, RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 0,
			MaxDelay:     0,
			Multiplier:   1.0,
			ShouldRetry:  nil, // use default
		}, func() error {
			callCount++
			return New(ErrBadRequest, "bad input") // ErrBadRequest is NOT retryable
		})

		if err == nil {
			t.Fatal("expected error")
		}
		if callCount != 1 {
			t.Errorf("expected 1 call (no retry for non-retryable), got %d", callCount)
		}
	})

	t.Run("does not retry non-AppError", func(t *testing.T) {
		callCount := 0
		err := RetryWithBackoff(ctx, RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 0,
			MaxDelay:     0,
			Multiplier:   1.0,
			ShouldRetry:  nil,
		}, func() error {
			callCount++
			return errors.New("plain error")
		})

		if err == nil {
			t.Fatal("expected error")
		}
		if callCount != 1 {
			t.Errorf("expected 1 call (no retry for plain error), got %d", callCount)
		}
	})
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name    string
		attempt int
		config  RetryConfig
		want    time.Duration
	}{
		{
			name:    "first attempt",
			attempt: 0,
			config: RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     10 * time.Second,
				Multiplier:   2.0,
			},
			want: 100 * time.Millisecond,
		},
		{
			name:    "second attempt",
			attempt: 1,
			config: RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     10 * time.Second,
				Multiplier:   2.0,
			},
			want: 200 * time.Millisecond,
		},
		{
			name:    "third attempt",
			attempt: 2,
			config: RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     10 * time.Second,
				Multiplier:   2.0,
			},
			want: 400 * time.Millisecond,
		},
		{
			name:    "capped at max delay",
			attempt: 20,
			config: RetryConfig{
				InitialDelay: 1 * time.Second,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
			},
			want: 30 * time.Second,
		},
		{
			name:    "linear backoff multiplier 1",
			attempt: 3,
			config: RetryConfig{
				InitialDelay: 500 * time.Millisecond,
				MaxDelay:     10 * time.Second,
				Multiplier:   1.0,
			},
			want: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBackoff(tt.attempt, tt.config)
			if got != tt.want {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestRetryImmediate(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := RetryImmediate(ctx, 3, func() error {
		callCount++
		if callCount < 2 {
			return New(ErrTimeout, "timeout")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestSafeExecute_Success(t *testing.T) {
	ctx := context.Background()
	err := SafeExecute(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSafeExecute_Error(t *testing.T) {
	ctx := context.Background()
	expected := errors.New("some error")
	err := SafeExecute(ctx, func() error {
		return expected
	})
	if err != expected {
		t.Fatalf("expected original error, got %v", err)
	}
}

func TestSafeExecute_RecoversPanic(t *testing.T) {
	ctx := context.Background()
	err := SafeExecute(ctx, func() error {
		panic("something bad happened")
	})

	if err == nil {
		t.Fatal("expected error from recovered panic")
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError from recovered panic")
	}
	if appErr.Type != ErrInternal {
		t.Errorf("expected type %s, got %s", ErrInternal, appErr.Type)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	var callCount int32

	err := WithRetry(ctx, func() error {
		atomic.AddInt32(&callCount, 1)
		return New(ErrBadRequest, "bad input")
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 call, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	err := WithRetry(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWithRetry_PlainError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	expected := errors.New("plain error")
	err := WithRetry(ctx, func() error {
		callCount++
		return expected
	})

	if err != expected {
		t.Fatalf("expected original error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (non-AppError not retried), got %d", callCount)
	}
}
