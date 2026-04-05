package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/circuitbreaker"
)

func TestBreaker_ClosedState_AllowsRequests(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected CLOSED, got %s", cb.State())
	}
}

func TestBreaker_OpensAfterThreshold(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})
	fail := errors.New("service down")
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error { return fail })
	}
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected OPEN after 3 failures, got %s", cb.State())
	}
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
	})
	fail := errors.New("service down")
	_ = cb.Execute(func() error { return fail })
	_ = cb.Execute(func() error { return fail })
	time.Sleep(60 * time.Millisecond)
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected success in HALF_OPEN, got %v", err)
	}
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected CLOSED after successful HALF_OPEN, got %s", cb.State())
	}
}

func TestBreaker_HalfOpenFailureReOpens(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 1,
		Timeout:          50 * time.Millisecond,
	})
	fail := errors.New("down")
	_ = cb.Execute(func() error { return fail })
	time.Sleep(60 * time.Millisecond)
	err := cb.Execute(func() error { return fail })
	if err == nil {
		t.Fatal("expected error in HALF_OPEN failure")
	}
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected OPEN after HALF_OPEN failure, got %s", cb.State())
	}
}

func TestBreaker_ExecuteWithFallback(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 1,
		Timeout:          1 * time.Second,
	})

	// Trip the breaker
	_ = cb.Execute(func() error { return errors.New("fail") })

	fallbackCalled := false
	err := cb.ExecuteWithFallback(
		func() error { return nil },
		func() error {
			fallbackCalled = true
			return nil
		},
	)

	if err != nil {
		t.Fatalf("expected fallback to succeed, got %v", err)
	}
	if !fallbackCalled {
		t.Fatal("expected fallback to be called")
	}
}

func TestBreaker_SuccessResetsFailureCount(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})
	fail := errors.New("fail")
	_ = cb.Execute(func() error { return fail })
	_ = cb.Execute(func() error { return fail })
	_ = cb.Execute(func() error { return nil })
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected CLOSED after success reset, got %s", cb.State())
	}
	_ = cb.Execute(func() error { return fail })
	_ = cb.Execute(func() error { return fail })
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected still CLOSED after 2 failures, got %s", cb.State())
	}
}
